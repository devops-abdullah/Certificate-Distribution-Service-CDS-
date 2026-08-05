package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/pkg/logger"
)

// AuditLog records who did what: method, path, resulting status, latency,
// remote address, and the authenticated role (if any), tagged with
// "audit": true so these events can be filtered out of general application
// logs downstream. Run it after APIKeyAuth if role information matters.
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		role, _ := c.Get(roleContextKey)

		logger.Log.WithFields(map[string]interface{}{
			"audit":    true,
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"role":     role,
			"remoteIP": c.ClientIP(),
			"duration": time.Since(start).String(),
		}).Info("audit event")
	}
}
