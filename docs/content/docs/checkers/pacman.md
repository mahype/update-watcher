---
title: "Pacman Update Notifications - Monitor Arch Linux & Manjaro Updates"
description: "Automatically check for Pacman package updates on Arch Linux and Manjaro. Get notified via Slack, Discord, Telegram, Email and 12 more notification channels."
weight: 3
---

Update-Watcher's Pacman checker monitors Arch Linux, Manjaro, and other Pacman-based distributions for available package updates. It syncs the package database and lists all upgradeable packages, reporting results through any of the 16 supported notification channels.

The setup wizard auto-detects Pacman and offers to enable this checker on Arch-based systems.

## Prerequisites

{{< callout type="info" >}}
- An Arch Linux, Manjaro, or Pacman-based system.
- Sudo access for the user running Update-Watcher (unless `use_sudo` is disabled). Syncing the package database typically requires root privileges.
{{< /callout >}}

## Adding via CLI

Add a Pacman watcher:

```bash {filename="Terminal"}
update-watcher watch pacman
```

Disable sudo if running as root:

```bash {filename="Terminal"}
update-watcher watch pacman --no-sudo
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_sudo` | bool | `true` | Run `pacman -Sy` with sudo. Disable if running as root or with appropriate permissions. |

## YAML Configuration Example

Basic Pacman configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: pacman
```

Configuration with sudo disabled (e.g., running as root):

```yaml {filename="config.yaml"}
watchers:
  - type: pacman
    use_sudo: false
```

## How It Works

The Pacman checker performs two steps:

{{% steps %}}

### Step 1: Sync package database

Runs `pacman -Sy` (with sudo if configured) to synchronize the local package database with the remote repositories. This downloads the latest package lists without upgrading anything.

### Step 2: List available upgrades

Runs `pacman -Qu` to query the local database for packages where a newer version is available in the synced repositories.

{{% /steps %}}

The checker reports each available update with the package name, currently installed version, and available version.

## Tips

{{< callout emoji="💡" >}}
**AUR Packages:** The Pacman checker only monitors official repository packages. AUR (Arch User Repository) packages managed through helpers like `yay` or `paru` are not included. The checker focuses on packages from the repositories configured in `/etc/pacman.conf`.

**Manjaro and Derivatives:** Manjaro uses the same Pacman package manager with its own repositories. The checker works identically on Manjaro -- it syncs from whatever repositories are configured on the system.

**Rolling Release Considerations:** Arch Linux is a rolling release distribution, so the Pacman checker may report a large number of updates on systems that have not been updated recently. Running Update-Watcher on a daily cron schedule helps you stay aware of accumulated updates.
{{< /callout >}}

## Related

Send Pacman update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
