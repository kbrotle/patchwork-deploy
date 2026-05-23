package rollback_test

import (
	"os"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/rollback"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "api", Dir: "/srv/api"},
		},
	}
}

func TestRecord_And_Latest(t *testing.T) {
	dir := t.TempDir()
	m := rollback.NewManager(rollback.NewFileStore(dir))
	cfg := baseConfig()

	if err := m.Record("api", "prod-1", "abc123", cfg); err != nil {
		t.Fatalf("Record: %v", err)
	}

	snap, err := m.Latest("api", "prod-1")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if snap.Revision != "abc123" {
		t.Errorf("expected revision abc123, got %s", snap.Revision)
	}
	if snap.App != "api" {
		t.Errorf("expected app api, got %s", snap.App)
	}
}

func TestRecord_UnknownApp(t *testing.T) {
	dir := t.TempDir()
	m := rollback.NewManager(rollback.NewFileStore(dir))
	err := m.Record("unknown", "prod-1", "abc123", baseConfig())
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestLatest_NoSnapshot(t *testing.T) {
	dir := t.TempDir()
	m := rollback.NewManager(rollback.NewFileStore(dir))
	_, err := m.Latest("api", "prod-1")
	if err == nil {
		t.Fatal("expected error when no snapshot exists")
	}
}

func TestRecord_OverwritesPrevious(t *testing.T) {
	dir := t.TempDir()
	m := rollback.NewManager(rollback.NewFileStore(dir))
	cfg := baseConfig()

	_ = m.Record("api", "prod-1", "first", cfg)
	_ = m.Record("api", "prod-1", "second", cfg)

	snap, err := m.Latest("api", "prod-1")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if snap.Revision != "second" {
		t.Errorf("expected second, got %s", snap.Revision)
	}
}

func TestFileStore_BadDir(t *testing.T) {
	// Use a file as the base dir to force a mkdir failure.
	tmp, _ := os.CreateTemp("", "notadir")
	tmp.Close()
	defer os.Remove(tmp.Name())

	fs := rollback.NewFileStore(tmp.Name() + "/subdir")
	snap := rollback.Snapshot{App: "api", Host: "h", Revision: "r"}
	if err := fs.Save(snap); err == nil {
		t.Fatal("expected error saving to bad dir")
	}
}
