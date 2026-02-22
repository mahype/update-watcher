# Update-Watcher

Go CLI tool for monitoring software updates across systems. Single static binary, zero runtime dependencies.

- **Module**: `github.com/mahype/update-watcher`
- **Go**: 1.25.7
- **License**: MIT

## Build & Development Commands

```bash
make build          # Static binary (CGO_ENABLED=0) → bin/update-watcher
make test           # go test -v -race -coverprofile=coverage.out ./...
make lint           # golangci-lint run ./...
make fmt            # gofumpt -l -w .
make vet            # go vet ./...
make snapshot       # goreleaser build --snapshot --clean
make install        # Build + install to /usr/local/bin/
```

Test a single package: `go test ./checker/apt/...`

## Architecture

**Entry**: `main.go` → `cmd.Execute()` (Cobra CLI)

**Flow**: CLI command → Runner → parallel Checkers → aggregate results → Notifiers

- **Plugin pattern**: Checkers and Notifiers register via `init()` + factory registry
- **Config**: Viper-based YAML with env var substitution (`${VAR}` / `${VAR:-default}`)
- **Runner** (`runner/runner.go`): Orchestrates parallel checker execution, aggregates results, dispatches to notifiers
- **Key interfaces**:
  - `checker.Checker`: `Name() string` + `Check(ctx) (*CheckResult, error)`
  - `notifier.Notifier`: `Name() string` + `Send(ctx, hostname, results) error`

## Directory Layout

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Cobra commands (run, setup, watch, status, validate, install-cron, selfupdate, etc.) |
| `checker/` | 14 checker implementations (apt, dnf, pacman, zypper, apk, homebrew, macos, docker, snap, flatpak, wordpress, webproject, distro, openclaw) |
| `notifier/` | 16 notifier implementations (slack, discord, teams, telegram, email, ntfy, pushover, gotify, homeassistant, googlechat, matrix, mattermost, rocketchat, pagerduty, pushbullet, webhook) |
| `notifier/formatting/` | Message formatting (plaintext, markdown, helpers) |
| `config/` | Config structs, loading, validation, defaults |
| `runner/` | Orchestration engine |
| `output/` | Terminal formatting (tables, icons, update commands) |
| `wizard/` | Interactive TUI setup (charmbracelet/huh) |
| `internal/executil/` | Command execution (timeouts, sudo, user switching) |
| `internal/fsutil/` | Filesystem utilities |
| `internal/hostname/` | Hostname detection |
| `internal/httputil/` | HTTP client utilities |
| `internal/selfupdate/` | Binary self-update from GitHub releases |
| `internal/version/` | Version info (injected via ldflags) |
| `cron/` | Cron job management |
| `docs/` | Documentation site (GitHub Pages) |

## Key Patterns

### Adding a new Checker

```go
// checker/mychecker/mychecker.go
func init() {
    checker.Register("mychecker", NewFromConfig)
}

func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
    return &MyChecker{
        option1: cfg.Options.GetString("option1", "default"),
        option2: cfg.Options.GetBool("option2", false),
    }, nil
}

func (c *MyChecker) Name() string { return "mychecker" }
func (c *MyChecker) Check(ctx context.Context) (*checker.CheckResult, error) { ... }
```

Then add a `cmd/watch_mychecker.go` for the CLI registration.

### Adding a new Notifier

Same pattern via `notifier.Register()` + `notifier.RegisterMeta()` for UI metadata.

### Config Options Access

```go
cfg.Options.GetString("key", "default")
cfg.Options.GetBool("key", false)
cfg.Options.GetInt("key", 0)
cfg.Options.GetStringSlice("key", []string{})
cfg.Options.GetMapSlice("key")  // for nested configs like wordpress.sites
```

### Command Execution

Always use `internal/executil` — handles timeouts (60s default), sudo, user switching:
```go
executil.Run("apt", "list", "--upgradable")
executil.RunWithTimeout(timeout, "cmd", args...)
executil.RunMaybeSudo(useSudo, "cmd", args...)
executil.RunAsSudo("cmd", args...)
executil.RunAsUser("username", "cmd", args...)
```

### Update Types & Priorities

- **Types**: `security`, `regular`, `plugin`, `theme`, `core`, `image`, `distro`
- **Priorities**: `critical`, `high`, `normal`, `low`

### Config Paths

- **Linux**: `/etc/update-watcher/config.yaml`
- **macOS**: `~/.config/update-watcher/config.yaml`

## Error Handling

- Checkers return `(*CheckResult, error)` — errors are aggregated in the runner, not fatal
- Runner collects all results and errors, notifiers receive both
- Send policy: `only-on-updates` (skip if no updates/errors) or `always`

## Conventions

- **User language**: German (communication), English (code, docs, commits)
- **Commit messages**: English, imperative mood, concise first line
- **Static binary**: Always `CGO_ENABLED=0`
- **Version**: Injected via ldflags at build time (`internal/version`)
- **Releases**: GoReleaser for multi-platform (Linux amd64/arm64/armv7, macOS amd64/arm64)
- **Self-update**: Checks GitHub releases for `mahype/update-watcher`
- **Dependencies**: Cobra (CLI), Viper (config), charmbracelet/huh (TUI), go-yaml/v3
