---
title: Notifiers
description: "Update-Watcher supports 16 notification channels. Send server update alerts via Slack, Discord, Teams, Telegram, Email, push notifications, and more."
weight: 40
---

Notifiers determine where Update-Watcher sends its results after checking for available updates. Each notifier formats the update report for a specific platform or protocol and delivers it to the configured destination. Notifiers are delivery-only -- they never modify your system, install updates, or interact with package managers.

You can enable any number of notifiers simultaneously. For example, you might send a Slack message to your DevOps channel, trigger a PagerDuty incident for security updates, and archive a JSON payload to a webhook endpoint -- all from a single `update-watcher run` invocation.

## All Notifiers

| Notifier | Type | Description |
|----------|------|-------------|
| [Slack](slack/) | Chat | Rich Block Kit messages to Slack channels via incoming webhooks. |
| [Discord](discord/) | Chat | Embedded messages to Discord channels via webhooks. |
| [Microsoft Teams](teams/) | Chat | Adaptive Card messages to Teams channels via Workflow webhooks. |
| [Telegram](telegram/) | Chat | Markdown-formatted messages to Telegram chats and groups via Bot API. |
| [Email](email/) | Email | HTML-formatted email via any SMTP server with STARTTLS support. |
| [ntfy](ntfy/) | Push | Push notifications via ntfy.sh or self-hosted ntfy servers. |
| [Pushover](pushover/) | Push | Push notifications to iOS, Android, and Desktop via Pushover API. |
| [Gotify](gotify/) | Push | Push notifications via self-hosted Gotify server. |
| [Home Assistant](homeassistant/) | Smart Home | Notifications through Home Assistant's notify service. |
| [Google Chat](googlechat/) | Chat | Messages to Google Chat spaces via webhooks. |
| [Matrix](matrix/) | Chat | Messages to Matrix rooms via the client-server API. |
| [Mattermost](mattermost/) | Chat | Messages to Mattermost channels via incoming webhooks. |
| [Rocket.Chat](rocketchat/) | Chat | Messages to Rocket.Chat channels via incoming webhooks. |
| [PagerDuty](pagerduty/) | Monitoring | Incident triggers via Events API v2 for security updates. |
| [Pushbullet](pushbullet/) | Push | Cross-device push notifications via Pushbullet API. |
| [Webhook](webhook/) | Generic | JSON payloads to any HTTP endpoint for custom integrations. |

## Categories

### Chat Platforms

Send update notifications directly into the channels where your team already communicates. All chat notifiers support rich formatting, security update highlighting, and optional mentions for critical updates.

- [Slack](slack/) -- Block Kit formatting with security mentions
- [Discord](discord/) -- Embedded messages with custom bot identity
- [Microsoft Teams](teams/) -- Adaptive Cards via Workflow webhooks
- [Telegram](telegram/) -- Bot API with Markdown formatting
- [Google Chat](googlechat/) -- Google Workspace webhook integration
- [Matrix](matrix/) -- Decentralized, self-hostable chat
- [Mattermost](mattermost/) -- Self-hosted Slack alternative
- [Rocket.Chat](rocketchat/) -- Self-hosted team communication

### Push Notifications

Receive update alerts on your phone, tablet, or desktop without needing a chat application open. Ideal for solo administrators or after-hours monitoring.

- [ntfy](ntfy/) -- Open-source, self-hostable push notifications
- [Pushover](pushover/) -- iOS, Android, and Desktop with priority levels
- [Gotify](gotify/) -- Fully self-hosted push notification server
- [Pushbullet](pushbullet/) -- Cross-device push notifications

### Email

Traditional email delivery for teams that rely on inbox-based workflows or need an audit trail.

- [Email](email/) -- HTML email via any SMTP server

### Monitoring and Incident Management

Integrate update notifications into your existing monitoring and incident response workflows.

- [PagerDuty](pagerduty/) -- Trigger incidents for security updates with configurable severity

### Smart Home

Integrate server update alerts into your home automation setup.

- [Home Assistant](homeassistant/) -- Notify service integration for dashboards and automations

### Generic

Send raw update data to any HTTP endpoint for custom processing, logging, or integration with systems not directly supported.

- [Webhook](webhook/) -- JSON payloads with custom headers and authentication

## Send Policy

The `send_policy` setting in your configuration controls when notifications are sent:

- **`only-on-updates`** (default) -- Notifications are sent only when at least one checker reports available updates. Silent when everything is up to date.
- **`always`** -- Notifications are sent after every run, even if no updates are found. Useful as a heartbeat to confirm Update-Watcher is running.

This setting applies globally to all configured notifiers. Set it in the top-level configuration:

```yaml {filename="config.yaml"}
send_policy: "only-on-updates"
```

## Security Mentions

Several notifiers support a `mention_on_security` option (or platform-specific equivalent like `mention_role` for Discord). When enabled, the notifier adds a mention or highlight to the notification whenever security updates are detected. This draws immediate attention to critical patches in busy channels.

## Next Steps

- [Getting Started](../getting-started/) -- Install Update-Watcher and run your first check.
- [Checkers](../checkers/) -- Configure which package managers to monitor.
- [Configuration](../configuration/) -- Full YAML configuration reference.
