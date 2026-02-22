---
title: "Environment Variables - Using Secrets in Update-Watcher Config"
description: "Use environment variables to manage secrets in Update-Watcher configuration. Supports ${VAR} and ${VAR:-default} syntax, .env files, and the UPDATE_WATCHER_ prefix."
weight: 2
---

Update-Watcher supports environment variable substitution in its YAML configuration file. This allows you to keep secrets like API tokens, webhook URLs, and passwords out of the config file itself, which is a best practice for security, CI/CD pipelines, and shared environments.

## Substitution Syntax

{{< callout type="info" >}}
All string values in `config.yaml` support the following substitution patterns:

| Pattern | Behavior | Example |
|---------|----------|---------|
| `${VAR}` | Replaced with the value of the environment variable `VAR`. If the variable is not set, the value is replaced with an empty string. | `${SLACK_WEBHOOK_URL}` |
| `${VAR:-default}` | Replaced with the value of `VAR` if set. If not set, uses the literal string `default`. | `${LOG_LEVEL:-info}` |

Substitution is applied to every string value in the config file, including values nested inside `options` objects. Non-string values (booleans, numbers, lists) are not affected.
{{< /callout >}}

### Examples in Config

```yaml {filename="config.yaml"}
notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"

  - type: telegram
    enabled: true
    options:
      bot_token: "${TELEGRAM_BOT_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"

  - type: email
    enabled: true
    options:
      smtp_host: "${SMTP_HOST:-smtp.gmail.com}"
      smtp_port: 587
      username: "${SMTP_USERNAME}"
      password: "${SMTP_PASSWORD}"
      from: "${SMTP_FROM:-alerts@example.com}"
      to: ["admin@example.com"]
```

## The UPDATE_WATCHER_ Prefix

