package slack

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/httputil"
	"github.com/mahype/update-watcher/notifier"
)

func init() {
	notifier.Register("slack", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "slack",
		DisplayName: "Slack",
		Description: "Send notifications via Slack webhooks",
	})
}

// SlackNotifier sends update reports via Slack webhooks.
type SlackNotifier struct {
	webhookURL        string
	mentionOnSecurity string
	useEmoji          bool
}

// NewFromConfig creates a SlackNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook_url is required")
	}

	return &SlackNotifier{
		webhookURL:        webhookURL,
		mentionOnSecurity: cfg.Options.GetString("mention_on_security", ""),
		useEmoji:          cfg.Options.GetBool("use_emoji", true),
	}, nil
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	blocks := BuildMessage(hostname, results, s.useEmoji)

	payload := map[string]interface{}{
		"blocks": blocks,
	}

	slog.Debug("sending slack notification", "webhook", s.webhookURL[:30]+"...")

	if err := httputil.PostJSON(s.webhookURL, payload); err != nil {
		return fmt.Errorf("slack: %w", err)
	}

	slog.Info("slack notification sent successfully")
	return nil
}
