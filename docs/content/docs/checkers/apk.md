---
title: "APK Update Notifications - Monitor Alpine Linux Package Updates"
description: "Automatically check for APK package updates on Alpine Linux servers and containers. Lightweight, cron-ready update notifications via 16 channels."
weight: 5
---

Update-Watcher's APK checker monitors Alpine Linux systems for available package updates. Alpine is widely used in Docker containers and lightweight server deployments, making this checker particularly useful for monitoring Alpine-based infrastructure.

The setup wizard auto-detects APK and offers to enable this checker on Alpine systems.

## Prerequisites

{{< callout type="info" >}}
- An Alpine Linux system with `apk` installed.
- Sudo access if required by your environment (Alpine containers typically run as root, so `use_sudo` defaults to `false`).
{{< /callout >}}

## Adding via CLI

Add an APK watcher:

```bash {filename="Terminal"}
update-watcher watch apk
```

Disable sudo (the default for APK, since Alpine containers typically run as root):

```bash {filename="Terminal"}
update-watcher watch apk --no-sudo
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_sudo` | bool | `false` | Run APK commands with sudo. Defaults to `false` because Alpine containers typically run as root. Enable if running as a non-root user on a full Alpine installation. |

## YAML Configuration Example

Basic APK configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: apk
```

Configuration with sudo enabled for a non-root Alpine installation:

```yaml {filename="config.yaml"}
watchers:
  - type: apk
    use_sudo: true
```

## How It Works

The APK checker performs two steps:

{{% steps %}}

### Step 1: Update package index

Runs `apk update` to fetch the latest package index from configured repositories.

### Step 2: List upgradeable packages

Runs `apk version -v -l '<'` to list all installed packages where the installed version is older than the version available in the repository. The `<` filter selects only packages that have newer versions available.

{{% /steps %}}

The checker reports each available update with the package name, currently installed version, and available version.

## Tips

{{< callout emoji="💡" >}}
**Alpine in Docker Containers:** If you are running Update-Watcher inside an Alpine-based Docker container, the default `use_sudo: false` setting is correct since containers typically run as root. If you run Update-Watcher on the host and check updates for a containerized Alpine system, consider using the [Docker checker](../docker/) instead.

**Minimal Footprint:** Alpine Linux is designed for minimal footprint environments. The APK checker is correspondingly lightweight -- it only invokes the two `apk` commands listed above and parses their text output.

**Edge and Testing Repositories:** The APK checker reports updates from all repositories configured in `/etc/apk/repositories`. If you have `edge` or `testing` repositories enabled, updates from those sources will also appear in the results.
{{< /callout >}}

## Related

Send APK update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
