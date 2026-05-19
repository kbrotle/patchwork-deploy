package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "deploy.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
hosts:
  - name: web
    address: 192.168.1.10
    user: deploy
    port: 22
    identity_file: ~/.ssh/id_rsa
apps:
  - name: api
    host: web
    dir: /srv/api
    steps:
      - git pull
      - make build
      - systemctl restart api
`
	cfg, err := Load(writeTempConfig(t, content))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cfg.Hosts) != 1 || cfg.Hosts[0].Name != "web" {
		t.Errorf("unexpected hosts: %+v", cfg.Hosts)
	}
	if len(cfg.Apps) != 1 || cfg.Apps[0].Name != "api" {
		t.Errorf("unexpected apps: %+v", cfg.Apps)
	}
	if len(cfg.Apps[0].Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(cfg.Apps[0].Steps))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/deploy.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_UnknownHostReference(t *testing.T) {
	content := `
hosts:
  - name: web
    address: 1.2.3.4
    user: deploy
apps:
  - name: api
    host: staging
    dir: /srv/api
`
	_, err := Load(writeTempConfig(t, content))
	if err == nil {
		t.Fatal("expected validation error for unknown host reference")
	}
}

func TestLoad_MissingAppDir(t *testing.T) {
	content := `
hosts:
  - name: web
    address: 1.2.3.4
    user: deploy
apps:
  - name: api
    host: web
`
	_, err := Load(writeTempConfig(t, content))
	if err == nil {
		t.Fatal("expected validation error for missing app dir")
	}
}
