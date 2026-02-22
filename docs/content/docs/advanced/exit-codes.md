---
title: "Exit Codes - Scripting Reference"
description: "Update-Watcher exit code reference for shell scripting. Exit codes 0-4 indicate success, updates found, partial/complete failure, and config errors."
weight: 2
---

Update-Watcher uses specific exit codes to communicate the outcome of each run. These codes are designed for shell scripting and automation, allowing you to take different actions based on whether updates were found, checkers failed, or the configuration is invalid.

## Exit Code Table

{{< callout type="info" >}}
Exit codes follow a severity hierarchy: 0 is clean, 1 means actionable updates, and 2+ indicates problems that may need investigation.
{{< /callout >}}

| Code | Name | Description |
|------|------|-------------|
| 0 | No updates | All configured checkers ran successfully. No updates are available. |
| 1 | Updates found | At least one checker reported available updates. All checkers ran successfully. |
| 2 | Partial failure | Some checkers succeeded but at least one failed. Updates may or may not have been found by the successful checkers. |
| 3 | Complete failure | All configured checkers failed. No update information is available. |
| 4 | Config error | The configuration file is missing, contains syntax errors, or has invalid values. No checkers were run. |

## Bash Script Examples

{{< tabs items="Basic,With Notifications,Security Only" >}}
{{< tab >}}

### Basic Exit Code Check

```bash {filename="Terminal"}
#!/bin/bash
update-watcher run --quiet
EXIT_CODE=$?

case $EXIT_CODE in
  0)
    echo "All systems up to date."
    ;;
  1)
    echo "Updates are available."
    ;;
  2)
    echo "Warning: some checkers failed."
    ;;
  3)
    echo "Error: all checkers failed."
    ;;
  4)
    echo "Configuration error. Run: update-watcher validate"
    ;;
esac
```

### Simple Update Detection

Use the exit code directly in a conditional:

```bash {filename="Terminal"}
#!/bin/bash
if update-watcher run --quiet --notify=false; then
  echo "No updates - nothing to do."
else
  echo "Updates found or errors occurred (exit code: $?)."
fi
```

{{< callout type="info" >}}
In Bash, `if command; then` succeeds only when the exit code is 0. Any non-zero exit code (updates found, failures, or config errors) takes the `else` branch.
{{< /callout >}}

### Distinguish Updates from Errors

```bash {filename="Terminal"}
#!/bin/bash
update-watcher run --quiet --notify=false
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  echo "No updates."
elif [ $EXIT_CODE -eq 1 ]; then
  echo "Updates available - review and apply."
elif [ $EXIT_CODE -ge 2 ]; then
  echo "Error detected (code $EXIT_CODE) - investigate."
  exit 1
fi
```

{{< /tab >}}
{{< tab >}}

### Monitoring Integration

Map exit codes to monitoring system severity levels:

```bash {filename="Terminal"}
#!/bin/bash
update-watcher run --quiet --notify=false
EXIT_CODE=$?

case $EXIT_CODE in
  0)
    echo "OK"
    ;;
  1)
    echo "WARNING - Updates available"
    ;;
  2)
    echo "WARNING - Partial checker failure"
    ;;
  3)
    echo "CRITICAL - All checkers failed"
    ;;
  4)
    echo "CRITICAL - Configuration error"
    ;;
esac

exit $EXIT_CODE
```

{{< callout emoji="💡" >}}
This pattern is compatible with Nagios-style monitoring plugins (0=OK, 1=WARNING, 2+=CRITICAL).
{{< /callout >}}

### Automated Update Workflow

Trigger downstream actions based on the exit code:

```bash {filename="Terminal"}
#!/bin/bash
update-watcher run --quiet
EXIT_CODE=$?

if [ $EXIT_CODE -eq 1 ]; then
  # Updates found - trigger a maintenance workflow
  echo "Updates detected on $(hostname) at $(date)" >> /var/log/update-alerts.log

  # Optionally trigger a maintenance window
  # /usr/local/bin/schedule-maintenance.sh
fi

if [ $EXIT_CODE -ge 3 ]; then
  # Critical failure - alert the on-call team
  echo "Update-Watcher failure on $(hostname)" | \
    mail -s "CRITICAL: Update-Watcher failure" oncall@example.com
fi
```

{{< /tab >}}
{{< tab >}}

### Combining with JSON Output

Exit codes work alongside JSON output for rich scripting:

```bash {filename="Terminal"}
#!/bin/bash
REPORT=$(update-watcher run --format json --notify=false)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 1 ]; then
  TOTAL=$(echo "$REPORT" | jq '.total_updates')
  SECURITY=$(echo "$REPORT" | jq '.total_security_updates')
  echo "Found $TOTAL updates ($SECURITY security) on $(hostname)"
fi
```

For more JSON examples, see [JSON Output](../json-output/).

### Security-Only Alerting

Combine JSON parsing with exit codes to alert only on security updates:

```bash {filename="Terminal"}
#!/bin/bash
REPORT=$(update-watcher run --format json --notify=false)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 1 ]; then
  SECURITY=$(echo "$REPORT" | jq '.total_security_updates')
  if [ "$SECURITY" -gt 0 ]; then
    echo "SECURITY: $SECURITY security updates on $(hostname)"
    # Trigger high-priority alert
  fi
fi
```

{{< /tab >}}
{{< /tabs >}}

## Exit Codes in CI/CD

Use exit codes in CI/CD pipelines to gate deployments on update status:

```yaml {filename=".github/workflows/check-updates.yml"}
# GitHub Actions example
- name: Check for security updates
  run: |
    update-watcher run --quiet --notify=false
    if [ $? -eq 1 ]; then
      echo "::warning::Updates available on the target server"
    fi
```

## Related

- [JSON Output](../json-output/) -- Machine-readable output for scripting.
- [run](../../cli/run/) -- Full `run` command reference with all flags.
- [Send Policy](../send-policy/) -- Control notification delivery independently of exit codes.
