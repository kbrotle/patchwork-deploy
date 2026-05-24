package version_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/version"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: map[string]config.App{
			"api": {Dir: "/srv/api"},
		},
	}
}

func TestNewTracker_NotNil(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if tr == nil {
		t.Fatal("expected non-nil Tracker")
	}
}

func TestRecord_UnknownApp(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if err := tr.Record("ghost", "v1.0.0", "abc123"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRecord_EmptyTag(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if err := tr.Record("api", "", "abc123"); err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestRecord_And_Current(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if err := tr.Record("api", "v2.3.1", "deadbeef"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, err := tr.Current("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Tag != "v2.3.1" {
		t.Errorf("expected tag v2.3.1, got %q", e.Tag)
	}
	if e.Commit != "deadbeef" {
		t.Errorf("expected commit deadbeef, got %q", e.Commit)
	}
	if e.DeployedAt.IsZero() {
		t.Error("expected non-zero DeployedAt")
	}
}

func TestCurrent_UnknownApp(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if _, err := tr.Current("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestCurrent_NoRecord(t *testing.T) {
	tr := version.NewTracker(baseConfig(), version.NewMockStore())
	if _, err := tr.Current("api"); err == nil {
		t.Fatal("expected error when no record exists")
	}
}

func TestRecord_SaveFnError(t *testing.T) {
	ms := version.NewMockStore()
	ms.SaveFn = func(_ string, _ version.Entry) error {
		return fmt.Errorf("disk full")
	}
	tr := version.NewTracker(baseConfig(), ms)
	if err := tr.Record("api", "v1.0.0", "abc"); err == nil {
		t.Fatal("expected error from SaveFn")
	}
}
