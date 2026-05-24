package retry

import (
	"context"
	"fmt"
	"time"
)

// Policy defines retry behaviour for an operation.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     float64 // multiplier applied to Delay after each attempt (1.0 = no backoff)
}

// DefaultPolicy returns a sensible default retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
		Backoff:     2.0,
	}
}

// Retryer executes operations with retry logic.
type Retryer struct {
	policy Policy
}

// NewRetryer creates a Retryer with the given policy.
func NewRetryer(p Policy) *Retryer {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 1
	}
	if p.Backoff < 1.0 {
		p.Backoff = 1.0
	}
	return &Retryer{policy: p}
}

// Do runs fn up to MaxAttempts times, waiting between attempts.
// It stops early if ctx is cancelled or fn returns nil.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := r.policy.Delay

	for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled after %d attempt(s): %w", attempt-1, ctx.Err())
		default:
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt < r.policy.MaxAttempts {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			}
			delay = time.Duration(float64(delay) * r.policy.Backoff)
		}
	}

	return fmt.Errorf("all %d attempt(s) failed: %w", r.policy.MaxAttempts, lastErr)
}
