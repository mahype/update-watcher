---
title: "First Run - Your First Update Check with Update-Watcher"
description: "Walk through your first update check with Update-Watcher. Use the interactive wizard or manual YAML configuration, run your first check, and send a test notification."
weight: 3
---

After [installing](../installation/) Update-Watcher, this guide walks you through configuring it, running your first update check, and sending your first notification.

{{< tabs items="Setup Wizard,Manual YAML" >}}

{{< tab >}}

## Interactive Setup Wizard

The fastest way to configure Update-Watcher is through the built-in setup wizard. It auto-detects installed package managers, guides you through enabling watchers and notifiers, and writes the config file for you.

```bash {filename="Terminal"}
update-watcher setup
```

The wizard displays a menu-driven interface:

```text {filename="Setup Wizard"}
=== update-watcher setup ===

Current configuration:
  Hostname:   web-prod-01
  Watchers:   (none configured)
  Notifiers:  (none configured)
  Cron:       not configured

What would you like to do?
  > Manage Watchers
  > Notifications
  > Settings (hostname: web-prod-01)
  > Cron Job (not configured)
  > Run Test Check (dry-run)
  > Save & Exit
```

{{% steps %}}

### Manage Watchers

The wizard lists only the package managers detected on your system. For example, on an Ubuntu server you might see APT, Snap, and Docker. Select the ones you want to monitor.

### Notifications

Choose a notification channel and enter the required credentials. For Slack, you need a webhook URL. For Email, you need SMTP server details. For Telegram, you need a bot token and chat ID.

### Settings

Optionally set a custom hostname (auto-detected by default) and adjust the notification policy.

### Run Test Check

The wizard offers a built-in dry-run option. Use it to confirm that your checkers are working before you save.

### Save & Exit

Writes the configuration to disk.

{{% /steps %}}

### Where the Config File is Saved

| Platform | System-wide | User |
|----------|------------|------|
| Linux | `/etc/update-watcher/config.yaml` | `~/.config/update-watcher/config.yaml` |
| macOS | -- | `~/.config/update-watcher/config.yaml` |

{{< callout type="info" >}}
  On Linux, if you ran the server setup during installation, the wizard writes to the system-wide location. Otherwise, it uses the user config directory.
{{< /callout >}}

{{< /tab >}}

{{< tab >}}

## Manual YAML Configuration

If you prefer to write the config file by hand, create it at the appropriate path and set secure permissions.

{{% steps %}}

### Create the config directory and file

```bash {filename="Terminal"}
mkdir -p ~/.config/update-watcher
touch ~/.config/update-watcher/config.yaml
chmod 600 ~/.config/update-watcher/config.yaml
```

### Write the configuration

Open the file in your editor and add a minimal configuration. Here is an example that checks APT for updates and sends notifications to Slack:

```yaml {filename="config.yaml"}
hostname: ""  # Auto-detected if left empty

watchers:
  - type: apt
    enabled: true
    options:
      use_sudo: true
      security_only: false
      hide_phased: true

notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"

settings:
  send_policy: "only-on-updates"
```

This configuration:

- Checks for APT package updates, using `sudo` to refresh the package list first.
- Hides phased updates (updates that Ubuntu is still rolling out gradually).
- Sends notifications to Slack, reading the webhook URL from an environment variable.
- Only sends notifications when updates are actually found.

### Set the environment variable

```bash {filename="Terminal"}
export SLACK_WEBHOOK_URL="<your-slack-webhook-url>"
```

{{% /steps %}}

{{< callout emoji="💡" >}}
  Instead of putting webhook URLs, tokens, and passwords directly in the config file, reference environment variables with `${VAR}`. This keeps secrets out of the config file itself. See [Environment Variables](../../configuration/environment-variables/) for more details.
{{< /callout >}}

For the full list of watchers, notifiers, and their options, see the [Config File Reference](../../configuration/config-file/).

{{< /tab >}}

{{< /tabs >}}

## Validating the Configuration

Before running a check, validate the config file to catch syntax errors or missing required fields:

```bash {filename="Terminal"}
update-watcher validate
```

If the configuration is valid, you will see a confirmation message. If there are errors, the output describes what needs to be fixed.

## Running Your First Check

### Without Notifications

Run all configured watchers and print results to the terminal without sending notifications:

```bash {filename="Terminal"}
update-watcher run --notify=false
```

Example output on an Ubuntu server with APT configured:

```text {filename="Output"}
Update-Watcher v1.x.x

Checking APT packages...
  3 updates available:
    - libssl3 (3.0.13-0ubuntu3.1 -> 3.0.13-0ubuntu3.2) [security]
    - curl (8.5.0-2ubuntu10.1 -> 8.5.0-2ubuntu10.2)
    - wget (1.21.4-1ubuntu4 -> 1.21.4-1ubuntu4.1)

Summary: 3 updates available (1 security)
```

{{< callout type="info" >}}
  Items marked `[security]` are security updates. Update-Watcher highlights these in notifications and can trigger special mentions (e.g., `@channel` in Slack).
{{< /callout >}}

### Understanding Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success, no updates found |
| 1 | Success, updates found |
| 2 | Partial failure (some checkers failed, others succeeded) |
| 3 | Complete failure (all checkers failed) |
| 4 | Configuration missing or invalid |

{{< callout type="info" >}}
  Exit code `1` is not an error -- it means updates were found. This is useful for scripting and monitoring.
{{< /callout >}}

## Sending Your First Notification

Once the dry run output looks correct, run with notifications enabled:

```bash {filename="Terminal"}
update-watcher run
```

Check your notification channel (Slack, Email, Telegram, etc.) to confirm the message was delivered. The notification includes:

- The hostname of the server.
- A list of available updates grouped by package manager.
- Security updates highlighted and called out separately.

{{< callout type="warning" >}}
  If the notification does not arrive, use verbose mode to see detailed debug output including HTTP request/response information: `update-watcher run --verbose`
{{< /callout >}}

## Checking Specific Watchers

To run only a specific checker (useful for testing individual configurations):

```bash {filename="Terminal"}
update-watcher run --only apt
```

This runs only the APT checker and skips all others. Replace `apt` with any configured watcher type.

## Scheduling Automatic Checks

Once you have confirmed that checks and notifications are working, schedule a daily cron job:

```bash {filename="Terminal"}
update-watcher install-cron
```

The default schedule is daily at 07:00. To customize the time:

```bash {filename="Terminal"}
update-watcher install-cron --time 09:00
```

See the [Quickstart](../quickstart/) for more details on cron scheduling.

## Next Steps

- [Configuration Reference](../../configuration/) -- Full guide to the YAML config, environment variables, and security best practices.
- [Checkers](../../checkers/) -- Detailed documentation for each of the 14 supported package manager checkers.
- [Notifiers](../../notifiers/) -- Configuration details for all 16 notification channels.
- [Server Setup](../../server-setup/) -- Production-ready setup with dedicated system user and minimal permissions.
