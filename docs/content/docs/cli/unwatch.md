---
title: "unwatch - Remove a Watcher"
description: "Remove a configured update checker from your configuration."
weight: 4
---

The `unwatch` command removes a configured update checker (watcher) from your configuration file. For multi-instance checker types like WordPress and Web Project, use the `--name` flag to target a specific instance.

## Usage

```bash {filename="Terminal"}
update-watcher unwatch <type> [--name NAME]
```

Where `<type>` is the checker type to remove (e.g., `apt`, `docker`, `wordpress`).

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--name NAME` | string | Name of the specific instance to remove. Required for multi-instance types when more than one instance is configured. |

{{< callout type="info" >}}
The `--name` flag is only required for multi-instance checker types (`wordpress` and `webproject`) when more than one instance is configured. For singleton types and single-instance multi-instance types, it can be omitted.
{{< /callout >}}

## Examples

### Remove a Singleton Watcher

Remove the APT watcher:

```bash {filename="Terminal"}
update-watcher unwatch apt
```

Remove the Docker watcher:

```bash {filename="Terminal"}
update-watcher unwatch docker
```

For singleton types (all types except `wordpress` and `webproject`), no `--name` flag is needed since there can only be one instance of each type.

### Remove a Specific WordPress Site

If you have multiple WordPress sites configured, use `--name` to target the one you want to remove:

```bash {filename="Terminal"}
update-watcher unwatch wordpress --name "Production Blog"
```

### Remove a Specific Web Project

```bash {filename="Terminal"}
update-watcher unwatch webproject --name "Frontend App"
```

### Remove When Only One Instance Exists

If only a single WordPress or Web Project instance is configured, the `--name` flag is optional:

```bash {filename="Terminal"}
update-watcher unwatch wordpress
```

## Multi-Instance Checkers

The following checker types support multiple instances in a single configuration:

- **wordpress** -- Multiple WordPress sites, each identified by path and name.
- **webproject** -- Multiple web projects, each identified by path and name.

All other checker types are singletons. Running `unwatch` on a singleton type removes the only instance.

When multiple instances of a multi-instance type exist and no `--name` flag is provided, the command will prompt for clarification or return an error indicating that the `--name` flag is required.

## What Happens

When you run `update-watcher unwatch`, the command:

1. Loads the existing configuration file.
2. Finds the watcher entry matching the specified type (and name, if provided).
3. Removes the entry from the watchers list.
4. Saves the updated configuration to disk.

The watcher is immediately removed and will not be included in the next `update-watcher run` invocation.

## Verifying Removal

After removing a watcher, verify the change with:

```bash {filename="Terminal"}
update-watcher status
```

The removed watcher should no longer appear in the watchers list.

## Related

- [watch](../watch/) -- Add a new watcher to the configuration.
- [status](../status/) -- View the current configuration.
- [setup](../setup/) -- Interactive wizard for managing watchers.
