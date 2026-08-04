package acme

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_DetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acme.json")

	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	w := NewWatcher(path)
	w.debounce = 50 * time.Millisecond

	stop := make(chan struct{})
	defer close(stop)

	changed := make(chan struct{}, 1)
	errs := make(chan error, 1)

	go func() {
		if err := w.Start(stop, func() {
			select {
			case changed <- struct{}{}:
			default:
			}
		}, nil); err != nil {
			errs <- err
		}
	}()

	// Give the watcher goroutine time to start and register the directory
	// before we trigger the change it's supposed to observe.
	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(path, []byte(`{"letsencrypt":{}}`), 0o600); err != nil {
		t.Fatalf("write updated file: %v", err)
	}

	select {
	case <-changed:
	case err := <-errs:
		t.Fatalf("watcher failed to start: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("expected onChange to be called after file write")
	}
}

func TestWatcher_StopsCleanly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acme.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	w := NewWatcher(path)
	stop := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- w.Start(stop, func() {}, nil)
	}()

	time.Sleep(100 * time.Millisecond)
	close(stop)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected Start to return nil after stop, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("expected Start to return after stop was closed")
	}
}
