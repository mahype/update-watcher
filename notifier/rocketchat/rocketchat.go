package rocketchat

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/httputil"
	"github.com/mahype/update-watcher/notifier"
	"github.com/mahype/update-watcher/notifier/formatting"
)

func init() {
	notifier.Register("rocketchat", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "rocketchat",
		DisplayName: "Rocket.Chat",
		Description: "Send notifications to Rocket.Chat via incoming webhooks",
	})
}

// RocketChatNotifier sends update reports via Rocket.Chat incoming webhooks.
type RocketChatNotifier struct {
	webhookURL string
	channel    string
	username   string
}

// NewFromConfig creates a RocketChatNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("rocketchat: webhook_url is required")
	}

	return &RocketChatNotifier{
		webhookURL: webhookURL,
		channel:    cfg.Options.GetString("channel", ""),
		username:   cfg.Options.GetString("username", "Update Watcher"),
	}, nil
}

func (r *RocketChatNotifier) Name() string { return "rocketchat" }

func (r *RocketChatNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	text := fmt.Sprintf("### %s\n\n%s", title, body)

	payload := map[string]interface{}{
		"text":     text,
		"username": r.username,
	}
	if r.channel != "" {
		payload["channel"] = r.channel
	}

	slog.Debug("sending rocket.chat notification")

	if err := httputil.PostJSON(r.webhookURL, payload); err != nil {
		return fmt.Errorf("rocketchat: %w", err)
	}

	slog.Info("rocket.chat notification sent successfully")
	return nil
}
