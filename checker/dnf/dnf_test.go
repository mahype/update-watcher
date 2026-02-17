package dnf

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleCheckUpdate = `
vim-enhanced.x86_64              9.0.2136-1.fc39        updates
curl.x86_64                      8.2.1-3.fc39           updates
openssl-libs.x86_64              3.1.4-2.fc39           updates
kernel.x86_64                    6.6.8-200.fc39         updates
podman.x86_64                    4.8.2-1.fc39           updates
`

const sampleSecurityInfo = `FEDORA-2024-abc123    Important/Sec.  openssl-libs-3.1.4-2.fc39.x86_64
FEDORA-2024-def456    Important/Sec.  curl-8.2.1-3.fc39.x86_64
`

func TestParseCheckUpdate(t *testing.T) {
	updates := parseCheckUpdate(sampleCheckUpdate, nil)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	if updates[0].Name != "vim-enhanced" {
		t.Errorf("expected first package to be vim-enhanced, got %s", updates[0].Name)
	}
	if updates[0].NewVersion != "9.0.2136-1.fc39" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}
}

func TestParseCheckUpdateWithSecurity(t *testing.T) {
	secPkgs := parseSecurityInfo(sampleSecurityInfo)
	updates := parseCheckUpdate(sampleCheckUpdate, secPkgs)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	// curl should be security
	found := false
	for _, u := range updates {
		if u.Name == "curl" {
			found = true
			if u.Type != checker.UpdateTypeSecurity {
				t.Errorf("expected curl to be security update, got %s", u.Type)
			}
			if u.Priority != checker.PriorityHigh {
				t.Errorf("expected curl to have high priority, got %s", u.Priority)
			}
		}
	}
	if !found {
		t.Error("curl not found in updates")
	}

	// vim should be regular
	for _, u := range updates {
		if u.Name == "vim-enhanced" {
			if u.Type != checker.UpdateTypeRegular {
				t.Errorf("expected vim-enhanced to be regular update, got %s", u.Type)
			}
		}
	}
}

func TestParseCheckUpdateEmpty(t *testing.T) {
	updates := parseCheckUpdate("", nil)
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseCheckUpdateWithMetadataLine(t *testing.T) {
	output := `Last metadata expiration check: 0:23:45 ago on Mon Dec 16 2024.
vim-enhanced.x86_64              9.0.2136-1.fc39        updates
`
	updates := parseCheckUpdate(output, nil)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].Name != "vim-enhanced" {
		t.Errorf("expected vim-enhanced, got %s", updates[0].Name)
	}
}

func TestParseSecurityInfo(t *testing.T) {
	pkgs := parseSecurityInfo(sampleSecurityInfo)

	if !pkgs["openssl-libs"] {
		t.Error("expected openssl-libs in security packages")
	}
	if !pkgs["curl"] {
		t.Error("expected curl in security packages")
	}
	if pkgs["vim-enhanced"] {
		t.Error("vim-enhanced should not be in security packages")
	}
}

func TestParseSecurityUpdates(t *testing.T) {
	updates := parseSecurityUpdates(sampleSecurityInfo)

	if len(updates) != 2 {
		t.Fatalf("expected 2 security updates, got %d", len(updates))
	}

	for _, u := range updates {
		if u.Type != checker.UpdateTypeSecurity {
			t.Errorf("expected security type, got %s for %s", u.Type, u.Name)
		}
	}
}
