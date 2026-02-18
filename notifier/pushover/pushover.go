package pushover

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/notifier"
	"github.com/mahype/update-watcher/notifier/formatting"
)

const apiURL = "https://api.pushover.net/1/messages.json"

func init() {
	notifier.Register("pushover", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "pushover",
		DisplayName: "Pushover",
		Description: "Push notifications via Pushover (iOS, Android, Desktop)",
	})
}

// PushoverNotifier sends update reports via Pushover.
type PushoverNotifier struct {
	appToken   string
	userKey    string
	device     string
	priority   int
	sound      string
	httpClient *http.Client
}

// NewFromConfig creates a PushoverNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	appToken := cfg.Options.GetString("app_token", "")
	if appToken == "" {
		return nil, fmt.Errorf("pushover: app_token is required")
	}
	userKey := cfg.Options.GetString("user_key", "")
	if userKey == "" {
		return nil, fmt.Errorf("pushover: user_key is required")
	}

	priority := cfg.Options.GetInt("priority", 0)

	return &PushoverNotifier{
		appToken:   appToken,
		userKey:    userKey,
		device:     cfg.Options.GetString("device", ""),
		priority:   priority,
		sound:      cfg.Options.GetString("sound", ""),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *PushoverNotifier) Name() string { return "pushover" }

func (p *PushoverNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	summary := formatting.SummarizeResults(results)

	// Escalate priority for security updates
	priority := p.priority
	if summary.SecurityCount > 0 && priority < 1 {
		priority = 1
	}

	form := url.Values{
		"token":   {p.appToken},
		"user":    {p.userKey},
		"title":   {title},
		"message": {body},
	}

	form.Set("priority", fmt.Sprintf("%d", priority))

	if p.device != "" {
		form.Set("device", p.device)
	}
	if p.sound != "" {
		form.Set("sound", p.sound)
	}

	slog.Debug("sending pushover notification")

	resp, err := p.httpClient.PostForm(apiURL, form)
	if err != nil {
		return fmt.Errorf("pushover: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pushover: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("pushover notification sent successfully")
	return nil
}
