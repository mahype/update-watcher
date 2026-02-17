package formatting

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
)

func sampleResults() []*checker.CheckResult {
	return []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "3 packages (1 security)",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13", NewVersion: "3.0.14", Type: checker.UpdateTypeSecurity, Priority: checker.PriorityHigh},
				{Name: "curl", CurrentVersion: "8.5.0-2", NewVersion: "8.5.0-6", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
				{Name: "git", CurrentVersion: "2.43.0-1", NewVersion: "2.43.0-2", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
			},
		},
		{
			CheckerName: "docker",
			Summary:     "1 container",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "nginx-proxy", CurrentVersion: "abc123", NewVersion: "789xyz", Type: checker.UpdateTypeImage, Priority: checker.PriorityNormal, Source: "nginx:1.25"},
			},
		},
	}
}

func TestSummarizeResults(t *testing.T) {
	results := sampleResults()
	s := SummarizeResults(results)

	if s.TotalUpdates != 4 {
		t.Errorf("expected 4 total updates, got %d", s.TotalUpdates)
	}
	if s.SecurityCount != 1 {
		t.Errorf("expected 1 security update, got %d", s.SecurityCount)
	}
	if s.CheckerCount != 2 {
		t.Errorf("expected 2 checkers, got %d", s.CheckerCount)
	}
}

func TestSummarizeResultsEmpty(t *testing.T) {
	s := SummarizeResults(nil)
	if s.TotalUpdates != 0 || s.SecurityCount != 0 || s.CheckerCount != 0 {
		t.Errorf("expected all zeros for nil results")
	}
}

func TestCheckerEmoji(t *testing.T) {
	if e := CheckerEmoji("apt", true); e != "\U0001f427" {
		t.Errorf("expected penguin emoji for apt, got %q", e)
	}
	if e := CheckerEmoji("docker", true); e != "\U0001f433" {
		t.Errorf("expected whale emoji for docker, got %q", e)
	}
	if e := CheckerEmoji("homebrew", true); e != "\U0001f37a" {
		t.Errorf("expected beer mug emoji for homebrew, got %q", e)
	}
	if e := CheckerEmoji("apt", false); e != "" {
		t.Errorf("expected empty string when emoji disabled, got %q", e)
	}
}

