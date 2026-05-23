package health

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/patchwork-deploy/internal/config"
)

func baseConfig() *config.Config {
	return &config.Config{
		Hosts: map[string]config.Host{
			"web": {Address: "127.0.0.1", User: "deploy", KeyFile: "/tmp/key"},
		},
		Apps: map[string]config.App{
			"myapp": {Host: "web", Dir: "/srv/myapp"},
		},
	}
}

func TestCheck_UnknownApp(t *testing.T) {
	c := NewChecker(baseConfig(), 2*time.Second)
	s := c.Check("nonexistent")
	if s.Healthy {
		t.Fatal("expected unhealthy for unknown app")
	}
	if s.Message != "unknown app" {
		t.Fatalf("unexpected message: %s", s.Message)
	}
}

func TestCheck_NoHealthConfig(t *testing.T) {
	c := NewChecker(baseConfig(), 2*time.Second)
	s := c.Check("myapp")
	if !s.Healthy {
		t.Fatalf("expected healthy when no check configured, got: %s", s.Message)
	}
}

func TestCheck_HTTPCheck_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := baseConfig()
	cfg.Apps["myapp"] = config.App{Host: "web", Dir: "/srv/myapp", HealthURL: ts.URL + "/health"}

	c := NewChecker(cfg, 2*time.Second)
	s := c.Check("myapp")
	if !s.Healthy {
		t.Fatalf("expected healthy, got: %s", s.Message)
	}
}

func TestCheck_HTTPCheck_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	cfg := baseConfig()
	cfg.Apps["myapp"] = config.App{Host: "web", Dir: "/srv/myapp", HealthURL: ts.URL + "/health"}

	c := NewChecker(cfg, 2*time.Second)
	s := c.Check("myapp")
	if s.Healthy {
		t.Fatal("expected unhealthy for 503 response")
	}
}

func TestCheck_TCPCheck_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	cfg := baseConfig()
	cfg.Apps["myapp"] = config.App{Host: "web", Dir: "/srv/myapp", HealthPort: port}

	c := NewChecker(cfg, 2*time.Second)
	s := c.Check("myapp")
	if !s.Healthy {
		t.Fatalf("expected healthy tcp check, got: %s", s.Message)
	}
}

func TestCheck_TCPCheck_Failure(t *testing.T) {
	// Use a port that is not listening to simulate a failed TCP health check.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // close immediately so nothing is listening

	cfg := baseConfig()
	cfg.Apps["myapp"] = config.App{Host: "web", Dir: "/srv/myapp", HealthPort: port}

	c := NewChecker(cfg, 2*time.Second)
	s := c.Check("myapp")
	if s.Healthy {
		t.Fatal("expected unhealthy for closed tcp port")
	}
}

func TestNewChecker_DefaultTimeout(t *testing.T) {
	c := NewChecker(baseConfig(), 0)
	if c.timeout != 5*time.Second {
		t.Fatalf("expected default 5s timeout, got %v", c.timeout)
	}
}
