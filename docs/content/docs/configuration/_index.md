---
title: "Configuration - Update-Watcher YAML Config Guide"
description: "Configure Update-Watcher via YAML config, CLI flags, or environment variables. File locations, structure overview, and validation."
weight: 20
---

Update-Watcher is configured through a YAML file that defines which package managers to check (watchers), where to send notifications (notifiers), and global settings. This section covers every aspect of configuration.

## Configuration Pages

{{< cards >}}
  {{< card link="config-file" title="Config File Reference" subtitle="Complete YAML reference with annotated examples for all watchers, notifiers, and settings." icon="document-text" >}}
  {{< card link="environment-variables" title="Environment Variables" subtitle="Use environment variables for secrets, .env files, and the UPDATE_WATCHER_ prefix." icon="key" >}}
  {{< card link="security" title="Security Best Practices" subtitle="File permissions, secret management, dedicated users, and network security." icon="shield-check" >}}
{{< /cards >}}

## Config File Locations

Update-Watcher looks for its configuration file in the following locations, in order of priority:

| Priority | Platform | Path | Typical Use |
|----------|----------|------|-------------|
| 1 | Any | Path passed via `--config` flag | Explicit override |
| 2 | Linux | `/etc/update-watcher/config.yaml` | System-wide (server setup) |
| 3 | Linux, macOS | `~/.config/update-watcher/config.yaml` | Per-user |

On Linux, the system-wide path (`/etc/update-watcher/`) is created by the server setup during installation. If both files exist, the system-wide config takes precedence. On macOS, only the user config path is used.

To use a custom config file path:

```bash {filename="Terminal"}
update-watcher run --config /path/to/my-config.yaml
```

## Configuration Structure Overview

A complete configuration file has three top-level sections:

```yaml {filename="config.yaml"}
# Optional: server hostname (auto-detected if empty)
hostname: ""

# What to check
watchers:
  - type: apt
    enabled: true
    options:
      use_sudo: true

# Where to send notifications
notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"

# Global settings
settings:
  send_policy: "only-on-updates"
  log_file: "/var/log/update-watcher.log"
  schedule: "0 7 * * *"
```

- **`hostname`** -- Identifies the server in notifications. Leave empty for auto-detection.
- **`watchers`** -- A list of package manager checkers. Each entry specifies a `type`, an `enabled` flag, and type-specific `options`.
- **`notifiers`** -- A list of notification channels. Same structure as watchers: `type`, `enabled`, and `options`.
- **`settings`** -- Global behavior: notification policy, log file path, and cron schedule.

For the full annotated reference with all options, see [Config File Reference](config-file/).

## Configuration Precedence

{{< callout type="info" >}}
When the same setting can be specified in multiple places, the following precedence applies (highest to lowest):

1. **CLI flags** -- Flags like `--quiet`, `--verbose`, `--config`, and `--only` always take the highest priority.
2. **Environment variables** -- Variables with the `UPDATE_WATCHER_` prefix override config file values via Viper. Additionally, `${VAR}` substitution is applied to all string values in the YAML file.
3. **Config file** -- The YAML configuration file provides the base values.

For example, if your config file sets `send_policy: "only-on-updates"` but you set `UPDATE_WATCHER_SETTINGS_SEND_POLICY=always` in the environment, the environment variable wins.
{{< /callout >}}

See [Environment Variables](environment-variables/) for details on the substitution syntax and the `UPDATE_WATCHER_` prefix.

## Validating Your Configuration

Run the built-in validation command to check for syntax errors, missing required fields, and invalid option values:

```bash {filename="Terminal"}
update-watcher validate
```

If the config is valid:

```
Configuration is valid.
```

If there are issues, the output describes each problem and where it occurs.

You can also view the resolved configuration (after environment variable substitution) with:

```bash {filename="Terminal"}
update-watcher status
```

For machine-readable output:

```bash {filename="Terminal"}
update-watcher status --format json
```

## Creating Your First Config

The easiest way to create a config file is through the interactive setup wizard:

```bash {filename="Terminal"}
update-watcher setup
```

The wizard auto-detects installed package managers, walks you through enabling watchers and notifiers, and writes the file with correct permissions. See [First Run](../getting-started/first-run/) for a full walkthrough.

Alternatively, you can write the YAML by hand. Start with the minimal example in the [Config File Reference](config-file/) and add watchers and notifiers as needed.

## Next Steps

- [Config File Reference](config-file/) -- Full annotated YAML with all options.
- [Environment Variables](environment-variables/) -- Secrets management and `.env` files.
- [Security Best Practices](security/) -- Permissions, secrets, and hardening.
