package discord

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
}

// NewFromConfig creates a DiscordNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("discord: webhook_url is required")
	}

	return &DiscordNotifier{
		webhookURL:  webhookURL,
		username:    cfg.Options.GetString("username", "Update Watcher"),
		avatarURL:   cfg.Options.GetString("avatar_url", ""),
		mentionRole: cfg.Options.GetString("mention_role", ""),
	}, nil
}

func (d *DiscordNotifier) Name() string { return "discord" }

func (d *DiscordNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
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

	slog.Debug("sending discord notification")

	if err := httputil.PostJSON(d.webhookURL, payload); err != nil {
		return fmt.Errorf("discord: %w", err)
	}

	slog.Info("discord notification sent successfully")
	return nil
}
