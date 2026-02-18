package matrix

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
	notifier.Register("matrix", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "matrix",
		DisplayName: "Matrix",
		Description: "Send notifications to Matrix rooms via client-server API",
	})
}

// MatrixNotifier sends update reports to a Matrix room.
type MatrixNotifier struct {
	homeserver  string
	accessToken string
	roomID      string
	httpClient  *http.Client
	txnCounter  int
}

// NewFromConfig creates a MatrixNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	opts := config.WatcherConfig{Options: cfg.Options}
	homeserver := opts.GetString("homeserver", "")
	if homeserver == "" {
		return nil, fmt.Errorf("matrix: homeserver is required")
	}
	accessToken := opts.GetString("access_token", "")
	if accessToken == "" {
		return nil, fmt.Errorf("matrix: access_token is required")
	}
	roomID := opts.GetString("room_id", "")
	if roomID == "" {
		return nil, fmt.Errorf("matrix: room_id is required")
	}

	return &MatrixNotifier{
		homeserver:  strings.TrimRight(homeserver, "/"),
		accessToken: accessToken,
		roomID:      roomID,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (m *MatrixNotifier) Name() string { return "matrix" }

func (m *MatrixNotifier) Send(hostname string, results []*checker.CheckResult) error {
	_, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())
	plainText := formatting.BuildPlainTextMessage(hostname, results)

	m.txnCounter++
	txnID := fmt.Sprintf("update-watcher-%d-%d", time.Now().UnixNano(), m.txnCounter)

	payload := map[string]interface{}{
		"msgtype":        "m.text",
		"body":           plainText,
		"format":         "org.matrix.custom.html",
		"formatted_body": markdownToBasicHTML(body),
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("matrix: failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		m.homeserver, m.roomID, txnID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("matrix: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.accessToken)

	slog.Debug("sending matrix notification", "room", m.roomID)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("matrix: failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("matrix: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("matrix notification sent successfully")
	return nil
}

// markdownToBasicHTML converts simple markdown to HTML for Matrix.
func markdownToBasicHTML(md string) string {
	// Replace markdown bold **text** with <strong>text</strong>
	result := md
	for {
		start := strings.Index(result, "**")
		if start == -1 {
			break
		}
		end := strings.Index(result[start+2:], "**")
		if end == -1 {
			break
		}
		end += start + 2
		inner := result[start+2 : end]
		result = result[:start] + "<strong>" + inner + "</strong>" + result[end+2:]
	}
	// Replace backtick `code` with <code>code</code>
	for {
		start := strings.Index(result, "`")
		if start == -1 {
			break
		}
		end := strings.Index(result[start+1:], "`")
		if end == -1 {
			break
		}
		end += start + 1
		inner := result[start+1 : end]
		result = result[:start] + "<code>" + inner + "</code>" + result[end+1:]
	}
	// Replace newlines with <br>
	result = strings.ReplaceAll(result, "\n", "<br>\n")
	return result
}
