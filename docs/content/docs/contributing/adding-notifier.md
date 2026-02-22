---
title: "Adding a New Notifier"
description: "Step-by-step guide to implementing a new notification backend for Update-Watcher. Includes code templates, registration, formatting helpers, and testing."
weight: 3
---

# Adding a New Notifier

This guide walks you through adding a new notification backend to Update-Watcher. A notifier delivers update check results to a specific service (chat platform, push service, webhook, etc.).

## Overview

Adding a new notifier involves these steps:

1. Create a new package under `notifier/`
2. Implement the `Notifier` interface
3. Register the notifier and its metadata in `init()`
4. Add a blank import in `runner/runner.go`
5. (Optional) Add to the setup wizard

{{% steps %}}

## Step 1: Create the Package

Create a new directory under `notifier/`:

{{< filetree/container >}}
  {{< filetree/folder name="notifier" >}}
    {{< filetree/folder name="myservice" >}}
      {{< filetree/file name="myservice.go" >}}
      {{< filetree/file name="myservice_test.go" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

## Step 2: Implement the Notifier Interface

Your notifier must implement the `Notifier` interface:

```go {filename="notifier/notifier.go"}
type Notifier interface {
    Name() string
    Send(ctx context.Context, hostname string, results []*checker.CheckResult) error
}
```

Here is a real-world example based on the ntfy notifier:

```go {filename="notifier/myservice/myservice.go"}
package myservice

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/mahype/update-watcher/checker"
    "github.com/mahype/update-watcher/config"
    "github.com/mahype/update-watcher/notifier"
    "github.com/mahype/update-watcher/notifier/formatting"
)

func init() {
    notifier.Register("myservice", NewFromConfig)
    notifier.RegisterMeta(notifier.NotifierMeta{
        Type:        "myservice",
        DisplayName: "My Service",
        Description: "Send notifications via My Service",
    })
}

// MyServiceNotifier sends update reports to My Service.
type MyServiceNotifier struct {
    webhookURL string
    httpClient *http.Client
}

// NewFromConfig creates a MyServiceNotifier from a notifier configuration.
func NewFromConfig(cfg config.NotifierConfig) (notifier.Notifier, error) {
    url := cfg.Options.GetString("webhook_url", "")
    if url == "" {
        return nil, fmt.Errorf("myservice: webhook_url is required")
    }

    return &MyServiceNotifier{
        webhookURL: url,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }, nil
}

func (n *MyServiceNotifier) Name() string { return "myservice" }

func (n *MyServiceNotifier) Send(ctx context.Context, hostname string, results []*checker.CheckResult) error {
    // Use the shared formatting helpers
    title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())

    // Send to your service
    req, err := http.NewRequest("POST", n.webhookURL, strings.NewReader(body))
    if err != nil {
        return fmt.Errorf("myservice: %w", err)
    }
    req.Header.Set("Content-Type", "text/plain")

    resp, err := n.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("myservice: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("myservice: server returned %d", resp.StatusCode)
    }

    return nil
}
```

{{< callout type="info" >}}
**Key points:**
- **`init()`** registers both the factory (`notifier.Register`) and display metadata (`notifier.RegisterMeta`). The metadata is used by the setup wizard.
- **`NewFromConfig`** reads options from `cfg.Options` using typed accessors. Validate required options and return errors for missing ones.
- **`Send()`** receives the hostname and all check results. Use the `formatting` package to build messages.
{{< /callout >}}

{{< callout type="info" >}}
**Formatting helpers available in `notifier/formatting/`:**
- Use **`formatting.BuildMarkdownMessage()`** for Markdown-formatted messages.
- Use **`formatting.BuildPlainTextMessage()`** for plain text.
- Use **`formatting.SummarizeResults()`** to get aggregate counts (total updates, security count).
{{< /callout >}}

## Step 3: Message Formatting

The `notifier/formatting/` package provides shared helpers:

```go {filename="notifier/formatting/formatting.go"}
// Build a Markdown-formatted message
title, body := formatting.BuildMarkdownMessage(hostname, results, formatting.DefaultOptions())

// Build a plain text message
title, body := formatting.BuildPlainTextMessage(hostname, results)

// Get aggregate stats
summary := formatting.SummarizeResults(results)
// summary.TotalUpdates, summary.SecurityCount, summary.CheckerCount
```

{{< callout emoji="💡" >}}
For services that require structured data (like Slack Block Kit or Teams Adaptive Cards), build the payload manually using the `results` data directly.
{{< /callout >}}

## Step 4: Register in runner.go

Add a blank import in `runner/runner.go`:

```go {filename="runner/runner.go"}
import (
    // ... existing imports
    _ "github.com/mahype/update-watcher/notifier/myservice"
)
```

## Step 5: Test Your Notifier

Create `notifier/myservice/myservice_test.go`:

```go {filename="notifier/myservice/myservice_test.go"}
package myservice

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/mahype/update-watcher/checker"
)

func TestSend(t *testing.T) {
    // Mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    n := &MyServiceNotifier{
        webhookURL: server.URL,
        httpClient: &http.Client{Timeout: 5 * time.Second},
    }

    results := []*checker.CheckResult{
        {
            CheckerName: "apt",
            Updates: []checker.Update{
                {Name: "openssl", CurrentVersion: "3.0.1", NewVersion: "3.0.2", Type: checker.UpdateTypeSecurity},
            },
            Summary:   "1 package (1 security)",
            CheckedAt: time.Now(),
        },
    }

    err := n.Send(context.Background(), "test-server", results)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

{{% /steps %}}

## Add to Config Example

Update `config.example.yaml` with your new notifier:

```yaml {filename="config.yaml"}
notifiers:
  # Secrets: webhook_url
  - type: myservice
    enabled: false
    options:
      webhook_url: "${MYSERVICE_WEBHOOK_URL}"
```

## Checklist

Before submitting a PR for a new notifier:

- [ ] Implements `Notifier` interface (`Name()`, `Send()`)
- [ ] Registered via `init()` with `notifier.Register()` and `notifier.RegisterMeta()`
- [ ] Blank import added to `runner/runner.go`
- [ ] Uses `formatting` package for message building
- [ ] Validates required options in `NewFromConfig`
- [ ] Secrets use `${ENV_VAR}` references in config example
- [ ] Tests written and passing (`make test`)
- [ ] Code formatted (`make fmt`) and linted (`make lint`)
- [ ] Documentation updated (notifier options in README)
