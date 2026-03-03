package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadClientConfigFromServerAddr(t *testing.T) {
	t.Setenv("SERVER_ADDR", "example.com")
	t.Setenv("LOCAL_FORWARD_ADDR", "127.0.0.1:3000")
	t.Setenv("CONN_POOL_SIZE", "2")
	t.Setenv("DIAL_TIMEOUT", "10s")

	cfg, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.ServerControlURL != "https://control.example.com/register" {
		t.Fatalf("unexpected control url: %s", cfg.ServerControlURL)
	}
	if cfg.ServerTunnelAddr != "example.com:8081" {
		t.Fatalf("unexpected tunnel addr: %s", cfg.ServerTunnelAddr)
	}
	if cfg.DialTimeout != 10*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.DialTimeout)
	}
}

func TestLoadClientConfigFromServerAddrWithOverrides(t *testing.T) {
	t.Setenv("SERVER_ADDR", "example.com")
	t.Setenv("SERVER_CONTROL_HOST", "api.example.com")
	t.Setenv("SERVER_CONTROL_SCHEME", "http")
	t.Setenv("SERVER_TUNNEL_PORT", "9001")
	t.Setenv("LOCAL_FORWARD_ADDR", "127.0.0.1:3000")
	t.Setenv("CONN_POOL_SIZE", "2")

	cfg, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.ServerControlURL != "http://api.example.com/register" {
		t.Fatalf("unexpected control url: %s", cfg.ServerControlURL)
	}
	if cfg.ServerTunnelAddr != "example.com:9001" {
		t.Fatalf("unexpected tunnel addr: %s", cfg.ServerTunnelAddr)
	}
}

func TestLoadClientConfigLegacyMode(t *testing.T) {
	t.Setenv("SERVER_CONTROL_URL", "https://control.example.com/register")
	t.Setenv("SERVER_TUNNEL_ADDR", "example.com:8081")
	t.Setenv("LOCAL_FORWARD_ADDR", "127.0.0.1:3000")
	t.Setenv("CONN_POOL_SIZE", "2")

	cfg, err := LoadClientConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.ServerControlURL != "https://control.example.com/register" {
		t.Fatalf("unexpected control url: %s", cfg.ServerControlURL)
	}
	if cfg.ServerTunnelAddr != "example.com:8081" {
		t.Fatalf("unexpected tunnel addr: %s", cfg.ServerTunnelAddr)
	}
}

func TestLoadClientConfigInvalidServerAddr(t *testing.T) {
	t.Setenv("SERVER_ADDR", "https://example.com")
	t.Setenv("LOCAL_FORWARD_ADDR", "127.0.0.1:3000")
	t.Setenv("CONN_POOL_SIZE", "2")

	_, err := LoadClientConfig()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "invalid SERVER_ADDR") {
		t.Fatalf("unexpected error: %v", err)
	}
}
