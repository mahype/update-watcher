---
title: "install-cron - Schedule Automated Tasks"
description: "Install cron jobs for automated update checks and self-updates. Configurable time, cron expression, and job type."
weight: 8
---

The `install-cron` command creates a cron job in the current user's crontab. It supports two job types: scheduled update checks (`check`) and automated binary self-updates (`self-update`).

## Usage

```bash {filename="Terminal"}
update-watcher install-cron [--type TYPE] [--time HH:MM] [--cron-expr "EXPR"]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type TYPE` | string | `check` | Job type: `check` (run update checks) or `self-update` (update the binary). |
| `--time HH:MM` | string | `07:00` | Time of day for daily runs (24-hour format). Creates a daily cron expression. |
| `--cron-expr "EXPR"` | string | (none) | A full cron expression for custom schedules. Overrides `--time` if both are provided. |

## Default Behavior

{{< callout type="info" >}}
Without any flags, `install-cron` creates a daily **check** cron job that runs at **07:00**.
{{< /callout >}}

```bash {filename="Terminal"}
update-watcher install-cron
```

This adds the following entry to the current user's crontab:

```text {filename="Crontab"}
# update-watcher scheduled check
0 7 * * * /usr/local/bin/update-watcher run --quiet 2>&1 | logger -t update-watcher
```

The entry includes an identifying comment that is used by the [uninstall-cron](../uninstall-cron/) command to locate and remove it.

## Examples

### Daily at a Custom Time

Run the check every day at 9:30 AM:

```bash {filename="Terminal"}
update-watcher install-cron --time 09:30
```

### Custom Cron Expression

Run twice daily at 7:00 AM and 7:00 PM:

```bash {filename="Terminal"}
update-watcher install-cron --cron-expr "0 7,19 * * *"
```

Run every Monday at 6:00 AM:

```bash {filename="Terminal"}
update-watcher install-cron --cron-expr "0 6 * * 1"
```

### Automated Self-Update

Schedule a weekly self-update every Sunday at 3:00 AM:

```bash {filename="Terminal"}
update-watcher install-cron --type self-update --cron-expr "0 3 * * 0"
```

This keeps the binary up to date automatically. The self-update job runs independently from the check job.

### Both Jobs Together

Install a daily check and a weekly self-update:

```bash {filename="Terminal"}
update-watcher install-cron --time 07:00
update-watcher install-cron --type self-update --cron-expr "0 3 * * 0"
```

### Verify the Cron Jobs

After installation, verify the entries exist in your crontab:

```bash {filename="Terminal"}
crontab -l
```

You should see:

```text {filename="Crontab"}
# update-watcher scheduled check
0 7 * * * /usr/local/bin/update-watcher run --quiet 2>&1 | logger -t update-watcher
# update-watcher self-update
0 3 * * 0 /usr/local/bin/update-watcher self-update 2>&1 | logger -t update-watcher
```

## How It Works

The `install-cron` command modifies the current user's crontab using the `crontab` utility:

1. Reads the existing crontab.
2. Checks for an existing entry of the same job type (by the identifying comment).
3. If an entry exists, replaces it with the new schedule.
4. If no entry exists, appends the new cron entry.
5. Writes the updated crontab.

Each job type has its own comment marker, so check and self-update jobs are managed independently. The command is idempotent. Running it multiple times updates the schedule rather than creating duplicate entries.

## Dedicated Service User

{{< callout emoji="💡" >}}
On production servers, install the cron job under a dedicated `update-watcher` system user rather than root for better security isolation.
{{< /callout >}}

On Linux servers with a dedicated `update-watcher` user, install the cron job under that user's crontab:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -e
```

And add manually:

```text {filename="Crontab"}
# update-watcher scheduled check
0 7 * * * /usr/local/bin/update-watcher run --quiet --as-service-user 2>&1 | logger -t update-watcher
```

Alternatively, use the `install-cron` command while running as the service user:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher install-cron
```

For complete server setup instructions, see [Linux Server Setup](../../server-setup/linux/).

## Updating the Schedule

To change the schedule, simply run `install-cron` again with the same `--type` and new time or expression. The existing entry is replaced:

```bash {filename="Terminal"}
# Originally set to 07:00
update-watcher install-cron --time 07:00

# Later changed to 09:00
update-watcher install-cron --time 09:00
```

## Related

- [uninstall-cron](../uninstall-cron/) -- Remove cron jobs.
- [self-update](../self-update/) -- Update the binary manually.
- [Cron Scheduling](../../server-setup/cron/) -- Detailed cron scheduling guide with logging and verification.
- [Linux Server Setup](../../server-setup/linux/) -- Production-ready server setup with dedicated user and cron.
