package matrix

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "matrix",
		Enabled: true,
		Options: map[string]interface{}{
			"homeserver":   "https://matrix.example.com",
			"access_token": "syt_test_token",
			"room_id":      "!abc123:matrix.example.com",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*MatrixNotifier)
	if nn.homeserver != "https://matrix.example.com" {
		t.Errorf("expected homeserver, got %q", nn.homeserver)
	}
	if nn.accessToken != "syt_test_token" {
		t.Errorf("expected access_token, got %q", nn.accessToken)
	}
	if nn.roomID != "!abc123:matrix.example.com" {
		t.Errorf("expected room_id, got %q", nn.roomID)
	}
}

func TestNewFromConfigTrailingSlash(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"homeserver":   "https://matrix.org/",
			"access_token": "token",
			"room_id":      "!room:matrix.org",
		},
	}

	n, _ := NewFromConfig(cfg)
	nn := n.(*MatrixNotifier)
	if nn.homeserver != "https://matrix.org" {
		t.Errorf("expected trailing slash trimmed, got %q", nn.homeserver)
	}
}

func TestNewFromConfigMissingHomeserver(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"access_token": "token",
			"room_id":      "!room:matrix.org",
		},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing homeserver")
	}
}

func TestNewFromConfigMissingAccessToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"homeserver": "https://matrix.org",
			"room_id":    "!room:matrix.org",
		},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing access_token")
	}
}

func TestNewFromConfigMissingRoomID(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"homeserver":   "https://matrix.org",
			"access_token": "token",
		},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing room_id")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"homeserver":   "https://matrix.org",
			"access_token": "t",
			"room_id":      "!r:m.org",
		},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "matrix" {
		t.Errorf("expected name 'matrix', got %q", n.Name())
	}
}
