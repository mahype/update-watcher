package mattermost

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "mattermost",
		Enabled: true,
		Options: map[string]interface{}{
			"webhook_url": "https://mattermost.example.com/hooks/xxx",
			"channel":     "town-square",
			"username":    "Bot",
			"icon_url":    "https://example.com/icon.png",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*MattermostNotifier)
	if nn.webhookURL != "https://mattermost.example.com/hooks/xxx" {
		t.Errorf("unexpected webhook URL: %q", nn.webhookURL)
	}
	if nn.channel != "town-square" {
		t.Errorf("expected channel 'town-square', got %q", nn.channel)
	}
	if nn.username != "Bot" {
		t.Errorf("expected username 'Bot', got %q", nn.username)
	}
	if nn.iconURL != "https://example.com/icon.png" {
		t.Errorf("expected icon_url, got %q", nn.iconURL)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"webhook_url": "https://mm.local/hooks/abc",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*MattermostNotifier)
	if nn.username != "Update Watcher" {
		t.Errorf("expected default username, got %q", nn.username)
	}
	if nn.channel != "" {
		t.Errorf("expected empty channel, got %q", nn.channel)
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
		Options: map[string]interface{}{"webhook_url": "https://mm.local/hooks/x"},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "mattermost" {
		t.Errorf("expected name 'mattermost', got %q", n.Name())
	}
}
