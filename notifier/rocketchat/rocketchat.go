package rocketchat

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
	httpClient *http.Client
}

// NewFromConfig creates a RocketChatNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("rocketchat: webhook_url is required")
	}

	return &RocketChatNotifier{
		webhookURL: webhookURL,
		channel:    opts.GetString("channel", ""),
		username:   opts.GetString("username", "Update Watcher"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (r *RocketChatNotifier) Name() string { return "rocketchat" }

func (r *RocketChatNotifier) Send(hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	text := fmt.Sprintf("### %s\n\n%s", title, body)

	payload := map[string]interface{}{
		"text":     text,
		"username": r.username,
	}
	if r.channel != "" {
		payload["channel"] = r.channel
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rocketchat: failed to marshal payload: %w", err)
	}

	slog.Debug("sending rocket.chat notification")

	resp, err := r.httpClient.Post(r.webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("rocketchat: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rocketchat: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("rocket.chat notification sent successfully")
	return nil
}
