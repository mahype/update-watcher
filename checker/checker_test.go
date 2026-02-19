package checker

import "testing"

func TestBuildSummaryNoUpdates(t *testing.T) {
	got := BuildSummary(nil, "packages")
	if got != "all packages are up to date" {
		t.Errorf("unexpected summary: %s", got)
	}
}

func TestBuildSummaryRegularOnly(t *testing.T) {
	updates := []Update{
		{Name: "curl", Type: UpdateTypeRegular},
		{Name: "vim", Type: UpdateTypeRegular},
	}
	got := BuildSummary(updates, "packages")
	if got != "2 packages" {
		t.Errorf("unexpected summary: %s", got)
	}
}

func TestBuildSummarySecurity(t *testing.T) {
	updates := []Update{
		{Name: "curl", Type: UpdateTypeRegular},
		{Name: "libssl3", Type: UpdateTypeSecurity},
	}
	got := BuildSummary(updates, "packages")
	if got != "2 packages (1 security)" {
		t.Errorf("unexpected summary: %s", got)
	}
}

func TestBuildSummaryPhasedOnly(t *testing.T) {
	updates := []Update{
		{Name: "gcc-13", Type: UpdateTypeRegular, Phasing: "deferred"},
		{Name: "libstdc++6", Type: UpdateTypeRegular, Phasing: "deferred"},
	}
	got := BuildSummary(updates, "packages")
	if got != "2 packages (2 phased)" {
		t.Errorf("unexpected summary: %s", got)
	}
}

func TestBuildSummaryMixed(t *testing.T) {
	updates := []Update{
		{Name: "curl", Type: UpdateTypeRegular},
		{Name: "gcc-13", Type: UpdateTypeRegular, Phasing: "deferred"},
		{Name: "libssl3", Type: UpdateTypeSecurity},
		{Name: "libstdc++6", Type: UpdateTypeRegular, Phasing: "50%"},
	}
	got := BuildSummary(updates, "packages")
	if got != "4 packages (2 phased, 1 security)" {
		t.Errorf("unexpected summary: %s", got)
	}
}
