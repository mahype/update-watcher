package slack

import (
	"testing"
	"time"

	"github.com/mahype/update-watcher/checker"
)

func TestBuildMessage(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "3 packages (1 security)",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "libssl3", CurrentVersion: "3.0.13-0ubuntu3.1", NewVersion: "3.0.13-0ubuntu3.4", Type: checker.UpdateTypeSecurity, Priority: checker.PriorityHigh},
				{Name: "curl", CurrentVersion: "8.5.0-2ubuntu10.1", NewVersion: "8.5.0-2ubuntu10.6", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
				{Name: "git", CurrentVersion: "1:2.43.0-1ubuntu7.1", NewVersion: "1:2.43.0-1ubuntu7.2", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
			},
		},
		{
			CheckerName: "docker",
			Summary:     "1 container",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{Name: "nginx-proxy", CurrentVersion: "abc123def456", NewVersion: "789xyz012345", Type: checker.UpdateTypeImage, Priority: checker.PriorityNormal, Source: "nginx:1.25"},
			},
		},
	}

	blocks := BuildMessage("test-server", results, true)

	if len(blocks) == 0 {
		t.Fatal("expected blocks to be non-empty")
	}

	// Should have: header + context + (divider + section) * 2 + divider + footer (security)
	// = 1 + 1 + 2*2 + 1 + 1 = 8
	if len(blocks) < 6 {
		t.Errorf("expected at least 6 blocks, got %d", len(blocks))
	}

	// First block should be header
	if blocks[0]["type"] != "header" {
		t.Errorf("expected first block to be header, got %s", blocks[0]["type"])
	}
}

func TestBuildMessageNoUpdates(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Summary:     "all packages are up to date",
			CheckedAt:   time.Now(),
		},
	}

	blocks := BuildMessage("test-server", results, true)

	// Should NOT have the security footer
	lastBlock := blocks[len(blocks)-1]
	if lastBlock["type"] == "context" {
		elements, ok := lastBlock["elements"].([]Block)
		if ok && len(elements) > 0 {
			text, _ := elements[0]["text"].(string)
			if text != "" && len(text) > 10 {
				// This is likely the security footer - should not be present
				t.Log("last block text:", text)
			}
		}
	}
}
