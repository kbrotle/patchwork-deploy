package secret_test

import (
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/secret"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: map[string]config.App{
			"api": {
				Secrets: map[string]string{
					"DB_PASS": "literal-password",
					"API_KEY": "env:TEST_API_KEY",
				},
			},
			"worker": {
				Secrets: map[string]string{},
			},
		},
	}
}

func TestResolve_UnknownApp(t *testing.T) {
	r := secret.NewResolver(baseConfig())
	_, err := r.Resolve("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestResolve_LiteralValue(t *testing.T) {
	r := secret.NewResolver(baseConfig())
	secrets, err := r.Resolve("api")
	if err != nil {
		// env var may not be set; only check literal
		t.Skipf("skipping: %v", err)
	}
	if secrets["DB_PASS"] != "literal-password" {
		t.Errorf("expected literal-password, got %q", secrets["DB_PASS"])
	}
}

func TestResolve_EnvVar(t *testing.T) {
	t.Setenv("TEST_API_KEY", "supersecret")
	r := secret.NewResolver(baseConfig())
	secrets, err := r.Resolve("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets["API_KEY"] != "supersecret" {
		t.Errorf("expected supersecret, got %q", secrets["API_KEY"])
	}
}

func TestResolve_MissingEnvVar(t *testing.T) {
	// Ensure the env var is unset
	t.Setenv("TEST_API_KEY", "")
	cfg := &config.Config{
		Apps: map[string]config.App{
			"api": {
				Secrets: map[string]string{
					"TOKEN": "env:DEFINITELY_NOT_SET_XYZ123",
				},
			},
		},
	}
	r := secret.NewResolver(cfg)
	_, err := r.Resolve("api")
	if err == nil {
		t.Fatal("expected error for missing env var")
	}
}

func TestToEnvSlice_EmptySecrets(t *testing.T) {
	r := secret.NewResolver(baseConfig())
	slice, err := r.ToEnvSlice("worker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slice) != 0 {
		t.Errorf("expected empty slice, got %v", slice)
	}
}

// TestToEnvSlice_Format verifies that resolved secrets are formatted as KEY=VALUE
// entries suitable for use as process environment variables.
func TestToEnvSlice_Format(t *testing.T) {
	t.Setenv("TEST_API_KEY", "mykey")
	r := secret.NewResolver(baseConfig())
	slice, err := r.ToEnvSlice("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"DB_PASS": "literal-password",
		"API_KEY": "mykey",
	}
	if len(slice) != len(want) {
		t.Fatalf("expected %d entries, got %d: %v", len(want), len(slice), slice)
	}
	for _, entry := range slice {
		for k, v := range want {
			if entry == k+"="+v {
				delete(want, k)
				break
			}
		}
	}
	if len(want) != 0 {
		t.Errorf("missing or incorrect entries for keys: %v", want)
	}
}

func TestNewResolver_NotNil(t *testing.T) {
	r := secret.NewResolver(baseConfig())
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}
