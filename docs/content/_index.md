---
title: Update-Watcher
description: "Monitor 14 package managers for available updates and get notified via Slack, Discord, Teams, Telegram, Email & more. Single binary, zero dependencies, cron-ready."
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <span>Open Source</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-headline >}}
  Never Miss a Server Update Again
{{< /hextra/hero-headline >}}
</div>

<div class="hx-mb-12">
{{< hextra/hero-subtitle >}}
  A single binary CLI tool that checks for available software updates&nbsp;<br class="sm:hx-block hx-hidden" />across 14 package managers and sends notifications through 16 channels.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx-mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started/quickstart" >}}
{{< hextra/hero-button text="View on GitHub" link="https://github.com/mahype/update-watcher" style="margin-left: 8px;" >}}
</div>

<div class="hx-mt-6"></div>

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

<div class="hx-mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="14 Package Managers"
    subtitle="APT, DNF, Pacman, Zypper, APK, macOS, Homebrew, Snap, Flatpak, Docker, WordPress, npm/yarn/pnpm/Composer, Distro Release, and OpenClaw."
    class="hx-aspect-auto md:hx-aspect-[1.1/1] max-md:hx-min-h-[340px]"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(59,130,246,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="16 Notification Channels"
    subtitle="Slack, Discord, Teams, Telegram, Email, ntfy, Pushover, Gotify, Home Assistant, Google Chat, Matrix, Mattermost, Rocket.Chat, PagerDuty, Pushbullet, and Webhooks."
    class="hx-aspect-auto md:hx-aspect-[1.1/1] max-md:hx-min-h-[340px]"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(142,53,234,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Zero Dependencies"
    subtitle="Single static binary. No runtime dependencies, no database, no Docker required. Just download, configure, and schedule with cron."
    class="hx-aspect-auto md:hx-aspect-[1.1/1] max-md:hx-min-h-[340px]"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(16,185,129,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Security Update Highlighting"
    subtitle="Automatically classifies and highlights security updates. Supports @mentions and priority escalation for critical patches."
  >}}
  {{< hextra/feature-card
    title="Interactive Setup Wizard"
    subtitle="Menu-driven TUI wizard that auto-detects installed package managers and guides you through configuration."
  >}}
  {{< hextra/feature-card
    title="Notification Only"
    subtitle="Read-only checks. Never installs updates, pulls Docker images, or modifies your system. You stay in full control."
  >}}
{{< /hextra/feature-grid >}}

## How It Works

Update-Watcher is designed for servers and workstations managed via cron. The workflow is simple:

{{% steps %}}

### Install

Download the single binary on your server with one command.

### Configure

Set up watchers (what to check) and notifiers (where to send alerts) via the interactive wizard or YAML config.

### Schedule

Add a cron job for daily checks -- get notified only when updates are available.

{{% /steps %}}

## Supported Platforms

| Platform | Architectures |
|----------|--------------|
| Linux    | amd64, arm64, armv7 |
| macOS    | amd64 (Intel), arm64 (Apple Silicon) |

## Frequently Asked Questions

{{< details title="What is Update-Watcher?" >}}
Update-Watcher is an open-source CLI tool that checks for available software updates across 14 package managers -- including APT, DNF, Pacman, Docker, Homebrew, WordPress, npm, yarn, pnpm, and Composer -- and sends notifications through 16 channels including Slack, Discord, Teams, Telegram, and Email.
{{< /details >}}

{{< details title="Does Update-Watcher automatically install updates?" >}}
No. Update-Watcher is notification-only. It checks for available updates and sends alerts -- it never modifies packages, pulls Docker images, or changes your system. You decide when and how to apply updates.
{{< /details >}}

{{< details title="Is Update-Watcher free?" >}}
Yes. Update-Watcher is free, open-source software released under the [MIT license](https://github.com/mahype/update-watcher/blob/main/LICENSE).
{{< /details >}}

{{< details title="How does Update-Watcher compare to unattended-upgrades?" >}}
unattended-upgrades automatically installs updates on Debian/Ubuntu only. Update-Watcher supports 14 package managers across multiple distributions and macOS, notifies you through 16 channels, but never installs updates -- giving you full control. See the [comparison page](docs/comparison) for a detailed feature comparison.
{{< /details >}}
