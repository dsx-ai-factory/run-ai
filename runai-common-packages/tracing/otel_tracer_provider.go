/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OtelTracerProvider implements the OtelTracerProvider interface.
type OtelTracerProvider struct {
	Provider  *sdktrace.TracerProvider
	TracerObj trace.Tracer
}

func NewOtelTracerProvider(config *Config) (*OtelTracerProvider, error) {
	ctx := context.Background()

	// Set up a connection to the collector.
	conn, err := grpc.NewClient(config.CollectorEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Create the gRPC trace exporter.
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// Create the trace provider with the exporter.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
		)),
		sdktrace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return &OtelTracerProvider{
		Provider:  tp,
		TracerObj: tp.Tracer(config.ServiceName),
	}, nil
}

// ForceFlush flushes the tracer.
func (t *OtelTracerProvider) ForceFlush(ctx context.Context) error {
	return t.Provider.ForceFlush(ctx)
}

// Tracer returns an instance of Tracer.
func (t *OtelTracerProvider) Tracer() Tracer {
	return &OtelTracer{tracer: t.TracerObj}
}

// Shutdown shuts down the tracer provider.
func (t *OtelTracerProvider) Shutdown(ctx context.Context) error {
	return t.Provider.Shutdown(ctx)
}

// initOtelPropagation initializes the basic propagator for OpenTelemetry.
// This allows us to retrieve traceId and spanId in the headers regardless of the tracer provider used.
func initOtelPropagation() {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}
