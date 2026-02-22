---
title: "Homebrew Update Notifications - Monitor Formulae & Cask Updates"
description: "Automatically check for outdated Homebrew formulae and casks on macOS and Linux. Get notified of available updates via Slack, Discord, Telegram and 13 more."
weight: 7
---

Update-Watcher's Homebrew checker monitors installed Homebrew formulae and casks for available updates. It covers both command-line tools installed as formulae and GUI applications installed as casks, giving you a complete picture of outdated Homebrew packages.

The setup wizard auto-detects Homebrew and offers to enable this checker on macOS and Linux systems where `brew` is installed.

## Prerequisites

{{< callout type="info" >}}
- [Homebrew](https://brew.sh/) installed on macOS or Linux.
- The `brew` command accessible from the user running Update-Watcher.
{{< /callout >}}

## Adding via CLI

Add a Homebrew watcher:

```bash {filename="Terminal"}
update-watcher watch homebrew
```

Exclude casks and only check formulae:

```bash {filename="Terminal"}
update-watcher watch homebrew --no-casks
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `include_casks` | bool | `true` | Also check for outdated casks (GUI applications). Set to `false` to only monitor formulae. |

## YAML Configuration Example

Basic Homebrew configuration (formulae and casks):

```yaml {filename="config.yaml"}
watchers:
  - type: homebrew
```

Formulae only (no casks):

```yaml {filename="config.yaml"}
watchers:
  - type: homebrew
    include_casks: false
```

Combined with the macOS checker for complete coverage:

```yaml {filename="config.yaml"}
watchers:
  - type: macos
  - type: homebrew
    include_casks: true
```

## How It Works

The Homebrew checker performs the following steps:

{{% steps %}}

### Step 1: Update Homebrew

Runs `brew update` to fetch the latest formulae and cask definitions from the Homebrew taps.

### Step 2: Check outdated formulae

Runs `brew outdated --json` to list all installed formulae with available updates. The JSON output provides package name, installed version, and available version.

### Step 3: Check outdated casks (if enabled)

Runs `brew outdated --cask --json` to list all installed casks with available updates.

{{% /steps %}}

The checker reports each outdated package with the name, installed version, available version, and whether it is a formula or a cask.

## Tips

{{< callout emoji="💡" >}}
**Homebrew on Linux:** Homebrew also works on Linux (known as Linuxbrew). The checker operates identically on Linux, though cask support is limited on Linux since most casks are macOS GUI applications.

**Auto-Update Interference:** Homebrew has a built-in auto-update feature that runs periodically when you use `brew` commands. This does not interfere with Update-Watcher -- the checker simply reads the current state of outdated packages after ensuring the latest definitions are fetched.

**Large Numbers of Outdated Packages:** If you have many Homebrew packages installed and rarely update them, the checker may report a large list. Consider running `brew upgrade` periodically or filtering the notification output to focus on specific packages you care about.
{{< /callout >}}

## Related

Send Homebrew update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
