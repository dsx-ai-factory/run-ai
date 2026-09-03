# Run:AI Installer Library

The Run:AI Installer Library provides a programmatic interface to the Run:AI installation process. It can be used by:

- **TUI (Terminal User Interface)**: Interactive installation with user confirmation between steps
- **API Services**: Automated installation without user interaction
- **CLI Tools**: Command-line installation scripts
- **Automation Tools**: CI/CD pipelines and orchestration systems

## Features

- **Programmatic Installation**: Full control over the installation process via Go API
- **User Confirmation Mode**: Optional user confirmation between installation screens (like TUI)
- **Automatic Mode**: Fully automated installation without user interaction (for APIs/scripts)
- **Progress Callbacks**: Real-time progress updates during installation
- **Validation**: Configuration validation before installation begins
- **Error Handling**: Comprehensive error reporting and recovery
- **Context Support**: Cancellable operations using Go context

## Installation Flow

The library orchestrates the following installation screens:

1. **Prerequisites Selection**: Validates selected prerequisites
2. **Configuration**: Validates configuration inputs (FQDN, credentials, etc.)
3. **Input Validation**: Deep validation of all inputs with prerequisites
4. **Prerequisite Installation**: Installs selected prerequisites (Ingress, Prometheus, etc.)
5. **Run:AI Installation**: Installs Run:AI Control Plane and Cluster

## Usage

### Basic Usage (Automatic Mode)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/dsx-ai-factory/run-ai/runai-installer/pkg/common"
    "github.com/dsx-ai-factory/run-ai/runai-installer/pkg/installer"
    configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/pkg/inputs"
    "github.com/dsx-ai-factory/run-ai/runai-installer/pkg/installations/prereq"
)

