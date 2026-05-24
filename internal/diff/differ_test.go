package diff_test

import (
	"strings"
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/diff"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "api", Host: "web1", Dir: "/srv/api"},
		},
		Hosts: []config.Host{
			{Name: "web1", Address: "10.0.0.1"},
		},
	}
}

func TestNewDiffer_NotNil(t *testing.T) {
	d := diff.NewDiffer(baseConfig())
	if d == nil {
		t.Fatal("expected non-nil Differ")
	}
}

func TestCompare_UnknownApp(t *testing.T) {
	d := diff.NewDiffer(baseConfig())
	_, err := d.Compare("ghost", nil, nil)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
	if !strings.Contains(err.Error(), "unknown app") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	d := diff.NewDiffer(baseConfig())
	prev := map[string]string{"image": "v1", "port": "8080"}
	next := map[string]string{"image": "v1", "port": "8080"}
	r, err := d.Compare("api", prev, next)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.HasChanges() {
		t.Errorf("expected no changes, got %d", len(r.Changes))
	}
}

func TestCompare_DetectsModified(t *testing.T) {
	d := diff.NewDiffer(baseConfig())
	prev := map[string]string{"image": "v1"}
	next := map[string]string{"image": "v2"}
	r, err := d.Compare("api", prev, next)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.HasChanges() {
		t.Fatal("expected changes")
	}
	if r.Changes[0].Old != "v1" || r.Changes[0].New != "v2" {
		t.Errorf("unexpected change: %+v", r.Changes[0])
	}
}

func TestCompare_DetectsAddedAndRemoved(t *testing.T) {
	d := diff.NewDiffer(baseConfig())
	prev := map[string]string{"a": "1"}
	next := map[string]string{"b": "2"}
	r, err := d.Compare("api", prev, next)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(r.Changes))
	}
}

func TestSummary_NoChanges(t *testing.T) {
	r := &diff.Result{App: "api"}
	if !strings.Contains(r.Summary(), "no changes") {
		t.Errorf("unexpected summary: %s", r.Summary())
	}
}

func TestSummary_WithChanges(t *testing.T) {
	r := &diff.Result{
		App:     "api",
		Changes: []diff.Change{{Field: "image", Old: "v1", New: "v2"}},
	}
	if !strings.Contains(r.Summary(), "image") {
		t.Errorf("summary missing field name: %s", r.Summary())
	}
}
