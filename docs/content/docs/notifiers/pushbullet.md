---
title: "Pushbullet Notifications for Server Updates - Cross-Device Push Notifications"
description: "Receive server update push notifications across all your devices via Pushbullet. Simple access token setup for cross-platform update alerts."
weight: 15
---

Update-Watcher sends server update notifications to all your devices through [Pushbullet](https://www.pushbullet.com). Pushbullet delivers push notifications across Android, iOS, Windows, and browser extensions simultaneously. The setup requires only a single access token, making it one of the simplest notifiers to configure for individual administrators who want update alerts on their personal devices.

## Setup on Pushbullet

{{% steps %}}

### Step 1: Create an Account

1. Sign up or log in at [pushbullet.com](https://www.pushbullet.com).
2. Install the Pushbullet app or browser extension on the devices where you want to receive notifications:
   - **Android** -- Available on [Google Play](https://play.google.com/store/apps/details?id=com.pushbullet.android).
   - **Windows** -- Desktop client available at [pushbullet.com/apps](https://www.pushbullet.com/apps).
   - **Browser** -- Extensions for Chrome, Firefox, and Opera.

### Step 2: Create an Access Token

1. Go to [pushbullet.com/account](https://www.pushbullet.com/#settings/account) (or click your avatar -> Settings).
2. Scroll to the **Access Tokens** section.
3. Click **Create Access Token**.
4. Copy the token. It looks like `o.aBcDeFgHiJkLmNoPqRsTuVwXyZ012345`.

### Step 3: Store the Token

Store the token in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_PUSHBULLET_TOKEN="<your-pushbullet-access-token>"
```

### Step 4: Optional Device Targeting

If you want to send notifications to a specific device instead of all devices:

1. Go to [pushbullet.com/account](https://www.pushbullet.com/#settings/account) and click **Devices**.
2. Use the Pushbullet API to list devices:

```bash {filename="Terminal"}
curl -H "Access-Token: YOUR_TOKEN" https://api.pushbullet.com/v2/devices
```

3. Copy the `iden` field for the target device.

{{% /steps %}}

{{< callout type="warning" >}}
Store the access token in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `access_token` | Yes | -- | Pushbullet access token. Use an environment variable reference. |
| `device_iden` | No | -- | Target a specific device by its identifier string. If omitted, the notification is sent to all devices associated with the account. |
| `channel_tag` | No | -- | Publish to a Pushbullet channel by its tag. Useful for broadcasting update notifications to multiple subscribers. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: pushbullet
    access_token: ${UPDATE_WATCHER_PUSHBULLET_TOKEN}
    device_iden: "ujpah72o0sjAoRtnM0jc"
```

### Minimal Configuration (All Devices)

```yaml {filename="config.yaml"}
notifiers:
  - type: pushbullet
    access_token: ${UPDATE_WATCHER_PUSHBULLET_TOKEN}
```

### Publish to a Channel

If you run a Pushbullet channel for your team, you can publish update notifications to all subscribers:

```yaml {filename="config.yaml"}
notifiers:
  - type: pushbullet
    access_token: ${UPDATE_WATCHER_PUSHBULLET_TOKEN}
    channel_tag: "server-updates"
```

## Notification Format

{{< callout type="info" >}}
Pushbullet notifications are delivered as "note" type pushes:

- **Title** -- Server hostname and update count (e.g., "12 updates available on web-prod-01").
- **Body** -- A text summary listing each checker with available updates, package names, and version information. Security updates are labeled distinctly.

The notification appears as a standard push notification on all configured devices, with the full body available when expanded.
{{< /callout >}}

## Device Targeting

Pushbullet offers flexible targeting:

- **All devices** (default) -- When neither `device_iden` nor `channel_tag` is set, the notification is sent to all devices linked to the account.
- **Specific device** -- Set `device_iden` to target a single device (e.g., your phone but not your laptop).
- **Channel** -- Set `channel_tag` to broadcast to all subscribers of a Pushbullet channel. This is useful for teams where multiple people want to subscribe to update notifications.

## Testing

Run Update-Watcher to verify the Pushbullet notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Pushbullet app or browser extension on your devices. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` with a JSON body containing the push details. Common errors include an invalid access token (`401 Unauthorized`) or an invalid device identifier.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
