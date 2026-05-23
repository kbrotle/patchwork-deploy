package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/patchwork-deploy/internal/config"
)

// Resolver resolves environment variables for a given app,
// merging global and app-level env declarations from config.
type Resolver struct {
	cfg *config.Config
}

// NewResolver creates a new Resolver backed by the given config.
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// Resolve returns the merged environment map for the named app.
// App-level values override global values. Values of the form
// "$VAR" or "${VAR}" are expanded from the host process environment.
func (r *Resolver) Resolve(appName string) (map[string]string, error) {
	var app *config.App
	for i := range r.cfg.Apps {
		if r.cfg.Apps[i].Name == appName {
			app = &r.cfg.Apps[i]
			break
		}
	}
	if app == nil {
		return nil, fmt.Errorf("env: unknown app %q", appName)
	}

	merged := make(map[string]string)

	for k, v := range r.cfg.GlobalEnv {
		merged[k] = expand(v)
	}
	for k, v := range app.Env {
		merged[k] = expand(v)
	}

	return merged, nil
}

// ToSlice converts an env map to a slice of "KEY=VALUE" strings
// suitable for use in SSH commands or exec calls.
func ToSlice(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

// expand replaces $VAR / ${VAR} references using os.Getenv.
func expand(val string) string {
	return os.Expand(val, func(key string) string {
		key = strings.TrimSpace(key)
		return os.Getenv(key)
	})
}
