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

const sampleDryRunOutput = `Reading package lists...
Building dependency tree...
Reading state information...
Calculating upgrade...
The following upgrades have been deferred due to phasing:
  cpp-13 cpp-13-x86-64-linux-gnu g++-13 gcc-13 gcc-13-base gcc-14-base
  libasan8 libatomic1 libstdc++6
0 upgraded, 0 newly installed, 0 to remove and 9 not upgraded.
`

func TestParseDeferredPackages(t *testing.T) {
	deferred := parseDeferredPackages(sampleDryRunOutput)

	expected := []string{
		"cpp-13", "cpp-13-x86-64-linux-gnu", "g++-13", "gcc-13",
		"gcc-13-base", "gcc-14-base", "libasan8", "libatomic1", "libstdc++6",
	}

	if len(deferred) != len(expected) {
		t.Fatalf("expected %d deferred packages, got %d", len(expected), len(deferred))
	}

	for _, name := range expected {
		if !deferred[name] {
			t.Errorf("expected %q to be in deferred set", name)
		}
	}
}

func TestParseDeferredPackagesEmpty(t *testing.T) {
	output := `Reading package lists...
Building dependency tree...
Reading state information...
Calculating upgrade...
0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
`
	deferred := parseDeferredPackages(output)
	if len(deferred) != 0 {
		t.Fatalf("expected 0 deferred packages, got %d", len(deferred))
	}
}

func TestParseDeferredPackagesMixed(t *testing.T) {
	output := `Reading package lists...
Building dependency tree...
Reading state information...
Calculating upgrade...
The following packages will be upgraded:
  curl vim
The following upgrades have been deferred due to phasing:
  gcc-13 libstdc++6
2 upgraded, 0 newly installed, 0 to remove and 2 not upgraded.
`
	deferred := parseDeferredPackages(output)

	if len(deferred) != 2 {
		t.Fatalf("expected 2 deferred packages, got %d", len(deferred))
	}
	if !deferred["gcc-13"] {
		t.Error("expected gcc-13 to be deferred")
	}
	if !deferred["libstdc++6"] {
		t.Error("expected libstdc++6 to be deferred")
	}
	if deferred["curl"] {
		t.Error("curl should not be deferred")
	}
}

func TestDeferredMarksHiddenPhased(t *testing.T) {
	// Simulate updates from apt list (no phased marker)
	listOutput := `Listing... Done
gcc-13/noble-updates 13.3.0-6ubuntu2~24.04.1 amd64 [upgradable from: 13.3.0-6ubuntu2~24.04]
curl/noble-updates 8.5.0-2ubuntu10.6 amd64 [upgradable from: 8.5.0-2ubuntu10.1]
`
	updates := parseUpgradable(listOutput, false)

	// Simulate dry-run output marking gcc-13 as deferred
	dryRunOutput := `The following packages will be upgraded:
  curl
The following upgrades have been deferred due to phasing:
  gcc-13
1 upgraded, 0 newly installed, 0 to remove and 1 not upgraded.
`
	deferred := parseDeferredPackages(dryRunOutput)

	// Apply deferred detection
	for i := range updates {
		if updates[i].Phasing == "" && deferred[updates[i].Name] {
			updates[i].Phasing = "deferred"
		}
	}

	// gcc-13 should now be marked as phased
	if updates[0].Phasing != "deferred" {
		t.Errorf("expected gcc-13 to be marked as deferred, got %q", updates[0].Phasing)
	}
	// curl should remain unphased
	if updates[1].Phasing != "" {
		t.Errorf("expected curl to have no phasing, got %q", updates[1].Phasing)
	}
}

func TestHidePhased(t *testing.T) {
	updates := []checker.Update{
		{Name: "curl", Phasing: ""},
		{Name: "gcc-13", Phasing: "deferred"},
		{Name: "libssl3", Phasing: "50%"},
		{Name: "vim", Phasing: ""},
	}

	// Simulate hide_phased filter
	filtered := updates[:0]
	for _, u := range updates {
		if u.Phasing == "" {
			filtered = append(filtered, u)
		}
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 non-phased updates, got %d", len(filtered))
	}
	if filtered[0].Name != "curl" {
		t.Errorf("expected first to be curl, got %s", filtered[0].Name)
	}
	if filtered[1].Name != "vim" {
		t.Errorf("expected second to be vim, got %s", filtered[1].Name)
	}
}
