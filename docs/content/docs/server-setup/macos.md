---
title: "macOS Setup - Update-Watcher on macOS"
description: "Set up Update-Watcher on macOS. Configuration paths, Homebrew and softwareupdate checkers, and scheduling options."
weight: 2
---

Update-Watcher runs on macOS for monitoring Homebrew packages and native macOS software updates. The macOS setup is simpler than Linux since most checkers do not require elevated permissions or a dedicated system user.

## Configuration Location

On macOS, the configuration file is stored in the user config directory:

```text {filename="config.yaml"}
~/.config/update-watcher/config.yaml
```

Create the directory if it does not exist:

```bash {filename="Terminal"}
mkdir -p ~/.config/update-watcher
```

The setup wizard (`update-watcher setup`) creates this directory and file automatically.

## Available Checkers on macOS

Two checkers are specific to macOS:

### macOS Software Updates

The `macos` checker uses the native `softwareupdate` command to detect available system updates, including macOS version upgrades, security patches, and Safari updates.

```bash {filename="Terminal"}
update-watcher watch macos
```

For security-only filtering:

```bash {filename="Terminal"}
update-watcher watch macos --security-only
```

No sudo is required. The `softwareupdate --list` command runs without elevated permissions.

### Homebrew

The `homebrew` checker detects outdated Homebrew formulae and casks:

```bash {filename="Terminal"}
update-watcher watch homebrew
```

To skip cask (GUI application) updates and only report formulae:

```bash {filename="Terminal"}
update-watcher watch homebrew --no-casks
```

No sudo is required. Homebrew runs entirely under the current user.

### When to Use Which

{{< callout type="info" >}}
Both checkers can be enabled simultaneously. They report independently since they monitor different update sources.
{{< /callout >}}

| Scenario | Recommended Checkers |
|----------|---------------------|
| macOS workstation with Homebrew | `macos` + `homebrew` |
| macOS server (no Homebrew) | `macos` only |
| Homebrew-only monitoring | `homebrew` only |
| CI runner with Homebrew | `homebrew` (macOS checker may be noisy on CI) |

### Other Checkers

The following checkers also work on macOS without modification:

- **docker** -- If Docker Desktop is installed.
- **wordpress** -- If WordPress sites are accessible on the local filesystem.
- **webproject** -- If web projects with npm, yarn, pnpm, or Composer are present.
- **distro** -- Not applicable on macOS (only checks Linux distributions).
- **openclaw** -- Works on macOS.

## Example Configuration

A typical macOS configuration:

```yaml {filename="config.yaml"}
hostname: "macbook-pro"

watchers:
  - type: macos
  - type: homebrew
  - type: docker

notifiers:
  - type: slack
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"

settings:
  send_policy: "only-on-updates"
```

## Scheduling

{{< tabs items="Cron,launchd" >}}

{{< tab >}}

### Using install-cron

The simplest way to schedule daily checks on macOS:

```bash {filename="Terminal"}
update-watcher install-cron --time 09:00
```

This creates an entry in the user's crontab. Verify with:

```bash {filename="Terminal"}
crontab -l
```

{{< /tab >}}

{{< tab >}}

### Using launchd

macOS uses `launchd` as its native scheduling system. While cron works on macOS, a `launchd` plist is the "macOS-native" approach and handles sleep/wake correctly (running missed jobs when the machine wakes up).

Create a plist file at `~/Library/LaunchAgents/com.update-watcher.daily.plist`:

```xml {filename="com.update-watcher.plist"}
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.update-watcher.daily</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/update-watcher</string>
        <string>run</string>
        <string>--quiet</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>9</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>/tmp/update-watcher.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/update-watcher.err</string>
</dict>
</plist>
```

Load the plist:

```bash {filename="Terminal"}
launchctl load ~/Library/LaunchAgents/com.update-watcher.daily.plist
```

Verify it is loaded:

```bash {filename="Terminal"}
launchctl list | grep update-watcher
```

To unload:

```bash {filename="Terminal"}
launchctl unload ~/Library/LaunchAgents/com.update-watcher.daily.plist
```

{{< /tab >}}

{{< /tabs >}}

### Cron vs launchd

{{< callout type="info" >}}
For laptops that sleep frequently, `launchd` is the better choice since it catches up on missed runs. For always-on Mac Minis or CI runners, cron works well.
{{< /callout >}}

| Feature | cron | launchd |
|---------|------|---------|
| Setup | One command (`install-cron`) | Manual plist creation |
| Missed jobs | Not re-run after sleep | Re-runs missed jobs on wake |
| macOS native | Legacy but functional | Recommended by Apple |
| Removal | `uninstall-cron` | Manual `launchctl unload` |

## Permissions

Most macOS checkers require no special permissions:

- `softwareupdate --list` runs as the current user.
- `brew outdated` runs as the current user.
- Docker Desktop manages its own socket permissions.

No sudoers configuration is needed for a standard macOS setup.

## Related

- [Linux Server Setup](../linux/) -- Production-ready Linux server setup.
- [Cron Scheduling](../cron/) -- Detailed cron scheduling guide.
- [Homebrew Checker](../../checkers/homebrew/) -- Full documentation for the Homebrew checker.
- [macOS Checker](../../checkers/macos/) -- Full documentation for the macOS software update checker.
