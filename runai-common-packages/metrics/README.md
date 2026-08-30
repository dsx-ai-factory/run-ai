## Metrics Library

The Metrics library is a Go package for easily instrumenting your Go applications with Prometheus metrics. It provides a convenient way to register and expose custom metrics, as well as default Go runtime metrics.

## Features

- Register and expose custom metrics such as counters, gauges, histograms, and summaries.
- Automatically include default Go runtime metrics.
- Simple API for adding labels to metrics.
- HTTP handler for exposing metrics via an HTTP endpoint.
- Push metrics to Prometheus Pushgateway.

## Installation

To use the Metrics library in your Go project, you can simply install it using go get:

```bash
go get github.com/run-ai/runai/runai-common-packages/metrics
```

## Usage

Importing the Library

```go
import "github.com/run-ai/runai/runai-common-packages/metrics"
```

## Initializing Metrics

To initialize the Metrics instance, simply call the NewMetrics function:

```go
metrics.Init()
```

## Registering Metrics

Counter

```go
metrics.Counter("my_counter").Register("A custom counter", nil)
```

Gauge

```go
gauge := metrics.Gauge("my_gauge").Register("A custom gauge", nil)
```

Histogram

```go
metrics.Histogram("my_histogram").Register("A custom histogram", nil, []float64{0.1, 1, 10})
```

Summary

```go
metrics.Summary("my_summary").Register("A custom summary", nil, map[float64]float64{0.5: 0.05, 0.9: 0.01})
```

Retrieve a metric from the registry is possible using the Get method:
```go
counter := metrics.Counter("my_counter").Get()
counter.Inc()
```

## Pushgateway Integration

To use Push logic and register metrics to the Pushgateway, use Pushable(), which returns a MetricsRegistrar that holds the metrics locally. After registering metrics to the registry, use the Push() function to push the metrics to the Pushgateway.

When using Pushgateway, make sure you provide the following environment variables:

`SERVICE_NAME`: This identifies the name of the service and will be used as the job name in Pushgateway.

`METRICS_PUSH_GATEWAY_URL`: The URL of the Prometheus Pushgateway where metrics will be pushed.

Registering Metrics to Pushgateway:

Counter

```go
metrics.Pushable().Counter("my_counter").Register("A custom counter", nil)
```

Gauge

```go
gauge := metrics.Pushable().Gauge("my_gauge").Register("A custom gauge", nil)
```

Histogram

```go
metrics.Pushable().Histogram("my_histogram").Register("A custom histogram", nil, []float64{0.1, 1, 10})
```

Summary

```go
metrics.Pushable().Summary("my_summary").Register("A custom summary", nil, map[float64]float64{0.5: 0.05, 0.9: 0.01})
```

Push Metrics to Pushgateway after registering your metrics:

**Important:** The usage of `Push` function is required in order to send your metrics to Pushgateway!

```go
metrics.Push()
```

## Examples

Basic:

```go
package main

import (
    "net/http"
    "github.com/run-ai/runai/runai-common-packages/metrics"
    httpmetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
)

func main() {
    metrics.Init(nil)
    // Register custom metrics
    metrics.Counter("requests_total").Register("Total number of requests received.", nil)

    // Start HTTP server
    http.Handle("/", http.HandlerFunc(yourHandler))
    http.ListenAndServe(":8080", httpmetrics.HttpMiddleware(http.DefaultServeMux))
}

func yourHandler(w http.ResponseWriter, r *http.Request) {
    // Increment the counter
    metrics.Counter("requests_total").Get().Inc()
	w.Write([]byte("Hello, world!"))
}
```

Using Pushgateway:

```go
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/run-ai/runai/runai-common-packages/metrics"
	httpmetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
)

func main() {
	metrics.Init(nil)
	// Register Pushgateway metrics
	metrics.Pushable().Counter("push_requests_total").Register("Total number of requests pushed.", nil)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChannel
		fmt.Println("Service shutting down... making last push of metrics.")
		if err := metrics.Push(); err != nil {
			fmt.Printf("Failed to push final metrics to Pushgateway: %v\n", err)
		} else {
			fmt.Println("Final metrics pushed successfully.")
		}
		os.Exit(0)
	}()

	// Start HTTP server
	http.Handle("/", http.HandlerFunc(yourHandler))
	http.ListenAndServe(":8080", httpmetrics.HttpMiddleware(http.DefaultServeMux))
}

func yourHandler(w http.ResponseWriter, r *http.Request) {
	// Increment the counter
	metrics.Pushable().Counter("push_requests_total").Get().Inc()
	err := metrics.Push() // Push metrics to Pushgateway
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to push metrics to Pushgateway: %v", err), http.StatusInternalServerError)
	}
	w.Write([]byte("Hello, world!"))
}
```

## Testing

The Metrics library provides a mock implementation for testing purposes. This allows you to validate that your code interacts with the metrics library as expected without actually registering or exposing metrics.

### Key Points:

1. **Metrics Ignored Without Initialization**:
   - If you do not call `metrics.Init`, all metrics calls in the code will be ignored. This ensures that metrics are only active when explicitly initialized.

2. **Mocking Metrics with `SetMock`**:
   - By calling `testing.SetMock(t)`, you get back a mock `MetricsRegistrar` instance. This mock instance allows you to validate that the same calls made in your code are expected during testing.
   - Each call you make on the mock instance will be expected in the code, ensuring consistency between your tests and the actual implementation.

### Examples:

#### Ignoring Metrics Without Initialization
If `metrics.Init` is not called, the following code will not register or expose any metrics:

```go
metrics.Counter("my_counter").Get().Inc() // This call will be ignored if metrics.Init is not called.
```

#### Using `SetMock` for Testing
You can use SetMock to validate that your code interacts with the metrics library as expected. Here's an example:

```go
package testing

import (
    "testing"

    "github.com/run-ai/runai/runai-common-packages/metrics"
)

func TestGaugeSet(t *testing.T) {
    // Set up the mock
    manager := SetMock(t)

    // Expect the following calls in the code
    manager.Gauge("test_gauge").Get().Set(42)
    manager.Gauge("test_gauge").Register("A test gauge", nil).Set(42)

    // Simulate the actual code
    metrics.Gauge("test_gauge").Get().Set(42)
    metrics.Gauge("test_gauge").Register("A test gauge", nil).Set(42)
}
```

In this example:

* SetMock(t) sets up a mock MetricsRegistrar.
* The test validates that the Gauge metric is registered and updated as expected.

#### Summary
* Use `metrics.Init` to enable metrics in your application.
* Use `testing.SetMock(t)` to mock metrics during testing and validate interactions with the metrics library.
* The mock implementation ensures that your tests are consistent with the actual metrics calls in your code.


## Contributing

Contributions are welcome! If you have suggestions for improvements or find any issues, please open an issue or submit a pull request.
