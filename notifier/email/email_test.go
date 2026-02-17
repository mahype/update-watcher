package email

import (
	"strings"
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"smtp_host": "smtp.example.com",
			"smtp_port": 587,
			"username":  "user@example.com",
			"password":  "secret",
			"from":      "noreply@example.com",
			"to":        []interface{}{"admin@example.com", "ops@example.com"},
			"tls":       true,
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	en := n.(*EmailNotifier)
	if en.smtpHost != "smtp.example.com" {
		t.Errorf("expected smtp host, got %q", en.smtpHost)
	}
	if en.smtpPort != 587 {
		t.Errorf("expected port 587, got %d", en.smtpPort)
	}
	if len(en.to) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(en.to))
	}
}

func TestNewFromConfigMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
	}{
		{"missing smtp_host", map[string]interface{}{"username": "u", "password": "p", "from": "f", "to": []interface{}{"t"}}},
		{"missing username", map[string]interface{}{"smtp_host": "h", "password": "p", "from": "f", "to": []interface{}{"t"}}},
		{"missing password", map[string]interface{}{"smtp_host": "h", "username": "u", "from": "f", "to": []interface{}{"t"}}},
		{"missing from", map[string]interface{}{"smtp_host": "h", "username": "u", "password": "p", "to": []interface{}{"t"}}},
		{"missing to", map[string]interface{}{"smtp_host": "h", "username": "u", "password": "p", "from": "f"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NotifierConfig{Options: tt.options}
			_, err := NewFromConfig(cfg)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestBuildHTMLMessage(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "2 packages (1 security)",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13", NewVersion: "3.0.14", Type: checker.UpdateTypeSecurity},
				{Name: "curl", CurrentVersion: "8.5.0", NewVersion: "8.5.1", Type: checker.UpdateTypeRegular},
			},
		},
	}

	html := BuildHTMLMessage("test-server", results)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected HTML doctype")
	}
	if !strings.Contains(html, "test-server") {
		t.Error("expected hostname in HTML")
	}
	if !strings.Contains(html, "libssl3") {
		t.Error("expected package name in HTML")
	}
	if !strings.Contains(html, "security-banner") {
		t.Error("expected security banner class")
	}
	if !strings.Contains(html, "<table>") {
		t.Error("expected HTML table")
	}
}

func TestBuildHTMLMessageNoSecurity(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "1 package",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "curl", CurrentVersion: "8.5.0", NewVersion: "8.5.1", Type: checker.UpdateTypeRegular},
			},
		},
	}

	html := BuildHTMLMessage("test-server", results)

	if strings.Contains(html, `class="security-banner">`) {
		t.Error("should not contain security banner div when no security updates")
	}
}
