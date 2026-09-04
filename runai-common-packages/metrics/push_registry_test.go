/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

func TestPush_Gauge(t *testing.T) {
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service1", gateway.URL)

	gauge := metricsPush.GetRegistry().Gauge("test_gauge").Register("A custom gauge with labels", map[string]string{"label1": "value1", "label2": "value2"})
	gauge.Set(54)

	err := metricsPush.Push()
	assert.NoError(t, err)

	gaugeMetric, ok := metricsMap["runai_test_gauge"]
	assert.True(t, ok)
	assert.Equal(t, float64(54), *gaugeMetric.GetMetric()[0].Gauge.Value)
	assert.Equal(t, 2, len(gaugeMetric.GetMetric()[0].GetLabel()))
}

func TestPush_Counter(t *testing.T) {
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service1", gateway.URL)

	counter := metricsPush.GetRegistry().Counter("test_counter").Register("description", nil)
	counter.Add(5)

	err := metricsPush.Push()
	assert.NoError(t, err)

	counterMetric, ok := metricsMap["runai_test_counter"]
	assert.True(t, ok)
	assert.Equal(t, float64(5), *counterMetric.GetMetric()[0].Counter.Value)
	assert.Equal(t, 0, len(counterMetric.GetMetric()[0].GetLabel()))
}

func TestPush_Histogram(t *testing.T) {
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service1", gateway.URL)

	histogramBuckets := []float64{0.1, 0.2, 0.5, 1, 2, 5, 10}
	histogram := metricsPush.GetRegistry().Histogram("test_histogram").Register("A custom histogram", map[string]string{"label1": "value1"}, histogramBuckets)
	histogram.Observe(0.3)

	err := metricsPush.Push()
	assert.NoError(t, err)

	histogramMetric, ok := metricsMap["runai_test_histogram"]
	assert.True(t, ok)
	assert.Equal(t, float64(0.3), *histogramMetric.GetMetric()[0].Histogram.SampleSum)
	assert.Equal(t, 1, len(histogramMetric.GetMetric()[0].GetLabel()))

	histogramMetricBuckets := histogramMetric.GetMetric()[0].Histogram.GetBucket()
	assert.Equal(t, len(histogramBuckets), len(histogramMetricBuckets))
	for i := range histogramBuckets {
		assert.Equal(t, histogramBuckets[i], *histogramMetricBuckets[i].UpperBound)
	}
}

func TestPush_Summary(t *testing.T) {
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service1", gateway.URL)

	summaryQuantiles := map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	summary := metricsPush.GetRegistry().Summary("test_summary").Register("A custom summary with labels", map[string]string{"label1": "value1", "label2": "value2"}, summaryQuantiles)
	summary.Observe(0.7)

	err := metricsPush.Push()
	assert.NoError(t, err)

	summaryMetric, ok := metricsMap["runai_test_summary"]
	assert.True(t, ok)
	assert.Equal(t, float64(0.7), *summaryMetric.GetMetric()[0].Summary.SampleSum)
	assert.Equal(t, 2, len(summaryMetric.GetMetric()[0].GetLabel()))

	summaryMetricQuantiles := summaryMetric.GetMetric()[0].Summary.GetQuantile()
	assert.Equal(t, len(summaryQuantiles), len(summaryMetricQuantiles))
}

func TestPush_HappyFlow(t *testing.T) {
	var counter int
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service3", gateway.URL)

	gauge := metricsPush.GetRegistry().Gauge("my_custom_gauge_with_labels").Register("A custom gauge with labels", map[string]string{"label1": "value1", "label2": "value2"})
	gauge.Set(42)

	err := metricsPush.Push()
	assert.NoError(t, err)
	assert.Equal(t, counter, 1)

	gauge1Metric, ok := metricsMap["runai_my_custom_gauge_with_labels"]
	assert.True(t, ok)
	assert.Equal(t, float64(42), *gauge1Metric.GetMetric()[0].Gauge.Value)
	assert.Equal(t, 2, len(gauge1Metric.GetMetric()[0].GetLabel()))

	gauge2 := metricsPush.GetRegistry().Gauge("my_second_custom_gauge_with_labels").Register("A second custom gauge with labels", map[string]string{"label3": "value3", "label4": "value4"})
	gauge2.Set(52)

	err = metricsPush.Push()
	assert.NoError(t, err)
	assert.Equal(t, counter, 2)

	gauge2Metric, ok := metricsMap["runai_my_second_custom_gauge_with_labels"]
	assert.True(t, ok)
	assert.Equal(t, float64(52), *gauge2Metric.GetMetric()[0].Gauge.Value)
	assert.Equal(t, 2, len(gauge2Metric.GetMetric()[0].GetLabel()))
}

func TestPush_RePushOfTheSameMetricUpdatesValues(t *testing.T) {
	var metricsMap map[string]*io_prometheus_client.MetricFamily

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMap = decodeMetricsResponse(t, r)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service-concurrent", gateway.URL)
	gauge := metricsPush.GetRegistry().Gauge("some_gauge").Register("Some gauge", nil)

	for i := range 3 {
		gauge.Set(float64(i))
		err := metricsPush.Push()
		assert.NoError(t, err)

		metric, ok := metricsMap["runai_some_gauge"]
		assert.True(t, ok)
		assert.Equal(t, float64(i), *metric.GetMetric()[0].Gauge.Value)
	}
}

func TestPush_MetricsDontGetToGateway_WhenPushIsNotCalled(t *testing.T) {
	var counter int
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(http.StatusAccepted)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service2", gateway.URL)
	metricsPush.GetRegistry().Gauge("my_custom_gauge_with_labels").Register("A custom gauge with labels", nil)
	assert.Equal(t, counter, 0)
}

func TestPush_FailedToPush(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer gateway.Close()

	metricsPush := newMetricsPusher(prometheus.NewRegistry(), "test-service-error", gateway.URL)

	gauge := metricsPush.GetRegistry().Gauge("error_gauge").Register("Test error handling", nil)
	gauge.Set(99)

	err := metricsPush.Push()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("failed to push metrics, err: unexpected status code 500 while pushing to %s/metrics/job/test-service-error: ", gateway.URL))
}

func decodeMetricsResponse(t *testing.T, r *http.Request) map[string]*io_prometheus_client.MetricFamily {
	metricsMap := make(map[string]*io_prometheus_client.MetricFamily)

	dec := expfmt.NewDecoder(r.Body, expfmt.ResponseFormat(r.Header))
	for {
		metricFamily := &io_prometheus_client.MetricFamily{}
		err := dec.Decode(metricFamily)
		if err != nil && err.Error() == "EOF" {
			break
		}
		assert.NoError(t, err)
		metricsMap[metricFamily.GetName()] = metricFamily
	}

	return metricsMap
}
