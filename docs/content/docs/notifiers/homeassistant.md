---
title: "Home Assistant Notifications for Server Updates - Smart Home Integration"
description: "Send server update notifications to Home Assistant via its notify service. Integrate Linux and macOS update alerts into your smart home automation dashboard."
weight: 9
---

Update-Watcher sends server update notifications to [Home Assistant](https://www.home-assistant.io) through its REST API and notify service. This integration allows you to display update alerts on your Home Assistant dashboard, trigger automations based on available updates, or forward notifications to any device connected to your smart home setup. It bridges server administration with home automation for administrators who already run Home Assistant.

## Setup on Home Assistant

{{% steps %}}

### Step 1: Create a Long-Lived Access Token

1. Open your Home Assistant web interface.
2. Click your profile icon in the bottom-left corner of the sidebar.
3. Scroll down to the **Long-Lived Access Tokens** section.
4. Click **Create Token**.
5. Give it a name (e.g., "Update Watcher") and click **OK**.
6. Copy the token immediately -- it is only shown once.

### Step 2: Store the Token

Store the token in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_HA_TOKEN="<your-home-assistant-long-lived-access-token>"
```

### Step 3: Verify the Notify Service

By default, Update-Watcher sends notifications through the `notify` service, which uses whatever notification platform you have configured as the default in Home Assistant (e.g., the mobile app, persistent notifications, or a custom notify target).

To check available notify services, go to **Developer Tools** in Home Assistant, select the **Services** tab, and search for services starting with `notify.`.

{{% /steps %}}

{{< callout type="warning" >}}
Store the Long-Lived Access Token in an environment variable rather than placing it directly in your configuration file. The token is only shown once when created.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `url` | Yes | -- | Base URL of your Home Assistant instance (e.g., `http://homeassistant.local:8123` or `https://ha.example.com`). |
| `token` | Yes | -- | Long-Lived Access Token for authentication. Use an environment variable reference. |
| `service` | No | `notify` | The Home Assistant notify service to call. Use the service name without the `notify.` prefix for alternate targets (e.g., `mobile_app_my_phone` to use `notify.mobile_app_my_phone`). |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: homeassistant
    url: "http://homeassistant.local:8123"
    token: ${UPDATE_WATCHER_HA_TOKEN}
    service: "notify"
```

### Send to a Specific Mobile Device

If you have the Home Assistant Companion App installed, you can target a specific device:

```yaml {filename="config.yaml"}
notifiers:
  - type: homeassistant
    url: "http://homeassistant.local:8123"
    token: ${UPDATE_WATCHER_HA_TOKEN}
    service: "mobile_app_my_phone"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: homeassistant
    url: "http://homeassistant.local:8123"
    token: ${UPDATE_WATCHER_HA_TOKEN}
```

## Notification Format

{{< callout type="info" >}}
Update-Watcher sends a service call to Home Assistant's notify service with:

- **Title** -- Server hostname and update count.
- **Message** -- A text summary listing each checker with available updates, package names, and versions. Security updates are labeled.

The exact rendering depends on the notification platform configured in Home Assistant. For the mobile companion app, it appears as a standard push notification. For persistent notifications, it appears in the Home Assistant notifications panel.
{{< /callout >}}

## Automation Ideas

Once Update-Watcher posts notifications to Home Assistant, you can build automations around them:

- **Dashboard card** -- Display the latest update status on a Lovelace dashboard using a Markdown card or the notification panel.
- **Colored lights** -- Trigger a scene that changes a smart light to red when security updates are available.
- **TTS announcement** -- Use a text-to-speech service to announce updates through a smart speaker.
- **Conditional forwarding** -- Create an automation that forwards the notification to a different platform (e.g., Telegram) only when security updates are detected.

## Network Considerations

The machine running Update-Watcher must be able to reach your Home Assistant instance over the network:

- **Local network** -- Use the local hostname or IP (e.g., `http://homeassistant.local:8123` or `http://192.168.1.100:8123`).
- **Remote access** -- If Update-Watcher runs on a remote server, use the external URL and ensure HTTPS is configured.
- **Nabu Casa** -- If using Home Assistant Cloud, the remote URL works but introduces an external dependency.

## Testing

Run Update-Watcher to verify the Home Assistant notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Home Assistant notifications panel or the targeted device for the notification. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` from the Home Assistant API. Common errors include an invalid or expired token (`401 Unauthorized`) or an unreachable Home Assistant instance.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
