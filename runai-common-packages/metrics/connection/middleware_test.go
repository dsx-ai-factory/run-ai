package connection

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/runaimetrics"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	"github.com/stretchr/testify/assert"
)

const (
	tenantId       = "1"
	tenantIdHeader = "X-Runai-Tenant-Id"
)

func resetVectors() {
	metrics.CounterVec(RequestCount).Get().(*runaimetrics.CounterVec).PromCounterVec.Reset()
	metrics.HistogramVec(RequestDurationSeconds).Get().(*runaimetrics.HistogramVec).PromHistogramVec.Reset()
	metrics.GaugeVec(ActiveConnection).Get().(*runaimetrics.GaugeVec).PromGaugeVec.Reset()
}

func TestConnectionMiddleware_OneRequest(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(MetricMiddleware)
		r.Method("GET", "/test/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/test/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsResp.Body.Read(metricsBody)
	metricsData := string(metricsBody)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedCount200Metric := `runai_connection_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedActiveMetric := `connection_active_count{path="/test/{id}",tenant_id="1"} 0`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
	assert.Contains(t, metricsData, expectedActiveMetric)
}

func TestConnectionMiddleware_MultipleRequestsSameResponseDifferentPathParams(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(MetricMiddleware)
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

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsResp.Body.Read(metricsBody)
	metricsData := string(metricsBody)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 2`
	expectedCount200Metric := `runai_connection_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 2`
	expectedActiveetric := `connection_active_count{path="/test/{id}",tenant_id="1"} 0`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
	assert.Contains(t, metricsData, expectedActiveetric)
}

func TestConnectionMiddleware_MultipleRequestsDifferentResponses(t *testing.T) {
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
		r.Use(MetricMiddleware)
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

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsResp.Body.Read(metricsBody)
	metricsData := string(metricsBody)

	// Check that the metrics include the expected observation
	expectedDuration200Metric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedCount200Metric := `runai_connection_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedActiveMetric := `connection_active_count{path="/test/{id}",tenant_id="1"} 0`
	assert.Contains(t, metricsData, expectedDuration200Metric)
	assert.Contains(t, metricsData, expectedCount200Metric)
	assert.Contains(t, metricsData, expectedActiveMetric)

	expectedDuration403Metric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="403",tenant_id="1"} 1`
	expectedCount403Metric := `runai_connection_request_count{method="GET",path="/test/{id}",status_code="403",tenant_id="1"} 1`
	assert.Contains(t, metricsData, expectedDuration403Metric)
	assert.Contains(t, metricsData, expectedCount403Metric)
	assert.Contains(t, metricsData, expectedActiveMetric)
}

func TestConnectionMiddleware_MultipleRequestsDifferentPaths(t *testing.T) {
	metrics.Init(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Forbidden"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(MetricMiddleware)
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

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsResp.Body.Read(metricsBody)
	metricsData := string(metricsBody)

	// Check that the metrics include the expected observation
	expectedDurationFirstPathMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedCountFirstPathMetric := `runai_connection_request_count{method="GET",path="/test/{id}",status_code="200",tenant_id="1"} 1`
	expectedActiveMetric := `connection_active_count{path="/test/{id}",tenant_id="1"} 0`
	assert.Contains(t, metricsData, expectedDurationFirstPathMetric)
	assert.Contains(t, metricsData, expectedCountFirstPathMetric)
	assert.Contains(t, metricsData, expectedActiveMetric)

	expectedDurationSecondPathMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/test/{id}/diff",status_code="200",tenant_id="1"} 1`
	expectedCountSecondPathMetric := `runai_connection_request_count{method="GET",path="/test/{id}/diff",status_code="200",tenant_id="1"} 1`
	assert.Contains(t, metricsData, expectedDurationSecondPathMetric)
	assert.Contains(t, metricsData, expectedCountSecondPathMetric)
	assert.Contains(t, metricsData, expectedActiveMetric)
}

func TestConnectionMiddleware_NoRouteContext(t *testing.T) {
	// Test that the middleware does not panic when there is no route context
	// We will use no router here, just plain middleware
	// We expect the 'path' label to be empty
	metrics.Init(nil)

	handler := MetricMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsResp.Body.Read(metricsBody)
	metricsData := string(metricsBody)

	// Check that the metrics include the expected observation
	expectedDurationMetric := `runai_connection_request_duration_seconds_count{method="GET",path="",status_code="200",tenant_id=""} 1`
	expectedCount200Metric := `runai_connection_request_count{method="GET",path="",status_code="200",tenant_id=""} 1`
	expectedActiveMetric := `connection_active_count{path="",tenant_id=""} 0`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCount200Metric)
	assert.Contains(t, metricsData, expectedActiveMetric)
}

func makeRequestToServer(url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(tenantIdHeader, tenantId)
	return client.Do(req)
}

func TestConnectionMiddleware_ConnectionUpgrade_WebSocket(t *testing.T) {
	metrics.Init(nil)

	// Handler that simulates WebSocket upgrade
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the underlying ResponseWriter supports hijacking
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}

		// Set WebSocket upgrade headers
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Sec-WebSocket-Accept", "test-accept-key")
		w.WriteHeader(http.StatusSwitchingProtocols)

		// Hijack the connection
		conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Write a simple WebSocket frame (for testing purposes)
		conn.Write([]byte("WebSocket connection established"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(MetricMiddleware)
		r.Method("GET", "/ws/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	// Make WebSocket upgrade request
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL+"/ws/1", nil)
	req.Header.Set(tenantIdHeader, tenantId)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsData := string(metricsBody)

	// Check that the metrics include the WebSocket upgrade
	expectedDurationMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/ws/{id}",status_code="101",tenant_id="1"} 1`
	expectedCountMetric := `runai_connection_request_count{method="GET",path="/ws/{id}",status_code="101",tenant_id="1"} 1`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCountMetric)
}

func TestConnectionMiddleware_ConnectionUpgrade_HijackSupport(t *testing.T) {
	metrics.Init(nil)

	// Handler that checks for hijack support but doesn't attempt it
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if hijacking is supported
		_, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}

		// If we get here, hijacking is supported (shouldn't happen with httptest.Server)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hijacking supported"))
	})

	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(runaictx.ContextMiddleware)
		r.Use(MetricMiddleware)
		r.Method("GET", "/ws/{id}", handler)
	})
	resetVectors()

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := makeRequestToServer(server.URL + "/ws/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	// The ResponseWriterWrapper should support hijacking, but the underlying httptest.ResponseRecorder doesn't
	// So we expect this to succeed since our wrapper implements the Hijacker interface
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Expose the metrics endpoint
	metricsServer := httptest.NewServer(metrics.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	assert.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	metricsData := string(metricsBody)

	// Check that the metrics include the successful response
	expectedDurationMetric := `runai_connection_request_duration_seconds_count{method="GET",path="/ws/{id}",status_code="200",tenant_id="1"} 1`
	expectedCountMetric := `runai_connection_request_count{method="GET",path="/ws/{id}",status_code="200",tenant_id="1"} 1`
	assert.Contains(t, metricsData, expectedDurationMetric)
	assert.Contains(t, metricsData, expectedCountMetric)
}
