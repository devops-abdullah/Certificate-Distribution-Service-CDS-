package v1

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	r := gin.New()
	r.GET("/health", Health)

	w := performRequest(r, "/health")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Fatalf("expected success=true, got %v", resp)
	}
}
