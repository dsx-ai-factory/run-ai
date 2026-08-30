package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// NewHTTPClientWithCA creates an HTTP client configured to trust a custom CA certificate.
// If caCertPath is empty, returns a standard HTTP client with default system trust store.
func NewHTTPClientWithCA(caCertPath string, timeout time.Duration) (*http.Client, error) {
	if caCertPath == "" {
		return &http.Client{
			Timeout: timeout,
		}, nil
	}

	tlsConfig, err := newTLSConfig(caCertPath)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

// NewTransportWithCA creates an http.Transport configured to trust a custom CA certificate.
// If caCertPath is empty, returns a standard transport with default system trust store.
func NewTransportWithCA(caCertPath string) (*http.Transport, error) {
	if caCertPath == "" {
		return &http.Transport{}, nil
	}

	tlsConfig, err := newTLSConfig(caCertPath)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}, nil
}

// newTLSConfig creates a tls.Config that trusts the specified CA certificate
func newTLSConfig(caCertPath string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate from %s: %w", caCertPath, err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate from %s", caCertPath)
	}

	return &tls.Config{
		RootCAs: caCertPool,
	}, nil
}
