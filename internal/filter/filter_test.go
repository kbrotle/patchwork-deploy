package filter_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/filter"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web-host": {Addr: "1.2.3.4"},
			"db-host":  {Addr: "5.6.7.8"},
		},
		Apps: map[string]config.App{
			"api":     {Host: "web-host", Dir: "/app/api", Tags: []string{"backend", "prod"}},
			"web":     {Host: "web-host", Dir: "/app/web", Tags: []string{"frontend", "prod"}},
			"migrate": {Host: "db-host", Dir: "/app/migrate", Tags: []string{"backend"}},
		},
	}
}

func TestNewFilter_NotNil(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	if f == nil {
		t.Fatal("expected non-nil Filter")
	}
}

func TestApply_EmptyCriteria_ReturnsAll(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	apps, err := f.Apply(filter.Criteria{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 3 {
		t.Fatalf("expected 3 apps, got %d", len(apps))
	}
}

func TestApply_ByAppName_ExplicitList(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	apps, err := f.Apply(filter.Criteria{Apps: []string{"api", "web"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
}

func TestApply_UnknownApp_ReturnsError(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	_, err := f.Apply(filter.Criteria{Apps: []string{"ghost"}})
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestApply_ByHost_FiltersCorrectly(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	apps, err := f.Apply(filter.Criteria{Hosts: []string{"db-host"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 || apps[0] != "migrate" {
		t.Fatalf("expected [migrate], got %v", apps)
	}
}

func TestApply_ByTag_FiltersCorrectly(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	apps, err := f.Apply(filter.Criteria{Tags: []string{"frontend"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 || apps[0] != "web" {
		t.Fatalf("expected [web], got %v", apps)
	}
}

func TestApply_ByTag_CaseInsensitive(t *testing.T) {
	f := filter.NewFilter(baseConfig())
	apps, err := f.Apply(filter.Criteria{Tags: []string{"PROD"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 prod apps, got %d", len(apps))
	}
}
