package output

import (
	"testing"

	"github.com/mahype/update-watcher/config"
)

func TestNotifierHost(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.NotifierConfig
		want string
	}{
		{
			name: "slack webhook_url",
			cfg: config.NotifierConfig{
				Type:    "slack",
				Options: config.OptionsMap{"webhook_url": "https://hooks.slack.com/services/T00/B00/xxx"},
			},
			want: "hooks.slack.com",
		},
		{
			name: "discord webhook_url",
			cfg: config.NotifierConfig{
				Type:    "discord",
				Options: config.OptionsMap{"webhook_url": "https://discord.com/api/webhooks/123/abc"},
			},
			want: "discord.com",
		},
		{
			name: "gotify server_url",
			cfg: config.NotifierConfig{
				Type:    "gotify",
				Options: config.OptionsMap{"server_url": "https://gotify.example.com"},
			},
			want: "gotify.example.com",
		},
		{
			name: "ntfy server_url",
			cfg: config.NotifierConfig{
				Type:    "ntfy",
				Options: config.OptionsMap{"server_url": "https://ntfy.example.com", "topic": "updates"},
			},
			want: "ntfy.example.com",
		},
		{
			name: "ntfy topic fallback",
			cfg: config.NotifierConfig{
				Type:    "ntfy",
				Options: config.OptionsMap{"topic": "my-updates"},
			},
			want: "my-updates",
		},
		{
			name: "webhook url",
			cfg: config.NotifierConfig{
				Type:    "webhook",
				Options: config.OptionsMap{"url": "https://example.com/hook"},
			},
			want: "example.com",
		},
		{
			name: "email smtp_host",
			cfg: config.NotifierConfig{
				Type:    "email",
				Options: config.OptionsMap{"smtp_host": "mail.example.com"},
			},
			want: "mail.example.com",
		},
		{
			name: "telegram chat_id",
			cfg: config.NotifierConfig{
				Type:    "telegram",
				Options: config.OptionsMap{"chat_id": "-1001234567"},
			},
			want: "-1001234567",
		},
		{
			name: "matrix homeserver",
			cfg: config.NotifierConfig{
				Type:    "matrix",
				Options: config.OptionsMap{"homeserver": "https://matrix.org"},
			},
			want: "matrix.org",
		},
		{
			name: "updatewall url",
			cfg: config.NotifierConfig{
				Type:    "updatewall",
				Options: config.OptionsMap{"url": "https://wall.example.com/api"},
			},
			want: "wall.example.com",
		},
		{
			name: "pushover has no host",
			cfg: config.NotifierConfig{
				Type:    "pushover",
				Options: config.OptionsMap{"app_token": "xxx"},
			},
			want: "",
		},
		{
			name: "env var not resolved",
			cfg: config.NotifierConfig{
				Type:    "slack",
				Options: config.OptionsMap{"webhook_url": "${SLACK_WEBHOOK_URL}"},
			},
			want: "",
		},
		{
			name: "empty options",
			cfg: config.NotifierConfig{
				Type:    "slack",
				Options: config.OptionsMap{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := notifierHost(tt.cfg)
			if got != tt.want {
				t.Errorf("notifierHost() = %q, want %q", got, tt.want)
			}
		})
	}
}
