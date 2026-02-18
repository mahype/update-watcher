package mattermost

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
	notifier.Register("mattermost", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "mattermost",
		DisplayName: "Mattermost",
		Description: "Send notifications to Mattermost via incoming webhooks",
	})
}

// MattermostNotifier sends update reports via Mattermost incoming webhooks.
type MattermostNotifier struct {
	webhookURL string
	channel    string
	username   string
	iconURL    string
}

// NewFromConfig creates a MattermostNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("mattermost: webhook_url is required")
	}

	return &MattermostNotifier{
		webhookURL: webhookURL,
		channel:    cfg.Options.GetString("channel", ""),
		username:   cfg.Options.GetString("username", "Update Watcher"),
		iconURL:    cfg.Options.GetString("icon_url", ""),
	}, nil
}

func (m *MattermostNotifier) Name() string { return "mattermost" }

func (m *MattermostNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	text := fmt.Sprintf("### %s\n\n%s", title, body)

	payload := map[string]interface{}{
		"text":     text,
		"username": m.username,
	}
	if m.channel != "" {
		payload["channel"] = m.channel
	}
	if m.iconURL != "" {
		payload["icon_url"] = m.iconURL
	}

	slog.Debug("sending mattermost notification")

	if err := httputil.PostJSON(m.webhookURL, payload); err != nil {
		return fmt.Errorf("mattermost: %w", err)
	}

	slog.Info("mattermost notification sent successfully")
	return nil
}
