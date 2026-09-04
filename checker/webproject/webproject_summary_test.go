package webproject

import "testing"

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name                              string
		updates, security, projects, fail int
		want                              string
	}{
		{"no updates, no errors", 0, 0, 2, 0, "all projects are up to date"},
		{"updates without security", 4, 0, 2, 0, "4 outdated packages across 2 projects"},
		{"updates with security", 4, 1, 2, 0, "4 outdated packages (1 security) across 2 projects"},
		{"all projects failed", 0, 0, 1, 1, "check failed for all 1 projects"},
		{"some failed, no updates", 0, 0, 3, 1, "2 of 3 projects up to date, check failed for 1"},
		{"some failed, with updates", 4, 1, 3, 1, "4 outdated packages (1 security) across 3 projects, check failed for 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSummary(tt.updates, tt.security, tt.projects, tt.fail); got != tt.want {
				t.Errorf("buildSummary(%d, %d, %d, %d) = %q, want %q",
					tt.updates, tt.security, tt.projects, tt.fail, got, tt.want)
			}
		})
	}
}
