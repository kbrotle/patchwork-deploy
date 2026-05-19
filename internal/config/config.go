package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Host describes an SSH target.
type Host struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	User    string `yaml:"user"`
	KeyFile string `yaml:"key_file"`
}

// App describes a deployable application.
type App struct {
	Host       string   `yaml:"host"`
	LocalDir   string   `yaml:"local_dir"`
	RemotePath string   `yaml:"remote_path"`
	Commands   []string `yaml:"commands"`
}

// Config is the top-level configuration structure.
type Config struct {
	Hosts map[string]Host `yaml:"hosts"`
	Apps  map[string]App  `yaml:"apps"`
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	for name, app := range cfg.Apps {
		if _, ok := cfg.Hosts[app.Host]; !ok {
			return fmt.Errorf("app %q references unknown host %q", name, app.Host)
		}
		if app.LocalDir == "" {
			return fmt.Errorf("app %q missing local_dir", name)
		}
	}
	for name, host := range cfg.Hosts {
		if host.Port == 0 {
			cfg.Hosts[name] = Host{
				Address: host.Address,
				Port:    22,
				User:    host.User,
				KeyFile: host.KeyFile,
			}
		}
	}
	return nil
}
