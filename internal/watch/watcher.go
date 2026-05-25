package watch

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// FileEvent describes a change detected on a watched file.
type FileEvent struct {
	App  string
	Path string
	Hash string
}

// Handler is called when a file change is detected.
type Handler func(FileEvent)

// Watcher polls local app directories for file changes.
type Watcher struct {
	cfg      *config.Config
	interval time.Duration
	mu       sync.Mutex
	hashes   map[string]string // path -> last sha256
}

// NewWatcher returns a Watcher that checks for changes at the given interval.
func NewWatcher(cfg *config.Config, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Watcher{
		cfg:      cfg,
		interval: interval,
		hashes:   make(map[string]string),
	}
}

// Watch starts polling and calls h whenever a file change is detected.
// It blocks until ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context, appName string, h Handler) error {
	app, ok := w.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("watch: unknown app %q", appName)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if ev, changed := w.checkFile(appName, app.Dir); changed {
				h(ev)
			}
		}
	}
}

func (w *Watcher) checkFile(appName, path string) (FileEvent, bool) {
	hash, err := hashPath(path)
	if err != nil {
		return FileEvent{}, false
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	prev, seen := w.hashes[path]
	w.hashes[path] = hash

	if !seen || prev != hash {
		return FileEvent{App: appName, Path: path, Hash: hash}, true
	}
	return FileEvent{}, false
}

func hashPath(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
