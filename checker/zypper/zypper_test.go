package zypper

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleListUpdates = `Loading repository data...
Reading installed packages...

S | Repository | Name       | Current Version | Available Version | Arch
--+------------+------------+-----------------+-------------------+-------
v | repo-oss   | vim        | 9.0.1234        | 9.0.1500          | x86_64
v | repo-oss   | curl       | 8.4.0           | 8.5.0             | x86_64
v | repo-oss   | openssl    | 3.1.3           | 3.1.4             | x86_64
v | repo-oss   | kernel-default | 6.6.1       | 6.6.3             | x86_64
v | repo-oss   | git        | 2.43.0          | 2.43.1            | x86_64
`

const sampleSecurityPatches = `Loading repository data...
Reading installed packages...

Repository  | Name           | Category | Severity  | Interactive | Status | Summary
------------+----------------+----------+-----------+-------------+--------+-----------------------------------
repo-update | openSUSE-SU-1  | security | important | ---         | needed | Security update for openssl
repo-update | openSUSE-SU-2  | security | important | ---         | needed | Security update for curl
repo-update | openSUSE-SU-3  | security | moderate  | ---         | applied | Security update for vim
`

func TestParseListUpdates(t *testing.T) {
	updates := parseListUpdates(sampleListUpdates, nil)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	if updates[0].Name != "vim" {
		t.Errorf("expected first package to be vim, got %s", updates[0].Name)
	}
	if updates[0].CurrentVersion != "9.0.1234" {
		t.Errorf("unexpected current version: %s", updates[0].CurrentVersion)
	}
	if updates[0].NewVersion != "9.0.1500" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}
}

func TestParseListUpdatesWithSecurity(t *testing.T) {
	secPkgs := parseSecurityPatches(sampleSecurityPatches)
	updates := parseListUpdates(sampleListUpdates, secPkgs)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	// Check that vim is NOT security (patch is "applied", not "needed")
	for _, u := range updates {
		if u.Name == "vim" && u.Type == checker.UpdateTypeSecurity {
			t.Error("vim should not be security (patch already applied)")
		}
	}
}

func TestParseListUpdatesEmpty(t *testing.T) {
	output := `Loading repository data...
Reading installed packages...
No updates found.
`
	updates := parseListUpdates(output, nil)
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseSecurityPatches(t *testing.T) {
	pkgs := parseSecurityPatches(sampleSecurityPatches)

	// Only "needed" patches should be included
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 needed security patches, got %d", len(pkgs))
	}

	if !pkgs["openSUSE-SU-1"] {
		t.Error("expected openSUSE-SU-1 in security packages")
	}
	if !pkgs["openSUSE-SU-2"] {
		t.Error("expected openSUSE-SU-2 in security packages")
	}
	// vim patch is "applied", should not be included
	if pkgs["openSUSE-SU-3"] {
		t.Error("openSUSE-SU-3 should not be in security packages (applied)")
	}
}
