---
title: Checkers
description: "Update-Watcher supports 14 package manager checkers. Monitor system packages, Docker containers, WordPress sites, and web project dependencies."
weight: 30
---

Checkers are the core building blocks of Update-Watcher. Each checker knows how to query a specific package manager or platform for available updates, parse the results, and report them in a consistent format. Checkers are read-only -- they never install updates, pull images, or modify your system.

You can enable any combination of checkers in your configuration. The interactive setup wizard (`update-watcher setup`) auto-detects which package managers are installed on your system and offers to enable the corresponding checkers.

## All Checkers

| Checker | Platform | Description |
|---------|----------|-------------|
| [APT](apt/) | Debian, Ubuntu | System package updates via `apt-get`, with security-only filtering and phased rollout detection. |
| [DNF](dnf/) | Fedora, RHEL, Rocky, AlmaLinux | System package updates via `dnf`, with automatic security update classification. |
| [Pacman](pacman/) | Arch Linux, Manjaro | System package updates via `pacman`. |
| [Zypper](zypper/) | openSUSE, SLES | System package updates and security patches via `zypper`. |
| [APK](apk/) | Alpine Linux | System package updates via `apk`, ideal for Alpine containers. |
| [macOS](macos/) | macOS | Native macOS software updates via `softwareupdate`. |
| [Homebrew](homebrew/) | macOS, Linux | Outdated Homebrew formulae and casks. |
| [Snap](snap/) | Ubuntu, Linux | Snap package updates via `snap refresh --list`. |
| [Flatpak](flatpak/) | Linux | Flatpak application updates across all configured remotes. |
| [Docker](docker/) | Linux, macOS | Detect newer Docker images for running containers without pulling anything. |
| [WordPress](wordpress/) | Linux, macOS | WordPress core, plugin, and theme updates across multiple sites and 11 environments. |
| [Web Project](webproject/) | Linux, macOS | Outdated packages and security audits for npm, yarn, pnpm, and Composer projects. |
| [Distro](distro/) | Linux | New distribution release notifications for Ubuntu, Debian, and Fedora. |
| [OpenClaw](openclaw/) | Linux, macOS | OpenClaw application update notifications with configurable channels. |

## Categories

### System Package Managers

The primary use case for Update-Watcher. These checkers monitor the native package manager on your Linux distribution and report available system-level updates.

- [APT](apt/) -- Debian and Ubuntu
- [DNF](dnf/) -- Fedora, RHEL, Rocky Linux, AlmaLinux
- [Pacman](pacman/) -- Arch Linux, Manjaro
- [Zypper](zypper/) -- openSUSE, SLES
- [APK](apk/) -- Alpine Linux

### macOS

Monitor updates on macOS workstations and CI runners.

- [macOS](macos/) -- Native software updates via `softwareupdate`
- [Homebrew](homebrew/) -- Formulae and cask updates via `brew`

### Application Stores

Desktop Linux application stores that distribute sandboxed packages.

- [Snap](snap/) -- Canonical's Snap store
- [Flatpak](flatpak/) -- Flatpak remotes (Flathub, etc.)

### Containers

Monitor running Docker containers for image updates without pulling or modifying anything.

- [Docker](docker/) -- Image digest comparison for running containers

### Web Applications

Monitor dependencies and CMS updates for web projects and WordPress sites.

- [WordPress](wordpress/) -- Core, plugin, and theme updates across 11 environments
- [Web Project](webproject/) -- npm, yarn, pnpm, and Composer dependency updates with security audits

### System

Broader system-level checks beyond individual packages.

- [Distro](distro/) -- New distribution release notifications (e.g., Ubuntu 24.04 LTS)
- [OpenClaw](openclaw/) -- OpenClaw application updates

## Auto-Detection

{{< callout type="info" >}}
When you run `update-watcher setup`, the wizard scans your system for installed package managers and offers to enable the corresponding checkers automatically. For example, on an Ubuntu server with Docker installed, the wizard will detect and offer to enable both APT and Docker checkers. You can always add or remove checkers later through the wizard or by editing the YAML configuration file directly.
{{< /callout >}}

## Next Steps

- [Getting Started](../getting-started/) -- Install Update-Watcher and run your first check.
- [Configuration](../configuration/) -- Full YAML configuration reference.
- [Notifiers](../notifiers/) -- Configure where update notifications are sent.
