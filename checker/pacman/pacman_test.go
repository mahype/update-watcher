package pacman

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = `linux 6.6.7.arch1-1 -> 6.6.8.arch1-1
vim 9.0.2136-1 -> 9.0.2155-1
curl 8.5.0-1 -> 8.6.0-1
firefox 121.0-1 -> 121.0.1-1
openssh 9.6p1-1 -> 9.6p1-2
`

func TestParseUpgradable(t *testing.T) {
	updates := parseUpgradable(sampleOutput)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	if updates[0].Name != "linux" {
		t.Errorf("expected first package to be linux, got %s", updates[0].Name)
	}
	if updates[0].CurrentVersion != "6.6.7.arch1-1" {
		t.Errorf("unexpected current version: %s", updates[0].CurrentVersion)
	}
	if updates[0].NewVersion != "6.6.8.arch1-1" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}
}

func TestParseUpgradableEmpty(t *testing.T) {
	updates := parseUpgradable("")
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseUpgradableIgnored(t *testing.T) {
	output := `linux 6.6.7.arch1-1 -> 6.6.8.arch1-1
nvidia 545.29.06-3 -> 545.29.06-4 [ignored]
curl 8.5.0-1 -> 8.6.0-1
`
	updates := parseUpgradable(output)
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates (ignored excluded), got %d", len(updates))
	}
	for _, u := range updates {
		if u.Name == "nvidia" {
			t.Error("nvidia should be excluded (ignored)")
		}
	}
}
