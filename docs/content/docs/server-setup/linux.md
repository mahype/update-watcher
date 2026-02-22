---
title: "Linux Server Setup - Production-Ready Configuration"
description: "Set up Update-Watcher on a Linux production server with a dedicated system user, minimal sudo permissions, Docker access, and cron scheduling."
weight: 1
---

This guide walks through setting up Update-Watcher on a Linux production server with proper security boundaries. The result is a dedicated system user with minimal permissions that runs automated update checks on a daily schedule.

{{% steps %}}

## Step 1: Create a Dedicated System User

Create a system user with no login shell and a home directory for storing state:

```bash {filename="Terminal"}
sudo useradd -r -s /usr/sbin/nologin -m -d /var/lib/update-watcher update-watcher
```

| Flag | Purpose |
|------|---------|
| `-r` | Create a system account (low UID, no aging). |
| `-s /usr/sbin/nologin` | Prevent interactive login. |
| `-m -d /var/lib/update-watcher` | Create a home directory for the user. |

## Step 2: Configuration Directory and Permissions

Create the configuration directory and set ownership:

```bash {filename="Terminal"}
sudo mkdir -p /etc/update-watcher
sudo chown update-watcher:update-watcher /etc/update-watcher
sudo chmod 755 /etc/update-watcher
```

If a configuration file already exists, move it and set restrictive permissions:

```bash {filename="Terminal"}
sudo mv config.yaml /etc/update-watcher/config.yaml
sudo chown update-watcher:update-watcher /etc/update-watcher/config.yaml
sudo chmod 600 /etc/update-watcher/config.yaml
```

{{< callout type="warning" >}}
The `600` permission ensures only the `update-watcher` user can read the file, which is important because it may contain webhook URLs, API tokens, and SMTP credentials.
{{< /callout >}}

## Step 3: Log File Setup (Optional)

If you want persistent logging beyond cron mail:

```bash {filename="Terminal"}
sudo touch /var/log/update-watcher.log
sudo chown update-watcher:update-watcher /var/log/update-watcher.log
sudo chmod 644 /var/log/update-watcher.log
```

Then configure the log file path in your config:

```yaml {filename="config.yaml"}
settings:
  log_file: "/var/log/update-watcher.log"
```

## Step 4: Sudoers Configuration

Most system package manager checkers need sudo to refresh package lists. Grant the `update-watcher` user passwordless sudo for only the specific commands it needs.

Create a sudoers drop-in file:

```bash {filename="Terminal"}
sudo visudo -f /etc/sudoers.d/update-watcher
```

Add the rules for your distribution's package manager. Only include the section for the package manager(s) you use.

{{< tabs items="Ubuntu/Debian,Fedora/RHEL,Arch,openSUSE,Alpine" >}}

{{< tab >}}
### APT (Debian, Ubuntu)

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/apt-get update
```
{{< /tab >}}

{{< tab >}}
### DNF (Fedora, RHEL, Rocky, AlmaLinux)

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/dnf check-update
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/dnf updateinfo
```
{{< /tab >}}

{{< tab >}}
### Pacman (Arch Linux, Manjaro)

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/pacman -Sy
```
{{< /tab >}}

{{< tab >}}
### Zypper (openSUSE, SLES)

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/zypper refresh
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/zypper list-updates
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/zypper list-patches
```
{{< /tab >}}

{{< tab >}}
### APK (Alpine Linux)

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /sbin/apk update
```
{{< /tab >}}

{{< /tabs >}}

### Multiple Package Managers

If you monitor multiple package managers on the same host (uncommon but possible), combine the rules in a single file:

```text {filename="/etc/sudoers.d/update-watcher"}
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/apt-get update
update-watcher ALL=(ALL) NOPASSWD: /usr/bin/snap refresh --list
```

{{< callout type="error" >}}
Always validate the sudoers file syntax after editing to avoid locking yourself out of sudo access.
{{< /callout >}}

```bash {filename="Terminal"}
sudo visudo -c -f /etc/sudoers.d/update-watcher
```

## Step 5: Docker Access

If you are using the Docker checker to monitor running containers, add the `update-watcher` user to the `docker` group:

```bash {filename="Terminal"}
sudo usermod -aG docker update-watcher
```

{{< callout type="warning" >}}
This grants access to the Docker socket (`/var/run/docker.sock`) without requiring sudo. The Docker checker queries image digests to detect newer versions but never pulls images or modifies containers.
{{< /callout >}}

## Step 6: WordPress and Web Project Access

If you are monitoring WordPress sites or web projects, the `update-watcher` user needs read access to the project directories. Add the user to the web server group:

```bash {filename="Terminal"}
sudo usermod -aG www-data update-watcher
```

On some distributions, the web server group may be `nginx`, `apache`, or `http` instead of `www-data`. Check with:

```bash {filename="Terminal"}
stat -c '%G' /var/www
```

Ensure the project directories are group-readable:

```bash {filename="Terminal"}
sudo chmod -R g+r /var/www/mysite
```

## Step 7: Cron Scheduling

Install a cron job under the dedicated user:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -e
```

Add the following entry:

```text {filename="Crontab"}
# update-watcher: daily update check
0 7 * * * /usr/local/bin/update-watcher run --quiet --as-service-user
```

Alternatively, use the built-in cron management while running as the service user:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher install-cron
```

For more scheduling options including twice-daily runs and logging, see [Cron Scheduling](../cron/).

{{% /steps %}}

## Summary Table

| Resource | Path | Permissions | Owner |
|----------|------|-------------|-------|
| Binary | `/usr/local/bin/update-watcher` | `755` | `root:root` |
| Config directory | `/etc/update-watcher/` | `755` | `update-watcher:update-watcher` |
| Config file | `/etc/update-watcher/config.yaml` | `600` | `update-watcher:update-watcher` |
| Log file | `/var/log/update-watcher.log` | `644` | `update-watcher:update-watcher` |
| Home directory | `/var/lib/update-watcher/` | `700` | `update-watcher:update-watcher` |
| Sudoers drop-in | `/etc/sudoers.d/update-watcher` | `440` | `root:root` |

## Security Notes

{{< callout type="error" >}}
Review these security properties to ensure your deployment meets your requirements.
{{< /callout >}}

- **No inbound ports** -- Update-Watcher does not listen on any ports. All network activity is outbound HTTPS to notification services and the GitHub API.
- **Read-only operations** -- Checkers never install updates, pull Docker images, or modify your system. They only query for available updates.
- **Minimal sudo** -- The sudoers configuration grants access only to the specific package manager commands needed for refreshing package lists.
- **Secret protection** -- The config file (which may contain webhook URLs, API tokens, and credentials) is readable only by the `update-watcher` user (`chmod 600`).

## Verifying the Setup

{{< callout emoji="💡" >}}
Run these verification commands after completing all setup steps to confirm everything is working correctly.
{{< /callout >}}

Run a test check as the service user to verify everything works:

```bash {filename="Terminal"}
sudo -u update-watcher update-watcher run --verbose --notify=false
```

Check each component:

```bash {filename="Terminal"}
# Verify config is readable
sudo -u update-watcher update-watcher validate

# Verify status
sudo -u update-watcher update-watcher status

# Verify cron is installed
sudo crontab -u update-watcher -l
```

## Related

- [macOS Setup](../macos/) -- Setup guide for macOS.
- [Cron Scheduling](../cron/) -- Detailed cron options and logging.
- [Configuration](../../configuration/) -- Full YAML configuration reference.
- [Security Best Practices](../../configuration/security/) -- Additional hardening guidance.
