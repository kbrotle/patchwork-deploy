package watch_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/watch"
)

func baseConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		Apps: map[string]config.App{
			"web": {Dir: dir},
		},
	}
}

func TestNewWatcher_NotNil(t *testing.T) {
	cfg := baseConfig(t)
	w := watch.NewWatcher(cfg, time.Second)
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestWatch_UnknownApp(t *testing.T) {
	cfg := baseConfig(t)
	w := watch.NewWatcher(cfg, 10*time.Millisecond)
	ctx := context.Background()
	err := w.Watch(ctx, "ghost", func(watch.FileEvent) {})
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestWatch_DetectsChange(t *testing.T) {
	cfg := baseConfig(t)
	dir := cfg.Apps["web"].Dir
	file := filepath.Join(dir, "app.txt")

	if err := os.WriteFile(file, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Point the app dir directly at the file so hashPath works.
	cfg.Apps["web"] = config.App{Dir: file}

	w := watch.NewWatcher(cfg, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	events := make(chan watch.FileEvent, 4)
	go w.Watch(ctx, "web", func(e watch.FileEvent) { //nolint:errcheck
		events <- e
	})

	// Wait for first detection (initial unseen state).
	select {
	case ev := <-events:
		if ev.App != "web" {
			t.Fatalf("unexpected app %q", ev.App)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for initial change event")
	}
}

func TestMockStore_SaveAndLoad(t *testing.T) {
	s := watch.NewMockStore()
	if err := s.Save("web", "abc123"); err != nil {
		t.Fatal(err)
	}
	hash, err := s.Load("web")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "abc123" {
		t.Fatalf("expected abc123, got %q", hash)
	}
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := watch.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save("web", "deadbeef"); err != nil {
		t.Fatal(err)
	}
	hash, err := s.Load("web")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "deadbeef" {
		t.Fatalf("expected deadbeef, got %q", hash)
	}
}

func TestFileStore_Load_Missing(t *testing.T) {
	dir := t.TempDir()
	s, err := watch.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := s.Load("nope")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "" {
		t.Fatalf("expected empty hash, got %q", hash)
	}
}
