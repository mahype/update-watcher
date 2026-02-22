---
title: "Linux Distribution Release Upgrade Notifications - Ubuntu, Debian, Fedora"
description: "Get notified when a new Linux distribution release is available. Supports Ubuntu (LTS-only option), Debian, and Fedora. Never miss a major distribution upgrade."
weight: 13
---

Update-Watcher's Distro checker monitors whether a newer release of your Linux distribution is available. Unlike the system package manager checkers (APT, DNF, etc.) that track individual package updates, this checker watches for entirely new distribution releases -- such as Ubuntu 24.04 LTS becoming available while you are running 22.04 LTS.

This is useful for long-term server planning and ensuring you are aware of upcoming migration targets.

## Prerequisites

{{< callout type="info" >}}
- A Linux system running Ubuntu, Debian, or Fedora.
- `lsb_release` command or a valid `/etc/os-release` file (present on virtually all modern Linux distributions).
{{< /callout >}}

## Adding via CLI

Add a Distro watcher:

```bash {filename="Terminal"}
update-watcher watch distro
```

Enable LTS-only mode (default) to only be notified about long-term support releases:

```bash {filename="Terminal"}
update-watcher watch distro --lts-only
```

Disable LTS-only to be notified about all releases including short-term support:

```bash {filename="Terminal"}
update-watcher watch distro --lts-only=false
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `lts_only` | bool | `true` | Only report LTS (Long Term Support) releases. When enabled, short-term or interim releases are ignored. This is particularly relevant for Ubuntu, which alternates between LTS and non-LTS releases. |

## YAML Configuration Example

Basic Distro configuration (LTS only, the default):

```yaml {filename="config.yaml"}
watchers:
  - type: distro
```

Notify about all releases, including non-LTS:

```yaml {filename="config.yaml"}
watchers:
  - type: distro
    lts_only: false
```

Combined with APT for complete Ubuntu server monitoring:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    security_only: true
    hide_phased: true
  - type: distro
    lts_only: true
```

## How It Works

The Distro checker performs the following steps:

{{% steps %}}

### Step 1: Identify current distribution

Reads `/etc/os-release` or runs `lsb_release` to determine the distribution name (Ubuntu, Debian, Fedora) and the currently installed version.

### Step 2: Check for newer releases

Compares the current version against the list of available releases for the detected distribution.

### Step 3: Apply LTS filter

If `lts_only` is enabled, non-LTS releases are filtered out. For example, on Ubuntu 22.04 LTS, the checker would report Ubuntu 24.04 LTS as available but not Ubuntu 23.04 or 23.10.

### Step 4: Report results

If a newer release is available, the checker reports the current version and the available version.

{{% /steps %}}

## Tips

{{< callout emoji="💡" >}}
**Ubuntu LTS Tracking:** Ubuntu releases a new LTS version every two years (April of even-numbered years). The `lts_only: true` default is designed for production servers that follow the LTS upgrade path. If you run Ubuntu on desktops or development machines and want to track interim releases, set `lts_only: false`.

**Debian:** Debian releases new stable versions roughly every two years. Since Debian does not have an LTS/non-LTS distinction in the same way Ubuntu does, the `lts_only` flag has limited effect on Debian systems. All Debian stable releases are reported.

**Fedora:** Fedora releases a new version approximately every six months. There is no LTS variant of Fedora. The `lts_only` setting is less relevant for Fedora, but the checker still reports when a newer Fedora release is available.

**Notification Frequency:** The Distro checker will report the same available release on every run until you upgrade. If you schedule Update-Watcher as a daily cron job, you will receive a daily notification about the available upgrade. To reduce noise, consider configuring your notification settings to use digest mode or adjusting the `send_policy` for this checker.

**Pairing with Package Checkers:** The Distro checker complements the system package manager checkers. While [APT](../apt/), [DNF](../dnf/), and other checkers track individual package updates within your current release, the Distro checker alerts you when an entirely new release is available for migration.
{{< /callout >}}

## Related

Send distribution upgrade notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
