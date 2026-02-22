---
title: "JSON Output - Machine-Readable Update Reports"
description: "Use Update-Watcher JSON output for scripting and automation. Parse results with jq, integrate with monitoring tools, and build custom workflows."
weight: 1
---

Update-Watcher can output check results and configuration status as structured JSON. This enables integration with monitoring tools, custom notification pipelines, dashboards, and scripting workflows.

## Enabling JSON Output

### Run Results

Add `--format json` to the `run` command:

```bash {filename="Terminal"}
update-watcher run --format json
```

The JSON output contains the full check results: each checker's status, the list of available updates, error messages for failed checkers, and summary totals.

### Configuration Status

The `status` command also supports JSON:

```bash {filename="Terminal"}
update-watcher status --format json
```

This outputs the resolved configuration as JSON, including all watchers, notifiers, settings, and cron status.

## JSON Structure

The `run` command produces output with the following structure:

```json {filename="output.json"}
{
  "hostname": "web-prod-01",
  "timestamp": "2025-04-15T07:00:12Z",
  "total_updates": 5,
  "total_security_updates": 2,
  "checkers": [
    {
      "type": "apt",
      "status": "success",
      "updates": 3,
      "security_updates": 2,
      "packages": [
        {
          "name": "openssl",
          "current_version": "3.0.2-0ubuntu1.14",
          "available_version": "3.0.2-0ubuntu1.15",
          "security": true
        },
        {
          "name": "libssl3",
          "current_version": "3.0.2-0ubuntu1.14",
          "available_version": "3.0.2-0ubuntu1.15",
          "security": true
        },
        {
          "name": "curl",
          "current_version": "7.81.0-1ubuntu1.15",
          "available_version": "7.81.0-1ubuntu1.16",
          "security": false
        }
      ]
    },
    {
      "type": "docker",
      "status": "success",
      "updates": 2,
      "security_updates": 0,
      "containers": [
        {
          "name": "nginx-proxy",
          "image": "nginx:1.25",
          "update_available": true
        },
        {
          "name": "redis-cache",
          "image": "redis:7.2",
          "update_available": true
        }
      ]
    }
  ]
}
```

## Parsing with jq

{{< callout emoji="💡" >}}
`jq` is the standard tool for processing JSON on the command line. Pipe the JSON output directly to `jq` for filtering, counting, and transforming results.
{{< /callout >}}

{{< details title="Example: Count Total Updates" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq '.total_updates'
```
{{< /details >}}

{{< details title="Example: List Only Security Updates" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq '[.checkers[].packages[]? | select(.security == true)]'
```
{{< /details >}}

{{< details title="Example: Get Checker Types with Updates" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq '[.checkers[] | select(.updates > 0) | .type]'
```
{{< /details >}}

{{< details title="Example: Extract Package Names" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq '[.checkers[].packages[]?.name] | unique'
```
{{< /details >}}

{{< details title="Example: Filter by Checker Type" >}}
Get only APT results:

```bash {filename="Terminal"}
update-watcher run --format json | jq '.checkers[] | select(.type == "apt")'
```
{{< /details >}}

{{< details title="Example: Check for Failed Checkers" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq '[.checkers[] | select(.status != "success") | .type]'
```
{{< /details >}}

{{< details title="Example: Pretty-Print the Full Report" >}}
```bash {filename="Terminal"}
update-watcher run --format json | jq .
```
{{< /details >}}

## Integration Examples

### Write Results to a File

Save the JSON report for later processing:

```bash {filename="Terminal"}
update-watcher run --format json > /var/log/update-watcher-report.json
```

### Send to a Custom Webhook

{{< callout emoji="💡" >}}
Use `--notify=false` when sending to a custom webhook to avoid duplicate notifications through the built-in notifiers.
{{< /callout >}}

Post the JSON report to a custom HTTP endpoint:

```bash {filename="Terminal"}
update-watcher run --format json --notify=false | \
  curl -s -X POST -H "Content-Type: application/json" \
    -d @- https://monitoring.example.com/api/updates
```

### Conditional Alerting Script

Build a custom alerting workflow based on security updates:

```bash {filename="Terminal"}
#!/bin/bash
REPORT=$(update-watcher run --format json --notify=false)
SECURITY_COUNT=$(echo "$REPORT" | jq '.total_security_updates')

if [ "$SECURITY_COUNT" -gt 0 ]; then
  echo "CRITICAL: $SECURITY_COUNT security updates available" | \
    mail -s "Security Updates on $(hostname)" admin@example.com
fi
```

### Feed into Monitoring Systems

Combine JSON output with exit codes for monitoring integration:

```bash {filename="Terminal"}
#!/bin/bash
RESULT=$(update-watcher run --format json --notify=false)
EXIT_CODE=$?

case $EXIT_CODE in
  0) echo "OK - No updates available" ;;
  1) echo "WARNING - $(echo $RESULT | jq -r '.total_updates') updates available" ;;
  2) echo "WARNING - Partial checker failure" ;;
  3) echo "CRITICAL - All checkers failed" ;;
  4) echo "CRITICAL - Configuration error" ;;
esac

exit $EXIT_CODE
```

### Aggregate Across Servers

Collect reports from multiple servers into a central location:

```bash {filename="Terminal"}
for server in web-01 web-02 db-01; do
  ssh "$server" "update-watcher run --format json --notify=false" > "reports/${server}.json"
done

# Summarize total updates across all servers
cat reports/*.json | jq -s '[.[].total_updates] | add'
```

## Combining with Other Flags

JSON output can be combined with other `run` flags:

```bash {filename="Terminal"}
# JSON output for a single checker
update-watcher run --format json --only apt

# JSON output without sending notifications
update-watcher run --format json --notify=false

# Verbose JSON (debug info goes to stderr, JSON to stdout)
update-watcher run --format json --verbose
```

{{< callout type="info" >}}
When `--verbose` is used with `--format json`, the debug output is written to stderr so it does not interfere with the JSON on stdout. This means you can redirect them separately.
{{< /callout >}}

```bash {filename="Terminal"}
update-watcher run --format json --verbose 2>/var/log/debug.log > report.json
```

## Related

- [Exit Codes](../exit-codes/) -- Use exit codes alongside JSON for scripting.
- [run](../../cli/run/) -- Full `run` command reference.
- [status](../../cli/status/) -- JSON output for configuration status.
- [Webhook Notifier](../../notifiers/webhook/) -- Send JSON payloads to custom endpoints natively.
