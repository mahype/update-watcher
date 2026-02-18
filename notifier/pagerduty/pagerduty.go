package pagerduty

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
}

// NewFromConfig creates a PagerDutyNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	routingKey := cfg.Options.GetString("routing_key", "")
	if routingKey == "" {
		return nil, fmt.Errorf("pagerduty: routing_key is required")
	}

	return &PagerDutyNotifier{
		routingKey: routingKey,
		severity:   cfg.Options.GetString("severity", "warning"),
	}, nil
}

func (p *PagerDutyNotifier) Name() string { return "pagerduty" }

func (p *PagerDutyNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
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

	slog.Debug("sending pagerduty event", "severity", severity)

	if err := httputil.PostJSON(eventsAPIURL, payload); err != nil {
		return fmt.Errorf("pagerduty: %w", err)
	}

	slog.Info("pagerduty event sent successfully", "severity", severity)
	return nil
}
