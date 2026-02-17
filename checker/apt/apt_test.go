package apt

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = `Listing... Done
libssl3/jammy-security 3.0.13-0ubuntu3.4 amd64 [upgradable from: 3.0.13-0ubuntu3.1]
openssl/jammy-security 3.0.13-0ubuntu3.4 amd64 [upgradable from: 3.0.13-0ubuntu3.1]
curl/jammy-updates 8.5.0-2ubuntu10.6 amd64 [upgradable from: 8.5.0-2ubuntu10.1]
git/jammy-updates 1:2.43.0-1ubuntu7.2 amd64 [upgradable from: 1:2.43.0-1ubuntu7.1] [phased 50%]
vim/jammy-updates 2:9.0.0242-1ubuntu1.1 amd64 [upgradable from: 2:9.0.0242-1ubuntu1]
`

func TestParseUpgradable(t *testing.T) {
	updates := parseUpgradable(sampleOutput, false)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	// First should be security
	if updates[0].Name != "libssl3" {
		t.Errorf("expected first package to be libssl3, got %s", updates[0].Name)
	}
	if updates[0].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected libssl3 to be security update, got %s", updates[0].Type)
	}
	if updates[0].CurrentVersion != "3.0.13-0ubuntu3.1" {
		t.Errorf("unexpected current version: %s", updates[0].CurrentVersion)
	}
	if updates[0].NewVersion != "3.0.13-0ubuntu3.4" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}

	// curl should be regular, no phasing
	if updates[2].Name != "curl" {
		t.Errorf("expected third package to be curl, got %s", updates[2].Name)
	}
	if updates[2].Type != checker.UpdateTypeRegular {
		t.Errorf("expected curl to be regular update, got %s", updates[2].Type)
	}
	if updates[2].Phasing != "" {
		t.Errorf("expected curl to have no phasing, got %s", updates[2].Phasing)
	}

	// git should be phased
	if updates[3].Name != "git" {
		t.Errorf("expected fourth package to be git, got %s", updates[3].Name)
	}
	if updates[3].Phasing != "50%" {
		t.Errorf("expected git phasing to be 50%%, got %q", updates[3].Phasing)
	}
}

func TestParseUpgradableSecurityOnly(t *testing.T) {
	updates := parseUpgradable(sampleOutput, true)

	if len(updates) != 2 {
		t.Fatalf("expected 2 security updates, got %d", len(updates))
	}

	for _, u := range updates {
		if u.Type != checker.UpdateTypeSecurity {
			t.Errorf("expected only security updates, got %s for %s", u.Type, u.Name)
		}
	}
}

func TestParseUpgradableEmpty(t *testing.T) {
	updates := parseUpgradable("Listing... Done\n", false)
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}
