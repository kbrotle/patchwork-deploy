package drain_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/drain"
	"github.com/patchwork-deploy/internal/ssh"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Addr: "127.0.0.1", User: "root", KeyFile: "/nonexistent"},
		},
		Apps: map[string]config.App{
			"myapp": {
				Host: "web",
				Dir:  "/srv/myapp",
			},
			"drained": {
				Host: "web",
				Dir:  "/srv/drained",
				Drain: &config.DrainConfig{
					Command: "systemctl stop myapp",
					Timeout: 0, // zero → use default (won't actually sleep in test)
				},
			},
		},
	}
}

func TestNewDrainer_NotNil(t *testing.T) {
	d := drain.NewDrainer(baseConfig(), ssh.NewPool(baseConfig()))
	if d == nil {
		t.Fatal("expected non-nil Drainer")
	}
}

func TestDrain_UnknownApp(t *testing.T) {
	d := drain.NewDrainer(baseConfig(), ssh.NewPool(baseConfig()))
	if err := d.Drain("ghost"); err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestDrain_NoDrainConfig_IsNoop(t *testing.T) {
	// "myapp" has no Drain config — should return nil immediately.
	d := drain.NewDrainer(baseConfig(), ssh.NewPool(baseConfig()))
	if err := d.Drain("myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrain_SSHFailure(t *testing.T) {
	// "drained" has a Drain config but the SSH key doesn't exist,
	// so pool.Get should fail.
	d := drain.NewDrainer(baseConfig(), ssh.NewPool(baseConfig()))
	if err := d.Drain("drained"); err == nil {
		t.Fatal("expected SSH error for bad key file")
	}
}
