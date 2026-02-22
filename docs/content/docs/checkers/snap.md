---
title: "Snap Update Notifications - Monitor Ubuntu Snap Package Updates"
description: "Check for available Snap package updates on Ubuntu and Linux. Receive automated update notifications via Slack, Discord, Telegram and 13 more channels."
weight: 8
---

Update-Watcher's Snap checker monitors installed Snap packages for available updates. Snaps are self-contained application packages used primarily on Ubuntu, but also available on other Linux distributions that support the Snap daemon.

The setup wizard auto-detects Snap and offers to enable this checker on systems where `snapd` is installed.

## Prerequisites

{{< callout type="info" >}}
- A Linux system with `snapd` installed and the `snap` command available.
- At least one Snap package installed on the system.
{{< /callout >}}

## Adding via CLI

Add a Snap watcher:

```bash {filename="Terminal"}
update-watcher watch snap
```

The Snap checker has no additional configuration flags.

## Configuration Reference

The Snap checker has no checker-specific options. It uses the default settings.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| (none) | -- | -- | The Snap checker requires no additional configuration. |

## YAML Configuration Example

Basic Snap configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: snap
```

Combined with APT for full Ubuntu coverage:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    security_only: false
    hide_phased: true
  - type: snap
```

## How It Works

The Snap checker runs a single command:

```text {filename="Output"}
snap refresh --list
```

This command queries the Snap daemon for all installed snaps that have newer revisions available in the store. It does not actually refresh (update) any snaps -- it only lists what is available.

The checker parses the output to extract the snap name, current version, available version, and the channel (e.g., `stable`, `edge`) the snap is tracking.

## Tips

{{< callout emoji="💡" >}}
**Snap Auto-Updates:** By default, Snap packages auto-update in the background. However, the Snap daemon may defer updates for various reasons (metered connections, snap in use, held refreshes, etc.). The Snap checker catches these deferred updates and notifies you that newer versions are pending.

This is particularly useful in server environments where you want visibility into what is about to change, or on systems where you have disabled or deferred Snap auto-updates.

**Server Environments:** On Ubuntu Server, several system components are delivered as Snaps (e.g., `lxd`, `snapd` itself). Monitoring these with the Snap checker alongside the [APT checker](../apt/) gives you complete visibility into available system updates.
{{< /callout >}}

Combining with other checkers for a complete Ubuntu monitoring setup:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    hide_phased: true
  - type: snap
  - type: flatpak
```

## Related

Send Snap update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
