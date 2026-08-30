# `toggles` Package Documentation

The `toggles` package provides feature flag management using [Unleash](https://www.getunleash.io/). It supports tenant-aware evaluation, automatic configuration from environment variables, and a fail-safe disabled mode that defaults all flags to enabled.

---

## Table of Contents

- [Features and Public API](#features-and-public-api)
  - [Initialization (`Init`)](#initialization-init)
  - [SetLogger](#setlogger)
  - [IsEnabled](#isenabled)
  - [Closing (`Close`)](#closing-close)
  - [SetMock](#setmock)
- [Disabled Mode](#disabled-mode)
- [Tenant-Aware Evaluation](#tenant-aware-evaluation)
- [Configuration](#configuration)
- [Example Usage](#example-usage)
- [Testing and Mocking](#testing-and-mocking)

---

## Features and Public API

### Initialization (`Init`)

Initializes the feature flag system and connects to the Unleash server. Uses `sync.Once` internally, so it is safe to call multiple times — only the first call takes effect.

> **Important:** If `Init` returns an error, all subsequent `IsEnabled` calls will return `false`. We recommend stopping the service (e.g. `log.Fatal`) on init failure rather than running with all features disabled.

You can pass a configuration struct to `Init` to programmatically set options instead of (or in addition to) using environment variables.

> **Note:** As a general practice, it is usually preferred to allow the package to load the configuration by itself (using environment variables) rather than providing a config struct to the `Init` function. This ensures a consistent configuration standard across all services using this package.

**Example:**

```go
enabled := true
err := toggles.Init(&toggles.Config{
    Enabled:                &enabled,
    Url:                    "http://unleash:4242/api",
    Environment:            toggles.FFSEDev,
    ApiKey:                 "<client-api-token>",
    RefreshIntervalSeconds: 15,
    ServiceName:            "my-service",
    ProjectName:            "default",
})
if err != nil {
    log.Fatal(err)
}
```

Pass `nil` to load defaults from environment variables:

```go
err := toggles.Init(nil)
if err != nil {
    log.Fatal(err)
}
```

- **Blocking**: Yes

---

### SetLogger

Sets a custom logger for the toggles package. Should be called **before** `Init`. If not set, a default no-op logger is used (no logs).

The logger must implement:

```go
type Logger interface {
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

**Example:**

```go
toggles.SetLogger(myLogger)
```

- **Blocking**: No

---

### IsEnabled

Checks whether a feature flag is enabled. Extracts the tenant ID from the context (if present) and passes it to Unleash for strategy evaluation. Returns `false` if the Unleash client is not initialized.

```go
if toggles.IsEnabled(ctx, "my-feature") {
    // feature is enabled
}
```

- **Blocking**: Yes

---

### Closing (`Close`)

Shuts down the Unleash client and resets internal state. After calling `Close`, the package reverts to the dummy toggler and can be re-initialized with a new `Init` call.

> **Note:** `Close` is not safe for concurrent use with `Init`, `IsEnabled`, or other `Close` calls.

```go
err := toggles.Close()
```

- **Blocking**: Yes

---

### SetMock

Replaces the global toggler instance with a mock for testing purposes. See [Testing and Mocking](#testing-and-mocking) for details.

```go
toggles.SetMock(myMockToggler)
```

- **Blocking**: No

---

## Disabled Mode

When `Enabled` is set to `false`, the package uses a dummy toggler that returns `true` for all feature flag checks. This is useful for self-hosted deployments where feature flags are not needed — all features are enabled by default.

```go
enabled := false
err := toggles.Init(&toggles.Config{
    Enabled: &enabled,
})
// All IsEnabled calls now return true
```

---

## Tenant-Aware Evaluation

Pass a tenant ID in the context to evaluate flags per-tenant:

```go
ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "tenant-123")
toggles.IsEnabled(ctx, "my-feature")
```

The tenant ID is sent to Unleash as a `tenantId` context property, which can be used in strategy constraints. If no tenant ID is present in the context, the flag is evaluated without tenant context.

---

## Configuration

All configuration options can be set via environment variables or via the `Init` function. Below is the full list of supported options:

| ENV Variable | Description | Default |
|---|---|---|
| `feature_flag_system_enabled` | Enable or disable the feature flag system | `true` |
| `feature_flag_system_url` | Unleash server URL (e.g. `http://unleash:4242/api`) | (required) |
| `feature_flag_system_env` | Environment: `production` or `development` | (required) |
| `feature_flag_system_api_key` | Unleash client API token | (required) |
| `feature_flag_system_refresh_interval_seconds` | How often to poll Unleash for updates (seconds) | `15` |
| `service_name` | App name registered with Unleash | (required) |
| `feature_flag_system_project_name` | Unleash project name | `Runai` |

When passing a `*toggles.Config` struct to `Init`, the corresponding fields are:

| Field | Description | Default |
|---|---|---|
| `Enabled` | Enable/disable the feature flag system (`*bool`) | `true` |
| `Url` | Unleash server URL | (required) |
| `Environment` | `toggles.FFSEDev` (`"development"`) or `toggles.FFSEProd` (`"production"`) | (required) |
| `ApiKey` | Unleash client API token | (required) |
| `RefreshIntervalSeconds` | Poll interval in seconds (must be >= 1) | `15` |
| `ServiceName` | App name registered with Unleash | (required) |
| `ProjectName` | Unleash project name | `"Runai"` |

---

## Example Usage

```go
import (
    "context"
    "fmt"

    "github.com/run-ai/runai/runai-common-packages/toggles"
    "github.com/run-ai/runai/runai-common-packages/runaictx"
)

func main() {
    // Optional: set a custom logger before Init
    toggles.SetLogger(myLogger)

    // Initialize (blocking, connects to Unleash)
    enabled := true
    err := toggles.Init(&toggles.Config{
        Enabled:                &enabled,
        Url:                    "http://unleash:4242/api",
        Environment:            toggles.FFSEDev,
        ApiKey:                 "<client-api-token>",
        RefreshIntervalSeconds: 15,
        ServiceName:            "my-service",
        ProjectName:            "default",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer toggles.Close()

    ctx := context.Background()

    // Check a feature flag
    if toggles.IsEnabled(ctx, "my-feature") {
        fmt.Println("Feature is enabled!")
    }

    // Check with tenant context
    tenantCtx := context.WithValue(ctx, runaictx.CfTenantId, "tenant-123")
    if toggles.IsEnabled(tenantCtx, "tenant-specific-feature") {
        fmt.Println("Feature enabled for this tenant!")
    }
}
```

---

## Testing and Mocking

### SetMock

Replaces the global toggler with a mock for unit testing. The mock must implement the `Toggler` interface:

```go
type Toggler interface {
    IsEnabled(ctx context.Context, name string) bool
    Close() error
}
```

Generated mocks are available in two packages:

- `toggles/testing` — provides `MockToggler` (testify mock)
- `toggles/mocks` — provides `Toggler` (testify mock)

**Example using the generated mock:**

```go
import (
    "context"
    "testing"

    "github.com/run-ai/runai/runai-common-packages/toggles"
    toggles_testing "github.com/run-ai/runai/runai-common-packages/toggles/testing"
    "github.com/stretchr/testify/mock"
)

func TestMyFunction(t *testing.T) {
    mockToggler := toggles_testing.NewMockToggler(t)

    // Set expectations
    mockToggler.On("IsEnabled", mock.Anything, "my-feature").Return(true)
    mockToggler.On("Close").Return(nil)

    // Inject the mock
    toggles.SetMock(mockToggler)

    // Your test code
    ctx := context.Background()
    result := toggles.IsEnabled(ctx, "my-feature")
    // result == true
}
```

### Running Unit Tests

```bash
make test
```

### Running Integration Tests

Integration tests spin up a local Unleash server via Docker Compose (PostgreSQL + Unleash v6):

```bash
make integration-tests
```

To start Unleash without running tests (for manual testing):

```bash
make integration-tests-debug
```

---

For more details, see the code and integration tests under `toggles/`.
