package teams

import (
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"webhook_url": "https://prod.workflows.office.com/test",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tn := n.(*TeamsNotifier)
	if tn.webhookURL != "https://prod.workflows.office.com/test" {
		t.Errorf("expected webhook URL, got %q", tn.webhookURL)
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

func TestBuildAdaptiveCard(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "3 packages (1 security)",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13", NewVersion: "3.0.14", Type: checker.UpdateTypeSecurity},
				{Name: "curl", CurrentVersion: "8.5.0", NewVersion: "8.5.1", Type: checker.UpdateTypeRegular},
			},
		},
	}

	card := BuildAdaptiveCard("test-server", results)

	if card["type"] != "AdaptiveCard" {
		t.Errorf("expected AdaptiveCard type, got %v", card["type"])
	}
	if card["version"] != "1.5" {
		t.Errorf("expected version 1.5, got %v", card["version"])
	}

	body, ok := card["body"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected body to be a slice of maps")
	}

	// Should have: header + context + factset + container + security footer = 5
	if len(body) < 4 {
		t.Errorf("expected at least 4 body elements, got %d", len(body))
	}
}
