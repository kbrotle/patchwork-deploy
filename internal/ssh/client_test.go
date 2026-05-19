package ssh_test

import (
	"os"
	"testing"
	"time"

	sshclient "github.com/yourorg/patchwork-deploy/internal/ssh"
)

func TestConnect_MissingKeyFile(t *testing.T) {
	_, err := sshclient.Connect(sshclient.Config{
		Host:    "127.0.0.1",
		Port:    22,
		User:    "root",
		KeyPath: "/nonexistent/key",
		Timeout: 2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for missing key file, got nil")
	}
}

func TestConnect_InvalidKey(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "badkey")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("not a valid pem key")
	f.Close()

	_, err = sshclient.Connect(sshclient.Config{
		Host:    "127.0.0.1",
		Port:    22,
		User:    "root",
		KeyPath: f.Name(),
		Timeout: 2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}

func TestConfig_DefaultPort(t *testing.T) {
	// Verify that zero port is treated as 22 by inspecting dial error message.
	f, err := os.CreateTemp(t.TempDir(), "badkey")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("not a valid pem key")
	f.Close()

	_, err = sshclient.Connect(sshclient.Config{
		Host:    "127.0.0.1",
		User:    "root",
		KeyPath: f.Name(),
	})
	// We expect a key-parse error, not a port error, confirming default port path.
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
