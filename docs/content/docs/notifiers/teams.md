---
title: "Microsoft Teams Notifications for Server Updates - Adaptive Card Integration"
description: "Send server update notifications to Microsoft Teams via Workflow webhooks with Adaptive Card formatting. Enterprise-ready update alerting for IT teams."
weight: 3
---

Update-Watcher sends server update notifications to Microsoft Teams channels using Workflow webhooks with Adaptive Card formatting. This integration uses the modern Teams Workflows approach, which replaces the deprecated Office 365 Connectors. Adaptive Cards provide a rich, structured layout that renders consistently across Teams desktop, web, and mobile clients.

## Setup on Microsoft Teams

Create a Workflow webhook in your Teams channel:

{{% steps %}}

### Step 1: Open Channel Settings

Open Microsoft Teams and navigate to the channel where you want notifications. Click the three-dot menu next to the channel name and select **Manage channel**.

### Step 2: Go to Workflows

Select **Connectors & workflows** (or go to the **Workflows** tab).

### Step 3: Select the Webhook Workflow

Search for **"Post to a channel when a webhook request is received"** and select it.

### Step 4: Name the Workflow

Give the workflow a name (e.g., "Update Watcher Notifications") and click **Next**.

### Step 5: Confirm Target Channel

Confirm the target team and channel, then click **Add workflow**.

### Step 6: Copy the Webhook URL

Copy the webhook URL that is generated. It looks like `https://prod-XX.westus.logic.azure.com:443/workflows/...`.

{{% /steps %}}

{{< callout type="warning" >}}
Store the webhook URL in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

```bash {filename="Terminal"}
export UPDATE_WATCHER_TEAMS_WEBHOOK="<your-teams-webhook-url>"
```

{{< callout type="warning" >}}
**Do not use Office 365 Connectors.** Microsoft has deprecated the legacy Office 365 Connector webhooks. Update-Watcher uses the newer Workflow-based webhooks that are fully supported going forward. If you have an existing O365 Connector URL, it will stop working -- migrate to a Workflow webhook using the steps above.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `webhook_url` | Yes | -- | Microsoft Teams Workflow webhook URL. Use an environment variable reference to keep it out of version control. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: teams
    webhook_url: ${UPDATE_WATCHER_TEAMS_WEBHOOK}
```

## Message Format

{{< callout type="info" >}}
Update-Watcher builds an Adaptive Card payload that includes:

- **Header** -- Server hostname in a prominent heading with an update count summary.
- **Checker sections** -- Each checker that found updates is rendered as a separate container with a bold title and a list of packages with their current and available versions.
- **Security highlighting** -- Security updates are displayed with a distinct accent color and label to draw attention to critical patches.
- **Compact layout** -- The card design is optimized for Teams' column widths and ensures readability on both desktop and mobile.

Adaptive Cards render natively in Teams without requiring any additional apps or permissions beyond the Workflow webhook.
{{< /callout >}}

## Enterprise Considerations

The Workflow webhook approach has several advantages for enterprise Teams environments:

- **No third-party app approval required** -- Workflow webhooks are a built-in Teams feature. No need to request IT approval for an external app.
- **Per-channel control** -- Each webhook is scoped to a single channel, following the principle of least privilege.
- **Audit trail** -- Workflow runs are logged in Power Automate, providing visibility into notification delivery.
- **No expiration** -- Unlike some connector types, Workflow webhooks do not expire automatically.

## Testing

Run Update-Watcher to verify the Teams notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the configured Teams channel for the Adaptive Card message. If no updates are available and `send_policy` is set to `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `202 Accepted` from the Workflow endpoint. If you receive a `404` or `401`, verify that the Workflow is still active in the Teams channel settings.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
