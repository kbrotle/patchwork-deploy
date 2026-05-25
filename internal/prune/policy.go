package prune

import (
	"fmt"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Policy describes retention rules for a single app.
type Policy struct {
	KeepLast    int
	MaxAgeDays  int
	AppName     string
}

// Manager evaluates pruning policies against a list of timestamped entries.
type Manager struct {
	cfg *config.Config
}

// NewManager returns a Manager for the given config.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

// PolicyFor returns the pruning policy for the named app.
func (m *Manager) PolicyFor(appName string) (Policy, error) {
	for _, app := range m.cfg.Apps {
		if app.Name == appName {
			kl := app.Prune.KeepLast
			if kl <= 0 {
				kl = 5
			}
			return Policy{
				AppName:    appName,
				KeepLast:   kl,
				MaxAgeDays: app.Prune.MaxAgeDays,
			}, nil
		}
	}
	return Policy{}, fmt.Errorf("prune: unknown app %q", appName)
}

// Apply filters entries according to the policy, returning those that should
// be removed. Entries must be sorted oldest-first.
func (m *Manager) Apply(p Policy, entries []Entry) []Entry {
	if len(entries) == 0 {
		return nil
	}

	var remove []Entry
	now := time.Now()

	// Mark entries that exceed max age.
	if p.MaxAgeDays > 0 {
		cutoff := now.AddDate(0, 0, -p.MaxAgeDays)
		for _, e := range entries {
			if e.CreatedAt.Before(cutoff) {
				remove = append(remove, e)
			}
		}
	}

	// After age filtering, enforce KeepLast on the remaining entries.
	remaining := without(entries, remove)
	if len(remaining) > p.KeepLast {
		excess := remaining[:len(remaining)-p.KeepLast]
		remove = append(remove, excess...)
	}

	return unique(remove)
}

// Entry represents a single prunable artifact (e.g. a build or snapshot).
type Entry struct {
	ID        string
	CreatedAt time.Time
}

func without(all, exclude []Entry) []Entry {
	exSet := make(map[string]struct{}, len(exclude))
	for _, e := range exclude {
		exSet[e.ID] = struct{}{}
	}
	out := make([]Entry, 0, len(all))
	for _, e := range all {
		if _, skip := exSet[e.ID]; !skip {
			out = append(out, e)
		}
	}
	return out
}

func unique(entries []Entry) []Entry {
	seen := make(map[string]struct{}, len(entries))
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if _, ok := seen[e.ID]; !ok {
			seen[e.ID] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}
