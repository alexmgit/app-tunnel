package config

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type ClientConfig struct {
	ServerControlURL   string
	ServerTunnelAddr   string
	LocalForwardAddr   string
	RequestedSubdomain string
	ConnPoolSize       int
	DialTimeout        time.Duration
	LogLevel           string
}

func LoadClientConfig() (ClientConfig, error) {
	localForwardAddr, err := RequireString("LOCAL_FORWARD_ADDR")
	if err != nil {
		return ClientConfig{}, err
	}

	serverControlURL, serverTunnelAddr, err := resolveServerEndpoints()
	if err != nil {
		return ClientConfig{}, err
	}

	requestedSubdomain := OptionalString("REQUESTED_SUBDOMAIN")
	connPoolSize, err := OptionalInt("CONN_POOL_SIZE", 4)
	if err != nil {
		return ClientConfig{}, err
	}
	dialTimeout, err := OptionalDuration("DIAL_TIMEOUT", 10*time.Second)
	if err != nil {
		return ClientConfig{}, err
	}
	logLevel := OptionalString("LOG_LEVEL")

	return ClientConfig{
		ServerControlURL:   serverControlURL,
		ServerTunnelAddr:   serverTunnelAddr,
		LocalForwardAddr:   localForwardAddr,
		RequestedSubdomain: requestedSubdomain,
		ConnPoolSize:       connPoolSize,
		DialTimeout:        dialTimeout,
		LogLevel:           logLevel,
	}, nil
}

func resolveServerEndpoints() (string, string, error) {
	serverAddr := OptionalString("SERVER_ADDR")
	if serverAddr == "" {
		serverControlURL, err := RequireString("SERVER_CONTROL_URL")
		if err != nil {
			return "", "", err
		}
		serverTunnelAddr, err := RequireString("SERVER_TUNNEL_ADDR")
		if err != nil {
			return "", "", err
		}
		return serverControlURL, serverTunnelAddr, nil
	}

	if strings.Contains(serverAddr, "://") || strings.Contains(serverAddr, "/") || strings.Contains(serverAddr, ":") {
		return "", "", fmt.Errorf("invalid SERVER_ADDR: expected host/domain without scheme, path, or port")
	}

	controlURL := OptionalString("SERVER_CONTROL_URL")
	if controlURL == "" {
		controlHost := OptionalString("SERVER_CONTROL_HOST")
		if controlHost == "" {
			controlHost = "control." + serverAddr
		}
		controlScheme := OptionalString("SERVER_CONTROL_SCHEME")
		if controlScheme == "" {
			controlScheme = "https"
		}
		controlURL = fmt.Sprintf("%s://%s/register", controlScheme, controlHost)
	}

	tunnelAddr := OptionalString("SERVER_TUNNEL_ADDR")
	if tunnelAddr == "" {
		tunnelPort := OptionalString("SERVER_TUNNEL_PORT")
		if tunnelPort == "" {
			tunnelPort = "8081"
		}
		tunnelAddr = net.JoinHostPort(serverAddr, tunnelPort)
	}

	return controlURL, tunnelAddr, nil
}
