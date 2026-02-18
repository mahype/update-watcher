package homeassistant

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
	notifier.Register("homeassistant", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "homeassistant",
		DisplayName: "Home Assistant",
		Description: "Push notifications via Home Assistant notify service",
	})
}

// HomeAssistantNotifier sends update reports via the Home Assistant REST API.
type HomeAssistantNotifier struct {
	url     string
	token   string
	service string
}

// NewFromConfig creates a HomeAssistantNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	url := cfg.Options.GetString("url", "")
	if url == "" {
		return nil, fmt.Errorf("homeassistant: url is required")
	}
	token := cfg.Options.GetString("token", "")
	if token == "" {
		return nil, fmt.Errorf("homeassistant: token is required")
	}

	service := cfg.Options.GetString("service", "notify")

	return &HomeAssistantNotifier{
		url:     strings.TrimRight(url, "/"),
		token:   token,
		service: service,
	}, nil
}

func (h *HomeAssistantNotifier) Name() string { return "homeassistant" }

func (h *HomeAssistantNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())

	payload := map[string]interface{}{
		"message": body,
		"title":   title,
	}

	url := fmt.Sprintf("%s/api/services/notify/%s", h.url, h.service)

	slog.Debug("sending home assistant notification", "url", url)

	headers := map[string]string{
		"Authorization": "Bearer " + h.token,
	}
	if err := httputil.PostJSONWithHeaders(url, payload, headers); err != nil {
		return fmt.Errorf("homeassistant: %w", err)
	}

	slog.Info("home assistant notification sent successfully")
	return nil
}
