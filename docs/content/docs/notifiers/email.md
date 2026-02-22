---
title: "Email Notifications for Server Updates - SMTP / HTML Email Alerts"
description: "Receive server update notifications via HTML email through any SMTP server. STARTTLS support, multiple recipients, professional formatting. Easy SMTP setup."
weight: 5
---

Update-Watcher sends server update notifications as HTML-formatted email through any standard SMTP server. Messages are professionally formatted with clear structure, security update highlighting, and compatibility across all major email clients. Multiple recipients are supported, making this notifier ideal for teams that prefer inbox-based workflows or need an audit trail of update reports.

## Setup for Email

Update-Watcher works with any SMTP server that supports authentication. Common options include:

- **Existing mail server** -- Use the SMTP credentials from your organization's mail infrastructure.
- **Gmail** -- Create an [App Password](https://myaccount.google.com/apppasswords) (requires 2FA enabled on the account). Use `smtp.gmail.com` on port `587`.
- **Outlook / Microsoft 365** -- Create an App Password or use the SMTP relay. Use `smtp.office365.com` on port `587`.
- **Amazon SES** -- Use the SMTP credentials from the SES console. Use the regional SES endpoint on port `587`.
- **Self-hosted** -- Postfix, Mailcow, or any SMTP server on your network.

{{< tabs items="Gmail,Custom SMTP" >}}

{{< tab >}}

{{% steps %}}

### Step 1: Generate an App Password

Go to [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords).

### Step 2: Select App and Device

Select **Mail** and your device, then click **Generate**.

### Step 3: Copy the Password

Copy the 16-character app password.

{{% /steps %}}

Store the credentials in environment variables:

```bash {filename="Terminal"}
export UPDATE_WATCHER_SMTP_USER="alerts@example.com"
export UPDATE_WATCHER_SMTP_PASS="xxxx-xxxx-xxxx-xxxx"
```

```yaml {filename="config.yaml"}
notifiers:
  - type: email
    smtp_host: smtp.gmail.com
    smtp_port: 587
    username: ${UPDATE_WATCHER_SMTP_USER}
    password: ${UPDATE_WATCHER_SMTP_PASS}
    from: "alerts@example.com"
    to:
      - "admin@example.com"
      - "ops-team@example.com"
    tls: true
```

{{< /tab >}}

{{< tab >}}

{{% steps %}}

### Step 1: Gather SMTP Credentials

Obtain the SMTP host, port, username, and password from your email provider or mail server administrator.

### Step 2: Store Credentials Securely

Store the credentials in environment variables:

```bash {filename="Terminal"}
export UPDATE_WATCHER_SMTP_USER="alerts@example.com"
export UPDATE_WATCHER_SMTP_PASS="your-smtp-password"
```

{{% /steps %}}

```yaml {filename="config.yaml"}
notifiers:
  - type: email
    smtp_host: mail.example.com
    smtp_port: 587
    username: ${UPDATE_WATCHER_SMTP_USER}
    password: ${UPDATE_WATCHER_SMTP_PASS}
    from: "alerts@example.com"
    to:
      - "admin@example.com"
    tls: true
```

### Local Relay Without Authentication

If you have a local mail relay (e.g., Postfix configured as a relay host on the same machine), you may not need authentication or TLS:

```yaml {filename="config.yaml"}
notifiers:
  - type: email
    smtp_host: localhost
    smtp_port: 25
    username: ""
    password: ""
    from: "update-watcher@myserver.local"
    to:
      - "admin@example.com"
    tls: false
```

{{< /tab >}}

{{< /tabs >}}

{{< callout type="warning" >}}
Store SMTP credentials in environment variables rather than placing them directly in your configuration file.
{{< /callout >}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `smtp_host` | Yes | -- | Hostname or IP address of the SMTP server (e.g., `smtp.gmail.com`). |
| `smtp_port` | No | `587` | SMTP port number. Use `587` for STARTTLS (recommended), `465` for implicit TLS, or `25` for unencrypted relay. |
| `username` | Yes | -- | SMTP authentication username. Typically the email address. Use an environment variable reference. |
| `password` | Yes | -- | SMTP authentication password or app password. Use an environment variable reference. |
| `from` | Yes | -- | Sender email address displayed in the From header. |
| `to` | Yes | -- | List of recipient email addresses. Can be a single address or multiple. |
| `tls` | No | `true` | Whether to use STARTTLS to encrypt the SMTP connection. Disable only for local relay servers that do not support TLS. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: email
    smtp_host: smtp.gmail.com
    smtp_port: 587
    username: ${UPDATE_WATCHER_SMTP_USER}
    password: ${UPDATE_WATCHER_SMTP_PASS}
    from: "alerts@example.com"
    to:
      - "admin@example.com"
      - "ops-team@example.com"
    tls: true
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: email
    smtp_host: smtp.gmail.com
    username: ${UPDATE_WATCHER_SMTP_USER}
    password: ${UPDATE_WATCHER_SMTP_PASS}
    from: "alerts@example.com"
    to:
      - "admin@example.com"
```

## Email Format

{{< callout type="info" >}}
Update-Watcher generates HTML emails that render correctly across major email clients (Gmail, Outlook, Apple Mail, Thunderbird). The email includes:

- **Subject line** -- Contains the server hostname and a count of available updates (e.g., "Update Watcher: 12 updates available on web-prod-01").
- **Body** -- A structured HTML layout with sections for each checker, listing package names, current versions, and available versions in a formatted table.
- **Security highlights** -- Security updates are displayed in a distinct section with visual emphasis.
- **Plain text fallback** -- A plain text version is included as a multipart alternative for clients that do not render HTML.
{{< /callout >}}

## Testing

Run Update-Watcher to verify email delivery:

```bash {filename="Terminal"}
update-watcher run
```

Check the inbox of the configured recipients. Also check spam/junk folders, especially on the first delivery. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a notification.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

Verbose output shows the SMTP connection process, including TLS negotiation and authentication. Common issues include:

- **Authentication failed** -- Verify username and password. For Gmail, ensure you are using an App Password, not your regular account password.
- **Connection refused** -- Verify the SMTP host and port. Ensure your server's firewall allows outbound connections on the SMTP port.
- **Certificate errors** -- If using a self-signed certificate on a local SMTP server, you may need to set `tls: false` for testing.

{{< callout emoji="💡" >}}
**Testing tip:** If no updates are available on your system, temporarily set `send_policy: "always"` in your configuration to force a notification. Also check your spam/junk folder on the first delivery.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
