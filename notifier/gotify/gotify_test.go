package gotify

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "gotify",
		Enabled: true,
		Options: map[string]interface{}{
			"server_url": "https://gotify.example.com",
			"token":      "AKsjdf83jsd",
			"priority":   7,
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*GotifyNotifier)
	if nn.serverURL != "https://gotify.example.com" {
		t.Errorf("expected server URL, got %q", nn.serverURL)
	}
	if nn.token != "AKsjdf83jsd" {
		t.Errorf("expected token, got %q", nn.token)
	}
	if nn.priority != 7 {
		t.Errorf("expected priority 7, got %d", nn.priority)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "gotify",
		Enabled: true,
		Options: map[string]interface{}{
			"server_url": "https://gotify.local",
			"token":      "test-token",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*GotifyNotifier)
	if nn.priority != 5 {
		t.Errorf("expected default priority 5, got %d", nn.priority)
	}
}

func TestNewFromConfigTrailingSlash(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"server_url": "https://gotify.local/",
			"token":      "t",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*GotifyNotifier)
	if nn.serverURL != "https://gotify.local" {
		t.Errorf("expected trailing slash trimmed, got %q", nn.serverURL)
	}
}

func TestNewFromConfigMissingServerURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"token": "test-token",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing server_url")
	}
}

func TestNewFromConfigMissingToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"server_url": "https://gotify.local",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"server_url": "https://gotify.local",
			"token":      "t",
		},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "gotify" {
		t.Errorf("expected name 'gotify', got %q", n.Name())
	}
}
