# Run:AI Installer API

This package provides the public API for the Run:AI installer library. It's designed to be used by external services and applications that need to programmatically install Run:AI.

## Features

- **Global Instance Management**: Singleton pattern for easy access throughout your application
- **Type-Safe Configuration**: Strongly-typed configuration flags and options
- **Event-Based Progress Tracking**: Receive real-time installation status updates
- **Clean API Surface**: Minimal, focused API that re-exports only what clients need

## Quick Start

### Basic Usage

```go
package main

import (
    "log"
    "time"
    
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/api"
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/common/config"
)

func main() {
    // Create configuration
    cfg := &api.Config{
        Mode:        config.ModeLib,
        StepTimeout: 7 * time.Minute,
        GPUType:     config.GPUTypeReal,
        Env:         config.EnvProd,
        ConfigValues: map[string]string{
            api.FlagFQDN:        "runai.example.com",
            api.FlagClusterName: "production-cluster",
            api.FlagJfrogCreds:  "base64-encoded-credentials",
        },
    }

    // Create installation data
    runaiData, err := api.NewRunAIInstallData(cfg)
    if err != nil {
        log.Fatalf("Failed to create installation data: %v", err)
    }

    // Initialize global installer
    if err := api.Init(runaiData); err != nil {
        log.Fatalf("Failed to initialize installer: %v", err)
    }

    // Get the installer instance
    inst := api.Get()

    // Run installation with event listener
    eventsChan, err := inst.RunInstallationWithEventListner()
    if err != nil {
        log.Fatalf("Failed to start installation: %v", err)
    }

    // Process events
    for event := range eventsChan {
        handleEvent(event)
    }
    
    log.Println("Installation completed successfully!")
}

func handleEvent(event *api.FlowStatusReportingMsg) {
    switch event.FlowStatus {
    case api.NotInstalled:
        log.Printf("[%s] Not installed", event.FlowName)
    case api.Installing:
        log.Printf("[%s] Installing...", event.FlowName)
    case api.Installed:
        log.Printf("[%s] ✓ Installed", event.FlowName)
    case api.Failed:
        log.Printf("[%s] ✗ Failed", event.FlowName)
    case api.Timeout:
        log.Printf("[%s] ⏱ Timeout", event.FlowName)
    }
}
```

### Async Installation with Timeout

```go
func InstallWithTimeout(timeout time.Duration) error {
    // Create configuration
    cfg := &api.Config{
        Mode:        config.ModeLib,
        StepTimeout: 7 * time.Minute,
        GPUType:     config.GPUTypeReal,
        Env:         config.EnvProd,
    }
    
    // Create installation data
    runaiData, err := api.NewRunAIInstallData(cfg)
    if err != nil {
        return err
    }
    
    // Initialize global installer
    if err := api.Init(runaiData); err != nil {
        return err
    }
    
    // Get installer and start installation
    inst := api.Get()
    eventsChan, err := inst.RunInstallationWithEventListner()
    if err != nil {
        return err
    }
    
    // Wait with timeout
    timeoutTimer := time.NewTimer(timeout)
    defer timeoutTimer.Stop()
    
    for {
        select {
        case event, ok := <-eventsChan:
            if !ok {
                // Channel closed, installation completed
                return nil
            }
            
            // Check for failure
            if event.FlowStatus == api.Failed {
                return fmt.Errorf("installation flow '%s' failed", event.FlowName)
            }
            if event.FlowStatus == api.Timeout {
                return fmt.Errorf("installation flow '%s' timed out", event.FlowName)
            }
            
        case <-timeoutTimer.C:
            return fmt.Errorf("installation timed out after %v", timeout)
        }
    }
}
```

## Configuration Flags

The following configuration flags are available:

| Flag | Description | Example |
|------|-------------|---------|
| `FlagJfrogCreds` | JFrog credentials (base64) | `"YWRtaW46cGFzc3dvcmQ="` |
| `FlagFQDN` | Fully Qualified Domain Name | `"runai.example.com"` |
| `FlagClusterName` | Kubernetes cluster name | `"my-cluster"` |
| `FlagIngressControllerHTTPSPort` | HTTPS NodePort (30000-32767) | `"30443"` |
| `FlagSSLCertPath` | SSL certificate path | `"/path/to/cert.pem"` |
| `FlagSSLKeyPath` | SSL key path | `"/path/to/key.pem"` |
| `FlagCAChainPath` | CA chain path | `"/path/to/ca-chain.pem"` |
| `FlagWildcardFQDN` | Wildcard FQDN for inference | `"*.inference.runai.example.com"` |
| `FlagWildcardCertPath` | Wildcard cert path | `"/path/to/wildcard-cert.pem"` |
| `FlagWildcardCertKeyPath` | Wildcard key path | `"/path/to/wildcard-key.pem"` |

