---
title: "Telegram Notifications for Server Updates - Bot API Integration"
description: "Send server update notifications to Telegram chats and groups via Bot API. Markdown-formatted messages with security highlighting. Easy bot setup guide."
weight: 4
---

Update-Watcher sends server update notifications to Telegram chats and groups using the Telegram Bot API. Messages are formatted with Markdown for clear structure, with security updates distinctly highlighted. The setup process involves creating a bot through BotFather and adding it to your target chat or group.

## Setup on Telegram

{{% steps %}}

### Step 1: Create a Bot

1. Open Telegram and search for **@BotFather**.
2. Send the `/newbot` command.
3. Follow the prompts: provide a display name and a username for the bot (must end in `bot`).
4. BotFather replies with your **bot token**. It looks like `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`.
5. Store the token securely -- anyone with this token can control the bot.

### Step 2: Get the Chat ID

For a **private chat** with the bot:

1. Send any message to your new bot in Telegram.
2. Open `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates` in a browser.
3. Look for `"chat":{"id":123456789}` in the JSON response. That number is your chat ID.

For a **group chat**:

1. Add the bot to your group.
2. Send a message in the group that mentions the bot (e.g., `/start@yourbot`).
3. Check the `getUpdates` URL as above. Group chat IDs are negative numbers (e.g., `-1001234567890`).

### Step 3: Store Credentials

Store the credentials in environment variables:

```bash {filename="Terminal"}
export UPDATE_WATCHER_TELEGRAM_TOKEN="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
export UPDATE_WATCHER_TELEGRAM_CHAT_ID="-1001234567890"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the bot token in an environment variable rather than placing it directly in your configuration file. Anyone with the token can control the bot.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `bot_token` | Yes | -- | Telegram Bot API token from BotFather. Use an environment variable reference. |
| `chat_id` | Yes | -- | Target chat ID. Positive for private chats, negative for groups. Can be a string or number. |
| `disable_notification` | No | `false` | When `true`, sends the message silently. Users receive the message without a push notification or sound. Useful for non-urgent update reports. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: telegram
    bot_token: ${UPDATE_WATCHER_TELEGRAM_TOKEN}
    chat_id: ${UPDATE_WATCHER_TELEGRAM_CHAT_ID}
    disable_notification: false
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: telegram
    bot_token: ${UPDATE_WATCHER_TELEGRAM_TOKEN}
    chat_id: ${UPDATE_WATCHER_TELEGRAM_CHAT_ID}
```

### Silent Notifications for Groups

To post updates to a group without triggering push notifications for all members:

```yaml {filename="config.yaml"}
notifiers:
  - type: telegram
    bot_token: ${UPDATE_WATCHER_TELEGRAM_TOKEN}
    chat_id: ${UPDATE_WATCHER_TELEGRAM_CHAT_ID}
    disable_notification: true
```

## Message Format

{{< callout type="info" >}}
Telegram notifications use Markdown formatting and include:

- **Header** -- Server hostname in bold with an update count.
- **Checker sections** -- Each checker is listed with its name in bold, followed by a formatted list of packages with current and new versions.
- **Security labels** -- Security updates are marked with a distinct label in the message text.
- **Clean layout** -- The message is structured to be readable on mobile devices, where most Telegram usage occurs.
{{< /callout >}}

## Bot Permissions

The bot requires minimal permissions:

- For **private chats**, no special permissions are needed. Simply start a conversation with the bot.
- For **groups**, the bot must be a member of the group. It does not need administrator privileges. If the group has **"Who can send messages"** restricted, the bot must be added as an exception or granted posting rights.

## Testing

Run Update-Watcher to verify the Telegram notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Telegram chat for the message. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns a JSON response with `"ok": true` from the Telegram API. Common errors include an invalid bot token (`401 Unauthorized`) or the bot not being a member of the target chat (`400 Bad Request`).

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
