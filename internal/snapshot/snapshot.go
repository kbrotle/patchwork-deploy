package snapshot

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/ssh"
)

// Manager handles remote directory snapshots for rollback support.
type Manager struct {
	cfg  *config.Config
	pool *ssh.Pool
}

// Snapshot represents a point-in-time copy of a deployed app directory.
type Snapshot struct {
	App       string
	Timestamp time.Time
	RemotePath string
}

// NewManager creates a new snapshot Manager.
func NewManager(cfg *config.Config, pool *ssh.Pool) *Manager {
	return &Manager{cfg: cfg, pool: pool}
}

// Take creates a remote snapshot of the app's current directory via cp -a.
func (m *Manager) Take(appName string) (*Snapshot, error) {
	app, ok := m.cfg.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("snapshot: unknown app %q", appName)
	}

	host, ok := m.cfg.Hosts[app.Host]
	if !ok {
		return nil, fmt.Errorf("snapshot: unknown host %q for app %q", app.Host, appName)
	}

	client, err := m.pool.Get(app.Host, host)
	if err != nil {
		return nil, fmt.Errorf("snapshot: ssh connect: %w", err)
	}

	ts := time.Now().UTC()
	snapshotDir := filepath.Join(app.Dir, ".snapshots", ts.Format("20060102T150405Z"))
	cmd := fmt.Sprintf("cp -a %s %s", app.Dir, snapshotDir)

	if err := client.Run(cmd); err != nil {
		return nil, fmt.Errorf("snapshot: remote cp failed: %w", err)
	}

	return &Snapshot{
		App:        appName,
		Timestamp:  ts,
		RemotePath: snapshotDir,
	}, nil
}

// Restore replaces the app directory with a previously taken snapshot.
func (m *Manager) Restore(appName, snapshotPath string) error {
	app, ok := m.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("snapshot: unknown app %q", appName)
	}

	host, ok := m.cfg.Hosts[app.Host]
	if !ok {
		return fmt.Errorf("snapshot: unknown host %q for app %q", app.Host, appName)
	}

	client, err := m.pool.Get(app.Host, host)
	if err != nil {
		return fmt.Errorf("snapshot: ssh connect: %w", err)
	}

	cmd := fmt.Sprintf("rm -rf %s && cp -a %s %s", app.Dir, snapshotPath, app.Dir)
	if err := client.Run(cmd); err != nil {
		return fmt.Errorf("snapshot: restore failed: %w", err)
	}
	return nil
}
