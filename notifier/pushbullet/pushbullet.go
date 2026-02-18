package pushbullet

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
	httpClient  *http.Client
}

// NewFromConfig creates a PushbulletNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	accessToken := opts.GetString("access_token", "")
	if accessToken == "" {
		return nil, fmt.Errorf("pushbullet: access_token is required")
	}

	return &PushbulletNotifier{
		accessToken: accessToken,
		deviceIden:  opts.GetString("device_iden", ""),
		channelTag:  opts.GetString("channel_tag", ""),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *PushbulletNotifier) Name() string { return "pushbullet" }

func (p *PushbulletNotifier) Send(hostname string, results []*checker.CheckResult) error {
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

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pushbullet: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("pushbullet: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", p.accessToken)

	slog.Debug("sending pushbullet notification")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pushbullet: failed to send push: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pushbullet: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("pushbullet notification sent successfully")
	return nil
}
