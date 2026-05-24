package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScheduleRecord persists a schedule entry to disk.
type ScheduleRecord struct {
	AppName  string        `json:"app_name"`
	Interval time.Duration `json:"interval_ns"`
	Created  time.Time     `json:"created"`
}

// FileStore persists schedule records as JSON files.
type FileStore struct {
	dir string
}

// NewFileStore creates a FileStore rooted at dir.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("schedule store: mkdir: %w", err)
	}
	return &FileStore{dir: dir}, nil
}

// Save writes a ScheduleRecord for the given app.
func (fs *FileStore) Save(rec ScheduleRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("schedule store: marshal: %w", err)
	}
	path := filepath.Join(fs.dir, rec.AppName+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("schedule store: write: %w", err)
	}
	return nil
}

// Load reads the ScheduleRecord for the given app.
func (fs *FileStore) Load(appName string) (ScheduleRecord, error) {
	path := filepath.Join(fs.dir, appName+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ScheduleRecord{}, fmt.Errorf("schedule store: no record for %q", appName)
		}
		return ScheduleRecord{}, fmt.Errorf("schedule store: read: %w", err)
	}
	var rec ScheduleRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return ScheduleRecord{}, fmt.Errorf("schedule store: unmarshal: %w", err)
	}
	return rec, nil
}

// Delete removes the stored record for the given app.
func (fs *FileStore) Delete(appName string) error {
	path := filepath.Join(fs.dir, appName+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("schedule store: delete: %w", err)
	}
	return nil
}
