---
title: "Zypper Update Notifications - Monitor openSUSE & SLES Updates"
description: "Check for Zypper package updates on openSUSE and SLES. Filter security patches, configure sudo access, and receive automated notifications via 16 channels."
weight: 4
---

Update-Watcher's Zypper checker monitors openSUSE and SUSE Linux Enterprise Server (SLES) for available package updates and security patches. It refreshes repositories, lists available updates, and optionally filters for security-only patches.

The setup wizard auto-detects Zypper and offers to enable this checker on openSUSE and SLES systems.

## Prerequisites

{{< callout type="info" >}}
- An openSUSE, SLES, or Zypper-based system.
- Sudo access for the user running Update-Watcher (unless `use_sudo` is disabled).
{{< /callout >}}

## Adding via CLI

Add a Zypper watcher:

```bash {filename="Terminal"}
update-watcher watch zypper
```

Enable security-only filtering:

```bash {filename="Terminal"}
update-watcher watch zypper --security-only
```

Disable sudo:

```bash {filename="Terminal"}
update-watcher watch zypper --no-sudo
```

Combine flags:

```bash {filename="Terminal"}
update-watcher watch zypper --security-only --no-sudo
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_sudo` | bool | `true` | Run Zypper commands with sudo. Disable if running as root. |
| `security_only` | bool | `false` | Only report security patches. Regular updates are silently filtered out. |

## YAML Configuration Example

Basic Zypper configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: zypper
```

Full configuration with all options:

```yaml {filename="config.yaml"}
watchers:
  - type: zypper
    use_sudo: true
    security_only: false
```

Security-only monitoring for a production SLES server:

```yaml {filename="config.yaml"}
watchers:
  - type: zypper
    security_only: true
```

## How It Works

The Zypper checker performs three operations:

{{% steps %}}

### Step 1: Refresh repositories

Runs `zypper refresh` to download the latest package metadata from all configured repositories.

### Step 2: List available updates

Runs `zypper list-updates` to enumerate all packages with newer versions available.

### Step 3: Identify security patches

Runs `zypper list-patches --category security` to identify which available patches are classified as security updates. When `security_only` is enabled, only these patches are reported.

{{% /steps %}}

The checker reports each available update with the package name, current version, and available version. Security patches are flagged separately.

## Tips

{{< callout emoji="💡" >}}
**openSUSE Tumbleweed:** openSUSE Tumbleweed is a rolling release distribution. Like Arch Linux, it may report a large number of updates if the system has not been updated recently. Daily cron scheduling keeps you informed of accumulated updates.

**openSUSE Leap and SLES:** openSUSE Leap and SLES use a traditional release model with well-defined security advisories. Security-only filtering is especially useful on these platforms for production servers where you want to prioritize security patches.

**Repository Refresh:** The `zypper refresh` step may take a few seconds depending on the number of configured repositories and network speed. This is the same operation you would run manually before checking for updates.
{{< /callout >}}

## Related

Send Zypper update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Teams](/docs/notifiers/teams/), or any of the other [16 supported notification channels](/docs/notifiers/).
