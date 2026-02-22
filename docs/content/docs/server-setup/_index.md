---
title: "Server Setup"
description: "Production-ready server setup guides for Update-Watcher on Linux and macOS. Dedicated users, permissions, and cron scheduling."
weight: 55
---

This section covers production-ready deployment of Update-Watcher on servers. While the [Quickstart](../getting-started/quickstart/) gets you running in minutes, a production setup adds proper security boundaries: a dedicated system user, minimal sudo permissions, correct file ownership, and reliable scheduling.

## Setup Guides

{{< cards >}}
  {{< card link="linux" title="Linux Server Setup" subtitle="Dedicated user, sudoers, Docker access, and file permissions for production Linux servers." icon="server" >}}
  {{< card link="macos" title="macOS Setup" subtitle="Configuration paths, Homebrew and softwareupdate checkers, and scheduling on macOS." icon="desktop-computer" >}}
  {{< card link="cron" title="Cron Scheduling" subtitle="Automate update checks with cron. Built-in management, manual setup, and logging." icon="clock" >}}
{{< /cards >}}

## Why a Dedicated Setup

Running Update-Watcher under your personal user account works for testing and single-user workstations. On production servers, a dedicated setup provides:

- **Least privilege** -- The `update-watcher` system user has only the permissions it needs, nothing more.
- **Isolation** -- The service user cannot log in interactively and has no shell access.
- **Auditability** -- Cron jobs and log files are tied to a specific service account.
- **Security** -- Config files with API tokens and webhook URLs are readable only by the service user.

## Quick Overview

A typical Linux server setup involves:

{{% steps %}}

### Creating a dedicated `update-watcher` system user.

### Setting up the configuration directory with correct permissions.

### Configuring minimal sudo access for package manager commands.

### Granting Docker socket access if monitoring containers.

### Installing a cron job under the service user.

{{% /steps %}}

For step-by-step instructions, see [Linux Server Setup](linux/).

On macOS, the setup is simpler since most checkers do not require elevated permissions. See [macOS Setup](macos/).

## Network Requirements

{{< callout type="info" >}}
Update-Watcher is outbound-only. It does not listen on any ports or accept inbound connections. No firewall rules need to be opened for inbound traffic.
{{< /callout >}}

The only network activity is:

- **HTTPS requests** to notification services (Slack, Discord, Telegram, etc.).
- **HTTPS requests** to GitHub Releases API (for `self-update` and the OpenClaw checker).
- **Docker socket** access (local Unix socket, not a network connection) for the Docker checker.

## Next Steps

- [Linux Server Setup](linux/) -- Full guide for Debian, Ubuntu, Fedora, RHEL, Arch, openSUSE, and Alpine.
- [macOS Setup](macos/) -- Setup for macOS workstations and CI runners.
- [Cron Scheduling](cron/) -- Detailed scheduling options with logging and verification.
