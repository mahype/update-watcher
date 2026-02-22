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

func TestPriorityLevel(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"critical", 4},
		{"high", 3},
		{"normal", 2},
		{"low", 1},
		{"", 2},       // empty → normal
		{"unknown", 2}, // unknown → normal
	}
	for _, tt := range tests {
		got := PriorityLevel(tt.input)
		if got != tt.want {
			t.Errorf("PriorityLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestValidPriority(t *testing.T) {
	for _, p := range []string{"critical", "high", "normal", "low"} {
		if !ValidPriority(p) {
			t.Errorf("ValidPriority(%q) = false, want true", p)
		}
	}
	for _, p := range []string{"", "unknown", "CRITICAL"} {
		if ValidPriority(p) {
			t.Errorf("ValidPriority(%q) = true, want false", p)
		}
	}
}

func TestFilterByMinPriority(t *testing.T) {
	updates := []Update{
		{Name: "a", Priority: PriorityCritical},
		{Name: "b", Priority: PriorityHigh},
		{Name: "c", Priority: PriorityNormal},
		{Name: "d", Priority: PriorityLow},
		{Name: "e", Priority: ""}, // empty → treated as normal
	}

	tests := []struct {
		minPriority string
		wantNames   []string
	}{
		{"", []string{"a", "b", "c", "d", "e"}},          // no filter
		{"low", []string{"a", "b", "c", "d", "e"}},       // all pass
		{"normal", []string{"a", "b", "c", "e"}},          // low filtered out
		{"high", []string{"a", "b"}},                       // normal+low filtered
		{"critical", []string{"a"}},                        // only critical
	}
	for _, tt := range tests {
		got := FilterByMinPriority(updates, tt.minPriority)
		if len(got) != len(tt.wantNames) {
			t.Errorf("FilterByMinPriority(min=%q): got %d updates, want %d", tt.minPriority, len(got), len(tt.wantNames))
			continue
		}
		for i, u := range got {
			if u.Name != tt.wantNames[i] {
				t.Errorf("FilterByMinPriority(min=%q)[%d].Name = %q, want %q", tt.minPriority, i, u.Name, tt.wantNames[i])
			}
		}
	}
}

func TestFilterResultsByPriority(t *testing.T) {
	results := []*CheckResult{
		{
			CheckerName: "apt",
			Updates: []Update{
				{Name: "curl", Priority: PriorityCritical},
				{Name: "vim", Priority: PriorityLow},
			},
		},
		{
			CheckerName: "docker",
			Updates: []Update{
				{Name: "nginx", Priority: PriorityHigh},
			},
		},
	}

	filtered, total := FilterResultsByPriority(results, "high")
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(filtered[0].Updates) != 1 || filtered[0].Updates[0].Name != "curl" {
		t.Errorf("expected only curl in apt results, got %v", filtered[0].Updates)
	}
	if len(filtered[1].Updates) != 1 || filtered[1].Updates[0].Name != "nginx" {
		t.Errorf("expected nginx in docker results, got %v", filtered[1].Updates)
	}

	// Original must not be mutated.
	if len(results[0].Updates) != 2 {
		t.Errorf("original results mutated: got %d updates, want 2", len(results[0].Updates))
	}
}

func TestFilterResultsByPriorityEmpty(t *testing.T) {
	results := []*CheckResult{
		{CheckerName: "apt", Updates: []Update{{Name: "a", Priority: PriorityNormal}}},
	}
	filtered, total := FilterResultsByPriority(results, "")
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	// With empty minPriority, should return same slice (no copy needed).
	if &filtered[0] != &results[0] {
		t.Error("expected same slice reference for empty minPriority")
	}
}
