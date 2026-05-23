package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Locker manages per-app deployment locks to prevent concurrent deploys.
type Locker struct {
	dir string
}

// NewLocker creates a Locker that stores lock files under dir.
func NewLocker(dir string) (*Locker, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("lock: create dir: %w", err)
	}
	return &Locker{dir: dir}, nil
}

// Acquire attempts to acquire a lock for appName.
// Returns an error if the lock is already held.
func (l *Locker) Acquire(appName string) error {
	path := l.lockPath(appName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			owner, _ := l.readLock(path)
			return fmt.Errorf("lock: app %q is already being deployed (lock: %s)", appName, owner)
		}
		return fmt.Errorf("lock: acquire: %w", err)
	}
	defer f.Close()
	pid := os.Getpid()
	_, err = fmt.Fprintf(f, "%d\n%d", pid, time.Now().Unix())
	return err
}

// Release removes the lock for appName.
func (l *Locker) Release(appName string) error {
	path := l.lockPath(appName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("lock: release %q: %w", appName, err)
	}
	return nil
}

// IsLocked reports whether appName currently holds a lock.
func (l *Locker) IsLocked(appName string) bool {
	_, err := os.Stat(l.lockPath(appName))
	return err == nil
}

func (l *Locker) lockPath(appName string) string {
	return filepath.Join(l.dir, appName+".lock")
}

func (l *Locker) readLock(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
	if len(parts) == 2 {
		ts, _ := strconv.ParseInt(parts[1], 10, 64)
		t := time.Unix(ts, 0).Format(time.RFC3339)
		return fmt.Sprintf("pid=%s since=%s", parts[0], t), nil
	}
	return string(data), nil
}
