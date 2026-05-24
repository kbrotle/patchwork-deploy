package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/pipeline"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "web", Host: "h1", Dir: "/app/web"},
		},
		Hosts: []config.Host{
			{Name: "h1", Addr: "127.0.0.1", User: "deploy", KeyFile: "/id_rsa"},
		},
	}
}

func noop(_ context.Context, _ string) error { return nil }

func TestNewPipeline_NotNil(t *testing.T) {
	p := pipeline.NewPipeline(baseConfig(), nil)
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
}

func TestRun_UnknownApp(t *testing.T) {
	p := pipeline.NewPipeline(baseConfig(), []pipeline.Stage{{Name: "s1", Run: noop}})
	err := p.Run(context.Background(), "ghost")
	if err == nil || !containsStr(err.Error(), "unknown app") {
		t.Fatalf("expected unknown app error, got %v", err)
	}
}

func TestRun_ExecutesAllStages(t *testing.T) {
	executed := []string{}
	makeStage := func(name string) pipeline.Stage {
		return pipeline.Stage{
			Name: name,
			Run: func(_ context.Context, _ string) error {
				executed = append(executed, name)
				return nil
			},
		}
	}
	p := pipeline.NewPipeline(baseConfig(), []pipeline.Stage{
		makeStage("build"), makeStage("upload"), makeStage("restart"),
	})
	if err := p.Run(context.Background(), "web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 3 {
		t.Fatalf("expected 3 stages executed, got %d", len(executed))
	}
}

func TestRun_StopsOnFirstError(t *testing.T) {
	ran := 0
	boom := errors.New("boom")
	stages := []pipeline.Stage{
		{Name: "ok", Run: func(_ context.Context, _ string) error { ran++; return nil }},
		{Name: "fail", Run: func(_ context.Context, _ string) error { ran++; return boom }},
		{Name: "skip", Run: func(_ context.Context, _ string) error { ran++; return nil }},
	}
	p := pipeline.NewPipeline(baseConfig(), stages)
	err := p.Run(context.Background(), "web")
	if !errors.Is(err, boom) {
		t.Fatalf("expected boom error, got %v", err)
	}
	if ran != 2 {
		t.Fatalf("expected 2 stages ran, got %d", ran)
	}
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := pipeline.NewPipeline(baseConfig(), []pipeline.Stage{{Name: "s", Run: noop}})
	err := p.Run(ctx, "web")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestStages_ReturnsNames(t *testing.T) {
	p := pipeline.NewPipeline(baseConfig(), []pipeline.Stage{
		{Name: "a", Run: noop},
		{Name: "b", Run: noop},
	})
	names := p.Stages()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("unexpected stage names: %v", names)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
