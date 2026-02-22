---
title: "setup - Interactive Setup Wizard"
description: "Launch the interactive menu-driven setup wizard. Auto-detects installed package managers and guides you through configuration."
weight: 2
---

The `setup` command launches Update-Watcher's interactive menu-driven setup wizard. It is the recommended way to create and modify your configuration. The wizard auto-detects installed package managers, guides you through enabling watchers and notifiers, and writes the configuration file with correct permissions.

## Usage

```bash {filename="Terminal"}
update-watcher setup
```

The `setup` command takes no additional flags beyond the global flags (`--config`, `--quiet`, `--verbose`).

## What the Wizard Does

When you launch the wizard, it performs the following:

{{% steps %}}

### Step 1: Load Existing Configuration

If a config file already exists at the resolved path, the wizard loads it and shows the current state. Otherwise, it starts with an empty configuration.

### Step 2: Auto-Detect Package Managers

The wizard scans your system for installed tools (`apt-get`, `dnf`, `pacman`, `brew`, `docker`, etc.) and marks them as available for configuration.

### Step 3: Present the Main Menu

A menu-driven interface lets you configure each section of Update-Watcher.

### Step 4: Save on Exit

When you select "Save & Exit," the wizard writes the configuration to the appropriate file path.

{{% /steps %}}

## Main Menu

After launching the wizard, you see a summary of the current configuration followed by the main menu:

```text {filename="Setup Wizard"}
=== update-watcher setup (0.14.0) ===

  Hostname:      web-prod-01
  Watchers:      APT, Docker
  Notifiers:     slack configured
  Cron:          daily at 07:00

What would you like to do?
  > Manage Watchers (APT, Docker)
  > Notifications (slack configured)
  > Settings (hostname: web-prod-01)
  > Cron Job (daily at 07:00)
  > Run Test Check
  > Send Test Notification
  > Self-Update
  > Save & Exit
```

### Manage Watchers

Lists all checker types that are available on your system, with detected tools marked. You can enable or disable individual checkers and configure their options (e.g., security-only filtering for APT, LTS-only for distro checks).

For checkers like WordPress and Web Project that support multiple instances, the wizard lets you add, edit, or remove individual sites and projects.

### Notifications

Guides you through configuring notification channels. The wizard prompts for the required credentials for each notifier type (e.g., webhook URLs for Slack, bot tokens for Telegram, SMTP settings for email).

You can configure multiple notifiers to receive the same update report through different channels simultaneously.

### Settings

Configure global settings:

- **Hostname** -- The server name included in notifications. Auto-detected if left empty.
- **Send Policy** -- Choose between `only-on-updates` (default) and `always`.
- **Log File** -- Optional log file path for persistent logging.

### Cron Job

Set up or modify the cron schedule for automated update checks. The wizard can:

- Install a cron job at a specified time (default: 07:00 daily).
- Show the current cron status.
- Remove an existing cron job.

This is equivalent to the [install-cron](../install-cron/) and [uninstall-cron](../uninstall-cron/) commands.

### Run Test Check

Runs all configured checkers immediately and displays the results in the terminal without sending notifications. This lets you verify that your watchers are working correctly before saving and exiting.

The wizard remains open after the test check completes, so you can review the results and make adjustments.

### Send Test Notification

Sends a test notification through all configured notifiers. This lets you verify that your notification channels are working correctly before relying on scheduled checks.

This option only appears when at least one notifier is configured.

### Self-Update

Checks for a newer version of Update-Watcher on GitHub Releases and offers to download and install it. If an update is performed, the wizard automatically restarts with the new version.

The current version is always shown in the wizard title bar: `=== update-watcher setup (0.14.0) ===`.

### Save and Exit

Writes the configuration to the appropriate file path and exits. On Linux, the system-wide path `/etc/update-watcher/config.yaml` is used for service user setups. On macOS and per-user setups, the file is written to `~/.config/update-watcher/config.yaml`.

## Auto-Detection

{{< callout type="info" >}}
The wizard automatically detects installed package managers on your system and marks them as available. You do not need to know which tools are installed -- the wizard handles this for you.
{{< /callout >}}

The wizard detects the following tools and maps them to checkers:

| Detected Tool | Checker Type |
|---------------|-------------|
| `apt-get` | apt |
| `dnf` | dnf |
| `pacman` | pacman |
| `zypper` | zypper |
| `apk` | apk |
| `softwareupdate` | macos |
| `brew` | homebrew |
| `snap` | snap |
| `flatpak` | flatpak |
| `docker` | docker |

WordPress and Web Project checkers are not auto-detected since they require explicit paths. Distro and OpenClaw checkers are always available regardless of detected tools.

## Example Session

A typical first-time setup session on an Ubuntu server with Docker:

```text {filename="Setup Wizard"}
$ update-watcher setup

=== update-watcher setup (0.14.0) ===

  Hostname:      (auto-detect)
  Watchers:      (none configured)
  Notifiers:     (none configured)
  Cron:          not configured

What would you like to do?
  > Manage Watchers (none configured)
```

After enabling APT and Docker, configuring Slack notifications, and setting up a daily cron job, the wizard writes the configuration and exits.

## Related

- [Quickstart](../../getting-started/quickstart/) -- End-to-end setup guide using the wizard.
- [Configuration](../../configuration/) -- Full YAML configuration reference for manual editing.
- [Checkers](../../checkers/) -- All 14 checker types with configuration details.
- [Notifiers](../../notifiers/) -- All 16 notification channels.
