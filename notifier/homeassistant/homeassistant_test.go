package homeassistant

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "homeassistant",
		Enabled: true,
		Options: map[string]interface{}{
			"url":     "http://homeassistant.local:8123",
			"token":   "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.test",
			"service": "mobile_app_iphone",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*HomeAssistantNotifier)
	if nn.url != "http://homeassistant.local:8123" {
		t.Errorf("expected url, got %q", nn.url)
	}
	if nn.token != "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.test" {
		t.Errorf("expected token, got %q", nn.token)
	}
	if nn.service != "mobile_app_iphone" {
		t.Errorf("expected service 'mobile_app_iphone', got %q", nn.service)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "homeassistant",
		Enabled: true,
		Options: map[string]interface{}{
			"url":   "http://homeassistant.local:8123",
			"token": "test-token",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*HomeAssistantNotifier)
	if nn.service != "notify" {
		t.Errorf("expected default service 'notify', got %q", nn.service)
	}
}

func TestNewFromConfigTrailingSlash(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"url":   "http://homeassistant.local:8123/",
			"token": "t",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*HomeAssistantNotifier)
	if nn.url != "http://homeassistant.local:8123" {
		t.Errorf("expected trailing slash trimmed, got %q", nn.url)
	}
}

func TestNewFromConfigMissingURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"token": "test-token",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestNewFromConfigMissingToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"url": "http://homeassistant.local:8123",
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
			"url":   "http://homeassistant.local:8123",
			"token": "t",
		},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "homeassistant" {
		t.Errorf("expected name 'homeassistant', got %q", n.Name())
	}
}
