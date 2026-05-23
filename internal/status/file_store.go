package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileStore persists AppStatus entries as JSON files on disk.
type FileStore struct {
	dir string
}

// NewFileStore creates a FileStore rooted at dir, creating it if necessary.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("status: create store dir: %w", err)
	}
	return &FileStore{dir: dir}, nil
}

// Save writes the AppStatus for appName to disk.
func (f *FileStore) Save(appName string, s AppStatus) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("status: marshal: %w", err)
	}
	return os.WriteFile(f.path(appName), data, 0o644)
}

// Load reads the AppStatus for appName from disk.
func (f *FileStore) Load(appName string) (AppStatus, error) {
	data, err := os.ReadFile(f.path(appName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AppStatus{}, fmt.Errorf("status: no record for %q", appName)
		}
		return AppStatus{}, fmt.Errorf("status: read: %w", err)
	}
	var s AppStatus
	if err := json.Unmarshal(data, &s); err != nil {
		return AppStatus{}, fmt.Errorf("status: unmarshal: %w", err)
	}
	return s, nil
}

func (f *FileStore) path(appName string) string {
	return filepath.Join(f.dir, appName+".json")
}
