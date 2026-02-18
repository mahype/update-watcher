package pushbullet

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNewFromConfig(t *testing.T) {
	cfg := config.NotifierConfig{
		Type:    "pushbullet",
		Enabled: true,
		Options: map[string]interface{}{
			"access_token": "o.ABCDEF123456",
			"device_iden":  "ujpah72o0sjAoRtnM0jc",
			"channel_tag":  "updates",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PushbulletNotifier)
	if nn.accessToken != "o.ABCDEF123456" {
		t.Errorf("expected access_token, got %q", nn.accessToken)
	}
	if nn.deviceIden != "ujpah72o0sjAoRtnM0jc" {
		t.Errorf("expected device_iden, got %q", nn.deviceIden)
	}
	if nn.channelTag != "updates" {
		t.Errorf("expected channel_tag 'updates', got %q", nn.channelTag)
	}
}

func TestNewFromConfigDefaults(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{
			"access_token": "test-token",
		},
	}

	n, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nn := n.(*PushbulletNotifier)
	if nn.deviceIden != "" {
		t.Errorf("expected empty device_iden, got %q", nn.deviceIden)
	}
	if nn.channelTag != "" {
		t.Errorf("expected empty channel_tag, got %q", nn.channelTag)
	}
}

func TestNewFromConfigMissingAccessToken(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{},
	}
	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing access_token")
	}
}

func TestName(t *testing.T) {
	cfg := config.NotifierConfig{
		Options: map[string]interface{}{"access_token": "t"},
	}
	n, _ := NewFromConfig(cfg)
	if n.Name() != "pushbullet" {
		t.Errorf("expected name 'pushbullet', got %q", n.Name())
	}
}
