package ntfy

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "ntfy",
		Enabled: true,
		Options: map[string]interface{}{
			"topic":      "test-topic",
			"server_url": "https://ntfy.example.com",
			"token":      "tk_test123",
			"priority":   "high",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*NtfyNotifier)
	if nn.topic != "test-topic" {
		t.Errorf("expected topic 'test-topic', got %q", nn.topic)
	}
	if nn.serverURL != "https://ntfy.example.com" {
		t.Errorf("expected custom server URL, got %q", nn.serverURL)
	}
	if nn.token != "tk_test123" {
		t.Errorf("expected token, got %q", nn.token)
	}
	if nn.priority != "high" {
		t.Errorf("expected priority 'high', got %q", nn.priority)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "ntfy",
		Enabled: true,
		Options: map[string]interface{}{
			"topic": "my-topic",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*NtfyNotifier)
	if nn.serverURL != "https://ntfy.sh" {
		t.Errorf("expected default server URL, got %q", nn.serverURL)
	}
	if nn.priority != "default" {
		t.Errorf("expected default priority, got %q", nn.priority)
	}
}

func TestNewFromConfigMissingTopic(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "ntfy",
		Enabled: true,
		Options: map[string]interface{}{},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{"topic": "t"},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "ntfy" {
		t.Errorf("expected name 'ntfy', got %q", n.Name())
	}
}
