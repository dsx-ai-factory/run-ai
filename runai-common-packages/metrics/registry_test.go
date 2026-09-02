/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/runaimetrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var globalMetrics, _ = Init(nil)

func TestHistogram_GetFromRegistryReturnSame(t *testing.T) {
	globalMetrics.Histogram("test_histogram").Register("description", make(map[string]string), []float64{})
	metric := globalMetrics.Histogram("test_histogram").Get()
	promMetric := metric.(*runaimetrics.Histogram)
	assert.NotNil(t, metric)
	desc := promMetric.PromHistogram.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_histogram"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestSummary_GetFromRegistryReturnSame(t *testing.T) {
	globalMetrics.Summary("test_summary").Register("description", make(map[string]string), make(map[float64]float64))
	metric := globalMetrics.Summary("test_summary").Get()
	promMetric := metric.(*runaimetrics.Summary)
	assert.NotNil(t, metric)
	desc := promMetric.PromSummary.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_summary"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestGauge_GetFromRegistryReturnSame(t *testing.T) {
	globalMetrics.Gauge("test_gauge").Register("description", make(map[string]string))
	metric := globalMetrics.Gauge("test_gauge").Get()
	promMetric := metric.(*runaimetrics.Gauge)
	assert.NotNil(t, metric)
	desc := promMetric.PromGauge.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_gauge"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestCounter_GetFromRegistryReturnSame(t *testing.T) {
	globalMetrics.Counter("test_counter").Register("description", make(map[string]string))
	metric := globalMetrics.Counter("test_counter").Get()
	promMetric := metric.(*runaimetrics.Counter)
	assert.NotNil(t, metric)
	desc := promMetric.PromCounter.Desc().String()
	assert.Contains(t, desc, `fqName: "runai_test_counter"`)
	assert.Contains(t, desc, `help: "description"`)
}

func TestNewMetrics_DefaultCollectorsRegistered(t *testing.T) {
	metricsHandler := globalMetrics.Handler()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsHandler.ServeHTTP(resp, req)

	body := resp.Body.String()
	assert.Contains(t, body, "go_gc_duration_seconds", "Go collector should be registered")
	if runtime.GOOS == "linux" {
		assert.Contains(t, body, "process_start_time_seconds", "Process collector should be registered") // available only on Linux
	}
}

func TestRegisterCounter(t *testing.T) {
	counter := globalMetrics.Counter("my_counter").Register("A custom counter", make(map[string]string))
	counter.Inc()
	promCounter := counter.(*runaimetrics.Counter).PromCounter
	assert.Equal(t, 1.0, testutil.ToFloat64(promCounter), "Custom counter should be 1")
}

func TestRegisterGauge(t *testing.T) {
	gauge := globalMetrics.Gauge("my_custom_gauge").Register("A custom gauge", nil)
	gauge.Set(42)
	promGauge := gauge.(*runaimetrics.Gauge).PromGauge
	assert.Equal(t, 42.0, testutil.ToFloat64(promGauge), "Custom gauge should be 42")
}

func TestRegisterHistogram(t *testing.T) {
	histogram := globalMetrics.Histogram("my_custom_histogram").Register("A custom histogram", map[string]string{}, []float64{0.1, 0.2, 0.5, 1, 2, 5, 10})
	histogram.Observe(0.3)
	expected := `
	# HELP runai_my_custom_histogram A custom histogram
	# TYPE runai_my_custom_histogram histogram
	runai_my_custom_histogram_bucket{le="0.1"} 0
	runai_my_custom_histogram_bucket{le="0.2"} 0
	runai_my_custom_histogram_bucket{le="0.5"} 1
	runai_my_custom_histogram_bucket{le="1"} 1
	runai_my_custom_histogram_bucket{le="2"} 1
	runai_my_custom_histogram_bucket{le="5"} 1
	runai_my_custom_histogram_bucket{le="10"} 1
	runai_my_custom_histogram_bucket{le="+Inf"} 1
	runai_my_custom_histogram_sum 0.3
	runai_my_custom_histogram_count 1
	`
	promHistogram := histogram.(*runaimetrics.Histogram).PromHistogram
	err := testutil.CollectAndCompare(promHistogram, strings.NewReader(expected), "runai_my_custom_histogram")
	assert.NoError(t, err, "Custom histogram should observe 0.3")
}

func TestRegisterSummary(t *testing.T) {
	summary := globalMetrics.Summary("my_custom_summary").Register("A custom summary", nil, map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001})
	summary.Observe(0.7)
	expected := `
    # HELP runai_my_custom_summary A custom summary
    # TYPE runai_my_custom_summary summary
	runai_my_custom_summary{quantile="0.5"} 0.7
	runai_my_custom_summary{quantile="0.9"} 0.7
	runai_my_custom_summary{quantile="0.99"} 0.7
    runai_my_custom_summary_sum 0.7
    runai_my_custom_summary_count 1
    `
	promSummary := summary.(*runaimetrics.Summary).PromSummary
	err := testutil.CollectAndCompare(promSummary, strings.NewReader(expected), "runai_my_custom_summary")
	assert.NoError(t, err, "Custom summary should observe 0.7")
}

func TestMetricsHandler_ExposesCustomCounter(t *testing.T) {
	counter := globalMetrics.Counter("my_custom_counter").Register("A custom counter", nil)
	counter.Inc()

	metricsHandler := globalMetrics.Handler()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsHandler.ServeHTTP(resp, req)

	body := resp.Body.String()
	assert.Contains(t, body, "my_custom_counter", "Metrics endpoint should include the custom counter")
}

