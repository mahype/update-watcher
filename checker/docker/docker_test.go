package docker

import (
	"testing"
)

func TestShortDigest(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sha256:abc123def456789", "abc123def456"},
		{"abc123def456789", "abc123def456"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		result := shortDigest(tt.input)
		if result != tt.expected {
			t.Errorf("shortDigest(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"nginx", "redis", "postgres"}

	if !containsString(slice, "nginx") {
		t.Error("expected to find nginx")
	}
	if containsString(slice, "mysql") {
		t.Error("did not expect to find mysql")
	}
	if containsString(nil, "anything") {
		t.Error("nil slice should not contain anything")
	}
}
