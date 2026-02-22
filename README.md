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

With server setup (dedicated user, sudoers, cron):

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --server
```

Without server setup (binary only):

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --no-server
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

## Linux Server: Recommended Setup

On production servers, run update-watcher under a **dedicated system user** with minimal privileges (principle of least privilege).

#### 1. Create a dedicated user

```bash
sudo useradd -r -s /usr/sbin/nologin -m -d /var/lib/update-watcher update-watcher
```

#### 2. Set up config directory and permissions

```bash
sudo mkdir -p /etc/update-watcher
sudo chown update-watcher:update-watcher /etc/update-watcher
sudo chmod 755 /etc/update-watcher

# Config file must be owner-readable only (contains tokens/secrets)
sudo touch /etc/update-watcher/config.yaml
sudo chown update-watcher:update-watcher /etc/update-watcher/config.yaml
sudo chmod 600 /etc/update-watcher/config.yaml
```

#### 3. Set up log file (optional)

```bash
sudo touch /var/log/update-watcher.log
sudo chown update-watcher:update-watcher /var/log/update-watcher.log
sudo chmod 640 /var/log/update-watcher.log
```

Then add `log_file: "/var/log/update-watcher.log"` to the `settings` section in your config.

#### 4. Grant sudo rights for package manager refresh (optional)

If you want update-watcher to refresh package lists before checking (e.g. `apt-get update`), create `/etc/sudoers.d/update-watcher`:

```bash
sudo visudo -f /etc/sudoers.d/update-watcher
```

Add the lines for your package manager(s):

```
# APT (Debian/Ubuntu)
update-watcher ALL=(root) NOPASSWD: /usr/bin/apt-get update

# DNF (Fedora/RHEL)
update-watcher ALL=(root) NOPASSWD: /usr/bin/dnf check-update
update-watcher ALL=(root) NOPASSWD: /usr/bin/dnf updateinfo list --security

# Pacman (Arch)
update-watcher ALL=(root) NOPASSWD: /usr/bin/pacman -Sy

# Zypper (openSUSE)
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive refresh
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-patches --category security
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-updates

# APK (Alpine)
update-watcher ALL=(root) NOPASSWD: /sbin/apk update
```

**Alternative:** If your server already refreshes package lists automatically (e.g. via `unattended-upgrades`), you can skip this and set `use_sudo: false` in the checker options.

#### 5. Docker access (optional)

If you want to monitor Docker containers, add the user to the `docker` group:

```bash
sudo usermod -aG docker update-watcher
```

#### 6. WordPress / Web project access (optional)

If you monitor WordPress sites or web projects, grant read access via group membership:

```bash
sudo usermod -aG www-data update-watcher
```

#### 7. Schedule via cron

```bash
sudo crontab -u update-watcher -e
```

Add a line like:

```
0 7 * * * /usr/local/bin/update-watcher run --quiet
```

Or use the built-in command (run as the `update-watcher` user):

```bash
sudo -u update-watcher update-watcher install-cron
```

#### Summary

| Resource | Path | Permissions |
|---|---|---|
| Binary | `/usr/local/bin/update-watcher` | `0755`, root:root |
| Config directory | `/etc/update-watcher/` | `0755`, update-watcher:update-watcher |
| Config file | `/etc/update-watcher/config.yaml` | `0600`, update-watcher:update-watcher |
| Log file | `/var/log/update-watcher.log` | `0640`, update-watcher:update-watcher |
| Sudoers | `/etc/sudoers.d/update-watcher` | `0440`, root:root |

The application requires **no inbound network ports**, no database, and no persistent state beyond the config file. All network access is outbound-only (HTTPS to notification services).

For the full server setup guide with additional details, see the [Server Setup documentation](https://mahype.github.io/update-watcher/docs/server-setup/).

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
