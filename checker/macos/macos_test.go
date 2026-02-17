package macos

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = `Software Update Tool

Finding available software updates...

Software Update found the following new or updated software:
* Label: macOS Sonoma 14.3.1
	Title: macOS Sonoma 14.3.1, Version: 14.3.1, Size: 1234K, Recommended: YES, Action: restart,
* Label: Safari 17.3.1
	Title: Safari, Version: 17.3.1, Size: 123K, Recommended: YES,
* Label: Security Update 2024-001
	Title: Security Update 2024-001, Version: 2024-001, Size: 500K, Recommended: YES, Action: restart,
`

const sampleNoUpdates = `Software Update Tool

Finding available software updates...

No new software available.
`

const sampleRapidResponse = `Software Update Tool

Finding available software updates...

Software Update found the following new or updated software:
* Label: macOS Ventura 13.4.1 (a) Rapid Security Response
	Title: macOS Ventura 13.4.1 (a), Version: 13.4.1, Size: 300K, Recommended: YES, Action: restart,
`

func TestParseSoftwareUpdate(t *testing.T) {
	updates := parseSoftwareUpdate(sampleOutput, false)

	if len(updates) != 3 {
		t.Fatalf("expected 3 updates, got %d", len(updates))
	}

	// macOS Sonoma: recommended but not security
	if updates[0].Name != "macOS Sonoma 14.3.1" {
		t.Errorf("expected name 'macOS Sonoma 14.3.1', got %q", updates[0].Name)
	}
	if updates[0].NewVersion != "14.3.1" {
		t.Errorf("expected version '14.3.1', got %q", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected type regular, got %q", updates[0].Type)
	}
	if updates[0].Priority != checker.PriorityHigh {
		t.Errorf("expected priority high (recommended), got %q", updates[0].Priority)
	}

	// Safari: recommended but not security
	if updates[1].Name != "Safari 17.3.1" {
		t.Errorf("expected name 'Safari 17.3.1', got %q", updates[1].Name)
	}
	if updates[1].Type != checker.UpdateTypeRegular {
		t.Errorf("expected type regular, got %q", updates[1].Type)
	}
	if updates[1].Priority != checker.PriorityHigh {
		t.Errorf("expected priority high (recommended), got %q", updates[1].Priority)
	}

	// Security Update: security
	if updates[2].Name != "Security Update 2024-001" {
		t.Errorf("expected name 'Security Update 2024-001', got %q", updates[2].Name)
	}
	if updates[2].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected type security, got %q", updates[2].Type)
	}
	if updates[2].Priority != checker.PriorityCritical {
		t.Errorf("expected priority critical, got %q", updates[2].Priority)
	}
}

func TestParseSoftwareUpdateSecurityOnly(t *testing.T) {
	updates := parseSoftwareUpdate(sampleOutput, true)

	if len(updates) != 1 {
		t.Fatalf("expected 1 security update, got %d", len(updates))
	}
	if updates[0].Name != "Security Update 2024-001" {
		t.Errorf("expected 'Security Update 2024-001', got %q", updates[0].Name)
	}
}

func TestParseSoftwareUpdateNoUpdates(t *testing.T) {
	updates := parseSoftwareUpdate(sampleNoUpdates, false)

	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseSoftwareUpdateEmptyOutput(t *testing.T) {
	updates := parseSoftwareUpdate("", false)

	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseSoftwareUpdateRapidSecurityResponse(t *testing.T) {
	updates := parseSoftwareUpdate(sampleRapidResponse, false)

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected type security for Rapid Security Response, got %q", updates[0].Type)
	}
	if updates[0].Priority != checker.PriorityCritical {
		t.Errorf("expected priority critical, got %q", updates[0].Priority)
	}
}
