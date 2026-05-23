package lifecycle_test

import (
	"errors"
	"testing"

	"github.com/user/patchwork-deploy/internal/config"
	"github.com/user/patchwork-deploy/internal/lifecycle"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Address: "10.0.0.1", User: "deploy", KeyFile: "/id_rsa"},
		},
		Apps: map[string]config.App{
			"api": {
				Host: "web",
				Dir:  "/srv/api",
				Hooks: config.Hooks{
					PreDeploy:  []string{"echo pre"},
					PostDeploy: []string{"systemctl restart api", "echo done"},
					OnFailure:  []string{"echo fail"},
				},
			},
			"nohooks": {
				Host: "web",
				Dir:  "/srv/nohooks",
			},
		},
	}
}

func TestRun_UnknownApp(t *testing.T) {
	r := lifecycle.NewRunner(baseConfig(), func(_, _ string) error { return nil })
	err := r.Run("ghost", lifecycle.HookPostDeploy)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRun_NoHooks_IsNoop(t *testing.T) {
	r := lifecycle.NewRunner(baseConfig(), func(_, _ string) error {
		t.Fatal("execFn should not be called")
		return nil
	})
	if err := r.Run("nohooks", lifecycle.HookPreDeploy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_ExecutesAllCommands(t *testing.T) {
	var executed []string
	r := lifecycle.NewRunner(baseConfig(), func(host, cmd string) error {
		executed = append(executed, cmd)
		return nil
	})
	if err := r.Run("api", lifecycle.HookPostDeploy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 2 {
		t.Fatalf("expected 2 commands executed, got %d", len(executed))
	}
}

func TestRun_StopsOnFirstError(t *testing.T) {
	callCount := 0
	boom := errors.New("cmd failed")
	r := lifecycle.NewRunner(baseConfig(), func(_, _ string) error {
		callCount++
		return boom
	})
	err := r.Run("api", lifecycle.HookPostDeploy)
	if err == nil {
		t.Fatal("expected error")
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call before stop, got %d", callCount)
	}
}

func TestRun_CorrectHostPassed(t *testing.T) {
	var gotHost string
	r := lifecycle.NewRunner(baseConfig(), func(host, _ string) error {
		gotHost = host
		return nil
	})
	_ = r.Run("api", lifecycle.HookPreDeploy)
	if gotHost != "10.0.0.1" {
		t.Fatalf("expected host 10.0.0.1, got %q", gotHost)
	}
}
