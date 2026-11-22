//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/integration-tests/prometheus-with-agent/client"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	vllmWorkload1        = "vllm-token-usage-w1"
	vllmWorkload2        = "vllm-token-usage-w2"
	vllmIdleWorkload     = "vllm-token-usage-idle"
	vllmMixedWorkload    = "vllm-token-usage-mixed"
	vllmPartialWorkload  = "vllm-token-usage-partial"
	vllmUpgradedWorkload = "vllm-token-usage-upgraded"
	// vllmUpgradedCluster upgrades from 3.0 to 3.1 one hour before metrics-service starts
	// (see the clusters client mocks).
	vllmUpgradedCluster = "cluster_4"
	vllmFullWindow      = 6 * time.Minute // the samples cover minutes 0 to 6
)

var vllmTokenUsageSeries = []string{
	"runai_vllm_prompt_tokens_total_per_pod",
	"runai_vllm_generation_tokens_total_per_pod",
	"runai_vllm_prompt_tokens_cached_total_per_pod",
	"runai_vllm_request_prompt_tokens_count_per_pod",
	"runai_vllm_request_prompt_tokens_sum_per_pod",
	"runai_vllm_request_prompt_tokens_bucket_per_pod",
	"runai_vllm_request_generation_tokens_count_per_pod",
	"runai_vllm_request_generation_tokens_sum_per_pod",
	"runai_vllm_request_generation_tokens_bucket_per_pod",
}

// VLLMTokenUsageTestSuite runs the real vLLM token usage query keys against seeded per-pod
// recording-rule series. Samples are one minute apart and the queries end exactly on the last
// sample, so increase() needs no extrapolation and the expected values are exact.
type VLLMTokenUsageTestSuite struct {
	suite.Suite
	end time.Time
}

func TestVLLMTokenUsageTestSuite(t *testing.T) {
	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "redis"))
	t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "localhost:6379"))
	t.Setenv("SERVICE_NAME", getEnvOrDefault("SERVICE_NAME", "metrics-service"))
	t.Setenv("DB_HOST", getEnvOrDefault("DB_HOST", "localhost"))
	t.Setenv("DB_PORT", getEnvOrDefault("DB_PORT", "5432"))
	t.Setenv("DB_NAME", getEnvOrDefault("DB_NAME", "runai"))
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../secrets/"))
	t.Setenv("METRICS_SERVER_HOST", getEnvOrDefault("METRICS_SERVER_HOST", "localhost:8080"))

	suite.Run(t, new(VLLMTokenUsageTestSuite))
}

