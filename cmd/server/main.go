package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/acme"
	"github.com/devops-abdullah/cds/internal/api"
	v1 "github.com/devops-abdullah/cds/internal/api/v1"
	"github.com/devops-abdullah/cds/internal/certs"
	"github.com/devops-abdullah/cds/internal/config"
	"github.com/devops-abdullah/cds/internal/server"
	"github.com/devops-abdullah/cds/internal/storage"
	"github.com/devops-abdullah/cds/pkg/logger"
)

func main() {
	// Load config
	config.Load()
	// Config log
	logger.Init(config.App.LogLevel)

	certStore := storage.New()
	extractor := certs.NewExtractor(config.App.ExportDir)
	warningWindow := time.Duration(config.App.CertExpiryWarnDays) * 24 * time.Hour

	refresh(certStore, extractor, warningWindow)
	v1.SetCertificateStore(certStore)
	v1.SetReloadFunc(func() { refresh(certStore, extractor, warningWindow) })

	startWatcher(certStore, extractor, warningWindow)

	// API routes
	router := api.SetupRouter()

	runServer(router)
}

// runServer starts the HTTP server, upgrading to TLS (and to mutual TLS, if
// a client CA is configured) when TLS_CERT/TLS_KEY are set. With neither
// configured it serves plain HTTP, e.g. for local development or behind an
// external TLS terminator.
func runServer(router *gin.Engine) {
	srv := &http.Server{
		Addr:    ":" + config.App.Port,
		Handler: router,
	}

	if config.App.TLSCert == "" || config.App.TLSKey == "" {
		logger.Log.Info("Certificate Manager started on :" + config.App.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.WithError(err).Fatal("Failed to start HTTP server")
		}
		return
	}

	tlsConfig, err := server.BuildTLSConfig(config.App.TLSClientCA)
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to configure TLS")
	}
	srv.TLSConfig = tlsConfig

	mode := "TLS"
	if tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert {
		mode = "mTLS"
	}
	logger.Log.Info("Certificate Manager started on :" + config.App.Port + " (" + mode + ")")

	if err := srv.ListenAndServeTLS(config.App.TLSCert, config.App.TLSKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.WithError(err).Fatal("Failed to start HTTPS server")
	}
}

// refresh reloads the ACME store from disk, updates the in-memory
// certificate inventory, and re-exports fullchain.pem/privkey.pem for every
// certificate found. A missing or invalid ACME file is logged as a warning
// rather than fatal, so the service keeps running (e.g. before Traefik has
// issued its first certificate) and keeps serving the last known inventory.
func refresh(store storage.Repository, extractor *certs.Extractor, warningWindow time.Duration) {
	parser := acme.NewParser()

	acmeStore, err := parser.Parse(config.App.AcmeFile)
	if err != nil {
		logger.Log.WithError(err).Warn("Failed to load ACME store; keeping the last known certificate inventory")
		return
	}

	inputs := acmeStore.ToCertificateInputs()

	metadata := make([]certs.Metadata, 0, len(inputs))
	for _, in := range inputs {
		meta, err := certs.FromInput(in, warningWindow)
		if err != nil {
			logger.Log.WithError(err).WithField("domain", in.Domain).Warn("Failed to parse certificate; skipping")
			continue
		}
		if meta.ExpiringSoon {
			logger.Log.WithField("domain", meta.Domain).WithField("notAfter", meta.NotAfter).Warn("Certificate is expiring soon")
		}
		metadata = append(metadata, meta)
	}

	store.Replace(metadata)
	logger.Log.WithField("count", len(metadata)).Info("Loaded certificate inventory")

	for _, exportErr := range extractor.Export(acmeStore.ToExportInputs()) {
		logger.Log.WithError(exportErr).Warn("Failed to export certificate")
	}
}

// startWatcher watches the ACME file for changes and calls refresh whenever
// it changes, so newly issued or renewed certificates are picked up without
// restarting the service. Failure to start the watcher (e.g. the ACME
// file's directory doesn't exist yet) is logged rather than fatal; the
// service still serves whatever was loaded at startup.
func startWatcher(store storage.Repository, extractor *certs.Extractor, warningWindow time.Duration) {
	watcher := acme.NewWatcher(config.App.AcmeFile)

	go func() {
		err := watcher.Start(nil, func() {
			logger.Log.Info("ACME file changed; reloading certificate inventory")
			refresh(store, extractor, warningWindow)
		}, func(err error) {
			logger.Log.WithError(err).Warn("ACME file watcher error")
		})
		if err != nil {
			logger.Log.WithError(err).Warn("Failed to start ACME file watcher; live reload disabled")
		}
	}()
}
