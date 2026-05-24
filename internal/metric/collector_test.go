package metric_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/metric"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "web"},
			{Name: "worker"},
		},
	}
}

func TestNewCollector_NotNil(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	if c == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestRecord_UnknownApp(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	if err := c.Record("ghost", "cpu", 0.5); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRecord_And_Get(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	if err := c.Record("web", "cpu", 0.42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Record("web", "mem", 128); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, err := c.Get("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestGet_UnknownApp(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	_, err := c.Get("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestReset_ClearsEntries(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	_ = c.Record("web", "cpu", 1.0)
	_ = c.Record("web", "cpu", 2.0)
	if err := c.Reset("web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, _ := c.Get("web")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after reset, got %d", len(entries))
	}
}

func TestReset_UnknownApp(t *testing.T) {
	c := metric.NewCollector(baseConfig())
	if err := c.Reset("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}
