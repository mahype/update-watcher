package pagerduty

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pagerduty",
		Enabled: true,
		Options: map[string]interface{}{
			"routing_key": "R0123456789ABCDEF",
			"severity":    "error",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PagerDutyNotifier)
	if nn.routingKey != "R0123456789ABCDEF" {
		t.Errorf("expected routing_key, got %q", nn.routingKey)
	}
	if nn.severity != "error" {
		t.Errorf("expected severity 'error', got %q", nn.severity)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"routing_key": "test-key",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PagerDutyNotifier)
	if nn.severity != "warning" {
		t.Errorf("expected default severity 'warning', got %q", nn.severity)
	}
}

func TestNewFromConfigMissingRoutingKey(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing routing_key")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{"routing_key": "k"},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "pagerduty" {
		t.Errorf("expected name 'pagerduty', got %q", n.Name())
	}
}
