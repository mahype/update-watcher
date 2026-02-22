---
title: "CLI Reference"
description: "Complete command-line reference for Update-Watcher. All commands, flags, and exit codes."
weight: 50
---

Update-Watcher is a single-binary CLI tool. All functionality is accessed through the `update-watcher` command and its subcommands. This section provides a complete reference for every command, flag, and exit code.

## Global Flags

The following flags are available on all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config PATH` | `-c` | (auto) | Path to a specific configuration file. Overrides the default search paths. |
| `--quiet` | `-q` | `false` | Suppress all output except errors. Useful for cron jobs. |
| `--verbose` | `-v` | `false` | Enable verbose debug output. Shows each step as it executes. |
| `--as-service-user` | | `false` | Run as the dedicated system user (`update-watcher`). Used internally by cron entries. |

{{< callout type="info" >}}
The `--quiet` and `--verbose` flags are mutually exclusive. If both are specified, `--verbose` takes precedence.
{{< /callout >}}

## Commands

| Command | Description |
|---------|-------------|
| [run](run/) | Execute all configured update checks and send notifications. |
| [setup](setup/) | Launch the interactive menu-driven setup wizard. |
| [watch](watch/) | Add a new update checker to the configuration. |
| [unwatch](unwatch/) | Remove a configured update checker from the configuration. |
| [status](status/) | Display the current configuration including watchers, notifiers, and settings. |
| [validate](validate/) | Validate the configuration file for syntax errors and missing fields. |
| [self-update](self-update/) | Update the Update-Watcher binary to the latest release. |
| [install-cron](install-cron/) | Install a cron job for daily automated update checks. |
| [uninstall-cron](uninstall-cron/) | Remove the Update-Watcher cron job from the user's crontab. |
| [version](version/) | Display the version, git commit hash, and build date. |

## Config File Resolution

When no `--config` flag is provided, Update-Watcher searches for a configuration file in the following locations, in order of priority:

| Priority | Path | Use Case |
|----------|------|----------|
| 1 | `/etc/update-watcher/config.yaml` | System-wide config (Linux server setups) |
| 2 | `~/.config/update-watcher/config.yaml` | Per-user config (macOS, desktop Linux) |
| 3 | `./config.yaml` | Current working directory (development, testing) |

The first file found is used. On Linux servers with a dedicated system user, the system-wide path is typical. On macOS, the per-user path under `~/.config` is the default.

To explicitly specify a config file:

```bash {filename="Terminal"}
update-watcher run --config /path/to/config.yaml
```

## Exit Codes

All commands return meaningful exit codes for scripting. The `run` command uses the full range:

| Code | Meaning |
|------|---------|
| 0 | Success -- no updates found. |
| 1 | Updates found -- at least one checker reported available updates. |
| 2 | Partial failure -- some checkers succeeded but at least one failed. |
| 3 | Complete failure -- all checkers failed. |
| 4 | Configuration error -- the config file is missing, malformed, or invalid. |

For scripting examples using exit codes, see [Exit Codes](../advanced/exit-codes/).

## Next Steps

- [Getting Started](../getting-started/) -- Install Update-Watcher and run your first check.
- [Configuration](../configuration/) -- Full YAML configuration reference.
- [Server Setup](../server-setup/) -- Production-ready setup for Linux and macOS servers.
