---
title: "Mattermost Notifications for Server Updates - Self-Hosted Chat Integration"
description: "Send server update notifications to Mattermost channels via incoming webhooks. Slack-compatible webhook format for self-hosted team communication."
weight: 12
---

Update-Watcher sends server update notifications to [Mattermost](https://mattermost.com) channels using incoming webhooks. Mattermost is a self-hosted team communication platform that supports Slack-compatible webhook payloads, making it a natural fit for Update-Watcher. You can customize the bot's display name, avatar, and target channel, and keep your notification infrastructure entirely on your own servers.

## Setup on Mattermost

{{% steps %}}

### Step 1: Enable Incoming Webhooks

Incoming webhooks must be enabled by a system administrator:

1. Go to **System Console** (Main Menu -> System Console).
2. Navigate to **Integrations** -> **Integration Management**.
3. Ensure **Enable Incoming Webhooks** is set to `true`.
4. Optionally, set **Enable integrations to override usernames** and **Enable integrations to override profile picture icons** to `true` if you want Update-Watcher to use a custom display name and avatar.

### Step 2: Create an Incoming Webhook

1. In Mattermost, click **Main Menu** (hamburger icon) -> **Integrations**.
2. Select **Incoming Webhooks** -> **Add Incoming Webhook**.
3. Fill in the details:
   - **Title**: Update Watcher
   - **Description**: Server update notifications
   - **Channel**: Select the default channel for notifications
4. Click **Save**.
5. Copy the webhook URL. It looks like `https://mattermost.example.com/hooks/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`.

### Step 3: Store the Webhook URL

Store the webhook URL in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_MATTERMOST_WEBHOOK="https://mattermost.example.com/hooks/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Mattermost incoming webhook URL. Use an environment variable reference. |
| `channel` | No | -- | Override the default channel set in the webhook configuration. Use the channel name (e.g., `server-updates`) or ID. Requires the webhook to have permission to post to the target channel. |
| `username` | No | `Update Watcher` | Display name for the bot when posting messages. Requires "Enable integrations to override usernames" in system settings. |
| `icon_url` | No | -- | URL to an image to use as the bot's avatar. Requires "Enable integrations to override profile picture icons" in system settings. Must be a publicly accessible URL. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: mattermost
    webhook_url: ${UPDATE_WATCHER_MATTERMOST_WEBHOOK}
    channel: "server-updates"
    username: "Update Watcher"
    icon_url: "https://example.com/update-watcher-icon.png"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: mattermost
    webhook_url: ${UPDATE_WATCHER_MATTERMOST_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Mattermost notifications use Markdown formatting and include:

- **Header** -- Server hostname and update count in bold.
- **Checker sections** -- Each checker with available updates is listed with a bold heading and a formatted table or list of packages with their current and available versions.
- **Security highlights** -- Security updates are marked with a distinct label.
- **Attachments** -- Optional colored attachment bars can be used to visually distinguish regular updates from security updates.

Mattermost renders the Markdown natively, producing a clean and readable notification in both desktop and mobile clients.
{{< /callout >}}

## Channel Override

The `channel` option allows you to send notifications to a different channel than the one configured in the webhook:

- This is useful if you want to route notifications dynamically or use a single webhook for multiple purposes.
- The channel name should not include the `~` prefix -- just use the plain name (e.g., `server-updates`).
- The Mattermost system administrator must enable "Allow webhooks to post to any channel" for this to work with channels other than the webhook's default.

## Testing

Run Update-Watcher to verify the Mattermost notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the configured Mattermost channel for the notification. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` from the Mattermost webhook endpoint. Common errors include a disabled webhook (`403 Forbidden`) or a mistyped webhook URL (`404 Not Found`).

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
