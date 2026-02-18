# 🔄 Update-Watcher

[![CI](https://img.shields.io/github/actions/workflow/status/mahype/update-watcher/test.yaml?branch=main&style=for-the-badge&label=Tests)](https://github.com/mahype/update-watcher/actions)
[![Release](https://img.shields.io/github/v/release/mahype/update-watcher?style=for-the-badge)](https://github.com/mahype/update-watcher/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)]()

A modular CLI tool that checks for available software updates and sends notifications. Designed to run on servers, scheduled via cron. Single binary, no runtime dependencies.

## ✨ Features

### 📦 Checkers

- 🐧 **APT** — Debian/Ubuntu package updates (with security-only filter and phased rollout detection)
- 🎩 **DNF** — Fedora/RHEL/Rocky/AlmaLinux package updates (with security classification)
- 👻 **Pacman** — Arch/Manjaro package updates
- 🦎 **Zypper** — openSUSE/SLES package updates (with security patches)
- 🏔️ **APK** — Alpine Linux package updates
- 🍎 **macOS** — macOS software updates via `softwareupdate` (with security-only filter)
- 🍺 **Homebrew** — macOS/Linux Homebrew package updates (formulae and casks)
- 📦 **Snap** — Ubuntu/Linux Snap package updates
- 📦 **Flatpak** — Linux Flatpak application updates
- 🐳 **Docker** — Detects newer images for running containers (read-only, no image pulls)
- 📝 **WordPress** — Core, plugin, and theme updates across 11 environments
- 📦 **Web Projects** — Outdated packages and security audits for npm, yarn, pnpm, and Composer

### 🔔 Notifiers

- 💬 **Slack** — Rich Block Kit messages with security highlighting
- 🎮 **Discord** — Embedded messages via webhooks
- 🟦 **Microsoft Teams** — Adaptive Card messages via Workflow webhooks
- ✈️ **Telegram** — Bot API messages with Markdown formatting
- 📧 **Email** — HTML emails via SMTP (with STARTTLS)
- 📲 **ntfy** — Push notifications via [ntfy.sh](https://ntfy.sh) or self-hosted
- 📢 **Pushover** — Push notifications for iOS, Android, Desktop
- 🔔 **Gotify** — Push notifications via self-hosted Gotify server
- 🏠 **Home Assistant** — Push notifications via Home Assistant notify service
- 💬 **Google Chat** — Messages to Google Workspace spaces via webhooks
- 🟢 **Matrix** — Messages to Matrix rooms via client-server API
- 💬 **Mattermost** — Incoming webhook messages (Slack-compatible)
- 🚀 **Rocket.Chat** — Incoming webhook messages
- 🚨 **PagerDuty** — Incident triggers via Events API v2
- 📌 **Pushbullet** — Push notifications to all devices
- 🌐 **Webhook** — JSON payloads to any HTTP endpoint

### ⚙️ Other

- 💡 **Update hints** — Copy-paste-ready commands shown after each checker's updates
- 🕐 **Cron scheduling** — Built-in cron job management
- 🧙 **Interactive setup** — Menu-driven wizard with auto-detection
- 💻 **Multi-platform** — Linux (amd64, arm64, armv7), macOS (amd64, arm64)

## 📥 Installation

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the latest release, and installs it to `/usr/local/bin`. On Linux, the script offers an optional **server setup** that creates a dedicated system user with proper permissions (see [Linux Server: Recommended Setup](#linux-server-recommended-setup)).

For non-interactive use (e.g. in provisioning scripts):

```bash
# With server setup (dedicated user, sudoers, cron)
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --server

# Without server setup (binary only)
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --no-server
```

### Manual Install

Download the latest release for your platform from [GitHub Releases](https://github.com/mahype/update-watcher/releases):

```bash
# Download (replace OS and ARCH as needed: linux/darwin, amd64/arm64/armv7)
curl -sSL -o update-watcher.tar.gz \
  https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_linux_amd64.tar.gz

# Extract and install
tar xzf update-watcher.tar.gz
sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
rm update-watcher.tar.gz

# Verify
update-watcher version
```

### Build from Source

Requires Go 1.21+.

```bash
git clone https://github.com/mahype/update-watcher.git
cd update-watcher
make build
sudo make install
```

### Linux Server: Recommended Setup

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

## 🗑️ Uninstallation

### Uninstall Script

The uninstall script automatically detects all installed components (binary, config, cron, log, sudoers, dedicated user) and removes them:

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/uninstall.sh | bash
```

For non-interactive use:

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/uninstall.sh | bash -s -- --yes
```

### Manual Removal

Quick removal (binary and config only):

```bash
update-watcher uninstall-cron
sudo rm /usr/local/bin/update-watcher
sudo rm -rf /etc/update-watcher
```

Full removal including the dedicated server user:

```bash
# 1. Remove cron job
sudo crontab -u update-watcher -r

# 2. Remove binary
sudo rm /usr/local/bin/update-watcher

# 3. Remove config
sudo rm -rf /etc/update-watcher

# 4. Remove log file
sudo rm -f /var/log/update-watcher.log

# 5. Remove sudoers file
sudo rm -f /etc/sudoers.d/update-watcher

# 6. Remove dedicated user
sudo userdel -r update-watcher
```

## 🚀 Quick Start

```bash
# Interactive setup wizard
update-watcher setup

# Test run without sending notifications
update-watcher run --dry-run

# Schedule daily checks (default: 7:00 AM)
update-watcher install-cron
```

## 📋 Commands

| Command | Description |
|---|---|
| `setup` | Interactive setup wizard |
| `run [--dry-run] [--only TYPE]` | Execute all configured checks |
| `status [--format json]` | Show current configuration |
| `validate` | Validate configuration file |
| `watch apt [--security-only] [--no-sudo]` | Add APT watcher |
| `watch dnf [--security-only] [--no-sudo]` | Add DNF watcher |
| `watch pacman [--no-sudo]` | Add Pacman watcher |
| `watch zypper [--security-only] [--no-sudo]` | Add Zypper watcher |
| `watch apk [--no-sudo]` | Add APK watcher |
| `watch macos [--security-only]` | Add macOS software update watcher |
| `watch homebrew [--no-casks]` | Add Homebrew watcher |
| `watch snap` | Add Snap watcher |
| `watch flatpak` | Add Flatpak watcher |
| `watch docker` | Add Docker watcher |
| `watch webproject --path PATH [--name NAME] [--env TYPE]` | Add web project watcher |
| `watch wordpress --path PATH [--name NAME] [--env TYPE]` | Add WordPress watcher |
| `unwatch <type> [--name NAME]` | Remove a watcher |
| `install-cron [--time HH:MM]` | Schedule daily cron job |
| `uninstall-cron` | Remove cron job |
| `version` | Show version info |

### Global Flags

| Flag | Description |
|---|---|
| `--config, -c` | Path to config file |
| `--quiet, -q` | Suppress terminal output |
| `--verbose, -v` | Enable debug logging |

## 📝 WordPress Environments

The WordPress checker auto-detects the development environment and uses the correct command to run WP-CLI. Supported environments:

| Environment | Command | Auto-detection |
|---|---|---|
| **Native** (default) | `wp --path=...` | `wp-config.php` exists |
| **ddev** | `ddev wp` | `.ddev/config.yaml` |
| **Lando** | `lando wp` | `.lando.yml` |
| **wp-env** | `npx wp-env run cli wp` | `.wp-env.json` |
| **Docker Compose** | `docker compose exec ... wp` | `docker-compose.yml` with wordpress image |
| **Bedrock** | `wp --path=web/wp` | `composer.json` with `roots/bedrock` |
| **LocalWP** | `wp` | Path contains `/Local Sites/` |
| **MAMP** | `wp` | `/Applications/MAMP/` exists |
| **XAMPP** | `wp` | `/Applications/XAMPP/` or `/opt/lampp/` |
| **Laragon** | `wp` | `C:\laragon\` exists |
| **Laravel Valet** | `wp` | `~/.config/valet/` exists |

Override auto-detection with `--env`:

```bash
update-watcher watch wordpress --path /path/to/project --name "My Site" --env ddev
```

## 📦 Web Project Environments

The web project checker auto-detects the package managers used in a project and the development environment. It supports multiple projects, each potentially using multiple package managers (e.g., a PHP project with both Composer and npm).

### Supported Package Managers

| Manager | Detection | Outdated Command | Security Audit |
|---|---|---|---|
| **npm** | `package-lock.json` | `npm outdated --json` | `npm audit --json` |
| **yarn** | `yarn.lock` | `yarn outdated --json` (v1) / plugin (v2+) | `yarn audit --json` |
| **pnpm** | `pnpm-lock.yaml` | `pnpm outdated --format json` | `pnpm audit --json` |
| **Composer** | `composer.json` | `composer outdated --format=json --direct` | `composer audit --format=json` |

When multiple Node.js lock files exist, the manager is chosen by priority: pnpm > yarn > npm. Non-Node managers (Composer) are always included alongside a Node.js manager.

### Supported Environments

| Environment | Command Prefix | Auto-detection |
|---|---|---|
| **Native** (default) | direct execution | No container markers found |
| **ddev** | `ddev exec <tool>` | `.ddev/config.yaml` |
| **Lando** | `lando ssh -c "<tool>"` | `.lando.yml` |
| **Docker Compose** | `docker compose exec app <tool>` | `docker-compose.yml` / `compose.yml` |

### Usage

```bash
# Auto-detect everything
update-watcher watch webproject --path /var/www/myapp --name "My App"

# Explicit package managers
update-watcher watch webproject --path /var/www/myapp --name "My App" --managers composer,npm

# Skip security audits
update-watcher watch webproject --path /var/www/myapp --name "My App" --no-audit

# Specify environment
update-watcher watch webproject --path /var/www/myapp --name "My App" --env ddev
```

## 🔧 Configuration

Config file location:
- **Linux:** `/etc/update-watcher/config.yaml` (system-wide), `~/.config/update-watcher/config.yaml` (user)
- **macOS:** `~/.config/update-watcher/config.yaml`

Environment variables with the `UPDATE_WATCHER_` prefix override config values.

### Example

```yaml
hostname: "web-prod-01"  # Auto-detected if empty

watchers:
  - type: apt
    enabled: true
    options:
      use_sudo: true
      security_only: false

  - type: macos
    enabled: true
    options:
      security_only: false

  - type: homebrew
    enabled: true
    options:
      include_casks: true

  - type: snap
    enabled: true

  - type: flatpak
    enabled: true

  - type: docker
    enabled: true
    options:
      containers: "all"
      exclude: ["watchtower", "traefik"]

  - type: wordpress
    enabled: true
    options:
      sites:
        - name: "Main Blog"
          path: "/var/www/html/blog"
          run_as: "www-data"
          environment: "native"
        - name: "Dev Site"
          path: "/home/user/projects/my-site"
          environment: "ddev"
      check_core: true
      check_plugins: true
      check_themes: true

  - type: webproject
    enabled: true
    options:
      check_audit: true
      projects:
        - name: "Laravel App"
          path: "/var/www/myapp"
          environment: "ddev"
          managers:
            - composer
            - npm
        - name: "React Frontend"
          path: "/var/www/frontend"
          # auto-detect managers and environment

notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "https://hooks.slack.com/services/..."
      mention_on_security: "@channel"
      use_emoji: true

  - type: discord
    enabled: false
    options:
      webhook_url: "https://discord.com/api/webhooks/..."
      username: "Update Watcher"
      mention_role: "123456789"

  - type: teams
    enabled: false
    options:
      webhook_url: "https://prod.workflows.microsoft.com/..."

  - type: telegram
    enabled: false
    options:
      bot_token: "123456:ABC-..."
      chat_id: "-100123456789"

  - type: email
    enabled: false
    options:
      smtp_host: "smtp.example.com"
      smtp_port: 587
      username: "alerts@example.com"
      password: "secret"
      from: "alerts@example.com"
      to: ["admin@example.com"]
      tls: true

  - type: ntfy
    enabled: false
    options:
      topic: "update-watcher"
      server_url: "https://ntfy.sh"

  - type: pushover
    enabled: false
    options:
      app_token: "azGDORePK8gMaC0QOYAMyEEuzJnyUi"
      user_key: "uQiRzpo4DXghDmr9QzzfQu27cmVRsG"

  - type: gotify
    enabled: false
    options:
      server_url: "https://gotify.example.com"
      token: "AKsjdf83jsd"

  - type: homeassistant
    enabled: false
    options:
      url: "http://homeassistant.local:8123"
      token: "eyJ0eXAiOiJKV1Qi..."
      service: "notify"

  - type: googlechat
    enabled: false
    options:
      webhook_url: "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=xxx&token=yyy"

  - type: matrix
    enabled: false
    options:
      homeserver: "https://matrix.org"
      access_token: "syt_bot_token_here"
      room_id: "!abc123:matrix.org"

  - type: mattermost
    enabled: false
    options:
      webhook_url: "https://mattermost.example.com/hooks/xxx"
      username: "Update Watcher"

  - type: rocketchat
    enabled: false
    options:
      webhook_url: "https://rocket.example.com/hooks/xxx"
      username: "Update Watcher"

  - type: pagerduty
    enabled: false
    options:
      routing_key: "R0123456789ABCDEF"
      severity: "warning"

  - type: pushbullet
    enabled: false
    options:
      access_token: "o.ABCDEF123456"

  - type: webhook
    enabled: false
    options:
      url: "https://api.example.com/updates"
      method: "POST"
      auth_header: "Bearer token123"

settings:
  send_policy: "only-on-updates"  # or "always"
  log_file: "/var/log/update-watcher.log"
  schedule: "0 7 * * *"
```

### Checker Options

| Checker | Option | Default | Description |
|---|---|---|---|
| `apt` | `use_sudo` | `true` | Use sudo for apt-get update |
| `apt` | `security_only` | `false` | Only report security updates |
| `dnf` | `use_sudo` | `true` | Use sudo for dnf operations |
| `dnf` | `security_only` | `false` | Only report security updates |
| `pacman` | `use_sudo` | `true` | Use sudo for pacman -Sy |
| `zypper` | `use_sudo` | `true` | Use sudo for zypper operations |
| `zypper` | `security_only` | `false` | Only report security updates |
| `apk` | `use_sudo` | `false` | Use sudo for apk operations |
| `macos` | `security_only` | `false` | Only report security updates |
| `homebrew` | `include_casks` | `true` | Also check cask updates |
| `docker` | `containers` | `"all"` | `"all"` or comma-separated names |
| `docker` | `exclude` | `[]` | Container names to skip |
| `wordpress` | `sites` | `[]` | List of site objects (name, path, run_as, environment) |
| `wordpress` | `check_core` | `true` | Check WordPress core updates |
| `wordpress` | `check_plugins` | `true` | Check plugin updates |
| `wordpress` | `check_themes` | `true` | Check theme updates |
| `webproject` | `projects` | `[]` | List of project objects (name, path, environment, managers, run_as) |
| `webproject` | `check_audit` | `true` | Run security audits (npm audit, composer audit, etc.) |

### Notifier Options

| Notifier | Option | Required | Description |
|---|---|---|---|
| `slack` | `webhook_url` | Yes | Slack incoming webhook URL |
| `slack` | `mention_on_security` | No | User/group to mention on security updates |
| `slack` | `use_emoji` | No | Enable emoji in messages |
| `discord` | `webhook_url` | Yes | Discord webhook URL |
| `discord` | `username` | No | Bot display name |
| `discord` | `avatar_url` | No | Bot avatar URL |
| `discord` | `mention_role` | No | Role ID to mention on security updates |
| `teams` | `webhook_url` | Yes | Teams Workflow webhook URL |
| `telegram` | `bot_token` | Yes | Telegram bot token |
| `telegram` | `chat_id` | Yes | Telegram chat/group ID |
| `telegram` | `disable_notification` | No | Send silently |
| `email` | `smtp_host` | Yes | SMTP server hostname |
| `email` | `smtp_port` | No | SMTP port (default: 587) |
| `email` | `username` | Yes | SMTP username |
| `email` | `password` | Yes | SMTP password |
| `email` | `from` | Yes | Sender address |
| `email` | `to` | Yes | Recipient addresses (list) |
| `email` | `tls` | No | Enable STARTTLS |
| `ntfy` | `topic` | Yes | ntfy topic name |
| `ntfy` | `server_url` | No | ntfy server (default: https://ntfy.sh) |
| `ntfy` | `token` | No | Authentication token |
| `ntfy` | `priority` | No | Message priority |
| `pushover` | `app_token` | Yes | Pushover application token |
| `pushover` | `user_key` | Yes | Pushover user or group key |
| `pushover` | `device` | No | Send to specific device only |
| `pushover` | `priority` | No | Priority (-2 to 2, default: 0) |
| `pushover` | `sound` | No | Notification sound |
| `gotify` | `server_url` | Yes | Gotify server URL |
| `gotify` | `token` | Yes | Gotify application token |
| `gotify` | `priority` | No | Priority (0-10, default: 5) |
| `homeassistant` | `url` | Yes | Home Assistant base URL (e.g. http://homeassistant.local:8123) |
| `homeassistant` | `token` | Yes | Long-Lived Access Token from HA profile |
| `homeassistant` | `service` | No | Notify service name (default: notify) |
| `googlechat` | `webhook_url` | Yes | Google Chat webhook URL |
| `googlechat` | `thread_key` | No | Group messages in a thread |
| `matrix` | `homeserver` | Yes | Matrix homeserver URL (e.g. https://matrix.org) |
| `matrix` | `access_token` | Yes | Bot access token |
| `matrix` | `room_id` | Yes | Room ID (e.g. !abc123:matrix.org) |
| `mattermost` | `webhook_url` | Yes | Mattermost incoming webhook URL |
| `mattermost` | `channel` | No | Override channel |
| `mattermost` | `username` | No | Bot display name (default: Update Watcher) |
| `mattermost` | `icon_url` | No | Bot avatar URL |
| `rocketchat` | `webhook_url` | Yes | Rocket.Chat incoming webhook URL |
| `rocketchat` | `channel` | No | Override channel |
| `rocketchat` | `username` | No | Bot display name (default: Update Watcher) |
| `pagerduty` | `routing_key` | Yes | Events API v2 integration key |
| `pagerduty` | `severity` | No | Default severity (info/warning/error/critical, default: warning) |
| `pushbullet` | `access_token` | Yes | Pushbullet access token |
| `pushbullet` | `device_iden` | No | Send to specific device only |
| `pushbullet` | `channel_tag` | No | Send to a Pushbullet channel |
| `webhook` | `url` | Yes | Target URL |
| `webhook` | `method` | No | HTTP method (default: POST) |
| `webhook` | `content_type` | No | Content-Type header |
| `webhook` | `auth_header` | No | Authorization header value |
| `webhook` | `headers` | No | Additional HTTP headers |

## 🔐 Security

### Config file permissions

The config file is written with mode `0600` (owner-readable only) because it may contain sensitive credentials like API tokens, webhook URLs, and passwords. After manual edits, verify permissions:

```bash
ls -la ~/.config/update-watcher/config.yaml
# Should show: -rw------- (600)
```

### Environment variable references

Instead of storing secrets in plain text, you can use `${ENV_VAR}` references in your `config.yaml`. This is recommended for CI/CD, Docker, and shared environments.

```yaml
notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"

  - type: email
    enabled: true
    options:
      smtp_host: "smtp.example.com"
      password: "${SMTP_PASSWORD}"
```

Supported syntax:

| Pattern | Behavior |
|---|---|
| `${VAR}` | Replaced with env var value, empty if unset |
| `${VAR:-default}` | Replaced with env var value, `default` if unset |

### Using a `.env` file

You can set environment variables in a `.env` file (already in `.gitignore`):

```bash
# .env
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T00/B00/xxx
TELEGRAM_BOT_TOKEN=123456:ABC-DEF
SMTP_PASSWORD=my-app-password
PUSHOVER_APP_TOKEN=azGDORePK8gMaC0QOYAMyEEuzJnyUi
PUSHOVER_USER_KEY=uQiRzpo4DXghDmr9QzzfQu27cmVRsG
```

Then load it before running:

```bash
export $(grep -v '^#' .env | xargs) && update-watcher run
```

A `config.example.yaml` template with all available options and `${ENV_VAR}` placeholders is included in the repository.

## 🧙 Setup Wizard

The `setup` command launches a menu-driven wizard that shows the current configuration and lets you add/remove watchers, configure notifiers, and manage settings.

The wizard auto-detects available tools -- it only shows package manager options for tools that are actually installed on the system (e.g., APT on Debian, DNF on Fedora). WordPress is always available since environment detection handles tool requirements.

```
=== update-watcher setup ===

Current configuration:
  Hostname:   web-prod-01
  Watchers:   APT, Docker, WordPress
  Notifiers:  slack configured
  Cron:       daily at 07:00

What would you like to do?
  > Manage Watchers (APT, Docker, WordPress)
  > Notifications (slack configured)
  > Settings (hostname: web-prod-01)
  > Cron Job (daily at 07:00)
  > Run Test Check (dry-run)
  > Save & Exit
```

## 🚦 Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success, no updates found |
| 1 | Success, updates found |
| 2 | Partial failure (some checkers failed) |
| 3 | Complete failure |
| 4 | Configuration missing or invalid |

## 📄 License

MIT
