---
title: "Google Chat Notifications for Server Updates - Google Workspace Integration"
description: "Send server update notifications to Google Chat spaces via webhooks. Simple webhook-based integration for Google Workspace environments."
weight: 10
---

Update-Watcher sends server update notifications to Google Chat spaces using incoming webhooks. This integration is designed for teams that use Google Workspace and want update alerts delivered directly into their existing communication channels. The webhook setup is straightforward and requires no additional Google Cloud APIs or service accounts.

## Setup on Google Chat

{{% steps %}}

### Step 1: Open a Google Chat Space

Open **Google Chat** in a browser or the desktop app. Navigate to the space where you want to receive notifications.

### Step 2: Open Webhook Settings

Click the space name at the top to open the dropdown menu, then select **Manage webhooks** (or go to **Apps & integrations**).

### Step 3: Create a Webhook

Click **Create a webhook**. Enter a name (e.g., "Update Watcher") and optionally provide an avatar URL. Click **Save**.

### Step 4: Copy the Webhook URL

Copy the webhook URL. It looks like `https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN`.

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file. The URL contains authentication tokens -- protect it like a password.
{{< /callout >}}

```bash {filename="Terminal"}
export UPDATE_WATCHER_GCHAT_WEBHOOK="https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN"
```

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Google Chat incoming webhook URL. Use an environment variable reference to keep it out of version control. |
| `thread_key` | No | -- | Thread key for posting to a specific thread. When set, all notifications are grouped into a single thread in the space, keeping the conversation organized. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: googlechat
    webhook_url: ${UPDATE_WATCHER_GCHAT_WEBHOOK}
    thread_key: "update-watcher"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: googlechat
    webhook_url: ${UPDATE_WATCHER_GCHAT_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Google Chat notifications are delivered as card messages with:

- **Header** -- Server hostname and update count displayed in a card header.
- **Sections** -- Each checker with available updates is rendered as a card section with package names and version information.
- **Security labels** -- Security updates are highlighted within the card layout.

Google Chat's card format provides a clean, structured appearance that is easy to scan in busy spaces.
{{< /callout >}}

## Thread Grouping

The `thread_key` option controls how messages are organized in the space:

- **Without `thread_key`** -- Each notification creates a new top-level message in the space.
- **With `thread_key`** -- All notifications with the same thread key are grouped into a single thread. This keeps the space tidy and allows you to follow the update history in one conversation thread.

Using a consistent `thread_key` like `"update-watcher"` is recommended for spaces that receive other messages, so update notifications do not clutter the main conversation.

## Google Workspace Considerations

- **Space permissions** -- The webhook creator must have permission to manage apps and integrations in the space. Organization-level policies may restrict webhook creation.
- **Webhook limits** -- Google Chat imposes rate limits on incoming webhooks. For typical Update-Watcher usage (a few notifications per day), this is not a concern.
- **No authentication required** -- The webhook URL itself contains all necessary authentication tokens. Protect it like a password.

## Testing

Run Update-Watcher to verify the Google Chat notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Google Chat space for the card message. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` with a JSON response containing the message resource. Common errors include an invalid webhook URL (`404 Not Found`) or a disabled webhook.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
