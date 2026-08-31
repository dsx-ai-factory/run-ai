/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	HttpRequestDurationSeconds = "http_request_duration_seconds"
	HttpRequestCount           = "http_request_count"
	MethodLabel                = "method"
	PathLabel                  = "path"
	StatusCodeLabel            = "status_code"
	TenantIdLabel              = "tenant_id"
	UserAgentLabel             = "user_agent"
)

var (
	once sync.Once

	userAgentRegex = regexp.MustCompile(`^[\w-]{0,30}$`)
)

func HttpMiddleware(next http.Handler) http.Handler {
	once.Do(func() {
		metrics.HistogramVec(HttpRequestDurationSeconds).Register("Duration of HTTP requests", []string{MethodLabel, PathLabel, StatusCodeLabel, TenantIdLabel, UserAgentLabel}, []float64{0.1, 0.5, 1, 2, 5, 10})
		metrics.CounterVec(HttpRequestCount).Register("Number of HTTP requests", []string{MethodLabel, PathLabel, StatusCodeLabel, TenantIdLabel, UserAgentLabel})
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := NewResponseWriterWrapper(w)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		updateDurationMetric(ww, r, duration)
		updateCountMetric(ww, r)
	})
}

func updateDurationMetric(w *ResponseWriterWrapper, r *http.Request, duration time.Duration) {

	labels := map[string]string{
		MethodLabel:     r.Method,
		PathLabel:       getRoutePattern(r),
		StatusCodeLabel: strconv.Itoa(w.StatusCode),
		TenantIdLabel:   getTenantIdFromRequest(r),
		UserAgentLabel:  normalizeUserAgent(r.UserAgent()),
	}
	metrics.HistogramVec(HttpRequestDurationSeconds).Get().WithLabels(labels).Observe(duration.Seconds())
}

func updateCountMetric(w *ResponseWriterWrapper, r *http.Request) {
	labels := map[string]string{
		MethodLabel:     r.Method,
		PathLabel:       getRoutePattern(r),
		StatusCodeLabel: strconv.Itoa(w.StatusCode),
		TenantIdLabel:   getTenantIdFromRequest(r),
		UserAgentLabel:  normalizeUserAgent(r.UserAgent()),
	}
	metrics.CounterVec(HttpRequestCount).Get().WithLabels(labels).Inc()
}

func getRoutePattern(r *http.Request) string {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		return routeCtx.RoutePattern()
	}
	return ""
}

func getTenantIdFromRequest(r *http.Request) string {
	tenantId, _ := runaictx.ParseTenantId(r.Context())
	return tenantId
}

func normalizeUserAgent(ua string) string {
	product := extractUserAgentProduct(ua)
	if !userAgentRegex.MatchString(product) {
		return ""
	}
	return strings.ToLower(product)
}

func extractUserAgentProduct(ua string) string {
	for i, c := range ua {
		if c == '/' || c == ' ' {
			return ua[:i]
		}
	}
	return ua
}
