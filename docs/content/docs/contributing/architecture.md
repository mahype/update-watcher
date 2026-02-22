---
title: "Architecture Overview"
description: "Understand Update-Watcher's codebase, project structure, design patterns, and data flow. Essential reading for contributors."
weight: 1
---

# Architecture Overview

Update-Watcher follows a modular plugin architecture. Checkers and notifiers register themselves at startup, and a central runner orchestrates parallel execution.

## Project Structure

{{< filetree/container >}}
  {{< filetree/folder name="update-watcher" >}}
    {{< filetree/file name="main.go" >}}
    {{< filetree/folder name="cmd" >}}
      {{< filetree/file name="root.go" >}}
      {{< filetree/file name="run.go" >}}
      {{< filetree/file name="setup.go" >}}
      {{< filetree/file name="watch_*.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="checker" >}}
      {{< filetree/file name="checker.go" >}}
      {{< filetree/file name="registry.go" >}}
      {{< filetree/folder name="apt" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="docker" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="wordpress" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="notifier" >}}
      {{< filetree/file name="notifier.go" >}}
      {{< filetree/file name="registry.go" >}}
      {{< filetree/folder name="formatting" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="slack" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="discord" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="runner" >}}
      {{< filetree/file name="runner.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="config" >}}
      {{< filetree/file name="config.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="wizard" >}}
      {{< filetree/file name="wizard.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="output" >}}
      {{< filetree/file name="terminal.go" >}}
      {{< filetree/file name="table.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="cron" >}}
      {{< filetree/file name="cron.go" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="internal" >}}
      {{< filetree/folder name="executil" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="hostname" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="fsutil" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="httputil" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="rootcheck" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="selfupdate" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="version" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="scripts" >}}
      {{< filetree/file name="install.sh" >}}
      {{< filetree/file name="uninstall.sh" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

## Key Design Patterns

### Registry Pattern

{{< callout type="info" >}}
Checkers and notifiers use a registry pattern with factory functions. Each implementation registers itself during package initialization via `init()`.
{{< /callout >}}

**Checker registry** (`checker/registry.go`):

```go {filename="checker/registry.go"}
// FactoryFunc creates a Checker from a watcher configuration.
type FactoryFunc func(cfg config.WatcherConfig) (Checker, error)

var registry = map[string]FactoryFunc{}

// Register adds a checker factory to the global registry.
func Register(name string, factory FactoryFunc) {
    registry[name] = factory
}
```

**Notifier registry** (`notifier/registry.go`):

```go {filename="notifier/registry.go"}
// Register adds a notifier factory to the global registry.
func Register(name string, factory FactoryFunc)

// RegisterMeta adds display metadata for the setup wizard.
func RegisterMeta(meta NotifierMeta)
```

### Blank Imports for Registration

{{< callout type="info" >}}
The runner imports all checker and notifier packages as blank imports to trigger their `init()` functions.
{{< /callout >}}

```go {filename="runner/runner.go"}
// runner/runner.go
import (
    _ "github.com/mahype/update-watcher/checker/apt"
    _ "github.com/mahype/update-watcher/checker/docker"
    // ... all other checker packages

    _ "github.com/mahype/update-watcher/notifier/slack"
    _ "github.com/mahype/update-watcher/notifier/discord"
    // ... all other notifier packages
)
```

### Interface-Based Design

{{< callout type="info" >}}
Both checkers and notifiers are defined by simple interfaces, allowing new implementations to be added without modifying existing code.
{{< /callout >}}

```go {filename="checker/checker.go"}
// checker/checker.go
type Checker interface {
    Name() string
    Check(ctx context.Context) (*CheckResult, error)
}
```

```go {filename="notifier/notifier.go"}
// notifier/notifier.go
type Notifier interface {
    Name() string
    Send(ctx context.Context, hostname string, results []*checker.CheckResult) error
}
```

### Functional Options

The runner uses the functional options pattern for configuration:

```go {filename="runner/runner.go"}
runner.New(cfg, runner.WithNotify(&notify), runner.WithOnly("apt"))
```

## Data Flow

The execution flow during `update-watcher run`:

```
1. cmd/run.go
   ├── config.Load()           # Read YAML, resolve ${ENV_VAR} references
   └── runner.New(cfg).Run()
       ├── For each enabled watcher (parallel):
       │   ├── checker.Create(type, cfg)   # Registry lookup + factory
       │   └── checker.Check(ctx)          # Execute check
       ├── Self-update check               # Query GitHub Releases API
       ├── Aggregate results               # Count updates, detect security
       └── Notify (sequential):
           ├── Apply send_policy           # Skip if no updates + only-on-updates
           └── For each enabled notifier:
               ├── notifier.Create(type, cfg)
               └── notifier.Send(ctx, hostname, results)
```

{{< callout type="info" >}}
**Key points:**
- Checkers run in **parallel** using `sync.WaitGroup` with a shared mutex for results
- Notifiers run **sequentially** after all checkers complete
- The self-update check always runs (unless `--only` filters to a specific checker)
- Errors from individual checkers do not abort other checkers
{{< /callout >}}

## Core Types

### CheckResult

```go {filename="checker/checker.go"}
type CheckResult struct {
    CheckerName string    `json:"checker_name"`
    Updates     []Update  `json:"updates"`
    Summary     string    `json:"summary"`
    CheckedAt   time.Time `json:"checked_at"`
    Error       string    `json:"error,omitempty"`
    Notes       []string  `json:"notes,omitempty"`
}
```

### Update

```go {filename="checker/checker.go"}
type Update struct {
    Name           string `json:"name"`
    CurrentVersion string `json:"current_version"`
    NewVersion     string `json:"new_version"`
    Type           string `json:"type"`      // security, regular, plugin, theme, core, image, distro
    Priority       string `json:"priority"`  // critical, high, normal, low
    Source         string `json:"source,omitempty"`
    Phasing        string `json:"phasing,omitempty"`
}
```

### RunResult

```go {filename="runner/runner.go"}
type RunResult struct {
    Results      []*checker.CheckResult
    TotalUpdates int
    HasSecurity  bool
    Errors       []error
}
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [cobra](https://github.com/spf13/cobra) | CLI framework |
| [viper](https://github.com/spf13/viper) | Configuration management |
| [charmbracelet/huh](https://github.com/charmbracelet/huh) | Interactive TUI forms |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [goreleaser](https://goreleaser.com/) | Release automation |
