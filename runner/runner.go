package runner

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/internal/selfupdate"
	"github.com/mahype/update-watcher/internal/version"
	// Import checker implementations so they register themselves
	_ "github.com/mahype/update-watcher/checker/apk"
	_ "github.com/mahype/update-watcher/checker/apt"
	_ "github.com/mahype/update-watcher/checker/distro"
	_ "github.com/mahype/update-watcher/checker/dnf"
	_ "github.com/mahype/update-watcher/checker/docker"
	_ "github.com/mahype/update-watcher/checker/flatpak"
	_ "github.com/mahype/update-watcher/checker/homebrew"
	_ "github.com/mahype/update-watcher/checker/macos"
	_ "github.com/mahype/update-watcher/checker/openclaw"
	_ "github.com/mahype/update-watcher/checker/pacman"
	_ "github.com/mahype/update-watcher/checker/snap"
	_ "github.com/mahype/update-watcher/checker/webproject"
	_ "github.com/mahype/update-watcher/checker/wordpress"
	_ "github.com/mahype/update-watcher/checker/zypper"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/notifier"
	// Import notifier implementations so they register themselves
	_ "github.com/mahype/update-watcher/notifier/discord"
	_ "github.com/mahype/update-watcher/notifier/email"
	_ "github.com/mahype/update-watcher/notifier/googlechat"
	_ "github.com/mahype/update-watcher/notifier/gotify"
	_ "github.com/mahype/update-watcher/notifier/homeassistant"
	_ "github.com/mahype/update-watcher/notifier/matrix"
	_ "github.com/mahype/update-watcher/notifier/mattermost"
	_ "github.com/mahype/update-watcher/notifier/ntfy"
	_ "github.com/mahype/update-watcher/notifier/pagerduty"
	_ "github.com/mahype/update-watcher/notifier/pushbullet"
	_ "github.com/mahype/update-watcher/notifier/pushover"
	_ "github.com/mahype/update-watcher/notifier/rocketchat"
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

// Run executes all configured checkers in parallel and sends notifications.
func (r *Runner) Run() (*RunResult, error) {
	result := &RunResult{}
	ctx := context.Background()
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, wCfg := range r.cfg.Watchers {
		if !wCfg.Enabled {
			continue
		}
		if r.only != "" && wCfg.Type != r.only {
			continue
		}

		wCfg := wCfg // capture for goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()

			slog.Info("running checker", "type", wCfg.Type)

			c, err := checker.Create(wCfg.Type, wCfg)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, &checker.CheckError{
					CheckerName: wCfg.Type,
					Err:         err,
					Retryable:   false,
				})
				mu.Unlock()
				slog.Error("failed to create checker", "type", wCfg.Type, "error", err)
				return
			}

			cr, err := c.Check(ctx)

			mu.Lock()
			defer mu.Unlock()

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
				return
			}

			result.Results = append(result.Results, cr)
			slog.Info("checker completed", "type", c.Name(), "updates", len(cr.Updates))
		}()
	}

	wg.Wait()

	// Self-update check: always runs unless --only filters to a specific checker.
	if r.only == "" {
		if cr := r.checkSelfUpdate(); cr != nil {
			result.Results = append(result.Results, cr)
		}
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
		if err := r.notify(ctx, result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	} else {
		slog.Info("dry run mode, skipping notifications")
	}

	return result, nil
}

func (r *Runner) notify(ctx context.Context, result *RunResult) error {
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
		if err := n.Send(ctx, r.cfg.Hostname, result.Results); err != nil {
			notifyErrors = append(notifyErrors, fmt.Errorf("notifier %q: %w", n.Name(), err))
			slog.Error("notification failed", "type", n.Name(), "error", err)
		}
	}

	if len(notifyErrors) > 0 {
		return fmt.Errorf("notification errors: %v", notifyErrors)
	}
	return nil
}

// checkSelfUpdate queries GitHub for a newer version of update-watcher.
// Returns a CheckResult only if a newer version is available, nil otherwise.
func (r *Runner) checkSelfUpdate() *checker.CheckResult {
	slog.Info("checking for update-watcher self-update")

	release, err := selfupdate.LatestRelease()
	if err != nil {
		slog.Warn("self-update check failed", "error", err)
		return nil
	}

	if !selfupdate.NeedsUpdate(version.Version, release) {
		slog.Info("update-watcher is up to date", "version", version.Version)
		return nil
	}

	slog.Info("update-watcher update available", "current", version.Version, "latest", release.TagName)

	return &checker.CheckResult{
		CheckerName: "self-update",
		Updates: []checker.Update{
			{
				Name:           "update-watcher",
				CurrentVersion: strings.TrimPrefix(version.Version, "v"),
				NewVersion:     release.Version,
				Type:           checker.UpdateTypeCore,
				Priority:       checker.PriorityNormal,
			},
		},
		Summary:   fmt.Sprintf("update available: %s → %s", version.Version, release.TagName),
		CheckedAt: time.Now(),
	}
}
