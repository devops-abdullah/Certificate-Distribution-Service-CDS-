package v1

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVersionHandler(t *testing.T) {
	r := gin.New()
	r.GET("/version", VersionHandler)

	w := performRequest(r, "/version")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}

	resp := decodeResponse(t, w)
	data, ok := resp.Data.(map[string]interface{})
	if !ok || data["version"] != Version {
		t.Fatalf("expected version %q in response, got %v", Version, resp.Data)
	}
}
