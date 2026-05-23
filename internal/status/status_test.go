package status_test

import (
	"errors"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/status"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Address: "1.2.3.4"},
		},
		Apps: map[string]config.App{
			"myapp": {Host: "web", Dir: "/srv/myapp"},
		},
	}
}

func TestRecord_And_Get(t *testing.T) {
	store := status.NewMockStore()
	tracker := status.NewTracker(baseConfig(), store)

	if err := tracker.Record("myapp", "v1.2.3", "ok", true); err != nil {
		t.Fatalf("Record: %v", err)
	}
	s, err := tracker.Get("myapp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.Version != "v1.2.3" {
		t.Errorf("version: got %q, want %q", s.Version, "v1.2.3")
	}
	if !s.Deployed {
		t.Error("expected deployed=true")
	}
	if s.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestRecord_UnknownApp(t *testing.T) {
	tracker := status.NewTracker(baseConfig(), status.NewMockStore())
	if err := tracker.Record("ghost", "v1", "", true); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestGet_UnknownApp(t *testing.T) {
	tracker := status.NewTracker(baseConfig(), status.NewMockStore())
	if _, err := tracker.Get("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestGet_NoRecord(t *testing.T) {
	tracker := status.NewTracker(baseConfig(), status.NewMockStore())
	if _, err := tracker.Get("myapp"); err == nil {
		t.Fatal("expected error when no record exists")
	}
}

func TestRecord_SaveFnError(t *testing.T) {
	store := status.NewMockStore()
	store.SaveFn = func(_ string, _ status.AppStatus) error {
		return errors.New("disk full")
	}
	tracker := status.NewTracker(baseConfig(), store)
	if err := tracker.Record("myapp", "v1", "", true); err == nil {
		t.Fatal("expected error from SaveFn")
	}
}
