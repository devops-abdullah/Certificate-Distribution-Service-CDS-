// Package server builds the TLS configuration used to serve the API,
// including optional mutual TLS.
package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// BuildTLSConfig returns the TLS configuration to serve with. If
// clientCAPath is empty, the returned config serves plain server-side TLS
// with no client certificate requirement. If clientCAPath is set, clients
// must present a certificate signed by that CA (mutual TLS).
func BuildTLSConfig(clientCAPath string) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if clientCAPath == "" {
		return cfg, nil
	}

	caCert, err := os.ReadFile(clientCAPath)
	if err != nil {
		return nil, fmt.Errorf("read client CA: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("parse client CA: no valid certificates found in %s", clientCAPath)
	}

	cfg.ClientCAs = pool
	cfg.ClientAuth = tls.RequireAndVerifyClientCert

	return cfg, nil
}
