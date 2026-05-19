package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds parameters for an SSH connection.
type Config struct {
	Host    string
	Port    int
	User    string
	KeyFile string
}

// Client wraps an active SSH connection.
type Client struct {
	conn *ssh.Client
}

// Connect establishes an SSH connection using the given config.
func Connect(cfg Config) (*Client, error) {
	key, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", cfg.KeyFile, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	port := cfg.Port
	if port == 0 {
		port = 22
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))
	conn, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	return &Client{conn: conn}, nil
}

// Run executes a remote command and returns an error if it fails.
func (c *Client) Run(cmd string) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()
	return sess.Run(cmd)
}

// Close terminates the underlying SSH connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
