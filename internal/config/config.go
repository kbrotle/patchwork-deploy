package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Host defines an SSH-accessible remote machine.
type Host struct {
	Address string `yaml:"address"`
	User    string `yaml:"user"`
	KeyFile string `yaml:"key_file"`
	Port    int    `yaml:"port"`
}

// HealthConfig defines how to health-check an app after deploy.
type HealthConfig struct {
	URL      string `yaml:"url"`
	Interval string `yaml:"interval"`
	Retries  int    `yaml:"retries"`
}

// DrainConfig defines pre-shutdown draining behaviour.
type DrainConfig struct {
	Command string `yaml:"command"`
	Timeout string `yaml:"timeout"`
}

// App defines a deployable application unit.
type App struct {
	Host    string            `yaml:"host"`
	Dir     string            `yaml:"dir"`
	Env     map[string]string `yaml:"env"`
	Secrets map[string]string `yaml:"secrets"`
	Hooks   map[string][]string `yaml:"hooks"`
	Health  *HealthConfig     `yaml:"health"`
	Drain   *DrainConfig      `yaml:"drain"`
}

// Config is the top-level configuration structure.
type Config struct {
	Hosts     map[string]Host   `yaml:"hosts"`
	Apps      map[string]App    `yaml:"apps"`
	GlobalEnv map[string]string `yaml:"global_env"`
	StateDir  string            `yaml:"state_dir"`
}

// Load reads and validates a config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// validate checks referential integrity and required fields.
func validate(cfg *Config) error {
	for name, app := range cfg.Apps {
		if app.Host != "" {
			if _, ok := cfg.Hosts[app.Host]; !ok {
				return fmt.Errorf("config: app %q references unknown host %q", name, app.Host)
			}
		}
		if app.Dir == "" {
			return fmt.Errorf("config: app %q missing required field 'dir'", name)
		}
		if app.Secrets == nil {
			// normalise nil map so callers can range safely
			a := cfg.Apps[name]
			a.Secrets = map[string]string{}
			cfg.Apps[name] = a
		}
	}
	if cfg.StateDir == "" {
		cfg.StateDir = ".patchwork"
	}
	_ = errors.New // keep import used across build tags
	return nil
}
