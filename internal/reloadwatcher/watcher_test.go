package reloadwatcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_DetectsChangeInExistingSubdir(t *testing.T) {
	root := t.TempDir()
	domainDir := filepath.Join(root, "example.com")
	if err := os.MkdirAll(domainDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	certPath := filepath.Join(domainDir, "fullchain.pem")
	if err := os.WriteFile(certPath, []byte("v1"), 0o644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	w := New(root)
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

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(certPath, []byte("v2"), 0o644); err != nil {
		t.Fatalf("write updated file: %v", err)
	}

	select {
	case <-changed:
	case err := <-errs:
		t.Fatalf("watcher failed to start: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("expected onChange after updating a file in an existing subdirectory")
	}
}

func TestWatcher_DetectsNewSubdirectory(t *testing.T) {
	root := t.TempDir()

	w := New(root)
	w.debounce = 50 * time.Millisecond

	stop := make(chan struct{})
	defer close(stop)

	changed := make(chan struct{}, 1)

	go func() {
		_ = w.Start(stop, func() {
			select {
			case changed <- struct{}{}:
			default:
			}
		}, nil)
	}()

	time.Sleep(200 * time.Millisecond)

	newDomainDir := filepath.Join(root, "new-domain.example.com")
	if err := os.MkdirAll(newDomainDir, 0o755); err != nil {
		t.Fatalf("mkdir new domain: %v", err)
	}

	// The directory creation itself should trigger onChange...
	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("expected onChange for the new subdirectory being created")
	}

	// ...and the watcher should now also be watching inside it.
	if err := os.WriteFile(filepath.Join(newDomainDir, "fullchain.pem"), []byte("v1"), 0o644); err != nil {
		t.Fatalf("write file in new subdirectory: %v", err)
	}

	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("expected onChange after writing into the newly created subdirectory")
	}
}

func TestWatcher_StopsCleanly(t *testing.T) {
	root := t.TempDir()

	w := New(root)
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
