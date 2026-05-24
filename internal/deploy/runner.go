package deploy

import (
	"context"
	"fmt"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/retry"
	"github.com/yourorg/patchwork-deploy/internal/ssh"
)

// Runner orchestrates deployment of apps defined in the config.
type Runner struct {
	cfg     *config.Config
	pool    *ssh.Pool
	retryer *retry.Retryer
}

// NewRunner constructs a Runner with a default retry policy.
func NewRunner(cfg *config.Config, pool *ssh.Pool) *Runner {
	return &Runner{
		cfg:     cfg,
		pool:    pool,
		retryer: retry.NewRetryer(retry.DefaultPolicy()),
	}
}

// WithRetryer replaces the retry policy used during deployment.
func (r *Runner) WithRetryer(retryer *retry.Retryer) *Runner {
	r.retryer = retryer
	return r
}

// Deploy uploads and activates the specified app on its target host.
func (r *Runner) Deploy(ctx context.Context, appName string) error {
	app, ok := r.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("deploy: unknown app %q", appName)
	}

	host, ok := r.cfg.Hosts[app.Host]
	if !ok {
		return fmt.Errorf("deploy: unknown host %q for app %q", app.Host, appName)
	}
	_ = host

	return r.retryer.Do(ctx, func() error {
		client, err := r.pool.Get(app.Host)
		if err != nil {
			return fmt.Errorf("deploy: ssh connect to %q: %w", app.Host, err)
		}

		if err := ssh.Upload(client, app.Dir, app.RemoteDir); err != nil {
			return fmt.Errorf("deploy: upload for app %q: %w", appName, err)
		}
		return nil
	})
}
