---
title: "Gotify Notifications for Server Updates - Self-Hosted Push Notifications"
description: "Send server update notifications via self-hosted Gotify server. Full control over your notification infrastructure. Simple token-based authentication."
weight: 8
---

Update-Watcher sends push notifications through [Gotify](https://gotify.net), a self-hosted push notification server. Gotify gives you complete control over your notification infrastructure -- no third-party services, no subscription fees, and no data leaving your network. It is an excellent choice for privacy-conscious administrators and air-gapped or restricted environments.

## Setup on Gotify

{{% steps %}}

### Step 1: Deploy a Gotify Server

If you do not already have a Gotify instance running, deploy one using Docker:

```bash {filename="Terminal"}
docker run -d \
  --name gotify \
  -p 8080:80 \
  -v gotify_data:/app/data \
  gotify/server
```

Alternatively, download a binary release from [gotify.net/docs/install](https://gotify.net/docs/install) or deploy with your preferred orchestration tool.

### Step 2: Create an Application

1. Open the Gotify web interface (e.g., `http://your-server:8080`).
2. Log in with the default credentials (admin/admin) and change the password.
3. Go to **Apps** and click **Create Application**.
4. Enter a name (e.g., "Update Watcher") and optionally a description.
5. Copy the **application token** displayed after creation.

### Step 3: Install the Client

Install the Gotify client on your devices to receive push notifications:

- **Android** -- Available on [F-Droid](https://f-droid.org/packages/com.github.gotify/) and [Google Play](https://play.google.com/store/apps/details?id=com.github.gotify).
- **Web** -- The Gotify web interface itself shows notifications in real time.

### Step 4: Store Credentials

Store the credentials in an environment variable:

```bash {filename="Terminal"}
export UPDATE_WATCHER_GOTIFY_TOKEN="<your-gotify-app-token>"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the application token in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `server_url` | Yes | -- | Base URL of your Gotify server (e.g., `https://gotify.example.com`). |
| `token` | Yes | -- | Gotify application token. Use an environment variable reference. |
| `priority` | No | `5` | Message priority from `0` (minimum) to `10` (maximum). Higher values may trigger more prominent notifications on the client. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: gotify
    server_url: "https://gotify.example.com"
    token: ${UPDATE_WATCHER_GOTIFY_TOKEN}
    priority: 5
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: gotify
    server_url: "https://gotify.example.com"
    token: ${UPDATE_WATCHER_GOTIFY_TOKEN}
```

### High Priority for Urgent Notifications

```yaml {filename="config.yaml"}
notifiers:
  - type: gotify
    server_url: "https://gotify.example.com"
    token: ${UPDATE_WATCHER_GOTIFY_TOKEN}
    priority: 8
```

## Notification Format

{{< callout type="info" >}}
Gotify notifications include:

- **Title** -- Server hostname and update count summary.
- **Message body** -- A Markdown-formatted listing of each checker with available updates, showing package names and version details. Security updates are highlighted.
- **Priority** -- Mapped to Gotify's priority system, which the Android client uses to determine notification behavior (vibration, sound, heads-up display).
{{< /callout >}}

## Self-Hosting Considerations

Gotify is lightweight and runs comfortably on minimal hardware:

- **Resource usage** -- Gotify typically uses less than 30 MB of RAM and negligible CPU.
- **Reverse proxy** -- For production use, place Gotify behind a reverse proxy (Nginx, Caddy, Traefik) with TLS termination.
- **Persistence** -- Mount a volume for `/app/data` to persist messages and configuration across container restarts.
- **Network** -- If your Gotify server is behind a firewall, ensure the machine running Update-Watcher can reach it on the configured port.

## Testing

Run Update-Watcher to verify the Gotify notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Gotify web interface or Android app for the notification. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` with a JSON body containing the message ID. If you receive a `401 Unauthorized`, verify the application token. If the connection is refused, ensure the Gotify server is running and reachable.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
