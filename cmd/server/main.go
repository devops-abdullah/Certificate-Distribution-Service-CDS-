package main

import (
	"github.com/devops-abdullah/cds/internal/acme"
	"github.com/devops-abdullah/cds/internal/api"
	v1 "github.com/devops-abdullah/cds/internal/api/v1"
	"github.com/devops-abdullah/cds/internal/certs"
	"github.com/devops-abdullah/cds/internal/config"
	"github.com/devops-abdullah/cds/internal/storage"
	"github.com/devops-abdullah/cds/pkg/logger"
)

func main() {
	// Load config
	config.Load()
	// Config log
	logger.Init(config.App.LogLevel)

	// Load certificate metadata from the ACME store
	certStore := storage.New()
	loadCertificates(certStore)
	v1.SetCertificateStore(certStore)

	// API routes
	router := api.SetupRouter()

	logger.Log.Info("Certificate Manager started on :" + config.App.Port)
	// HTTP server start
	if err := router.Run(":" + config.App.Port); err != nil {
		logger.Log.WithError(err).Fatal("Failed to start HTTP server")
	}
}

// loadCertificates reads the configured ACME store and populates the
// in-memory certificate inventory. A missing or invalid ACME file is logged
// as a warning rather than a fatal error, so the service can still start
// (e.g. before Traefik has issued its first certificate).
func loadCertificates(store *storage.Store) {
	parser := acme.NewParser()

	acmeStore, err := parser.Parse(config.App.AcmeFile)
	if err != nil {
		logger.Log.WithError(err).Warn("Failed to load ACME store; starting with an empty certificate inventory")
		return
	}

	inputs := acmeStore.ToCertificateInputs()

	metadata := make([]certs.Metadata, 0, len(inputs))
	for _, in := range inputs {
		meta, err := certs.FromInput(in)
		if err != nil {
			logger.Log.WithError(err).WithField("domain", in.Domain).Warn("Failed to parse certificate; skipping")
			continue
		}
		metadata = append(metadata, meta)
	}

	store.Replace(metadata)
	logger.Log.WithField("count", len(metadata)).Info("Loaded certificate inventory")
}
