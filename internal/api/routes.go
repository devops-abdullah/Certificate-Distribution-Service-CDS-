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
	}

	// Certificate data and the manual reload trigger require an API key.
	// AuditLog runs first so failed/unauthorized attempts are recorded too,
	// not just successfully authenticated requests.
	protected := v1Routes.Group("")
	protected.Use(AuditLog(), APIKeyAuth())
	{
		protected.GET("/certificates", RequireRole(RoleReadOnly), v1.ListCertificates)
		protected.GET("/certificates/:domain", RequireRole(RoleReadOnly), v1.GetCertificate)
		protected.POST("/reload", RequireRole(RoleAdmin), v1.Reload)
	}

	return router
}
