/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tenantId       = "1"
	tenantIdHeader = "X-Runai-Tenant-Id"
)

func resetVectors() {
	metrics.CounterVec(HttpRequestCount).Get().Reset()
	metrics.HistogramVec(HttpRequestDurationSeconds).Get().Reset()
}

func scrapeMetrics(t *testing.T) string {
	t.Helper()

	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	metricsResp, err := http.Get(metricsServer.URL)
	require.NoError(t, err)
	defer metricsResp.Body.Close()

	metricsBody, err := io.ReadAll(metricsResp.Body)
	require.NoError(t, err)
	return string(metricsBody)
}

func TestHttpMiddleware_OneRequest(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(HttpMiddleware)
		r.Method("GET", "/test/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/test/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	metricsData := scrapeMetrics(t)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	expectedCount200Metric := `runai_http_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
}

func TestHttpMiddleware_MultipleRequestsSameResponseDifferentPathParams(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(HttpMiddleware)
		r.Method("GET", "/test/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/test/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = makeRequestToServer(server.URL + "/test/2")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	metricsData := scrapeMetrics(t)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 2`
	expectedCount200Metric := `runai_http_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 2`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
}

func TestHttpMiddleware_MultipleRequestsDifferentResponses(t *testing.T) {
	metrics.Init(nil)

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
		}
		callCount++
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(HttpMiddleware)
		r.Method("GET", "/test/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/test/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp2, err := makeRequestToServer(server.URL + "/test/2")
	assert.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp2.StatusCode)

	metricsData := scrapeMetrics(t)

	// Check that the metrics include the expected observation
	expectedDuration200Metric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	expectedCount200Metric := `runai_http_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDuration200Metric)
	assert.Contains(t, metricsData, expectedCount200Metric)

	expectedDuration403Metric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="403",tenant_id="1",user_agent="go-http-client"} 1`
	expectedCount403Metric := `runai_http_request_count{method="GET",path="/test/{id}",status_code="403",tenant_id="1",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDuration403Metric)
	assert.Contains(t, metricsData, expectedCount403Metric)
}

func TestHttpMiddleware_MultipleRequestsDifferentPaths(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Forbidden"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(HttpMiddleware)
		r.Method("GET", "/test/{id}", handler)
		r.Method("GET", "/test/{id}/diff", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/test/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp2, err := makeRequestToServer(server.URL + "/test/1/diff")
	assert.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	metricsData := scrapeMetrics(t)

	// Check that the metrics include the expected observation
	expectedDurationFirstPathMetric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	expectedCountFirstPathMetric := `runai_http_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDurationFirstPathMetric)
	assert.Contains(t, metricsData, expectedCountFirstPathMetric)

	expectedDurationSecondPathMetric := `runai_http_request_duration_seconds_count{method="GET",path="/test/{id}/diff",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	expectedCountSecondPathMetric := `runai_http_request_count{method="GET",path="/test/{id}/diff",status_code="200",tenant_id="1",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDurationSecondPathMetric)
	assert.Contains(t, metricsData, expectedCountSecondPathMetric)
}

func TestHttpMiddleware_NoRouteContext(t *testing.T) {
	// Test that the middleware does not panic when there is no route context
	// We will use no router here, just plain middleware
	// We expect the 'path' label to be empty
	metrics.Init(nil)

	handler := HttpMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	resetVectors()

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	metricsData := scrapeMetrics(t)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_http_request_duration_seconds_count{method="GET",path="",status_code="200",tenant_id="",user_agent="go-http-client"} 1`
	expectedCount200Metric := `runai_http_request_count{method="GET",path="",status_code="200",tenant_id="",user_agent="go-http-client"} 1`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
}

func makeRequestToServer(url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(tenantIdHeader, tenantId)
	return client.Do(req)
}

func TestNormalizeUserAgent(t *testing.T) {
	assert.Equal(t, "go-http-client", normalizeUserAgent("Go-http-client/1.1"))
	assert.Equal(t, "go-http-client", normalizeUserAgent("Go-http-client 1.1"))
	assert.Equal(t, "go-http-client", normalizeUserAgent("Go-http-client"))
	assert.Equal(t, "k8s_client", normalizeUserAgent("k8s_client/1.0"))
	assert.Equal(t, "", normalizeUserAgent("this-is-a-very-long-user-agent-name/1.0"))
}