## Installation Phases

The installer reports status through these phases:

- `NotInstalled` - Component is not yet installed
- `Installing` - Component is currently being installed
- `Installed` - Component was installed successfully
- `Failed` - Component installation failed
- `Timeout` - Component installation timed out

## API Reference

### Global Instance Management

```go
// Initialize the global installer
func Init(runaiData *RunAIInstallData) error

// Get the global installer instance
func Get() *Installer

// Check if initialized
func IsInitialized() bool
```

### Installation Data

```go
// Create installation data
func NewRunAIInstallData(opts *Config) (*RunAIInstallData, error)
```

### Event Types

```go
// FlowStatusReportingMsg represents an installation event
type FlowStatusReportingMsg struct {
    FlowStatus InstallPhase  // Current phase
    FlowName   string         // Name of the flow
}

// InstallPhase represents installation status
type InstallPhase int

const (
    NotInstalled InstallPhase = iota
    Installing
    Failed
    Installed
    Timeout
)
```

### Types

```go
// Installer is the main installer instance
type Installer = installer.Installer

// RunAIInstallData contains all installation configuration and state
type RunAIInstallData = common.RunAIInstallData

// Config is the installer configuration
type Config = config.Config
```

## Best Practices

1. **Use Global Instance**: The API is designed around a global singleton pattern. Initialize once with `Init()` and access anywhere with `Get()`

2. **Handle Events Asynchronously**: Installation events are sent on a channel. Make sure to consume them to avoid blocking

3. **Set Appropriate Timeouts**: Configure `StepTimeout` in the Config based on your cluster size and network speed

4. **Check Installation Phase**: Always check for `Failed` and `Timeout` phases in your event handler

5. **Clean Up Resources**: The event channel will close when installation completes. Always drain the channel properly

## Integration Example (Provisioner Service)

See how the provisioner-service uses this API:

```go
// provisioner-service/pkg/workflow/install.go
package workflow

import (
    "fmt"
    "time"
    
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/api"
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/common/config"
)

func InstallRunAI(envID string, mode string) error {
    // Create configuration
    cfg := &api.Config{
        Mode:        config.ModeLib,
        StepTimeout: 7 * time.Minute,
        GPUType:     config.GPUTypeReal,
        Env:         config.EnvProd,
    }

    // Create installation data
    runaiData, err := api.NewRunAIInstallData(cfg)
    if err != nil {
        return fmt.Errorf("failed to create installation data: %w", err)
    }

    // Initialize global installer
    if err := api.Init(runaiData); err != nil {
        return fmt.Errorf("failed to initialize installer: %w", err)
    }

    // Get installer and start installation
    inst := api.Get()
    eventsChan, err := inst.RunInstallationWithEventListner()
    if err != nil {
        return fmt.Errorf("failed to start installation: %w", err)
    }

    // Wait for completion
    for event := range eventsChan {
        if event.FlowStatus == api.Failed {
            return fmt.Errorf("installation failed: %s", event.FlowName)
        }
    }

    return nil
}
```

## Thread Safety

The API is designed to be thread-safe:

- Global instance initialization uses `sync.Once`
- The `Init()` function is idempotent (safe to call multiple times, only initializes once)
- Multiple goroutines can safely call `Get()` after initialization
- Event channels can be read from any goroutine

## Error Handling

The API provides detailed error messages. Always check errors from:

- `Init()` - Returns error if initialization fails
- `NewRunAIInstallData()` - Returns error if configuration is invalid
- `inst.RunInstallationWithEventListner()` - Returns error if installation cannot start
- Installation event channels - Check for `Failed` and `Timeout` phases

## Advanced Usage

### Testing with Mock K8s Client

For testing purposes, you can use the underlying `installer` package directly to inject mock dependencies:

```go
import (
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/installer"
    "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/k8s/mocks"
)

func TestInstallation(t *testing.T) {
    // Create mock K8s client
    mockK8s := mocks.NewMockK8sClients(t)
    
    // Create installation data
    runaiData, _ := api.NewRunAIInstallData(cfg)
    
    // Use the installer package directly for testing
    err := installer.InitWithK8sClient(runaiData, mockK8s)
    // ... rest of test
}
```

Note: For production use, always use the `api` package functions.