Update-Watcher uses [Viper](https://github.com/spf13/viper) for configuration management. Viper automatically maps environment variables with the `UPDATE_WATCHER_` prefix to config file keys.

The mapping follows this pattern:

- Nested keys are separated by underscores.
- All characters are uppercase.

| Environment Variable | Config Equivalent |
|---------------------|-------------------|
| `UPDATE_WATCHER_HOSTNAME` | `hostname` |
| `UPDATE_WATCHER_SETTINGS_SEND_POLICY` | `settings.send_policy` |
| `UPDATE_WATCHER_SETTINGS_LOG_FILE` | `settings.log_file` |

These prefixed environment variables **override** values from the config file, following the standard [configuration precedence](../#configuration-precedence):

1. CLI flags (highest priority)
2. `UPDATE_WATCHER_` environment variables
3. Config file values (lowest priority)

Note that the `UPDATE_WATCHER_` prefix mechanism is separate from the `${VAR}` substitution syntax. The prefix approach overrides config keys directly through Viper, while `${VAR}` performs string replacement within the YAML file itself. Both can be used together.

## Using a .env File

For local development or servers where you want to keep secrets in a separate file, you can use a `.env` file. This file contains one environment variable per line in `KEY=VALUE` format.

Create a `.env` file (for example, at `/etc/update-watcher/.env` on a server or next to the binary during development):

```bash {filename=".env"}
# /etc/update-watcher/.env
# Notification service credentials
SLACK_WEBHOOK_URL=<your-slack-webhook-url>
TELEGRAM_BOT_TOKEN=<your-telegram-bot-token>
TELEGRAM_CHAT_ID=<your-telegram-chat-id>
SMTP_PASSWORD=<your-smtp-password>

# Pushover
PUSHOVER_APP_TOKEN=<your-pushover-app-token>
PUSHOVER_USER_KEY=<your-pushover-user-key>
```

Load the `.env` file before running Update-Watcher:

```bash {filename="Terminal"}
export $(grep -v '^#' /etc/update-watcher/.env | xargs) && update-watcher run
```

### .env File in Cron

When running via cron, environment variables from your shell profile are not automatically available. To load a `.env` file in a cron job, include the export command in the cron entry:

```
0 7 * * * export $(grep -v '^\#' /etc/update-watcher/.env | xargs) && /usr/local/bin/update-watcher run --quiet
```

Alternatively, define the variables directly in the crontab:

```
SLACK_WEBHOOK_URL=<your-slack-webhook-url>
0 7 * * * /usr/local/bin/update-watcher run --quiet
```

### .env File Permissions

{{< callout type="warning" >}}
The `.env` file contains secrets and must be protected. Never commit `.env` files to version control. Always add `.env` to your `.gitignore`.
{{< /callout >}}

Set strict permissions:

```bash {filename="Terminal"}
chmod 600 /etc/update-watcher/.env
chown update-watcher:update-watcher /etc/update-watcher/.env
```

Make sure the `.env` file is listed in `.gitignore` if it exists inside a repository.

## Common Environment Variables

The following table lists environment variables commonly referenced in Update-Watcher configurations. These are not built-in to Update-Watcher; they are standard names used in the `${VAR}` substitution pattern in your config file.

| Variable | Used By | Description |
|----------|---------|-------------|
| `SLACK_WEBHOOK_URL` | Slack notifier | Slack incoming webhook URL |
| `DISCORD_WEBHOOK_URL` | Discord notifier | Discord webhook URL |
| `TEAMS_WEBHOOK_URL` | Teams notifier | Microsoft Teams Workflow webhook URL |
| `TELEGRAM_BOT_TOKEN` | Telegram notifier | Telegram Bot API token |
| `TELEGRAM_CHAT_ID` | Telegram notifier | Telegram chat or group ID |
| `SMTP_HOST` | Email notifier | SMTP server hostname |
| `SMTP_USERNAME` | Email notifier | SMTP username |
| `SMTP_PASSWORD` | Email notifier | SMTP password or app-specific password |
| `NTFY_TOKEN` | ntfy notifier | ntfy authentication token |
| `PUSHOVER_APP_TOKEN` | Pushover notifier | Pushover application token |
| `PUSHOVER_USER_KEY` | Pushover notifier | Pushover user or group key |
| `GOTIFY_TOKEN` | Gotify notifier | Gotify application token |
| `GOTIFY_URL` | Gotify notifier | Gotify server URL |
| `HA_TOKEN` | Home Assistant notifier | Home Assistant long-lived access token |
| `HA_URL` | Home Assistant notifier | Home Assistant base URL |
| `GOOGLECHAT_WEBHOOK_URL` | Google Chat notifier | Google Chat webhook URL |
| `MATRIX_ACCESS_TOKEN` | Matrix notifier | Matrix bot access token |
| `PAGERDUTY_ROUTING_KEY` | PagerDuty notifier | PagerDuty Events API v2 integration key |
| `PUSHBULLET_TOKEN` | Pushbullet notifier | Pushbullet access token |
| `WEBHOOK_AUTH` | Webhook notifier | Authorization header for generic webhooks |

## Best Practices

{{< callout type="info" >}}
1. **Never commit secrets to version control.** Use `${VAR}` references in `config.yaml` and keep actual values in environment variables or a `.env` file that is excluded from Git.

2. **Set strict file permissions.** Both `config.yaml` and `.env` files should be `0600` (owner-readable only). See [Security Best Practices](../security/) for details.

3. **Use `${VAR:-default}` for non-sensitive defaults.** This makes the config self-documenting and ensures sensible fallback values for settings like `smtp_host` or `smtp_port`.

4. **Prefer `.env` files over inline cron variables** for servers with many secrets. A single `.env` file is easier to manage and audit than scattered variable definitions.

5. **Validate after changes.** After modifying environment variables, run `update-watcher validate` to confirm the resolved config is correct.
{{< /callout >}}

## Next Steps

- [Security Best Practices](../security/) -- File permissions, dedicated users, and network security.
- [Config File Reference](../config-file/) -- Full YAML configuration reference.
- [First Run](../../getting-started/first-run/) -- Setting up your first configuration.
