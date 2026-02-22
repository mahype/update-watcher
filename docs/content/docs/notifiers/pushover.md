---
title: "Pushover Notifications for Server Updates - iOS, Android & Desktop Alerts"
description: "Receive server update push notifications on iOS, Android, and Desktop via Pushover. Priority levels, custom sounds, device targeting. Quick API setup."
weight: 7
---

Update-Watcher sends push notifications to your iOS, Android, and Desktop devices via the [Pushover](https://pushover.net) service. Pushover provides reliable delivery with configurable priority levels, custom sounds, and per-device targeting. It is a popular choice for server administrators who want immediate alerts on their phone without relying on a chat application.

## Setup on Pushover

{{% steps %}}

### Step 1: Create a Pushover Account

1. Sign up at [pushover.net](https://pushover.net).
2. Install the Pushover app on your devices ([iOS](https://pushover.net/clients/ios), [Android](https://pushover.net/clients/android), or [Desktop](https://pushover.net/clients/desktop)).
3. Log in on each device. Your **user key** is displayed on the main dashboard page after logging in at pushover.net.

### Step 2: Create an Application

1. Go to [pushover.net/apps/build](https://pushover.net/apps/build).
2. Enter a name (e.g., "Update Watcher") and optionally upload an icon.
3. Click **Create Application**.
4. Copy the **API Token/Key** displayed on the application page.

### Step 3: Store Credentials

Store the credentials in environment variables:

```bash {filename="Terminal"}
export UPDATE_WATCHER_PUSHOVER_APP_TOKEN="<your-pushover-app-token>"
export UPDATE_WATCHER_PUSHOVER_USER_KEY="<your-pushover-user-key>"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the API token and user key in environment variables rather than placing them directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `app_token` | Yes | -- | Pushover application API token. Use an environment variable reference. |
| `user_key` | Yes | -- | Your Pushover user key (or a group key for multiple recipients). Use an environment variable reference. |
| `device` | No | -- | Target a specific device name. If omitted, the notification is sent to all devices registered to the user. |
| `priority` | No | `0` | Notification priority. `-2` = lowest (no alert), `-1` = low (quiet), `0` = normal, `1` = high (bypass quiet hours), `2` = emergency (repeats until acknowledged). |
| `sound` | No | -- | Notification sound name. See [Pushover sound list](https://pushover.net/api#sounds) for available options (e.g., `pushover`, `bike`, `siren`, `spacealarm`). |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: pushover
    app_token: ${UPDATE_WATCHER_PUSHOVER_APP_TOKEN}
    user_key: ${UPDATE_WATCHER_PUSHOVER_USER_KEY}
    device: "my-phone"
    priority: 0
    sound: "pushover"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: pushover
    app_token: ${UPDATE_WATCHER_PUSHOVER_APP_TOKEN}
    user_key: ${UPDATE_WATCHER_PUSHOVER_USER_KEY}
```

### High Priority for Security Updates

You can configure a second Pushover notifier instance with a higher priority specifically for security-critical notifications, or rely on Update-Watcher's built-in security detection to highlight critical updates in the notification body:

```yaml {filename="config.yaml"}
notifiers:
  - type: pushover
    app_token: ${UPDATE_WATCHER_PUSHOVER_APP_TOKEN}
    user_key: ${UPDATE_WATCHER_PUSHOVER_USER_KEY}
    priority: 1
    sound: "siren"
```

## Notification Format

{{< callout type="info" >}}
Pushover notifications include:

- **Title** -- Server hostname and update summary.
- **Body** -- A text listing of each checker with available updates, showing package names and versions. Security updates are labeled distinctly.
- **Priority indicator** -- The configured priority level determines how the notification is displayed on the device (silent, normal alert, bypass quiet hours, or emergency repeat).
- **URL** -- A supplementary URL can be included linking back to Update-Watcher documentation or your server dashboard.
{{< /callout >}}

## Priority Levels

| Priority | Behavior |
|----------|----------|
| `-2` (Lowest) | No notification sound or vibration. Appears in the notification list only. |
| `-1` (Low) | Quiet notification. No sound but may show briefly. |
| `0` (Normal) | Standard notification with sound and display. |
| `1` (High) | Bypasses quiet hours. The notification always triggers an alert. |
| `2` (Emergency) | Repeats the notification every 30 seconds until the user acknowledges it. Use with caution. |

## Testing

Run Update-Watcher to verify the Pushover notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Pushover app on your devices. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns a JSON response with `"status": 1` from the Pushover API. Common errors include an invalid app token (`"token is invalid"`) or an invalid user key.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
