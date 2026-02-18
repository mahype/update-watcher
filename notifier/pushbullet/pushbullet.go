package pushbullet

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

const apiURL = "https://api.pushbullet.com/v2/pushes"

func init() {
	notifier.Register("pushbullet", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "pushbullet",
		DisplayName: "Pushbullet",
		Description: "Push notifications to all devices via Pushbullet",
	})
}

// PushbulletNotifier sends update reports via Pushbullet.
type PushbulletNotifier struct {
	accessToken string
	deviceIden  string
	channelTag  string
}

// NewFromConfig creates a PushbulletNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	accessToken := cfg.Options.GetString("access_token", "")
	if accessToken == "" {
		return nil, fmt.Errorf("pushbullet: access_token is required")
	}

	return &PushbulletNotifier{
		accessToken: accessToken,
		deviceIden:  cfg.Options.GetString("device_iden", ""),
		channelTag:  cfg.Options.GetString("channel_tag", ""),
	}, nil
}

func (p *PushbulletNotifier) Name() string { return "pushbullet" }

func (p *PushbulletNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	body := formatting.BuildPlainTextMessage(hostname, results)
	summary := formatting.SummarizeResults(results)

	title := fmt.Sprintf("\U0001f504 Update Report: %s", hostname)
	if summary.SecurityCount > 0 {
		title = fmt.Sprintf("\u26a0\ufe0f Update Report: %s (%d security)", hostname, summary.SecurityCount)
	}

	payload := map[string]interface{}{
		"type":  "note",
		"title": title,
		"body":  body,
	}
	if p.deviceIden != "" {
		payload["device_iden"] = p.deviceIden
	}
	if p.channelTag != "" {
		payload["channel_tag"] = p.channelTag
	}

	slog.Debug("sending pushbullet notification")

	headers := map[string]string{
		"Access-Token": p.accessToken,
	}
	if err := httputil.PostJSONWithHeaders(apiURL, payload, headers); err != nil {
		return fmt.Errorf("pushbullet: %w", err)
	}

	slog.Info("pushbullet notification sent successfully")
	return nil
}
