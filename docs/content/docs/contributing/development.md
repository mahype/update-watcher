---
title: "Development Workflow"
description: "Set up the development environment for Update-Watcher. Build, test, lint, format, and release process."
weight: 4
---

# Development Workflow

This guide covers the development environment setup, build process, testing, and release workflow for Update-Watcher.

## Prerequisites

{{< callout type="info" >}}
- **Go 1.21+** (currently using 1.25)
- **make** (for build targets)
- **golangci-lint** (optional, for linting)
- **gofumpt** (optional, for formatting)
- **goreleaser** (optional, for snapshot builds)
{{< /callout >}}

## Getting the Source

{{% steps %}}

### Clone the repository

```bash {filename="Terminal"}
git clone https://github.com/mahype/update-watcher.git
```

### Enter the project directory

```bash {filename="Terminal"}
cd update-watcher
```

{{% /steps %}}

## Build

Build the binary to `bin/update-watcher`:

```bash {filename="Terminal"}
make build
```

This compiles with CGO disabled (`CGO_ENABLED=0`) and injects version information via ldflags:

```text {filename="Makefile"}
-X github.com/mahype/update-watcher/internal/version.Version
-X github.com/mahype/update-watcher/internal/version.Commit
-X github.com/mahype/update-watcher/internal/version.Date
```

Install to `/usr/local/bin`:

```bash {filename="Terminal"}
sudo make install
```

## Testing

Run all tests with the race detector:

```bash {filename="Terminal"}
make test
```

This runs:

```bash {filename="Terminal"}
go test -v -race -coverprofile=coverage.out ./...
```

View coverage:

```bash {filename="Terminal"}
go tool cover -html=coverage.out
```

## Linting and Formatting

Format code:

```bash {filename="Terminal"}
make fmt
```

Run the linter:

```bash {filename="Terminal"}
make lint
```

Run Go vet:

```bash {filename="Terminal"}
make vet
```

## Available Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build binary to `bin/update-watcher` |
| `install` | Copy binary to `/usr/local/bin/` |
| `test` | Run tests with race detector and coverage |
| `lint` | Run golangci-lint |
| `fmt` | Format code with gofumpt |
| `vet` | Run go vet |
| `clean` | Remove `bin/`, `dist/`, `coverage.out` |
| `snapshot` | Create snapshot release with goreleaser |

## Snapshot Builds

Test the release process locally without publishing:

```bash {filename="Terminal"}
make snapshot
```

This uses goreleaser to build cross-platform archives in `dist/`.

{{< details title="Release Process" >}}

Releases are automated via GitHub Actions when a git tag is pushed:

1. Create a tag: `git tag v1.2.3`
2. Push the tag: `git push origin v1.2.3`
3. GitHub Actions runs tests, then goreleaser builds and publishes the release

The release workflow (`.github/workflows/release.yaml`):
- Runs tests with the race detector
- Builds archives for Linux (amd64, arm64, armv7) and macOS (amd64, arm64)
- Creates a GitHub Release with checksums and changelog
- Changelog excludes docs, test, ci, and chore commits

{{< /details >}}

## Commit Conventions

Follow conventional commit style for meaningful changelogs:

| Prefix | Use for |
|--------|---------|
| `feat:` | New features |
| `fix:` | Bug fixes |
| `docs:` | Documentation (excluded from changelog) |
| `test:` | Tests (excluded from changelog) |
| `ci:` | CI/CD changes (excluded from changelog) |
| `chore:` | Maintenance (excluded from changelog) |
| `refactor:` | Code refactoring |

## Project Conventions

- **No CGO**: The binary must compile with `CGO_ENABLED=0` for maximum portability
- **No runtime dependencies**: The binary must be fully self-contained
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
- **Structured logging**: Use `log/slog` for all log output
- **60-second timeout**: External commands use a default 60-second timeout via `executil.Run()`
- **Config options**: Use `cfg.Options.GetString()`, `GetBool()`, `GetInt()` for typed access
