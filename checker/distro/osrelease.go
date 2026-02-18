package distro

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// OSRelease holds parsed fields from /etc/os-release.
type OSRelease struct {
	ID              string // e.g. "ubuntu", "debian", "fedora"
	IDLike          string // e.g. "debian ubuntu" (space-separated)
	VersionID       string // e.g. "22.04", "12", "40"
	VersionCodename string // e.g. "jammy", "bookworm"
	PrettyName      string // e.g. "Ubuntu 22.04.4 LTS"
	Name            string // e.g. "Ubuntu"
}

// ParseOSRelease reads and parses an os-release file at the given path.
func ParseOSRelease(path string) (OSRelease, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return OSRelease{}, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseOSReleaseContent(string(data)), nil
}

// ParseOSReleaseContent parses os-release content from a string.
func ParseOSReleaseContent(content string) OSRelease {
	var rel OSRelease
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = unquote(value)
		switch key {
		case "ID":
			rel.ID = value
		case "ID_LIKE":
			rel.IDLike = value
		case "VERSION_ID":
			rel.VersionID = value
		case "VERSION_CODENAME":
			rel.VersionCodename = value
		case "PRETTY_NAME":
			rel.PrettyName = value
		case "NAME":
			rel.Name = value
		}
	}
	return rel
}

// unquote removes surrounding double quotes from a value.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
