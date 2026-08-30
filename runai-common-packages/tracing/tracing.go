package tracing

import (
	"context"
)

// TracerProvider defines an interface for the tracer provider.
type TracerProvider interface {
	Shutdown(ctx context.Context) error
	Tracer() Tracer
	ForceFlush(ctx context.Context) error
}

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
	FromContext(ctx context.Context) Span
}

type Span interface {
	End()
	SetAttribute(key, value string) Span
	SetStatus(status SpanStatus, description string) Span
	RecordError(err error) Span
	GetSpanID() string
}
