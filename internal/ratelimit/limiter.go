package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Limiter enforces a minimum interval between deploys for a given app.
type Limiter struct {
	cfg      *config.Config
	mu       sync.Mutex
	lastSeen map[string]time.Time
}

// NewLimiter creates a Limiter backed by the provided config.
func NewLimiter(cfg *config.Config) *Limiter {
	return &Limiter{
		cfg:      cfg,
		lastSeen: make(map[string]time.Time),
	}
}

// Allow returns nil if the app may proceed with a deploy, or an error if the
// minimum interval since the last deploy has not yet elapsed.
// minInterval of 0 means no rate limiting is applied.
func (l *Limiter) Allow(appName string, minInterval time.Duration) error {
	if !l.appExists(appName) {
		return fmt.Errorf("ratelimit: unknown app %q", appName)
	}

	if minInterval <= 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if last, ok := l.lastSeen[appName]; ok {
		elapsed := time.Since(last)
		if elapsed < minInterval {
			remaining := minInterval - elapsed
			return fmt.Errorf("ratelimit: app %q must wait %s before next deploy", appName, remaining.Round(time.Second))
		}
	}

	return nil
}

// Record marks the current time as the last deploy time for the app.
func (l *Limiter) Record(appName string) error {
	if !l.appExists(appName) {
		return fmt.Errorf("ratelimit: unknown app %q", appName)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.lastSeen[appName] = time.Now()
	return nil
}

// Reset clears the recorded deploy time for the app.
func (l *Limiter) Reset(appName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastSeen, appName)
}

func (l *Limiter) appExists(appName string) bool {
	for _, app := range l.cfg.Apps {
		if app.Name == appName {
			return true
		}
	}
	return false
}
