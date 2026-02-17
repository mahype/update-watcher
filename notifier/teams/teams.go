package teams

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
	httpClient *http.Client
}

// NewFromConfig creates a TeamsNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("teams: webhook_url is required")
	}

	return &TeamsNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (t *TeamsNotifier) Name() string { return "teams" }

func (t *TeamsNotifier) Send(hostname string, results []*checker.CheckResult) error {
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

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: failed to marshal payload: %w", err)
	}

	slog.Debug("sending teams notification")

	resp, err := t.httpClient.Post(t.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("teams: webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("teams notification sent successfully")
	return nil
}
