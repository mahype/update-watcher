package teams

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
	notifier.Register("teams", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "teams",
		DisplayName: "Microsoft Teams",
		Description: "Send notifications via Teams Workflow webhooks (Adaptive Cards)",
	})
}

// TeamsNotifier sends update reports via Microsoft Teams webhooks.
type TeamsNotifier struct {
	webhookURL string
}

// NewFromConfig creates a TeamsNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	webhookURL := cfg.Options.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("teams: webhook_url is required")
	}

	return &TeamsNotifier{
		webhookURL: webhookURL,
	}, nil
}

func (t *TeamsNotifier) Name() string { return "teams" }

func (t *TeamsNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	card := BuildAdaptiveCard(hostname, results)

	payload := map[string]interface{}{
		"type": "message",
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"contentUrl":  nil,
				"content":     card,
			},
		},
	}

	slog.Debug("sending teams notification")

	if err := httputil.PostJSON(t.webhookURL, payload); err != nil {
		return fmt.Errorf("teams: %w", err)
	}

	slog.Info("teams notification sent successfully")
	return nil
}
