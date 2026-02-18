package config

import (
	"os"
	"testing"
)

func TestResolveEnvVarsSet(t *testing.T) {
	os.Setenv("TEST_RESOLVE_VAR", "hello")
	defer os.Unsetenv("TEST_RESOLVE_VAR")

	result := resolveEnvVars("${TEST_RESOLVE_VAR}")
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestResolveEnvVarsUnset(t *testing.T) {
	os.Unsetenv("TEST_RESOLVE_UNSET")

	result := resolveEnvVars("${TEST_RESOLVE_UNSET}")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestResolveEnvVarsDefaultUnset(t *testing.T) {
	os.Unsetenv("TEST_RESOLVE_DEF")

	result := resolveEnvVars("${TEST_RESOLVE_DEF:-fallback}")
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

func TestResolveEnvVarsDefaultSet(t *testing.T) {
	os.Setenv("TEST_RESOLVE_DEF2", "real-value")
	defer os.Unsetenv("TEST_RESOLVE_DEF2")

	result := resolveEnvVars("${TEST_RESOLVE_DEF2:-fallback}")
	if result != "real-value" {
		t.Errorf("expected 'real-value', got %q", result)
	}
}

func TestResolveEnvVarsNoPattern(t *testing.T) {
	result := resolveEnvVars("just a plain string")
	if result != "just a plain string" {
		t.Errorf("expected unchanged string, got %q", result)
	}
}

func TestResolveEnvVarsMultiple(t *testing.T) {
	os.Setenv("TEST_A", "foo")
	os.Setenv("TEST_B", "bar")
	defer os.Unsetenv("TEST_A")
	defer os.Unsetenv("TEST_B")

	result := resolveEnvVars("${TEST_A}:${TEST_B}")
	if result != "foo:bar" {
		t.Errorf("expected 'foo:bar', got %q", result)
	}
}

func TestResolveEnvVarsEmbedded(t *testing.T) {
	os.Setenv("TEST_HOST", "example.com")
	defer os.Unsetenv("TEST_HOST")

	result := resolveEnvVars("https://${TEST_HOST}/api")
	if result != "https://example.com/api" {
		t.Errorf("expected 'https://example.com/api', got %q", result)
	}
}

func TestResolveOptionsEnvVars(t *testing.T) {
	os.Setenv("TEST_OPT_TOKEN", "secret123")
	defer os.Unsetenv("TEST_OPT_TOKEN")

	options := map[string]interface{}{
		"token":    "${TEST_OPT_TOKEN}",
		"plain":    "no-change",
		"priority": 5,
	}

	resolveOptionsEnvVars(options)

	if options["token"] != "secret123" {
		t.Errorf("expected 'secret123', got %q", options["token"])
	}
	if options["plain"] != "no-change" {
		t.Errorf("expected 'no-change', got %q", options["plain"])
	}
}

func TestResolveOptionsEnvVarsNested(t *testing.T) {
	os.Setenv("TEST_NESTED", "resolved")
	defer os.Unsetenv("TEST_NESTED")

	options := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"name":  "site",
				"token": "${TEST_NESTED}",
			},
		},
	}

	resolveOptionsEnvVars(options)

	items := options["items"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["token"] != "resolved" {
		t.Errorf("expected 'resolved', got %q", item["token"])
	}
}
