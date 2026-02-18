package distro

import "testing"

func TestParseLatestFedoraRelease(t *testing.T) {
	data := []byte(`[
		{"version": "39", "stable": "2023-11-07"},
		{"version": "40", "stable": "2024-04-23"},
		{"version": "41", "stable": "2024-10-29"},
		{"version": "42", "stable": ""}
	]`)

	got, err := parseLatestFedoraRelease(data)
	if err != nil {
		t.Fatalf("parseLatestFedoraRelease() error: %v", err)
	}
	want := 41
	if got != want {
		t.Errorf("parseLatestFedoraRelease() = %d, want %d", got, want)
	}
}

func TestParseLatestFedoraRelease_AllStable(t *testing.T) {
	data := []byte(`[
		{"version": "39", "stable": "2023-11-07"},
		{"version": "40", "stable": "2024-04-23"}
	]`)

	got, err := parseLatestFedoraRelease(data)
	if err != nil {
		t.Fatalf("parseLatestFedoraRelease() error: %v", err)
	}
	if got != 40 {
		t.Errorf("parseLatestFedoraRelease() = %d, want 40", got)
	}
}

func TestParseLatestFedoraRelease_NoStable(t *testing.T) {
	data := []byte(`[
		{"version": "42", "stable": ""},
		{"version": "43", "stable": ""}
	]`)

	_, err := parseLatestFedoraRelease(data)
	if err == nil {
		t.Error("parseLatestFedoraRelease() should return error when no stable releases")
	}
}

func TestParseLatestFedoraRelease_Empty(t *testing.T) {
	data := []byte(`[]`)

	_, err := parseLatestFedoraRelease(data)
	if err == nil {
		t.Error("parseLatestFedoraRelease() should return error for empty list")
	}
}

func TestParseLatestFedoraRelease_InvalidJSON(t *testing.T) {
	data := []byte(`invalid json`)

	_, err := parseLatestFedoraRelease(data)
	if err == nil {
		t.Error("parseLatestFedoraRelease() should return error for invalid JSON")
	}
}
