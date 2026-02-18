package distro

import "testing"

func TestParseUbuntuUpgradeOutput_Available(t *testing.T) {
	output := `Checking for a new Ubuntu release
New release '24.04 LTS' available.
Run 'do-release-upgrade' to upgrade to it.`

	got := parseUbuntuUpgradeOutput(output)
	want := "24.04 LTS"
	if got != want {
		t.Errorf("parseUbuntuUpgradeOutput() = %q, want %q", got, want)
	}
}

func TestParseUbuntuUpgradeOutput_NonLTS(t *testing.T) {
	output := `Checking for a new Ubuntu release
New release '23.10' available.
Run 'do-release-upgrade' to upgrade to it.`

	got := parseUbuntuUpgradeOutput(output)
	want := "23.10"
	if got != want {
		t.Errorf("parseUbuntuUpgradeOutput() = %q, want %q", got, want)
	}
}

func TestParseUbuntuUpgradeOutput_NoUpgrade(t *testing.T) {
	output := `Checking for a new Ubuntu release
No new release found.`

	got := parseUbuntuUpgradeOutput(output)
	if got != "" {
		t.Errorf("parseUbuntuUpgradeOutput() = %q, want empty", got)
	}
}

func TestParseUbuntuUpgradeOutput_Empty(t *testing.T) {
	got := parseUbuntuUpgradeOutput("")
	if got != "" {
		t.Errorf("parseUbuntuUpgradeOutput() = %q, want empty", got)
	}
}
