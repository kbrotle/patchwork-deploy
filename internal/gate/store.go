package gate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	dir string
}

func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("gate store: mkdir %s: %w", dir, err)
	}
	return &FileStore{dir: dir}, nil
}

type record struct {
	Open bool `json:"open"`
}

func (s *FileStore) Load(app string) (bool, error) {
	path := filepath.Join(s.dir, app+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("gate store: no record for %q", app)
		}
		return false, err
	}
	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return false, fmt.Errorf("gate store: decode %q: %w", app, err)
	}
	return r.Open, nil
}

func (s *FileStore) Save(app string, open bool) error {
	path := filepath.Join(s.dir, app+".json")
	data, err := json.Marshal(record{Open: open})
	if err != nil {
		return fmt.Errorf("gate store: encode %q: %w", app, err)
	}
	return os.WriteFile(path, data, 0o644)
}
