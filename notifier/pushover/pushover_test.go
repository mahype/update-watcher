package pushover

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pushover",
		Enabled: true,
		Options: map[string]interface{}{
			"app_token": "azGDORePK8gMaC0QOYAMyEEuzJnyUi",
			"user_key":  "uQiRzpo4DXghDmr9QzzfQu27cmVRsG",
			"device":    "iphone",
			"priority":  1,
			"sound":     "cosmic",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PushoverNotifier)
	if nn.appToken != "azGDORePK8gMaC0QOYAMyEEuzJnyUi" {
		t.Errorf("expected app_token, got %q", nn.appToken)
	}
	if nn.userKey != "uQiRzpo4DXghDmr9QzzfQu27cmVRsG" {
		t.Errorf("expected user_key, got %q", nn.userKey)
	}
	if nn.device != "iphone" {
		t.Errorf("expected device 'iphone', got %q", nn.device)
	}
	if nn.priority != 1 {
		t.Errorf("expected priority 1, got %d", nn.priority)
	}
	if nn.sound != "cosmic" {
		t.Errorf("expected sound 'cosmic', got %q", nn.sound)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pushover",
		Enabled: true,
		Options: map[string]interface{}{
			"app_token": "test-token",
			"user_key":  "test-user",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PushoverNotifier)
	if nn.priority != 0 {
		t.Errorf("expected default priority 0, got %d", nn.priority)
	}
	if nn.device != "" {
		t.Errorf("expected empty device, got %q", nn.device)
	}
	if nn.sound != "" {
		t.Errorf("expected empty sound, got %q", nn.sound)
	}
}

func TestNewFromConfigMissingAppToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pushover",
		Enabled: true,
		Options: map[string]interface{}{
			"user_key": "test-user",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing app_token")
	}
}

func TestNewFromConfigMissingUserKey(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pushover",
		Enabled: true,
		Options: map[string]interface{}{
			"app_token": "test-token",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing user_key")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"app_token": "t",
			"user_key":  "u",
		},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "pushover" {
		t.Errorf("expected name 'pushover', got %q", n.Name())
	}
}
