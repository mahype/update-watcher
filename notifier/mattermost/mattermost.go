package mattermost

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
	httpClient *http.Client
}

// NewFromConfig creates a MattermostNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("mattermost: webhook_url is required")
	}

	return &MattermostNotifier{
		webhookURL: webhookURL,
		channel:    opts.GetString("channel", ""),
		username:   opts.GetString("username", "Update Watcher"),
		iconURL:    opts.GetString("icon_url", ""),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (m *MattermostNotifier) Name() string { return "mattermost" }

func (m *MattermostNotifier) Send(hostname string, results []*checker.CheckResult) error {
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

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: failed to marshal payload: %w", err)
	}

	slog.Debug("sending mattermost notification")

	resp, err := m.httpClient.Post(m.webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("mattermost: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mattermost: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("mattermost notification sent successfully")
	return nil
}
