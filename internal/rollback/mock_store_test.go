package rollback_test

import (
	"errors"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/rollback"
)

func TestMockStore_SaveAndLoad(t *testing.T) {
	ms := rollback.NewMockStore()
	snap := rollback.Snapshot{App: "web", Host: "srv1", Revision: "deadbeef"}

	if err := ms.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := ms.Load("web", "srv1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Revision != "deadbeef" {
		t.Errorf("expected deadbeef, got %s", got.Revision)
	}
}

func TestMockStore_LoadMissing(t *testing.T) {
	ms := rollback.NewMockStore()
	_, err := ms.Load("nope", "srv1")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestMockStore_SaveFn_Error(t *testing.T) {
	ms := rollback.NewMockStore()
	ms.SaveFn = func(_ rollback.Snapshot) error {
		return errors.New("injected save error")
	}
	err := ms.Save(rollback.Snapshot{App: "x", Host: "h"})
	if err == nil {
		t.Fatal("expected injected error")
	}
}

func TestManager_WithMockStore(t *testing.T) {
	ms := rollback.NewMockStore()
	m := rollback.NewManager(ms)
	cfg := baseConfig()

	if err := m.Record("api", "prod-1", "rev42", cfg); err != nil {
		t.Fatalf("Record: %v", err)
	}
	snap, err := m.Latest("api", "prod-1")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if snap.Revision != "rev42" {
		t.Errorf("expected rev42, got %s", snap.Revision)
	}
}
