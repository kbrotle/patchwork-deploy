package schedule_test

import (
	"os"
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/schedule"
)

func tempStore(t *testing.T) *schedule.FileStore {
	t.Helper()
	dir, err := os.MkdirTemp("", "schedule-store-*")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	fs, err := schedule.NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return fs
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	fs := tempStore(t)
	rec := schedule.ScheduleRecord{
		AppName:  "web",
		Interval: 5 * time.Minute,
		Created:  time.Now().UTC().Truncate(time.Second),
	}
	if err := fs.Save(rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := fs.Load("web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AppName != rec.AppName || got.Interval != rec.Interval {
		t.Fatalf("mismatch: got %+v, want %+v", got, rec)
	}
}

func TestFileStore_Load_Missing(t *testing.T) {
	fs := tempStore(t)
	_, err := fs.Load("ghost")
	if err == nil {
		t.Fatal("expected error for missing record")
	}
}

func TestFileStore_Save_Overwrites(t *testing.T) {
	fs := tempStore(t)
	rec := schedule.ScheduleRecord{AppName: "web", Interval: time.Minute, Created: time.Now()}
	_ = fs.Save(rec)
	rec.Interval = 10 * time.Minute
	_ = fs.Save(rec)
	got, err := fs.Load("web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Interval != 10*time.Minute {
		t.Fatalf("expected overwritten interval, got %v", got.Interval)
	}
}

func TestFileStore_Delete(t *testing.T) {
	fs := tempStore(t)
	rec := schedule.ScheduleRecord{AppName: "web", Interval: time.Minute, Created: time.Now()}
	_ = fs.Save(rec)
	if err := fs.Delete("web"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := fs.Load("web")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestFileStore_Delete_NonExistent_IsNoop(t *testing.T) {
	fs := tempStore(t)
	if err := fs.Delete("ghost"); err != nil {
		t.Fatalf("expected noop delete, got: %v", err)
	}
}
