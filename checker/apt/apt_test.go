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

const sampleInstOutput = `Reading package lists...
Building dependency tree...
Reading state information...
Calculating upgrade...
The following packages will be upgraded:
  curl libssl3 openssl vim
Inst libssl3 [3.0.13-0ubuntu3.1] (3.0.13-0ubuntu3.4 Ubuntu:22.04/jammy-security [amd64])
Inst openssl [3.0.13-0ubuntu3.1] (3.0.13-0ubuntu3.4 Ubuntu:22.04/jammy-security [amd64])
Inst curl [8.5.0-2ubuntu10.1] (8.5.0-2ubuntu10.6 Ubuntu:22.04/jammy-updates [amd64])
Inst vim [2:9.0.0242-1ubuntu1] (2:9.0.0242-1ubuntu1.1 Ubuntu:22.04/jammy-updates [amd64])
Conf libssl3 (3.0.13-0ubuntu3.4 Ubuntu:22.04/jammy-security [amd64])
Conf openssl (3.0.13-0ubuntu3.4 Ubuntu:22.04/jammy-security [amd64])
Conf curl (8.5.0-2ubuntu10.6 Ubuntu:22.04/jammy-updates [amd64])
Conf vim (2:9.0.0242-1ubuntu1.1 Ubuntu:22.04/jammy-updates [amd64])
4 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
`

func TestParseInstSecurity(t *testing.T) {
	security := parseInstSecurity(sampleInstOutput)

	if len(security) != 2 {
		t.Fatalf("expected 2 security packages, got %d", len(security))
	}
	if !security["libssl3"] {
		t.Error("expected libssl3 to be security")
	}
	if !security["openssl"] {
		t.Error("expected openssl to be security")
	}
	if security["curl"] {
		t.Error("curl should not be security")
	}
	if security["vim"] {
		t.Error("vim should not be security")
	}
}

func TestParseInstSecurityDebian(t *testing.T) {
	output := `Inst dpkg [1.19.7] (1.19.8 Debian-Security:10/oldstable [amd64])
Inst git [1:2.20.1-2+deb10u8] (1:2.20.1-2+deb10u9 Debian:10.13/oldstable [amd64])
`
	security := parseInstSecurity(output)

	if len(security) != 1 {
		t.Fatalf("expected 1 security package, got %d", len(security))
	}
	if !security["dpkg"] {
		t.Error("expected dpkg to be security")
	}
	if security["git"] {
		t.Error("git should not be security")
	}
}

func TestParseInstSecurityEmpty(t *testing.T) {
	security := parseInstSecurity("0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.\n")
	if len(security) != 0 {
		t.Fatalf("expected 0 security packages, got %d", len(security))
	}
}

func TestSecurityCrossCheck(t *testing.T) {
	// Simulate: apt list shows curl as regular, but Inst line reveals it's from security
	updates := []checker.Update{
		{Name: "curl", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
		{Name: "vim", Type: checker.UpdateTypeRegular, Priority: checker.PriorityNormal},
	}

	instOutput := `Inst curl [8.5.0-2] (8.5.0-6 Ubuntu:22.04/jammy-security [amd64])
Inst vim [2:9.0.0242-1ubuntu1] (2:9.0.0242-1ubuntu1.1 Ubuntu:22.04/jammy-updates [amd64])
`
	instSecurity := parseInstSecurity(instOutput)

	for i := range updates {
		if updates[i].Type != checker.UpdateTypeSecurity && instSecurity[updates[i].Name] {
			updates[i].Type = checker.UpdateTypeSecurity
			updates[i].Priority = checker.PriorityHigh
		}
	}

	if updates[0].Type != checker.UpdateTypeSecurity {
		t.Errorf("expected curl to be reclassified as security, got %s", updates[0].Type)
	}
	if updates[0].Priority != checker.PriorityHigh {
		t.Errorf("expected curl priority to be high, got %s", updates[0].Priority)
	}
	if updates[1].Type != checker.UpdateTypeRegular {
		t.Errorf("expected vim to remain regular, got %s", updates[1].Type)
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
