package mocks

import (
	"context"
	"log"
	"net"
	"sync"

	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
)

// MockCollector implements the OTLP collector gRPC service.
type MockCollector struct {
	collectortrace.UnimplementedTraceServiceServer
	mu    sync.Mutex
	spans []*tracepb.Span
}

// NewMockCollector creates a new MockCollector.
func NewMockCollector() *MockCollector {
	return &MockCollector{}
}

// Export is a mock implementation of the gRPC Export method.
func (mc *MockCollector) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, resourceSpans := range req.ResourceSpans {
		for _, scopeSpans := range resourceSpans.ScopeSpans {
			mc.spans = append(mc.spans, scopeSpans.Spans...)
		}
	}
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

// GetSpans returns the collected spans.
func (mc *MockCollector) GetSpans() []*tracepb.Span {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.spans
}

// StartMockCollector starts the mock OTLP collector.
func StartMockCollector(address string) (*grpc.Server, *MockCollector, error) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", address)
	if err != nil {
		return nil, nil, err
	}

	mockCollector := NewMockCollector()
	server := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(server, mockCollector)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to start mock OTLP collector: %v", err)
		}
	}()

	return server, mockCollector, nil
}
