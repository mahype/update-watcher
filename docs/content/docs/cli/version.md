---
title: "version - Show Version Info"
description: "Display the Update-Watcher version, git commit hash, and build date."
weight: 10
---

The `version` command displays the current Update-Watcher version, the git commit hash it was built from, and the build date. Use the `--short` flag to output only the version number.

## Usage

```bash {filename="Terminal"}
update-watcher version [--short]
```

## Flags

| Flag | Description |
|------|-------------|
| `--short` | Output only the version number (e.g., `v1.3.0`). Omits the commit hash and build date. |

## Output

### Full Version Info

```bash {filename="Terminal"}
update-watcher version
```

```text {filename="Output"}
update-watcher v1.3.0
  commit: a1b2c3d
  built:  2025-04-15T10:30:00Z
```

The output includes:

- **Version** -- The semantic version number (e.g., `v1.3.0`).
- **Commit** -- The short git commit hash of the source code this binary was built from.
- **Built** -- The date and time the binary was compiled, in ISO 8601 format.

### Short Version

```bash {filename="Terminal"}
update-watcher version --short
```

```text {filename="Output"}
v1.3.0
```

The short output is useful in scripts where you need just the version string:

```bash {filename="Terminal"}
CURRENT_VERSION=$(update-watcher version --short)
echo "Running Update-Watcher $CURRENT_VERSION"
```

## Use Cases

- **Bug reports** -- Include the full version output when reporting issues.
- **Deployment verification** -- Confirm which version is installed on a server after deployment.
- **Scripting** -- Use `--short` to capture the version number for comparison or logging.
- **Inventory** -- Collect version information across multiple servers for fleet management.

## Examples

### Check If Installed

```bash {filename="Terminal"}
update-watcher version > /dev/null 2>&1 && echo "Installed" || echo "Not found"
```

### Compare Versions in a Script

```bash {filename="Terminal"}
INSTALLED=$(update-watcher version --short)
REQUIRED="v1.3.0"

if [ "$INSTALLED" != "$REQUIRED" ]; then
  echo "Version mismatch: installed $INSTALLED, required $REQUIRED"
fi
```

## Related

- [self-update](../self-update/) -- Update to the latest version.
- [Installation](../../getting-started/installation/) -- All installation methods.
