---
title: "DNF Update Notifications - Monitor Fedora, RHEL & Rocky Linux Updates"
description: "Check for DNF package updates on Fedora, RHEL, Rocky Linux, and AlmaLinux. Classify security updates automatically and receive notifications via Slack, Teams & more."
weight: 2
---

Update-Watcher's DNF checker monitors Fedora, RHEL, Rocky Linux, AlmaLinux, and other DNF-based distributions for available package updates. It automatically classifies security updates separately from regular updates, so you can prioritize critical patches in your notification workflow.

The setup wizard auto-detects DNF on your system and offers to enable this checker.

## Prerequisites

{{< callout type="info" >}}
- A Fedora, RHEL, Rocky Linux, AlmaLinux, or compatible system with `dnf` installed.
- Sudo access for the user running Update-Watcher (unless `use_sudo` is disabled).
{{< /callout >}}

## Adding via CLI

Add a DNF watcher:

```bash {filename="Terminal"}
update-watcher watch dnf
```

Enable security-only filtering:

```bash {filename="Terminal"}
update-watcher watch dnf --security-only
```

Disable sudo if running as root:

```bash {filename="Terminal"}
update-watcher watch dnf --no-sudo
```

Combine flags:

```bash {filename="Terminal"}
update-watcher watch dnf --security-only --no-sudo
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_sudo` | bool | `true` | Run DNF commands with sudo. Disable if running as root or with appropriate permissions. |
| `security_only` | bool | `false` | Only report security updates. Regular package updates are silently filtered out. |

## YAML Configuration Example

Basic DNF configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: dnf
```

Full configuration with all options:

```yaml {filename="config.yaml"}
watchers:
  - type: dnf
    use_sudo: true
    security_only: false
```

Security-only monitoring for a production RHEL server:

```yaml {filename="config.yaml"}
watchers:
  - type: dnf
    security_only: true
```

## How It Works

The DNF checker performs two operations:

{{% steps %}}

### Step 1: Check for updates

Runs `dnf check-update` to list all packages with available updates. DNF returns a non-zero exit code when updates are available, which the checker handles correctly.

### Step 2: Classify security updates

Runs `dnf updateinfo list --security` to identify which available updates are security-related. Each update is tagged as either a regular update or a security update in the results.

{{% /steps %}}

The checker reports each available update with the package name, current version, available version, and whether it is a security update. Notification templates can highlight security updates differently from regular updates.

## Tips

{{< callout emoji="💡" >}}
**RHEL and Rocky Linux:** On RHEL, Rocky Linux, and AlmaLinux, the security metadata is provided by the distribution vendor. Security classification is reliable and based on the advisory information in the repositories.

**Fedora:** Fedora also provides security advisory metadata through the Bodhi update system. Security classification works the same way as on RHEL-family distributions.

**CentOS Stream:** CentOS Stream uses DNF and is fully supported. Security metadata availability depends on the specific repositories configured.
{{< /callout >}}

## Related

Send DNF update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Teams](/docs/notifiers/teams/), or any of the other [16 supported notification channels](/docs/notifiers/).
