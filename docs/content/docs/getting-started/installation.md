---
title: "Install Update-Watcher - Quick Setup for Linux & macOS"
description: "Install Update-Watcher in one command. Supports Linux (Debian, Ubuntu, Fedora, Arch, Alpine) and macOS. Single binary, no dependencies, optional server setup."
weight: 2
---

Update-Watcher ships as a single static binary with no runtime dependencies. Choose the installation method that fits your workflow.

## Quick Install (Recommended)

The install script detects your operating system and architecture, downloads the latest release from GitHub, and places the binary in `/usr/local/bin`.

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
```

{{< callout type="info" >}}
  On Linux, the script will interactively ask whether you want to perform the recommended **server setup**, which creates a dedicated `update-watcher` system user with restricted permissions. This is the recommended approach for production servers. See [Server Setup](../../server-setup/) for details on what it configures.
{{< /callout >}}

{{< details title="Non-Interactive Installation" >}}

For automated provisioning, CI/CD pipelines, or configuration management tools, pass flags to skip the interactive prompt.

**With server setup** (creates dedicated user, sudoers file, cron job):

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --server
```

**Without server setup** (binary only):

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash -s -- --no-server
```

{{< /details >}}

## Manual Download

Download the pre-built binary for your platform from [GitHub Releases](https://github.com/mahype/update-watcher/releases).

### Available Platforms

| Operating System | Architecture | Archive Filename |
|-----------------|-------------|-----------------|
| Linux | amd64 (x86_64) | `update-watcher_linux_amd64.tar.gz` |
| Linux | arm64 (aarch64) | `update-watcher_linux_arm64.tar.gz` |
| Linux | armv7 | `update-watcher_linux_armv7.tar.gz` |
| macOS | amd64 (Intel) | `update-watcher_darwin_amd64.tar.gz` |
| macOS | arm64 (Apple Silicon) | `update-watcher_darwin_arm64.tar.gz` |

### Download and Install

{{< tabs items="Linux amd64,Linux arm64,macOS Apple Silicon,macOS Intel,Raspberry Pi" >}}

  {{< tab >}}
  ```bash {filename="Terminal"}
  curl -sSL -o update-watcher.tar.gz \
    https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_linux_amd64.tar.gz
  tar xzf update-watcher.tar.gz
  sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
  rm update-watcher.tar.gz
  ```
  {{< /tab >}}

  {{< tab >}}
  ```bash {filename="Terminal"}
  curl -sSL -o update-watcher.tar.gz \
    https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_linux_arm64.tar.gz
  tar xzf update-watcher.tar.gz
  sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
  rm update-watcher.tar.gz
  ```
  {{< /tab >}}

  {{< tab >}}
  ```bash {filename="Terminal"}
  curl -sSL -o update-watcher.tar.gz \
    https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_darwin_arm64.tar.gz
  tar xzf update-watcher.tar.gz
  sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
  rm update-watcher.tar.gz
  ```
  {{< /tab >}}

  {{< tab >}}
  ```bash {filename="Terminal"}
  curl -sSL -o update-watcher.tar.gz \
    https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_darwin_amd64.tar.gz
  tar xzf update-watcher.tar.gz
  sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
  rm update-watcher.tar.gz
  ```
  {{< /tab >}}

  {{< tab >}}
  ```bash {filename="Terminal"}
  curl -sSL -o update-watcher.tar.gz \
    https://github.com/mahype/update-watcher/releases/latest/download/update-watcher_linux_armv7.tar.gz
  tar xzf update-watcher.tar.gz
  sudo install -m 0755 update-watcher /usr/local/bin/update-watcher
  rm update-watcher.tar.gz
  ```
  {{< /tab >}}

{{< /tabs >}}

## Build from Source

Building from source requires Go 1.21 or later.

```bash {filename="Terminal"}
git clone https://github.com/mahype/update-watcher.git
cd update-watcher
make build
sudo make install
```

The `make install` target copies the compiled binary to `/usr/local/bin/update-watcher`.

## Verify the Installation

After installing with any method, verify that the binary is available and working:

```bash {filename="Terminal"}
update-watcher version
```

{{< callout type="warning" >}}
  If the command is not found, ensure `/usr/local/bin` is in your `PATH`.
{{< /callout >}}

This prints the installed version, build date, and commit hash.

## Supported Linux Distributions

Update-Watcher itself runs on any Linux distribution. The checkers that are relevant depend on which package managers are installed:

| Distribution | Primary Checker | Notes |
|-------------|----------------|-------|
| Debian, Ubuntu | APT | Security-only filter, phased rollout detection |
| Fedora, RHEL, Rocky, AlmaLinux | DNF | Security classification support |
| Arch, Manjaro | Pacman | |
| openSUSE, SLES | Zypper | Security patch support |
| Alpine | APK | |
| Any with Docker | Docker | Read-only image comparison |
| Any with Homebrew | Homebrew | Formulae and casks |
| Any with Snap | Snap | |
| Any with Flatpak | Flatpak | |

Additional cross-platform checkers (WordPress, web projects, distro release, OpenClaw) work on any supported OS.

## Updating Update-Watcher

Update-Watcher can update itself to the latest release:

```bash {filename="Terminal"}
update-watcher self-update
```

{{< callout emoji="💡" >}}
  To check for a newer version without installing it: `update-watcher self-update --status`
{{< /callout >}}

## Uninstallation

### Uninstall Script

The uninstall script automatically detects and removes all installed components, including the binary, config files, cron job, log file, sudoers entry, and the dedicated system user (if created during server setup).

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/uninstall.sh | bash
```

For non-interactive use (skips confirmation prompts):

```bash {filename="Terminal"}
curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/uninstall.sh | bash -s -- --yes
```

{{< details title="Manual Removal" >}}

**Quick removal** (binary and config only):

```bash {filename="Terminal"}
update-watcher uninstall-cron
sudo rm /usr/local/bin/update-watcher
sudo rm -rf /etc/update-watcher
rm -rf ~/.config/update-watcher
```

**Full removal** including the dedicated server user and all supporting files:

```bash {filename="Terminal"}
sudo crontab -u update-watcher -r
sudo rm /usr/local/bin/update-watcher
sudo rm -rf /etc/update-watcher
sudo rm -f /var/log/update-watcher.log
sudo rm -f /etc/sudoers.d/update-watcher
sudo userdel -r update-watcher
```

{{< /details >}}

## Next Steps

- [Quickstart](../quickstart/) -- Get running in 5 minutes.
- [First Run](../first-run/) -- Walk through configuration and your first update check.
- [Server Setup](../../server-setup/) -- Production-ready Linux setup with a dedicated system user.
