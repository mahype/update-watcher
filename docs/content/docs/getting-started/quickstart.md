---
title: "Quickstart - Set Up Server Update Notifications in 5 Minutes"
description: "Install Update-Watcher and configure automated server update notifications in 5 minutes. Quick install, interactive setup, and cron scheduling."
weight: 1
---

This guide takes you from zero to automated update notifications in four steps. The entire process takes about five minutes.

## Prerequisites

- A Linux or macOS machine with internet access.
- Credentials for at least one notification channel (e.g., a Slack incoming webhook URL, a Telegram bot token, or an SMTP server).
- `curl` installed (present by default on most systems).

{{% steps %}}

### Step 1: Install

Run the one-line installer. It detects your OS and architecture, downloads the latest release, and installs the binary to `/usr/local/bin`.

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

{{< callout type="info" >}}
  On Linux, the installer will ask whether you want to set up a dedicated system user with proper permissions. For a quick start, you can skip this and set it up later (see [Server Setup](../../server-setup/)).
{{< /callout >}}

Verify the installation:

```bash {filename="Terminal"}
update-watcher version
```

For alternative installation methods, including manual download and building from source, see the [Installation](../installation/) page.

### Step 2: Configure with the Setup Wizard

Launch the interactive setup wizard:

```bash {filename="Terminal"}
update-watcher setup
```

The wizard presents a menu-driven interface that:

1. **Auto-detects** installed package managers on your system (e.g., APT on Debian/Ubuntu, DNF on Fedora, Homebrew on macOS).
2. **Guides you** through enabling watchers for the detected package managers.
3. **Walks you through** configuring at least one notification channel (Slack, Discord, Email, Telegram, etc.).
4. **Saves** the configuration to the appropriate config file.

```text {filename="Setup Wizard"}
=== update-watcher setup ===

Current configuration:
  Hostname:   web-prod-01
  Watchers:   APT, Docker
  Notifiers:  slack configured
  Cron:       not configured

What would you like to do?
  > Manage Watchers (APT, Docker)
  > Notifications (slack configured)
  > Settings (hostname: web-prod-01)
  > Cron Job (not configured)
  > Run Test Check (dry-run)
  > Save & Exit
```

For a detailed walkthrough of the wizard and manual YAML configuration, see [First Run](../first-run/).

### Step 3: Test Run

Run a check without sending notifications to verify everything works:

```bash {filename="Terminal"}
update-watcher run --notify=false
```

This executes all configured watchers and prints the results to the terminal. You should see output listing any available updates for each enabled checker. If a checker fails (for example, because a package manager is not installed), the output will indicate the error.

Once the dry run looks correct, run with notifications enabled to confirm delivery:

```bash {filename="Terminal"}
update-watcher run
```

Check your notification channel to verify the message arrived.

### Step 4: Schedule with Cron

Set up a daily cron job so Update-Watcher runs automatically. The built-in command creates a cron entry for the current user:

```bash {filename="Terminal"}
update-watcher install-cron
```

{{< callout emoji="💡" >}}
  By default, the cron job runs daily at **07:00**. To choose a different time: `update-watcher install-cron --time 09:00`
{{< /callout >}}

To verify the cron job was created:

```bash {filename="Terminal"}
crontab -l
```

You should see a line similar to:

```text {filename="Crontab"}
0 7 * * * /usr/local/bin/update-watcher run --quiet
```

{{% /steps %}}

That is it. Update-Watcher will now check for updates daily and notify you through your configured channels.

## What Happens Next

{{< callout type="info" >}}
  - **Updates found**: You receive a notification listing all available updates, grouped by package manager. Security updates are highlighted.
  - **No updates found**: By default (`send_policy: "only-on-updates"`), no notification is sent. Change this to `"always"` in the [configuration](../../configuration/config-file/) if you want daily confirmation messages.
  - **Errors**: If a checker fails, the exit code and log output will indicate what went wrong. Use `--verbose` for detailed debug output.
{{< /callout >}}

## Next Steps

- [Installation](../installation/) -- All installation methods and platform details.
- [First Run](../first-run/) -- Detailed walkthrough of the wizard and manual YAML configuration.
- [Configuration](../../configuration/) -- Full configuration reference including environment variables and security.
- [Server Setup](../../server-setup/) -- Production-ready setup with a dedicated system user, proper permissions, and sudoers configuration.
