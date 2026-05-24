package version

import (
	"fmt"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Entry represents a recorded version for a deployed app.
type Entry struct {
	App       string    `json:"app"`
	Tag       string    `json:"tag"`
	Commit    string    `json:"commit"`
	DeployedAt time.Time `json:"deployed_at"`
}

// Store persists and retrieves version entries.
type Store interface {
	Save(app string, e Entry) error
	Load(app string) (Entry, error)
}

// Tracker records and retrieves version information for apps.
type Tracker struct {
	cfg   *config.Config
	store Store
}

// NewTracker returns a Tracker backed by the provided store.
func NewTracker(cfg *config.Config, store Store) *Tracker {
	return &Tracker{cfg: cfg, store: store}
}

// Record saves a version entry for the given app.
func (t *Tracker) Record(app, tag, commit string) error {
	if _, ok := t.cfg.Apps[app]; !ok {
		return fmt.Errorf("version: unknown app %q", app)
	}
	if tag == "" {
		return fmt.Errorf("version: tag must not be empty")
	}
	e := Entry{
		App:        app,
		Tag:        tag,
		Commit:     commit,
		DeployedAt: time.Now().UTC(),
	}
	return t.store.Save(app, e)
}

// Current returns the latest recorded version for the given app.
func (t *Tracker) Current(app string) (Entry, error) {
	if _, ok := t.cfg.Apps[app]; !ok {
		return Entry{}, fmt.Errorf("version: unknown app %q", app)
	}
	return t.store.Load(app)
}
