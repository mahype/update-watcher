package wordpress

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapWPCLIError_SudoPermissionErrors(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
	}{
		{"password required", "sudo: a password is required"},
		{"terminal required", "sudo: a terminal is required to read the password"},
		{"not allowed", "user update-watcher is not allowed to execute"},
		{"not in sudoers", "Sorry, user update-watcher is not in the sudoers file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the wrapped error that wpcli.Run produces.
			inner := errors.New("exit status 1")
			fullErr := errors.New("wp-cli failed: " + inner.Error() + " (stderr: " + tc.stderr + ")")

			wrapped := wrapWPCLIError(fullErr, "www-data")
			if wrapped == nil {
				t.Fatal("expected non-nil wrapped error")
			}
			msg := wrapped.Error()
			if !strings.Contains(msg, "missing sudoers rule") {
				t.Errorf("expected 'missing sudoers rule' in error, got: %s", msg)
			}
			if !strings.Contains(msg, "www-data") {
				t.Errorf("expected run_as user 'www-data' in error, got: %s", msg)
			}
			if !strings.Contains(msg, "install-cron") {
				t.Errorf("expected actionable hint about 'install-cron' in error, got: %s", msg)
			}
		})
	}
}

func TestWrapWPCLIError_NonSudoErrorUnchanged(t *testing.T) {
	inner := errors.New("wp-cli failed: exit status 1 (stderr: Error: 'core check-update' is not a registered subcommand)")
	wrapped := wrapWPCLIError(inner, "www-data")
	if wrapped == nil || wrapped.Error() != inner.Error() {
		t.Errorf("expected unchanged error for non-sudo failures, got: %v", wrapped)
	}
}

func TestWrapWPCLIError_NoRunAsUnchanged(t *testing.T) {
	// Without run_as the message cannot point at a missing sudoers rule.
	inner := errors.New("wp-cli failed: exit status 1 (stderr: sudo: a password is required)")
	wrapped := wrapWPCLIError(inner, "")
	if wrapped == nil || wrapped.Error() != inner.Error() {
		t.Errorf("expected unchanged error when run_as is empty, got: %v", wrapped)
	}
}

func TestWrapWPCLIError_NilUnchanged(t *testing.T) {
	if wrapWPCLIError(nil, "www-data") != nil {
		t.Error("expected nil for nil input")
	}
}
