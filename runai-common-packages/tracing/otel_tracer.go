package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Tracer implements the Tracer interface.
type OtelTracer struct {
	tracer trace.Tracer
}

// StartSpan starts a new span.
func (t *OtelTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, &OtelSpan{span: span}
}

// FromContext returns the span from the context.
func (t *OtelTracer) FromContext(ctx context.Context) Span {
	span := trace.SpanFromContext(ctx)
	return &OtelSpan{span: span}
}
