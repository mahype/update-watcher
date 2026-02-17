package flatpak

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = "Firefox\torg.mozilla.firefox\t125.0.3\nLibreOffice\torg.libreoffice.LibreOffice\t24.2.2.1\nGIMP\torg.gimp.GIMP\t2.10.36\n"

func TestParseRemoteUpdates(t *testing.T) {
	updates := parseRemoteUpdates(sampleOutput)

	if len(updates) != 3 {
		t.Fatalf("expected 3 updates, got %d", len(updates))
	}

	if updates[0].Name != "org.mozilla.firefox" {
		t.Errorf("expected first app to be org.mozilla.firefox, got %s", updates[0].Name)
	}
	if updates[0].NewVersion != "125.0.3" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Source != "Firefox" {
		t.Errorf("expected source Firefox, got %s", updates[0].Source)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}

	if updates[2].Name != "org.gimp.GIMP" {
		t.Errorf("expected third app to be org.gimp.GIMP, got %s", updates[2].Name)
	}
}

func TestParseRemoteUpdatesEmpty(t *testing.T) {
	updates := parseRemoteUpdates("")
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseRemoteUpdatesMissingVersion(t *testing.T) {
	// Some flatpak entries may lack a version
	output := "Firefox\torg.mozilla.firefox\n"
	updates := parseRemoteUpdates(output)

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].NewVersion != "" {
		t.Errorf("expected empty version, got %s", updates[0].NewVersion)
	}
}
