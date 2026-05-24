package schedule

import (
	"fmt"
	"sync"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Entry holds a scheduled deploy job for an app.
type Entry struct {
	AppName  string
	Interval time.Duration
	LastRun  time.Time
	stopCh   chan struct{}
}

// DeployFunc is the function called when a scheduled deploy fires.
type DeployFunc func(appName string) error

// Scheduler manages periodic deploy triggers for configured apps.
type Scheduler struct {
	cfg     *config.Config
	entries map[string]*Entry
	mu      sync.Mutex
	deploy  DeployFunc
}

// NewScheduler creates a Scheduler backed by the given config and deploy function.
func NewScheduler(cfg *config.Config, fn DeployFunc) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		entries: make(map[string]*Entry),
		deploy:  fn,
	}
}

// Register adds a recurring deploy schedule for the named app.
func (s *Scheduler) Register(appName string, interval time.Duration) error {
	if _, ok := s.cfg.Apps[appName]; !ok {
		return fmt.Errorf("schedule: unknown app %q", appName)
	}
	if interval <= 0 {
		return fmt.Errorf("schedule: interval must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if e, exists := s.entries[appName]; exists {
		close(e.stopCh)
	}

	e := &Entry{
		AppName:  appName,
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
	s.entries[appName] = e
	go s.run(e)
	return nil
}

// Unregister stops the schedule for the named app.
func (s *Scheduler) Unregister(appName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[appName]; ok {
		close(e.stopCh)
		delete(s.entries, appName)
	}
}

// StopAll stops all running schedules.
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, e := range s.entries {
		close(e.stopCh)
		delete(s.entries, name)
	}
}

func (s *Scheduler) run(e *Entry) {
	ticker := time.NewTicker(e.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.LastRun = time.Now()
			_ = s.deploy(e.AppName)
		case <-e.stopCh:
			return
		}
	}
}
