package diff

import (
	"fmt"
	"strings"

	"github.com/patchwork-deploy/internal/config"
)

// Change represents a single detected change between two snapshots.
type Change struct {
	Field string
	Old   string
	New   string
}

// Result holds the full diff result for an app.
type Result struct {
	App     string
	Changes []Change
}

// HasChanges returns true if any changes were detected.
func (r *Result) HasChanges() bool {
	return len(r.Changes) > 0
}

// Summary returns a human-readable summary of the diff.
func (r *Result) Summary() string {
	if !r.HasChanges() {
		return fmt.Sprintf("app %s: no changes", r.App)
	}
	lines := make([]string, 0, len(r.Changes))
	for _, c := range r.Changes {
		lines = append(lines, fmt.Sprintf("  %s: %q -> %q", c.Field, c.Old, c.New))
	}
	return fmt.Sprintf("app %s:\n%s", r.App, strings.Join(lines, "\n"))
}

// Differ compares app configurations to detect deployment-relevant changes.
type Differ struct {
	cfg *config.Config
}

// NewDiffer creates a new Differ for the given config.
func NewDiffer(cfg *config.Config) *Differ {
	return &Differ{cfg: cfg}
}

// Compare computes the diff between two app config snapshots represented
// as key-value maps (e.g. serialised from config at two points in time).
func (d *Differ) Compare(appName string, prev, next map[string]string) (*Result, error) {
	found := false
	for _, app := range d.cfg.Apps {
		if app.Name == appName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("diff: unknown app %q", appName)
	}

	result := &Result{App: appName}

	// Detect changed or removed keys.
	for k, oldVal := range prev {
		newVal, ok := next[k]
		if !ok {
			result.Changes = append(result.Changes, Change{Field: k, Old: oldVal, New: ""})
		} else if oldVal != newVal {
			result.Changes = append(result.Changes, Change{Field: k, Old: oldVal, New: newVal})
		}
	}

	// Detect added keys.
	for k, newVal := range next {
		if _, ok := prev[k]; !ok {
			result.Changes = append(result.Changes, Change{Field: k, Old: "", New: newVal})
		}
	}

	return result, nil
}
