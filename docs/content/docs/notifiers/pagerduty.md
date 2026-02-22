---
title: "PagerDuty Integration for Server Updates - Security Update Incident Triggers"
description: "Trigger PagerDuty incidents when security updates are available on your servers. Events API v2 integration with configurable severity levels for IT teams."
weight: 14
---

Update-Watcher integrates with [PagerDuty](https://www.pagerduty.com) through the Events API v2 to trigger incidents when updates are available on your servers. This integration is designed for operations teams that use PagerDuty for incident management and want security updates to follow the same escalation policies as other critical alerts. Severity levels are configurable, so you can control how urgently PagerDuty routes the notification.

## Setup on PagerDuty

{{% steps %}}

### Step 1: Create a Service (or Use an Existing One)

1. Log in to PagerDuty and go to **Services** -> **Service Directory**.
2. Click **New Service** (or select an existing service where you want update alerts).
3. Give the service a name (e.g., "Server Updates") and assign an escalation policy.
4. Under **Integrations**, select **Events API v2** and click **Create Service**.

### Step 2: Get the Routing Key

1. On the service page, go to the **Integrations** tab.
2. If you used an existing service, click **Add Integration** and select **Events API v2**.
3. Copy the **Integration Key** (also called the routing key). It looks like a 32-character hexadecimal string.

### Step 3: Store the Routing Key

Store the routing key in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_PAGERDUTY_KEY="<your-pagerduty-routing-key>"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the routing key in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `routing_key` | Yes | -- | PagerDuty Events API v2 integration key (routing key). Use an environment variable reference. |
| `severity` | No | `warning` | Default severity level for triggered events. Values: `info`, `warning`, `error`, `critical`. This determines how PagerDuty routes and prioritizes the incident. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: pagerduty
    routing_key: ${UPDATE_WATCHER_PAGERDUTY_KEY}
    severity: "warning"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: pagerduty
    routing_key: ${UPDATE_WATCHER_PAGERDUTY_KEY}
```

### Critical Severity for Security Updates

If your PagerDuty escalation policy treats different severity levels differently, you might configure a higher severity:

```yaml {filename="config.yaml"}
notifiers:
  - type: pagerduty
    routing_key: ${UPDATE_WATCHER_PAGERDUTY_KEY}
    severity: "critical"
```

## Event Payload

{{< callout type="info" >}}
Update-Watcher sends a trigger event to PagerDuty's Events API v2 with the following structure:

- **Summary** -- Server hostname and count of available updates (e.g., "12 updates available on web-prod-01").
- **Severity** -- The configured severity level (`info`, `warning`, `error`, or `critical`).
- **Source** -- The server hostname.
- **Component** -- "Update Watcher".
- **Custom details** -- A structured breakdown of all available updates grouped by checker, including package names, current versions, available versions, and security flags.

PagerDuty uses the summary and severity to route the incident through your escalation policies. The custom details are available in the incident view for on-call responders.
{{< /callout >}}

## Severity Levels

| Severity | Typical Use | PagerDuty Behavior |
|----------|-------------|-------------------|
| `info` | Routine updates, no urgency | Low-urgency notification, may not page on-call |
| `warning` | Updates available, should be addressed soon | Standard notification through escalation policy |
| `error` | Security updates or critical packages | Higher urgency, faster escalation |
| `critical` | Immediate action required | Highest urgency, immediate paging |

Choose the severity level that matches your organization's incident response expectations. For most teams, `warning` is appropriate for general updates, with `error` or `critical` reserved for environments where unpatched servers represent a compliance risk.

## Deduplication

Update-Watcher includes a deduplication key in each PagerDuty event based on the server hostname. This means:

- Multiple runs reporting updates on the same server do not create duplicate incidents.
- PagerDuty merges the events into a single incident until it is resolved.
- Once you patch the server and Update-Watcher reports no more updates, the incident can be resolved manually or through PagerDuty automation.

## Best Practices

- **Pair with another notifier** -- Use PagerDuty alongside a chat notifier (Slack, Teams, etc.) for full visibility. PagerDuty handles escalation for critical updates, while the chat notification provides the daily summary.
- **Use `send_policy: "only-on-updates"`** -- Avoid triggering PagerDuty events when there are no updates. This is the default behavior.
- **Match severity to your escalation policy** -- Ensure your PagerDuty service's escalation policy is configured to handle the chosen severity level appropriately.

## Testing

Run Update-Watcher to verify the PagerDuty integration:

```bash {filename="Terminal"}
update-watcher run
```

Check the PagerDuty service for a triggered incident. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a trigger.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `202 Accepted` with a JSON body containing `"status": "success"` and a `dedup_key`. Common errors include an invalid routing key (`"Invalid Routing Key"`) or a deactivated integration.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a trigger and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
