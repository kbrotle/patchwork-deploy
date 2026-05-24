package gate_test

import (
	"errors"
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/gate"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "web"},
			{Name: "worker"},
		},
	}
}

func TestIsOpen_DefaultsToTrue_WhenNoRecord(t *testing.T) {
	g := gate.NewGatekeeper(baseConfig(), gate.NewMockStore())
	ok, err := g.IsOpen("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected gate to be open by default")
	}
}

func TestHold_ClosesGate(t *testing.T) {
	g := gate.NewGatekeeper(baseConfig(), gate.NewMockStore())
	if err := g.Hold("web"); err != nil {
		t.Fatalf("Hold: %v", err)
	}
	ok, err := g.IsOpen("web")
	if err != nil {
		t.Fatalf("IsOpen: %v", err)
	}
	if ok {
		t.Fatal("expected gate to be closed after Hold")
	}
}

func TestRelease_OpensGate(t *testing.T) {
	g := gate.NewGatekeeper(baseConfig(), gate.NewMockStore())
	_ = g.Hold("web")
	if err := g.Release("web"); err != nil {
		t.Fatalf("Release: %v", err)
	}
	ok, err := g.IsOpen("web")
	if err != nil {
		t.Fatalf("IsOpen: %v", err)
	}
	if !ok {
		t.Fatal("expected gate to be open after Release")
	}
}

func TestIsOpen_UnknownApp_ReturnsError(t *testing.T) {
	g := gate.NewGatekeeper(baseConfig(), gate.NewMockStore())
	_, err := g.IsOpen("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestHold_UnknownApp_ReturnsError(t *testing.T) {
	g := gate.NewGatekeeper(baseConfig(), gate.NewMockStore())
	if err := g.Hold("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestSaveFn_Error_PropagatesOnHold(t *testing.T) {
	ms := gate.NewMockStore()
	ms.SaveFn = func(_ string, _ bool) error { return errors.New("disk full") }
	g := gate.NewGatekeeper(baseConfig(), ms)
	if err := g.Hold("web"); err == nil {
		t.Fatal("expected store error to propagate")
	}
}
