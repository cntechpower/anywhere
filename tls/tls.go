package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"
)

var defaultDialer = &net.Dialer{Timeout: 5 * time.Second}

func loadKeyPair(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load key pair (%s, %s): %w", certFile, keyFile, err)
	}
	return cert, nil
}

func loadCertPool(caFile string) (*x509.CertPool, error) {
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read ca file %s: %w", caFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("no certificate found in %s", caFile)
	}
	return pool, nil
}

// ParseServerTlsConfig builds a mutual TLS server configuration. Client
// certificates are required and verified against caFile.
func ParseServerTlsConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := loadKeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	pool, err := loadCertPool(caFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// ParseClientTlsConfig builds the agent side configuration.
//
// serverName must match a Subject Alternative Name of the server certificate.
// insecure disables server certificate verification and should only be used to
// talk to servers that still run with a legacy certificate.
func ParseClientTlsConfig(certFile, keyFile, caFile, serverName string, insecure bool) (*tls.Config, error) {
	cert, err := loadKeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	pool, err := loadCertPool(caFile)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   serverName,
		//nolint:gosec // explicit opt-in escape hatch for legacy certificates
		InsecureSkipVerify: insecure,
	}
	return cfg, nil
}

// ParseTlsConfig is kept for backwards compatibility and returns a server
// configuration. New code should use ParseServerTlsConfig or
// ParseClientTlsConfig explicitly.
func ParseTlsConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	return ParseServerTlsConfig(certFile, keyFile, caFile)
}

func DialTlsServer(ip string, port int, config *tls.Config) (*tls.Conn, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("wrong format of ip: %s", ip)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("wrong format of port: %d", port)
	}
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	return tls.DialWithDialer(defaultDialer, "tcp", addr, config)
}
