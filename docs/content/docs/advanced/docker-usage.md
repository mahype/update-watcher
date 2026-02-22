---
title: "Running Update-Watcher in Docker"
description: "Run Update-Watcher inside a Docker container to monitor a Docker host. Dockerfile and Docker Compose examples with environment variable injection."
weight: 3
---

Update-Watcher can run inside a Docker container while monitoring the host's Docker daemon and other services. This is useful for environments where installing binaries on the host is restricted, for centralized monitoring setups, and for consistent deployment across infrastructure.

## Use Case

Running Update-Watcher in a container is appropriate when:

- You manage infrastructure through Docker Compose and want monitoring to follow the same pattern.
- Host-level binary installation is restricted by policy or permissions.
- You want to version-lock the monitoring tool alongside your application stack.
- You need a reproducible monitoring setup that can be deployed across multiple hosts.

The containerized instance can monitor the host's Docker containers by mounting the Docker socket, and it can send notifications through any of the 16 supported channels.

## Dockerfile

A multi-stage Dockerfile that produces a minimal image with only the Update-Watcher binary:

```dockerfile {filename="Dockerfile"}
# Stage 1: Download the binary
FROM alpine:3.19 AS downloader

ARG TARGETARCH
ARG VERSION=latest

RUN apk add --no-cache curl jq

RUN if [ "$VERSION" = "latest" ]; then \
      VERSION=$(curl -s https://api.github.com/repos/mahype/update-watcher/releases/latest | jq -r .tag_name); \
    fi && \
    curl -sSL "https://github.com/mahype/update-watcher/releases/download/${VERSION}/update-watcher_linux_${TARGETARCH}" \
      -o /usr/local/bin/update-watcher && \
    chmod +x /usr/local/bin/update-watcher

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=downloader /usr/local/bin/update-watcher /usr/local/bin/update-watcher

# Create a non-root user
RUN adduser -D -s /sbin/nologin watcher
USER watcher

ENTRYPOINT ["update-watcher"]
CMD ["run", "--quiet"]
```

Build the image:

```bash {filename="Terminal"}
docker build -t update-watcher:latest .
```

## Docker Compose

A Docker Compose configuration that mounts the config file and Docker socket:

```yaml {filename="docker-compose.yml"}
services:
  update-watcher:
    image: update-watcher:latest
    container_name: update-watcher
    restart: "no"
    volumes:
      - ./config.yaml:/home/watcher/.config/update-watcher/config.yaml:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    env_file:
      - .env
    user: "1000:${DOCKER_GID:-999}"
```

{{< callout type="warning" >}}
Mounting the Docker socket (`/var/run/docker.sock`) gives the container access to the host's Docker daemon. This is equivalent to root access on the host. Only mount it in trusted containers and use read-only mode (`:ro`).
{{< /callout >}}

### Volume Mounts

| Mount | Purpose |
|-------|---------|
| `config.yaml` | The Update-Watcher configuration file, mounted read-only into the container. |
| `/var/run/docker.sock` | The Docker socket, required for the Docker checker to query running containers and image digests. |

### User and Group

The container runs as a non-root user. To access the Docker socket, the user must belong to the Docker group on the host. Set the `DOCKER_GID` environment variable to the GID of the `docker` group on your host:

```bash {filename="Terminal"}
# Find the Docker group GID
DOCKER_GID=$(getent group docker | cut -d: -f3)

# Pass it to Docker Compose
DOCKER_GID=$DOCKER_GID docker compose run update-watcher
```

Or set it in your `.env` file:

```bash {filename=".env"}
DOCKER_GID=999
```

## Environment Variable Injection

Use an `.env` file to inject secrets into the container without hardcoding them in the config file. The config file references environment variables with `${VAR}` syntax:

**config.yaml:**

```yaml {filename="config.yaml"}
hostname: "docker-host-01"

watchers:
  - type: docker

notifiers:
  - type: slack
    options:
      webhook_url: "${SLACK_WEBHOOK_URL}"
  - type: email
    options:
      smtp_host: "${SMTP_HOST}"
      smtp_port: 587
      smtp_user: "${SMTP_USER}"
      smtp_pass: "${SMTP_PASS}"
      from: "update-watcher@example.com"
      to: "admin@example.com"

settings:
  send_policy: "only-on-updates"
```

**.env:**

```bash {filename=".env"}
SLACK_WEBHOOK_URL=<your-slack-webhook-url>
SMTP_HOST=smtp.example.com
SMTP_USER=update-watcher@example.com
SMTP_PASS=app-specific-password
```

{{< callout type="warning" >}}
The `.env` file contains secrets and should not be committed to version control. Add it to `.gitignore`.
{{< /callout >}}

## Scheduling Inside the Container

Since the container does not run a cron daemon by default, you have several options for scheduling:

{{< tabs items="Docker Compose,Docker CLI" >}}
{{< tab >}}

### Docker Compose with Restart Policy

Run as a one-shot container triggered by the host's cron:

```bash {filename="Terminal"}
# In the host's crontab
0 7 * * * docker compose -f /opt/update-watcher/docker-compose.yaml run --rm update-watcher
```

{{< /tab >}}
{{< tab >}}

### External Scheduler

Use any external scheduler (systemd timer, Kubernetes CronJob, CI/CD pipeline) to trigger the container:

```bash {filename="Terminal"}
# systemd timer example
0 7 * * * docker run --rm \
  -v /opt/update-watcher/config.yaml:/home/watcher/.config/update-watcher/config.yaml:ro \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  --env-file /opt/update-watcher/.env \
  update-watcher:latest
```

{{< /tab >}}
{{< /tabs >}}

## Monitoring Host Package Managers

{{< callout type="info" >}}
The Docker checker works natively from inside a container via the socket mount. However, monitoring host-level package managers (APT, DNF, etc.) from inside a container is not supported because the container does not have access to the host's package manager.
{{< /callout >}}

For host-level package monitoring, install Update-Watcher directly on the host. The Docker container approach is best suited for:

- Docker container image monitoring.
- WordPress and web project monitoring (with appropriate volume mounts).
- Notification-only setups where the host's package managers are monitored by a separate instance.

## Full Example

A complete setup for monitoring Docker containers from a containerized Update-Watcher:

{{< filetree/container >}}
  {{< filetree/folder name="project" >}}
    {{< filetree/file name="docker-compose.yaml" >}}
    {{< filetree/file name="config.yaml" >}}
    {{< filetree/file name=".env" >}}
    {{< filetree/file name=".gitignore" >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

**docker-compose.yaml:**

```yaml {filename="docker-compose.yml"}
services:
  update-watcher:
    build: .
    container_name: update-watcher
    restart: "no"
    volumes:
      - ./config.yaml:/home/watcher/.config/update-watcher/config.yaml:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    env_file:
      - .env
    user: "1000:${DOCKER_GID:-999}"
```

**.gitignore:**

```text {filename=".gitignore"}
.env
```

Run manually or via cron:

```bash {filename="Terminal"}
docker compose run --rm update-watcher
```

## Related

- [Docker Checker](../../checkers/docker/) -- Full documentation for the Docker checker.
- [Configuration](../../configuration/) -- YAML config reference.
- [Environment Variables](../../configuration/environment-variables/) -- Environment variable substitution and `.env` files.
