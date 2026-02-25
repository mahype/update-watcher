---
title: "Update Wall Dashboard Notifier for Server Update Monitoring"
description: "Send structured server update reports to a self-hosted Update Wall dashboard. Track update history, compare servers, and filter by package type or priority."
weight: 16
---

Update-Watcher sends structured JSON reports to [Update Wall](https://github.com/mahype/update-wall), a self-hosted dashboard for monitoring server updates. Unlike chat or push notifiers that deliver human-readable messages, Update Wall receives a rich, machine-readable payload -- hostname, checker results, package names, current and new versions, update type, and priority -- and stores it persistently for historical tracking, multi-server comparison, and filtering.

## Setup

{{% steps %}}

### Step 1: Deploy Update Wall

Deploy Update Wall on a server or machine reachable from all servers running Update-Watcher. Follow the [Update Wall installation guide](https://github.com/mahype/update-wall) and log in as admin.

### Step 2: Create an API Token

In the Update Wall admin panel, navigate to **API Tokens** and create a new token. Copy it immediately -- it is shown only once.

Store the token in an environment variable rather than placing it directly in your configuration file:

```bash {filename="Terminal"}
export UPDATE_WATCHER_UPDATEWALL_TOKEN="<your-updatewall-token>"
```

### Step 3: Configure the Notifier

Add the `updatewall` notifier to your `config.yaml`:

```yaml {filename="config.yaml"}
notifiers:
  - type: updatewall
    url: "https://your-updatewall.example.com/api/v1/report"
    api_token: ${UPDATE_WATCHER_UPDATEWALL_TOKEN}
```

### Step 4: Enable Heartbeat Monitoring (Optional)

To use Update Wall as a heartbeat monitor -- confirming that Update-Watcher ran, even when no updates are available -- set `send_policy: "always"` at the top level of your configuration:

```yaml {filename="config.yaml"}
send_policy: "always"

notifiers:
  - type: updatewall
    url: "https://your-updatewall.example.com/api/v1/report"
    api_token: ${UPDATE_WATCHER_UPDATEWALL_TOKEN}
```

{{% /steps %}}

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `url` | Yes | -- | Full API endpoint URL of your Update Wall instance, e.g. `https://your-domain.example.com/api/v1/report`. |
| `api_token` | Yes | -- | Bearer token created in the Update Wall admin panel. Use an environment variable reference to avoid storing the token in plain text. |

## What Update Wall Receives

Each run sends a single JSON POST to `url` with the following structure:

- **`hostname`** -- The server name that ran Update-Watcher.
- **`timestamp`** -- RFC 3339 UTC timestamp of when the check ran.
- **`total_updates`** -- Aggregate count of available updates across all checkers.
- **`has_security`** -- `true` if any checker found security updates.
- **`checkers`** -- Array of per-checker entries, each containing:
  - `name` -- Checker identifier (e.g. `apt`, `docker`, `wordpress`).
  - `summary` -- Human-readable summary line.
  - `error` -- Error message if the checker failed (omitted on success).
  - `updates` -- Array of individual updates, each with `name`, current and new version, `type` (e.g. `security`, `regular`, `plugin`), and `priority` (e.g. `critical`, `high`, `normal`).

The Update Wall dashboard uses these fields to populate its per-checker expandable tables, making it possible to drill into exactly which packages are pending on each server.

## Configuration Example

```yaml {filename="config.yaml"}
send_policy: "always"

checkers:
  - type: apt
  - type: docker

notifiers:
  - type: updatewall
    url: "https://updatewall.example.com/api/v1/report"
    api_token: ${UPDATE_WATCHER_UPDATEWALL_TOKEN}
```

## Testing

Run Update-Watcher to send a report to Update Wall:

```bash {filename="Terminal"}
update-watcher run
```

Open the Update Wall dashboard and verify that your server appears with the correct checker data. Click on a checker row to expand the package table.

If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a report and confirm the integration is working.

For troubleshooting, run with verbose output to see the HTTP response:

```bash {filename="Terminal"}
update-watcher run --verbose
```

A successful delivery returns an HTTP `2xx` response from the Update Wall server. If you receive `401 Unauthorized`, verify that the `api_token` value matches the token shown in the Update Wall admin panel.

{{< callout emoji="💡" >}}
**Testing tip:** Set `send_policy: "always"` while testing so Update-Watcher always sends a report, even when all packages are up to date.
{{< /callout >}}

## Related

Receive update reports for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
