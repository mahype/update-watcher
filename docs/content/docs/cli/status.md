---
title: "status - Show Configuration Status"
description: "Display the current Update-Watcher configuration including configured watchers, notifiers, and settings."
weight: 5
---

The `status` command displays the current Update-Watcher configuration in a human-readable table or machine-readable JSON format. It shows all configured watchers, notifiers, global settings, and cron status at a glance.

## Usage

```bash {filename="Terminal"}
update-watcher status [--format table|json]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | `table\|json` | `table` | Output format. Use `json` for machine-readable output suitable for scripting. |

## Output

{{< tabs items="Table,JSON" >}}

{{< tab >}}

The default table output provides a quick overview of the configuration:

```text {filename="Output"}
=== Update-Watcher Status ===

Hostname:    web-prod-01
Config File: /etc/update-watcher/config.yaml

Watchers:
  TYPE        ENABLED   OPTIONS
  apt         yes       security_only=false, use_sudo=true, hide_phased=true
  docker      yes       (defaults)
  wordpress   yes       name="Blog", path=/var/www/blog

Notifiers:
  TYPE        ENABLED   DESTINATION
  slack       yes       #ops-alerts
  email       yes       admin@example.com

Settings:
  Send Policy:  only-on-updates
  Log File:     /var/log/update-watcher.log

Cron:
  Status:     active
  Schedule:   0 7 * * * (daily at 07:00)
```

### What Is Shown

- **Hostname** -- The server name used in notifications. Shows "(auto-detect)" if not explicitly set.
- **Config File** -- The resolved path to the configuration file in use.
- **Watchers** -- Each configured checker with its type, enabled/disabled state, and non-default options.
- **Notifiers** -- Each configured notification channel with its type, enabled/disabled state, and target destination.
- **Settings** -- Global settings including send policy and log file path.
- **Cron** -- Whether a cron job is installed, and if so, its schedule expression.

Disabled watchers and notifiers are still listed, marked with "no" in the ENABLED column.

{{< /tab >}}

{{< tab >}}

For scripting and automation, request JSON output:

```bash {filename="Terminal"}
update-watcher status --format json
```

The JSON output includes the same information in a structured format suitable for parsing with tools like `jq`:

```bash {filename="Terminal"}
update-watcher status --format json | jq '.watchers[] | .type'
```

For detailed examples of working with JSON output, see [JSON Output](../advanced/json-output/).

{{< /tab >}}

{{< /tabs >}}

## Examples

### Quick Status Check

```bash {filename="Terminal"}
update-watcher status
```

### Check Which Watchers Are Enabled

```bash {filename="Terminal"}
update-watcher status --format json | jq '.watchers[] | select(.enabled == true) | .type'
```

### Verify Config File Location

```bash {filename="Terminal"}
update-watcher status --format json | jq '.config_file'
```

### Use a Specific Config File

```bash {filename="Terminal"}
update-watcher status --config /path/to/config.yaml
```

## Use Cases

- **Verify setup** -- After running the setup wizard or editing the YAML config, check that all watchers and notifiers are configured as expected.
- **Debug issues** -- Confirm which config file is being loaded and what settings are active.
- **Audit** -- Review the configuration on a server during maintenance or handoff.
- **Scripting** -- Use JSON output to programmatically check configuration state across multiple servers.

## Related

- [validate](../validate/) -- Validate the configuration for errors.
- [setup](../setup/) -- Interactive wizard for modifying the configuration.
- [Configuration](../../configuration/) -- Full YAML configuration reference.
- [JSON Output](../advanced/json-output/) -- Working with machine-readable output.
