package lock

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func tempLocker(t *testing.T) *Locker {
	t.Helper()
	dir := t.TempDir()
	l, err := NewLocker(dir)
	if err != nil {
		t.Fatalf("NewLocker: %v", err)
	}
	return l
}

func TestAcquire_And_Release(t *testing.T) {
	l := tempLocker(t)

	if err := l.Acquire("myapp"); err != nil {
		t.Fatalf("expected no error on first acquire, got: %v", err)
	}
	if !l.IsLocked("myapp") {
		t.Fatal("expected app to be locked")
	}
	if err := l.Release("myapp"); err != nil {
		t.Fatalf("release: %v", err)
	}
	if l.IsLocked("myapp") {
		t.Fatal("expected app to be unlocked after release")
	}
}

func TestAcquire_FailsWhenAlreadyLocked(t *testing.T) {
	l := tempLocker(t)

	if err := l.Acquire("myapp"); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer l.Release("myapp") //nolint

	if err := l.Acquire("myapp"); err == nil {
		t.Fatal("expected error on second acquire, got nil")
	}
}

func TestRelease_NonExistent_IsNoop(t *testing.T) {
	l := tempLocker(t)
	if err := l.Release("ghost"); err != nil {
		t.Fatalf("expected no error releasing non-existent lock, got: %v", err)
	}
}

func TestNewLocker_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "locks")
	_, err := NewLocker(dir)
	if err != nil {
		t.Fatalf("NewLocker with nested path: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
}

func TestAcquire_DifferentApps_Independent(t *testing.T) {
	l := tempLocker(t)
	if err := l.Acquire("app1"); err != nil {
		t.Fatalf("app1 acquire: %v", err)
	}
	if err := l.Acquire("app2"); err != nil {
		t.Fatalf("app2 acquire: %v", err)
	}
	l.Release("app1") //nolint
	l.Release("app2") //nolint
}

func TestAcquire_Concurrent_OnlyOneSucceeds(t *testing.T) {
	l := tempLocker(t)
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		success int
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Acquire("concurrent-app"); err == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success != 1 {
		t.Fatalf("expected exactly 1 successful acquire, got %d", success)
	}
}
