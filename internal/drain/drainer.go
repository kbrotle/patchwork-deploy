package drain

import (
	"fmt"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/ssh"
)

// Drainer sends a drain signal to an app before deployment,
// waiting for in-flight requests to complete.
type Drainer struct {
	cfg  *config.Config
	pool *ssh.Pool
}

// NewDrainer creates a new Drainer.
func NewDrainer(cfg *config.Config, pool *ssh.Pool) *Drainer {
	return &Drainer{cfg: cfg, pool: pool}
}

// Drain signals the app to stop accepting new connections and waits
// for the configured drain timeout before returning.
func (d *Drainer) Drain(appName string) error {
	app, ok := d.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("drain: unknown app %q", appName)
	}

	if app.Drain == nil {
		// No drain config — nothing to do.
		return nil
	}

	client, err := d.pool.Get(app.Host)
	if err != nil {
		return fmt.Errorf("drain: ssh connect to %q: %w", app.Host, err)
	}

	if app.Drain.Command != "" {
		if err := client.Run(app.Drain.Command); err != nil {
			return fmt.Errorf("drain: command failed for %q: %w", appName, err)
		}
	}

	timeout := app.Drain.Timeout
	if timeout <= 0 {
		timeout = 15
	}
	time.Sleep(time.Duration(timeout) * time.Second)

	return nil
}
