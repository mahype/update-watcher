package webhook

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "webhook",
		Enabled: true,
		Options: map[string]interface{}{
			"url":         "https://example.com/hook",
			"method":      "PUT",
			"auth_header": "Bearer secret",
			"headers": map[string]interface{}{
				"X-Custom": "value",
			},
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wn := n.(*WebhookNotifier)
	if wn.url != "https://example.com/hook" {
		t.Errorf("expected url, got %q", wn.url)
	}
	if wn.method != "PUT" {
		t.Errorf("expected PUT, got %q", wn.method)
	}
	if wn.authHeader != "Bearer secret" {
		t.Errorf("expected auth header, got %q", wn.authHeader)
	}
	if wn.headers["X-Custom"] != "value" {
		t.Errorf("expected custom header")
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"url": "https://example.com/hook",
		},
	}

	n, _ := NewFromConfig(cfg)
	wn := n.(*WebhookNotifier)
	if wn.method != "POST" {
		t.Errorf("expected default method POST, got %q", wn.method)
	}
	if wn.contentType != "application/json" {
		t.Errorf("expected default content type, got %q", wn.contentType)
	}
}

func TestNewFromConfigMissingURL(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}
