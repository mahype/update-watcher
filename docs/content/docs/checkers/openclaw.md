---
title: "OpenClaw Update Notifications - Monitor OpenClaw Application Updates"
description: "Check for OpenClaw application updates with configurable update channels. Receive notifications when new versions are available."
weight: 14
---

Update-Watcher's OpenClaw checker monitors the OpenClaw application for available updates. It supports configurable update channels, allowing you to track stable releases, beta builds, or other release channels depending on your needs.

## Prerequisites

{{< callout type="info" >}}
- OpenClaw installed on the system.
- Network access to check the configured update channel.
{{< /callout >}}

## Adding via CLI

Add an OpenClaw watcher:

```bash {filename="Terminal"}
update-watcher watch openclaw
```

Specify an update channel:

```bash {filename="Terminal"}
update-watcher watch openclaw --channel stable
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `channel` | string | `""` | The update channel to check for new releases. Leave empty for the default channel. Set to a specific channel name (e.g., `stable`, `beta`) to track that release stream. |

## YAML Configuration Example

Basic OpenClaw configuration (default channel):

```yaml {filename="config.yaml"}
watchers:
  - type: openclaw
```

Track a specific update channel:

```yaml {filename="config.yaml"}
watchers:
  - type: openclaw
    channel: stable
```

Track the beta channel:

```yaml {filename="config.yaml"}
watchers:
  - type: openclaw
    channel: beta
```

Combined with other checkers:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    security_only: true
  - type: docker
  - type: openclaw
    channel: stable
```

## How It Works

The OpenClaw checker queries the configured update channel for new releases. It compares the currently installed version against the latest available version in the specified channel and reports when a newer version is available.

{{< callout type="info" >}}
The check is a lightweight network request that retrieves version metadata from the update channel endpoint. No downloads or installations are performed.
{{< /callout >}}

## Tips

{{< callout emoji="💡" >}}
**Update Channels:** Update channels allow you to control which release stream you track. Common channel configurations include:

- **Default (empty string)** -- Uses the default release channel configured by the OpenClaw installation.
- **stable** -- Track stable production releases only.
- **beta** -- Track beta releases for early testing.

**Pairing with System Checkers:** If OpenClaw is running on a server alongside other services, combine it with the appropriate system package manager checker for comprehensive update monitoring.
{{< /callout >}}

Combined with system package manager:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    hide_phased: true
  - type: openclaw
    channel: stable
```

## Related

Send OpenClaw update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
