---
title: Getting Started
description: "Get started with Update-Watcher. Install, configure, and schedule automated update checks on Linux and macOS servers."
weight: 10
---

Update-Watcher is a single-binary CLI tool that checks for available software updates across 14 package managers and sends notifications through 16 channels. It is designed to run on servers and workstations, scheduled via cron, and requires no runtime dependencies.

This section walks you through everything you need to go from zero to a fully automated update notification pipeline.

## Where to Begin

{{< cards >}}
  {{< card link="quickstart" title="Quickstart" subtitle="Install, configure, and schedule your first check in under 5 minutes." icon="play" >}}
  {{< card link="installation" title="Installation" subtitle="All installation methods: quick script, manual download, and building from source." icon="download" >}}
  {{< card link="first-run" title="First Run" subtitle="Walk through configuration, your first update check, and your first notification." icon="check-circle" >}}
{{< /cards >}}

## Overview

The typical workflow for setting up Update-Watcher involves three steps:

1. **Install** the binary on your server or workstation.
2. **Configure** which package managers to check (watchers) and where to send alerts (notifiers) using the interactive setup wizard or a YAML config file.
3. **Schedule** a daily cron job so checks run automatically.

Update-Watcher is notification-only. It never installs updates, pulls Docker images, or modifies your system in any way. You stay in full control of when and how updates are applied.

## Supported Platforms

| Platform | Architectures |
|----------|---------------|
| Linux    | amd64, arm64, armv7 |
| macOS    | amd64 (Intel), arm64 (Apple Silicon) |

Linux distributions with first-class checker support include Debian, Ubuntu, Fedora, RHEL, Rocky Linux, AlmaLinux, Arch, Manjaro, openSUSE, SLES, and Alpine. macOS is supported via the native `softwareupdate` command and Homebrew.

## What You Will Need

- A Linux or macOS machine (server, VM, or workstation).
- Roughly 2 minutes for the quick install and interactive setup.
- Credentials for at least one notification channel (e.g., a Slack webhook URL, Telegram bot token, or SMTP server).

If you are deploying on a production Linux server and want a hardened setup with a dedicated system user, see the [Server Setup](../server-setup/) section after completing the quickstart.

## Next Steps

Start with the [Quickstart](quickstart/) to get up and running in under 5 minutes, or jump to [Installation](installation/) if you need details on a specific install method.
