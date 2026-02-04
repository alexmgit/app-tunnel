package config

import "time"

type ClientConfig struct {
	ServerControlURL string
	ServerTunnelAddr string
	LocalForwardAddr string
	RequestedSubdomain string
	ConnPoolSize     int
	DialTimeout      time.Duration
	LogLevel         string
}

func LoadClientConfig() (ClientConfig, error) {
	serverControlURL, err := RequireString("SERVER_CONTROL_URL")
	if err != nil {
		return ClientConfig{}, err
	}
	serverTunnelAddr, err := RequireString("SERVER_TUNNEL_ADDR")
	if err != nil {
		return ClientConfig{}, err
	}
	localForwardAddr, err := RequireString("LOCAL_FORWARD_ADDR")
	if err != nil {
		return ClientConfig{}, err
	}
	requestedSubdomain := OptionalString("REQUESTED_SUBDOMAIN")
	connPoolSize, err := RequireInt("CONN_POOL_SIZE")
	if err != nil {
		return ClientConfig{}, err
	}
	dialTimeout, err := OptionalDuration("DIAL_TIMEOUT", 10*time.Second)
	if err != nil {
		return ClientConfig{}, err
	}
	logLevel := OptionalString("LOG_LEVEL")

	return ClientConfig{
		ServerControlURL: serverControlURL,
		ServerTunnelAddr: serverTunnelAddr,
		LocalForwardAddr: localForwardAddr,
		RequestedSubdomain: requestedSubdomain,
		ConnPoolSize:     connPoolSize,
		DialTimeout:      dialTimeout,
		LogLevel:         logLevel,
	}, nil
}
