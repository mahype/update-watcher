---
title: "Send Policy - Control When Notifications Are Sent"
description: "Configure when Update-Watcher sends notifications. Choose between always or only-on-updates, with CLI overrides for testing."
weight: 4
---

The `send_policy` setting controls when Update-Watcher sends notifications after a check run. You can choose to receive notifications only when updates are found, or on every run regardless of results. CLI flags provide per-run overrides for testing and special cases.

## Send Policy Options

### only-on-updates (Default)

```yaml {filename="config.yaml"}
settings:
  send_policy: "only-on-updates"
```

{{< callout type="info" >}}
`only-on-updates` is the default behavior. If you do not set `send_policy` in your configuration, this is what you get. It keeps notification channels quiet when there is nothing to report, reducing alert fatigue.
{{< /callout >}}

Notifications are sent only when at least one checker reports available updates. If all checkers run successfully and no updates are found, no notification is sent.

**When notifications are sent:**
- At least one package has an available update.
- A new Docker image version is available.
- A WordPress core, plugin, or theme update is available.

**When notifications are not sent:**
- All checkers report no available updates.
- The run completes with no changes since the last notification.

### always

```yaml {filename="config.yaml"}
settings:
  send_policy: "always"
```

Notifications are sent after every run, even when no updates are found. The notification includes the results of all checkers, with a "no updates" message for checkers that found nothing.

**Use cases for "always":**

- **Heartbeat monitoring** -- Confirm that Update-Watcher is running and the cron job is working. If you stop receiving daily reports, something is wrong.
- **Audit trail** -- Maintain a record of every check, including runs with no updates, for compliance or operational logging.
- **Peace of mind** -- Some administrators prefer explicit confirmation that the system was checked, rather than interpreting silence as "no updates."

## CLI Overrides

The `--notify` flag on the `run` command overrides the configured `send_policy` for a single invocation. This is useful for testing, debugging, and one-off operations.

### Force Notifications

Send notifications regardless of the send policy and regardless of whether updates were found:

```bash {filename="Terminal"}
update-watcher run --notify=true
```

This is useful for:
- Testing notification delivery after configuring a new notifier.
- Verifying that webhook URLs, bot tokens, and SMTP credentials are correct.
- Sending a manual status report on demand.

### Suppress Notifications

Run all checks but do not send any notifications:

```bash {filename="Terminal"}
update-watcher run --notify=false
```

This is useful for:
- Verifying that checkers are working without triggering notifications.
- Running a quick check during maintenance without alerting the team.
- Testing configuration changes before committing to them.

### No Override

When `--notify` is omitted, the configured `send_policy` applies:

```bash {filename="Terminal"}
update-watcher run
```

## Per-Notifier Send Policy

Each notifier can override the global `send_policy`. If a notifier does not specify its own, the global setting applies.

```yaml {filename="config.yaml"}
settings:
  send_policy: "only-on-updates"    # global default

notifiers:
  - type: slack
    send_policy: "always"           # Slack always receives notifications
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
  - type: email
    # no send_policy → uses global "only-on-updates"
    options:
      smtp_host: "smtp.example.com"
      to: ["admin@example.com"]
```

## Priority Filtering

The `min_priority` setting filters which updates are included in notifications. Updates below the minimum priority are excluded from that notifier's message.

Priority levels from highest to lowest: `critical` > `high` > `normal` > `low`.

```yaml {filename="config.yaml"}
settings:
  min_priority: "normal"            # global: skip low-priority updates

notifiers:
  - type: slack
    min_priority: "low"             # Slack sees everything
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
  - type: email
    min_priority: "high"            # email only for high + critical
    options:
      smtp_host: "smtp.example.com"
      to: ["admin@example.com"]
  - type: pagerduty
    min_priority: "critical"        # PagerDuty only for critical
    options:
      routing_key: "${PAGERDUTY_ROUTING_KEY}"
```

{{< callout type="info" >}}
If `min_priority` is not set (empty), no filtering is applied and all updates are included. Updates without an explicit priority are treated as `normal`.
{{< /callout >}}

### Priority Filtering and Send Policy Interaction

After priority filtering, if no updates remain for a notifier, the `only-on-updates` policy treats it as "no updates" and skips the notification. With `always`, the notification is still sent (but the update list will be empty for that notifier).

The `--notify=true` CLI flag overrides the send policy (forces send) but does **not** bypass priority filtering. This means you can force-send notifications while still respecting each notifier's content filter.

## Configuration Example

A complete configuration with per-notifier policies:

```yaml {filename="config.yaml"}
hostname: "web-prod-01"

watchers:
  - type: apt
  - type: docker

notifiers:
  - type: slack
    send_policy: "always"
    min_priority: "low"
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
  - type: email
    min_priority: "high"
    options:
      smtp_host: "smtp.example.com"
      smtp_port: 587
      username: "${SMTP_USER}"
      password: "${SMTP_PASS}"
      from: "updates@example.com"
      to: ["ops-team@example.com"]

settings:
  send_policy: "only-on-updates"
  min_priority: "normal"
```

## Decision Guide

{{< callout emoji="💡" >}}
Not sure which policy to pick? Start with `only-on-updates` (the default). Switch to `always` if you need a heartbeat confirmation or an audit trail. You can always use `--notify=true` for one-off tests.
{{< /callout >}}

| Scenario | Recommended Policy |
|----------|-------------------|
| Production server, daily checks | `only-on-updates` |
| Single critical server, need confirmation | `always` |
| Multiple servers, shared Slack channel | `only-on-updates` |
| Compliance requirement for check records | `always` |
| Personal workstation | `only-on-updates` |
| First-time setup, verifying configuration | Use `--notify=true` override |

## Behavior with Checker Failures

The send policy interacts with checker outcomes as follows:

| Scenario | only-on-updates | always |
|----------|-----------------|--------|
| No updates found, all checkers OK | No notification | Notification sent |
| Updates found, all checkers OK | Notification sent | Notification sent |
| Partial failure, updates found | Notification sent | Notification sent |
| Partial failure, no updates from successful checkers | No notification | Notification sent |
| Complete failure (all checkers fail) | No notification | Notification sent |

When using the "always" policy, notifications include error information for failed checkers, giving you visibility into both update status and checker health.

## Environment Variable Override

You can override the send policy via environment variable:

```bash {filename="Terminal"}
UPDATE_WATCHER_SETTINGS_SEND_POLICY=always update-watcher run
```

See [Environment Variables](../../configuration/environment-variables/) for details on the `UPDATE_WATCHER_` prefix.

## Related

- [run](../../cli/run/) -- The `--notify` flag and other run options.
- [Configuration](../../configuration/) -- Full YAML configuration reference.
- [Notifiers](../../notifiers/) -- All 16 notification channels.
- [Exit Codes](../exit-codes/) -- Exit codes for scripting alongside notification policies.
