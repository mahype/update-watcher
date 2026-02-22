---
title: "Docker Image Update Notifications - Monitor Container Updates Without Pulling"
description: "Detect newer Docker images for running containers without pulling or modifying anything. Read-only update detection with notifications via Slack, Discord & more."
weight: 10
---

Update-Watcher's Docker checker detects when newer versions of Docker images are available for your running containers. It compares local image digests against remote registry digests without pulling images or modifying any containers. This is a fully read-only operation -- nothing on your system changes.

{{< callout type="info" >}}
The Docker checker is **completely read-only**. It never pulls images, never stops or restarts containers, and never modifies anything on your system. It only queries remote registries for manifest metadata to compare digests.
{{< /callout >}}

This makes Update-Watcher ideal for production Docker hosts where you want to know about available image updates but control the upgrade process manually.

## Prerequisites

{{< callout type="info" >}}
- Docker installed and running.
- The user running Update-Watcher must be in the `docker` group (or run as root) to access the Docker socket.
- `docker buildx` must be available (included in Docker Desktop and modern Docker Engine installations).
- Internet access to query remote registries.
{{< /callout >}}

{{< callout type="warning" >}}
The user running Update-Watcher needs access to the Docker socket (`/var/run/docker.sock`). This typically means the user must be in the `docker` group or run as root. Being in the `docker` group is effectively equivalent to root access on the host.
{{< /callout >}}

## Adding via CLI

Add a Docker watcher for all running containers:

```bash {filename="Terminal"}
update-watcher watch docker
```

The Docker checker monitors all running containers by default.

## Configuration Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `containers` | string | `"all"` | Which containers to monitor. Set to `"all"` to check every running container. |
| `exclude` | string list | `[]` | List of container names to exclude from monitoring. Useful for skipping locally-built development containers. |

## YAML Configuration Example

Basic Docker configuration (all running containers):

```yaml {filename="config.yaml"}
watchers:
  - type: docker
```

Exclude specific containers:

```yaml {filename="config.yaml"}
watchers:
  - type: docker
    exclude:
      - my-dev-container
      - local-test-app
```

Combined with APT for a complete server monitoring setup:

```yaml {filename="config.yaml"}
watchers:
  - type: apt
    security_only: true
    hide_phased: true
  - type: docker
    exclude:
      - dev-sandbox
```

## How It Works

The Docker checker performs the following steps for each running container:

{{% steps %}}

### Step 1: List running containers

Queries the Docker daemon for all currently running containers.

### Step 2: Get image references

For each container, retrieves the image reference (repository, tag, and local digest) it was started from.

### Step 3: Skip locally-built images

Images without a remote registry reference (i.e., locally built images with no push history) are automatically skipped.

### Step 4: Compare digests

Uses `docker buildx imagetools inspect` to query the remote registry for the current manifest digest of each image tag. Compares the remote digest against the local digest.

### Step 5: Report mismatches

If the remote digest differs from the local digest, the image has been updated in the registry since it was last pulled. The checker reports this container as having an available update.

{{% /steps %}}

All container checks run concurrently for performance. The checker never pulls images, never stops or restarts containers, and never modifies anything on the system.

### Digest Comparison

The checker uses manifest digests (sha256 hashes) for comparison rather than tag names. This means it detects updates even when the tag stays the same (e.g., when a `latest` tag or a `1.2` tag is rebuilt with newer contents). This is the only reliable way to detect image updates for mutable tags.

## FAQ

{{< details title="FAQ: Does the Docker checker pull images?" >}}
No. The Docker checker never pulls images. It uses `docker buildx imagetools inspect` to query the remote registry for the manifest digest. This is a metadata-only operation that downloads a few kilobytes of JSON -- no image layers are transferred.
{{< /details >}}

{{< details title="FAQ: How does Update-Watcher compare to Watchtower?" >}}
[Watchtower](https://containrrr.dev/watchtower/) is a popular Docker container that automatically updates running containers to the latest image. The key differences are:

| Feature | Watchtower | Update-Watcher |
|---------|-----------|----------------|
| **Purpose** | Auto-updates containers | Notification only |
| **Modifies system** | Yes (pulls images, restarts containers) | No (read-only) |
| **Scope** | Docker only | 14 checkers (Docker + APT, DNF, WordPress, etc.) |
| **Notifications** | Limited | 16 channels |
| **Control** | Automated | You decide when to update |

If you want full control over when containers are updated, Update-Watcher notifies you about available updates and lets you apply them on your own schedule. If you want hands-off auto-updates, Watchtower is the right tool.
{{< /details >}}

{{< details title="FAQ: How does Update-Watcher compare to Diun?" >}}
[Diun](https://crazymax.dev/diun/) (Docker Image Update Notifier) is a dedicated Docker image update notification tool. The key differences are:

| Feature | Diun | Update-Watcher |
|---------|------|----------------|
| **Scope** | Docker/OCI images only | 14 checkers (Docker, APT, WordPress, npm, etc.) |
| **Notifications** | Multiple channels | 16 channels |
| **Installation** | Docker container or binary | Single binary |
| **Config approach** | Dedicated Docker tool | Multi-purpose update watcher |

Diun is excellent if you only need Docker image monitoring. Update-Watcher is the better choice if you want a single tool that monitors Docker images alongside system packages, WordPress sites, web project dependencies, and more.
{{< /details >}}

{{< details title="FAQ: Can I exclude specific containers?" >}}
Yes. Use the `exclude` option to list container names that should be skipped:

```yaml {filename="config.yaml"}
watchers:
  - type: docker
    exclude:
      - my-dev-container
      - build-runner
      - test-app
```

This is useful for excluding locally-built development containers that have no remote registry image, or containers you intentionally pin to a specific version.
{{< /details >}}

## Tips

{{< callout emoji="💡" >}}
**Locally-Built Images:** The checker automatically skips images that appear to be locally built (no remote registry reference). You do not need to manually exclude these -- they are detected and skipped silently.

**Private Registries:** The checker uses `docker buildx imagetools inspect`, which respects your Docker credential configuration. If you are authenticated to a private registry (via `docker login` or credential helpers), the checker can query that registry for digest updates.

**Multi-Architecture Images:** The digest comparison works correctly with multi-architecture images. The checker compares the manifest list digest, which changes whenever any platform variant of the image is updated.
{{< /callout >}}

## Related

Send Docker update notifications to [Slack](/docs/notifiers/slack/), [Discord](/docs/notifiers/discord/), [Email](/docs/notifiers/email/), [Telegram](/docs/notifiers/telegram/), or any of the other [16 supported notification channels](/docs/notifiers/).