func (s *VLLMTokenUsageTestSuite) SetupSuite() {
	s.end = time.Now().Truncate(time.Minute).Add(-time.Minute)
	metrics := client.NewMetricCollection()
	addSeries := func(name string, labels [][2]string, values ...float64) {
		for i, value := range values {
			metric := client.NewGaugeMetric(name, value).
				WithTime(s.end.Add(time.Duration(i-len(values)+1) * time.Minute))
			for _, label := range labels {
				metric.AddLabel(label[0], label[1])
			}
			metrics.AddMetric(metric)
		}
	}
	pod := func(workloadId, podName string, extra ...[2]string) [][2]string {
		return append([][2]string{{"workload_id", workloadId}, {"pod", podName}, {"engine", "0"}}, extra...)
	}

	// pod-1 restarts after 2 minutes; pod-2 keeps counting.
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 100, 200, 0, 100, 200)
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmWorkload1, "w1-pod-2"), 0, 50, 100, 150, 200, 250)
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmWorkload2, "w2-pod-1"), 0, 1000, 2000, 3000, 4000, 5000)
	addSeries("runai_vllm_generation_tokens_total_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 10, 20, 30, 40, 50)
	// Only w1 runs a vLLM version with the cached counter.
	addSeries("runai_vllm_prompt_tokens_cached_total_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 30, 60, 0, 30, 60)

	addSeries("runai_vllm_request_prompt_tokens_count_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 2, 4, 6, 8, 10)
	addSeries("runai_vllm_request_prompt_tokens_sum_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 200, 400, 600, 800, 1000)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "100"}), 0, 1, 2, 3, 4, 5)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "500"}), 0, 2, 4, 6, 8, 10)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "+Inf"}), 0, 2, 4, 6, 8, 10)
	addSeries("runai_vllm_request_generation_tokens_count_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 2, 4, 6, 8, 10)
	addSeries("runai_vllm_request_generation_tokens_sum_per_pod", pod(vllmWorkload1, "w1-pod-1"), 0, 100, 200, 300, 400, 500)
	addSeries("runai_vllm_request_generation_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "10"}), 0, 1, 2, 3, 4, 5)
	addSeries("runai_vllm_request_generation_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "100"}), 0, 2, 4, 6, 8, 10)
	addSeries("runai_vllm_request_generation_tokens_bucket_per_pod", pod(vllmWorkload1, "w1-pod-1", [2]string{"le", "+Inf"}), 0, 2, 4, 6, 8, 10)

	// The idle workload is serving but finishes no requests.
	for _, histogram := range []string{"prompt", "generation"} {
		prefix := "runai_vllm_request_" + histogram + "_tokens_"
		addSeries(prefix+"count_per_pod", pod(vllmIdleWorkload, "idle-pod-1"), 0, 0, 0, 0, 0, 0)
		addSeries(prefix+"sum_per_pod", pod(vllmIdleWorkload, "idle-pod-1"), 0, 0, 0, 0, 0, 0)
		addSeries(prefix+"bucket_per_pod", pod(vllmIdleWorkload, "idle-pod-1", [2]string{"le", "100"}), 0, 0, 0, 0, 0, 0)
		addSeries(prefix+"bucket_per_pod", pod(vllmIdleWorkload, "idle-pod-1", [2]string{"le", "+Inf"}), 0, 0, 0, 0, 0, 0)
	}

	// One pod finishes a single 100-token request and the other nine 10-token requests.
	addSeries("runai_vllm_request_prompt_tokens_count_per_pod", pod(vllmMixedWorkload, "mixed-pod-1"), 0, 1, 1, 1, 1, 1)
	addSeries("runai_vllm_request_prompt_tokens_sum_per_pod", pod(vllmMixedWorkload, "mixed-pod-1"), 0, 100, 100, 100, 100, 100)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-1", [2]string{"le", "10"}), 0, 0, 0, 0, 0, 0)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-1", [2]string{"le", "100"}), 0, 1, 1, 1, 1, 1)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-1", [2]string{"le", "+Inf"}), 0, 1, 1, 1, 1, 1)
	addSeries("runai_vllm_request_prompt_tokens_count_per_pod", pod(vllmMixedWorkload, "mixed-pod-2"), 0, 3, 6, 9, 9, 9)
	addSeries("runai_vllm_request_prompt_tokens_sum_per_pod", pod(vllmMixedWorkload, "mixed-pod-2"), 0, 30, 60, 90, 90, 90)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-2", [2]string{"le", "10"}), 0, 3, 6, 9, 9, 9)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-2", [2]string{"le", "100"}), 0, 3, 6, 9, 9, 9)
	addSeries("runai_vllm_request_prompt_tokens_bucket_per_pod", pod(vllmMixedWorkload, "mixed-pod-2", [2]string{"le", "+Inf"}), 0, 3, 6, 9, 9, 9)

	// pod-2 has only one sample in the period, so increase() has nothing to compute for it.
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmPartialWorkload, "partial-pod-1"), 0, 100, 200, 300, 400, 500)
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmPartialWorkload, "partial-pod-2"), 40)

	// The per-pod rules exist only from 3.1, so the upgraded cluster has samples only after its upgrade.
	addSeries("runai_vllm_prompt_tokens_total_per_pod", pod(vllmUpgradedWorkload, "upgraded-pod-1", [2]string{"clusterId", vllmUpgradedCluster}),
		0, 50, 100, 150, 200, 250)

	require.NoError(s.T(), client.PostMetrics(metrics.ToOpenMetrics()))
	require.NoError(s.T(), client.WaitForMetrics(vllmTokenUsageSeries, 5, 120))
}

func (s *VLLMTokenUsageTestSuite) TearDownSuite() {
	client.ClearMetrics(vllmTokenUsageSeries)
}

