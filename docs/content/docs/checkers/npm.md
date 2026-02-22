---
title: "npm Global Update Notifications - Monitor Globally Installed npm Packages"
description: "Check for outdated globally installed npm packages. Receive automated update notifications via Slack, Discord, Telegram and 13 more channels."
weight: 7
---

Update-Watcher's npm checker monitors globally installed npm packages for available updates. It reports each outdated package with its current and latest version, so you always know when your global CLI tools need updating.

The setup wizard auto-detects npm and offers to enable this checker on any system where `npm` is installed.

{{< callout emoji="💡" >}}
**Looking for project-level npm monitoring?** If you want to track outdated dependencies in a Node.js project (`package.json`), use the [Web Project checker](../webproject/) instead. It supports npm, yarn, pnpm, and Composer with auto-detection, security audits, and multi-project support.
{{< /callout >}}

## Prerequisites

{{< callout type="info" >}}
- Node.js and `npm` installed and available in `PATH`.
- At least one globally installed npm package (e.g., `npm install -g typescript`).
{{< /callout >}}

## Adding via CLI

Add an npm global watcher:

```bash {filename="Terminal"}
update-watcher watch npm
```

The npm checker has no additional configuration flags.

## Configuration Reference

The npm checker has no checker-specific options. It uses the default settings.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| (none) | -- | -- | The npm checker requires no additional configuration. |

## YAML Configuration Example

Basic npm configuration:

```yaml {filename="config.yaml"}
watchers:
  - type: npm
```

Combined with other checkers for a full development machine setup:

```yaml {filename="config.yaml"}
watchers:
  - type: homebrew
  - type: npm
  - type: macos
```

## How It Works

The npm checker runs a single command:

```text
npm outdated -g --json
```

This queries the npm registry for all globally installed packages that have newer versions available. It does not update any packages -- it only reports what is outdated.

The JSON output contains each outdated package with its `current`, `wanted`, and `latest` versions. For global packages, the checker compares `current` against `latest` to determine if an update is available.

## Tips

{{< callout emoji="💡" >}}
**Global CLI tools:** Many developers install CLI tools globally with `npm install -g` (e.g., `typescript`, `eslint`, `@angular/cli`, `vercel`). These tools do not auto-update, so they can become outdated quickly. The npm checker helps you stay on top of these updates without having to remember to check manually.

**Scoped packages:** The checker fully supports scoped packages like `@angular/cli` and `@vue/cli`.
{{< /callout >}}

To update all outdated global packages:

```bash {filename="Terminal"}
npm update -g
```

Or update a specific package:

```bash {filename="Terminal"}
npm install -g <package-name>@latest
```

## Related

Send npm update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
