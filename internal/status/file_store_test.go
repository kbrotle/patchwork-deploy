package status_test

import (
	"os"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/status"
)

func tempStore(t *testing.T) *status.FileStore {
	t.Helper()
	dir := t.TempDir()
	store, err := status.NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return store
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	store := tempStore(t)
	s := status.AppStatus{App: "myapp", Host: "web", Deployed: true, Version: "v2.0.0"}
	if err := store.Save("myapp", s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load("myapp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Version != "v2.0.0" {
		t.Errorf("version: got %q, want %q", got.Version, "v2.0.0")
	}
}

func TestFileStore_Load_Missing(t *testing.T) {
	store := tempStore(t)
	if _, err := store.Load("nope"); err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestFileStore_Save_Overwrites(t *testing.T) {
	store := tempStore(t)
	s1 := status.AppStatus{App: "myapp", Version: "v1"}
	s2 := status.AppStatus{App: "myapp", Version: "v2"}
	_ = store.Save("myapp", s1)
	_ = store.Save("myapp", s2)
	got, err := store.Load("myapp")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Version != "v2" {
		t.Errorf("expected v2, got %q", got.Version)
	}
}

func TestNewFileStore_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/nested/status"
	if _, err := status.NewFileStore(dir); err != nil {
		t.Fatalf("NewFileStore should create dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("dir not created: %v", err)
	}
}
