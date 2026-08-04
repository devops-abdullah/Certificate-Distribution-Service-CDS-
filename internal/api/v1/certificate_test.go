package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/devops-abdullah/cds/internal/certs"
	"github.com/devops-abdullah/cds/internal/storage"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/certificates", ListCertificates)
	r.GET("/certificates/:domain", GetCertificate)
	return r
}

func performRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) APIResponse {
	t.Helper()

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestListCertificates_StoreNotInitialized(t *testing.T) {
	SetCertificateStore(nil)

	w := performRequest(newTestRouter(), "/certificates")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestListCertificates_Populated(t *testing.T) {
	store := storage.New()
	store.Replace([]certs.Metadata{{Domain: "example.com"}})
	SetCertificateStore(store)
	defer SetCertificateStore(nil)

	w := performRequest(newTestRouter(), "/certificates")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}

	resp := decodeResponse(t, w)
	list, ok := resp.Data.([]interface{})
	if !ok || len(list) != 1 {
		t.Fatalf("expected 1 certificate in response, got %v", resp.Data)
	}
}

func TestGetCertificate_Found(t *testing.T) {
	store := storage.New()
	store.Replace([]certs.Metadata{{Domain: "example.com", Issuer: "Test CA"}})
	SetCertificateStore(store)
	defer SetCertificateStore(nil)

	w := performRequest(newTestRouter(), "/certificates/example.com")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}

	resp := decodeResponse(t, w)
	data, ok := resp.Data.(map[string]interface{})
	if !ok || data["domain"] != "example.com" {
		t.Fatalf("expected domain example.com in response, got %v", resp.Data)
	}
}

func TestGetCertificate_NotFound(t *testing.T) {
	SetCertificateStore(storage.New())
	defer SetCertificateStore(nil)

	w := performRequest(newTestRouter(), "/certificates/missing.example.com")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetCertificate_StoreNotInitialized(t *testing.T) {
	SetCertificateStore(nil)

	w := performRequest(newTestRouter(), "/certificates/example.com")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}