func main() {
    // Create configuration
    config := installer.NewInstallConfig()

    // Set prerequisites
    config.Prerequisites = []string{
        prereq.PrerequisiteIdentifierIngress,    // Mandatory
        prereq.PrerequisiteIdentifierPrometheus, // Optional
    }

    // Set configuration inputs
    config.Inputs = map[configinputs.InstallInputType]string{
        configinputs.InstallInputJfrogCreds:  "base64encodedcredentials==",
        configinputs.InstallInputFQDN:        "runai.example.com",
        configinputs.InstallInputClusterName: "my-cluster",
        configinputs.InstallInputSSLPath:     "/path/to/cert.pem",
        configinputs.InstallInputSSLKeyPath:  "/path/to/key.pem",
    }

    // Set options
    config.Options = &common.Options{
        StepTimeout: 7 * time.Minute,
        GPUType:     common.GPUTypeReal,
        Env:         common.EnvProd,
    }

    // No user confirmation - runs automatically
    config.UserConfirmation = false

    // Set progress callback
    config.ProgressCallback = func(screenName string, stepName string, progress int, message string) {
        fmt.Printf("[%d%%] %s: %s\n", progress, screenName, message)
    }

    // Create installer
    inst, err := installer.NewInstaller(config)
    if err != nil {
        panic(err)
    }

    // Run installation
    result, err := inst.RunInstallation(context.Background())
    if err != nil {
        panic(err)
    }

    // Check results
    if result.Success {
        fmt.Printf("Installation completed in %v\n", result.Duration)
    } else {
        fmt.Printf("Installation failed: %v\n", result.Errors)
    }
}
```

### Interactive Mode with User Confirmation

```go
func main() {
    config := installer.NewInstallConfig()
    
    // ... set prerequisites and inputs ...

    // Enable user confirmation
    config.UserConfirmation = true

    // Set confirmation callback
    config.ConfirmCallback = func(screenName string, data *installer.ScreenData) (bool, error) {
        fmt.Printf("\nAbout to execute: %s\n", screenName)
        fmt.Printf("Summary: %s\n", data.Summary)
        
        // Prompt user
        fmt.Print("Continue? (y/n): ")
        var response string
        fmt.Scanln(&response)
        
        return response == "y" || response == "Y", nil
    }

    // Set progress callback
    config.ProgressCallback = func(screenName string, stepName string, progress int, message string) {
        fmt.Printf("[%d%%] %s: %s - %s\n", progress, screenName, stepName, message)
    }

    inst, err := installer.NewInstaller(config)
    if err != nil {
        panic(err)
    }

    result, err := inst.RunInstallation(context.Background())
    // ... handle result ...
}
```

### API Usage with Background Execution

```go
// API endpoint handler
func handleInstallRequest(w http.ResponseWriter, r *http.Request) {
    // Parse request body into config
    var req InstallRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Create installer config
    config := installer.NewInstallConfig()
    config.Prerequisites = req.Prerequisites
    config.Inputs = req.Inputs
    config.Options = req.Options
    config.UserConfirmation = false // API mode - no user interaction

    // Store progress updates
    jobID := generateJobID()
    config.ProgressCallback = func(screenName string, stepName string, progress int, message string) {
        updateJobProgress(jobID, progress, message)
    }

    // Create installer
    inst, err := installer.NewInstaller(config)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Run installation in background
    go func() {
        result, err := inst.RunInstallation(context.Background())
        if err != nil {
            updateJobStatus(jobID, "failed", err.Error())
        } else if result.Success {
            updateJobStatus(jobID, "completed", "")
        } else {
            updateJobStatus(jobID, "failed", strings.Join(result.Errors, ", "))
        }
    }()

    // Return job ID for polling
    json.NewEncoder(w).Encode(map[string]string{
        "job_id": jobID,
        "status": "running",
    })
}
```

### Context Cancellation

```go
func main() {
    config := installer.NewInstallConfig()
    // ... configure ...

    inst, err := installer.NewInstaller(config)
    if err != nil {
        panic(err)
    }

    // Create cancellable context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    // Handle interrupt signal
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan
        fmt.Println("\nReceived interrupt signal, cancelling installation...")
        cancel()
    }()

    // Run installation with cancellable context
    result, err := inst.RunInstallation(ctx)
    if err == context.Canceled {
        fmt.Println("Installation was cancelled")
    } else if err != nil {
        fmt.Printf("Installation failed: %v\n", err)
    } else if result.Success {
        fmt.Println("Installation completed successfully")
    }
}
```

## Configuration

### InstallConfig

| Field | Type | Description |
|-------|------|-------------|
| `Prerequisites` | `[]string` | List of prerequisites to install (use `prereq.PrerequisiteIdentifier*` constants) |
| `Inputs` | `map[InstallInputType]string` | Configuration inputs (FQDN, credentials, SSL paths, etc.) |
| `Options` | `*common.Options` | Installation options (timeouts, GPU type, environment) |
| `UserConfirmation` | `bool` | If `true`, waits for user confirmation between screens; if `false`, runs automatically |
| `ConfirmCallback` | `func(string, *ScreenData) (bool, error)` | Called before each screen when `UserConfirmation` is `true` |
| `ProgressCallback` | `func(string, string, int, string)` | Called to report progress (screen name, step name, progress %, message) |

### Prerequisites

Available prerequisites (use constants from `pre_req` package):

- `PrerequisiteIdentifierIngress` - NGINX Ingress Controller (Mandatory)
- `PrerequisiteIdentifierPrometheus` - Prometheus monitoring (Optional)
- `PrerequisiteIdentifierGpuOperator` - NVIDIA GPU Operator (Optional)
- `PrerequisiteIdentifierNetworkOperator` - NVIDIA Network Operator (Optional)
- `PrerequisiteIdentifierKnative` - Knative Serving (Optional)
- `PrerequisiteIdentifierTrainingOperator` - Kubeflow Training Operator (Optional)
- `PrerequisiteIdentifierLWS` - LeaderWorkerSet for distributed inference (Optional)

### Configuration Inputs

Required inputs (use constants from `configinputs` package):

- `InstallInputJfrogCreds` - JFROG credentials (base64 encoded)
- `InstallInputFQDN` - Fully Qualified Domain Name
- `InstallInputClusterName` - Kubernetes cluster name

Optional inputs:

- `InstallInputIngressControllerHTTPSPort` - Custom HTTPS port (30000-32767)
- `InstallInputSSLPath` - Path to SSL certificate
- `InstallInputSSLKeyPath` - Path to SSL private key
- `InstallInputCAChainPath` - Path to CA chain certificate
- `InstallInputWildcardFQDN` - Wildcard FQDN for inference
- `InstallInputWildcardCertPath` - Wildcard SSL certificate path
- `InstallInputWildcardCertKeyPath` - Wildcard SSL key path

## Validation

The library provides comprehensive validation:

```go
config := installer.NewInstallConfig()
// ... configure ...

