/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connection

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	RequestDurationSeconds = "connection_request_duration_seconds"
	RequestCount           = "connection_request_count"
	ActiveConnection       = "connection_active_count"
	MethodLabel            = "method"
	PathLabel              = "path"
	StatusCodeLabel        = "status_code"
	TenantIdLabel          = "tenant_id"
)

var once sync.Once

func MetricMiddleware(next http.Handler) http.Handler {
	once.Do(func() {
		metrics.HistogramVec(RequestDurationSeconds).Register("Duration of connection requests", []string{MethodLabel, PathLabel, StatusCodeLabel, TenantIdLabel}, []float64{0.1, 0.5, 1, 5, 10, 600, 3600, 86400})
		metrics.CounterVec(RequestCount).Register("Number of connection requests", []string{MethodLabel, PathLabel, StatusCodeLabel, TenantIdLabel})
		metrics.GaugeVec(ActiveConnection).Register("Number of connection active", []string{PathLabel, TenantIdLabel})
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := NewResponseWriterWrapper(w)

		path := getRoutePattern(r)
		tenantId := getTenantIdFromRequest(r)

		incrementActive(path, tenantId)
		defer decrementActive(path, tenantId)
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
	}
	metrics.HistogramVec(RequestDurationSeconds).Get().WithLabels(labels).Observe(duration.Seconds())
}

func updateCountMetric(w *ResponseWriterWrapper, r *http.Request) {
	labels := map[string]string{
		MethodLabel:     r.Method,
		PathLabel:       getRoutePattern(r),
		StatusCodeLabel: strconv.Itoa(w.StatusCode),
		TenantIdLabel:   getTenantIdFromRequest(r),
	}
	metrics.CounterVec(RequestCount).Get().WithLabels(labels).Inc()
}

func incrementActive(path, tenantId string) {
	labels := map[string]string{
		PathLabel:     path,
		TenantIdLabel: tenantId,
	}
	metrics.GaugeVec(ActiveConnection).Get().WithLabels(labels).Inc()
}

func decrementActive(path, tenantId string) {
	labels := map[string]string{
		PathLabel:     path,
		TenantIdLabel: tenantId,
	}
	metrics.GaugeVec(ActiveConnection).Get().WithLabels(labels).Dec()
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
