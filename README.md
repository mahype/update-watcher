# Update-Watcher

A modular CLI tool that checks for available software updates and sends notifications via Slack. Designed to run on servers, scheduled via cron.

## Features

- **APT packages** — Monitors Debian/Ubuntu package updates (with security-only filter)
- **Docker containers** — Detects newer images for running containers
- **WordPress sites** — Checks core, plugin, and theme updates across 10+ environments (ddev, Lando, Docker Compose, Bedrock, native, and more)
- **Slack notifications** — Rich Block Kit messages with security highlighting
- **Cron scheduling** — Built-in cron job management
- **Interactive setup** — Guided wizard for first-time configuration

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the latest release, and installs it to `/usr/local/bin`.

**Supported platforms:** Linux (amd64, arm64, armv7), macOS (amd64, arm64)

## Quick Start

```bash
# Interactive setup wizard
update-watcher setup

# Test run without sending notifications
update-watcher run --dry-run

# Schedule daily checks (default: 7:00 AM)
update-watcher install-cron
```

## Commands

| Command | Description |
|---|---|
| `setup` | Interactive setup wizard |
| `run` | Execute all configured checks |
| `run --dry-run` | Run checks without sending notifications |
| `status` | Show current configuration |
| `validate` | Validate configuration file |
| `watch apt` | Add APT package watcher |
| `watch docker` | Add Docker container watcher |
| `watch wordpress --path <PATH>` | Add WordPress site watcher |
| `unwatch <type>` | Remove a watcher |
| `install-cron [--time HH:MM]` | Schedule daily cron job |
| `uninstall-cron` | Remove cron job |
| `version` | Show version info |

## Configuration

Config file location:
- **Linux:** `/etc/update-watcher/config.yaml`
- **macOS:** `~/.config/update-watcher/config.yaml`

All settings can be overridden with environment variables using the `UPDATE_WATCHER_` prefix.

## Build from Source

```bash
git clone https://github.com/mahype/update-watcher.git
cd update-watcher
make build
make install
```

## License

MIT
