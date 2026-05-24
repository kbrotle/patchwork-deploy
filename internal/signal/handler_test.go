package signal_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/signal"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Addr: "127.0.0.1", User: "deploy"},
		},
		Apps: map[string]config.App{
			"api": {Host: "web", Dir: "/app"},
		},
	}
}

func TestNewHandler_NotNil(t *testing.T) {
	h := signal.NewHandler(baseConfig())
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestWait_ContextCancelled_ReturnsError(t *testing.T) {
	h := signal.NewHandler(baseConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := h.Wait(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestWait_RunsHooksOnSignal(t *testing.T) {
	called := false
	hook := func(ctx context.Context) error {
		called = true
		return nil
	}

	h := signal.NewHandler(baseConfig(), hook)
	h.WithSignals(syscall.SIGUSR1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		h.Notify(syscall.SIGUSR1)
	}()

	_ = h.Wait(ctx)
	if !called {
		t.Fatal("expected shutdown hook to be called")
	}
}

func TestWait_ReturnsFirstHookError(t *testing.T) {
	expected := errors.New("hook failed")
	hook := func(ctx context.Context) error {
		return expected
	}

	h := signal.NewHandler(baseConfig(), hook)
	h.WithSignals(syscall.SIGUSR1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		h.Notify(syscall.SIGUSR1)
	}()

	err := h.Wait(ctx)
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestWithSignals_Chainable(t *testing.T) {
	h := signal.NewHandler(baseConfig()).WithSignals(syscall.SIGTERM)
	if h == nil {
		t.Fatal("expected non-nil handler after WithSignals")
	}
}
