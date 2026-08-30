package http

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	runaiTraceIdHeader = "x-runai-trace-id"
	runaiSpanIdHeader  = "x-runai-span-id"
)

var tracingHeaders = []string{
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-sampled",
	"x-b3-parentspanid",
	"x-b3-flags",
	"x-ot-span-context",
	"tracestate",
	"traceparent",
}

func SetHeaders(ctx context.Context, r *http.Request) {
	for _, headerName := range tracingHeaders {
		if value, ok := ctx.Value(contextKey(headerName)).(string); ok && value != "" {
			r.Header.Add(headerName, value)
		}
	}
}

func saveTracingHeadersInCtx(r *http.Request, ctx context.Context) context.Context {
	rctx := ctx
	for _, headerName := range tracingHeaders {
		if value := r.Header.Get(headerName); value != "" {
			rctx = context.WithValue(rctx, contextKey(headerName), value)
		}
	}
	return rctx
}

func setRunaictxHeaders(r *http.Request, span trace.Span) {
	if span.SpanContext().TraceID().IsValid() {
		r.Header.Set(runaiTraceIdHeader, span.SpanContext().TraceID().String())
	}
	if span.SpanContext().SpanID().IsValid() {
		r.Header.Set(runaiSpanIdHeader, span.SpanContext().SpanID().String())
	}
}
