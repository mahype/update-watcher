package homebrew

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleJSON = `{
  "formulae": [
    {
      "name": "git",
      "installed_versions": ["2.43.0"],
      "current_version": "2.44.0",
      "pinned": false,
      "pinned_version": null
    },
    {
      "name": "curl",
      "installed_versions": ["8.5.0"],
      "current_version": "8.6.0",
      "pinned": false,
      "pinned_version": null
    },
    {
      "name": "node",
      "installed_versions": ["21.5.0"],
      "current_version": "21.6.0",
      "pinned": true,
      "pinned_version": "21.5.0"
    }
  ],
  "casks": [
    {
      "name": "firefox",
      "installed_versions": ["121.0"],
      "current_version": "122.0"
    }
  ]
}`

func TestParseOutdated(t *testing.T) {
	updates, err := parseOutdated(sampleJSON, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 formulae (node is pinned) + 1 cask = 3
	if len(updates) != 3 {
		t.Fatalf("expected 3 updates, got %d", len(updates))
	}

	// First formula
	if updates[0].Name != "git" {
		t.Errorf("expected first package to be git, got %s", updates[0].Name)
	}
	if updates[0].CurrentVersion != "2.43.0" {
		t.Errorf("unexpected current version: %s", updates[0].CurrentVersion)
	}
	if updates[0].NewVersion != "2.44.0" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}
	if updates[0].Source != "formulae" {
		t.Errorf("expected source formulae, got %s", updates[0].Source)
	}

	// Cask
	if updates[2].Name != "firefox" {
		t.Errorf("expected cask to be firefox, got %s", updates[2].Name)
	}
	if updates[2].Source != "casks" {
		t.Errorf("expected source casks, got %s", updates[2].Source)
	}
}

func TestParseOutdatedNoCasks(t *testing.T) {
	updates, err := parseOutdated(sampleJSON, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 formulae only (node is pinned, casks excluded)
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates (no casks), got %d", len(updates))
	}

	for _, u := range updates {
		if u.Source == "casks" {
			t.Error("cask should be excluded when includeCasks is false")
		}
	}
}

func TestParseOutdatedEmpty(t *testing.T) {
	updates, err := parseOutdated("", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseOutdatedEmptyObject(t *testing.T) {
	updates, err := parseOutdated(`{"formulae":[],"casks":[]}`, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseOutdatedPinned(t *testing.T) {
	input := `{
  "formulae": [
    {
      "name": "node",
      "installed_versions": ["21.5.0"],
      "current_version": "21.6.0",
      "pinned": true,
      "pinned_version": "21.5.0"
    }
  ],
  "casks": []
}`
	updates, err := parseOutdated(input, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates (all pinned), got %d", len(updates))
	}
}
