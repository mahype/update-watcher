package snap

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = `Name           Version                        Rev    Size    Publisher       Notes
core           16-2.45.1+git2022.b6b3c25      9584   97MB    canonical*      core
firefox        125.0.3-2                       4336   283MB   mozilla**       -
get-iplayer    3.26                            250    15MB    snapcrafters    -
`

func TestParseRefreshList(t *testing.T) {
	updates := parseRefreshList(sampleOutput)

	if len(updates) != 3 {
		t.Fatalf("expected 3 updates, got %d", len(updates))
	}

	if updates[0].Name != "core" {
		t.Errorf("expected first package to be core, got %s", updates[0].Name)
	}
	if updates[0].NewVersion != "16-2.45.1+git2022.b6b3c25" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}

	if updates[1].Name != "firefox" {
		t.Errorf("expected second package to be firefox, got %s", updates[1].Name)
	}
	if updates[1].NewVersion != "125.0.3-2" {
		t.Errorf("unexpected new version: %s", updates[1].NewVersion)
	}
}

func TestParseRefreshListEmpty(t *testing.T) {
	updates := parseRefreshList("")
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseRefreshListHeaderOnly(t *testing.T) {
	updates := parseRefreshList("Name           Version          Rev    Size    Publisher       Notes\n")
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates (header only), got %d", len(updates))
	}
}
