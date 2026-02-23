package telegram

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
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

const maxMessageLength = 4096

func init() {
	notifier.Register("telegram", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "telegram",
		DisplayName: "Telegram",
		Description: "Send notifications via Telegram Bot API",
	})
}

// TelegramNotifier sends update reports via Telegram Bot API.
type TelegramNotifier struct {
	botToken            string
	chatID              string
	parseMode           string
	disableNotification bool
	httpClient          *http.Client
}

// NewFromConfig creates a TelegramNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
	botToken := cfg.Options.GetString("bot_token", "")
	if botToken == "" {
		return nil, fmt.Errorf("telegram: bot_token is required")
	}
	chatID := cfg.Options.GetString("chat_id", "")
	if chatID == "" {
		return nil, fmt.Errorf("telegram: chat_id is required")
	}

	return &TelegramNotifier{
		botToken:            botToken,
		chatID:              chatID,
		parseMode:           cfg.Options.GetString("parse_mode", "HTML"),
		disableNotification: cfg.Options.GetBool("disable_notification", false),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	message := buildHTMLMessage(hostname, results)

	// Split into chunks if too long
	chunks := splitMessage(message, maxMessageLength)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	for i, chunk := range chunks {
		payload := map[string]interface{}{
			"chat_id":              t.chatID,
			"text":                 chunk,
			"parse_mode":           t.parseMode,
			"disable_notification": t.disableNotification,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("telegram: failed to marshal payload: %w", err)
		}

		slog.Debug("sending telegram message", "chunk", i+1, "of", len(chunks))

		resp, err := t.httpClient.Post(apiURL, "application/json", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("telegram: failed to send message: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("telegram: API returned %d: %s", resp.StatusCode, string(respBody))
		}
	}

	slog.Info("telegram notification sent successfully")
	return nil
}

func buildHTMLMessage(hostname string, results []*checker.CheckResult) string {
	summary := formatting.SummarizeResults(results)
	var parts []string

	// Header
	parts = append(parts, fmt.Sprintf("<b>\U0001f504 Update Report: %s</b>", html.EscapeString(hostname)))
	parts = append(parts, fmt.Sprintf("<i>Checked at %s | %d checkers | %d updates found</i>",
		time.Now().UTC().Format("2006-01-02 15:04 UTC"), summary.CheckerCount, summary.TotalUpdates))

	// Per-checker sections
	for _, r := range results {
		icon := formatting.CheckerEmoji(r.CheckerName, true)
		section := fmt.Sprintf("<b>%s %s</b> — %s",
			icon, html.EscapeString(formatting.CheckerDisplayName(r.CheckerName)), html.EscapeString(r.Summary))

		if r.Error != "" {
			section += fmt.Sprintf("\n\u26a0\ufe0f %s", html.EscapeString(r.Error))
		}

		updates := formatUpdatesHTML(r)
		if updates != "" {
			section += "\n\n" + updates
		}

		if cmd := formatting.UpdateCommandForResult(r.CheckerName, r.Updates); cmd != "" && len(r.Updates) > 0 {
			section += fmt.Sprintf("\n\n\U0001f4a1 Update: <code>%s</code>", html.EscapeString(cmd))
		}

		if count, cmd := formatting.PhasingNote(r.CheckerName, r.Updates); count > 0 {
			section += fmt.Sprintf("\n\u23f3 %d phased update(s) cannot be installed via regular upgrade. Use:\n<code>%s</code>", count, html.EscapeString(cmd))
		}

		if count, cmd := formatting.KeptBackNote(r.CheckerName, r.Updates); count > 0 {
			section += fmt.Sprintf("\n\u23f3 %d package(s) held back \u2014 need new dependencies or removals. Use:\n<code>%s</code>", count, html.EscapeString(cmd))
		}

		for _, note := range r.Notes {
			section += fmt.Sprintf("\n\u23f3 %s", html.EscapeString(note))
		}

		parts = append(parts, section)
	}

	// Security footer
	if summary.SecurityCount > 0 {
		parts = append(parts, fmt.Sprintf("\u26a0\ufe0f <b>Security updates require attention</b> (%d security updates)", summary.SecurityCount))
	}

	return strings.Join(parts, "\n\n")
}

func formatUpdatesHTML(r *checker.CheckResult) string {
	if len(r.Updates) == 0 {
		return ""
	}

	if r.CheckerName == "wordpress" {
		return formatWordPressUpdatesHTML(r.Updates)
	}

	var lines []string
	for _, u := range r.Updates {
		indicator := formatting.PriorityIndicator(u, true)
		var line string
		if u.Type == checker.UpdateTypeSecurity {
			line = fmt.Sprintf("%s <b><code>%s</code></b> %s \u2192 %s \u26a0\ufe0f <b>SECURITY</b>",
				indicator, html.EscapeString(u.Name), html.EscapeString(u.CurrentVersion), html.EscapeString(u.NewVersion))
		} else {
			line = fmt.Sprintf("%s <code>%s</code> %s \u2192 %s",
				indicator, html.EscapeString(u.Name), html.EscapeString(u.CurrentVersion), html.EscapeString(u.NewVersion))
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", html.EscapeString(u.Source))
		}
		if u.Phasing == "held" {
			line += " <i>(kept back)</i>"
		} else if u.Phasing != "" {
			line += fmt.Sprintf(" <i>(phased %s)</i>", html.EscapeString(u.Phasing))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWordPressUpdatesHTML(updates []checker.Update) string {
	grouped := make(map[string][]checker.Update)
	var order []string
	for _, u := range updates {
		if _, exists := grouped[u.Source]; !exists {
			order = append(order, u.Source)
		}
		grouped[u.Source] = append(grouped[u.Source], u)
	}

	var sections []string
	for _, source := range order {
		siteUpdates := grouped[source]
		lines := []string{fmt.Sprintf("<b>%s</b>", html.EscapeString(source))}
		for _, u := range siteUpdates {
			indicator := formatting.PriorityIndicator(u, true)
			typeName := strings.ToUpper(u.Type[:1]) + u.Type[1:]
			var line string
			if u.Type == checker.UpdateTypeSecurity {
				line = fmt.Sprintf("%s %s: <b><code>%s</code></b> %s \u2192 %s \u26a0\ufe0f <b>SECURITY</b>",
					indicator, typeName, html.EscapeString(u.Name), html.EscapeString(u.CurrentVersion), html.EscapeString(u.NewVersion))
			} else {
				line = fmt.Sprintf("%s %s: <code>%s</code> %s \u2192 %s",
					indicator, typeName, html.EscapeString(u.Name), html.EscapeString(u.CurrentVersion), html.EscapeString(u.NewVersion))
			}
			lines = append(lines, line)
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

// splitMessage splits a message into chunks that fit within Telegram's limit.
func splitMessage(msg string, maxLen int) []string {
	if len(msg) <= maxLen {
		return []string{msg}
	}

	var chunks []string
	for len(msg) > 0 {
		if len(msg) <= maxLen {
			chunks = append(chunks, msg)
			break
		}

		// Find a good split point (newline before maxLen)
		splitAt := maxLen
		if idx := strings.LastIndex(msg[:maxLen], "\n"); idx > 0 {
			splitAt = idx
		}

		chunks = append(chunks, msg[:splitAt])
		msg = msg[splitAt:]
		// Trim leading newline from next chunk
		msg = strings.TrimPrefix(msg, "\n")
	}

	return chunks
}
