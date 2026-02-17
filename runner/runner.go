package runner

import (
	"fmt"
	"log/slog"

	"github.com/mahype/update-watcher/checker"
	// Import checker implementations so they register themselves
	_ "github.com/mahype/update-watcher/checker/apk"
	_ "github.com/mahype/update-watcher/checker/apt"
	_ "github.com/mahype/update-watcher/checker/dnf"
	_ "github.com/mahype/update-watcher/checker/docker"
	_ "github.com/mahype/update-watcher/checker/macos"
	_ "github.com/mahype/update-watcher/checker/pacman"
	_ "github.com/mahype/update-watcher/checker/webproject"
	_ "github.com/mahype/update-watcher/checker/wordpress"
	_ "github.com/mahype/update-watcher/checker/zypper"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/notifier"
	// Import notifier implementations so they register themselves
	_ "github.com/mahype/update-watcher/notifier/discord"
	_ "github.com/mahype/update-watcher/notifier/email"
	_ "github.com/mahype/update-watcher/notifier/ntfy"
	_ "github.com/mahype/update-watcher/notifier/slack"
	_ "github.com/mahype/update-watcher/notifier/teams"
	_ "github.com/mahype/update-watcher/notifier/telegram"
	_ "github.com/mahype/update-watcher/notifier/webhook"
)

// RunResult is the aggregate outcome of a full run.
type RunResult struct {
	Results      []*checker.CheckResult
	TotalUpdates int
	HasSecurity  bool
	Errors       []error
}

// Runner orchestrates checker execution and notification dispatch.
type Runner struct {
	cfg    *config.Config
	dryRun bool
	only   string
}

// Option configures the runner.
type Option func(*Runner)

// WithDryRun disables notifications.
func WithDryRun(dryRun bool) Option {
	return func(r *Runner) { r.dryRun = dryRun }
}

// WithOnly restricts the run to a single checker type.
func WithOnly(only string) Option {
	return func(r *Runner) { r.only = only }
}

// New creates a new Runner.
func New(cfg *config.Config, opts ...Option) *Runner {
	r := &Runner{cfg: cfg}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes all configured checkers and sends notifications.
func (r *Runner) Run() (*RunResult, error) {
	result := &RunResult{}

	for _, wCfg := range r.cfg.Watchers {
		if !wCfg.Enabled {
			continue
		}
		if r.only != "" && wCfg.Type != r.only {
			continue
		}

		slog.Info("running checker", "type", wCfg.Type)

		c, err := checker.Create(wCfg.Type, wCfg)
		if err != nil {
			result.Errors = append(result.Errors, &checker.CheckError{
				CheckerName: wCfg.Type,
				Err:         err,
				Retryable:   false,
			})
			slog.Error("failed to create checker", "type", wCfg.Type, "error", err)
			continue
		}

		cr, err := c.Check()
		if err != nil {
			result.Errors = append(result.Errors, &checker.CheckError{
				CheckerName: c.Name(),
				Err:         err,
				Retryable:   true,
			})
			slog.Error("checker failed", "type", c.Name(), "error", err)

			if cr != nil {
				cr.Error = err.Error()
				result.Results = append(result.Results, cr)
			}
			continue
		}

		result.Results = append(result.Results, cr)
		slog.Info("checker completed", "type", c.Name(), "updates", len(cr.Updates))
	}

	// Aggregate
	for _, cr := range result.Results {
		result.TotalUpdates += len(cr.Updates)
		if cr.HasSecurityUpdates() {
			result.HasSecurity = true
		}
	}

	// Notify
	if !r.dryRun {
		if err := r.notify(result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	} else {
		slog.Info("dry run mode, skipping notifications")
	}

	return result, nil
}

func (r *Runner) notify(result *RunResult) error {
	policy := r.cfg.Settings.SendPolicy
	if policy == "only-on-updates" && result.TotalUpdates == 0 && len(result.Errors) == 0 {
		slog.Info("no updates found, skipping notification (send_policy: only-on-updates)")
		return nil
	}

	var notifyErrors []error
	for _, nCfg := range r.cfg.Notifiers {
		if !nCfg.Enabled {
			continue
		}

		n, err := notifier.Create(nCfg.Type, nCfg)
		if err != nil {
			notifyErrors = append(notifyErrors, fmt.Errorf("notifier %q: %w", nCfg.Type, err))
			continue
		}

		slog.Info("sending notification", "type", n.Name())
		if err := n.Send(r.cfg.Hostname, result.Results); err != nil {
			notifyErrors = append(notifyErrors, fmt.Errorf("notifier %q: %w", n.Name(), err))
			slog.Error("notification failed", "type", n.Name(), "error", err)
		}
	}

	if len(notifyErrors) > 0 {
		return fmt.Errorf("notification errors: %v", notifyErrors)
	}
	return nil
}
