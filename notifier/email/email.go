package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/notifier"
	"github.com/mahype/update-watcher/notifier/formatting"
)

func init() {
	notifier.Register("email", NewFromConfig)
	notifier.RegisterMeta(notifier.NotifierMeta{
		Type:        "email",
		DisplayName: "E-Mail",
		Description: "Send notifications via SMTP email",
	})
}

// EmailNotifier sends update reports via SMTP email.
type EmailNotifier struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
	useTLS   bool
}

// NewFromConfig creates an EmailNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {

	smtpHost := cfg.Options.GetString("smtp_host", "")
	if smtpHost == "" {
		return nil, fmt.Errorf("email: smtp_host is required")
	}

	username := cfg.Options.GetString("username", "")
	if username == "" {
		return nil, fmt.Errorf("email: username is required")
	}

	password := cfg.Options.GetString("password", "")
	if password == "" {
		return nil, fmt.Errorf("email: password is required")
	}

	from := cfg.Options.GetString("from", "")
	if from == "" {
		return nil, fmt.Errorf("email: from is required")
	}

	to := cfg.Options.GetStringSlice("to", nil)
	if len(to) == 0 {
		return nil, fmt.Errorf("email: at least one recipient in 'to' is required")
	}

	smtpPort := cfg.Options.GetInt("smtp_port", 587)

	return &EmailNotifier{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
		to:       to,
		useTLS:   cfg.Options.GetBool("tls", true),
	}, nil
}

func (e *EmailNotifier) Name() string { return "email" }

func (e *EmailNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
	subject := fmt.Sprintf("Update Report: %s", hostname)
	htmlBody := BuildHTMLMessage(hostname, results)
	plainBody := formatting.BuildPlainTextMessage(hostname, results)

	// Build MIME message
	msg := buildMIMEMessage(e.from, e.to, subject, plainBody, htmlBody)

	addr := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)

	slog.Debug("sending email notification", "host", e.smtpHost, "to", e.to)

	var err error
	if e.useTLS {
		err = sendWithSTARTTLS(addr, e.smtpHost, e.username, e.password, e.from, e.to, msg)
	} else {
		err = sendPlain(addr, e.smtpHost, e.username, e.password, e.from, e.to, msg)
	}

	if err != nil {
		return fmt.Errorf("email: failed to send: %w", err)
	}

	slog.Info("email notification sent successfully")
	return nil
}

func buildMIMEMessage(from string, to []string, subject, plainBody, htmlBody string) []byte {
	boundary := "boundary-update-watcher-7f2e8a"

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(plainBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return []byte(msg.String())
}

func sendWithSTARTTLS(addr, host, username, password, from string, to []string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer c.Close()

	// STARTTLS
	tlsConfig := &tls.Config{ServerName: host}
	if err := c.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// Auth
	auth := smtp.PlainAuth("", username, password, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return sendMail(c, from, to, msg)
}

func sendPlain(addr, host, username, password, from string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", username, password, host)
	return smtp.SendMail(addr, auth, from, to, msg)
}

func sendMail(c *smtp.Client, from string, to []string, msg []byte) error {
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", addr, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return c.Quit()
}
