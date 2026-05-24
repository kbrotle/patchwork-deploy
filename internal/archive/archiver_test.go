package archive_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/archive"
	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/ssh"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Addr: "127.0.0.1", User: "deploy", KeyFile: "/nonexistent/key"},
		},
		Apps: map[string]config.App{
			"api": {Host: "web", Dir: "/srv/api"},
		},
	}
}

func TestNewArchiver_NotNil(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool()
	a := archive.NewArchiver(cfg, pool)
	if a == nil {
		t.Fatal("expected non-nil Archiver")
	}
}

func TestArchive_UnknownApp(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool()
	a := archive.NewArchiver(cfg, pool)

	_, err := a.Archive("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestList_UnknownApp(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool()
	a := archive.NewArchiver(cfg, pool)

	_, err := a.List("ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestArchive_SSHFailure(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool()
	a := archive.NewArchiver(cfg, pool)

	// Key file does not exist, so SSH connection must fail.
	_, err := a.Archive("api")
	if err == nil {
		t.Fatal("expected SSH error")
	}
}

func TestList_SSHFailure(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool()
	a := archive.NewArchiver(cfg, pool)

	_, err := a.List("api")
	if err == nil {
		t.Fatal("expected SSH error")
	}
}
