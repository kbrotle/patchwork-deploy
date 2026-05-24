package pipeline

import (
	"context"
	"fmt"

	"github.com/yourorg/patchwork-deploy/internal/config"
)

// Stage represents a named deployment pipeline step.
type Stage struct {
	Name string
	Run  func(ctx context.Context, appName string) error
}

// Pipeline executes an ordered sequence of stages for a given app.
type Pipeline struct {
	cfg    *config.Config
	stages []Stage
}

// NewPipeline creates a Pipeline bound to the provided config.
func NewPipeline(cfg *config.Config, stages []Stage) *Pipeline {
	return &Pipeline{cfg: cfg, stages: stages}
}

// Run executes each stage in order. It stops and returns the first error
// encountered, annotating it with the failing stage name.
func (p *Pipeline) Run(ctx context.Context, appName string) error {
	if !p.appExists(appName) {
		return fmt.Errorf("pipeline: unknown app %q", appName)
	}

	for _, stage := range p.stages {
		select {
		case <-ctx.Done():
			return fmt.Errorf("pipeline: context cancelled before stage %q: %w", stage.Name, ctx.Err())
		default:
		}

		if err := stage.Run(ctx, appName); err != nil {
			return fmt.Errorf("pipeline: stage %q failed: %w", stage.Name, err)
		}
	}

	return nil
}

// Stages returns the names of all registered stages.
func (p *Pipeline) Stages() []string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.Name
	}
	return names
}

func (p *Pipeline) appExists(appName string) bool {
	for _, app := range p.cfg.Apps {
		if app.Name == appName {
			return true
		}
	}
	return false
}
