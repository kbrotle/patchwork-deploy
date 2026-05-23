package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/patchwork-deploy/internal/config"
)

// Event represents a deployment lifecycle event.
type Event struct {
	App    string
	Status string // "started", "success", "failure"
	Msg    string
}

// Notifier sends deployment event notifications.
type Notifier struct {
	cfg    *config.Config
	client *http.Client
}

// NewNotifier creates a Notifier backed by the provided config.
func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{
		cfg: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches an Event according to the global notify config.
// It is a no-op when no webhook URL is configured.
func (n *Notifier) Send(evt Event) error {
	if n.cfg.Notify.WebhookURL == "" {
		return nil
	}

	payload := map[string]string{
		"app":    evt.App,
		"status": evt.Status,
		"msg":    evt.Msg,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.cfg.Notify.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned status %d", resp.StatusCode)
	}

	return nil
}
