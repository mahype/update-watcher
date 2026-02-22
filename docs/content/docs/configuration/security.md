---
title: "Security Best Practices - Config Permissions and Secret Management"
description: "Secure your Update-Watcher deployment. Config file permissions, environment variable secrets, dedicated system user, and network security best practices."
weight: 3
---

Update-Watcher handles sensitive data including webhook URLs, API tokens, SMTP passwords, and bot credentials. This page covers how to secure your deployment through file permissions, secret management, system user isolation, and network considerations.

## Config File Permissions

{{< callout type="error" >}}
The configuration file may contain secrets in plain text (or references to environment variables that resolve to secrets). Both `config.yaml` and `.env` files **must** have mode `0600` (owner read/write only). World-readable or group-readable permissions expose secrets to any user on the system.
{{< /callout >}}

Update-Watcher writes the config file with mode `0600` by default. After any manual edit, verify the permissions have not changed:

```bash {filename="Terminal"}
ls -la /etc/update-watcher/config.yaml
# Expected: -rw------- 1 update-watcher update-watcher ... config.yaml
```

For a user-level config:

```bash {filename="Terminal"}
ls -la ~/.config/update-watcher/config.yaml
# Expected: -rw------- 1 youruser youruser ... config.yaml
```

If permissions are incorrect, fix them:

```bash {filename="Terminal"}
chmod 600 /etc/update-watcher/config.yaml
```

Or for the user config:

```bash {filename="Terminal"}
chmod 600 ~/.config/update-watcher/config.yaml
```

### Why 0600?

A config file with world-readable or group-readable permissions exposes secrets to any user on the system. The `0600` permission ensures that only the file owner can read or write the file. On a multi-user system, this is critical.

## Environment Variable References for Secrets

The most secure way to handle credentials in Update-Watcher is to keep them out of the config file entirely. Instead of writing secrets as plain text, use `${VAR}` references that are resolved at runtime from environment variables.

**Less secure** -- secret in plain text:

```yaml {filename="config.yaml"}
notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
```

**More secure** -- secret referenced via environment variable:

```yaml {filename="config.yaml"}
notifiers:
  - type: slack
    enabled: true
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
```

The actual webhook URL is then set in the environment (or a `.env` file) and never appears in the config file. This approach has several advantages:

- Secrets are not stored in the YAML file, which may be backed up, copied, or accidentally committed to version control.
- Different environments (staging, production) can use different credentials without changing the config file.
- Secrets can be managed through system-level tools (e.g., HashiCorp Vault, AWS Secrets Manager, systemd credentials).

See [Environment Variables](../environment-variables/) for the full substitution syntax and `.env` file usage.

## .env File Safety

If you use a `.env` file to store secrets, treat it with the same care as the config file itself:

```bash {filename="Terminal"}
# Set strict permissions
chmod 600 /etc/update-watcher/.env
chown update-watcher:update-watcher /etc/update-watcher/.env
```

Additional precautions:

- **Add `.env` to `.gitignore`** in any repository that might contain it. Update-Watcher's own repository already includes this.
- **Do not place `.env` files in web-accessible directories.** If your server runs a web server, ensure the `.env` file is outside the document root.
- **Avoid logging the `.env` contents.** Be cautious with scripts that export variables -- redirect output away from log files that other users might access.

## Dedicated System User

On production Linux servers, run Update-Watcher under a dedicated system user rather than as root or your personal account. This follows the principle of least privilege and limits the damage if the process is compromised.

{{% steps %}}

### Create the system user

The recommended server setup creates a user called `update-watcher`:

```bash {filename="Terminal"}
sudo useradd -r -s /usr/sbin/nologin -m -d /var/lib/update-watcher update-watcher
```

### Verify user properties

Key properties of this user:

- **System user** (`-r`) -- does not appear in login screens.
- **No login shell** (`-s /usr/sbin/nologin`) -- cannot be used for interactive login.
- **Owns its config and log files** -- only this user can read the config containing secrets.

### Set up file ownership

The install script can create this user automatically during server setup. See [Server Setup](../../server-setup/) for the full walkthrough including directory permissions, log file setup, and cron configuration.

{{% /steps %}}

### File Ownership Summary

| Resource | Path | Owner | Permissions |
|----------|------|-------|-------------|
| Binary | `/usr/local/bin/update-watcher` | `root:root` | `0755` |
| Config directory | `/etc/update-watcher/` | `update-watcher:update-watcher` | `0755` |
| Config file | `/etc/update-watcher/config.yaml` | `update-watcher:update-watcher` | `0600` |
| .env file | `/etc/update-watcher/.env` | `update-watcher:update-watcher` | `0600` |
| Log file | `/var/log/update-watcher.log` | `update-watcher:update-watcher` | `0640` |
| Sudoers file | `/etc/sudoers.d/update-watcher` | `root:root` | `0440` |

