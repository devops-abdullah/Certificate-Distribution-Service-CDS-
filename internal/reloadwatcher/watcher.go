// Package reloadwatcher watches a directory tree of installed certificates
// (the layout the Certificate Agent writes: <root>/<domain>/*.pem) and
// triggers a reload command whenever something inside changes. It exists
// for deployments where the agent and Nginx run in separate containers
// sharing only a certificate volume: the agent has no way to signal a
// process in another container, so a small watcher living alongside Nginx
// (sharing its PID namespace) reacts to file changes instead.
package reloadwatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches root and its immediate subdirectories for changes.
type Watcher struct {
	root     string
	debounce time.Duration
}

// New creates a watcher for root.
func New(root string) *Watcher {
	return &Watcher{root: root, debounce: 500 * time.Millisecond}
}

// Start watches root (creating it if necessary) and every existing
// immediate subdirectory, adding a watch to any subdirectory created
// afterward (a newly onboarded domain). It calls onChange (debounced)
// whenever a file inside changes, and onError for watcher errors. Start
// blocks until stop is closed.
func (w *Watcher) Start(stop <-chan struct{}, onChange func(), onError func(error)) error {
	if err := os.MkdirAll(w.root, 0o755); err != nil {
		return fmt.Errorf("ensure watch root exists: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	if err := watcher.Add(w.root); err != nil {
		return fmt.Errorf("watch %s: %w", w.root, err)
	}

	entries, err := os.ReadDir(w.root)
	if err != nil {
		return fmt.Errorf("read watch root: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			_ = watcher.Add(filepath.Join(w.root, entry.Name()))
		}
	}

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

			// A newly created subdirectory (a newly onboarded domain)
			// needs its own watch added so changes inside it are seen.
			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = watcher.Add(event.Name)
				}
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
