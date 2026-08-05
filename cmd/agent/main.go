package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/devops-abdullah/cds/internal/agent"
	"github.com/devops-abdullah/cds/pkg/logger"
)

func main() {
	agent.Load()
	logger.Init("info")

	if len(agent.App.Domains) == 0 {
		logger.Log.Fatal("No DOMAINS configured; nothing for the agent to do")
	}
	if agent.App.APIKey == "" {
		logger.Log.Fatal("API_KEY is required")
	}

	httpClient, err := buildHTTPClient(agent.App)
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to configure TLS for the manager connection")
	}

	client := agent.NewClient(agent.App.ManagerURL, agent.App.APIKey, httpClient)
	installer := agent.NewInstaller(agent.App.InstallDir)

	logger.Log.
		WithField("domains", agent.App.Domains).
		WithField("manager", agent.App.ManagerURL).
		WithField("pollInterval", agent.App.PollInterval.String()).
		Info("Certificate Agent started")

	stop := make(chan struct{})
	agent.Run(agent.App, client, installer, logEvent, stop)
}

func logEvent(event string, fields map[string]interface{}) {
	entry := logger.Log.WithField("event", event)
	for k, v := range fields {
		entry = entry.WithField(k, v)
	}
	entry.Info("agent poll")
}

// buildHTTPClient returns nil (agent.NewClient then uses a plain default)
// unless a client certificate or a CA to trust is configured.
func buildHTTPClient(cfg agent.Config) (*http.Client, error) {
	if cfg.TLSClientCert == "" && cfg.TLSCA == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	if cfg.TLSClientCert != "" && cfg.TLSClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSClientCert, cfg.TLSClientKey)
		if err != nil {
			return nil, fmt.Errorf("load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.TLSCA != "" {
		caCert, err := os.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("read CA: %w", err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("parse CA: no valid certificates found in %s", cfg.TLSCA)
		}
		tlsConfig.RootCAs = pool
	}

	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}, nil
}
