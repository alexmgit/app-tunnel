package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"app-tunnel/internal/config"
	"app-tunnel/internal/logging"
	"app-tunnel/internal/server"
)

func main() {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	level, err := logging.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("log level error: %v", err)
	}
	logger := logging.NewLogger("server ", level)

	store, err := server.NewSubdomainStore(cfg.SubdomainStorePath, cfg.SubdomainLength)
	if err != nil {
		logger.Errorf("store error: %v", err)
		os.Exit(1)
	}

	registry := server.NewTunnelRegistry(store, cfg.TunnelTimeout)
	svc := server.NewServer(registry, cfg.Domain, cfg.TunnelTimeout, logger)
	if err := svc.ValidateTLSConfig(cfg.HTTPSAddr, cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
		logger.Errorf("tls config error: %v", err)
		os.Exit(1)
	}

	stop := make(chan struct{})
	registry.StartReaper(stop)

	errCh := make(chan error, 4)
	go func() {
		logger.Infof("tunnel listener starting addr=%s", cfg.TunnelAddr)
		errCh <- server.StartTunnelListener(cfg.TunnelAddr, svc.HandleTunnel, logger)
	}()
	go func() {
		controlMux := server.NewControlMux(httpHandlerFunc(svc.ControlHandler), httpHandlerFunc(svc.CaddyAskHandler))
		logger.Infof("control listener starting addr=%s", cfg.ControlAddr)
		errCh <- server.StartHTTPServer(cfg.ControlAddr, controlMux)
	}()
	go func() {
		proxyMux := server.NewProxyMux(httpHandlerFunc(svc.ProxyHandler), httpNotFoundHandler{})
		logger.Infof("http listener starting addr=%s", cfg.HTTPAddr)
		errCh <- server.StartHTTPServer(cfg.HTTPAddr, proxyMux)
	}()
	if cfg.HTTPSAddr != "" {
		go func() {
			proxyMux := server.NewProxyMux(httpHandlerFunc(svc.ProxyHandler), httpNotFoundHandler{})
			logger.Infof("https listener starting addr=%s", cfg.HTTPSAddr)
			errCh <- server.StartHTTPSServer(cfg.HTTPSAddr, cfg.TLSCertFile, cfg.TLSKeyFile, proxyMux)
		}()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Errorf("server error: %v", err)
		os.Exit(1)
	case <-interrupt:
		close(stop)
		logger.Infof("shutdown")
	}
}

type httpHandlerFunc func(w http.ResponseWriter, r *http.Request)

type httpNotFoundHandler struct{}

func (f httpHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) { f(w, r) }

func (httpNotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}
