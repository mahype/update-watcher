---
title: "Web Project Dependency Monitoring - npm, Yarn, pnpm & Composer Update Notifications"
description: "Monitor outdated packages and security vulnerabilities in npm, yarn, pnpm, and Composer projects. Auto-detects package managers. Notifications via 16 channels."
weight: 12
---

Update-Watcher's Web Project checker monitors web application dependencies for outdated packages and known security vulnerabilities. It supports four package managers -- npm, yarn, pnpm, and Composer -- and auto-detects which ones are in use based on lock files present in your project directory.

This checker is designed for web agencies, DevOps teams, and developers who manage multiple projects and need centralized visibility into dependency freshness and security status.

## Prerequisites

{{< callout type="info" >}}
- At least one of the following package managers installed:
  - **npm** (ships with Node.js)
  - **yarn** (v1 or v2+)
  - **pnpm**
  - **Composer** (PHP package manager)
- The project directory must contain the relevant lock files for auto-detection.
- For containerized environments (ddev, Lando, Docker Compose), the respective tool must be installed and the project must be running.
{{< /callout >}}

## Adding via CLI

Add a web project watcher:

```bash {filename="Terminal"}
update-watcher watch webproject --path /var/www/myapp --name "My App"
```

Specify which package managers to check (skip auto-detection):

```bash {filename="Terminal"}
update-watcher watch webproject --path /var/www/myapp --name "My App" --managers npm,composer
```

Disable security audit checks:

```bash {filename="Terminal"}
update-watcher watch webproject --path /var/www/myapp --name "My App" --no-audit
```

Specify the environment:

```bash {filename="Terminal"}
update-watcher watch webproject --path /var/www/myapp --name "My App" --env ddev
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `projects` | list | `[]` | List of web project objects to monitor. Each project has its own name, path, and settings. |
| `check_audit` | bool | `true` | Run security audit commands in addition to checking for outdated packages. |

### Project Object Properties

Each entry in the `projects` list supports the following properties:

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `name` | string | (required) | A human-readable name for the project, used in notifications. |
| `path` | string | (required) | Absolute path to the project directory. |
| `environment` | string | `""` | Override the auto-detected environment. Leave empty for auto-detection. |
| `managers` | string list | `[]` | Explicitly specify which package managers to check. Leave empty for auto-detection. |
| `run_as` | string | `""` | Run commands as a specific user. |

## YAML Configuration Example

Single project with auto-detection:

```yaml {filename="config.yaml"}
watchers:
  - type: webproject
    projects:
      - name: "Company Website"
        path: /var/www/company-site
```

Multiple projects with different configurations:

```yaml {filename="config.yaml"}
watchers:
  - type: webproject
    check_audit: true
    projects:
      - name: "Frontend App"
        path: /var/www/frontend
        managers:
          - pnpm
      - name: "Backend API"
        path: /var/www/api
        managers:
          - npm
          - composer
      - name: "Client Project"
        path: /home/dev/client-project
        environment: ddev
```

Disable security audits (only check for outdated packages):

```yaml {filename="config.yaml"}
watchers:
  - type: webproject
    check_audit: false
    projects:
      - name: "My App"
        path: /var/www/myapp
