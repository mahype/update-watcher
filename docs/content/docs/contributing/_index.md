---
title: "Contributing"
description: "How to contribute to Update-Watcher. Bug reports, feature requests, documentation, and code contributions."
weight: 70
---

# Contributing to Update-Watcher

Thank you for your interest in contributing to Update-Watcher. Whether you are reporting a bug, suggesting a feature, improving documentation, or writing code, every contribution is welcome.

## Ways to Contribute

- **Bug reports** -- Open an issue on [GitHub](https://github.com/mahype/update-watcher/issues) with steps to reproduce
- **Feature requests** -- Suggest new checkers, notifiers, or improvements via issues
- **Documentation** -- Fix typos, improve explanations, or add examples
- **Code** -- Add new checkers, notifiers, or improve existing functionality

## Getting Started

{{% steps %}}

### Fork the repository on GitHub

### Clone your fork

```bash {filename="Terminal"}
git clone https://github.com/YOUR_USERNAME/update-watcher.git
```

### Create a feature branch

```bash {filename="Terminal"}
git checkout -b feature/my-feature
```

### Make your changes

### Run tests

```bash {filename="Terminal"}
make test
```

### Push and open a Pull Request

{{% /steps %}}

## Guides

{{< cards >}}
  {{< card link="architecture" title="Architecture Overview" subtitle="Understand the codebase structure and design patterns." >}}
  {{< card link="adding-checker" title="Adding a Checker" subtitle="Step-by-step guide to implementing a new update checker." >}}
  {{< card link="adding-notifier" title="Adding a Notifier" subtitle="Step-by-step guide to implementing a new notification backend." >}}
  {{< card link="development" title="Development Workflow" subtitle="Build, test, lint, and release process." >}}
{{< /cards >}}
