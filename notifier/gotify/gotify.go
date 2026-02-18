package gotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/notifier"
	"github.com/mahype/update-watcher/notifier/formatting"
)

func init() {
	notifier.Register("gotify", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "gotify",
		DisplayName: "Gotify",
		Description: "Push notifications via self-hosted Gotify server",
	})
}

// GotifyNotifier sends update reports via Gotify.
type GotifyNotifier struct {
	serverURL  string
	token      string
	priority   int
	httpClient *http.Client
}

// NewFromConfig creates a GotifyNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	serverURL := opts.GetString("server_url", "")
	if serverURL == "" {
		return nil, fmt.Errorf("gotify: server_url is required")
	}
	token := opts.GetString("token", "")
	if token == "" {
		return nil, fmt.Errorf("gotify: token is required")
	}

	priority := 5
	if v, ok := cfg.Options["priority"]; ok {
		switch p := v.(type) {
		case int:
			priority = p
		case float64:
			priority = int(p)
		}
	}

	return &GotifyNotifier{
		serverURL:  strings.TrimRight(serverURL, "/"),
		token:      token,
		priority:   priority,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (g *GotifyNotifier) Name() string { return "gotify" }

func (g *GotifyNotifier) Send(hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	summary := formatting.SummarizeResults(results)

	// Escalate priority for security updates
	priority := g.priority
	if summary.SecurityCount > 0 && priority < 8 {
		priority = 8
	}

	payload := map[string]interface{}{
		"title":    title,
		"message":  body,
		"priority": priority,
		"extras": map[string]interface{}{
			"client::display": map[string]interface{}{
				"contentType": "text/markdown",
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gotify: failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/message", g.serverURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("gotify: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", g.token)

	slog.Debug("sending gotify notification", "url", url)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gotify: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gotify: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("gotify notification sent successfully")
	return nil
}
