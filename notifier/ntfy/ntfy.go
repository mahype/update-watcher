package ntfy

import (
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
	notifier.Register("ntfy", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "ntfy",
		DisplayName: "ntfy",
		Description: "Push notifications via ntfy.sh or self-hosted ntfy server",
	})
}

// NtfyNotifier sends update reports via ntfy.sh.
type NtfyNotifier struct {
	serverURL  string
	topic      string
	token      string
	priority   string
	httpClient *http.Client
}

// NewFromConfig creates a NtfyNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	topic := opts.GetString("topic", "")
	if topic == "" {
		return nil, fmt.Errorf("ntfy: topic is required")
	}

	return &NtfyNotifier{
		serverURL: opts.GetString("server_url", "https://ntfy.sh"),
		topic:     topic,
		token:     opts.GetString("token", ""),
		priority:  opts.GetString("priority", "default"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (n *NtfyNotifier) Name() string { return "ntfy" }

func (n *NtfyNotifier) Send(hostname string, results []*checker.CheckResult) error {
	title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	summary := formatting.SummarizeResults(results)

	url := fmt.Sprintf("%s/%s", strings.TrimRight(n.serverURL, "/"), n.topic)
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("ntfy: failed to create request: %w", err)
	}

	req.Header.Set("Title", title)
	req.Header.Set("Markdown", "yes")

	// Set priority — escalate to high if security updates present
	priority := n.priority
	if summary.SecurityCount > 0 && (priority == "default" || priority == "low" || priority == "min") {
		priority = "high"
	}
	req.Header.Set("Priority", priority)

	// Tags
	tags := []string{"package"}
	if summary.SecurityCount > 0 {
		tags = append(tags, "warning")
	}
	req.Header.Set("Tags", strings.Join(tags, ","))

	if n.token != "" {
		req.Header.Set("Authorization", "Bearer "+n.token)
	}

	slog.Debug("sending ntfy notification", "url", url)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ntfy: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("ntfy notification sent successfully")
	return nil
}