```

## Package Manager Detection

The checker auto-detects package managers based on lock files present in the project directory:

| Package Manager | Detection File | Outdated Command | Audit Command |
|----------------|---------------|-----------------|---------------|
| npm | `package-lock.json` | `npm outdated --json` | `npm audit --json` |
| yarn | `yarn.lock` | `yarn outdated --json` | `yarn audit --json` |
| pnpm | `pnpm-lock.yaml` | `pnpm outdated --json` | `pnpm audit --json` |
| Composer | `composer.json` | `composer outdated --format=json` | `composer audit --format=json` |

### Priority When Multiple Node Managers Are Detected

{{< callout type="info" >}}
If a project contains lock files for multiple Node.js package managers, the checker uses the following priority order:

1. **pnpm** (highest priority)
2. **yarn**
3. **npm** (lowest priority)

Only one Node.js package manager is used per project to avoid duplicate results. Composer is always checked alongside the selected Node.js manager if `composer.json` is present, since it manages a separate dependency tree (PHP).
{{< /callout >}}

## Supported Environments

The web project checker supports running in containerized development environments:

| Environment | Detection Method | Command Prefix |
|-------------|-----------------|----------------|
| Native | Default fallback | Commands run directly |
| ddev | `.ddev/` directory present | `ddev exec` |
| Lando | `.lando.yml` present | `lando` |
| Docker Compose | `docker-compose.yml` present | `docker compose exec <service>` |

## How It Works

For each configured project, the checker performs the following steps:

{{% steps %}}

### Step 1: Detect environment

Scans the project path for environment indicators (`.ddev/`, `.lando.yml`, `docker-compose.yml`) to determine how to run commands. If the `environment` property is set, auto-detection is skipped.

### Step 2: Detect package managers

Scans the project path for lock files (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `composer.json`) to determine which package managers are in use. If `managers` is explicitly set, auto-detection is skipped.

### Step 3: Check outdated packages

Runs the outdated command for each detected package manager and parses the JSON output to identify packages with newer versions available.

### Step 4: Run security audits (if enabled)

Runs the audit command for each detected package manager to check installed dependencies against known vulnerability databases.

{{% /steps %}}

The checker reports each outdated package with the package name, current version, available version, and the package manager. Security audit results are reported separately, including the vulnerability severity level when available.

## FAQ

{{< details title="FAQ: Which package managers are supported?" >}}
The checker supports four package managers: npm, yarn, pnpm, and Composer. For Node.js projects, the checker auto-detects which manager is in use based on the lock file. For PHP projects, Composer is detected via the `composer.json` file.
{{< /details >}}

{{< details title="FAQ: Does it run security audits?" >}}
Yes, by default. The `check_audit` option (enabled by default) runs the audit command for each detected package manager. This checks your installed dependencies against public vulnerability databases (npm advisory database, GitHub Advisory Database, Packagist security advisories, etc.).

You can disable audits with `check_audit: false` if you only want to track outdated packages.
{{< /details >}}

{{< details title="FAQ: Does it work with ddev, Lando, or Docker Compose?" >}}
Yes. The checker auto-detects containerized environments and routes commands through the appropriate tool. For example, in a ddev environment, `npm outdated` becomes `ddev exec npm outdated`. This works transparently for all supported package managers.
{{< /details >}}

## Tips

{{< callout emoji="💡" >}}
**Agency and Multi-Project Setups:** If you manage many web projects (common for agencies), you can add all of them to a single Update-Watcher configuration. Each project's results appear separately in notifications, grouped by project name.

**Combining with WordPress Checker:** For projects that include both a WordPress installation and custom frontend assets (e.g., a theme with npm dependencies), you can use both the WordPress checker and the Web Project checker.

**Monorepos:** For monorepo projects where multiple lock files exist at the root, you can explicitly set the `managers` list to control which package managers are checked. The auto-detection priority (pnpm > yarn > npm) handles most cases correctly, but explicit configuration gives you full control.
{{< /callout >}}

Agency multi-project example:

```yaml {filename="config.yaml"}
watchers:
  - type: webproject
    check_audit: true
    projects:
      - name: "Client A - Website"
        path: /var/www/client-a
      - name: "Client B - E-Commerce"
        path: /var/www/client-b
        environment: ddev
      - name: "Client C - API"
        path: /var/www/client-c
        managers:
          - composer
```

Combining with WordPress checker:

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    sites:
      - name: "My WP Site"
        path: /var/www/html
  - type: webproject
    projects:
      - name: "My WP Theme"
        path: /var/www/html/wp-content/themes/my-theme
        managers:
          - npm
```

## Related

Send web project update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
