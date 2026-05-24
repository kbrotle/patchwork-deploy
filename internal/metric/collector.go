package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Entry holds a single recorded metric data point.
type Entry struct {
	App       string
	Name      string
	Value     float64
	Timestamp time.Time
}

// Collector records and retrieves named metrics per app.
type Collector struct {
	mu      sync.RWMutex
	cfg     *config.Config
	records map[string][]Entry
}

// NewCollector returns a Collector backed by the given config.
func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		cfg:     cfg,
		records: make(map[string][]Entry),
	}
}

// Record stores a named metric value for the given app.
func (c *Collector) Record(app, name string, value float64) error {
	if !c.appExists(app) {
		return fmt.Errorf("metric: unknown app %q", app)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e := Entry{
		App:       app,
		Name:      name,
		Value:     value,
		Timestamp: time.Now().UTC(),
	}
	c.records[app] = append(c.records[app], e)
	return nil
}

// Get returns all recorded entries for the given app.
func (c *Collector) Get(app string) ([]Entry, error) {
	if !c.appExists(app) {
		return nil, fmt.Errorf("metric: unknown app %q", app)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entries := make([]Entry, len(c.records[app]))
	copy(entries, c.records[app])
	return entries, nil
}

// Reset clears all recorded metrics for the given app.
func (c *Collector) Reset(app string) error {
	if !c.appExists(app) {
		return fmt.Errorf("metric: unknown app %q", app)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, app)
	return nil
}

func (c *Collector) appExists(app string) bool {
	for _, a := range c.cfg.Apps {
		if a.Name == app {
			return true
		}
	}
	return false
}
