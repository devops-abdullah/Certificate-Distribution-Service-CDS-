package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setKeys(t *testing.T, readOnly, admin string) {
	t.Helper()

	original := config.App
	config.App.APIKeyReadOnly = readOnly
	config.App.APIKeyAdmin = admin

	t.Cleanup(func() { config.App = original })
}

func newAuthTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/protected", APIKeyAuth(), RequireRole(RoleReadOnly), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/admin-only", APIKeyAuth(), RequireRole(RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func doRequest(r http.Handler, path, apiKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPIKeyAuth_NoKeysConfigured(t *testing.T) {
	setKeys(t, "", "")

	w := doRequest(newAuthTestRouter(), "/protected", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	setKeys(t, "read-secret", "admin-secret")

	w := doRequest(newAuthTestRouter(), "/protected", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIKeyAuth_WrongKey(t *testing.T) {
	setKeys(t, "read-secret", "admin-secret")

	w := doRequest(newAuthTestRouter(), "/protected", "not-the-right-key")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAPIKeyAuth_ReadOnlyKey_AccessesReadOnlyRoute(t *testing.T) {
	setKeys(t, "read-secret", "admin-secret")

	w := doRequest(newAuthTestRouter(), "/protected", "read-secret")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAPIKeyAuth_ReadOnlyKey_DeniedFromAdminRoute(t *testing.T) {
	setKeys(t, "read-secret", "admin-secret")

	w := doRequest(newAuthTestRouter(), "/admin-only", "read-secret")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestAPIKeyAuth_AdminKey_AccessesBothRoutes(t *testing.T) {
	setKeys(t, "read-secret", "admin-secret")

	router := newAuthTestRouter()

	if w := doRequest(router, "/protected", "admin-secret"); w.Code != http.StatusOK {
		t.Fatalf("admin on read-only route: expected %d, got %d", http.StatusOK, w.Code)
	}

	if w := doRequest(router, "/admin-only", "admin-secret"); w.Code != http.StatusOK {
		t.Fatalf("admin on admin route: expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAPIKeyAuth_BearerToken(t *testing.T) {
	setKeys(t, "read-secret", "")

	router := newAuthTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer read-secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