// Validate configuration before creating installer
result := config.Validate()
if !result.Valid {
    fmt.Println("Configuration errors:")
    for inputType, errMsg := range result.Errors {
        fmt.Printf("  %v: %s\n", inputType, errMsg)
    }
}

if len(result.Warnings) > 0 {
    fmt.Println("Configuration warnings:")
    for _, warning := range result.Warnings {
        fmt.Printf("  - %s\n", warning)
    }
}
```

## Error Handling

The library provides detailed error reporting:

```go
result, err := inst.RunInstallation(ctx)

if err != nil {
    // Fatal error that stopped installation
    fmt.Printf("Installation error: %v\n", err)
}

if result != nil {
    if !result.Success {
        // Installation completed but with errors
        fmt.Println("Installation errors:")
        for _, errMsg := range result.Errors {
            fmt.Printf("  - %s\n", errMsg)
        }
    }

    if len(result.Warnings) > 0 {
        // Non-fatal warnings
        fmt.Println("Warnings:")
        for _, warning := range result.Warnings {
            fmt.Printf("  - %s\n", warning)
        }
    }
}
```

## TUI Integration

The TUI can use this library for installation:

```go
// Collect configuration from TUI screens
config := collectConfigFromTUIScreens()

// Enable user confirmation for screen transitions
config.UserConfirmation = true

// Use TUI to show confirmation dialogs
config.ConfirmCallback = func(screenName string, data *installer.ScreenData) (bool, error) {
    return showTUIConfirmationDialog(screenName, data)
}

// Use TUI to show progress
config.ProgressCallback = func(screenName string, stepName string, progress int, message string) {
    updateTUIProgress(screenName, stepName, progress, message)
}

inst, _ := installer.NewInstaller(config)
result, _ := inst.RunInstallation(context.Background())
```

## Best Practices

1. **Always validate configuration** before creating installer
2. **Use context with timeout** to prevent hanging operations
3. **Handle cancellation** gracefully with context
4. **Set appropriate progress callbacks** for monitoring
5. **Log all operations** using the logging package
6. **Check result.Success** even if err is nil
7. **Review warnings** as they may indicate configuration issues

## Thread Safety

The installer is **not thread-safe**. Do not share an `Installer` instance across goroutines. Create a new instance for each installation operation.

## Examples

See `example.go` for complete examples of:
- Library usage without user confirmation (automatic mode)
- Library usage with user confirmation (interactive mode)
- API usage with background execution

## Migration from TUI

If you're migrating from the TUI to library usage:

1. Extract configuration from TUI screens into `InstallConfig`
2. Set `UserConfirmation = false` for automatic operation
3. Implement progress callbacks to replace TUI progress display
4. Handle errors and results instead of relying on TUI error screens

## Future Enhancements

- [ ] Dry-run mode for validation without installation
- [ ] Resume capability for interrupted installations
- [ ] Rollback support for failed installations
- [ ] Detailed step-by-step logs in InstallResult
- [ ] Custom prerequisite plugins
- [ ] Installation templates/presets

