package archive

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/ssh"
)

// Archiver creates remote tar archives of deployed app directories.
type Archiver struct {
	cfg  *config.Config
	pool *ssh.Pool
}

// NewArchiver returns a new Archiver.
func NewArchiver(cfg *config.Config, pool *ssh.Pool) *Archiver {
	return &Archiver{cfg: cfg, pool: pool}
}

// Archive creates a timestamped tar.gz of the app's remote directory.
// Returns the remote path of the created archive.
func (a *Archiver) Archive(appName string) (string, error) {
	app, ok := a.cfg.Apps[appName]
	if !ok {
		return "", fmt.Errorf("archive: unknown app %q", appName)
	}

	host, ok := a.cfg.Hosts[app.Host]
	if !ok {
		return "", fmt.Errorf("archive: unknown host %q for app %q", app.Host, appName)
	}

	client, err := a.pool.Get(app.Host, host)
	if err != nil {
		return "", fmt.Errorf("archive: ssh connect: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	archiveDir := filepath.Join(app.Dir, "..", "archives")
	archiveName := fmt.Sprintf("%s_%s.tar.gz", appName, timestamp)
	destPath := filepath.Join(archiveDir, archiveName)

	cmds := []string{
		fmt.Sprintf("mkdir -p %s", archiveDir),
		fmt.Sprintf("tar -czf %s -C %s .", destPath, app.Dir),
	}

	for _, cmd := range cmds {
		if err := client.Run(cmd); err != nil {
			return "", fmt.Errorf("archive: command %q failed: %w", cmd, err)
		}
	}

	return destPath, nil
}

// List returns all archive paths for the given app on its remote host.
func (a *Archiver) List(appName string) ([]string, error) {
	app, ok := a.cfg.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("archive: unknown app %q", appName)
	}

	host, ok := a.cfg.Hosts[app.Host]
	if !ok {
		return nil, fmt.Errorf("archive: unknown host %q for app %q", app.Host, appName)
	}

	client, err := a.pool.Get(app.Host, host)
	if err != nil {
		return nil, fmt.Errorf("archive: ssh connect: %w", err)
	}

	archiveDir := filepath.Join(app.Dir, "..", "archives")
	cmd := fmt.Sprintf("ls %s/%s_*.tar.gz 2>/dev/null || true", archiveDir, appName)

	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("archive: list failed: %w", err)
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}
