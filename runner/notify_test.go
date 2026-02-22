package runner

import (
	"fmt"
	"testing"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

// shouldNotify is the pure decision logic extracted for testability.
// It returns whether the notifier should fire and the filtered results.
func shouldNotify(
	nCfg config.NotifierConfig,
	globalPolicy string,
	globalMinPriority string,
	results []*checker.CheckResult,
	errors []error,
	notifyOverride *bool,
) (bool, []*checker.CheckResult, int) {
	effectivePolicy := nCfg.SendPolicy
	if effectivePolicy == "" {
		effectivePolicy = globalPolicy
	}

	effectiveMinPriority := nCfg.MinPriority
	if effectiveMinPriority == "" {
		effectiveMinPriority = globalMinPriority
	}

	filtered, total := checker.FilterResultsByPriority(results, effectiveMinPriority)

	if notifyOverride == nil {
		if effectivePolicy == "only-on-updates" && total == 0 && len(errors) == 0 {
			return false, filtered, total
		}
	}

	return true, filtered, total
}

func boolPtr(b bool) *bool { return &b }

func TestShouldNotify_PerNotifierPolicy(t *testing.T) {
	results := []*checker.CheckResult{
		{CheckerName: "apt", Updates: []checker.Update{{Name: "curl", Priority: "normal"}}},
	}

	// Notifier with "always" policy should send even with no updates.
	send, _, _ := shouldNotify(
		config.NotifierConfig{SendPolicy: "always"},
		"only-on-updates", "", nil, nil, nil,
	)
	if !send {
		t.Error("expected send=true for notifier with always policy and no updates")
	}

	// Notifier without policy falls back to global "only-on-updates".
	send, _, _ = shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "", nil, nil, nil,
	)
	if send {
		t.Error("expected send=false for empty notifier policy with global only-on-updates and no updates")
	}

	// Notifier with "only-on-updates" should send when updates exist.
	send, _, total := shouldNotify(
		config.NotifierConfig{SendPolicy: "only-on-updates"},
		"always", "", results, nil, nil,
	)
	if !send || total != 1 {
		t.Errorf("expected send=true, total=1; got send=%v, total=%d", send, total)
	}
}

func TestShouldNotify_MinPriorityFilter(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Updates: []checker.Update{
				{Name: "curl", Priority: "critical"},
				{Name: "vim", Priority: "low"},
			},
		},
	}

	// Per-notifier min_priority=high should filter out "low".
	send, filtered, total := shouldNotify(
		config.NotifierConfig{MinPriority: "high"},
		"only-on-updates", "", results, nil, nil,
	)
	if !send || total != 1 {
		t.Errorf("expected send=true, total=1; got send=%v, total=%d", send, total)
	}
	if filtered[0].Updates[0].Name != "curl" {
		t.Errorf("expected curl, got %s", filtered[0].Updates[0].Name)
	}

	// Per-notifier min_priority=critical should filter out both low and vim.
	send, _, total = shouldNotify(
		config.NotifierConfig{MinPriority: "critical"},
		"only-on-updates", "", results, nil, nil,
	)
	if !send || total != 1 {
		t.Errorf("expected send=true, total=1; got send=%v, total=%d", send, total)
	}
}

func TestShouldNotify_MinPriorityFallbackToGlobal(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Updates: []checker.Update{
				{Name: "curl", Priority: "high"},
				{Name: "vim", Priority: "low"},
			},
		},
	}

	// No per-notifier min_priority, global min_priority=high.
	send, _, total := shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "high", results, nil, nil,
	)
	if !send || total != 1 {
		t.Errorf("expected send=true, total=1; got send=%v, total=%d", send, total)
	}

	// Per-notifier min_priority overrides global.
	send, _, total = shouldNotify(
		config.NotifierConfig{MinPriority: "low"},
		"only-on-updates", "high", results, nil, nil,
	)
	if !send || total != 2 {
		t.Errorf("expected send=true, total=2; got send=%v, total=%d", send, total)
	}
}

func TestShouldNotify_FilteredOutSkipsOnlyOnUpdates(t *testing.T) {
	results := []*checker.CheckResult{
		{
			CheckerName: "apt",
			Updates: []checker.Update{
				{Name: "vim", Priority: "low"},
			},
		},
	}

	// min_priority=high filters all updates → only-on-updates should skip.
	send, _, total := shouldNotify(
		config.NotifierConfig{MinPriority: "high"},
		"only-on-updates", "", results, nil, nil,
	)
	if send || total != 0 {
		t.Errorf("expected send=false, total=0; got send=%v, total=%d", send, total)
	}

	// Same but with "always" policy → should still send.
	send, _, _ = shouldNotify(
		config.NotifierConfig{MinPriority: "high", SendPolicy: "always"},
		"only-on-updates", "", results, nil, nil,
	)
	if !send {
		t.Error("expected send=true for always policy even with all updates filtered")
	}
}

func TestShouldNotify_ErrorsTriggerNotification(t *testing.T) {
	// No updates but errors → only-on-updates should still send.
	send, _, _ := shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "",
		nil,
		[]error{fmt.Errorf("checker failed")},
		nil,
	)
	if !send {
		t.Error("expected send=true when errors exist even with only-on-updates")
	}
}

func TestShouldNotify_NotifyOverride(t *testing.T) {
	// --notify=true forces send regardless of policy.
	send, _, _ := shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "", nil, nil, boolPtr(true),
	)
	if !send {
		t.Error("expected send=true with --notify=true override")
	}
}

func TestShouldNotify_BackwardCompatibility(t *testing.T) {
	// Config with no per-notifier fields behaves like before.
	results := []*checker.CheckResult{
		{CheckerName: "apt", Updates: []checker.Update{{Name: "curl", Priority: "normal"}}},
	}

	// With updates → should send.
	send, filtered, total := shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "", results, nil, nil,
	)
	if !send || total != 1 || len(filtered) != 1 {
		t.Errorf("backward compat: expected send=true, total=1; got send=%v, total=%d", send, total)
	}

	// Without updates → should not send.
	send, _, _ = shouldNotify(
		config.NotifierConfig{},
		"only-on-updates", "", nil, nil, nil,
	)
	if send {
		t.Error("backward compat: expected send=false with no updates and only-on-updates")
	}
}
