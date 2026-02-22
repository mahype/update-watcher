---
title: "Adding a New Checker"
description: "Step-by-step guide to implementing a new update checker for Update-Watcher. Includes code templates, registration, CLI integration, and testing."
weight: 2
---

# Adding a New Checker

This guide walks you through adding a new update checker to Update-Watcher. A checker is a module that detects available updates from a specific source (package manager, application, service, etc.).

## Overview

Adding a new checker involves these steps:

1. Create a new package under `checker/`
2. Implement the `Checker` interface
3. Register the checker in `init()`
4. Add a blank import in `runner/runner.go`
5. Add a CLI subcommand in `cmd/`
6. (Optional) Add to the setup wizard

{{% steps %}}

## Step 1: Create the Package

Create a new directory under `checker/` for your checker:

{{< filetree/container >}}
  {{< filetree/folder name="checker" >}}
    {{< filetree/folder name="mychecker" >}}
      {{< filetree/file name="mychecker.go" >}}
      {{< filetree/file name="mychecker_test.go" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

## Step 2: Implement the Checker Interface

Your checker must implement the `Checker` interface:

```go {filename="checker/checker.go"}
type Checker interface {
    Name() string
    Check(ctx context.Context) (*CheckResult, error)
}
```

Here is a minimal checker implementation:

```go {filename="checker/mychecker/mychecker.go"}
package mychecker

import (
    "context"
    "time"

    "github.com/mahype/update-watcher/checker"
    "github.com/mahype/update-watcher/config"
    "github.com/mahype/update-watcher/internal/executil"
)

func init() {
    checker.Register("mychecker", NewFromConfig)
}

// MyChecker checks for available updates from MySource.
type MyChecker struct {
    useSudo bool
}

// NewFromConfig creates a MyChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
    return &MyChecker{
        useSudo: cfg.Options.GetBool("use_sudo", false),
    }, nil
}

func (c *MyChecker) Name() string { return "mychecker" }

func (c *MyChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
    result := &checker.CheckResult{
        CheckerName: c.Name(),
        CheckedAt:   time.Now(),
    }

    // Run external commands to detect updates
    out, err := executil.Run("my-tool", "check-updates")
    if err != nil {
        return result, fmt.Errorf("my-tool failed: %w", err)
    }

    // Parse the output into Update structs
    result.Updates = parseUpdates(out.Stdout)
    result.Summary = checker.BuildSummary(result.Updates, "packages")

    return result, nil
}
```

{{< callout type="info" >}}
**Key points:**
- **`init()`** registers the checker with the global registry. The first argument is the type name used in YAML config.
- **`NewFromConfig`** is the factory function. Read options from `cfg.Options` using typed accessors (`GetBool`, `GetString`, `GetInt`, `GetStringSlice`, `GetMapSlice`).
- **`Name()`** returns the human-readable name shown in output and notifications.
- **`Check()`** performs the actual check. Return a `CheckResult` even on error if you have partial results.
- Use **`checker.BuildSummary()`** to generate consistent summary strings.
- Use **`executil.Run()`** for running external commands with timeout support.
{{< /callout >}}

### Update types

Set the `Type` field on each `Update` to classify it:

```go {filename="checker/checker.go"}
checker.UpdateTypeSecurity  // "security"
checker.UpdateTypeRegular   // "regular"
checker.UpdateTypePlugin    // "plugin"
checker.UpdateTypeTheme     // "theme"
checker.UpdateTypeCore      // "core"
checker.UpdateTypeImage     // "image"
checker.UpdateTypeDistro    // "distro"
```

## Step 3: Register in runner.go

Add a blank import for your new checker package in `runner/runner.go`:

```go {filename="runner/runner.go"}
import (
    // ... existing imports
    _ "github.com/mahype/update-watcher/checker/mychecker"
)
```

{{< callout emoji="💡" >}}
This triggers the `init()` function and registers the checker.
{{< /callout >}}

## Step 4: Add a CLI Subcommand

Create `cmd/watch_mychecker.go`:

```go {filename="cmd/watch_mychecker.go"}
package cmd

func init() {
    watchCmd := addWatchCommand("mychecker", "Add MyChecker watcher", nil)
    // Add checker-specific flags:
    watchCmd.Flags().Bool("my-flag", false, "Description of my flag")
}
```

The `addWatchCommand` helper creates a `watch mychecker` subcommand that adds the watcher to the configuration.

## Step 5: Test Your Checker

Create `checker/mychecker/mychecker_test.go`:

```go {filename="checker/mychecker/mychecker_test.go"}
package mychecker

import (
    "testing"
)

func TestParseUpdates(t *testing.T) {
    input := `package1 1.0.0 -> 2.0.0
package2 3.1.0 -> 3.2.0`

    updates := parseUpdates(input)

    if len(updates) != 2 {
        t.Errorf("expected 2 updates, got %d", len(updates))
    }
}
```

Run tests with:

```bash {filename="Terminal"}
make test
```

## Step 6: Add to the Setup Wizard (Optional)

If your checker can be auto-detected (e.g., by checking if a command exists in PATH), you can add it to the wizard in `wizard/wizard.go`. Look at how existing checkers are detected and follow the same pattern.

{{% /steps %}}

## Checklist

Before submitting a PR for a new checker:

- [ ] Implements `Checker` interface (`Name()`, `Check()`)
- [ ] Registered via `init()` with `checker.Register()`
- [ ] Blank import added to `runner/runner.go`
- [ ] CLI subcommand created in `cmd/watch_*.go`
- [ ] Tests written and passing (`make test`)
- [ ] Code formatted (`make fmt`) and linted (`make lint`)
- [ ] Documentation updated (checker options in README)
