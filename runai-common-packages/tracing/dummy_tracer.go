package tracing

import "context"

type DummyTracerProvider struct {
}

func (t *DummyTracerProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (t *DummyTracerProvider) Tracer() Tracer {
	return &DummyTracer{}
}

func (t *DummyTracerProvider) ForceFlush(ctx context.Context) error {
	return nil
}

type DummyTracer struct {
}

func (t *DummyTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, &DummySpan{}
}

func (t *DummyTracer) FromContext(ctx context.Context) Span {
	return &DummySpan{}
}

type DummySpan struct {
}

func (s *DummySpan) End() {
}

func (s *DummySpan) SetAttribute(key, value string) Span {
	return s
}

func (s *DummySpan) SetStatus(status SpanStatus, description string) Span {
	return s
}

func (s *DummySpan) RecordError(err error) Span {
	return s
}

func (s *DummySpan) GetSpanID() string {
	return ""
}
