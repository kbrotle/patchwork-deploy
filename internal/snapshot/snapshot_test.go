package snapshot_test

import (
	"fmt"
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/snapshot"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"prod": {User: "deploy", Addr: "127.0.0.1", KeyFile: "/tmp/missing_key"},
		},
		Apps: map[string]config.App{
			"web": {Host: "prod", Dir: "/srv/web"},
		},
	}
}

func TestTake_UnknownApp(t *testing.T) {
	mgr := snapshot.NewManager(baseConfig(), nil)
	_, err := mgr.Take("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRestore_UnknownApp(t *testing.T) {
	mgr := snapshot.NewManager(baseConfig(), nil)
	err := mgr.Restore("ghost", "/srv/ghost/.snapshots/x")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestTake_SSHFailure(t *testing.T) {
	cfg := baseConfig()
	// pool is nil — Get will panic; use a config with bad host to trigger error path
	cfg.Apps["web"] = config.App{Host: "prod", Dir: "/srv/web"}
	cfg.Hosts["prod"] = config.Host{
		User:    "deploy",
		Addr:    "127.0.0.1",
		KeyFile: "/nonexistent/key",
	}

	// Use a real pool so Get returns an error on bad key
	pool := ssh.NewPool()
	defer pool.CloseAll()

	mgr := snapshot.NewManager(cfg, pool)
	_, err := mgr.Take("web")
	if err == nil {
		t.Fatal("expected SSH error")
	}
}

func TestNewManager_NotNil(t *testing.T) {
	mgr := snapshot.NewManager(baseConfig(), nil)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestSnapshot_Fields(t *testing.T) {
	snap := &snapshot.Snapshot{
		App:        "api",
		RemotePath: "/srv/api/.snapshots/20240101T000000Z",
	}
	if snap.App != "api" {
		t.Errorf("unexpected App: %s", snap.App)
	}
	if snap.RemotePath == "" {
		t.Error("RemotePath should not be empty")
	}
	_ = fmt.Sprintf("%+v", snap) // ensure Snapshot is printable
}
