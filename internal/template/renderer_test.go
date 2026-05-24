package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/template"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Addr: "10.0.0.1"},
		},
		Global: config.GlobalConfig{
			Env: map[string]string{
				"REGION": "us-east",
			},
		},
		Apps: map[string]config.App{
			"api": {
				Host: "web",
				Dir:  "/srv/api",
				Env:  map[string]string{"PORT": "8080"},
			},
		},
	}
}

func TestNewRenderer_NotNil(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestRenderString_UnknownApp(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	_, err := r.RenderString("ghost", "hello")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestRenderString_InjectsAppVars(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	out, err := r.RenderString("api", "port={{.PORT}} dir={{.Dir}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "port=8080 dir=/srv/api" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRenderString_GlobalVarAvailable(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	out, err := r.RenderString("api", "region={{.REGION}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "region=us-east" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRenderString_MissingKeyErrors(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	_, err := r.RenderString("api", "{{.UNDEFINED_VAR}}")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestRenderFile_ReadsAndRenders(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "deploy.conf.tmpl")
	content := "app={{.AppName}} host={{.Host}}"
	if err := os.WriteFile(tmplPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	r := template.NewRenderer(baseConfig())
	out, err := r.RenderFile("api", tmplPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "app=api host=web" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRenderFile_MissingFile(t *testing.T) {
	r := template.NewRenderer(baseConfig())
	_, err := r.RenderFile("api", "/nonexistent/path/deploy.tmpl")
	if err == nil {
		t.Fatal("expected error for missing template file")
	}
}
