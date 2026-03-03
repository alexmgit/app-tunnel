package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"app-tunnel/internal/logging"
	"app-tunnel/internal/protocol"
)

type Server struct {
	registry     *TunnelRegistry
	domain       string
	tunnelTimeout time.Duration
	log          *logging.Logger
}

func NewServer(registry *TunnelRegistry, domain string, tunnelTimeout time.Duration, logger *logging.Logger) *Server {
	return &Server{
		registry:      registry,
		domain:        domain,
		tunnelTimeout: tunnelTimeout,
		log:           logger,
	}
}

func (s *Server) ControlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.log.Warnf("control invalid method=%s remote=%s", r.Method, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload protocol.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.log.Warnf("control decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	subdomain, err := s.registry.RegisterSubdomain(payload.RequestedSubdomain)
	if err != nil {
		s.log.Warnf("control register error: %v", err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	s.log.Infof("registered subdomain=%s requested=%s remote=%s", subdomain, payload.RequestedSubdomain, r.RemoteAddr)

	response := protocol.RegisterResponse{
		Subdomain: subdomain,
		Domain:    s.domain,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) CaddyAskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	host := r.URL.Query().Get("domain")
	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !isAllowedCaddyHost(host, s.domain) {
		s.log.Warnf("caddy ask denied host=%s", host)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleTunnel(conn net.Conn) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		s.log.Warnf("tunnel handshake read error: %v", err)
		_ = conn.Close()
		return
	}
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 || parts[0] != "SUBDOMAIN" {
		s.log.Warnf("tunnel handshake invalid: %q", line)
		_ = conn.Close()
		return
	}
	subdomain := parts[1]
	if err := s.registry.RegisterTunnel(subdomain, NewTunnelConnWithReader(conn, reader)); err != nil {
		s.log.Warnf("tunnel register failed subdomain=%s error=%v", subdomain, err)
		_ = conn.Close()
		return
	}
	s.log.Debugf("tunnel registered subdomain=%s remote=%s", subdomain, conn.RemoteAddr().String())
}

func (s *Server) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	subdomain, ok := extractSubdomain(r.Host, s.domain)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tunnelConn, err := s.registry.Acquire(subdomain, s.tunnelTimeout)
	if err != nil {
		s.log.Warnf("no tunnel available subdomain=%s error=%v", subdomain, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	healthy := true
	defer func() {
		if !healthy {
			_ = tunnelConn.Conn.Close()
			return
		}
		if err := s.registry.RegisterTunnel(subdomain, tunnelConn); err != nil {
			_ = tunnelConn.Conn.Close()
			return
		}
		_ = tunnelConn.Conn.SetDeadline(time.Time{})
	}()

	ctx, cancel := context.WithTimeout(r.Context(), s.tunnelTimeout)
	defer cancel()

	s.log.Debugf("proxy start subdomain=%s method=%s path=%s host=%s remote=%s", subdomain, r.Method, r.URL.Path, r.Host, r.RemoteAddr)
	if err := writeRequest(ctx, tunnelConn.Conn, r); err != nil {
		healthy = false
		s.log.Warnf("proxy write error subdomain=%s error=%v", subdomain, err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	resp, err := readResponse(ctx, tunnelConn, r)
	if err != nil {
		healthy = false
		s.log.Warnf("proxy read error subdomain=%s error=%v", subdomain, err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	s.log.Debugf("proxy done subdomain=%s status=%d", subdomain, resp.StatusCode)
}

func extractSubdomain(host string, domain string) (string, bool) {
	host = stripPort(host)
	if host == domain {
		return "", false
	}
	if !strings.HasSuffix(host, "."+domain) {
		return "", false
	}
	trimmed := strings.TrimSuffix(host, "."+domain)
	if trimmed == "" || strings.Contains(trimmed, ".") {
		return "", false
	}
	return trimmed, true
}

func isAllowedCaddyHost(host string, domain string) bool {
	host = strings.ToLower(stripPort(strings.TrimSpace(host)))
	domain = strings.ToLower(strings.TrimSpace(domain))
	if host == "" || domain == "" {
		return false
	}
	if host == domain || host == "control."+domain {
		return true
	}
	_, ok := extractSubdomain(host, domain)
	return ok
}

func stripPort(host string) string {
	if strings.Contains(host, "]") {
		return host
	}
	parts := strings.Split(host, ":")
	if len(parts) > 1 {
		return parts[0]
	}
	return host
}

func copyHeader(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func writeRequest(ctx context.Context, conn net.Conn, req *http.Request) error {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetWriteDeadline(deadline)
	}
	if req.Body != nil {
		defer req.Body.Close()
	}
	return req.Write(conn)
}

func readResponse(ctx context.Context, conn *TunnelConn, req *http.Request) (*http.Response, error) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.Conn.SetReadDeadline(deadline)
	}
	return http.ReadResponse(conn.Reader, req)
}

func StartTunnelListener(addr string, handler func(net.Conn), logger *logging.Logger) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			if logger != nil {
				logger.Warnf("tunnel accept error: %v", err)
			}
			continue
		}
		if logger != nil {
			logger.Debugf("tunnel connection accepted remote=%s", conn.RemoteAddr().String())
		}
		go handler(conn)
	}
}

func StartHTTPServer(addr string, handler http.Handler) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return server.ListenAndServe()
}

func StartHTTPSServer(addr string, certFile string, keyFile string, handler http.Handler) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return server.ListenAndServeTLS(certFile, keyFile)
}

func NewProxyMux(proxyHandler http.Handler, controlHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/register", controlHandler)
	mux.Handle("/", proxyHandler)
	return mux
}

func NewControlMux(registerHandler http.Handler, caddyAskHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/register", registerHandler)
	mux.Handle("/caddy/allow", caddyAskHandler)
	mux.Handle("/", http.NotFoundHandler())
	return mux
}

func (s *Server) ValidateTLSConfig(httpsAddr, certFile, keyFile string) error {
	if httpsAddr == "" {
		return nil
	}
	if certFile == "" || keyFile == "" {
		return fmt.Errorf("TLS cert and key are required when HTTPS_ADDR is set")
	}
	return nil
}
