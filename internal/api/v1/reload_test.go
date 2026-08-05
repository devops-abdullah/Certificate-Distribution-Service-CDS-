package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func postReload(r http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/reload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestReload_NotConfigured(t *testing.T) {
	SetReloadFunc(nil)

	r := gin.New()
	r.POST("/reload", Reload)

	w := postReload(r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestReload_CallsWiredFunc(t *testing.T) {
	called := false
	SetReloadFunc(func() { called = true })
	defer SetReloadFunc(nil)

	r := gin.New()
	r.POST("/reload", Reload)

	w := postReload(r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Error("expected reload function to be called")
	}
}
