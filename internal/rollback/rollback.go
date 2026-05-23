package rollback

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/yourorg/patchwork-deploy/internal/config"
)

// Snapshot represents a recorded deployment state for a single app.
type Snapshot struct {
	App       string
	Host      string
	Timestamp time.Time
	Revision  string // git SHA or tag captured at deploy time
	RemoteDir string
}

// Manager handles saving and restoring deployment snapshots.
type Manager struct {
	store Store
}

// NewManager creates a Manager backed by the given Store.
func NewManager(s Store) *Manager {
	return &Manager{store: s}
}

// Record saves a snapshot after a successful deployment.
func (m *Manager) Record(app, host, revision string, cfg *config.Config) error {
	a, err := findApp(app, cfg)
	if err != nil {
		return err
	}
	snap := Snapshot{
		App:       app,
		Host:      host,
		Timestamp: time.Now().UTC(),
		Revision:  revision,
		RemoteDir: filepath.Join("/opt/patchwork", a.Name),
	}
	return m.store.Save(snap)
}

// Latest returns the most recent snapshot for the given app+host pair.
func (m *Manager) Latest(app, host string) (*Snapshot, error) {
	return m.store.Load(app, host)
}

func findApp(name string, cfg *config.Config) (config.App, error) {
	for _, a := range cfg.Apps {
		if a.Name == name {
			return a, nil
		}
	}
	return config.App{}, fmt.Errorf("rollback: app %q not found in config", name)
}
