package deploy_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/deploy"
	"github.com/yourorg/patchwork-deploy/internal/retry"
	"github.com/yourorg/patchwork-deploy/internal/ssh"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Address: "127.0.0.1", User: "deploy", KeyFile: "/nonexistent/key"},
		},
		Apps: map[string]config.App{
			"api": {Host: "web", Dir: "/local/api", RemoteDir: "/srv/api"},
		},
	}
}

func fastRetryer() *retry.Retryer {
	return retry.NewRetryer(retry.Policy{MaxAttempts: 1, Delay: time.Millisecond, Backoff: 1.0})
}

func TestNewRunner_NotNil(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	r := deploy.NewRunner(cfg, pool)
	if r == nil {
		t.Fatal("expected non-nil Runner")
	}
}

func TestDeploy_UnknownApp(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	r := deploy.NewRunner(cfg, pool).WithRetryer(fastRetryer())
	err := r.Deploy(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestDeploy_SSHFailure(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	r := deploy.NewRunner(cfg, pool).WithRetryer(fastRetryer())
	err := r.Deploy(context.Background(), "api")
	if err == nil {
		t.Fatal("expected SSH failure error")
	}
}

func TestDeploy_CancelledContext(t *testing.T) {
	cfg := baseConfig()
	pool := ssh.NewPool(cfg)
	r := deploy.NewRunner(cfg, pool).WithRetryer(
		retry.NewRetryer(retry.Policy{MaxAttempts: 3, Delay: 50 * time.Millisecond, Backoff: 1.0}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Deploy(ctx, "api")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
