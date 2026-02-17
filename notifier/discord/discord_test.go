package discord

import (
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"webhook_url":  "https://discord.com/api/webhooks/test",
			"username":     "TestBot",
			"mention_role": "12345",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dn := n.(*DiscordNotifier)
	if dn.username != "TestBot" {
		t.Errorf("expected username 'TestBot', got %q", dn.username)
	}
	if dn.mentionRole != "12345" {
		t.Errorf("expected mention role '12345', got %q", dn.mentionRole)
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

func TestBuildEmbeds(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "2 packages",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13", NewVersion: "3.0.14", Type: checker.UpdateTypeSecurity, Priority: checker.PriorityHigh},
				{Name: "curl", CurrentVersion: "8.5.0", NewVersion: "8.5.1", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
			},
		},
	}

	embeds := BuildEmbeds("test-server", results)

	if len(embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embeds))
	}

	embed := embeds[0]
	if embed.Color != colorRed {
		t.Errorf("expected red color for security updates, got %d", embed.Color)
	}
	if len(embed.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(embed.Fields))
	}
}

func TestBuildEmbedsNoUpdates(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "all up to date",
			CheckedAt:   time.Now(),
		},
	}

	embeds := BuildEmbeds("test-server", results)
	if embeds[0].Color != colorGreen {
		t.Errorf("expected green color for no updates, got %d", embeds[0].Color)
	}
}