func TestCheckerDisplayName(t *testing.T) {
	tests := map[string]string{
		"apt":       "APT Updates",
		"docker":    "Docker Updates",
		"wordpress": "WordPress Updates",
		"unknown":   "unknown Updates",
	}
	for input, expected := range tests {
		if got := CheckerDisplayName(input); got != expected {
			t.Errorf("CheckerDisplayName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestPriorityIndicator(t *testing.T) {
	secUpdate := checker.Update{Type: checker.UpdateTypeSecurity}
	regUpdate := checker.Update{Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal}
	critUpdate := checker.Update{Type: checker.UpdateTypeRegular, Priority: checker.PriorityCritical}

	if pi := PriorityIndicator(secUpdate, true); pi != "\U0001f534" {
		t.Errorf("expected red circle for security update with emoji")
	}
	if pi := PriorityIndicator(regUpdate, true); pi != "\u26aa" {
		t.Errorf("expected white circle for regular update with emoji")
	}
	if pi := PriorityIndicator(critUpdate, true); pi != "\U0001f534" {
		t.Errorf("expected red circle for critical update with emoji")
	}
	if pi := PriorityIndicator(secUpdate, false); pi != "[!]" {
		t.Errorf("expected [!] for security update without emoji")
	}
	if pi := PriorityIndicator(regUpdate, false); pi != "[-]" {
		t.Errorf("expected [-] for regular update without emoji")
	}
}

func TestBuildMarkdownMessage(t *testing.T) {
	results := sampleResults()
	title, body := BuildMarkdownMessage("test-server", results, DefaultOptions())

	if !strings.Contains(title, "test-server") {
		t.Error("title should contain hostname")
	}
	if !strings.Contains(title, "\U0001f504") {
		t.Error("title should contain emoji when enabled")
	}
	if !strings.Contains(body, "APT Updates") {
		t.Error("body should contain APT Updates section")
	}
	if !strings.Contains(body, "Docker Updates") {
		t.Error("body should contain Docker Updates section")
	}
	if !strings.Contains(body, "libssl3") {
		t.Error("body should contain update package name")
	}
	if !strings.Contains(body, "Security updates require attention") {
		t.Error("body should contain security footer")
	}
	if !strings.Contains(body, "**SECURITY**") {
		t.Error("body should contain bold SECURITY label for security updates")
	}
	if !strings.Contains(body, "**`libssl3`**") {
		t.Error("body should contain bold package name for security updates")
	}
}

func TestBuildMarkdownMessageNoUpdates(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "all packages are up to date",
			CheckedAt:   time.Now(),
		},
	}
	_, body := BuildMarkdownMessage("test-server", results, DefaultOptions())

	if strings.Contains(body, "Security updates require attention") {
		t.Error("body should not contain security footer when no security updates")
	}
}

func TestBuildPlainTextMessage(t *testing.T) {
	results := sampleResults()
	msg := BuildPlainTextMessage("test-server", results)

	if !strings.Contains(msg, "Update Report: test-server") {
		t.Error("message should contain header")
	}
	if !strings.Contains(msg, "APT Updates") {
		t.Error("message should contain APT section")
	}
	if !strings.Contains(msg, "libssl3") {
		t.Error("message should contain package name")
	}
	if !strings.Contains(msg, "Security updates require attention") {
		t.Error("message should contain security footer")
	}
	if !strings.Contains(msg, "[SECURITY]") {
		t.Error("message should contain [SECURITY] label for security updates")
	}
}

func TestFormatUpdatesMarkdownNoTruncation(t *testing.T) {
	var updates []checker.Update
	for i := 0; i < 25; i++ {
		updates = append(updates, checker.Update{
			Name:           fmt.Sprintf("pkg-%d", i),
			CurrentVersion: "1.0",
			NewVersion:     "2.0",
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
		})
	}
	r := &checker.CheckResult{
		CheckerName: "apt",
		Updates:     updates,
	}
	result := FormatUpdatesMarkdown(r, true)

	if strings.Contains(result, "more") {
		t.Error("should not truncate updates")
	}
	for i := 0; i < 25; i++ {
		if !strings.Contains(result, fmt.Sprintf("pkg-%d", i)) {
			t.Errorf("should contain all updates, missing pkg-%d", i)
		}
	}
}

func TestFormatUpdatesPlainTextNoTruncation(t *testing.T) {
	var updates []checker.Update
	for i := 0; i < 25; i++ {
		updates = append(updates, checker.Update{
			Name:           fmt.Sprintf("pkg-%d", i),
			CurrentVersion: "1.0",
			NewVersion:     "2.0",
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
		})
	}
	r := &checker.CheckResult{
		CheckerName: "apt",
		Updates:     updates,
	}
	result := FormatUpdatesPlainText(r)

	if strings.Contains(result, "more") {
		t.Error("should not truncate updates")
	}
	for i := 0; i < 25; i++ {
		if !strings.Contains(result, fmt.Sprintf("pkg-%d", i)) {
			t.Errorf("should contain all updates, missing pkg-%d", i)
		}
	}
}

func TestFormatUpdatesMarkdownPhasing(t *testing.T) {
	r := &checker.CheckResult{
		CheckerName: "apt",
		Updates: []checker.Update{
			{Name: "git", CurrentVersion: "2.43.0-1", NewVersion: "2.43.0-2", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal, Phasing: "50%"},
			{Name: "curl", CurrentVersion: "8.5.0-2", NewVersion: "8.5.0-6", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
		},
	}

	md := FormatUpdatesMarkdown(r, true)
	if !strings.Contains(md, "_(phased 50%)_") {
		t.Error("markdown should contain phasing info for git")
	}
	if strings.Count(md, "phased") != 1 {
		t.Error("only git should have phasing info, not curl")
	}

	plain := FormatUpdatesPlainText(r)
	if !strings.Contains(plain, "[phased 50%]") {
		t.Error("plaintext should contain phasing info for git")
	}
	if strings.Count(plain, "phased") != 1 {
		t.Error("only git should have phasing info, not curl")
	}
}

func TestFormatUpdatesMarkdownWordPress(t *testing.T) {
	r := &checker.CheckResult{
		CheckerName: "wordpress",
		Updates: []checker.Update{
			{Name: "woocommerce", CurrentVersion: "8.0", NewVersion: "8.1", Type: checker.UpdateTypePlugin, Source: "My Shop"},
			{Name: "storefront", CurrentVersion: "4.0", NewVersion: "4.1", Type: checker.UpdateTypeTheme, Source: "My Shop"},
			{Name: "akismet", CurrentVersion: "5.0", NewVersion: "5.1", Type: checker.UpdateTypePlugin, Source: "Blog"},
		},
	}

	result := FormatUpdatesMarkdown(r, true)

	if !strings.Contains(result, "**My Shop**") {
		t.Error("should group by site name")
	}
	if !strings.Contains(result, "**Blog**") {
		t.Error("should contain second site group")
	}
	if !strings.Contains(result, "Plugin:") {
		t.Error("should contain type label")
	}
}
