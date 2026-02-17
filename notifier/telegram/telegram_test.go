package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"bot_token":            "123456:ABC",
			"chat_id":              "-1001234",
			"disable_notification": true,
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tn := n.(*TelegramNotifier)
	if tn.botToken != "123456:ABC" {
		t.Errorf("expected bot token, got %q", tn.botToken)
	}
	if tn.chatID != "-1001234" {
		t.Errorf("expected chat ID, got %q", tn.chatID)
	}
	if !tn.disableNotification {
		t.Error("expected disable_notification to be true")
	}
}

func TestNewFromConfigMissingBotToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"chat_id": "-1001234",
		},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing bot_token")
	}
}

func TestNewFromConfigMissingChatID(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"bot_token": "123456:ABC",
		},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing chat_id")
	}
}

func TestBuildHTMLMessage(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "2 packages",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13", NewVersion: "3.0.14", Type: checker.UpdateTypeSecurity},
				{Name: "curl", CurrentVersion: "8.5.0", NewVersion: "8.5.1", Type: checker.UpdateTypeRegular},
			},
		},
	}

	msg := buildHTMLMessage("test-server", results)

	if !strings.Contains(msg, "<b>") {
		t.Error("expected HTML bold tags")
	}
	if !strings.Contains(msg, "test-server") {
		t.Error("expected hostname in message")
	}
	if !strings.Contains(msg, "<code>libssl3</code>") {
		t.Error("expected package name in code tags")
	}
	if !strings.Contains(msg, "Security updates require attention") {
		t.Error("expected security footer")
	}
}

func TestSplitMessage(t *testing.T) {
	short := "hello world"
	chunks := splitMessage(short, 100)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk for short message, got %d", len(chunks))
	}

	long := strings.Repeat("line\n", 1000)
	chunks = splitMessage(long, 100)
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks for long message, got %d", len(chunks))
	}

	for _, chunk := range chunks {
		if len(chunk) > 100 {
			t.Errorf("chunk exceeds max length: %d", len(chunk))
		}
	}
}
