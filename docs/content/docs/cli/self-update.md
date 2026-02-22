---
title: "self-update - Update Update-Watcher"
description: "Update Update-Watcher to the latest version from GitHub Releases. Check for updates without installing."
weight: 7
---

The `self-update` command updates the Update-Watcher binary to the latest version from GitHub Releases. It can also check whether a newer version is available without performing the update.

## Usage

```bash {filename="Terminal"}
update-watcher self-update [--status]
```

## Flags

| Flag | Description |
|------|-------------|
| `--status` | Check for a newer version without updating. Prints the current and latest versions, then exits. |

{{< callout emoji="💡" >}}
Use `--status` in monitoring scripts to check for available updates without applying them. Pair it with exit codes for automated workflows.
{{< /callout >}}

## How It Works

When `self-update` runs without `--status`:

1. **Query GitHub Releases** -- Contacts the GitHub Releases API for the Update-Watcher repository to determine the latest available version.
2. **Compare versions** -- Compares the latest release version with the currently installed version.
3. **Download** -- If a newer version is available, downloads the appropriate binary for your OS and architecture.
4. **Replace** -- Replaces the current binary in place. The old binary is overwritten with the new one.

The update preserves your configuration file. Only the binary is replaced.

## Examples

### Update to the Latest Version

```bash {filename="Terminal"}
update-watcher self-update
```

If a newer version is available:

```text {filename="Output"}
Current version: v1.2.0
Latest version:  v1.3.0
Downloading update-watcher v1.3.0 for linux/amd64...
Updated successfully to v1.3.0
```

If already up to date:

```text {filename="Output"}
Current version: v1.3.0
Already up to date.
```

### Check for Updates Without Installing

```bash {filename="Terminal"}
update-watcher self-update --status
```

```text {filename="Output"}
Current version: v1.2.0
Latest version:  v1.3.0
Update available.
```

Or if already up to date:

```text {filename="Output"}
Current version: v1.3.0
Already up to date.
```

This is useful in scripts where you want to report on available updates without applying them.

## Permissions

The `self-update` command needs write access to the directory where the binary is installed. If the binary is installed in `/usr/local/bin` (the default), you may need to run the command with `sudo`:

```bash {filename="Terminal"}
sudo update-watcher self-update
```

## Automation

### Via Cron

Schedule automatic self-updates using the built-in cron management:

```bash {filename="Terminal"}
update-watcher install-cron --type self-update --cron-expr "0 3 * * 0"
```

This runs `self-update` every Sunday at 3:00 AM, keeping the binary current without manual intervention. See [install-cron](../install-cron/) for details.

### Via Scripts

You can combine `--status` with exit codes in scripts to build custom update workflows:

```bash {filename="Terminal"}
update-watcher self-update --status
if [ $? -eq 1 ]; then
  echo "A newer version of Update-Watcher is available"
  # Optionally trigger the update
  sudo update-watcher self-update
fi
```

## Wizard Integration

The `self-update` feature is also available from the interactive setup wizard (`update-watcher setup`). The wizard's "Self-Update" menu option checks for updates and, if an update is performed, automatically restarts the wizard with the new version.

## Related

- [version](../version/) -- Display the current version, git commit, and build date.
- [install-cron](../install-cron/) -- Schedule automated self-updates via cron.
- [Installation](../../getting-started/installation/) -- All installation methods.
