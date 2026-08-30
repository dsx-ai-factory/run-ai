# Controller Runtime Metrics Extension

This package provides integration between the sigs.k8s.io/controller-runtime metrics and the Run:AI metrics package.
This extension allows controller-runtime metrics and your service's metrics to be registered to the same registry.

## Usage

To use this extension in a service that uses controller-runtime:

1. Import the package:
```go
import "github.com/run-ai/runai/runai-common-packages/metricsctr"
```

2. Initialize the metrics package and then the metricsctr package (metrics package should be initialized first):
```go
// Initialize the metrics package
_, err := metrics.Init(&metrics.Config{
    ServiceName: "your-service",
    ExposePort: 9090,
})
if err != nil {
    // handle error
}

// Initialize controller-runtime metrics to use our registry
metricsctr.Init()
```
