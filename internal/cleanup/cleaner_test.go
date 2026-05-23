package cleanup_test

import (
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/cleanup"
	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/ssh"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Address: "192.0.2.1", User: "deploy", KeyFile: "/nonexistent/key"},
		},
		Apps: map[string]config.App{
			"myapp": {
				Dir:   "/srv/myapp",
				Hosts: []string{"web"},
			},
		},
	}
}

func TestNewCleaner_NotNil(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	c := cleanup.NewCleaner(cfg, pool)
	if c == nil {
		t.Fatal("expected non-nil Cleaner")
	}
}

func TestPrune_UnknownApp(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	c := cleanup.NewCleaner(cfg, pool)

	err := c.Prune("ghost", 3)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestPrune_SSHFailure(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	c := cleanup.NewCleaner(cfg, pool)

	// Key file does not exist, so SSH connection will fail.
	err := c.Prune("myapp", 3)
	if err == nil {
		t.Fatal("expected SSH error when key file is missing")
	}
}

func TestPrune_DefaultKeep(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	c := cleanup.NewCleaner(cfg, pool)

	// keep=0 should be normalised to 3 internally; still fails on SSH, not on
	// validation — confirming the keep guard doesn't panic.
	err := c.Prune("myapp", 0)
	if err == nil {
		t.Fatal("expected SSH error, not a panic or nil")
	}
}
