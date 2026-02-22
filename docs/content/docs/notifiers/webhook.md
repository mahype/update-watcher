---
title: "Webhook Notifications for Server Updates - JSON Payloads to Any HTTP Endpoint"
description: "Send server update data as JSON payloads to any HTTP endpoint. Custom headers, authentication, configurable HTTP methods. Build custom update monitoring workflows."
weight: 16
---

Update-Watcher sends server update data as structured JSON payloads to any HTTP endpoint using its generic webhook notifier. This is the most flexible notifier -- it allows you to integrate Update-Watcher with any system that can receive HTTP requests, including custom dashboards, serverless functions, logging pipelines, ITSM tools, CI/CD systems, or any service not directly supported by the built-in notifiers.

## Setup for Webhooks

The webhook notifier works with any HTTP endpoint that accepts JSON payloads. Common targets include:

- **Custom API endpoints** -- Your own web server or API gateway.
- **Serverless functions** -- AWS Lambda (via API Gateway), Google Cloud Functions, Azure Functions, Cloudflare Workers.
- **Automation platforms** -- Zapier, Make (Integromat), n8n, or Node-RED webhooks.
- **Logging services** -- Elasticsearch, Loki, or any log ingestion endpoint that accepts JSON.
- **ITSM tools** -- ServiceNow, Jira Service Management, or other ticketing systems with webhook receivers.

No special setup is required on the Update-Watcher side beyond configuring the target URL and optional authentication.

## Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `url` | Yes | -- | The target HTTP endpoint URL. Use HTTPS for production endpoints. |
| `method` | No | `POST` | HTTP method to use. Typically `POST` or `PUT`. |
| `content_type` | No | `application/json` | Value for the `Content-Type` header. Change this if your endpoint expects a different content type. |
| `auth_header` | No | -- | Value for the `Authorization` header. Use an environment variable reference for tokens (e.g., `Bearer ${WEBHOOK_TOKEN}`). |
| `headers` | No | -- | Additional HTTP headers as a key-value map. Useful for API keys, custom routing headers, or correlation IDs. |

## Configuration Example

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://api.example.com/update-watcher/ingest"
    method: "POST"
    content_type: "application/json"
    auth_header: "Bearer ${UPDATE_WATCHER_WEBHOOK_TOKEN}"
    headers:
      X-Source: "update-watcher"
      X-Environment: "production"
```

### Minimal Configuration

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://api.example.com/update-watcher/ingest"
```

### With API Key Header

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://api.example.com/updates"
    headers:
      X-API-Key: ${UPDATE_WATCHER_API_KEY}
```

### Serverless Function

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://abc123.execute-api.us-east-1.amazonaws.com/prod/updates"
    auth_header: "Bearer ${AWS_LAMBDA_AUTH_TOKEN}"
```

{{< callout type="warning" >}}
Store authentication tokens and API keys in environment variables rather than placing them directly in your configuration file.
{{< /callout >}}

## JSON Payload Structure

Update-Watcher sends a JSON payload with the following structure:

```json {filename="Output"}
{
  "hostname": "web-prod-01",
  "timestamp": "2026-02-20T07:00:00Z",
  "total_updates": 12,
  "has_security_updates": true,
  "checkers": [
    {
      "name": "apt",
      "updates": [
        {
          "package": "libssl3",
          "current_version": "3.0.13-1ubuntu3.3",
          "new_version": "3.0.13-1ubuntu3.4",
          "security": true
        },
        {
          "package": "nginx",
          "current_version": "1.24.0-1ubuntu1",
          "new_version": "1.24.0-1ubuntu1.1",
          "security": false
        }
      ]
    },
    {
      "name": "docker",
      "updates": [
        {
          "package": "nginx:1.25",
          "current_version": "sha256:abc123...",
          "new_version": "sha256:def456...",
          "security": false
        }
      ]
    }
  ]
}
```

### Payload Fields

| Field | Type | Description |
|-------|------|-------------|
| `hostname` | string | The configured hostname of the server running Update-Watcher. |
| `timestamp` | string | ISO 8601 timestamp of when the check was performed. |
| `total_updates` | integer | Total number of available updates across all checkers. |
| `has_security_updates` | boolean | Whether any of the available updates are classified as security updates. |
| `checkers` | array | Array of checker results, each containing the checker name and its list of available updates. |
| `checkers[].name` | string | Name of the checker (e.g., `apt`, `docker`, `wordpress`). |
| `checkers[].updates` | array | Array of individual update objects. |
| `checkers[].updates[].package` | string | Package or image name. |
| `checkers[].updates[].current_version` | string | Currently installed version. |
| `checkers[].updates[].new_version` | string | Available version. |
| `checkers[].updates[].security` | boolean | Whether this update is classified as a security update. |

## Integration Examples

{{< details title="Example: Store in a Database" >}}

Point the webhook at an API endpoint that inserts the payload into a database for historical tracking and reporting:

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://internal-api.example.com/v1/server-updates"
    auth_header: "Bearer ${INTERNAL_API_TOKEN}"
```

{{< /details >}}

{{< details title="Example: Trigger a CI/CD Pipeline" >}}

Use the webhook to trigger an automated patching pipeline in your CI/CD system:

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://ci.example.com/api/v4/projects/42/trigger/pipeline"
    method: "POST"
    headers:
      PRIVATE-TOKEN: ${GITLAB_TRIGGER_TOKEN}
```

{{< /details >}}

{{< details title="Example: Forward to Another Monitoring Tool" >}}

Send update data to a monitoring tool that does not have a dedicated Update-Watcher notifier:

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://monitoring.example.com/api/events"
    auth_header: "Token ${MONITORING_TOKEN}"
    headers:
      X-Event-Type: "server-updates"
```

{{< /details >}}

## Testing

Run Update-Watcher to verify the webhook delivery:

```bash {filename="Terminal"}
update-watcher run
```

Check your endpoint's logs or dashboard for the received payload. If no updates are available and `send_policy` is `only-on-updates`, temporarily set `send_policy: "always"` to force a delivery.

For troubleshooting:

```bash {filename="Terminal"}
update-watcher run --verbose
```

Verbose output shows the full HTTP request and response, including status code, headers, and body. For local testing, you can use a tool like [webhook.site](https://webhook.site) or a simple local HTTP server to inspect the payload:

```yaml {filename="config.yaml"}
notifiers:
  - type: webhook
    url: "https://webhook.site/your-unique-id"
```

A successful delivery depends on your endpoint's expected response. Update-Watcher considers any `2xx` HTTP status code a success.

{{< callout emoji="💡" >}}
**Testing tip:** For local testing without a real endpoint, use [webhook.site](https://webhook.site) to inspect the exact JSON payload Update-Watcher sends. Set `send_policy: "always"` to force a delivery even when no updates are available.
{{< /callout >}}

## Related

Receive update notifications for [APT](/docs/checkers/apt), [Docker](/docs/checkers/docker), [WordPress](/docs/checkers/wordpress), and [11 more checkers](/docs/checkers/).
