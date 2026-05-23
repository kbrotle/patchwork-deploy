package rollback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Store is the persistence interface for snapshots.
type Store interface {
	Save(snap Snapshot) error
	Load(app, host string) (*Snapshot, error)
}

// FileStore persists snapshots as JSON files under a base directory.
type FileStore struct {
	BaseDir string
}

// NewFileStore creates a FileStore that writes under baseDir.
func NewFileStore(baseDir string) *FileStore {
	return &FileStore{BaseDir: baseDir}
}

func (f *FileStore) snapshotPath(app, host string) string {
	filename := fmt.Sprintf("%s_%s.json", app, host)
	return filepath.Join(f.BaseDir, filename)
}

// Save serialises the snapshot to disk, overwriting any previous entry.
func (f *FileStore) Save(snap Snapshot) error {
	if err := os.MkdirAll(f.BaseDir, 0o755); err != nil {
		return fmt.Errorf("filestore: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("filestore: marshal: %w", err)
	}
	path := f.snapshotPath(snap.App, snap.Host)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("filestore: write %s: %w", path, err)
	}
	return nil
}

// Load reads the latest snapshot for the given app+host pair.
func (f *FileStore) Load(app, host string) (*Snapshot, error) {
	path := f.snapshotPath(app, host)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("filestore: no snapshot for %s@%s", app, host)
		}
		return nil, fmt.Errorf("filestore: read %s: %w", path, err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("filestore: unmarshal: %w", err)
	}
	return &snap, nil
}
