package status

import (
	"fmt"
	"time"

	"github.com/yourorg/patchwork-deploy/internal/config"
)

// AppStatus holds the current deployment status of a single app.
type AppStatus struct {
	App       string    `json:"app"`
	Host      string    `json:"host"`
	Deployed  bool      `json:"deployed"`
	Version   string    `json:"version,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// Store is the interface for persisting and retrieving app statuses.
type Store interface {
	Save(appName string, s AppStatus) error
	Load(appName string) (AppStatus, error)
}

// Tracker records and retrieves deployment statuses.
type Tracker struct {
	cfg   *config.Config
	store Store
}

// NewTracker creates a Tracker backed by the given store.
func NewTracker(cfg *config.Config, store Store) *Tracker {
	return &Tracker{cfg: cfg, store: store}
}

// Record saves a new status entry for the given app.
func (t *Tracker) Record(appName, version, message string, deployed bool) error {
	app, ok := t.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("status: unknown app %q", appName)
	}
	s := AppStatus{
		App:       appName,
		Host:      app.Host,
		Deployed:  deployed,
		Version:   version,
		Timestamp: time.Now().UTC(),
		Message:   message,
	}
	return t.store.Save(appName, s)
}

// Get retrieves the latest status for the given app.
func (t *Tracker) Get(appName string) (AppStatus, error) {
	if _, ok := t.cfg.Apps[appName]; !ok {
		return AppStatus{}, fmt.Errorf("status: unknown app %q", appName)
	}
	return t.store.Load(appName)
}

// Summary returns a slice of AppStatus for all known apps, collecting any
// load errors into a combined error returned alongside the partial results.
func (t *Tracker) Summary() ([]AppStatus, error) {
	var statuses []AppStatus
	var errs []error
	for appName := range t.cfg.Apps {
		s, err := t.store.Load(appName)
		if err != nil {
			errs = append(errs, fmt.Errorf("status: loading %q: %w", appName, err))
			continue
		}
		statuses = append(statuses, s)
	}
	if len(errs) > 0 {
		return statuses, fmt.Errorf("status: summary encountered %d error(s): %v", len(errs), errs)
	}
	return statuses, nil
}
