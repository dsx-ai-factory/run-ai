/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const clusterUUID = "d1405b3f-761a-4d4b-acb6-a00327ca853c"

func clMeas(metricType string, vals ...[2]string) map[string]any {
	values := make([]map[string]any, 0, len(vals))
	for _, v := range vals {
		values = append(values, map[string]any{"value": v[0], "timestamp": v[1]})
	}
	return map[string]any{"type": metricType, "values": values}
}

func clusterMetricsPath() string { return "/api/v1/clusters/" + clusterUUID + "/metrics" }

func TestGetClusterMetricsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getClusterMetricsInput
		wantErr string
	}{
		{name: "happy", in: getClusterMetricsInput{ClusterID: clusterUUID}},
		{name: "happy_nodepool_scope", in: getClusterMetricsInput{ClusterID: clusterUUID, NodePoolName: "gpu-pool"}},
		{name: "missing_cluster", in: getClusterMetricsInput{}, wantErr: "clusterId"},
		{name: "bad_cluster_uuid", in: getClusterMetricsInput{ClusterID: "not-a-uuid"}, wantErr: "UUID"},
		{name: "unknown_metric", in: getClusterMetricsInput{ClusterID: clusterUUID, MetricTypes: []string{"diskIO"}}, wantErr: "unknown metric"},
		{name: "bad_window", in: getClusterMetricsInput{ClusterID: clusterUUID, Window: "last_year"}, wantErr: "window must be"},
		{name: "start_without_end", in: getClusterMetricsInput{ClusterID: clusterUUID, Start: "2026-06-01T00:00:00Z"}, wantErr: "together"},
		{name: "samples_over_max", in: getClusterMetricsInput{ClusterID: clusterUUID, NumberOfSamples: 1001}, wantErr: "numberOfSamples"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestGetClusterMetricsRun_FlatHappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotSamples string
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, r *http.Request) {
		gotSamples = r.URL.Query().Get("numberOfSamples")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				clMeas("ALLOCATED_GPU", [2]string{"3", "2026-06-01T00:00:00Z"}),
				clMeas("UNALLOCATED_GPU", [2]string{"13", "2026-06-01T00:00:00Z"}),
				clMeas("GPU_UTILIZATION", [2]string{"85.333", "2026-06-01T00:00:00Z"}),
			},
		})
	})

	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID:       clusterUUID,
		MetricTypes:     []string{"allocatedGpu", "unallocatedGpu", "gpuUtilization"},
		NumberOfSamples: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, "5", gotSamples)
	assert.Nil(t, out.NodePool)
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.AllocatedGpu, 1)
	assert.Equal(t, float64(3), out.Series.AllocatedGpu[0].V)
	require.Len(t, out.Series.UnallocatedGpu, 1)
	assert.Equal(t, float64(13), out.Series.UnallocatedGpu[0].V)
	require.Len(t, out.Series.GPUUtilization, 1)
	assert.Equal(t, 85.333, out.Series.GPUUtilization[0].V)
}

func TestGetClusterMetricsRun_TimeRange(t *testing.T) {
	cases := []struct {
		name      string
		in        getClusterMetricsInput
		wantStart string // exact when explicit; "" means "defaulted from window, just assert non-empty"
		wantEnd   string
	}{
		{
			name:      "explicit_range_overrides_window",
			in:        getClusterMetricsInput{ClusterID: clusterUUID, Window: "last_24h", Start: "2026-06-01T00:00:00Z", End: "2026-06-02T00:00:00Z"},
			wantStart: "2026-06-01T00:00:00Z",
			wantEnd:   "2026-06-02T00:00:00Z",
		},
		{
			name: "window_defaulted",
			in:   getClusterMetricsInput{ClusterID: clusterUUID, Window: "last_24h"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			var gotStart, gotEnd string
			mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, r *http.Request) {
				gotStart = r.URL.Query().Get("start")
				gotEnd = r.URL.Query().Get("end")
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"measurements": []map[string]any{clMeas("ALLOCATED_GPU", [2]string{"1", "2026-06-01T00:00:00Z"})},
				})
			})

			_, err := getClusterMetricsRun(context.Background(), client, tc.in)
			require.NoError(t, err)
			if tc.wantStart != "" {
				assert.Equal(t, tc.wantStart, gotStart, "explicit start forwarded verbatim")
				assert.Equal(t, tc.wantEnd, gotEnd, "explicit end forwarded verbatim")
				return
			}
			assert.NotEmpty(t, gotStart, "window resolves to a concrete start")
			assert.NotEmpty(t, gotEnd, "window resolves to a concrete end")
			assert.NotEqual(t, gotStart, gotEnd, "window spans a non-empty range")
		})
	}
}

