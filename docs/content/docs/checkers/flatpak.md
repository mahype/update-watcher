---
title: "Flatpak Update Notifications - Monitor Linux Flatpak Application Updates"
description: "Automatically check for Flatpak application updates on Linux. Get notified of available updates via Slack, Discord, Email, Telegram and 12 more channels."
weight: 9
---

Update-Watcher's Flatpak checker monitors installed Flatpak applications for available updates across all configured remotes (e.g., Flathub). Flatpak is a universal Linux packaging format used for desktop applications and is available on most Linux distributions.

The setup wizard auto-detects Flatpak and offers to enable this checker on systems where the `flatpak` command is available.

## Prerequisites

{{< callout type="info" >}}
- A Linux system with Flatpak installed and configured.
- At least one Flatpak remote (e.g., Flathub) configured.
- At least one Flatpak application installed.
{{< /callout >}}

## Adding via CLI

Add a Flatpak watcher:

```bash {filename="Terminal"}
update-watcher watch flatpak
```

The Flatpak checker has no additional configuration flags.

## Configuration Reference

The Flatpak checker has no checker-specific options.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| (none) | -- | -- | The Flatpak checker requires no additional configuration. |

## YAML Configuration Example

Basic Flatpak configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: flatpak
```

Combined with system package manager and Snap for full Linux desktop coverage:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    hide_phased: true
  - type: snap
  - type: flatpak
```

## How It Works

The Flatpak checker runs a single command:

```text {filename="Output"}
flatpak remote-ls --updates
```

This queries all configured Flatpak remotes for updates to installed applications and runtimes. It does not download or install any updates -- it only lists what is available.

The checker parses the output to extract the application name, current version, and available version for each outdated Flatpak.

## Tips

{{< callout emoji="💡" >}}
**Multiple Remotes:** If you have multiple Flatpak remotes configured (e.g., Flathub and a corporate remote), the checker queries all of them. Updates from all remotes are included in the results.

**Runtimes and Applications:** The checker reports updates for both Flatpak applications and the underlying runtimes. Runtime updates (e.g., `org.freedesktop.Platform`) are important for security, as they provide the shared libraries used by Flatpak applications.

**Desktop Workstations:** Flatpak is primarily used on desktop Linux workstations. If you manage a fleet of Linux desktops and want to track update compliance, combining this checker with a notification channel like [Email](/docs/notifiers/email/) or [Slack](/docs/notifiers/slack/) provides centralized visibility.

**Server Environments:** Flatpak is uncommon on servers. If you are monitoring a server, the system package manager checkers ([APT](../apt/), [DNF](../dnf/), [Pacman](../pacman/), etc.) are more relevant.
{{< /callout >}}

## Related

Send Flatpak update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
