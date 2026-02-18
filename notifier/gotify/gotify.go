package gotify

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/httputil"
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
	serverURL string
	token     string
	priority  int
}

// NewFromConfig creates a GotifyNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	serverURL := cfg.Options.GetString("server_url", "")
	if serverURL == "" {
		return nil, fmt.Errorf("gotify: server_url is required")
	}
	token := cfg.Options.GetString("token", "")
	if token == "" {
		return nil, fmt.Errorf("gotify: token is required")
	}

	priority := cfg.Options.GetInt("priority", 5)

	return &GotifyNotifier{
		serverURL: strings.TrimRight(serverURL, "/"),
		token:     token,
		priority:  priority,
	}, nil
}

func (g *GotifyNotifier) Name() string { return "gotify" }

func (g *GotifyNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
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

	url := fmt.Sprintf("%s/message", g.serverURL)

	slog.Debug("sending gotify notification", "url", url)

	headers := map[string]string{
		"X-Gotify-Key": g.token,
	}
	if err := httputil.PostJSONWithHeaders(url, payload, headers); err != nil {
		return fmt.Errorf("gotify: %w", err)
	}

	slog.Info("gotify notification sent successfully")
	return nil
}
