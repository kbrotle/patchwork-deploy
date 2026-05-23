package health

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/user/patchwork-deploy/internal/config"
)

// Status represents the health check result for an app.
type Status struct {
	App     string
	Healthy bool
	Message string
}

// Checker performs health checks against deployed apps.
type Checker struct {
	cfg     *config.Config
	client  *http.Client
	timeout time.Duration
}

// NewChecker creates a new Checker with the given config and timeout.
func NewChecker(cfg *config.Config, timeout time.Duration) *Checker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Checker{
		cfg:     cfg,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Check runs a health check for the named app.
// It uses the app's HealthURL if set, otherwise falls back to a TCP port check.
func (c *Checker) Check(appName string) Status {
	app, ok := c.cfg.Apps[appName]
	if !ok {
		return Status{App: appName, Healthy: false, Message: "unknown app"}
	}

	if app.HealthURL != "" {
		return c.httpCheck(appName, app.HealthURL)
	}

	if app.HealthPort > 0 {
		host, ok := c.cfg.Hosts[app.Host]
		if !ok {
			return Status{App: appName, Healthy: false, Message: "unknown host"}
		}
		addr := fmt.Sprintf("%s:%d", host.Address, app.HealthPort)
		return c.tcpCheck(appName, addr)
	}

	return Status{App: appName, Healthy: true, Message: "no health check configured"}
}

func (c *Checker) httpCheck(appName, url string) Status {
	resp, err := c.client.Get(url)
	if err != nil {
		return Status{App: appName, Healthy: false, Message: fmt.Sprintf("http check failed: %v", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return Status{App: appName, Healthy: true, Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return Status{App: appName, Healthy: false, Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (c *Checker) tcpCheck(appName, addr string) Status {
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return Status{App: appName, Healthy: false, Message: fmt.Sprintf("tcp check failed: %v", err)}
	}
	conn.Close()
	return Status{App: appName, Healthy: true, Message: fmt.Sprintf("tcp ok: %s", addr)}
}
