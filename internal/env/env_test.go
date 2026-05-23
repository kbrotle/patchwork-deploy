package env_test

import (
	"os"
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/env"
)

func baseConfig() *config.Config {
	return &config.Config{
		GlobalEnv: map[string]string{
			"REGION": "us-east-1",
			"LOG_LEVEL": "info",
		},
		Apps: []config.App{
			{
				Name: "api",
				Env: map[string]string{
					"PORT": "8080",
					"LOG_LEVEL": "debug", // overrides global
				},
			},
			{
				Name: "worker",
				Env:  map[string]string{},
			},
		},
	}
}

func TestResolve_UnknownApp(t *testing.T) {
	r := env.NewResolver(baseConfig())
	_, err := r.Resolve("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app, got nil")
	}
}

func TestResolve_MergesGlobalAndApp(t *testing.T) {
	r := env.NewResolver(baseConfig())
	got, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["REGION"] != "us-east-1" {
		t.Errorf("expected REGION=us-east-1, got %q", got["REGION"])
	}
	if got["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", got["PORT"])
	}
}

func TestResolve_AppOverridesGlobal(t *testing.T) {
	r := env.NewResolver(baseConfig())
	got, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["LOG_LEVEL"] != "debug" {
		t.Errorf("expected LOG_LEVEL=debug (app override), got %q", got["LOG_LEVEL"])
	}
}

func TestResolve_ExpandsEnvVars(t *testing.T) {
	os.Setenv("SECRET_KEY", "s3cr3t")
	t.Cleanup(func() { os.Unsetenv("SECRET_KEY") })

	cfg := baseConfig()
	cfg.Apps[0].Env["APP_SECRET"] = "$SECRET_KEY"

	r := env.NewResolver(cfg)
	got, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["APP_SECRET"] != "s3cr3t" {
		t.Errorf("expected APP_SECRET=s3cr3t, got %q", got["APP_SECRET"])
	}
}

func TestToSlice_Format(t *testing.T) {
	m := map[string]string{"FOO": "bar"}
	slice := env.ToSlice(m)
	if len(slice) != 1 || slice[0] != "FOO=bar" {
		t.Errorf("unexpected slice: %v", slice)
	}
}
