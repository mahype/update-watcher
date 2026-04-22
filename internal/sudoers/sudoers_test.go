package sudoers

import (
	"errors"
	"strings"
	"testing"

	"github.com/mahype/update-watcher/config"
)

// fakeLookPath builds a LookPathFunc that serves from a static map.
// Names not in the map return an error.
func fakeLookPath(paths map[string]string) LookPathFunc {
	return func(name string) (string, error) {
		if p, ok := paths[name]; ok {
			return p, nil
		}
		return "", errors.New("not found: " + name)
	}
}

func TestBuild_AptWithSudo(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "apt", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
		},
	}

	rules, warnings, err := BuildWith(cfg, fakeLookPath(map[string]string{"apt-get": "/usr/bin/apt-get"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d: %+v", len(rules), rules)
	}
	want := Rule{RunAs: "root", Command: "/usr/bin/apt-get", Args: []string{"update"}}
	if rules[0].RunAs != want.RunAs || rules[0].Command != want.Command ||
		strings.Join(rules[0].Args, " ") != strings.Join(want.Args, " ") {
		t.Fatalf("rule mismatch: got %+v want %+v", rules[0], want)
	}
}

func TestBuild_AptSudoDisabled(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "apt", Enabled: true, Options: config.OptionsMap{"use_sudo": false}},
		},
	}
	rules, _, err := BuildWith(cfg, fakeLookPath(map[string]string{"apt-get": "/usr/bin/apt-get"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules when use_sudo is false, got %+v", rules)
	}
}

func TestBuild_DisabledWatcherSkipped(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "apt", Enabled: false, Options: config.OptionsMap{"use_sudo": true}},
		},
	}
	rules, _, err := BuildWith(cfg, fakeLookPath(map[string]string{"apt-get": "/usr/bin/apt-get"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules for disabled watcher, got %+v", rules)
	}
}

func TestBuild_WordPressMultipleSitesDedup(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{
				Type:    "wordpress",
				Enabled: true,
				Options: config.OptionsMap{
					"sites": []interface{}{
						map[string]interface{}{"name": "a", "path": "/var/www/a", "run_as": "www-data"},
						map[string]interface{}{"name": "b", "path": "/var/www/b", "run_as": "www-data"},
						map[string]interface{}{"name": "c", "path": "/var/www/c", "run_as": "nginx"},
					},
				},
			},
		},
	}
	rules, warnings, err := BuildWith(cfg, fakeLookPath(map[string]string{"wp": "/usr/local/bin/wp"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 deduplicated rules (www-data, nginx), got %d: %+v", len(rules), rules)
	}
	users := map[string]bool{}
	for _, r := range rules {
		users[r.RunAs] = true
		if r.Command != "/usr/local/bin/wp" {
			t.Errorf("expected wp path /usr/local/bin/wp, got %q", r.Command)
		}
		if len(r.Args) != 0 {
			t.Errorf("expected no args restriction for wp-cli, got %v", r.Args)
		}
	}
	if !users["www-data"] || !users["nginx"] {
		t.Errorf("expected rules for www-data and nginx, got %+v", users)
	}
}

func TestBuild_WordPressNoRunAs(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{
				Type:    "wordpress",
				Enabled: true,
				Options: config.OptionsMap{
					"sites": []interface{}{
						map[string]interface{}{"name": "a", "path": "/var/www/a"},
					},
				},
			},
		},
	}
	rules, warnings, err := BuildWith(cfg, fakeLookPath(map[string]string{"wp": "/usr/local/bin/wp"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules when no site has run_as, got %+v", rules)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings when no run_as configured, got %v", warnings)
	}
}

