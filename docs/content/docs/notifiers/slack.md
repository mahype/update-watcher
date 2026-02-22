---
title: "Slack Notifications for Server Updates - Update-Watcher Slack Integration"
description: "Send server update notifications to Slack with rich Block Kit formatting. Security highlighting, @channel mentions, emoji support. Set up in under 2 minutes."
weight: 1
---

Update-Watcher sends server update notifications to any Slack channel using incoming webhooks. Messages are formatted with Slack's Block Kit for a clean, structured layout that groups updates by package manager and highlights security patches. Setup takes under two minutes and requires no bot permissions beyond posting messages.

## Setup on Slack

Follow these steps to create an incoming webhook for your Slack workspace:

{{% steps %}}

### Step 1: Create a Slack App

Go to [api.slack.com/apps](https://api.slack.com/apps) and click **Create New App**.

### Step 2: Select From Scratch

Select **From scratch**, give it a name (e.g., "Update Watcher"), and choose your workspace.

### Step 3: Enable Incoming Webhooks

In the left sidebar, click **Incoming Webhooks**.

### Step 4: Activate Webhooks

Toggle **Activate Incoming Webhooks** to On.

### Step 5: Add Webhook to Workspace

Click **Add New Webhook to Workspace** at the bottom of the page.

### Step 6: Select Channel

Select the channel where you want update notifications to appear and click **Allow**.

### Step 7: Copy the Webhook URL

Copy the webhook URL. It looks like `https://hooks.slack.com/services/T.../B.../...`.

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

```bash {filename="Terminal"}
export UPDATE_WATCHER_SLACK_WEBHOOK="<your-slack-webhook-url>"
```

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Slack incoming webhook URL. Use an environment variable reference to keep it out of version control. |
| `mention_on_security` | No | -- | User or group to mention when security updates are detected. Examples: `@channel`, `@here`, `<@U01234567>` for a specific user, `<!subteam^S01234567>` for a user group. |
| `use_emoji` | No | `true` | Whether to include emoji indicators in the notification. When enabled, package names and security labels are prefixed with contextual emoji. |

## Configuration Example

Add the Slack notifier to the `notifiers` section of your Update-Watcher configuration file:

```yaml {filename="config.yaml"}
notifiers:
  - type: slack
    webhook_url: ${UPDATE_WATCHER_SLACK_WEBHOOK}
    mention_on_security: "@channel"
    use_emoji: true
```

### Minimal Configuration

If you only need basic notifications without mentions or emoji:

```yaml {filename="config.yaml"}
notifiers:
  - type: slack
    webhook_url: ${UPDATE_WATCHER_SLACK_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Update-Watcher uses Slack's Block Kit to produce structured messages. Each notification includes:

- **Header** -- The hostname of the server and a summary of available updates.
- **Checker sections** -- One section per enabled checker that found updates, listing each package with its current and available version.
- **Security highlights** -- Security updates are called out with a distinct label. When `mention_on_security` is set, the configured mention is included at the top of the message to trigger an alert.
- **Footer** -- Timestamp and Update-Watcher version.

The Block Kit formatting ensures messages render cleanly on desktop, mobile, and in Slack's notification previews.
{{< /callout >}}

## Security Mentions

The `mention_on_security` option is particularly useful for operations teams. When Update-Watcher detects security updates (for example, through APT's security repository classification or DNF's advisory metadata), the configured mention string is prepended to the message. This triggers Slack notifications for the mentioned users or groups, ensuring critical patches are not overlooked in busy channels.

Common values:

- `@channel` -- Notifies everyone in the channel.
- `@here` -- Notifies only active members.
- `<@U01234567>` -- Notifies a specific user (use their Slack member ID).
- `<!subteam^S01234567>` -- Notifies a specific user group.

## Testing

After configuring the Slack notifier, run a test to verify delivery:

```bash {filename="Terminal"}
update-watcher run
```

Check the configured Slack channel for the notification. If no updates are available on your system and your `send_policy` is set to the default `only-on-updates`, force a notification by temporarily setting `send_policy: "always"` in your configuration.

For troubleshooting, run with verbose output:

```bash {filename="Terminal"}
update-watcher run --verbose
```

This prints detailed logs including the HTTP response from Slack's webhook endpoint. A `200 OK` response confirms successful delivery.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
