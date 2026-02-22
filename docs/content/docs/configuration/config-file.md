---
title: "Config File Reference - Complete YAML Configuration for Update-Watcher"
description: "Complete YAML configuration reference for Update-Watcher. All watcher options, notifier settings, global settings with annotated examples."
weight: 1
---

This page documents the complete YAML configuration file format for Update-Watcher, including all top-level fields, watcher options, notifier options, and global settings.

## Full Annotated Example

The following example shows a typical configuration with one watcher and one notifier. Every field is commented to explain its purpose.

```yaml {filename="config.yaml"}
# Server hostname included in notifications.
# Leave empty to auto-detect from the system hostname.
hostname: ""

# Watchers define which package managers to check for updates.
# You can have multiple watchers of the same type (e.g. multiple WordPress sites).
watchers:
  - type: apt
    enabled: true
    options:
      # Run 'sudo apt-get update' before checking. Set to false if your
      # system already refreshes package lists automatically.
      use_sudo: true
      # Only report security updates, ignoring regular package updates.
      security_only: false
      # Hide phased updates. Ubuntu gradually rolls out some updates;
      # enabling this hides updates not yet available to your machine.
      hide_phased: true

# Notifiers define where update notifications are sent.
# You can configure multiple notifiers; all enabled ones receive every notification.
notifiers:
  - type: slack
    enabled: true
    options:
      # Slack incoming webhook URL. Use an environment variable reference
      # to keep the secret out of the config file.
      webhook_url: "${SLACK_WEBHOOK_URL}"
      # Mention this user or group when security updates are found.
      mention_on_security: "@channel"
      # Include emoji in Slack messages for visual distinction.
      use_emoji: true

# Global settings that apply to the entire application.
settings:
  # "only-on-updates" sends notifications only when updates are found.
  # "always" sends a notification after every run, even if no updates exist.
  send_policy: "only-on-updates"
  # Path to the log file. Leave empty to disable file logging.
  log_file: "/var/log/update-watcher.log"
  # Cron schedule expression used by 'install-cron'. Does not affect 'run'.
  schedule: "0 7 * * *"
```

## Top-Level Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `hostname` | string | No | Auto-detected | Server name shown in notifications |
| `watchers` | list | Yes | `[]` | List of watcher configurations |
| `notifiers` | list | Yes | `[]` | List of notifier configurations |
| `settings` | object | No | See below | Global application settings |

## Watcher Configuration

Each entry in the `watchers` list has the following structure:

```yaml {filename="config.yaml"}
watchers:
  - type: <watcher-type>    # Required: checker identifier
    enabled: true            # Optional: default true
    options:                 # Type-specific options
      key: value
```

### Watcher Types and Options

| Watcher | Option | Type | Default | Description |
|---------|--------|------|---------|-------------|
| `apt` | `use_sudo` | bool | `true` | Run `sudo apt-get update` before checking |
| `apt` | `security_only` | bool | `false` | Only report security updates |
| `apt` | `hide_phased` | bool | `true` | Hide Ubuntu phased updates |
| `dnf` | `use_sudo` | bool | `true` | Use sudo for DNF operations |
| `dnf` | `security_only` | bool | `false` | Only report security updates |
| `pacman` | `use_sudo` | bool | `true` | Use sudo for `pacman -Sy` |
| `zypper` | `use_sudo` | bool | `true` | Use sudo for Zypper operations |
| `zypper` | `security_only` | bool | `false` | Only report security updates |
| `apk` | `use_sudo` | bool | `false` | Use sudo for APK operations |
| `macos` | `security_only` | bool | `false` | Only report security updates |
| `homebrew` | `include_casks` | bool | `true` | Also check cask updates |
| `docker` | `containers` | string | `"all"` | `"all"` or comma-separated container names |
| `docker` | `exclude` | list | `[]` | Container names to skip |
| `distro` | `lts_only` | bool | `true` | Only report LTS release upgrades (Ubuntu) |
| `openclaw` | `channel` | string | `""` | Update channel: stable, beta, or dev |

### WordPress Watcher

The WordPress watcher monitors one or more WordPress installations for core, plugin, and theme updates.

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    enabled: true
    options:
      sites:
        - name: "Production Blog"
          path: "/var/www/html/blog"
          run_as: "www-data"
          environment: "native"
        - name: "Dev Site"
          path: "/home/user/projects/my-site"
          environment: "ddev"
      check_core: true
      check_plugins: true
      check_themes: true
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `sites` | list | `[]` | List of site objects |
| `sites[].name` | string | -- | Display name for the site |
| `sites[].path` | string | -- | Filesystem path to WordPress root |
| `sites[].run_as` | string | -- | User to run WP-CLI as (optional) |
| `sites[].environment` | string | auto | Environment: native, ddev, lando, wp-env, docker, bedrock, etc. |
| `check_core` | bool | `true` | Check for WordPress core updates |
| `check_plugins` | bool | `true` | Check for plugin updates |
| `check_themes` | bool | `true` | Check for theme updates |