## Minimal Sudoers Permissions

{{< callout type="warning" >}}
Rather than granting broad sudo access, create a sudoers file that allows **only** the specific commands needed by the checkers you have enabled. Use full paths to binaries and always use `NOPASSWD` since the dedicated user has no login shell.
{{< /callout >}}

Some checkers need `sudo` to refresh package lists (e.g., `apt-get update`, `pacman -Sy`). Create the file:

```bash {filename="Terminal"}
sudo visudo -f /etc/sudoers.d/update-watcher
```

Add entries only for the package managers you actually use:

{{< tabs items="Ubuntu/Debian,Fedora/RHEL,Arch,openSUSE,Alpine" >}}

{{< tab >}}
```bash {filename="/etc/sudoers.d/update-watcher"}
# APT (Debian/Ubuntu)
update-watcher ALL=(root) NOPASSWD: /usr/bin/apt-get update
```
{{< /tab >}}

{{< tab >}}
```bash {filename="/etc/sudoers.d/update-watcher"}
# DNF (Fedora/RHEL)
update-watcher ALL=(root) NOPASSWD: /usr/bin/dnf check-update
update-watcher ALL=(root) NOPASSWD: /usr/bin/dnf updateinfo list --security
```
{{< /tab >}}

{{< tab >}}
```bash {filename="/etc/sudoers.d/update-watcher"}
# Pacman (Arch)
update-watcher ALL=(root) NOPASSWD: /usr/bin/pacman -Sy
```
{{< /tab >}}

{{< tab >}}
```bash {filename="/etc/sudoers.d/update-watcher"}
# Zypper (openSUSE)
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive refresh
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-patches --category security
update-watcher ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-updates
```
{{< /tab >}}

{{< tab >}}
```bash {filename="/etc/sudoers.d/update-watcher"}
# APK (Alpine)
update-watcher ALL=(root) NOPASSWD: /sbin/apk update
```
{{< /tab >}}

{{< /tabs >}}

**Important rules:**

- Only grant access to commands the `update-watcher` user actually needs.
- Use full paths to the binaries.
- Use `NOPASSWD` since the user has no login shell and cannot enter a password.
- The sudoers file must have permissions `0440` and be owned by `root:root`. The `visudo` command handles this automatically.

If your server already refreshes package lists on a schedule (e.g., via `unattended-upgrades` on Debian/Ubuntu), you can skip the sudoers configuration entirely and set `use_sudo: false` in each watcher's options.

## Network Security

Update-Watcher has a minimal network footprint:

- **No inbound ports.** Update-Watcher does not listen on any network port. It is a CLI tool that runs, performs checks, sends notifications, and exits.
- **Outbound HTTPS only.** Notifications are sent via HTTPS to external services (Slack, Discord, Telegram, SMTP servers, etc.). No unencrypted HTTP traffic is used for notifications.
- **No persistent state.** There is no database, no server process, and no inter-process communication. The only persistent file is the configuration YAML.

### Firewall Considerations

If your server runs a restrictive outbound firewall, ensure HTTPS traffic (port 443) is allowed to the notification services you use. For SMTP email, allow the configured SMTP port (typically 587 for STARTTLS or 465 for SMTPS).

No inbound firewall rules are needed for Update-Watcher.

## Audit and Logging

Enable the log file in your configuration to maintain an audit trail of all checks and notifications:

```yaml {filename="config.yaml"}
settings:
  log_file: "/var/log/update-watcher.log"
```

Set appropriate permissions on the log file:

```bash {filename="Terminal"}
sudo touch /var/log/update-watcher.log
sudo chown update-watcher:update-watcher /var/log/update-watcher.log
sudo chmod 640 /var/log/update-watcher.log
```

The `0640` permission allows the owner to write and the group to read, while preventing access by other users. If you want to integrate with a log management system, point it at this file or add the `update-watcher` group to your log collector's user.

## Security Checklist

{{< callout type="info" >}}
Use this checklist to verify your deployment is properly secured:

- [ ] Config file permissions are `0600`.
- [ ] Secrets are stored as `${VAR}` references, not in plain text.
- [ ] `.env` file (if used) has `0600` permissions and is listed in `.gitignore`.
- [ ] A dedicated system user runs Update-Watcher (not root, not a personal account).
- [ ] Sudoers grants only the specific commands needed.
- [ ] Log file has `0640` permissions, owned by the `update-watcher` user.
- [ ] No inbound firewall rules reference Update-Watcher.
- [ ] The `update-watcher validate` command reports no errors.
{{< /callout >}}

## Next Steps

- [Environment Variables](../environment-variables/) -- Full reference for `${VAR}` substitution and `.env` files.
- [Config File Reference](../config-file/) -- Complete YAML configuration reference.
- [Server Setup](../../server-setup/) -- Full production server setup guide.
