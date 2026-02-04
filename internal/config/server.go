package config

import "time"

type ServerConfig struct {
	ControlAddr        string
	TunnelAddr         string
	HTTPAddr           string
	HTTPSAddr          string
	TLSCertFile        string
	TLSKeyFile         string
	Domain             string
	SubdomainStorePath string
	SubdomainLength    int
	TunnelTimeout      time.Duration
	LogLevel           string
}

func LoadServerConfig() (ServerConfig, error) {
	controlAddr, err := RequireString("CONTROL_ADDR")
	if err != nil {
		return ServerConfig{}, err
	}
	tunnelAddr, err := RequireString("TUNNEL_ADDR")
	if err != nil {
		return ServerConfig{}, err
	}
	httpAddr, err := RequireString("HTTP_ADDR")
	if err != nil {
		return ServerConfig{}, err
	}
	httpsAddr := OptionalString("HTTPS_ADDR")
	tlsCertFile := OptionalString("TLS_CERT_FILE")
	tlsKeyFile := OptionalString("TLS_KEY_FILE")
	domain, err := RequireString("DOMAIN")
	if err != nil {
		return ServerConfig{}, err
	}
	subdomainStorePath, err := RequireString("SUBDOMAIN_STORE_PATH")
	if err != nil {
		return ServerConfig{}, err
	}
	subdomainLength, err := RequireInt("SUBDOMAIN_LENGTH")
	if err != nil {
		return ServerConfig{}, err
	}
	tunnelTimeout, err := OptionalDuration("TUNNEL_TIMEOUT", 30*time.Second)
	if err != nil {
		return ServerConfig{}, err
	}
	logLevel := OptionalString("LOG_LEVEL")

	return ServerConfig{
		ControlAddr:        controlAddr,
		TunnelAddr:         tunnelAddr,
		HTTPAddr:           httpAddr,
		HTTPSAddr:          httpsAddr,
		TLSCertFile:        tlsCertFile,
		TLSKeyFile:         tlsKeyFile,
		Domain:             domain,
		SubdomainStorePath: subdomainStorePath,
		SubdomainLength:    subdomainLength,
		TunnelTimeout:      tunnelTimeout,
		LogLevel:           logLevel,
	}, nil
}
