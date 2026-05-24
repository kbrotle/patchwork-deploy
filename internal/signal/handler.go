package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/patchwork-deploy/internal/config"
)

// ShutdownFunc is called when a termination signal is received.
type ShutdownFunc func(ctx context.Context) error

// Handler listens for OS signals and triggers graceful shutdown.
type Handler struct {
	cfg      *config.Config
	hooks    []ShutdownFunc
	signals  []os.Signal
}

// NewHandler creates a Handler that responds to SIGINT and SIGTERM by default.
func NewHandler(cfg *config.Config, fns ...ShutdownFunc) *Handler {
	return &Handler{
		cfg:     cfg,
		hooks:   fns,
		signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

// WithSignals overrides the default set of handled signals.
func (h *Handler) WithSignals(sigs ...os.Signal) *Handler {
	h.signals = sigs
	return h
}

// AddHook appends a ShutdownFunc to the list of hooks that are called on shutdown.
func (h *Handler) AddHook(fn ShutdownFunc) {
	h.hooks = append(h.hooks, fn)
}

// Wait blocks until a signal is received, then runs all shutdown hooks
// in registration order. Returns the first non-nil error encountered.
func (h *Handler) Wait(ctx context.Context) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, h.signals...)
	defer signal.Stop(ch)

	select {
	case <-ch:
	case <-ctx.Done():
		return ctx.Err()
	}

	for _, fn := range h.hooks {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Notify sends a signal to the handler's channel for testing purposes.
func (h *Handler) Notify(sig os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sig)
	ch <- sig
	signal.Stop(ch)
}
