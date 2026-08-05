package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_FetchBundle_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "agent-secret" {
			t.Errorf("expected X-API-Key header to be set")
		}
		if r.URL.Path != "/api/v1/certificates/example.com/bundle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Certificate bundle",
			"data": map[string]string{
				"domain":      "example.com",
				"certificate": "CERT-PEM",
				"key":         "KEY-PEM",
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "agent-secret", nil)
	bundle, err := client.FetchBundle("example.com")
	if err != nil {
		t.Fatalf("FetchBundle returned error: %v", err)
	}

	if bundle.Domain != "example.com" || string(bundle.Certificate) != "CERT-PEM" || string(bundle.Key) != "KEY-PEM" {
		t.Errorf("unexpected bundle: %+v", bundle)
	}
}

func TestClient_FetchBundle_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Certificate bundle not found",
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "agent-secret", nil)
	_, err := client.FetchBundle("missing.example.com")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClient_FetchBundle_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid or missing API key",
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "wrong-key", nil)
	_, err := client.FetchBundle("example.com")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}
