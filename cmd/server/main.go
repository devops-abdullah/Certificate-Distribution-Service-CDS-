package main

import (
	"github.com/devops-abdullah/cds/internal/api"
	"github.com/devops-abdullah/cds/internal/config"
	"github.com/devops-abdullah/cds/pkg/logger"
)

func main() {
	// Load config
	config.Load()
	// Config log
	logger.Init(config.App.LogLevel)
	// API routes
	router := api.SetupRouter()

	logger.Log.Info("Certificate Manager started on :" + config.App.Port)
	// HTTP server start
	if err := router.Run(":" + config.App.Port); err != nil {
		logger.Log.WithError(err).Fatal("Failed to start HTTP server")
	}
}