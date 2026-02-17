package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
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
	httpClient        *http.Client
}

// NewFromConfig creates a SlackNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook_url is required")
	}

	return &SlackNotifier{
		webhookURL:        webhookURL,
		mentionOnSecurity: opts.GetString("mention_on_security", ""),
		useEmoji:          opts.GetBool("use_emoji", true),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(hostname string, results []*checker.CheckResult) error {
	blocks := BuildMessage(hostname, results, s.useEmoji)

	payload := map[string]interface{}{
		"blocks": blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	slog.Debug("sending slack notification", "webhook", s.webhookURL[:30]+"...")

	resp, err := s.httpClient.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("slack notification sent successfully")
	return nil
}
