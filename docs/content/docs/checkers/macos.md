---
title: "macOS Update Notifications - Monitor macOS Software Updates"
description: "Check for macOS software updates automatically via softwareupdate. Filter security-only updates and receive notifications via Slack, Discord, Email & more."
weight: 6
---

Update-Watcher's macOS checker monitors Apple's native software update system for available updates, including macOS version upgrades, security patches, and system component updates. It uses the built-in `softwareupdate` command-line tool that ships with every Mac.

The setup wizard auto-detects macOS and offers to enable this checker on Apple systems.

## Prerequisites

{{< callout type="info" >}}
- A macOS system (any supported version).
- No additional software required -- `softwareupdate` is built into macOS.
{{< /callout >}}

## Adding via CLI

Add a macOS watcher:

```bash {filename="Terminal"}
update-watcher watch macos
```

Enable security-only filtering to ignore non-security updates:

```bash {filename="Terminal"}
update-watcher watch macos --security-only
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `security_only` | bool | `false` | Only report security updates. Feature updates and non-security system updates are silently filtered out. |

## YAML Configuration Example

Basic macOS configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: macos
```

Security-only monitoring:

```yaml {filename="config.yaml"}
watchers:
  - type: macos
    security_only: true
```

## How It Works

The macOS checker runs the following command:

```text {filename="Output"}
softwareupdate -l
```

This lists all available software updates from Apple's update servers. The checker parses the output to extract:

- **Update name** -- The display name of the update (e.g., "macOS Sonoma 14.3" or "Security Update 2024-001").
- **Version** -- The version number or build identifier.
- **Update type** -- Whether the update is flagged as a security update or a recommended/regular update.

When `security_only` is enabled, only updates that Apple classifies as security-related are included in the results.

## Tips

{{< callout emoji="💡" >}}
**Scheduled Checks on macOS:** macOS servers and workstations that run Update-Watcher via cron or launchd benefit from regular checks. Apple releases security updates on an irregular schedule, and this checker ensures you are notified promptly.

**Combining with Homebrew:** On most macOS systems, you will want to enable both the macOS checker and the [Homebrew checker](../homebrew/). The macOS checker covers system-level updates from Apple, while Homebrew covers third-party packages and applications installed via `brew`.

**CI and Build Servers:** If you manage macOS CI runners (GitHub Actions self-hosted runners, Jenkins agents, etc.), this checker helps ensure the underlying macOS system stays patched. Pair it with a [Slack](/docs/notifiers/slack/) or [Teams](/docs/notifiers/teams/) notifier to alert your infrastructure team.
{{< /callout >}}

Combined macOS and Homebrew configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: macos
    security_only: true
  - type: homebrew
    include_casks: true
```

## Related

Send macOS update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
