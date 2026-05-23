package audit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/audit"
)

func tempLogger(t *testing.T) *audit.Logger {
	t.Helper()
	dir := t.TempDir()
	l, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	return l
}

func TestRecord_And_ReadAll(t *testing.T) {
	l := tempLogger(t)

	if err := l.Record("myapp", "deploy", "success", "v1.2.3"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := l.Record("myapp", "rollback", "success", "v1.2.2"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := l.ReadAll("myapp")
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Action != "deploy" {
		t.Errorf("expected deploy, got %s", events[0].Action)
	}
	if events[1].Action != "rollback" {
		t.Errorf("expected rollback, got %s", events[1].Action)
	}
}

func TestReadAll_NoLog_ReturnsNil(t *testing.T) {
	l := tempLogger(t)

	events, err := l.ReadAll("unknown-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events != nil {
		t.Errorf("expected nil events, got %v", events)
	}
}

func TestRecord_IsolatedPerApp(t *testing.T) {
	l := tempLogger(t)

	_ = l.Record("app-a", "deploy", "success", "")
	_ = l.Record("app-b", "deploy", "failure", "timeout")

	eventsA, _ := l.ReadAll("app-a")
	eventsB, _ := l.ReadAll("app-b")

	if len(eventsA) != 1 || eventsA[0].App != "app-a" {
		t.Errorf("app-a isolation failed: %+v", eventsA)
	}
	if len(eventsB) != 1 || eventsB[0].Status != "failure" {
		t.Errorf("app-b isolation failed: %+v", eventsB)
	}
}

func TestNewLogger_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "audit")

	_, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected dir to be created: %s", dir)
	}
}

func TestRecord_EventFields(t *testing.T) {
	l := tempLogger(t)

	_ = l.Record("webapp", "health-check", "failure", "connection refused")

	events, _ := l.ReadAll("webapp")
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	e := events[0]
	if e.App != "webapp" || e.Action != "health-check" || e.Status != "failure" || e.Message != "connection refused" {
		t.Errorf("unexpected event fields: %+v", e)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
