package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newBundleServer(t *testing.T, cert string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]string{
				"domain":      "example.com",
				"certificate": cert,
				"key":         "KEY-" + cert,
			},
		})
	}))
}

func TestPoll_InstallsAndReloadsOnChange(t *testing.T) {
	srv := newBundleServer(t, "CERT-1")
	defer srv.Close()

	installDir := t.TempDir()
	reloadMarker := filepath.Join(t.TempDir(), "reloaded")

	cfg := Config{
		Domains:        []string{"example.com"},
		NginxReloadCmd: "touch " + reloadMarker,
	}
	client := NewClient(srv.URL, "agent-secret", nil)
	installer := NewInstaller(installDir)

	var events []string
	onLog := func(event string, fields map[string]interface{}) { events = append(events, event) }

	poll(cfg, client, installer, onLog)

	if _, err := os.Stat(reloadMarker); err != nil {
		t.Errorf("expected nginx reload to run on first install: %v", err)
	}

	certData, err := os.ReadFile(filepath.Join(installDir, "example.com", "fullchain.pem"))
	if err != nil || string(certData) != "CERT-1" {
		t.Fatalf("unexpected installed cert: %q, err=%v", certData, err)
	}

	if events[len(events)-1] != "reloaded" {
		t.Errorf("expected last event to be \"reloaded\", got %v", events)
	}
}

func TestPoll_SkipsReloadWhenUnchanged(t *testing.T) {
	srv := newBundleServer(t, "CERT-1")
	defer srv.Close()

	installDir := t.TempDir()
	reloadMarker := filepath.Join(t.TempDir(), "reloaded")

	cfg := Config{
		Domains:        []string{"example.com"},
		NginxReloadCmd: "touch " + reloadMarker,
	}
	client := NewClient(srv.URL, "agent-secret", nil)
	installer := NewInstaller(installDir)

	var events []string
	onLog := func(event string, fields map[string]interface{}) { events = append(events, event) }

	poll(cfg, client, installer, onLog) // first: installs + reloads
	if err := os.Remove(reloadMarker); err != nil {
		t.Fatalf("remove marker: %v", err)
	}

	poll(cfg, client, installer, onLog) // second: identical content, should skip reload

	if _, err := os.Stat(reloadMarker); err == nil {
		t.Error("expected no reload on an unchanged certificate")
	}
	if events[len(events)-1] != "unchanged" {
		t.Errorf("expected last event to be \"unchanged\", got %v", events)
	}
}

func TestRun_StopsCleanly(t *testing.T) {
	srv := newBundleServer(t, "CERT-1")
	defer srv.Close()

	cfg := Config{
		Domains:      []string{"example.com"},
		PollInterval: 10 * time.Millisecond,
	}
	client := NewClient(srv.URL, "agent-secret", nil)
	installer := NewInstaller(t.TempDir())

	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		Run(cfg, client, installer, func(string, map[string]interface{}) {}, stop)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	close(stop)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("expected Run to return after stop was closed")
	}
}
