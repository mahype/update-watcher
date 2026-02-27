package updatewall

import (
	"bytes"
	"context"
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
	notifier.Register("updatewall", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "updatewall",
		DisplayName: "Update Wall",
		Description: "Send update reports to an Update Wall dashboard",
		AlwaysSend:  true,
	})
}

// UpdateWallNotifier sends update reports to an Update Wall dashboard via its REST API.
type UpdateWallNotifier struct {
	url        string
	apiToken   string
	httpClient *http.Client
}

// NewFromConfig creates an UpdateWallNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	url := cfg.Options.GetString("url", "")
	if url == "" {
		return nil, fmt.Errorf("updatewall: url is required")
	}
	apiToken := cfg.Options.GetString("api_token", "")
	if apiToken == "" {
		return nil, fmt.Errorf("updatewall: api_token is required")
	}
	return &UpdateWallNotifier{
		url:        url,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (u *UpdateWallNotifier) Name() string { return "updatewall" }

type payload struct {
	Hostname     string         `json:"hostname"`
	Timestamp    string         `json:"timestamp"`
	TotalUpdates int            `json:"total_updates"`
	HasSecurity  bool           `json:"has_security"`
	Checkers     []checkerEntry `json:"checkers"`
}

type checkerEntry struct {
	Name          string           `json:"name"`
	Summary       string           `json:"summary"`
	Error         string           `json:"error,omitempty"`
	UpdateCommand string           `json:"update_command,omitempty"`
	Updates       []checker.Update `json:"updates,omitempty"`
}

func (u *UpdateWallNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	summary := formatting.SummarizeResults(results)

	p := payload{
		Hostname:     hostname,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalUpdates: summary.TotalUpdates,
		HasSecurity:  summary.SecurityCount > 0,
		Checkers:     make([]checkerEntry, 0, len(results)),
	}

	for _, r := range results {
		p.Checkers = append(p.Checkers, checkerEntry{
			Name:          r.CheckerName,
			Summary:       r.Summary,
			Error:         r.Error,
			UpdateCommand: formatting.UpdateCommandForResult(r.CheckerName, r.Updates),
			Updates:       r.Updates,
		})
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("updatewall: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("updatewall: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.apiToken)

	slog.Debug("sending updatewall notification", "url", u.url)

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("updatewall: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Debug("updatewall response", "status", resp.StatusCode, "headers", resp.Header, "body", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("updatewall: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var serverResp struct {
		Status    string `json:"status"`
		ReportID  int    `json:"report_id"`
		MachineID int    `json:"machine_id"`
	}
	if err := json.Unmarshal(respBody, &serverResp); err != nil {
		return fmt.Errorf("updatewall: failed to parse server response: %w", err)
	}
	if serverResp.Status != "ok" {
		return fmt.Errorf("updatewall: server responded with status %q", serverResp.Status)
	}

	slog.Info("updatewall notification sent successfully", "report_id", serverResp.ReportID, "machine_id", serverResp.MachineID)
	return nil
}
