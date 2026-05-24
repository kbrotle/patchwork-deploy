package schedule_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/schedule"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: map[string]config.App{
			"web": {Dir: "/app/web"},
		},
	}
}

func TestNewScheduler_NotNil(t *testing.T) {
	s := schedule.NewScheduler(baseConfig(), func(string) error { return nil })
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestRegister_UnknownApp(t *testing.T) {
	s := schedule.NewScheduler(baseConfig(), func(string) error { return nil })
	err := s.Register("ghost", time.Second)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRegister_InvalidInterval(t *testing.T) {
	s := schedule.NewScheduler(baseConfig(), func(string) error { return nil })
	err := s.Register("web", 0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestRegister_FiresDeploy(t *testing.T) {
	var calls int32
	s := schedule.NewScheduler(baseConfig(), func(app string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	if err := s.Register("web", 20*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s.StopAll()

	time.Sleep(70 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Fatalf("expected at least 2 deploy calls, got %d", got)
	}
}

func TestUnregister_StopsFiring(t *testing.T) {
	var calls int32
	s := schedule.NewScheduler(baseConfig(), func(string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	_ = s.Register("web", 20*time.Millisecond)
	time.Sleep(35 * time.Millisecond)
	s.Unregister("web")
	snap := atomic.LoadInt32(&calls)
	time.Sleep(40 * time.Millisecond)
	if after := atomic.LoadInt32(&calls); after != snap {
		t.Fatalf("deploy fired after unregister: before=%d after=%d", snap, after)
	}
}

func TestRegister_ReplacesExisting(t *testing.T) {
	s := schedule.NewScheduler(baseConfig(), func(string) error { return nil })
	_ = s.Register("web", 50*time.Millisecond)
	err := s.Register("web", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("re-register should succeed, got: %v", err)
	}
	s.StopAll()
}
