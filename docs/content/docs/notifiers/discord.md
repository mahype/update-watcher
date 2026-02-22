---
title: "Discord Notifications for Server Updates - Update-Watcher Discord Integration"
description: "Receive server update notifications in Discord via webhook embeds. Custom bot name, avatar, role mentions for security updates. Easy webhook configuration."
weight: 2
---

Update-Watcher delivers server update notifications to Discord channels using webhook embeds. Notifications appear as rich embedded messages with color-coded sections, grouped by package manager. You can customize the bot's display name and avatar, and configure role mentions to alert team members when security updates are detected.

## Setup on Discord

Create a webhook in your Discord server:

{{% steps %}}

### Step 1: Open Channel Settings

Open your Discord server and navigate to the channel where you want notifications. Click the gear icon next to the channel name to open **Channel Settings**.

### Step 2: Go to Integrations

Select **Integrations** from the left sidebar.

### Step 3: Create a Webhook

Click **Webhooks**, then **New Webhook**.

### Step 4: Configure the Webhook

Give the webhook a name (e.g., "Update Watcher"). Optionally upload an avatar image.

### Step 5: Copy the URL

Click **Copy Webhook URL**. It looks like `https://discord.com/api/webhooks/1234567890/abcdefghijklmnop`.

### Step 6: Save Changes

Click **Save Changes**.

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

```bash {filename="Terminal"}
export UPDATE_WATCHER_DISCORD_WEBHOOK="https://discord.com/api/webhooks/1234567890/abcdefghijklmnop"
```

### Finding a Role ID for Mentions

If you want to mention a Discord role when security updates are detected:

1. Open **Server Settings** and go to **Roles**.
2. Enable **Developer Mode** in your Discord user settings under **Advanced**.
3. Right-click the role you want to mention and select **Copy Role ID**.

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Discord webhook URL. Use an environment variable reference for security. |
| `username` | No | `Update Watcher` | Display name for the bot when posting messages. Overrides the name set in the webhook configuration. |
| `avatar_url` | No | -- | URL to an image to use as the bot's avatar. Must be a publicly accessible HTTPS URL. |
| `mention_role` | No | -- | Discord role ID to mention when security updates are detected. This is a numeric string (e.g., `"123456789012345678"`), not the role name. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: discord
    webhook_url: ${UPDATE_WATCHER_DISCORD_WEBHOOK}
    username: "Update Watcher"
    avatar_url: "https://example.com/update-watcher-avatar.png"
    mention_role: "123456789012345678"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: discord
    webhook_url: ${UPDATE_WATCHER_DISCORD_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Discord notifications are delivered as embedded messages with the following structure:

- **Title** -- Server hostname and update summary.
- **Color** -- The embed sidebar color changes based on content: orange for regular updates, red when security updates are present.
- **Fields** -- Each checker with available updates gets its own field listing packages and versions.
- **Role mention** -- When `mention_role` is configured and security updates are detected, the role is mentioned above the embed to trigger Discord notifications for all role members.
- **Timestamp** -- Displayed in the embed footer.

Discord's embed format ensures messages look clean on both desktop and mobile clients, and they collapse neatly in busy channels.
{{< /callout >}}

## Testing

Run Update-Watcher to verify the Discord notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the configured Discord channel for the embedded message. If no updates are available and `send_policy` is set to `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `204 No Content` from Discord's webhook endpoint.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