func TestBuild_BinaryNotFoundEmitsWarning(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "apt", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
			{
				Type:    "wordpress",
				Enabled: true,
				Options: config.OptionsMap{
					"sites": []interface{}{
						map[string]interface{}{"name": "a", "path": "/var/www/a", "run_as": "www-data"},
					},
				},
			},
		},
	}
	rules, warnings, err := BuildWith(cfg, fakeLookPath(map[string]string{})) // nothing in PATH
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules when binaries missing, got %+v", rules)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings (apt-get, wp), got %v", warnings)
	}
}

func TestBuild_DnfZypperPacmanApk(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "dnf", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
			{Type: "zypper", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
			{Type: "pacman", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
			{Type: "apk", Enabled: true, Options: config.OptionsMap{"use_sudo": true}},
		},
	}
	paths := map[string]string{
		"dnf":    "/usr/bin/dnf",
		"zypper": "/usr/bin/zypper",
		"pacman": "/usr/bin/pacman",
		"apk":    "/sbin/apk",
	}
	rules, _, err := BuildWith(cfg, fakeLookPath(paths))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// dnf: 2 rules, zypper: 3 rules, pacman: 1 rule, apk: 1 rule → 7 total
	if len(rules) != 7 {
		t.Fatalf("expected 7 rules, got %d: %+v", len(rules), rules)
	}
}

func TestBuild_ApkDefaultSudoIsFalse(t *testing.T) {
	cfg := &config.Config{
		Watchers: []config.WatcherConfig{
			{Type: "apk", Enabled: true, Options: config.OptionsMap{}}, // no use_sudo key
		},
	}
	rules, _, err := BuildWith(cfg, fakeLookPath(map[string]string{"apk": "/sbin/apk"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("apk default should be use_sudo=false, got rules: %+v", rules)
	}
}

func TestRender(t *testing.T) {
	rules := []Rule{
		{RunAs: "www-data", Command: "/usr/local/bin/wp"},
		{RunAs: "root", Command: "/usr/bin/apt-get", Args: []string{"update"}},
	}
	out := Render(rules, "update-watcher")

	if !strings.Contains(out, "# Managed by update-watcher") {
		t.Errorf("missing header comment, got:\n%s", out)
	}
	// root should come before www-data (alphabetical by RunAs)
	rootIdx := strings.Index(out, "update-watcher ALL=(root) NOPASSWD: /usr/bin/apt-get update")
	wwwIdx := strings.Index(out, "update-watcher ALL=(www-data) NOPASSWD: /usr/local/bin/wp")
	if rootIdx < 0 || wwwIdx < 0 {
		t.Fatalf("expected both rules in output, got:\n%s", out)
	}
	if rootIdx > wwwIdx {
		t.Errorf("expected root rule before www-data rule, got:\n%s", out)
	}
}

func TestRender_Empty(t *testing.T) {
	out := Render(nil, "update-watcher")
	if !strings.Contains(out, "# Managed by update-watcher") {
		t.Errorf("empty render should still contain header, got:\n%s", out)
	}
	// Actual rules start with the service user name at the beginning of a line.
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "update-watcher ") {
			t.Errorf("empty render should have no rule lines, found: %q", line)
		}
	}
}

func TestFormatRule(t *testing.T) {
	cases := []struct {
		name string
		rule Rule
		want string
	}{
		{
			name: "with args",
			rule: Rule{RunAs: "root", Command: "/usr/bin/apt-get", Args: []string{"update"}},
			want: "update-watcher ALL=(root) NOPASSWD: /usr/bin/apt-get update",
		},
		{
			name: "no args",
			rule: Rule{RunAs: "www-data", Command: "/usr/local/bin/wp"},
			want: "update-watcher ALL=(www-data) NOPASSWD: /usr/local/bin/wp",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatRule("update-watcher", tc.rule)
			if got != tc.want {
				t.Errorf("FormatRule: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuild_NilConfig(t *testing.T) {
	rules, warnings, err := BuildWith(nil, fakeLookPath(map[string]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 || len(warnings) != 0 {
		t.Errorf("expected empty for nil config, got rules=%+v warnings=%+v", rules, warnings)
	}
}
