---
title: "Rocket.Chat Notifications for Server Updates - Self-Hosted Chat Integration"
description: "Send server update notifications to Rocket.Chat channels via incoming webhooks. Self-hosted team chat integration for update monitoring."
weight: 13
---

Update-Watcher sends server update notifications to [Rocket.Chat](https://www.rocket.chat) channels using incoming webhooks. Rocket.Chat is a self-hosted, open-source team communication platform with a rich feature set. The webhook integration requires minimal setup and supports customizable bot names and target channels, allowing you to keep update notifications organized within your existing team communication.

## Setup on Rocket.Chat

{{% steps %}}

### Step 1: Enable Incoming Webhooks

An administrator must enable incoming webhooks on the Rocket.Chat instance:

1. Go to **Administration** (click the three-dot menu -> **Administration** or **Workspace** -> **Settings**).
2. Navigate to **Integrations**.
3. Ensure incoming webhooks are enabled.

### Step 2: Create an Incoming Webhook

1. In the Administration panel, go to **Integrations**.
2. Click **New Integration** and select **Incoming WebHook**.
3. Fill in the details:
   - **Enabled**: Toggle on.
   - **Name**: Update Watcher
   - **Post to Channel**: Select the target channel (e.g., `#server-updates`).
   - **Post as**: Choose a username for the bot or leave it to be overridden by the payload.
4. Click **Save Changes**.
5. Scroll down to find the **Webhook URL** and copy it. It looks like `https://rocketchat.example.com/hooks/xxxxxxxxxxxxxxxxxxxxxxxx/yyyyyyyyyyyyyyyyyyyyyyyy`.

### Step 3: Store the Webhook URL

Store the webhook URL in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_ROCKETCHAT_WEBHOOK="<your-rocketchat-webhook-url>"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Rocket.Chat incoming webhook URL. Use an environment variable reference. |
| `channel` | No | -- | Override the default channel set in the webhook configuration. Use the channel name with `#` prefix (e.g., `#server-updates`) or a username with `@` prefix for direct messages. |
| `username` | No | `Update Watcher` | Display name for the bot when posting messages. Overrides the default set in the webhook configuration. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: rocketchat
    webhook_url: ${UPDATE_WATCHER_ROCKETCHAT_WEBHOOK}
    channel: "#server-updates"
    username: "Update Watcher"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: rocketchat
    webhook_url: ${UPDATE_WATCHER_ROCKETCHAT_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Rocket.Chat notifications use Markdown formatting and include:

- **Header** -- Server hostname and update summary in bold text.
- **Checker sections** -- Each checker with available updates is listed with its name as a bold heading, followed by package names with current and available versions.
- **Security labels** -- Security updates are marked with a distinct text label for quick identification.
- **Attachments** -- Colored attachment bars may be used to differentiate regular updates from security-critical ones.

Rocket.Chat's Markdown renderer ensures clean formatting on desktop, mobile, and web clients.
{{< /callout >}}

## Channel Override

When the `channel` option is set, notifications are posted to the specified channel instead of the default one configured in the webhook:

- Use `#channel-name` for public or private channels.
- Use `@username` to send a direct message.
- The webhook must have permission to post to the target channel.

## Testing

Run Update-Watcher to verify the Rocket.Chat notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the configured Rocket.Chat channel for the notification. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` from the Rocket.Chat webhook endpoint. Common errors include a disabled integration (`403`) or an incorrect webhook URL (`404`).

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
