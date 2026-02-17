# Update-Watcher

A modular CLI tool that checks for available software updates and sends notifications. Designed to run on servers, scheduled via cron. Single binary, no runtime dependencies.

## Features

### Checkers

- **APT** — Debian/Ubuntu package updates (with security-only filter)
- **DNF** — Fedora/RHEL/Rocky/AlmaLinux package updates (with security classification)
- **Pacman** — Arch/Manjaro package updates
- **Zypper** — openSUSE/SLES package updates (with security patches)
- **APK** — Alpine Linux package updates
- **Docker** — Detects newer images for running containers (read-only, no image pulls)
- **WordPress** — Core, plugin, and theme updates across 11 environments

### Notifiers

- **Slack** — Rich Block Kit messages with security highlighting
- **Discord** — Embedded messages via webhooks
- **Microsoft Teams** — Adaptive Card messages via Workflow webhooks
- **Telegram** — Bot API messages with Markdown formatting
- **Email** — HTML emails via SMTP (with STARTTLS)
- **ntfy** — Push notifications via [ntfy.sh](https://ntfy.sh) or self-hosted
- **Webhook** — JSON payloads to any HTTP endpoint

### Other

- **Cron scheduling** — Built-in cron job management
- **Interactive setup** — Menu-driven wizard with auto-detection
- **Multi-platform** — Linux (amd64, arm64, armv7), macOS (amd64, arm64)

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the latest release, and installs it to `/usr/local/bin`.

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
| `run [--dry-run] [--only TYPE]` | Execute all configured checks |
| `status [--format json]` | Show current configuration |
| `validate` | Validate configuration file |
| `watch apt [--security-only] [--no-sudo]` | Add APT watcher |
| `watch dnf [--security-only] [--no-sudo]` | Add DNF watcher |
| `watch pacman [--no-sudo]` | Add Pacman watcher |
| `watch zypper [--security-only] [--no-sudo]` | Add Zypper watcher |
| `watch apk [--no-sudo]` | Add APK watcher |
| `watch docker` | Add Docker watcher |
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

## WordPress Environments

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

## Configuration

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
| `docker` | `containers` | `"all"` | `"all"` or comma-separated names |
| `docker` | `exclude` | `[]` | Container names to skip |
| `wordpress` | `sites` | `[]` | List of site objects (name, path, run_as, environment) |
| `wordpress` | `check_core` | `true` | Check WordPress core updates |
| `wordpress` | `check_plugins` | `true` | Check plugin updates |
| `wordpress` | `check_themes` | `true` | Check theme updates |

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
| `webhook` | `url` | Yes | Target URL |
| `webhook` | `method` | No | HTTP method (default: POST) |
| `webhook` | `content_type` | No | Content-Type header |
| `webhook` | `auth_header` | No | Authorization header value |
| `webhook` | `headers` | No | Additional HTTP headers |

## Setup Wizard

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

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success, no updates found |
| 1 | Success, updates found |
| 2 | Partial failure (some checkers failed) |
| 3 | Complete failure |
| 4 | Configuration missing or invalid |

## Build from Source

```bash
git clone https://github.com/mahype/update-watcher.git
cd update-watcher
make build
make install
```

Requires Go 1.21+.

## License

MIT
