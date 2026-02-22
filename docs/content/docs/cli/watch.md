---
title: "watch - Add a Watcher"
description: "Add a new update checker to your configuration. Supports 14 checker types with per-type flags."
weight: 3
---

The `watch` command adds a new update checker (watcher) to your configuration. Each watcher type corresponds to a package manager or update source. Some types accept additional flags to customize their behavior.

## Usage

```bash {filename="Terminal"}
update-watcher watch <type> [flags]
```

Where `<type>` is one of the 14 supported checker types listed below.

## Watcher Types and Flags

{{< tabs items="System Package Managers,macOS,Application Stores,Containers,Distribution Releases,Application Updates,Web Applications" >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch apt [--security-only] [--no-sudo] [--hide-phased]
```

| Flag | Description |
|------|-------------|
| `--security-only` | Only report security updates. Non-security updates are ignored. |
| `--no-sudo` | Do not use sudo for `apt-get update`. Use when running as root. |
| `--hide-phased` | Exclude phased rollout updates that your system cannot yet install. Enabled by default. |

```bash {filename="Terminal"}
update-watcher watch dnf [--security-only] [--no-sudo]
```

| Flag | Description |
|------|-------------|
| `--security-only` | Only report security updates. |
| `--no-sudo` | Do not use sudo for `dnf check-update`. |

```bash {filename="Terminal"}
update-watcher watch pacman [--no-sudo]
```

| Flag | Description |
|------|-------------|
| `--no-sudo` | Do not use sudo for `pacman -Sy`. |

```bash {filename="Terminal"}
update-watcher watch zypper [--security-only] [--no-sudo]
```

| Flag | Description |
|------|-------------|
| `--security-only` | Only report security patches. |
| `--no-sudo` | Do not use sudo for zypper commands. |

```bash {filename="Terminal"}
update-watcher watch apk [--no-sudo]
```

| Flag | Description |
|------|-------------|
| `--no-sudo` | Do not use sudo for `apk update`. |

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch macos [--security-only]
```

| Flag | Description |
|------|-------------|
| `--security-only` | Only report security updates from `softwareupdate`. |

```bash {filename="Terminal"}
update-watcher watch homebrew [--no-casks]
```

| Flag | Description |
|------|-------------|
| `--no-casks` | Only check formulae. Skip cask (GUI application) updates. |

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch snap
```

```bash {filename="Terminal"}
update-watcher watch flatpak
```

No additional flags. These checkers report all available updates from Snap and Flatpak respectively.

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch docker
```

No additional flags. Checks all running Docker containers for newer image versions by comparing digests.

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch distro [--lts-only]
```

| Flag | Description |
|------|-------------|
| `--lts-only` | Only notify about LTS (Long Term Support) releases. Ignores interim releases. |

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch openclaw [--channel CHANNEL]
```

| Flag | Description |
|------|-------------|
| `--channel CHANNEL` | Release channel to monitor (e.g., `stable`, `beta`). |

{{< /tab >}}

{{< tab >}}

```bash {filename="Terminal"}
update-watcher watch wordpress --path PATH [--name NAME] [--env TYPE]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--path PATH` | Yes | Absolute path to the WordPress installation directory. |
| `--name NAME` | No | Display name for this WordPress site in notifications. Defaults to the directory name. |
| `--env TYPE` | No | PHP environment type (e.g., `php`, `lando`, `ddev`, `docker`). Controls how WP-CLI is invoked. |

```bash {filename="Terminal"}
update-watcher watch webproject --path PATH [--name NAME] [--env TYPE] [--managers LIST] [--no-audit]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--path PATH` | Yes | Absolute path to the project root directory. |
| `--name NAME` | No | Display name for this project in notifications. Defaults to the directory name. |
| `--env TYPE` | No | Environment type that controls how package managers are invoked. |
| `--managers LIST` | No | Comma-separated list of package managers to check (e.g., `npm,composer`). Auto-detected if omitted. |
| `--no-audit` | No | Skip security audit checks. Only report outdated packages. |

{{< /tab >}}

{{< /tabs >}}

## Examples

Add an APT watcher with security-only filtering:

```bash {filename="Terminal"}
update-watcher watch apt --security-only
```

Add a Docker watcher:

```bash {filename="Terminal"}
update-watcher watch docker
```

Add a WordPress site with a custom name:

```bash {filename="Terminal"}
update-watcher watch wordpress --path /var/www/mysite --name "Production Blog"
```

Add a web project monitoring npm and Composer:

```bash {filename="Terminal"}
update-watcher watch webproject --path /var/www/app --managers npm,composer
```

Add a distro watcher that only reports LTS releases:

```bash {filename="Terminal"}
update-watcher watch distro --lts-only
```

## Multiple Instances

Most watcher types are singletons -- you can only have one per type in your configuration. The exceptions are **wordpress** and **webproject**, which support multiple instances. Each instance is identified by its `--path` and optional `--name`.

To add a second WordPress site:

```bash {filename="Terminal"}
update-watcher watch wordpress --path /var/www/site-a --name "Site A"
update-watcher watch wordpress --path /var/www/site-b --name "Site B"
```

To remove a specific instance, use the [unwatch](../unwatch/) command with `--name`.

## What Happens

When you run `update-watcher watch`, the command:

1. Loads the existing configuration file (or creates one if none exists).
2. Adds a new watcher entry with the specified type and options.
3. Saves the updated configuration to disk.

The watcher is immediately available for the next `update-watcher run` invocation.

## Related

- [unwatch](../unwatch/) -- Remove a configured watcher.
- [Checkers](../../checkers/) -- Detailed documentation for each checker type.
- [Configuration](../../configuration/) -- Full YAML configuration reference.
- [setup](../setup/) -- Interactive wizard for managing watchers.