// The happy-path test exercises the GPU-capacity trio; this one covers totalGpu,
// the utilization family, and the scheduling metrics so every series key and its
// assignClusterSeries branch is wired through.
func TestGetClusterMetricsRun_AllMetricFamilies(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				clMeas("TOTAL_GPU", [2]string{"16", "2026-06-01T00:00:00Z"}),
				clMeas("GPU_UTILIZATION", [2]string{"42.5", "2026-06-01T00:00:00Z"}),
				clMeas("GPU_MEMORY_UTILIZATION", [2]string{"55.25", "2026-06-01T00:00:00Z"}),
				clMeas("CPU_UTILIZATION", [2]string{"12.1", "2026-06-01T00:00:00Z"}),
				clMeas("CPU_MEMORY_UTILIZATION", [2]string{"33.3", "2026-06-01T00:00:00Z"}),
				clMeas("AVG_WORKLOAD_WAIT_TIME", [2]string{"7.5", "2026-06-01T00:00:00Z"}),
				clMeas("WORKLOADS_COUNT", [2]string{"9", "2026-06-01T00:00:00Z"}),
			},
		})
	})

	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID: clusterUUID,
		MetricTypes: []string{
			"totalGpu", "gpuUtilization", "gpuMemoryUtilization",
			"cpuUtilization", "cpuMemoryUtilization", "avgWorkloadWaitTime", "workloadsCount",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.TotalGpu, 1)
	assert.Equal(t, float64(16), out.Series.TotalGpu[0].V)
	require.Len(t, out.Series.GPUUtilization, 1)
	assert.Equal(t, 42.5, out.Series.GPUUtilization[0].V)
	require.Len(t, out.Series.GPUMemoryUtilization, 1)
	assert.Equal(t, 55.25, out.Series.GPUMemoryUtilization[0].V)
	require.Len(t, out.Series.CPUUtilization, 1)
	assert.Equal(t, 12.1, out.Series.CPUUtilization[0].V)
	require.Len(t, out.Series.CPUMemoryUtilization, 1)
	assert.Equal(t, 33.3, out.Series.CPUMemoryUtilization[0].V)
	require.Len(t, out.Series.AvgWorkloadWaitTime, 1)
	assert.Equal(t, 7.5, out.Series.AvgWorkloadWaitTime[0].V)
	require.Len(t, out.Series.WorkloadsCount, 1)
	assert.Equal(t, float64(9), out.Series.WorkloadsCount[0].V)
}

// NaN/Inf would break JSON encoding, and a duplicated metricType must not double
// the upstream request or the truncation estimate.
func TestGetClusterMetricsRun_DropsNonFiniteAndDedupes(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotMetricTypes []string
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, r *http.Request) {
		gotMetricTypes = r.URL.Query()["metricType"]
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				clMeas("ALLOCATED_GPU",
					[2]string{"NaN", "2026-06-01T00:00:00Z"},
					[2]string{"+Inf", "2026-06-01T00:01:00Z"},
					[2]string{"4", "2026-06-01T00:02:00Z"},
				),
			},
		})
	})

	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID:   clusterUUID,
		MetricTypes: []string{"allocatedGpu", "allocatedGpu"},
	})
	require.NoError(t, err)
	assert.Len(t, gotMetricTypes, 1, "duplicate metricType collapsed to one upstream value")
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.AllocatedGpu, 1, "only the finite point survives")
	assert.Equal(t, float64(4), out.Series.AllocatedGpu[0].V)
}

func TestGetClusterMetricsRun_NodePoolScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotNodePool string
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, r *http.Request) {
		gotNodePool = r.URL.Query().Get("nodepoolName")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{clMeas("ALLOCATED_GPU", [2]string{"1", "2026-06-01T00:00:00Z"})},
		})
	})

	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID: clusterUUID, NodePoolName: "gpu-pool", MetricTypes: []string{"allocatedGpu"},
	})
	require.NoError(t, err)
	assert.Equal(t, "gpu-pool", gotNodePool, "nodePoolName forwarded")
	require.NotNil(t, out.NodePool)
	assert.Equal(t, "gpu-pool", *out.NodePool)
	require.NotNil(t, out.Series)
	require.Len(t, out.Series.AllocatedGpu, 1)
}

func TestGetClusterMetricsRun_TruncationGuard(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t) // no handler: must not be called
	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID:       clusterUUID,
		NumberOfSamples: 1000,
		MetricTypes:     []string{"allocatedGpu", "unallocatedGpu", "totalGpu", "gpuUtilization", "gpuMemoryUtilization", "cpuUtilization"}, // 6 × 1000 > 5000
	})
	require.NoError(t, err)
	assert.True(t, out.SeriesTruncated)
	assert.Nil(t, out.Series)
}

func TestGetClusterMetricsRun_PartialData207(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{clMeas("ALLOCATED_GPU", [2]string{"2", "2026-06-01T00:00:00Z"})},
		})
	})

	out, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{
		ClusterID: clusterUUID, MetricTypes: []string{"allocatedGpu", "cpuUtilization"},
	})
	require.NoError(t, err)
	assert.True(t, out.PartialData)
	require.NotNil(t, out.Series)
	assert.Len(t, out.Series.AllocatedGpu, 1)
	assert.Empty(t, out.Series.CPUUtilization, "failed metric absent")
}

func TestGetClusterMetricsRun_ForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(clusterMetricsPath(), func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getClusterMetricsRun(context.Background(), client, getClusterMetricsInput{ClusterID: clusterUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}