func TestRegisterCounterWithLabels(t *testing.T) {
	counter := globalMetrics.Counter("my_custom_counter_with_labels").Register("A custom counter with labels", map[string]string{"label1": "value1", "label2": "value2"})
	counter.Inc()
	promCounter := counter.(*runaimetrics.Counter).PromCounter
	assert.Equal(t, 1.0, testutil.ToFloat64(promCounter), "Custom counter with labels should be 1")
}

func TestRegisterGaugeWithLabels(t *testing.T) {
	gauge := globalMetrics.Gauge("my_custom_gauge_with_labels").Register("A custom gauge with labels", map[string]string{"label1": "value1", "label2": "value2"})
	gauge.Set(42)

	promGauge := gauge.(*runaimetrics.Gauge).PromGauge
	assert.Equal(t, 42.0, testutil.ToFloat64(promGauge), "Custom gauge with labels should be 42")
}

func TestRegisterHistogramWithLabels(t *testing.T) {
	histogram := globalMetrics.Histogram("my_custom_histogram_with_labels").Register("A custom histogram with labels", map[string]string{"label1": "value1", "label2": "value2"}, []float64{0.1, 0.2, 0.5, 1, 2, 5, 10})
	histogram.Observe(0.3)
	expected := `
    # HELP runai_my_custom_histogram_with_labels A custom histogram with labels
    # TYPE runai_my_custom_histogram_with_labels histogram
    runai_my_custom_histogram_with_labels_bucket{le="0.1",label1="value1",label2="value2"} 0
    runai_my_custom_histogram_with_labels_bucket{le="0.2",label1="value1",label2="value2"} 0
    runai_my_custom_histogram_with_labels_bucket{le="0.5",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_bucket{le="1",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_bucket{le="2",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_bucket{le="5",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_bucket{le="10",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_bucket{le="+Inf",label1="value1",label2="value2"} 1
    runai_my_custom_histogram_with_labels_sum{label1="value1",label2="value2"} 0.3
    runai_my_custom_histogram_with_labels_count{label1="value1",label2="value2"} 1
    `
	promHistogram := histogram.(*runaimetrics.Histogram).PromHistogram
	err := testutil.CollectAndCompare(promHistogram, strings.NewReader(expected), "runai_my_custom_histogram_with_labels")
	assert.NoError(t, err, "Custom histogram with labels should observe 0.3")
}

func TestRegisterSummaryWithLabels(t *testing.T) {
	summary := globalMetrics.Summary("my_custom_summary_with_labels").Register("A custom summary with labels", map[string]string{"label1": "value1", "label2": "value2"}, map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001})
	summary.Observe(0.7)
	expected := `
    # HELP runai_my_custom_summary_with_labels A custom summary with labels
    # TYPE runai_my_custom_summary_with_labels summary
    runai_my_custom_summary_with_labels{label1="value1",label2="value2",quantile="0.5"} 0.7
    runai_my_custom_summary_with_labels{label1="value1",label2="value2",quantile="0.9"} 0.7
    runai_my_custom_summary_with_labels{label1="value1",label2="value2",quantile="0.99"} 0.7
    runai_my_custom_summary_with_labels_sum{label1="value1",label2="value2"} 0.7
    runai_my_custom_summary_with_labels_count{label1="value1",label2="value2"} 1
    `
	promSummary := summary.(*runaimetrics.Summary).PromSummary
	err := testutil.CollectAndCompare(promSummary, strings.NewReader(expected), "runai_my_custom_summary_with_labels")
	assert.NoError(t, err, "Custom summary with labels should observe 0.7")
}

func TestRegisterCounterVecWithLabelsAndReset(t *testing.T) {
	counterVec := globalMetrics.CounterVec("my_custom_counter_vec_with_labels").Register("A custom counter vec with labels", []string{"label1", "label2"})
	counterVec.WithLabels(map[string]string{"label1": "value1", "label2": "value2"}).Inc()
	expected := `
    # HELP runai_my_custom_counter_vec_with_labels A custom counter vec with labels
    # TYPE runai_my_custom_counter_vec_with_labels counter
    runai_my_custom_counter_vec_with_labels{label1="value1",label2="value2"} 1
    `
	promCounterVec := counterVec.(*runaimetrics.CounterVec).PromCounterVec
	err := testutil.CollectAndCompare(promCounterVec, strings.NewReader(expected), "runai_my_custom_counter_vec_with_labels")
	assert.NoError(t, err, "Custom counter vec with labels should be 1")

	counterVec.Reset()
	expected = `
    # HELP runai_my_custom_counter_vec_with_labels A custom counter vec with labels
    # TYPE runai_my_custom_counter_vec_with_labels counter
    runai_my_custom_counter_vec_with_labels{label1="value1",label2="value2"} 0
    `
	err = testutil.CollectAndCompare(promCounterVec, strings.NewReader(expected), "runai_my_custom_counter_vec_with_labels")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runai_my_custom_counter_vec_with_labels")
}

func TestGatherOpenMetricsString(t *testing.T) {
	r := Registry("test")
	counterVec := r.CounterVec("my_custom_counter_vec_with_labels_for_gather").Register("A custom counter vec with labels for gather", []string{"label1", "label2"})
	counterVec.WithLabels(map[string]string{"label1": "value1", "label2": "value2"}).Inc()
	counterVec.WithLabels(map[string]string{"label1": "value2", "label2": "value3"}).Add(3)

	metricsServer := httptest.NewServer(r.Handler())
	defer metricsServer.Close()

	// Scrape the metrics endpoint
	metricsResp, err := http.Get(metricsServer.URL)
	require.NoError(t, err)
	defer metricsResp.Body.Close()

	// Read the metrics
	metricsBody, err := io.ReadAll(metricsResp.Body)
	require.NoError(t, err)
	metricsData := string(metricsBody)

	res, err := r.BuildMetricsString()
	assert.NoError(t, err)
	assert.Equal(t, metricsData, res)
}
