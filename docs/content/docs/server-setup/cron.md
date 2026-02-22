---
title: "Cron Scheduling - Automate Update Checks"
description: "Schedule automated update checks with cron. Built-in cron management, manual crontab setup, logging, and verification."
weight: 3
---

Update-Watcher is designed to run on a schedule, checking for available updates and sending notifications automatically. This page covers all scheduling options, from the built-in `install-cron` command to manual crontab configuration and output logging.

## Setting Up a Schedule

{{< tabs items="install-cron Command,Manual Crontab" >}}

{{< tab >}}

### Using install-cron

The simplest way to set up a daily schedule is the built-in `install-cron` command:

```bash {filename="Terminal"}
update-watcher install-cron --time 07:00
```

This creates a cron entry that runs `update-watcher run --quiet` every day at 07:00. The `--quiet` flag suppresses terminal output, which is appropriate for unattended cron execution.

To use the default time of 07:00, the `--time` flag can be omitted:

```bash {filename="Terminal"}
update-watcher install-cron
```

For a custom cron expression (e.g., twice daily):

```bash {filename="Terminal"}
update-watcher install-cron --cron-expr "0 7,19 * * *"
```

See [install-cron](../../cli/install-cron/) for full command documentation.

{{< /tab >}}

{{< tab >}}

### Manual Crontab Setup

For more control, edit the crontab directly.

#### Current User

```bash {filename="Terminal"}
crontab -e
```

#### Dedicated Service User

On Linux servers with a dedicated `update-watcher` user:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -e
```

#### Dedicated Service User Entry

When running as a dedicated system user on Linux, include the `--as-service-user` flag:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet --as-service-user
```

{{< /tab >}}

{{< /tabs >}}

## Schedule Examples

{{< callout emoji="💡" >}}
All examples use the full path to the binary (`/usr/local/bin/update-watcher`). This is important because cron runs with a minimal PATH.
{{< /callout >}}

**Daily at 07:00** (the most common setup):

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet
```

**Twice daily** at 07:00 and 19:00:

```text {filename="Crontab"}
0 7,19 * * * /usr/local/bin/update-watcher run --quiet
```

**Every Monday at 06:00** (weekly):

```text {filename="Crontab"}
0 6 * * 1 /usr/local/bin/update-watcher run --quiet
```

**Every 6 hours**:

```text {filename="Crontab"}
0 */6 * * * /usr/local/bin/update-watcher run --quiet
```

**Weekdays only at 08:00**:

```text {filename="Crontab"}
0 8 * * 1-5 /usr/local/bin/update-watcher run --quiet
```

## Logging Cron Output

Cron jobs run silently by default. To capture output for debugging and auditing, you have several options.

### Redirect to a Log File

Append stdout and stderr to a log file directly in the crontab entry:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet 2>&1 >> /var/log/update-watcher-cron.log
```

### Use the Built-in Log File

Configure a log file in your `config.yaml`:

```yaml {filename="config.yaml"}
settings:
  log_file: "/var/log/update-watcher.log"
```

With this setting, Update-Watcher writes structured log output to the specified file on every run, regardless of cron redirection.

### Cron Mail

By default, cron emails any output to the crontab owner. If your server has a local mail transport agent (MTA) configured, you can rely on this for notifications about cron failures. However, the `--quiet` flag suppresses most output. Remove `--quiet` if you want cron mail to include the full report:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run 2>&1
```

## Verifying Cron Is Running

### Check the Crontab

View the active cron entries:

```bash {filename="Terminal"}
crontab -l
```

For the dedicated service user:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -l
```

### Check Cron Logs

On systems with `syslog`, cron execution is logged:

```bash {filename="Terminal"}
grep update-watcher /var/log/syslog
```

On systemd-based systems:

```bash {filename="Terminal"}
journalctl -u cron --grep update-watcher
```

### Check the Application Log

If a log file is configured, check the last entry:

```bash {filename="Terminal"}
tail -20 /var/log/update-watcher.log
```

### Verify Manually

Run the exact command that cron executes to verify it works:

```bash {filename="Terminal"}
# As your user
/usr/local/bin/update-watcher run --quiet

# As the dedicated service user
sudo -u update-watcher /usr/local/bin/update-watcher run --quiet --as-service-user
```

## Automated Self-Updates

In addition to scheduling update checks, you can schedule automatic self-updates of the Update-Watcher binary. This ensures your monitoring tool stays current with the latest features and fixes.

```bash {filename="Terminal"}
update-watcher install-cron --type self-update --cron-expr "0 3 * * 0"
```

This runs `update-watcher self-update` every Sunday at 3:00 AM. Both job types are managed independently -- you can install, update, or remove them separately.

A typical production setup with both jobs:

```text {filename="Crontab"}
# update-watcher scheduled check
0 7 * * * /usr/local/bin/update-watcher run --quiet 2>&1 | logger -t update-watcher
# update-watcher self-update
0 3 * * 0 /usr/local/bin/update-watcher self-update 2>&1 | logger -t update-watcher
```

See [install-cron](../../cli/install-cron/) for all `--type` options and flags.

## Removing Cron Jobs

### Using uninstall-cron

Remove the check job:

```bash {filename="Terminal"}
update-watcher uninstall-cron
```

Remove the self-update job:

```bash {filename="Terminal"}
update-watcher uninstall-cron --type self-update
```

Remove all Update-Watcher cron jobs:

```bash {filename="Terminal"}
update-watcher uninstall-cron --all
```

For the dedicated service user:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher uninstall-cron --all
```

### Manually

Edit the crontab and remove the Update-Watcher lines:

```bash {filename="Terminal"}
crontab -e
```

Or for the service user:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -e
```

## Troubleshooting

{{< details title="Troubleshooting: Cron Job Not Running" >}}

Common causes:

{{< callout type="warning" >}}
**PATH issues** -- Cron runs with a minimal PATH. Always use the full path to the binary (`/usr/local/bin/update-watcher`) in crontab entries.
{{< /callout >}}

- **Permission denied** -- The user's crontab may not have permission to execute the binary or read the config file.
- **Cron daemon not running** -- Check with `systemctl status cron` or `systemctl status crond`.

{{< /details >}}

{{< details title="Troubleshooting: Notifications Not Arriving" >}}

Run the command manually with `--verbose` to diagnose:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher run --verbose
```

Check that the `send_policy` is not set to `only-on-updates` when there are no updates to report. Use `--notify=true` to force a notification for testing:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher run --notify=true
```

{{< /details >}}

## Related

- [install-cron](../../cli/install-cron/) -- Built-in cron job management.
- [uninstall-cron](../../cli/uninstall-cron/) -- Remove the cron job.
- [Linux Server Setup](../linux/) -- Full server setup guide.
- [macOS Setup](../macos/) -- macOS scheduling with cron and launchd.
