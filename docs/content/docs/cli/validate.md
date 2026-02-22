---
title: "validate - Validate Configuration"
description: "Validate your Update-Watcher configuration file for syntax errors and missing required fields."
weight: 6
---

The `validate` command checks your Update-Watcher configuration file for correctness. It verifies YAML syntax, required fields, valid checker and notifier types, and option value types. This is useful after manually editing the config file or before deploying a new configuration to a server.

## Usage

```bash {filename="Terminal"}
update-watcher validate
```

The `validate` command uses the same config file resolution as all other commands: it checks the `--config` flag first, then the default search paths (`/etc/update-watcher/config.yaml`, `~/.config/update-watcher/config.yaml`, `./config.yaml`).

## What Is Validated

The validator checks the following:

### YAML Syntax

The file must be valid YAML. Common syntax errors include incorrect indentation, missing colons, and unescaped special characters.

### Required Fields

- The `watchers` section must exist and contain at least one watcher entry.
- Each watcher entry must have a `type` field.
- Notifier entries that require credentials (e.g., Slack webhook URL, Telegram bot token) are checked for the presence of those fields.

### Valid Types

- Watcher `type` values must be one of the 14 supported checker types: `apt`, `dnf`, `pacman`, `zypper`, `apk`, `macos`, `homebrew`, `snap`, `flatpak`, `docker`, `distro`, `openclaw`, `wordpress`, `webproject`.
- Notifier `type` values must be one of the 16 supported notifier types.

### Option Types

- Boolean options (like `security_only`, `use_sudo`) must be boolean values, not strings.
- String options (like `webhook_url`, `path`) must be strings.
- The `send_policy` setting must be either `"only-on-updates"` or `"always"`.

## Output

### Valid Configuration

```text {filename="Output"}
$ update-watcher validate
Configuration is valid.
```

### Invalid Configuration

```text {filename="Output"}
$ update-watcher validate
Configuration errors found:

  watchers[0]: unknown type "atp" (did you mean "apt"?)
  watchers[1].options.security_only: expected bool, got string "yes"
  notifiers[0].options: missing required field "webhook_url" for type "slack"
```

Each error message indicates the location in the config file and a description of the problem.

## Examples

### Validate the Default Config File

```bash {filename="Terminal"}
update-watcher validate
```

### Validate a Specific Config File

```bash {filename="Terminal"}
update-watcher validate --config /path/to/config.yaml
```

### Validate Before Deployment

A common pattern is to validate a config file before copying it to a production server:

```bash {filename="Terminal"}
update-watcher validate --config ./staging-config.yaml && \
  scp ./staging-config.yaml server:/etc/update-watcher/config.yaml
```

### Validate in CI/CD

{{< callout emoji="💡" >}}
Add validation to your deployment pipeline to catch configuration errors before they reach production.
{{< /callout >}}

```bash {filename="Terminal"}
update-watcher validate --config ./config.yaml
if [ $? -ne 0 ]; then
  echo "Config validation failed"
  exit 1
fi
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Configuration is valid. |
| 4 | Configuration errors were found. |

## Related

- [status](../status/) -- View the resolved configuration.
- [Configuration](../../configuration/) -- Full YAML configuration reference.
- [setup](../setup/) -- Interactive wizard for creating valid configurations.
