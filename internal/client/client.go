package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"app-tunnel/internal/logging"
	"app-tunnel/internal/protocol"
)

type Client struct {
	serverControlURL string
	serverTunnelAddr string
	localForwardAddr string
	requestedSubdomain string
	connPoolSize     int
	dialTimeout      time.Duration
	log              *logging.Logger
}

func NewClient(serverControlURL, serverTunnelAddr, localForwardAddr, requestedSubdomain string, connPoolSize int, dialTimeout time.Duration, logger *logging.Logger) *Client {
	return &Client{
		serverControlURL: serverControlURL,
		serverTunnelAddr: serverTunnelAddr,
		localForwardAddr: localForwardAddr,
		requestedSubdomain: requestedSubdomain,
		connPoolSize:     connPoolSize,
		dialTimeout:      dialTimeout,
		log:              logger,
	}
}

func (c *Client) Register() (protocol.RegisterResponse, error) {
	payload := protocol.RegisterRequest{RequestedSubdomain: c.requestedSubdomain}
	body, err := json.Marshal(payload)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}
	c.log.Debugf("register request url=%s requested_subdomain=%s", c.serverControlURL, c.requestedSubdomain)
	request, err := http.NewRequest(http.MethodPost, c.serverControlURL, bytes.NewReader(body))
	if err != nil {
		return protocol.RegisterResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: c.dialTimeout}
	resp, err := client.Do(request)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return protocol.RegisterResponse{}, fmt.Errorf("register failed: %s", resp.Status)
	}

	var response protocol.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol.RegisterResponse{}, err
	}
	c.log.Infof("registered subdomain=%s domain=%s", response.Subdomain, response.Domain)
	return response, nil
}

func (c *Client) Run(subdomain string) {
	for i := 0; i < c.connPoolSize; i++ {
		go c.tunnelWorker(subdomain)
	}
	select {}
}

func (c *Client) tunnelWorker(subdomain string) {
	for {
		if err := c.handleTunnel(subdomain); err != nil {
			c.log.Warnf("tunnel error: %v", err)
			backoff(c.dialTimeout)
		}
	}
}

func (c *Client) handleTunnel(subdomain string) error {
	dialer := net.Dialer{Timeout: c.dialTimeout}
	conn, err := dialer.Dial("tcp", c.serverTunnelAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "SUBDOMAIN %s\n", subdomain); err != nil {
		return err
	}
	c.log.Debugf("tunnel connected remote=%s subdomain=%s", conn.RemoteAddr().String(), subdomain)

	reader := bufio.NewReader(conn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			return err
		}
		req.RequestURI = req.URL.RequestURI()
		c.log.Debugf("request start method=%s path=%s host=%s", req.Method, req.URL.Path, req.Host)

		resp, err := c.forwardToLocal(req)
		if err != nil {
			c.log.Warnf("forward error: %v", err)
			return err
		}
		if err := resp.Write(conn); err != nil {
			_ = resp.Body.Close()
			return err
		}
		c.log.Debugf("request done status=%d", resp.StatusCode)
		_ = resp.Body.Close()
	}
}

func (c *Client) forwardToLocal(req *http.Request) (*http.Response, error) {
	dialer := net.Dialer{Timeout: c.dialTimeout}
	localConn, err := dialer.Dial("tcp", c.localForwardAddr)
	if err != nil {
		return nil, err
	}
	defer localConn.Close()

	if err := req.Write(localConn); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(localConn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func backoff(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}
