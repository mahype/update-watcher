---
title: "WordPress Update Monitoring - Core, Plugin & Theme Update Notifications"
description: "Monitor WordPress core, plugin, and theme updates across multiple sites. Auto-detects ddev, Lando, Docker Compose & 8 more environments. CLI-based, no WP plugin."
weight: 11
---

Update-Watcher's WordPress checker monitors one or more WordPress installations for available updates to the core software, plugins, and themes. It works entirely through WP-CLI and requires no WordPress plugin. The checker auto-detects 11 different development and hosting environments, including ddev, Lando, Docker Compose, and native installations.

This is one of the most powerful checkers in Update-Watcher. It can monitor dozens of WordPress sites from a single configuration file and report all pending updates through your notification channels.

## Prerequisites

{{< callout type="info" >}}
- [WP-CLI](https://wp-cli.org/) installed and accessible for native environments.
- For containerized environments (ddev, Lando, Docker Compose, etc.), the respective tool must be installed and the project must be running.
- File system access to the WordPress installation directory (or the project directory for containerized environments).
{{< /callout >}}

## Adding via CLI

Add a WordPress watcher for a single site:

```bash {filename="Terminal"}
update-watcher watch wordpress --path /var/www/html --name "My Site"
```

Specify the environment type explicitly:

```bash {filename="Terminal"}
update-watcher watch wordpress --path /var/www/html --name "My Site" --env ddev
```

Disable theme checking:

```bash {filename="Terminal"}
update-watcher watch wordpress --path /var/www/html --name "My Site" --no-themes
```

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `sites` | list | `[]` | List of WordPress site objects to monitor. Each site has its own name, path, and optional settings. |
| `check_core` | bool | `true` | Check for WordPress core updates. |
| `check_plugins` | bool | `true` | Check for plugin updates. |
| `check_themes` | bool | `true` | Check for theme updates. |

### Site Object Properties

Each entry in the `sites` list supports the following properties:

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `name` | string | (required) | A human-readable name for the site, used in notifications. |
| `path` | string | (required) | Absolute path to the WordPress installation or project directory. |
| `run_as` | string | `""` | Optional. Run WP-CLI as a specific user via sudo (e.g., `www-data`). Only needed when Update-Watcher runs as root. For dedicated service users, prefer group membership instead (see [Linux Server Setup](../../server-setup/linux/)). |
| `environment` | string | `""` | Override the auto-detected environment. Leave empty for auto-detection. |

## YAML Configuration Example

Single WordPress site with default settings:

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    sites:
      - name: "Company Blog"
        path: /var/www/wordpress
```

Multiple sites with different environments:

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    check_core: true
    check_plugins: true
    check_themes: true
    sites:
      - name: "Production Site"
        path: /var/www/production
        run_as: www-data
      - name: "Staging Site"
        path: /home/dev/staging-site
        environment: ddev
      - name: "Client Site"
        path: /var/www/client
        environment: docker-compose
```

Plugins and core only (skip themes):

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    check_core: true
    check_plugins: true
    check_themes: false
    sites:
      - name: "My Site"
        path: /var/www/html
```

## Supported Environments

The WordPress checker auto-detects the environment for each site. You can also set the environment explicitly if auto-detection does not work for your setup.

| Environment | Detection Method | WP-CLI Command |
|-------------|-----------------|----------------|
| Native | Default fallback | `wp` (directly) |
| ddev | `.ddev/` directory present | `ddev wp` |
| Lando | `.lando.yml` present | `lando wp` |
| wp-env | `.wp-env.json` present | `npx wp-env run cli wp` |
| Docker Compose | `docker-compose.yml` present | `docker compose exec <service> wp` |
| Bedrock | `web/wp/` directory structure | `wp` with adjusted paths |
| LocalWP | Local by Flywheel site structure | `wp` via Local's shell |
| MAMP | MAMP directory structure detected | `wp` with MAMP PHP |
| XAMPP | XAMPP directory structure detected | `wp` with XAMPP PHP |
| Laragon | Laragon directory structure detected | `wp` via Laragon paths |
| Laravel Valet | Valet-linked site detected | `wp` directly |

## How It Works

For each configured WordPress site, the checker performs the following steps:

{{% steps %}}

### Step 1: Detect environment

Scans the site path for environment indicators (`.ddev/`, `.lando.yml`, `docker-compose.yml`, etc.) to determine how to invoke WP-CLI. If the `environment` property is set, auto-detection is skipped.

### Step 2: Check core updates

Runs `wp core check-update` to see if a newer version of WordPress is available.

### Step 3: Check plugin updates

Runs `wp plugin list --update=available --format=json` to list all plugins with pending updates.

### Step 4: Check theme updates

Runs `wp theme list --update=available --format=json` to list all themes with pending updates.

{{% /steps %}}

All WP-CLI commands are executed in the context of the detected environment. For example, in a ddev environment, all commands are prefixed with `ddev` so they run inside the container.

The checker reports each available update with the component type (core, plugin, or theme), component name, current version, and available version.

## FAQ

{{< details title="FAQ: Do I need to install a WordPress plugin?" >}}
No. Update-Watcher uses WP-CLI externally to query for updates. No WordPress plugin is required. This means the checker works even if the WordPress admin dashboard is not accessible, and it introduces zero overhead to your WordPress site's runtime performance.
{{< /details >}}

{{< details title="FAQ: Can I monitor multiple WordPress sites?" >}}
Yes. Add multiple entries to the `sites` list in the configuration. Each site can have its own path, environment, and `run_as` setting:

```yaml {filename="config.yaml"}
watchers:
  - type: wordpress
    sites:
      - name: "Site A"
        path: /var/www/site-a
      - name: "Site B"
        path: /var/www/site-b
        run_as: www-data
      - name: "Dev Site"
        path: /home/dev/project
        environment: ddev
```

There is no limit to the number of sites you can monitor.
{{< /details >}}

{{< details title="FAQ: Which environments are supported?" >}}
The checker supports 11 environments: Native, ddev, Lando, wp-env, Docker Compose, Bedrock, LocalWP, MAMP, XAMPP, Laragon, and Laravel Valet. The environment is auto-detected based on the files and directory structure present at the site path. You can override auto-detection with the `environment` property if needed.
{{< /details >}}

## Tips

{{< callout type="warning" >}}
**File Permissions:** The user running Update-Watcher needs read access to `wp-config.php` and the WordPress directory. The recommended setup is to add the `update-watcher` service user to the web server group (e.g., `www-data`). See [Linux Server Setup](../../server-setup/linux/) for details.

If Update-Watcher runs as root instead, use `run_as` to execute WP-CLI as the file owner:

```yaml {filename="config.yaml"}
sites:
  - name: "Production"
    path: /var/www/html
    run_as: www-data
```
{{< /callout >}}

{{< callout emoji="💡" >}}
**Mixed Environments:** A single Update-Watcher configuration can monitor sites across different environments. For example, you might monitor a native production site, a ddev development site, and a Docker Compose staging site, all from the same config file.

**Notification Grouping:** When monitoring multiple WordPress sites, update notifications are grouped by site name. This makes it easy to see at a glance which sites need attention.
{{< /callout >}}

## Related

Send WordPress update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
