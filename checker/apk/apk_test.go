package apk

import (
	"testing"

	"github.com/mahype/update-watcher/checker"
)

const sampleOutput = `Installed:                          Available:
busybox-1.36.1-r15                < 1.36.1-r19
curl-8.5.0-r0                     < 8.6.0-r0
openssl-3.1.4-r2                  < 3.1.4-r5
musl-1.2.4-r4                     < 1.2.5-r0
zlib-1.3-r2                       < 1.3.1-r0
`

func TestParseVersionOutput(t *testing.T) {
	updates := parseVersionOutput(sampleOutput)

	if len(updates) != 5 {
		t.Fatalf("expected 5 updates, got %d", len(updates))
	}

	if updates[0].Name != "busybox" {
		t.Errorf("expected first package to be busybox, got %s", updates[0].Name)
	}
	if updates[0].CurrentVersion != "1.36.1-r15" {
		t.Errorf("unexpected current version: %s", updates[0].CurrentVersion)
	}
	if updates[0].NewVersion != "1.36.1-r19" {
		t.Errorf("unexpected new version: %s", updates[0].NewVersion)
	}
	if updates[0].Type != checker.UpdateTypeRegular {
		t.Errorf("expected regular update type, got %s", updates[0].Type)
	}
}

func TestParseVersionOutputEmpty(t *testing.T) {
	updates := parseVersionOutput("")
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}

func TestParseVersionOutputNoHeader(t *testing.T) {
	output := `busybox-1.36.1-r15 < 1.36.1-r19
curl-8.5.0-r0 < 8.6.0-r0
`
	updates := parseVersionOutput(output)
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}
	if updates[0].Name != "busybox" {
		t.Errorf("expected busybox, got %s", updates[0].Name)
	}
	if updates[1].Name != "curl" {
		t.Errorf("expected curl, got %s", updates[1].Name)
	}
}

func TestParseVersionOutputHeaderOnly(t *testing.T) {
	output := "Installed:                          Available:\n"
	updates := parseVersionOutput(output)
	if len(updates) != 0 {
		t.Fatalf("expected 0 updates, got %d", len(updates))
	}
}
