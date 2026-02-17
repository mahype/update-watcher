package discord

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
	notifier.Register("discord", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "discord",
		DisplayName: "Discord",
		Description: "Send notifications via Discord webhooks",
	})
}

// DiscordNotifier sends update reports via Discord webhooks.
type DiscordNotifier struct {
	webhookURL  string
	username    string
	avatarURL   string
	mentionRole string
	httpClient  *http.Client
}

// NewFromConfig creates a DiscordNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("discord: webhook_url is required")
	}

	return &DiscordNotifier{
		webhookURL:  webhookURL,
		username:    opts.GetString("username", "Update Watcher"),
		avatarURL:   opts.GetString("avatar_url", ""),
		mentionRole: opts.GetString("mention_role", ""),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (d *DiscordNotifier) Name() string { return "discord" }

func (d *DiscordNotifier) Send(hostname string, results []*checker.CheckResult) error {
	embeds := BuildEmbeds(hostname, results)

	payload := map[string]interface{}{
		"embeds": embeds,
	}

	if d.username != "" {
		payload["username"] = d.username
	}
	if d.avatarURL != "" {
		payload["avatar_url"] = d.avatarURL
	}

	// Mention role if security updates present
	if d.mentionRole != "" {
		for _, r := range results {
			if r.HasSecurityUpdates() {
				payload["content"] = fmt.Sprintf("<@&%s> Security updates found!", d.mentionRole)
				break
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: failed to marshal payload: %w", err)
	}

	slog.Debug("sending discord notification")

	resp, err := d.httpClient.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord: webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("discord notification sent successfully")
	return nil
}
