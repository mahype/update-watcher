package pacman

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

func TestParseArchAudit(t *testing.T) {
	output := `curl is affected by CVE-2023-46218. High risk!
vim is affected by CVE-2023-46218, CVE-2023-46219. Critical risk!
openssh is affected by CVE-2024-12345. Low risk!
`

	vulns := parseArchAudit(output)

	if len(vulns) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(vulns))
	}
	if vulns["curl"] != "High" {
		t.Errorf("expected curl=High, got %s", vulns["curl"])
	}
	if vulns["vim"] != "Critical" {
		t.Errorf("expected vim=Critical, got %s", vulns["vim"])
	}
	if vulns["openssh"] != "Low" {
		t.Errorf("expected openssh=Low, got %s", vulns["openssh"])
	}
}

func TestParseArchAuditHighestSeverityWins(t *testing.T) {
	output := `curl is affected by CVE-2023-00001. Low risk!
curl is affected by CVE-2023-00002. Critical risk!
curl is affected by CVE-2023-00003. High risk!
`

	vulns := parseArchAudit(output)

	if len(vulns) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(vulns))
	}
	if vulns["curl"] != "Critical" {
		t.Errorf("expected curl=Critical (highest), got %s", vulns["curl"])
	}
}

func TestParseArchAuditEmpty(t *testing.T) {
	vulns := parseArchAudit("")
	if len(vulns) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(vulns))
	}
}

func TestMapArchAuditSeverity(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"Critical", checker.PriorityCritical},
		{"High", checker.PriorityHigh},
		{"Medium", checker.PriorityNormal},
		{"Low", checker.PriorityLow},
		{"critical", checker.PriorityCritical}, // case-insensitive
		{"unknown", checker.PriorityNormal},    // fallback
		{"", checker.PriorityNormal},           // empty
	}

	for _, tt := range tests {
		got := mapArchAuditSeverity(tt.severity)
		if got != tt.want {
			t.Errorf("mapArchAuditSeverity(%q) = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

func TestEnrichWithArchAudit(t *testing.T) {
	updates := []checker.Update{
		{Name: "curl", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
		{Name: "vim", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
		{Name: "firefox", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
	}

	vulns := map[string]string{
		"curl": "High",
		"vim":  "Critical",
	}

	enriched := enrichWithArchAudit(updates, vulns)

	// curl → security, high
	if enriched[0].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected curl type=security, got %s", enriched[0].Type)
	}
	if enriched[0].Priority != checker.PriorityHigh {
		t.Errorf("expected curl priority=high, got %s", enriched[0].Priority)
	}

	// vim → security, critical
	if enriched[1].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected vim type=security, got %s", enriched[1].Type)
	}
	if enriched[1].Priority != checker.PriorityCritical {
		t.Errorf("expected vim priority=critical, got %s", enriched[1].Priority)
	}

	// firefox → unchanged
	if enriched[2].Type != checker.UpdateTypeRegular {
		t.Errorf("expected firefox type=regular, got %s", enriched[2].Type)
	}
	if enriched[2].Priority != checker.PriorityNormal {
		t.Errorf("expected firefox priority=normal, got %s", enriched[2].Priority)
	}
}

func TestEnrichWithArchAuditEmptyVulns(t *testing.T) {
	updates := []checker.Update{
		{Name: "curl", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
	}

	enriched := enrichWithArchAudit(updates, nil)

	if enriched[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected unchanged type, got %s", enriched[0].Type)
	}
}
