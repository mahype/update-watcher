package googlechat

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
	notifier.Register("googlechat", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "googlechat",
		DisplayName: "Google Chat",
		Description: "Send notifications to Google Chat spaces via webhooks",
	})
}

// GoogleChatNotifier sends update reports via Google Chat webhooks.
type GoogleChatNotifier struct {
	webhookURL string
	threadKey  string
}

// NewFromConfig creates a GoogleChatNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("googlechat: webhook_url is required")
	}

	return &GoogleChatNotifier{
		webhookURL: webhookURL,
		threadKey:  cfg.Options.GetString("thread_key", ""),
	}, nil
}

func (g *GoogleChatNotifier) Name() string { return "googlechat" }

func (g *GoogleChatNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	message := fmt.Sprintf("*%s*\n\n%s", title, body)

	payload := map[string]interface{}{
		"text": message,
	}

	url := g.webhookURL
	if g.threadKey != "" {
		url += "&threadKey=" + g.threadKey + "&messageReplyOption=REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD"
	}

	slog.Debug("sending google chat notification")

	if err := httputil.PostJSON(url, payload); err != nil {
		return fmt.Errorf("googlechat: %w", err)
	}

	slog.Info("google chat notification sent successfully")
	return nil
}
