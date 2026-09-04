/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	methodKey     = "method"
	pathKey       = "path"
	userAgentKey  = "user_agent"
	statusCodeKey = "status_code"
	durationKey   = "duration"
)

// loggingResponseWriter is a wrapper around http.ResponseWriter that stores the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// newLoggingResponseWriter creates a new loggingResponseWriter.
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, &bytes.Buffer{}}
}

// WriteHeader writes the status code to the response writer.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body.
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

// HttpMiddleware adds zerolog logging to an HTTP handler.
func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := newLoggingResponseWriter(w)

		defer func() {
			panicVal := recover()
			if panicVal != nil {
				lrw.statusCode = http.StatusInternalServerError // ensure that the status code is updated
				panic(panicVal)
			}

			l := logging.WithCtx(runaictx.LoggerHook(r.Context()))

			zerologger := l.Actual().(zerolog.Logger)
			zerologger.Debug().Str(methodKey, r.Method).Str(pathKey, r.URL.Path).Str(userAgentKey, r.UserAgent()).Int(statusCodeKey, lrw.statusCode).Dur(durationKey, time.Since(start)).Msg("incoming request")
			if lrw.statusCode >= 400 {
				zerologger.Debug().Int(statusCodeKey, lrw.statusCode).Msgf("Error response: %s", lrw.body.String())
			}
		}()

		next.ServeHTTP(lrw, r)
	})
}
