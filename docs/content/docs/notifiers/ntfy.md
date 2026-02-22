---
title: "ntfy Notifications for Server Updates - Push Notifications via ntfy.sh"
description: "Send server update push notifications via ntfy.sh or self-hosted ntfy server. Open-source push notifications for Linux update monitoring. Topic-based setup."
weight: 6
---

Update-Watcher sends push notifications through [ntfy](https://ntfy.sh), an open-source push notification service. ntfy uses a simple topic-based pub/sub model -- you publish to a topic, and any device subscribed to that topic receives the notification instantly. You can use the free public server at ntfy.sh or host your own ntfy instance for full control over your notification infrastructure.

## Setup for ntfy

{{< tabs items="ntfy.sh (Public),Self-hosted" >}}

{{< tab >}}

{{% steps %}}

### Step 1: Choose a Topic Name

Choose a unique topic name. Topics are public by default on ntfy.sh, so use something unguessable (e.g., `update-watcher-a7f3b2c1` rather than `updates`).

### Step 2: Subscribe on Your Devices

Subscribe to the topic on your phone by installing the [ntfy app](https://ntfy.sh/docs/subscribe/phone/) (available on F-Droid and Google Play) and adding your topic.

### Step 3: Optional Web Subscription

Optionally, subscribe in a web browser at `https://ntfy.sh/your-topic-name`.

{{% /steps %}}

No account or API key is required for the public server. However, if you want access control, you can create a free account on ntfy.sh and use token-based authentication.

```yaml {filename="config.yaml"}
notifiers:
  - type: ntfy
    topic: "update-watcher-a7f3b2c1"
    priority: "default"
```

{{< /tab >}}

{{< tab >}}

{{% steps %}}

### Step 1: Deploy ntfy

Deploy ntfy following the [self-hosting guide](https://docs.ntfy.sh/install/).

### Step 2: Configure Authentication

Create a user and access token if you have authentication enabled.

### Step 3: Choose a Topic

Choose a topic name for your Update-Watcher notifications.

### Step 4: Subscribe on Devices

Subscribe on your devices using your server URL and topic.

{{% /steps %}}

{{< callout type="warning" >}}
Store the access token in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

```bash {filename="Terminal"}
export UPDATE_WATCHER_NTFY_TOKEN="tk_your_access_token_here"
```

```yaml {filename="config.yaml"}
notifiers:
  - type: ntfy
    topic: "server-updates"
    server_url: "https://ntfy.example.com"
    token: ${UPDATE_WATCHER_NTFY_TOKEN}
    priority: "high"
```

{{< /tab >}}

{{< /tabs >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `topic` | Yes | -- | The ntfy topic name to publish to. Choose something unique and unguessable for public servers. |
| `server_url` | No | `https://ntfy.sh` | Base URL of the ntfy server. Change this if you are self-hosting ntfy. |
| `token` | No | -- | Access token for authenticated ntfy servers. Not required for the public ntfy.sh server without access control. Use an environment variable reference. |
| `priority` | No | -- | Notification priority level. Values: `min`, `low`, `default`, `high`, `urgent`. When security updates are detected, the priority may be automatically elevated. |

## Configuration Example

### Public ntfy.sh Server

```yaml {filename="config.yaml"}
notifiers:
  - type: ntfy
    topic: "update-watcher-a7f3b2c1"
    priority: "default"
```

### Self-Hosted ntfy with Authentication

```yaml {filename="config.yaml"}
notifiers:
  - type: ntfy
    topic: "server-updates"
    server_url: "https://ntfy.example.com"
    token: ${UPDATE_WATCHER_NTFY_TOKEN}
    priority: "high"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: ntfy
    topic: "update-watcher-a7f3b2c1"
```

## Notification Format

{{< callout type="info" >}}
ntfy notifications include:

- **Title** -- Server hostname and update count.
- **Body** -- A text summary listing each checker and its available updates, with package names and versions.
- **Priority** -- Maps to ntfy's priority system, which controls notification sound and display behavior on mobile devices.
- **Tags** -- Update-Watcher sets appropriate tags that ntfy renders as emoji indicators in the notification.

On mobile, ntfy notifications appear as standard push notifications with the configured priority level controlling the alert sound and vibration pattern.
{{< /callout >}}

## Topic Security

On the public ntfy.sh server, anyone who knows a topic name can subscribe to it. To protect your update notifications:

- **Use a long, random topic name** -- Treat the topic name like a password. A UUID or random string is ideal.
- **Self-host ntfy** -- Run your own server and configure authentication for full access control.
- **Use access tokens** -- On ntfy.sh or self-hosted instances with access control enabled, use token-based authentication to restrict who can publish and subscribe.

## Testing

Run Update-Watcher to verify the ntfy notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the ntfy app on your phone or the web interface for the notification. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` from the ntfy server. If you receive a `401 Unauthorized`, verify your access token. If using a self-hosted server, ensure it is reachable from the machine running Update-Watcher.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