func (s *VLLMTokenUsageTestSuite) query(queryKey string, window time.Duration, workloadIds ...string) *openapi.GetMetrics200Response {
	return s.queryFiltered(queryKey, window, VLLMTokenUsageFilters{WorkloadId: workloadIds})
}

func (s *VLLMTokenUsageTestSuite) queryFiltered(queryKey string, window time.Duration, filters VLLMTokenUsageFilters) *openapi.GetMetrics200Response {
	metric := &VLLMTokenUsageMetric{
		Name:           queryKey,
		Filters:        filters,
		ResponseLabels: VLLMTokenUsageResponseLabels{WorkloadId: true, QuantileLabel: true},
	}
	metricQuery, err := metrics_client.NewMetricQuery(metric)
	require.NoError(s.T(), err)
	metricQuery.SetRange(s.end.Add(-window), s.end)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	result, err := metricQuery.Query(ctx)
	require.NoError(s.T(), err)
	return result
}

func (s *VLLMTokenUsageTestSuite) assertSingleValue(result *openapi.GetMetrics200Response, workloadId, expected string) {
	require.Len(s.T(), result.Data, 1)
	assert.Equal(s.T(), workloadId, result.Data[0].MetricLabels["workload_id"])
	require.Len(s.T(), result.Data[0].Measurements, 1)
	assert.Equal(s.T(), expected, result.Data[0].Measurements[0].Value)
}

func (s *VLLMTokenUsageTestSuite) TestTotals_CountAcrossRestartsAndPods() {
	result := s.query("runai_vllm_prompt_tokens_total_per_workload", vllmFullWindow, vllmWorkload1)
	assert.Contains(s.T(), result.Query, fmt.Sprintf("[%ds]", int(vllmFullWindow.Seconds())))
	// pod-1: 200 + 200 across its restart, pod-2: 250. w2 is not included.
	s.assertSingleValue(result, vllmWorkload1, "650")

	s.assertSingleValue(s.query("runai_vllm_generation_tokens_total_per_workload", vllmFullWindow, vllmWorkload1), vllmWorkload1, "50")
	s.assertSingleValue(s.query("runai_vllm_prompt_tokens_cached_total_per_workload", vllmFullWindow, vllmWorkload1), vllmWorkload1, "120")
}

func (s *VLLMTokenUsageTestSuite) TestTotals_OneValuePerWorkload() {
	result := s.query("runai_vllm_prompt_tokens_total_per_workload", vllmFullWindow, vllmWorkload1, vllmWorkload2)
	values := map[string]string{}
	for _, data := range result.Data {
		require.Len(s.T(), data.Measurements, 1)
		values[data.MetricLabels["workload_id"]] = data.Measurements[0].Value
	}
	assert.Equal(s.T(), map[string]string{vllmWorkload1: "650", vllmWorkload2: "5000"}, values)
}

func (s *VLLMTokenUsageTestSuite) TestCachedTokens_NoValueWithoutTheCounter() {
	result := s.query("runai_vllm_prompt_tokens_cached_total_per_workload", vllmFullWindow, vllmWorkload2)
	assert.Empty(s.T(), result.Data)
}

func (s *VLLMTokenUsageTestSuite) TestRequestStatistics() {
	s.assertSingleValue(s.query("runai_vllm_finished_engine_request_count_per_workload", vllmFullWindow, vllmWorkload1), vllmWorkload1, "10")
	s.assertSingleValue(s.query("runai_vllm_prompt_tokens_per_request_avg_per_workload", vllmFullWindow, vllmWorkload1), vllmWorkload1, "100")

	// 5 of 10 requests are at most 100 tokens and all 10 are at most 500.
	s.assertPercentiles(s.query("runai_vllm_prompt_tokens_per_request_percentile_per_workload", vllmFullWindow, vllmWorkload1),
		map[string]float64{"50": 100, "95": 460, "99": 492})
}

func (s *VLLMTokenUsageTestSuite) TestGenerationRequestStatistics() {
	s.assertSingleValue(s.query("runai_vllm_generation_tokens_per_request_avg_per_workload", vllmFullWindow, vllmWorkload1), vllmWorkload1, "50")
	// 5 of 10 requests are at most 10 tokens and all 10 are at most 100.
	s.assertPercentiles(s.query("runai_vllm_generation_tokens_per_request_percentile_per_workload", vllmFullWindow, vllmWorkload1),
		map[string]float64{"50": 10, "95": 91, "99": 98.2})
}

