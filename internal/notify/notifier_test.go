package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork-deploy/internal/config"
	"github.com/user/patchwork-deploy/internal/notify"
)

func baseConfig(webhookURL string) *config.Config {
	return &config.Config{
		Notify: config.NotifyConfig{
			WebhookURL: webhookURL,
		},
	}
}

func TestSend_NoWebhook_IsNoop(t *testing.T) {
	n := notify.NewNotifier(baseConfig(""))
	if err := n.Send(notify.Event{App: "api", Status: "success"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSend_WebhookReceivesPayload(t *testing.T) {
	var got map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := notify.NewNotifier(baseConfig(srv.URL))
	evt := notify.Event{App: "worker", Status: "failure", Msg: "exit 1"}

	if err := n.Send(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["app"] != "worker" || got["status"] != "failure" || got["msg"] != "exit 1" {
		t.Errorf("payload mismatch: %v", got)
	}
}

func TestSend_WebhookErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := notify.NewNotifier(baseConfig(srv.URL))
	if err := n.Send(notify.Event{App: "api", Status: "started"}); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestNewNotifier_NotNil(t *testing.T) {
	if notify.NewNotifier(baseConfig("")) == nil {
		t.Fatal("expected non-nil notifier")
	}
}
