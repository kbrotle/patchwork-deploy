package snapshot_test

import (
	"os"
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/snapshot"
)

func tempStore(t *testing.T) *snapshot.FileStore {
	t.Helper()
	dir, err := os.MkdirTemp("", "snap-store-*")
	if err != nil {
		t.Fatalf("tempStore: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	store, err := snapshot.NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return store
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	store := tempStore(t)
	snap := &snapshot.Snapshot{
		App:        "web",
		Timestamp:  time.Now().UTC().Truncate(time.Second),
		RemotePath: "/srv/web/.snapshots/20240101T120000Z",
	}
	if err := store.Save("web", snap); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load("web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if got.RemotePath != snap.RemotePath {
		t.Errorf("RemotePath: got %q, want %q", got.RemotePath, snap.RemotePath)
	}
}

func TestFileStore_Load_Missing(t *testing.T) {
	store := tempStore(t)
	got, err := store.Load("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing app, got %+v", got)
	}
}

func TestFileStore_Save_Overwrites(t *testing.T) {
	store := tempStore(t)
	first := &snapshot.Snapshot{App: "api", RemotePath: "/srv/api/.snapshots/first"}
	second := &snapshot.Snapshot{App: "api", RemotePath: "/srv/api/.snapshots/second"}

	_ = store.Save("api", first)
	_ = store.Save("api", second)

	got, _ := store.Load("api")
	if got.RemotePath != second.RemotePath {
		t.Errorf("expected second snapshot, got %q", got.RemotePath)
	}
}

func TestMockStore_SaveAndLoad(t *testing.T) {
	store := snapshot.NewMockStore()
	snap := &snapshot.Snapshot{App: "worker", RemotePath: "/srv/worker/snap"}
	_ = store.Save("worker", snap)
	got := store.MustLoad("worker")
	if got.RemotePath != snap.RemotePath {
		t.Errorf("got %q, want %q", got.RemotePath, snap.RemotePath)
	}
}

func TestMockStore_SaveFn_Error(t *testing.T) {
	store := snapshot.NewMockStore()
	store.SaveFn = func(_ string, _ *snapshot.Snapshot) error {
		return fmt.Errorf("disk full")
	}
	err := store.Save("web", &snapshot.Snapshot{})
	if err == nil {
		t.Fatal("expected error from SaveFn")
	}
}
