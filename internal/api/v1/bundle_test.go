package v1

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/config"
)

func withExportDir(t *testing.T, dir string) {
	t.Helper()
	original := config.App.ExportDir
	config.App.ExportDir = dir
	t.Cleanup(func() { config.App.ExportDir = original })
}

func newBundleTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/certificates/:domain/bundle", GetCertificateBundle)
	return r
}

func TestGetCertificateBundle_Found(t *testing.T) {
	dir := t.TempDir()
	withExportDir(t, dir)

	domainDir := filepath.Join(dir, "example.com")
	if err := os.MkdirAll(domainDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(domainDir, "fullchain.pem"), []byte("CERT-DATA"), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(domainDir, "privkey.pem"), []byte("KEY-DATA"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	w := performRequest(newBundleTestRouter(), "/certificates/example.com/bundle")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	resp := decodeResponse(t, w)
	data, ok := resp.Data.(map[string]interface{})
	if !ok || data["certificate"] != "CERT-DATA" || data["key"] != "KEY-DATA" {
		t.Fatalf("unexpected bundle response: %v", resp.Data)
	}
}

func TestGetCertificateBundle_NotFound(t *testing.T) {
	withExportDir(t, t.TempDir())

	w := performRequest(newBundleTestRouter(), "/certificates/missing.example.com/bundle")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetCertificateBundle_PathTraversalRejected(t *testing.T) {
	withExportDir(t, t.TempDir())

	w := performRequest(newBundleTestRouter(), "/certificates/../bundle")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d for a \"..\" domain, got %d", http.StatusBadRequest, w.Code)
	}
}
