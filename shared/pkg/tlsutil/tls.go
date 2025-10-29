package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// ServerTLSConfig creates TLS configuration for gRPC servers
// certFile: path to server certificate file (.pem)
// keyFile: path to server private key file (.pem)
func ServerTLSConfig(certFile, keyFile string) (credentials.TransportCredentials, error) {
	// Load server certificate and private key
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server TLS credentials: %w", err)
	}

	return creds, nil
}

// ClientTLSConfig creates TLS configuration for gRPC clients
// caFile: path to CA certificate file (.pem)
// serverName: expected server name for certificate validation (optional, can be empty)
func ClientTLSConfig(caFile, serverName string) (credentials.TransportCredentials, error) {
	// Load CA certificate
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Create certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		ServerName:         serverName,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false, // Always verify in production
	}

	return credentials.NewTLS(tlsConfig), nil
}

// HTTPServerTLSConfig creates TLS configuration for HTTP servers (Gin, etc.)
// This provides a secure TLS config with modern cipher suites
func HTTPServerTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites (Go automatically includes these)
			// TLS 1.2 cipher suites
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}
}

// InsecureClientTLSConfig creates TLS config that skips verification
// WARNING: Only use this for development/testing!
func InsecureClientTLSConfig() (credentials.TransportCredentials, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip certificate verification
	}

	return credentials.NewTLS(tlsConfig), nil
}

// IsTLSEnabled checks if TLS is enabled based on certificate paths
func IsTLSEnabled(certFile, keyFile string) bool {
	if certFile == "" || keyFile == "" {
		return false
	}

	// Check if files exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return false
	}

	return true
}
