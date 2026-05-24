package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourorg/patchwork-deploy/internal/retry"
)

func fastPolicy() retry.Policy {
	return retry.Policy{MaxAttempts: 3, Delay: 1 * time.Millisecond, Backoff: 1.0}
}

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	r := retry.NewRetryer(fastPolicy())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	r := retry.NewRetryer(fastPolicy())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	r := retry.NewRetryer(fastPolicy())
	sentinel := errors.New("always fails")
	err := r.Do(context.Background(), func() error { return sentinel })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel wrapped in error, got %v", err)
	}
}

func TestDo_CancelledContext(t *testing.T) {
	r := retry.NewRetryer(retry.Policy{MaxAttempts: 5, Delay: 100 * time.Millisecond, Backoff: 1.0})
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		cancel()
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error after cancel")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call before cancel propagated, got %d", calls)
	}
}

func TestNewRetryer_NotNil(t *testing.T) {
	r := retry.NewRetryer(retry.DefaultPolicy())
	if r == nil {
		t.Fatal("expected non-nil Retryer")
	}
}

func TestDefaultPolicy_SaneValues(t *testing.T) {
	p := retry.DefaultPolicy()
	if p.MaxAttempts <= 0 {
		t.Error("MaxAttempts must be positive")
	}
	if p.Delay <= 0 {
		t.Error("Delay must be positive")
	}
	if p.Backoff < 1.0 {
		t.Error("Backoff must be >= 1.0")
	}
}
