/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	thttp "github.com/run-ai/runai/runai-common-packages/tracing/http"
	mocks "github.com/run-ai/runai/runai-common-packages/tracing/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// Helper function to start the mock OTLP collector.
func startMockCollector(address string) (*grpc.Server, *mocks.MockCollector) {
	server, mockCollector, err := mocks.StartMockCollector(address)
	if err != nil {
		log.Fatalf("Failed to start mock OTLP collector: %v", err)
	}
	return server, mockCollector
}

// Helper function to initialize the tracer provider.
func initializeTracer(t *testing.T, serviceName, address string) TracerProvider {
	syncOnce = sync.Once{}
	tracingRegistry = nil

	tp, err := Init(&Config{
		ServiceName:       serviceName,
		CollectorEndpoint: address,
		Enabled:           true,
		UnderlineTracer:   OtelTracerName,
	})
	assert.NoError(t, err, "expected no error initializing tracer")
	assert.NotNil(t, tp, "expected tracer provider to be non-nil")
	return tp
}

func TestInitTracer(t *testing.T) {
	address := "localhost:4317"
	server, _ := startMockCollector(address)
	defer server.GracefulStop()

	// Test tracer initialization.
	tp := initializeTracer(t, "test-service", address)
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err, "expected no error shutting down tracer provider")
}

func TestCreateSpan(t *testing.T) {
	address := "localhost:4317"
	server, mockCollector := startMockCollector(address)
	defer server.GracefulStop()

	// Initialize the tracer.
	tp := initializeTracer(t, "test-service", address)
	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err, "expected no error shutting down tracer provider")
	}()

	// Create a span.
	tracer := tp.Tracer()
	_, span := tracer.StartSpan(context.Background(), "test-span")
	span.End()

	// Force flush the spans.
	err := tp.ForceFlush(context.Background())
	assert.NoError(t, err, "expected no error flushing spans")

	// Verify the span was received by the mock collector.
	spans := mockCollector.GetSpans()
	assert.Len(t, spans, 1, "expected one span to be exported")
	assert.Equal(t, "test-span", spans[0].Name, "expected span name to match")
}

func TestShutdownTracer(t *testing.T) {
	tp := mocks.NewMockTracerProvider()
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err, "expected no error shutting down tracer provider")
	assert.True(t, tp.IsShutdownCalled(), "expected shutdown to be called")
}

func TestSpanFromContext(t *testing.T) {
	// Initialize the tracer provider with a mock collector.
	address := "localhost:4318"
	server, _ := startMockCollector(address)
	defer server.GracefulStop()

	// Initialize the tracer.
	tp := initializeTracer(t, "test-service", address)
	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err, "expected no error shutting down tracer provider")
	}()

	// Create a new span.
	tracer := tp.Tracer()
	ctx, span := tracer.StartSpan(context.Background(), "test-span")
	defer span.End()

	// Extract the span from the context.
	extractedSpan := tracer.FromContext(ctx)
	assert.NotNil(t, extractedSpan, "expected extracted span to be non-nil")
	assert.Equal(t, span, extractedSpan, "expected extracted span to match created span")
}

func TestSpanSetAttributes(t *testing.T) {
	// Start the mock OTLP collector.
	address := "localhost:4317"
	server, mockCollector := startMockCollector(address)
	defer server.GracefulStop()

	// Initialize the tracer.
	tp := initializeTracer(t, "test-service", address)
	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err, "expected no error shutting down tracer provider")
	}()

	// Create a span and set attributes.
	tracer := tp.Tracer()
	_, span := tracer.StartSpan(context.Background(), "test-span")
	span.SetAttribute("key1", "value1").SetAttribute("key2", "123")
	span.End()

	// Force flush the spans.
	time.Sleep(1 * time.Second)
	err := tp.ForceFlush(context.Background())
	assert.NoError(t, err, "expected no error flushing spans")

	// Verify the span was received by the mock collector.
	spans := mockCollector.GetSpans()
	assert.Len(t, spans, 1, "expected one span to be exported")
	assert.Equal(t, "test-span", spans[0].Name, "expected span name to match")

	// Verify the attributes.
	attributes := spans[0].Attributes
	assert.Len(t, attributes, 2, "expected two attributes to be set")
	assert.Equal(t, "value1", attributes[0].Value.GetStringValue(), "expected attribute key1 to match")
	assert.Equal(t, "123", attributes[1].Value.GetStringValue(), "expected attribute key2 to match")
}

func TestHTTPMiddleware(t *testing.T) {
	// Initialize the tracer provider with a mock collector.
	address := "localhost:4319"
	server, _ := startMockCollector(address)
	defer server.GracefulStop()

	// Initialize the tracer.
	tp := initializeTracer(t, "test-service", address)
	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err, "expected no error shutting down tracer provider")
	}()

	// Create an HTTP middleware.
	middleware := thttp.HttpMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler := middleware

	// Create a new request.
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/test", nil)
	rec := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "expected status code to be 200")
	assert.NotNil(t, rec.Header().Get("traceparent"), "expected traceparent header to be set")
}
