package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/ssh"
)

// Cleaner removes old release directories on remote hosts.
type Cleaner struct {
	cfg  *config.Config
	pool *ssh.Pool
}

// NewCleaner returns a new Cleaner.
func NewCleaner(cfg *config.Config, pool *ssh.Pool) *Cleaner {
	return &Cleaner{cfg: cfg, pool: pool}
}

// Prune keeps the most recent `keep` releases for the given app and removes
// older ones from every host the app is deployed to.
func (c *Cleaner) Prune(appName string, keep int) error {
	if keep < 1 {
		keep = 3
	}

	app, ok := c.cfg.Apps[appName]
	if !ok {
		return fmt.Errorf("cleanup: unknown app %q", appName)
	}

	for _, hostRef := range app.Hosts {
		client, err := c.pool.Get(hostRef)
		if err != nil {
			return fmt.Errorf("cleanup: connect to %s: %w", hostRef, err)
		}

		releasesDir := filepath.Join(app.Dir, "releases")

		// List release directories sorted oldest-first.
		listCmd := fmt.Sprintf(
			`ls -1dt %s/*/ 2>/dev/null | tail -n +%d`,
			releasesDir, keep+1,
		)

		out, err := client.Run(listCmd)
		if err != nil {
			// No old releases to clean — not an error.
			continue
		}

		if len(out) == 0 {
			continue
		}

		rmCmd := fmt.Sprintf("xargs -r rm -rf <<'EOF'\n%sEOF", out)
		if _, err := client.Run(rmCmd); err != nil {
			return fmt.Errorf("cleanup: remove old releases on %s: %w", hostRef, err)
		}
	}

	return nil
}
