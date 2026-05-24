package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Store persists snapshot metadata locally.
type Store interface {
	Save(appName string, snap *Snapshot) error
	Load(appName string) (*Snapshot, error)
	Delete(appName string) error
}

// FileStore is a JSON-backed snapshot store.
type FileStore struct {
	baseDir string
}

// NewFileStore creates a FileStore rooted at baseDir.
func NewFileStore(baseDir string) (*FileStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot store: mkdir: %w", err)
	}
	return &FileStore{baseDir: baseDir}, nil
}

func (s *FileStore) path(appName string) string {
	return filepath.Join(s.baseDir, appName+".json")
}

// Save writes snapshot metadata for appName to disk.
func (s *FileStore) Save(appName string, snap *Snapshot) error {
	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("snapshot store: marshal: %w", err)
	}
	if err := os.WriteFile(s.path(appName), data, 0o644); err != nil {
		return fmt.Errorf("snapshot store: write: %w", err)
	}
	return nil
}

// Load reads the latest snapshot metadata for appName.
// Returns (nil, nil) if no snapshot exists for the given appName.
func (s *FileStore) Load(appName string) (*Snapshot, error) {
	data, err := os.ReadFile(s.path(appName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshot store: read: %w", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("snapshot store: unmarshal: %w", err)
	}
	return &snap, nil
}

// Delete removes the snapshot file for appName from disk.
// Returns nil if the snapshot does not exist.
func (s *FileStore) Delete(appName string) error {
	if err := os.Remove(s.path(appName)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("snapshot store: delete: %w", err)
	}
	return nil
}
