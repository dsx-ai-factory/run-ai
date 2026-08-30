package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// httpHandler adds OpenTelemetry tracing to an HTTP handler.
func httpHandler(next http.Handler) http.Handler {
	handler := otelhttp.NewHandler(next, "api")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

// httpTraceToHeadersHandler adds OpenTelemetry tracing to an HTTP handler and saves the tracing headers in the context.
func httpTraceToHeadersHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		setRunaictxHeaders(r, span)
		ctx = saveTracingHeadersInCtx(r, ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HttpMiddleware adds OpenTelemetry tracing to an HTTP handler and saves the tracing headers in the context.
func HttpMiddleware(next http.Handler) http.Handler {
	return httpHandler(httpTraceToHeadersHandler(next))
}