### Web Project Watcher

The web project watcher checks for outdated packages and security vulnerabilities across npm, yarn, pnpm, and Composer.

```yaml {filename="config.yaml"}
watchers:
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
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `projects` | list | `[]` | List of project objects |
| `projects[].name` | string | -- | Display name for the project |
| `projects[].path` | string | -- | Filesystem path to project root |
| `projects[].environment` | string | auto | Environment: native, ddev, lando, docker |
| `projects[].managers` | list | auto | Package managers: npm, yarn, pnpm, composer |
| `projects[].run_as` | string | -- | User to run commands as (optional) |
| `check_audit` | bool | `true` | Run security audits |

## Notifier Configuration

Each entry in the `notifiers` list follows the same structure as watchers:

```yaml {filename="config.yaml"}
notifiers:
  - type: <notifier-type>
    enabled: true
    options:
      key: value
```

Multiple notifiers can be enabled simultaneously. All enabled notifiers receive every notification.

### Notifier Types and Options

For a compact reference of all notifier types and their options, see below. Required options are marked with **(R)**.

**Slack**

```yaml {filename="config.yaml"}
- type: slack
  options:
    webhook_url: "https://hooks.slack.com/..."   # (R)
    mention_on_security: "@channel"
    use_emoji: true
```

**Discord**

```yaml {filename="config.yaml"}
- type: discord
  options:
    webhook_url: "https://discord.com/api/..."   # (R)
    username: "Update Watcher"
    avatar_url: ""
    mention_role: "123456789"
```

**Microsoft Teams**

```yaml {filename="config.yaml"}
- type: teams
  options:
    webhook_url: "https://prod.workflows.microsoft.com/..."  # (R)
```

**Telegram**

```yaml {filename="config.yaml"}
- type: telegram
  options:
    bot_token: "123456:ABC-..."      # (R)
    chat_id: "-100123456789"         # (R)
    disable_notification: false
```

**Email (SMTP)**

```yaml {filename="config.yaml"}
- type: email
  options:
    smtp_host: "smtp.example.com"    # (R)
    smtp_port: 587
    username: "alerts@example.com"   # (R)
    password: "${SMTP_PASSWORD}"     # (R)
    from: "alerts@example.com"       # (R)
    to: ["admin@example.com"]        # (R)
    tls: true
```

**ntfy**

```yaml {filename="config.yaml"}
- type: ntfy
  options:
    topic: "update-watcher"          # (R)
    server_url: "https://ntfy.sh"
    token: ""
    priority: ""
```

**Pushover, Gotify, Home Assistant, Google Chat, Matrix, Mattermost, Rocket.Chat, PagerDuty, Pushbullet, Webhook** -- see the [Notifiers](../../notifiers/) section for full per-notifier documentation.

## Settings Section

The `settings` object controls global application behavior.

```yaml {filename="config.yaml"}
settings:
  send_policy: "only-on-updates"
  log_file: "/var/log/update-watcher.log"
  schedule: "0 7 * * *"
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `send_policy` | string | `"only-on-updates"` | `"only-on-updates"` skips notification when no updates are found. `"always"` sends a notification after every run. |
| `log_file` | string | `""` | Path to a log file. Leave empty to disable file logging. |
| `schedule` | string | `"0 7 * * *"` | Cron expression used by `install-cron`. Does not affect `run` directly. |

## Multiple Watchers of the Same Type

You can define multiple entries of the same watcher type. This is common for WordPress and web project watchers, but also works for any type. Each entry is independent:

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    enabled: true
    options:
      sites:
        - name: "Blog"
          path: "/var/www/blog"

  - type: webproject
    enabled: true
    options:
      projects:
        - name: "API"
          path: "/var/www/api"

  - type: webproject
    enabled: true
    options:
      projects:
        - name: "Dashboard"
          path: "/var/www/dashboard"
```

## Environment Variable Substitution

{{< callout type="info" >}}
All string values in the config file support `${VAR}` and `${VAR:-default}` substitution. This is the recommended way to handle secrets like webhook URLs, API tokens, and passwords. See [Environment Variables](../environment-variables/) for the full reference.
{{< /callout >}}

## Validating the Config

After editing, always validate:

```bash {filename="Terminal"}
update-watcher validate
```

To inspect the resolved configuration with all environment variables expanded:

```bash {filename="Terminal"}
update-watcher status
```

{{< callout emoji="💡" >}}
Run `update-watcher validate` after every config change to catch syntax errors and missing required fields before your next scheduled run.
{{< /callout >}}

## Next Steps

- [Environment Variables](../environment-variables/) -- Substitution syntax and `.env` file usage.
- [Security Best Practices](../security/) -- Permissions and secret management.
- [Checkers](../../checkers/) -- Detailed per-checker documentation.
- [Notifiers](../../notifiers/) -- Detailed per-notifier documentation.
