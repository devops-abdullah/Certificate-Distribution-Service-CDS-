package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	v1 "github.com/devops-abdullah/cds/internal/api/v1"
	"github.com/devops-abdullah/cds/internal/config"
)

// Role is the permission level granted to an authenticated request.
type Role string

const (
	// RoleReadOnly can read certificate inventory data.
	RoleReadOnly Role = "readonly"
	// RoleAdmin can read certificate inventory data and trigger mutating
	// actions (e.g. a manual reload).
	RoleAdmin Role = "admin"
	// RoleAgent is a least-privilege role for Certificate Agents: it can
	// only fetch a certificate bundle (cert + private key) for a domain,
	// never list inventory or trigger a reload.
	RoleAgent Role = "agent"

	roleContextKey = "role"
)

// APIKeyAuth authenticates requests against the configured read-only,
// admin, and agent API keys. If none are configured, the API fails closed
// (503) rather than serving unauthenticated. The matched role is stored
// in the request context for RequireRole to consult.
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.App.APIKeyReadOnly == "" && config.App.APIKeyAdmin == "" && config.App.APIKeyAgent == "" {
			v1.Error(c, http.StatusServiceUnavailable, "API authentication is not configured", nil)
			c.Abort()
			return
		}

		key := extractAPIKey(c)

		switch {
		case key != "" && config.App.APIKeyAdmin != "" && keysEqual(key, config.App.APIKeyAdmin):
			c.Set(roleContextKey, RoleAdmin)
		case key != "" && config.App.APIKeyReadOnly != "" && keysEqual(key, config.App.APIKeyReadOnly):
			c.Set(roleContextKey, RoleReadOnly)
		case key != "" && config.App.APIKeyAgent != "" && keysEqual(key, config.App.APIKeyAgent):
			c.Set(roleContextKey, RoleAgent)
		default:
			v1.Error(c, http.StatusUnauthorized, "Invalid or missing API key", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole aborts the request with 403 unless the authenticated role
// satisfies min. RoleAdmin always satisfies any requirement. Must run after
// APIKeyAuth.
func RequireRole(min Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(roleContextKey)
		have, _ := role.(Role)

		if have != RoleAdmin && have != min {
			v1.Error(c, http.StatusForbidden, "Insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	if key := c.GetHeader("X-API-Key"); key != "" {
		return key
	}

	const prefix = "Bearer "
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, prefix) {
		return strings.TrimPrefix(auth, prefix)
	}

	return ""
}

// keysEqual compares API keys in constant time to avoid leaking key
// contents through response-timing side channels.
func keysEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
