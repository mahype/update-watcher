---
title: "uninstall-cron - Remove Cron Jobs"
description: "Remove Update-Watcher cron jobs from the user's crontab. Remove individual job types or all jobs at once."
weight: 9
---

The `uninstall-cron` command removes Update-Watcher cron jobs from the current user's crontab. You can remove a specific job type or all Update-Watcher jobs at once.

## Usage

```bash {filename="Terminal"}
update-watcher uninstall-cron [--type TYPE] [--all]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type TYPE` | string | `check` | Job type to remove: `check` or `self-update`. |
| `--all` | bool | `false` | Remove all Update-Watcher cron jobs regardless of type. |

## How It Works

When the [install-cron](../install-cron/) command creates a cron entry, it includes an identifying comment for each job type:

```text {filename="Crontab"}
# update-watcher scheduled check
0 7 * * * /usr/local/bin/update-watcher run --quiet 2>&1 | logger -t update-watcher
# update-watcher self-update
0 3 * * 0 /usr/local/bin/update-watcher self-update 2>&1 | logger -t update-watcher
```

The `uninstall-cron` command:

1. Reads the current user's crontab.
2. Searches for lines matching the comment marker for the specified job type (or all types with `--all`).
3. Removes the comment line and the associated command line.
4. Writes the updated crontab.

## Examples

### Remove the Check Job

```bash {filename="Terminal"}
update-watcher uninstall-cron
```

```text {filename="Output"}
Update Check cron job removed successfully.
```

### Remove the Self-Update Job

```bash {filename="Terminal"}
update-watcher uninstall-cron --type self-update
```

```text {filename="Output"}
Self-Update cron job removed successfully.
```

### Remove All Jobs

```bash {filename="Terminal"}
update-watcher uninstall-cron --all
```

```text {filename="Output"}
All update-watcher cron jobs removed.
```

### Verify Removal

After removing cron jobs, verify they are gone:

```bash {filename="Terminal"}
crontab -l
```

The Update-Watcher entries should no longer appear in the output.

## Dedicated Service User

If the cron jobs were installed under a dedicated system user, remove them with:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher uninstall-cron --all
```

Or edit the user's crontab directly:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -e
```

## Related

- [install-cron](../install-cron/) -- Install cron jobs for automated tasks.
- [Cron Scheduling](../../server-setup/cron/) -- Detailed cron scheduling guide.
