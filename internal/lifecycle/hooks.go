package lifecycle

import (
	"fmt"
	"strings"

	"github.com/user/patchwork-deploy/internal/config"
)

// HookType represents a deployment lifecycle hook stage.
type HookType string

const (
	HookPreDeploy  HookType = "pre_deploy"
	HookPostDeploy HookType = "post_deploy"
	HookOnFailure  HookType = "on_failure"
)

// Runner executes lifecycle hooks for a given app.
type Runner struct {
	cfg    *config.Config
	execFn func(host, cmd string) error
}

// NewRunner creates a new lifecycle hook Runner.
// execFn is called with the target host and the shell command to run.
func NewRunner(cfg *config.Config, execFn func(host, cmd string) error) *Runner {
	return &Runner{cfg: cfg, execFn: execFn}
}

// Run executes all hooks of the given type for the named app.
// It stops and returns the first error encountered.
func (r *Runner) Run(appName string, hook HookType) error {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("lifecycle: unknown app %q", appName)
	}

	hooks := hooksForType(app, hook)
	if len(hooks) == 0 {
		return nil
	}

	host, err := resolveHost(r.cfg, app)
	if err != nil {
		return err
	}

	for _, cmd := range hooks {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		if err := r.execFn(host, cmd); err != nil {
			return fmt.Errorf("lifecycle: hook %s command %q failed: %w", hook, cmd, err)
		}
	}
	return nil
}

func hooksForType(app config.App, hook HookType) []string {
	switch hook {
	case HookPreDeploy:
		return app.Hooks.PreDeploy
	case HookPostDeploy:
		return app.Hooks.PostDeploy
	case HookOnFailure:
		return app.Hooks.OnFailure
	}
	return nil
}

func resolveHost(cfg *config.Config, app config.App) (string, error) {
	h, ok := cfg.Hosts[app.Host]
	if !ok {
		return "", fmt.Errorf("lifecycle: unknown host %q", app.Host)
	}
	return h.Address, nil
}
