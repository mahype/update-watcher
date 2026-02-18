package googlechat

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
	httpClient *http.Client
}

// NewFromConfig creates a GoogleChatNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	webhookURL := opts.GetString("webhook_url", "")
	if webhookURL == "" {
		return nil, fmt.Errorf("googlechat: webhook_url is required")
	}

	return &GoogleChatNotifier{
		webhookURL: webhookURL,
		threadKey:  opts.GetString("thread_key", ""),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (g *GoogleChatNotifier) Name() string { return "googlechat" }

func (g *GoogleChatNotifier) Send(hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	message := fmt.Sprintf("*%s*\n\n%s", title, body)

	payload := map[string]interface{}{
		"text": message,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlechat: failed to marshal payload: %w", err)
	}

	url := g.webhookURL
	if g.threadKey != "" {
		url += "&threadKey=" + g.threadKey + "&messageReplyOption=REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD"
	}

	slog.Debug("sending google chat notification")

	resp, err := g.httpClient.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("googlechat: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("googlechat: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("google chat notification sent successfully")
	return nil
}
