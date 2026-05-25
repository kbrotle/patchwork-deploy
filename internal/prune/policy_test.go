package prune

import (
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{
				Name: "web",
				Prune: config.PruneConfig{
					KeepLast:   3,
					MaxAgeDays: 7,
				},
			},
			{
				Name:  "worker",
				Prune: config.PruneConfig{},
			},
		},
	}
}

func TestNewManager_NotNil(t *testing.T) {
	m := NewManager(baseConfig())
	if m == nil {
		t.Fatal("expected non-nil Manager")
	}
}

func TestPolicyFor_UnknownApp(t *testing.T) {
	m := NewManager(baseConfig())
	_, err := m.PolicyFor("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestPolicyFor_DefaultKeepLast(t *testing.T) {
	m := NewManager(baseConfig())
	p, err := m.PolicyFor("worker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.KeepLast != 5 {
		t.Errorf("expected default KeepLast=5, got %d", p.KeepLast)
	}
}

func TestApply_NoEntries(t *testing.T) {
	m := NewManager(baseConfig())
	p, _ := m.PolicyFor("web")
	result := m.Apply(p, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestApply_RemovesOldEntries(t *testing.T) {
	m := NewManager(baseConfig())
	p, _ := m.PolicyFor("web") // MaxAgeDays=7, KeepLast=3

	now := time.Now()
	entries := []Entry{
		{ID: "a", CreatedAt: now.AddDate(0, 0, -10)}, // too old
		{ID: "b", CreatedAt: now.AddDate(0, 0, -8)},  // too old
		{ID: "c", CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "d", CreatedAt: now.AddDate(0, 0, -2)},
		{ID: "e", CreatedAt: now.AddDate(0, 0, -1)},
	}

	remove := m.Apply(p, entries)
	ids := make(map[string]bool)
	for _, e := range remove {
		ids[e.ID] = true
	}

	if !ids["a"] || !ids["b"] {
		t.Error("expected old entries a and b to be pruned")
	}
	if ids["c"] || ids["d"] || ids["e"] {
		t.Error("expected recent entries c, d, e to be kept")
	}
}

func TestApply_EnforcesKeepLast(t *testing.T) {
	m := NewManager(baseConfig())
	p, _ := m.PolicyFor("web") // KeepLast=3, MaxAgeDays=7

	now := time.Now()
	entries := []Entry{
		{ID: "a", CreatedAt: now.AddDate(0, 0, -6)},
		{ID: "b", CreatedAt: now.AddDate(0, 0, -5)},
		{ID: "c", CreatedAt: now.AddDate(0, 0, -4)},
		{ID: "d", CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "e", CreatedAt: now.AddDate(0, 0, -1)},
	}

	remove := m.Apply(p, entries)
	if len(remove) != 2 {
		t.Errorf("expected 2 entries removed, got %d", len(remove))
	}
}
