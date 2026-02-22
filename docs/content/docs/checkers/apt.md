---
title: "APT Update Notifications - Monitor Debian & Ubuntu Package Updates"
description: "Automatically check for APT package updates on Debian and Ubuntu servers. Filter security-only updates, detect phased rollouts, and get notified via 16 channels."
weight: 1
---

Update-Watcher's APT checker monitors Debian and Ubuntu servers for available package updates. It classifies updates as regular or security, detects phased rollouts that your system is not yet eligible for, and reports results through any of the 16 supported notification channels.

This is the most commonly used checker. If you are running Update-Watcher on a Debian-based server, the setup wizard will auto-detect APT and offer to enable it.

## Prerequisites

{{< callout type="info" >}}
- A Debian, Ubuntu, or Debian-derivative system with `apt-get` installed.
- Sudo access for the user running Update-Watcher (unless `use_sudo` is disabled).
{{< /callout >}}

## Adding via CLI

Add an APT watcher using the `watch` command:

```bash {filename="Terminal"}
update-watcher watch apt
```

Enable security-only filtering to ignore non-security updates:

```bash {filename="Terminal"}
update-watcher watch apt --security-only
```

Disable sudo if Update-Watcher runs as root or if your environment does not require it:

```bash {filename="Terminal"}
update-watcher watch apt --no-sudo
```

Hide phased updates (enabled by default) to exclude packages that Ubuntu is gradually rolling out:

```bash {filename="Terminal"}
update-watcher watch apt --hide-phased
```

Combine flags as needed:

```bash {filename="Terminal"}
update-watcher watch apt --security-only --no-sudo
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_sudo` | bool | `true` | Run `apt-get update` with sudo. Disable if running as root or with appropriate permissions. |
| `security_only` | bool | `false` | Only report security updates. Non-security package updates are silently ignored. |
| `hide_phased` | bool | `true` | Hide phased updates that your system is not yet eligible to install. Ubuntu uses phased rollouts to gradually release updates to a percentage of machines. |

## YAML Configuration Example

Basic APT configuration in your `config.yaml`:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
```

Full APT configuration with all options:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    use_sudo: true
    security_only: false
    hide_phased: true
```

Security-only monitoring for a production server:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    security_only: true
    hide_phased: true
```

## How It Works

The APT checker performs three steps when it runs:

{{% steps %}}

### Step 1: Refresh package lists

Runs `apt-get update` (with sudo if configured) to download the latest package index files from all configured repositories.

### Step 2: Check for upgrades

Queries APT for packages that have newer versions available. The initial security classification comes from parsing the `apt list --upgradable` output for repository origins (e.g., `*-security` on Ubuntu, `Debian-Security` on Debian).

### Step 3: Detect phased rollouts and cross-check security

Simulates a full upgrade via `apt-get -s dist-upgrade` (dry-run). This serves two purposes:

1. **Phased update detection** -- Identifies packages held back due to Ubuntu's phased update mechanism. When `hide_phased` is enabled (the default), these packages are excluded from the results.
2. **Security cross-check** -- Parses the `Inst` lines from the simulation output to detect packages originating from security repositories. This catches security updates that may not be flagged by the initial `apt list` parsing, ensuring more accurate security classification.

{{% /steps %}}

The checker reports each available update with the package name, current version, and available version. Security updates are flagged separately so that notification templates can highlight them.

## FAQ

{{< details title="FAQ: What are phased updates?" >}}
Ubuntu uses a system called phased updates to gradually roll out certain package updates. Instead of releasing an update to all machines at once, Ubuntu assigns each machine a random "phase" percentage. A package update might initially be available to only 10% of machines, then 50%, then 100% over a period of days.

When `hide_phased` is enabled (the default), Update-Watcher excludes packages that your machine is not yet eligible to install. This prevents notifications about updates you cannot actually apply yet.
{{< /details >}}

{{< details title="FAQ: How does security-only filtering work?" >}}
When `security_only` is set to `true`, the checker still runs a full package list refresh, but it only reports updates that originate from the distribution's security repository. On Ubuntu, this means the `*-security` suite (e.g., `jammy-security`). Regular updates from `*-updates` are silently filtered out.

This is useful for production servers where you only care about security patches and apply feature updates on a separate schedule.
{{< /details >}}

{{< details title="FAQ: How does this compare to apticron?" >}}
Apticron is a traditional Debian tool that emails a list of available updates. Update-Watcher's APT checker provides the same core functionality but adds several advantages:

- **16 notification channels** instead of email only (Slack, Discord, Teams, Telegram, ntfy, and more).
- **Phased update detection** to avoid false positives on Ubuntu.
- **Security-only filtering** built in.
- **Unified tool** that also monitors Docker images, WordPress sites, web project dependencies, and 10 other package managers from a single binary and config file.
{{< /details >}}

## Related

Send APT update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
