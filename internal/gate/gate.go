package gate

import (
	"fmt"
	"sync"

	"github.com/patchwork-deploy/internal/config"
)

// Gate controls whether a deployment is allowed to proceed based on
// per-app feature flags / manual hold toggles.

type Store interface {
	Load(app string) (bool, error)
	Save(app string, open bool) error
}

type Gatekeeper struct {
	cfg   *config.Config
	store Store
	mu    sync.RWMutex
}

func NewGatekeeper(cfg *config.Config, store Store) *Gatekeeper {
	return &Gatekeeper{cfg: cfg, store: store}
}

// IsOpen returns true when deployments for the given app are allowed.
func (g *Gatekeeper) IsOpen(app string) (bool, error) {
	if !g.appExists(app) {
		return false, fmt.Errorf("gate: unknown app %q", app)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	open, err := g.store.Load(app)
	if err != nil {
		// default open when no record exists
		return true, nil
	}
	return open, nil
}

// Hold prevents deployments for the given app.
func (g *Gatekeeper) Hold(app string) error {
	if !g.appExists(app) {
		return fmt.Errorf("gate: unknown app %q", app)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.store.Save(app, false)
}

// Release allows deployments for the given app.
func (g *Gatekeeper) Release(app string) error {
	if !g.appExists(app) {
		return fmt.Errorf("gate: unknown app %q", app)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.store.Save(app, true)
}

func (g *Gatekeeper) appExists(app string) bool {
	for _, a := range g.cfg.Apps {
		if a.Name == app {
			return true
		}
	}
	return false
}
