# runai-tracing

`tracing` is a Go library that wraps the OpenTelemetry client to provide auto instrumentation for multiple Go services. It simplifies the setup and usage of OpenTelemetry for tracing in Go applications and supports exporting spans to an external collector using gRPC.

## Features

- Easy setup for OpenTelemetry tracing in Go services
- Middleware for HTTP server auto-instrumentation
- Support for exporting spans to an OpenTelemetry Collector via gRPC
- Simple, interface-based design for ease of use and testing

## Installation

To install the library, run:

```sh
go get github.com/run-ai/runai/runai-common-packages/tracing
```

## Usage

To use the library, import it in your Go code:

```go
import "github.com/run-ai/runai/runai-common-packages/tracing"
```

Then, create a new `Tracer` instance and use it to instrument your code:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/tracing"
	trachinghttp "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

func main() {
	tracing.Init(nil)
	defer func() {
		if err := tracing.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown tracer: %v", err)
		}
	}()

	http.Handle("/", http.HandlerFunc(yourHandler))
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", trachinghttp.HttpMiddleware(http.DefaultServeMux)))
}

func yourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
```
** Note: the above example uses a collector agent running on localhost:4317. You can replace this with the address of your OpenTelemetry Collector.

## Using the OpenTelemetry Mock Collector

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/tracing"
	trachinghttp "github.com/run-ai/runai/runai-common-packages/tracing/http"
	"github.com/run-ai/runai/runai-common-packages/tracing/mock"
	"google.golang.org/grpc"
)

var (
	server        *grpc.Server
	mockCollector *mocks.MockCollector
	err           error
)

func main() {
	// Start the mock OTLP collector.
	address := "localhost:4317"
	server, mockCollector, err = mocks.StartMockCollector(address)
	if err != nil {
		log.Fatalf("Failed to start mock OTLP collector: %v", err)
	}
	defer server.GracefulStop()
	
	// Initialize tracer with the OpenTelemetry Collector endpoint.
	tracing.Init(nil)
	defer func() {
		if err := tracing.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown tracer: %v", err)
		}
	}()

	http.Handle("/", http.HandlerFunc(yourHandler))
	http.Handle("/get-spans", http.HandlerFunc(getSpansHandler))
	http.Handle("/create-span", http.HandlerFunc(createSpanHandler))
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", trachinghttp.HttpMiddleware(http.DefaultServeMux)))
}

func yourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}

func createSpanHandler(w http.ResponseWriter, r *http.Request) {
	spanName := r.URL.Query().Get("name")
	if spanName == "" {
		w.Write([]byte("Span name is required"))
		return
	}
	createSpan(spanName, r.Context())
	w.Write([]byte("Span created"))
}

func createSpan(name string, ctx context.Context) {
	tracer := tracing.GetTracer()
	_, span := tracer.StartSpan(ctx, name)
	span.SetAttribute("key", "value")
	span.SetAttribute("key2", "value2")
	span.End()
}

func getSpansHandler(w http.ResponseWriter, r *http.Request) {
	spans := mockCollector.GetSpans()
	for _, span := range spans {
		w.Write([]byte(span.String()))
	}
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

