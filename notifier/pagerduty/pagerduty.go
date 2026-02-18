package pagerduty

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

const eventsAPIURL = "https://events.pagerduty.com/v2/enqueue"

func init() {
	notifier.Register("pagerduty", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "pagerduty",
		DisplayName: "PagerDuty",
		Description: "Trigger PagerDuty incidents for critical/security updates",
	})
}

// PagerDutyNotifier triggers PagerDuty events via the Events API v2.
type PagerDutyNotifier struct {
	routingKey string
	severity   string
	httpClient *http.Client
}

// NewFromConfig creates a PagerDutyNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	routingKey := opts.GetString("routing_key", "")
	if routingKey == "" {
		return nil, fmt.Errorf("pagerduty: routing_key is required")
	}

	return &PagerDutyNotifier{
		routingKey: routingKey,
		severity:   opts.GetString("severity", "warning"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *PagerDutyNotifier) Name() string { return "pagerduty" }

func (p *PagerDutyNotifier) Send(hostname string, results []*checker.CheckResult) error {
	summary := formatting.SummarizeResults(results)
	_, body := formatting.BuildMarkdownMessage(hostname, results, formatting.MessageOptions{UseEmoji: false})

	// Determine severity
	severity := p.severity
	if summary.SecurityCount > 0 {
		severity = "critical"
	}

	eventSummary := fmt.Sprintf("Update Report: %s — %d updates found", hostname, summary.TotalUpdates)
	if summary.SecurityCount > 0 {
		eventSummary += fmt.Sprintf(" (%d security)", summary.SecurityCount)
	}

	payload := map[string]interface{}{
		"routing_key":  p.routingKey,
		"event_action": "trigger",
		"payload": map[string]interface{}{
			"summary":        eventSummary,
			"source":         hostname,
			"severity":       severity,
			"component":      "update-watcher",
			"custom_details": body,
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pagerduty: failed to marshal payload: %w", err)
	}

	slog.Debug("sending pagerduty event", "severity", severity)

	resp, err := p.httpClient.Post(eventsAPIURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("pagerduty: failed to send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pagerduty: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("pagerduty event sent successfully", "severity", severity)
	return nil
}
