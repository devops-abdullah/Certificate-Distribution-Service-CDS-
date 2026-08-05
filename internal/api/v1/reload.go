package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var reloadFunc func()

// SetReloadFunc wires the function that re-reads the ACME store and
// refreshes the certificate inventory. Call once during startup.
func SetReloadFunc(f func()) {
	reloadFunc = f
}

// Reload triggers an immediate re-read of the ACME store instead of
// waiting for the file watcher to notice a change.
func Reload(c *gin.Context) {
	if reloadFunc == nil {
		Error(c, http.StatusServiceUnavailable, "Reload not available", nil)
		return
	}

	reloadFunc()
	Success(c, "Certificate inventory reload triggered", nil)
}
