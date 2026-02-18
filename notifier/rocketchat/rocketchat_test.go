package rocketchat

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "rocketchat",
		Enabled: true,
		Options: map[string]interface{}{
			"webhook_url": "https://rocket.example.com/hooks/xxx",
			"channel":     "#general",
			"username":    "Bot",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*RocketChatNotifier)
	if nn.webhookURL != "https://rocket.example.com/hooks/xxx" {
		t.Errorf("unexpected webhook URL: %q", nn.webhookURL)
	}
	if nn.channel != "#general" {
		t.Errorf("expected channel '#general', got %q", nn.channel)
	}
	if nn.username != "Bot" {
		t.Errorf("expected username 'Bot', got %q", nn.username)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"webhook_url": "https://rc.local/hooks/abc",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*RocketChatNotifier)
	if nn.username != "Update Watcher" {
		t.Errorf("expected default username, got %q", nn.username)
	}
}

func TestNewFromConfigMissingWebhookURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing webhook_url")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{"webhook_url": "https://rc.local/hooks/x"},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "rocketchat" {
		t.Errorf("expected name 'rocketchat', got %q", n.Name())
	}
}
