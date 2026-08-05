package v1

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/config"
)

// CertificateBundle is the response body for GetCertificateBundle. Unlike
// Metadata, this includes private key material — it is only ever served to
// authenticated agent/admin callers, never to readonly ones.
type CertificateBundle struct {
	Domain      string `json:"domain"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

// GetCertificateBundle serves the exported certificate + private key for a
// domain, read from EXPORT_DIR (populated by the exporter on every load/
// reload). Returns 404 if nothing has been exported for that domain yet
// (e.g. its key couldn't be decoded, or the domain doesn't exist at all).
func GetCertificateBundle(c *gin.Context) {
	domain := c.Param("domain")

	if domain == "" || domain == "." || domain == ".." || strings.ContainsAny(domain, `/\`) {
		Error(c, http.StatusBadRequest, "Invalid domain", nil)
		return
	}

	dir := filepath.Join(config.App.ExportDir, domain)

	cert, err := os.ReadFile(filepath.Join(dir, "fullchain.pem"))
	if err != nil {
		Error(c, http.StatusNotFound, "Certificate bundle not found", nil)
		return
	}

	key, err := os.ReadFile(filepath.Join(dir, "privkey.pem"))
	if err != nil {
		Error(c, http.StatusNotFound, "Certificate bundle not found", nil)
		return
	}

	Success(c, "Certificate bundle", CertificateBundle{
		Domain:      domain,
		Certificate: string(cert),
		Key:         string(key),
	})
}
