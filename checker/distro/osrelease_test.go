package distro

import "testing"

func TestParseOSReleaseContent_Ubuntu(t *testing.T) {
	content := `NAME="Ubuntu"
VERSION="22.04.4 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 22.04.4 LTS"
VERSION_ID="22.04"
VERSION_CODENAME=jammy
`
	rel := ParseOSReleaseContent(content)

	if rel.ID != "ubuntu" {
		t.Errorf("ID = %q, want %q", rel.ID, "ubuntu")
	}
	if rel.IDLike != "debian" {
		t.Errorf("IDLike = %q, want %q", rel.IDLike, "debian")
	}
	if rel.VersionID != "22.04" {
		t.Errorf("VersionID = %q, want %q", rel.VersionID, "22.04")
	}
	if rel.VersionCodename != "jammy" {
		t.Errorf("VersionCodename = %q, want %q", rel.VersionCodename, "jammy")
	}
	if rel.PrettyName != "Ubuntu 22.04.4 LTS" {
		t.Errorf("PrettyName = %q, want %q", rel.PrettyName, "Ubuntu 22.04.4 LTS")
	}
	if rel.Name != "Ubuntu" {
		t.Errorf("Name = %q, want %q", rel.Name, "Ubuntu")
	}
}

func TestParseOSReleaseContent_Debian(t *testing.T) {
	content := `PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
NAME="Debian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=debian
`
	rel := ParseOSReleaseContent(content)

	if rel.ID != "debian" {
		t.Errorf("ID = %q, want %q", rel.ID, "debian")
	}
	if rel.VersionID != "12" {
		t.Errorf("VersionID = %q, want %q", rel.VersionID, "12")
	}
	if rel.VersionCodename != "bookworm" {
		t.Errorf("VersionCodename = %q, want %q", rel.VersionCodename, "bookworm")
	}
}

func TestParseOSReleaseContent_Fedora(t *testing.T) {
	content := `NAME="Fedora Linux"
VERSION="40 (Workstation Edition)"
ID=fedora
VERSION_ID=40
PRETTY_NAME="Fedora Linux 40 (Workstation Edition)"
`
	rel := ParseOSReleaseContent(content)

	if rel.ID != "fedora" {
		t.Errorf("ID = %q, want %q", rel.ID, "fedora")
	}
	if rel.VersionID != "40" {
		t.Errorf("VersionID = %q, want %q", rel.VersionID, "40")
	}
	if rel.Name != "Fedora Linux" {
		t.Errorf("Name = %q, want %q", rel.Name, "Fedora Linux")
	}
}

func TestParseOSReleaseContent_EmptyAndComments(t *testing.T) {
	content := `# This is a comment

ID=arch

# Another comment
VERSION_ID=
`
	rel := ParseOSReleaseContent(content)

	if rel.ID != "arch" {
		t.Errorf("ID = %q, want %q", rel.ID, "arch")
	}
	if rel.VersionID != "" {
		t.Errorf("VersionID = %q, want empty", rel.VersionID)
	}
}

func TestParseOSReleaseContent_Minimal(t *testing.T) {
	content := `ID=alpine
VERSION_ID=3.19.0
`
	rel := ParseOSReleaseContent(content)

	if rel.ID != "alpine" {
		t.Errorf("ID = %q, want %q", rel.ID, "alpine")
	}
	if rel.VersionID != "3.19.0" {
		t.Errorf("VersionID = %q, want %q", rel.VersionID, "3.19.0")
	}
	if rel.Name != "" {
		t.Errorf("Name = %q, want empty", rel.Name)
	}
}
