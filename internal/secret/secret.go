package secret

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/patchwork-deploy/internal/config"
)

// Resolver resolves secret values for apps, supporting env var references
// and literal values defined in the config.
type Resolver struct {
	cfg *config.Config
}

// NewResolver creates a new secret Resolver.
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// Resolve returns the resolved secrets map for the given app.
// Values prefixed with "env:" are looked up from the process environment.
// All other values are treated as literals.
func (r *Resolver) Resolve(appName string) (map[string]string, error) {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("secret: unknown app %q", appName)
	}

	resolved := make(map[string]string, len(app.Secrets))
	for key, val := range app.Secrets {
		v, err := resolveValue(val)
		if err != nil {
			return nil, fmt.Errorf("secret: app %q key %q: %w", appName, key, err)
		}
		resolved[key] = v
	}
	return resolved, nil
}

// ToEnvSlice returns secrets as a slice of KEY=VALUE strings suitable for
// passing to remote commands.
func (r *Resolver) ToEnvSlice(appName string) ([]string, error) {
	secrets, err := r.Resolve(appName)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(secrets))
	for k, v := range secrets {
		out = append(out, k+"="+v)
	}
	return out, nil
}

// resolveValue resolves a single secret value.
func resolveValue(val string) (string, error) {
	const envPrefix = "env:"
	if strings.HasPrefix(val, envPrefix) {
		envKey := strings.TrimPrefix(val, envPrefix)
		if envKey == "" {
			return "", errors.New("empty env var name after 'env:' prefix")
		}
		resolved, found := os.LookupEnv(envKey)
		if !found {
			return "", fmt.Errorf("env var %q not set", envKey)
		}
		return resolved, nil
	}
	return val, nil
}
