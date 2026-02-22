---
title: "Advanced Topics"
description: "Advanced Update-Watcher features: JSON output for scripting, exit codes, Docker containerization, and notification policies."
weight: 60
---

This section covers advanced features and integration patterns for Update-Watcher. These topics go beyond the basic setup and are aimed at users who want to integrate Update-Watcher into scripting pipelines, monitoring systems, or containerized environments.

## Topics

{{< cards >}}
  {{< card link="json-output" title="JSON Output" subtitle="Machine-readable output for scripting, parsing with jq, and integration with monitoring tools." icon="code" >}}
  {{< card link="exit-codes" title="Exit Codes" subtitle="Exit code reference for shell scripting. Automate responses based on update status." icon="terminal" >}}
  {{< card link="docker-usage" title="Docker Usage" subtitle="Run Update-Watcher inside a Docker container to monitor a Docker host." icon="cube" >}}
  {{< card link="send-policy" title="Send Policy" subtitle="Control when notifications are sent. Always, only on updates, or CLI overrides." icon="bell" >}}
{{< /cards >}}

## Overview

### JSON Output

Every check result can be output as structured JSON using the `--format json` flag. This makes it straightforward to parse results with `jq`, feed them into monitoring dashboards, or build custom notification workflows. See [JSON Output](json-output/).

### Exit Codes

The `run` command returns specific exit codes (0 through 4) that indicate the outcome of the check. These codes are designed for shell scripting: exit 0 means no updates, exit 1 means updates were found, and higher codes indicate failures. See [Exit Codes](exit-codes/).

### Docker Usage

You can run Update-Watcher inside a Docker container while monitoring the host's Docker daemon and other services. This is useful for centralized monitoring setups and environments where installing binaries on the host is restricted. See [Docker Usage](docker-usage/).

### Send Policy

The `send_policy` setting controls whether notifications are sent on every run or only when updates are found. CLI flags can override this for testing and debugging. See [Send Policy](send-policy/).

## Next Steps

- [CLI Reference](../cli/) -- Complete command-line reference.
- [Configuration](../configuration/) -- Full YAML configuration reference.
- [Server Setup](../server-setup/) -- Production-ready deployment guides.
