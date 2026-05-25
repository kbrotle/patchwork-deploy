package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Store persists the last-seen file hash for each app.
type Store interface {
	Load(app string) (string, error)
	Save(app, hash string) error
}

// FileStore is a file-backed Store.
type FileStore struct {
	dir string
}

// NewFileStore returns a FileStore rooted at dir.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("watch store: mkdir: %w", err)
	}
	return &FileStore{dir: dir}, nil
}

type record struct {
	Hash string `json:"hash"`
}

func (s *FileStore) Load(app string) (string, error) {
	p := filepath.Join(s.dir, app+".json")
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("watch store: load %q: %w", app, err)
	}
	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("watch store: decode %q: %w", app, err)
	}
	return r.Hash, nil
}

func (s *FileStore) Save(app, hash string) error {
	p := filepath.Join(s.dir, app+".json")
	data, err := json.Marshal(record{Hash: hash})
	if err != nil {
		return fmt.Errorf("watch store: encode %q: %w", app, err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("watch store: save %q: %w", app, err)
	}
	return nil
}
