package ratelimit_test

import (
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/ratelimit"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "web", Dir: "/app/web"},
		},
	}
}

func TestNewLimiter_NotNil(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	if l == nil {
		t.Fatal("expected non-nil Limiter")
	}
}

func TestAllow_UnknownApp(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	err := l.Allow("ghost", time.Minute)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestAllow_ZeroInterval_AlwaysPasses(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	if err := l.Record("web"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := l.Allow("web", 0); err != nil {
		t.Fatalf("expected nil with zero interval, got %v", err)
	}
}

func TestAllow_BlocksWithinInterval(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	if err := l.Record("web"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	err := l.Allow("web", 10*time.Minute)
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
}

func TestAllow_PassesAfterIntervalElapsed(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	if err := l.Record("web"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	// Use a very small interval that has definitely elapsed.
	time.Sleep(5 * time.Millisecond)
	if err := l.Allow("web", time.Millisecond); err != nil {
		t.Fatalf("expected nil after interval elapsed, got %v", err)
	}
}

func TestRecord_UnknownApp(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	if err := l.Record("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestReset_ClearsRecord(t *testing.T) {
	l := ratelimit.NewLimiter(baseConfig())
	_ = l.Record("web")
	l.Reset("web")
	// After reset, a long interval should still pass since there's no prior record.
	if err := l.Allow("web", time.Hour); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}
