package acme

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches an ACME store file for changes, debouncing rapid
// successive writes before notifying a caller.
type Watcher struct {
	path     string
	debounce time.Duration
}

// NewWatcher creates a watcher for the ACME store file at path.
func NewWatcher(path string) *Watcher {
	return &Watcher{path: path, debounce: 500 * time.Millisecond}
}

// Start watches the ACME file's parent directory (so atomic
// write-then-rename updates, as used by Traefik and most ACME clients, are
// still detected) and calls onChange whenever the file is created, written,
// or renamed into place. onError receives any watcher errors encountered
// along the way. Start blocks until stop is closed or the watcher fails to
// initialize.
func (w *Watcher) Start(stop <-chan struct{}, onChange func(), onError func(error)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	dir := filepath.Dir(w.path)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watch %s: %w", dir, err)
	}

	target := filepath.Base(w.path)

	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-stop:
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Base(event.Name) != target {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, onChange)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			if onError != nil {
				onError(err)
			}
		}
	}
}
