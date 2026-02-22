---
title: "run - Execute Update Checks"
description: "Run all configured update checks and send notifications. Supports JSON output, checker filtering, and notification control."
weight: 1
---

The `run` command is the primary entry point for Update-Watcher. It executes all configured update checkers, collects the results, and sends notifications through the configured channels. This is the command that cron jobs invoke on a schedule.

## Usage

```bash {filename="Terminal"}
update-watcher run [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | `text\|json` | `text` | Output format. Use `json` for machine-readable output suitable for scripting and pipelines. |
| `--only TYPE` | string | (all) | Run only the specified checker type (e.g., `apt`, `docker`, `wordpress`). All other checkers are skipped. |
| `--notify BOOL` | bool | (config) | Control notification delivery. `true` forces notifications regardless of send policy. `false` suppresses all notifications. Omit to follow the configured `send_policy`. |
| `--quiet` | | `false` | Suppress all output except errors. |
| `--verbose` | | `false` | Enable verbose debug output. |

{{< callout type="info" >}}
The `--notify` flag overrides the `send_policy` setting in your configuration. When omitted, the configured policy applies (see [Send Policy](../advanced/send-policy/)).
{{< /callout >}}

## Exit Codes

The `run` command uses exit codes to communicate results for scripting:

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | No updates | All checkers ran successfully and no updates were found. |
| 1 | Updates found | At least one checker reported available updates. |
| 2 | Partial failure | Some checkers succeeded, but at least one failed (e.g., network timeout, missing permissions). |
| 3 | Complete failure | All configured checkers failed. |
| 4 | Config error | The configuration file is missing, malformed, or contains invalid values. |

For detailed scripting examples using exit codes, see [Exit Codes](../advanced/exit-codes/).

## Examples

### Basic Run

Execute all configured checkers and send notifications according to the configured `send_policy`:

```bash {filename="Terminal"}
update-watcher run
```

### JSON Output

Output results as JSON for parsing by other tools:

```bash {filename="Terminal"}
update-watcher run --format json
```

Combine with `jq` to extract specific information:

```bash {filename="Terminal"}
update-watcher run --format json | jq '.checkers[] | select(.updates > 0)'
```

For more JSON output examples, see [JSON Output](../advanced/json-output/).

### Run a Single Checker

Run only the APT checker, skipping all other configured checkers:

```bash {filename="Terminal"}
update-watcher run --only apt
```

This is useful for testing a specific checker or debugging issues with a single package manager.

### Suppress Notifications

Run all checks but do not send any notifications. Results are printed to the terminal only:

```bash {filename="Terminal"}
update-watcher run --notify=false
```

This is equivalent to a "dry run" and is useful for verifying that checkers are working before enabling notifications.

### Force Notifications

Force notifications to be sent even if no updates are found, regardless of the configured `send_policy`:

```bash {filename="Terminal"}
update-watcher run --notify=true
```

This is useful for testing notification delivery or as a heartbeat confirmation.

### Quiet Mode for Cron

Run silently with no terminal output. Only errors are printed to stderr:

```bash {filename="Terminal"}
update-watcher run --quiet
```

This is the mode used by the cron job created with `update-watcher install-cron`.

### Verbose Debug Output

Enable detailed logging to diagnose issues:

```bash {filename="Terminal"}
update-watcher run --verbose
```

The verbose output shows each step: config loading, checker initialization, package manager queries, result parsing, and notification delivery.

### Combining Flags

Run only the Docker checker with JSON output and no notifications:

```bash {filename="Terminal"}
update-watcher run --only docker --format json --notify=false
```

## Typical Cron Entry

When scheduled via `update-watcher install-cron`, the cron entry looks like:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet
```

For dedicated service user setups:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet --as-service-user
```

{{< callout emoji="💡" >}}
Schedule checks during off-peak hours to avoid competing with package manager lock files and to receive notifications before you start your workday. See [Cron Scheduling](../../server-setup/cron/) for more scheduling options.
{{< /callout >}}

## Related

- [JSON Output](../advanced/json-output/) -- Detailed guide to parsing JSON results.
- [Exit Codes](../advanced/exit-codes/) -- Scripting reference for all exit codes.
- [Send Policy](../advanced/send-policy/) -- Control when notifications are sent.
- [Cron Scheduling](../../server-setup/cron/) -- Automate update checks on a schedule.
