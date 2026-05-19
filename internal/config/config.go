package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level deployment configuration.
type Config struct {
	Hosts []Host `yaml:"hosts"`
	Apps  []App  `yaml:"apps"`
}

// Host defines a remote server to deploy to.
type Host struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	User       string `yaml:"user"`
	Port       int    `yaml:"port"`
	IdentityFile string `yaml:"identity_file"`
}

// App defines a deployable application unit.
type App struct {
	Name    string   `yaml:"name"`
	Host    string   `yaml:"host"`
	Dir     string   `yaml:"dir"`
	Steps   []string `yaml:"steps"`
	EnvFile string   `yaml:"env_file"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded config.
func (c *Config) validate() error {
	hostNames := make(map[string]struct{}, len(c.Hosts))
	for _, h := range c.Hosts {
		if h.Name == "" {
			return fmt.Errorf("host is missing a name")
		}
		if h.Address == "" {
			return fmt.Errorf("host %q is missing an address", h.Name)
		}
		if h.User == "" {
			return fmt.Errorf("host %q is missing a user", h.Name)
		}
		hostNames[h.Name] = struct{}{}
	}

	for _, a := range c.Apps {
		if a.Name == "" {
			return fmt.Errorf("app is missing a name")
		}
		if a.Host == "" {
			return fmt.Errorf("app %q is missing a host reference", a.Name)
		}
		if _, ok := hostNames[a.Host]; !ok {
			return fmt.Errorf("app %q references unknown host %q", a.Name, a.Host)
		}
		if a.Dir == "" {
			return fmt.Errorf("app %q is missing a working directory", a.Name)
		}
	}

	return nil
}
