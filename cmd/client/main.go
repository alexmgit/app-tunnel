package main

import (
	"log"
	"os"

	"app-tunnel/internal/client"
	"app-tunnel/internal/config"
	"app-tunnel/internal/logging"
)

func main() {
	cfg, err := config.LoadClientConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	level, err := logging.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("log level error: %v", err)
	}
	logger := logging.NewLogger("client ", level)

	cl := client.NewClient(
		cfg.ServerControlURL,
		cfg.ServerTunnelAddr,
		cfg.LocalForwardAddr,
		cfg.RequestedSubdomain,
		cfg.ConnPoolSize,
		cfg.DialTimeout,
		logger,
	)

	response, err := cl.Register()
	if err != nil {
		logger.Errorf("register error: %v", err)
		os.Exit(1)
	}

	cl.Run(response.Subdomain)
}
