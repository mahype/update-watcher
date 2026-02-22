---
title: "Matrix Notifications for Server Updates - Decentralized Chat Integration"
description: "Send server update notifications to Matrix rooms via the client-server API. Self-hosted or public homeserver support. Privacy-focused, decentralized messaging."
weight: 11
---

Update-Watcher sends server update notifications to [Matrix](https://matrix.org) rooms using the client-server API. Matrix is a decentralized, open-standard communication protocol that supports self-hosted homeservers, end-to-end encryption, and federation between servers. This notifier is ideal for privacy-conscious teams and organizations that run their own communication infrastructure.

## Setup on Matrix

{{% steps %}}

### Step 1: Create a Bot Account

Create a dedicated Matrix account for Update-Watcher on your homeserver or on a public server like matrix.org:

1. Register a new account (e.g., `@update-watcher:your-homeserver.org`).
2. You can use [Element](https://element.io) or any Matrix client for the initial setup.

### Step 2: Get an Access Token

You need an access token to authenticate API calls. There are several ways to obtain one:

**Via Element (easiest):**

1. Log in to Element with the bot account.
2. Go to **Settings** (click the user avatar) and then **Help & About**.
3. Scroll to the bottom and click **Access Token** (you may need to expand an "Advanced" section).
4. Copy the access token.

**Via the API:**

```bash {filename="Terminal"}
curl -XPOST "https://your-homeserver.org/_matrix/client/r0/login" \
  -H "Content-Type: application/json" \
  -d '{"type":"m.login.password","user":"update-watcher","password":"your-password"}'
```

The response JSON contains the `access_token` field.

### Step 3: Create or Join a Room

1. Create a new room or identify the existing room where notifications should be posted.
2. Invite the bot account to the room and accept the invitation from the bot's session.
3. Copy the **room ID**. Room IDs look like `!abcdef123456:your-homeserver.org`. In Element, find it in room settings under **Advanced**.

### Step 4: Store Credentials

Store the credentials in environment variables:

```bash {filename="Terminal"}
export UPDATE_WATCHER_MATRIX_TOKEN="<your-matrix-access-token>"
export UPDATE_WATCHER_MATRIX_ROOM_ID="<your-matrix-room-id>"
```

{{% /steps %}}

{{< callout type="warning" >}}
Store the access token in an environment variable rather than placing it directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `homeserver` | Yes | -- | Base URL of the Matrix homeserver (e.g., `https://matrix.example.org` or `https://matrix-client.matrix.org`). |
| `access_token` | Yes | -- | Access token for the bot account. Use an environment variable reference. |
| `room_id` | Yes | -- | The Matrix room ID to send notifications to. Starts with `!` and includes the homeserver domain. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: matrix
    homeserver: "https://matrix.example.org"
    access_token: ${UPDATE_WATCHER_MATRIX_TOKEN}
    room_id: ${UPDATE_WATCHER_MATRIX_ROOM_ID}
```

## Message Format

{{< callout type="info" >}}
Matrix notifications are sent as formatted messages with both plain text and HTML bodies:

- **Plain text** -- A clean text representation for clients that do not render HTML.
- **HTML body** -- A structured HTML layout with bold headings for each checker, package lists with version information, and security update labels.
- **Room event** -- The notification is posted as a standard `m.room.message` event of type `m.text` with a `formatted_body`.

Messages render well in Element, FluffyChat, Nheko, and other popular Matrix clients.
{{< /callout >}}

## Self-Hosting Considerations

If you run your own Matrix homeserver (Synapse, Dendrite, Conduit):

- **Local network** -- The machine running Update-Watcher must be able to reach the homeserver's client-server API endpoint.
- **Federation** -- Not required for this integration. Update-Watcher communicates directly with the homeserver via the client-server API.
- **Rate limiting** -- Self-hosted servers may have rate limiting configured. For typical Update-Watcher usage, this is not a concern.
- **Token expiry** -- Access tokens created via the login API may expire depending on your homeserver configuration. Tokens created through Element's settings interface are typically long-lived.

## Testing

Run Update-Watcher to verify the Matrix notification:

```bash {filename="Terminal"}
update-watcher run
```

Check the Matrix room for the message in your preferred client. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns HTTP status `200 OK` with a JSON body containing the `event_id`. Common errors include an invalid access token (`401`), the bot not being a member of the room (`403`), or an unreachable homeserver.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification and verify the integration is working.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
