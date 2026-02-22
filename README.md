# Update-Watcher

[![CI](https://img.shields.io/github/actions/workflow/status/mahype/update-watcher/test.yaml?branch=main&style=for-the-badge&label=Tests)](https://github.com/mahype/update-watcher/actions)
[![Release](https://img.shields.io/github/v/release/mahype/update-watcher?style=for-the-badge)](https://github.com/mahype/update-watcher/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)]()
[![Docs](https://img.shields.io/badge/Docs-GitHub%20Pages-blue?style=for-the-badge)](https://mahype.github.io/update-watcher/)

A modular CLI tool that checks for available software updates and sends notifications. Designed to run on servers, scheduled via cron. Single binary, no runtime dependencies.

**[Full Documentation](https://mahype.github.io/update-watcher/)**

## Features

- **14 Checkers** -- APT, DNF, Pacman, Zypper, APK, macOS, Homebrew, Snap, Flatpak, Docker, WordPress, Web Projects (npm/yarn/pnpm/Composer), Distro Release, OpenClaw
- **16 Notifiers** -- Slack, Discord, Teams, Telegram, Email, ntfy, Pushover, Gotify, Home Assistant, Google Chat, Matrix, Mattermost, Rocket.Chat, PagerDuty, Pushbullet, Webhook
- **Zero dependencies** -- Single static binary, no runtime requirements
- **Multi-platform** -- Linux (amd64, arm64, armv7), macOS (amd64, arm64)
- **Interactive setup** -- Menu-driven wizard with auto-detection
- **Self-update** -- Update the binary itself from GitHub releases
- **Cron scheduling** -- Built-in cron job management

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

## Quick Start

```bash
# Interactive setup wizard
update-watcher setup

# Run checks (notifications suppressed)
update-watcher run --notify=false

# Schedule daily checks at 7:00 AM
update-watcher install-cron
```

## Documentation

The full documentation covers installation, configuration, all checkers and notifiers, CLI reference, server setup, and more:

**[https://mahype.github.io/update-watcher/](https://mahype.github.io/update-watcher/)**

| Section | Description |
|---------|-------------|
| [Getting Started](https://mahype.github.io/update-watcher/docs/getting-started/) | Installation, quickstart, first run |
| [Configuration](https://mahype.github.io/update-watcher/docs/configuration/) | YAML reference, environment variables, security |
| [Checkers](https://mahype.github.io/update-watcher/docs/checkers/) | All 14 update checkers with options |
| [Notifiers](https://mahype.github.io/update-watcher/docs/notifiers/) | All 16 notification channels with setup guides |
| [CLI Reference](https://mahype.github.io/update-watcher/docs/cli/) | Commands, flags, exit codes |
| [Server Setup](https://mahype.github.io/update-watcher/docs/server-setup/) | Linux, macOS, cron scheduling |
| [Contributing](https://mahype.github.io/update-watcher/docs/contributing/) | Architecture, adding checkers/notifiers |

## License

MIT
