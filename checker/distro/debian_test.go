package distro

import "testing"

func TestParseDebianReleaseVersion(t *testing.T) {
	content := `Origin: Debian
Label: Debian
Suite: stable
Version: 12.5
Codename: bookworm
Date: Sat, 10 Feb 2024 10:47:52 UTC
Architectures: all amd64 arm64 armel armhf i386 mips64el mipsel ppc64el s390x`

	got := parseDebianReleaseVersion(content)
	want := "12"
	if got != want {
		t.Errorf("parseDebianReleaseVersion() = %q, want %q", got, want)
	}
}

func TestParseDebianReleaseVersion_NoVersion(t *testing.T) {
	content := `Origin: Debian
Label: Debian
Suite: stable
Codename: bookworm`

	got := parseDebianReleaseVersion(content)
	if got != "" {
		t.Errorf("parseDebianReleaseVersion() = %q, want empty", got)
	}
}

func TestParseMajorVersion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"12", 12},
		{"12.5", 12},
		{"22.04", 22},
		{"3.19.0", 3},
		{"40", 40},
	}

	for _, tt := range tests {
		got, err := parseMajorVersion(tt.input)
		if err != nil {
			t.Errorf("parseMajorVersion(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseMajorVersion(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseMajorVersion_Invalid(t *testing.T) {
	_, err := parseMajorVersion("rolling")
	if err == nil {
		t.Error("parseMajorVersion(\"rolling\") should return error")
	}
}
