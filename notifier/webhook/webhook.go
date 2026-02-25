package webhook

import (
	"context"
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
	notifier.Register("webhook", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "webhook",
		DisplayName: "Webhook",
		Description: "Send JSON payloads to any HTTP endpoint",
	})
}

// WebhookNotifier sends update reports via HTTP webhooks.
type WebhookNotifier struct {
	url         string
	method      string
	contentType string
	authHeader  string
	headers     map[string]string
	httpClient  *http.Client
}

// NewFromConfig creates a WebhookNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	url := cfg.Options.GetString("url", "")
	if url == "" {
		return nil, fmt.Errorf("webhook: url is required")
	}

	headers := make(map[string]string)
	if hdrs, ok := cfg.Options["headers"]; ok {
		if hdrMap, ok := hdrs.(map[string]interface{}); ok {
			for k, v := range hdrMap {
				if s, ok := v.(string); ok {
					headers[k] = s
				}
			}
		}
	}

	return &WebhookNotifier{
		url:         url,
		method:      cfg.Options.GetString("method", "POST"),
		contentType: cfg.Options.GetString("content_type", "application/json"),
		authHeader:  cfg.Options.GetString("auth_header", ""),
		headers:     headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (w *WebhookNotifier) Name() string { return "webhook" }

// Payload is the JSON structure sent by the webhook notifier.
type Payload struct {
	Hostname     string           `json:"hostname"`
	Timestamp    string           `json:"timestamp"`
	TotalUpdates int              `json:"total_updates"`
	HasSecurity  bool             `json:"has_security"`
	Checkers     []CheckerPayload `json:"checkers"`
}

// CheckerPayload represents a single checker's results in the webhook payload.
type CheckerPayload struct {
	Name    string              `json:"name"`
	Summary string              `json:"summary"`
	Error   string              `json:"error,omitempty"`
	Updates []checker.Update    `json:"updates"`
}

func (w *WebhookNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	summary := formatting.SummarizeResults(results)

	payload := Payload{
		Hostname:     hostname,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalUpdates: summary.TotalUpdates,
		HasSecurity:  summary.SecurityCount > 0,
	}

	for _, r := range results {
		cp := CheckerPayload{
			Name:    r.CheckerName,
			Summary: r.Summary,
			Error:   r.Error,
			Updates: r.Updates,
		}
		if cp.Updates == nil {
			cp.Updates = []checker.Update{}
		}
		payload.Checkers = append(payload.Checkers, cp)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(w.method, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", w.contentType)
	req.Header.Set("User-Agent", "update-watcher")

	if w.authHeader != "" {
		req.Header.Set("Authorization", w.authHeader)
	}

	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	slog.Debug("sending webhook notification", "url", w.url, "method", w.method)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Debug("webhook response", "status", resp.StatusCode, "headers", resp.Header, "body", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("webhook notification sent successfully")
	return nil
}
