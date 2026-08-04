package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/storage"
)

var certificateStore *storage.Store

// SetCertificateStore wires the certificate metadata store used by the
// certificate inventory endpoints. Call once during startup.
func SetCertificateStore(store *storage.Store) {
	certificateStore = store
}

// ListCertificates returns metadata for every known certificate.
func ListCertificates(c *gin.Context) {
	if certificateStore == nil {
		Error(c, http.StatusServiceUnavailable, "Certificate store not initialized", nil)
		return
	}

	Success(c, "Certificate inventory", certificateStore.List())
}

// GetCertificate returns metadata for a single certificate by domain.
func GetCertificate(c *gin.Context) {
	if certificateStore == nil {
		Error(c, http.StatusServiceUnavailable, "Certificate store not initialized", nil)
		return
	}

	domain := c.Param("domain")

	cert, ok := certificateStore.Get(domain)
	if !ok {
		Error(c, http.StatusNotFound, "Certificate not found", nil)
		return
	}

	Success(c, "Certificate metadata", cert)
}
