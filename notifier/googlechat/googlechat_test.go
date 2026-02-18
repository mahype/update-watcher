package googlechat

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "googlechat",
		Enabled: true,
		Options: map[string]interface{}{
			"webhook_url": "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=xxx&token=yyy",
			"thread_key":  "update-watcher",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*GoogleChatNotifier)
	if nn.webhookURL != "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=xxx&token=yyy" {
		t.Errorf("unexpected webhook URL: %q", nn.webhookURL)
	}
	if nn.threadKey != "update-watcher" {
		t.Errorf("expected thread_key 'update-watcher', got %q", nn.threadKey)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "googlechat",
		Enabled: true,
		Options: map[string]interface{}{
			"webhook_url": "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=xxx",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*GoogleChatNotifier)
	if nn.threadKey != "" {
		t.Errorf("expected empty thread_key, got %q", nn.threadKey)
	}
}

func TestNewFromConfigMissingWebhookURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "googlechat",
		Enabled: true,
		Options: map[string]interface{}{},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing webhook_url")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"webhook_url": "https://example.com",
		},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "googlechat" {
		t.Errorf("expected name 'googlechat', got %q", n.Name())
	}
}
