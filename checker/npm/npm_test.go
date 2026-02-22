package npm

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleJSON = `{
  "typescript": {
    "current": "5.3.3",
    "wanted": "5.3.3",
    "latest": "5.7.3"
  },
  "@angular/cli": {
    "current": "17.0.0",
    "wanted": "17.0.0",
    "latest": "19.1.0"
  }
}`

func TestParseOutdated(t *testing.T) {
	updates, err := parseOutdated(sampleJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}

	// Map iteration order is non-deterministic; check by name.
	found := map[string]checker.Update{}
	for _, u := range updates {
		found[u.Name] = u
	}

	ts, ok := found["typescript"]
	if !ok {
		t.Fatal("expected typescript in updates")
	}
	if ts.CurrentVersion != "5.3.3" {
		t.Errorf("typescript current: got %s, want 5.3.3", ts.CurrentVersion)
	}
	if ts.NewVersion != "5.7.3" {
		t.Errorf("typescript latest: got %s, want 5.7.3", ts.NewVersion)
	}
	if ts.Type != checker.UpdateTypeRegular {
		t.Errorf("typescript type: got %s, want %s", ts.Type, checker.UpdateTypeRegular)
	}

	ng, ok := found["@angular/cli"]
	if !ok {
		t.Fatal("expected @angular/cli in updates")
	}
	if ng.CurrentVersion != "17.0.0" {
		t.Errorf("@angular/cli current: got %s, want 17.0.0", ng.CurrentVersion)
	}
	if ng.NewVersion != "19.1.0" {
		t.Errorf("@angular/cli latest: got %s, want 19.1.0", ng.NewVersion)
	}
}

func TestParseOutdatedEmpty(t *testing.T) {
	updates, err := parseOutdated("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseOutdatedEmptyObject(t *testing.T) {
	updates, err := parseOutdated("{}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseOutdatedAlreadyCurrent(t *testing.T) {
	input := `{
  "eslint": {
    "current": "9.0.0",
    "wanted": "9.0.0",
    "latest": "9.0.0"
  }
}`
	updates, err := parseOutdated(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates (already current), got %d", len(updates))
	}
}

func TestParseOutdatedInvalidJSON(t *testing.T) {
	_, err := parseOutdated("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
