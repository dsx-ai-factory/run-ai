/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestHttpMiddleware_WhenContext_SetsRunaiTraceHeaders(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	spanId := [8]byte{1, 2}
	traceId := [16]byte{1, 2, 3}
	span := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId,
		SpanID:  spanId,
	})
	ctx := trace.ContextWithSpanContext(incomeReq.Context(), span)
	incomeReq = incomeReq.WithContext(ctx)

	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get(runaiTraceIdHeader), hex.EncodeToString(traceId[:]))
		assert.Equal(t, r.Header.Get(runaiSpanIdHeader), hex.EncodeToString(spanId[:]))
	})

	handler := HttpMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}

func TestHttpMiddleware_WhenNoContext_NotSetRunaiTraceHeaders(t *testing.T) {
	incomeReq := httptest.NewRequest("POST", "http://some.url.com", nil)
	recorder := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get(runaiTraceIdHeader), "")
		assert.Equal(t, r.Header.Get(runaiSpanIdHeader), "")
	})

	handler := HttpMiddleware(nextHandler)
	handler.ServeHTTP(recorder, incomeReq)
}
