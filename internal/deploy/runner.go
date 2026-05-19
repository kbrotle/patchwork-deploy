package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/ssh"
)

// Runner orchestrates deployments for configured apps.
type Runner struct {
	cfg  *config.Config
	pool *ssh.Pool
}

// NewRunner creates a new Runner using the provided config.
func NewRunner(cfg *config.Config) *Runner {
	return &Runner{
		cfg:  cfg,
		pool: ssh.NewPool(),
	}
}

// Deploy runs the deployment for the named app.
func (r *Runner) Deploy(appName string) error {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("app %q not found in config", appName)
	}

	host, ok := r.cfg.Hosts[app.Host]
	if !ok {
		return fmt.Errorf("host %q not found in config", app.Host)
	}

	client, err := r.pool.Get(app.Host, ssh.Config{
		Host:       host.Address,
		Port:       host.Port,
		User:       host.User,
		KeyFile:    host.KeyFile,
	})
	if err != nil {
		return fmt.Errorf("ssh connect to %s: %w", app.Host, err)
	}

	remoteDir := filepath.Join(app.RemotePath, appName)

	if err := client.Upload(app.LocalDir, remoteDir); err != nil {
		return fmt.Errorf("upload %s: %w", appName, err)
	}

	for _, cmd := range app.Commands {
		if err := client.Run(cmd); err != nil {
			return fmt.Errorf("command %q failed: %w", cmd, err)
		}
	}

	return nil
}

// Close releases all SSH connections held by the runner.
func (r *Runner) Close() {
	r.pool.CloseAll()
}
