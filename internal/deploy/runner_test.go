package deploy_test

import (
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/deploy"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"prod": {
				Address: "192.0.2.1",
				Port:    22,
				User:    "deploy",
				KeyFile: "/nonexistent/key",
			},
		},
		Apps: map[string]config.App{
			"myapp": {
				Host:       "prod",
				LocalDir:   "/tmp/myapp",
				RemotePath: "/srv",
				Commands:   []string{"systemctl restart myapp"},
			},
		},
	}
}

func TestDeploy_UnknownApp(t *testing.T) {
	r := deploy.NewRunner(baseConfig())
	defer r.Close()

	err := r.Deploy("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app, got nil")
	}
}

func TestDeploy_SSHFailure(t *testing.T) {
	r := deploy.NewRunner(baseConfig())
	defer r.Close()

	// The key file doesn't exist, so SSH connection should fail.
	err := r.Deploy("myapp")
	if err == nil {
		t.Fatal("expected SSH error, got nil")
	}
}

func TestNewRunner_NotNil(t *testing.T) {
	r := deploy.NewRunner(baseConfig())
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
	r.Close()
}
