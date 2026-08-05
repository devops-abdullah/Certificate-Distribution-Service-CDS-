package api

import (
	"github.com/devops-abdullah/cds/internal/api/dashboard"
	v1 "github.com/devops-abdullah/cds/internal/api/v1"
	"github.com/devops-abdullah/cds/internal/metrics"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	router := gin.New()

	router.Use(Logger())
	router.Use(gin.Recovery())

	// Register Prometheus endpoint
	metrics.Register(router)

	// Development-only API status dashboard
	router.GET("/dashboard", dashboard.Handler)

	api := router.Group("/api")
	v1Routes := api.Group("/v1")

	{
		v1Routes.GET("/health", v1.Health)
		v1Routes.GET("/version", v1.VersionHandler)
		v1Routes.GET("/certificates", v1.ListCertificates)
		v1Routes.GET("/certificates/:domain", v1.GetCertificate)
	}

	return router
}
