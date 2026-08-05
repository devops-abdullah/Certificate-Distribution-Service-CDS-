package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/devops-abdullah/cds/pkg/logger"
)

func TestAuditLog_RecordsRequestDetails(t *testing.T) {
	var buf bytes.Buffer
	originalOut, originalFormatter := logger.Log.Out, logger.Log.Formatter
	logger.Log.SetOutput(&buf)
	logger.Log.SetFormatter(&logrus.JSONFormatter{})
	t.Cleanup(func() {
		logger.Log.SetOutput(originalOut)
		logger.Log.SetFormatter(originalFormatter)
	})

	r := gin.New()
	r.GET("/thing", AuditLog(), func(c *gin.Context) {
		c.Status(http.StatusTeapot)
	})

	req := httptest.NewRequest(http.MethodGet, "/thing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("decode audit log line: %v (raw: %s)", err, buf.String())
	}

	if entry["audit"] != true {
		t.Errorf("expected audit=true, got %v", entry["audit"])
	}
	if entry["method"] != http.MethodGet {
		t.Errorf("expected method=GET, got %v", entry["method"])
	}
	if entry["path"] != "/thing" {
		t.Errorf("expected path=/thing, got %v", entry["path"])
	}
	if entry["status"] != float64(http.StatusTeapot) {
		t.Errorf("expected status=418, got %v", entry["status"])
	}
}