func (s *VLLMTokenUsageTestSuite) TestRequestStatistics_NoFinishedRequests() {
	s.assertSingleValue(s.query("runai_vllm_finished_engine_request_count_per_workload", vllmFullWindow, vllmIdleWorkload), vllmIdleWorkload, "0")
	// No value instead of NaN, which the workloads service would report as "0".
	for _, queryKey := range []string{
		"runai_vllm_prompt_tokens_per_request_avg_per_workload",
		"runai_vllm_prompt_tokens_per_request_percentile_per_workload",
		"runai_vllm_generation_tokens_per_request_avg_per_workload",
		"runai_vllm_generation_tokens_per_request_percentile_per_workload",
	} {
		assert.Empty(s.T(), s.query(queryKey, vllmFullWindow, vllmIdleWorkload).Data, queryKey)
	}
}

// assertPercentiles compares the p50, p95 and p99 values by quantile label. The values are
// interpolated inside histogram buckets, so they are compared with a small tolerance.
func (s *VLLMTokenUsageTestSuite) assertPercentiles(result *openapi.GetMetrics200Response, expected map[string]float64) {
	require.Len(s.T(), result.Data, len(expected))
	for _, data := range result.Data {
		require.Len(s.T(), data.Measurements, 1)
		quantile := data.MetricLabels["quantile_label"]
		require.Contains(s.T(), expected, quantile)
		value, err := strconv.ParseFloat(data.Measurements[0].Value, 64)
		require.NoError(s.T(), err)
		assert.InDelta(s.T(), expected[quantile], value, 1e-9, "p%s", quantile)
	}
}

func (s *VLLMTokenUsageTestSuite) TestWindowShorterThanTwoSamples_NoValue() {
	assert.Empty(s.T(), s.query("runai_vllm_prompt_tokens_total_per_workload", 30*time.Second, vllmWorkload1).Data)
}

func (s *VLLMTokenUsageTestSuite) TestRequestStatistics_CombineRequestsFromAllPods() {
	s.assertSingleValue(s.query("runai_vllm_finished_engine_request_count_per_workload", vllmFullWindow, vllmMixedWorkload), vllmMixedWorkload, "10")
	// (100 + 9 * 10) / 10, not the average of the two pod averages (55).
	s.assertSingleValue(s.query("runai_vllm_prompt_tokens_per_request_avg_per_workload", vllmFullWindow, vllmMixedWorkload), vllmMixedWorkload, "19")
	// 9 of 10 requests are at most 10 tokens and all 10 are at most 100.
	s.assertPercentiles(s.query("runai_vllm_prompt_tokens_per_request_percentile_per_workload", vllmFullWindow, vllmMixedWorkload),
		map[string]float64{"50": 50.0 / 9, "95": 55, "99": 91})
}

func (s *VLLMTokenUsageTestSuite) TestTotals_PodWithOneSampleIsNotCounted() {
	// Only pod-1 contributes. The result is still a successful numeric value, so a total
	// can leave out the usage of pods that have fewer than two samples in the period.
	s.assertSingleValue(s.query("runai_vllm_prompt_tokens_total_per_workload", vllmFullWindow, vllmPartialWorkload), vllmPartialWorkload, "500")
}

func (s *VLLMTokenUsageTestSuite) TestTotals_PeriodStartsBeforeTheUpgradeTo31() {
	// The two-hour period starts on 3.0 and ends on 3.1, so metrics-service uses the cross-version
	// query. Only the usage recorded after the upgrade is returned.
	result := s.queryFiltered("runai_vllm_prompt_tokens_total_per_workload", 2*time.Hour,
		VLLMTokenUsageFilters{ClusterId: []string{vllmUpgradedCluster}, WorkloadId: []string{vllmUpgradedWorkload}})
	assert.Contains(s.T(), result.Query, fmt.Sprintf("clusterId=~'%s'", vllmUpgradedCluster))
	s.assertSingleValue(result, vllmUpgradedWorkload, "250")
}
