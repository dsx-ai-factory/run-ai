/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const validPodID = "0f3d6a2b-58a3-4f54-9bba-71f0c0a0a9d2"

func TestGetWorkloadPodsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadPodsInput
		wantErr bool
	}{
		{"empty", getWorkloadPodsInput{WorkloadID: ""}, true},
		{"whitespace", getWorkloadPodsInput{WorkloadID: "   "}, true},
		{"not a uuid", getWorkloadPodsInput{WorkloadID: "wl-1"}, true},
		{"valid", getWorkloadPodsInput{WorkloadID: validUUID}, false},
		{"valid max limit and offset", getWorkloadPodsInput{WorkloadID: validUUID, Limit: 500, Offset: 20}, false},
		{"valid metric limit", getWorkloadPodsInput{WorkloadID: validUUID, Limit: 20, IncludeMetrics: true}, false},
		{"valid advanced metrics", getWorkloadPodsInput{WorkloadID: validUUID, IncludeMetrics: true, IncludeAdvancedMetrics: true}, false},
		{"advanced metrics require baseline metrics", getWorkloadPodsInput{WorkloadID: validUUID, IncludeAdvancedMetrics: true}, true},
		{"metric limit over max", getWorkloadPodsInput{WorkloadID: validUUID, Limit: 21, IncludeMetrics: true}, true},
		{"negative limit", getWorkloadPodsInput{WorkloadID: validUUID, Limit: -1}, true},
		{"limit over API max", getWorkloadPodsInput{WorkloadID: validUUID, Limit: 501}, true},
		{"negative offset", getWorkloadPodsInput{WorkloadID: validUUID, Offset: -1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetWorkloadPodsRun_TrainingHappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	var gotLimit, gotOffset string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		gotOffset = r.URL.Query().Get("offset")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				{
					"name":              "t1-pod-0",
					"priorityClassName": "train",
					"id":                validPodID,
					"workloadId":        validUUID,
					"clusterId":         "c-1",
					"nodeName":          "node-a",
					"currentNodePool":   "pool-default",
					"k8sPhase":          "Running",
					"createdAt":         "2026-05-17T15:00:00Z",
					"k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
					"updatedAt":         "2026-05-17T15:00:00Z",
					"tenantId":          1,
					"containers":        []any{},
					"requestedResources": map[string]any{
						"gpuRequestType": "portion",
						"gpu":            map[string]any{"request": 1.0, "limit": 2.0},
						"gpuMemory":      map[string]any{"request": "2Gi", "limit": "4Gi"},
						"cpu":            map[string]any{"request": 0.5, "limit": 1.0},
					},
					"allocatedResources": map[string]any{
						"gpu":       1.0,
						"gpuMemory": "2Gi",
						"cpu":       0.5,
					},
				},
			},
		})
	})

	var gotMetricTypes, gotStart, gotEnd string
	var metricCalls int
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricCalls++
		gotMetricTypes = r.URL.Query().Get("metricType")
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{
					"type":   "GPU_UTILIZATION",
					"labels": map[string]string{"gpu": "0"},
					"values": []map[string]any{
						{"value": "0.42", "timestamp": "2026-05-17T14:51:00Z"},
						{"value": "0.50", "timestamp": "2026-05-17T15:00:00Z"},
					},
				},
			},
		})
	})

	out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{
		WorkloadID: validUUID, IncludeMetrics: true, IncludeAdvancedMetrics: true,
	})
	require.NoError(t, err)

	assert.Equal(t, validUUID, out.WorkloadID)
	assert.Equal(t, "Training", out.Type)
	require.Len(t, out.Pods, 1)

	p := out.Pods[0]
	assert.Equal(t, "t1-pod-0", p.Name)
	assert.Equal(t, validPodID, p.ID)
	assert.Equal(t, "node-a", p.NodeName)
	assert.Equal(t, "pool-default", p.CurrentNodePool)
	assert.Equal(t, "Running", p.K8sPhase)
	assert.Equal(t, "2026-05-17T15:00:00Z", p.CreatedAt, "createdAt forwarded as RFC3339 from upstream pods[].createdAt")
	assert.Equal(t, "c-1", p.ClusterID)
	assert.Empty(t, p.RevisionID, "revisionId is empty for non-inference pods")

	require.NotNil(t, p.RequestedResources)
	assert.Equal(t, "portion", p.RequestedResources.GpuRequestType)
	require.NotNil(t, p.RequestedResources.Gpu, "requested.gpu surfaced as a {request, limit} pair, not a single number")
	require.NotNil(t, p.RequestedResources.Gpu.Request)
	require.NotNil(t, p.RequestedResources.Gpu.Limit)
	assert.Equal(t, 1.0, *p.RequestedResources.Gpu.Request)
	assert.Equal(t, 2.0, *p.RequestedResources.Gpu.Limit)
	require.NotNil(t, p.RequestedResources.GpuMemory)
	assert.Equal(t, "2Gi", p.RequestedResources.GpuMemory.Request)
	assert.Equal(t, "4Gi", p.RequestedResources.GpuMemory.Limit)

	require.NotNil(t, p.AllocatedResources)
	require.NotNil(t, p.AllocatedResources.Gpu, "allocated.gpu surfaced as a single granted number")
	assert.Equal(t, 1.0, *p.AllocatedResources.Gpu)
	assert.Equal(t, "2Gi", p.AllocatedResources.GpuMemory)
	require.NotNil(t, p.AllocatedResources.Cpu)
	assert.Equal(t, 0.5, *p.AllocatedResources.Cpu)

	require.NotNil(t, p.Measurements, "measurements should be non-nil on success")
	require.Len(t, *p.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", (*p.Measurements)[0].Type)
	require.Len(t, (*p.Measurements)[0].Values, 2)
	assert.Equal(t, "0.50", (*p.Measurements)[0].Values[1].Value)
	assert.Equal(t, "2026-05-17T15:00:00Z", (*p.Measurements)[0].Values[1].Timestamp)

	assert.Equal(t, "21", gotLimit, "pods fetched with one lookahead record")
	assert.Equal(t, "0", gotOffset, "pods fetched with offset=0")
	for _, want := range []string{"GPU_UTILIZATION", "GPU_MEMORY_USAGE_BYTES", "CPU_USAGE_CORES", "CPU_MEMORY_USAGE_BYTES"} {
		assert.Contains(t, gotMetricTypes, want, "combined pod metrics request includes %s", want)
	}
	assert.Contains(t, gotMetricTypes, "GPU_SM_ACTIVITY_PER_GPU")
	assert.NotEmpty(t, gotStart)
	assert.NotEmpty(t, gotEnd)
	assert.Equal(t, 1, metricCalls, "supported advanced metrics use one combined request")
}

func TestGetWorkloadPodsRun_PaginatesAndOmitsMetricsByDefault(t *testing.T) {
	for _, tc := range []struct {
		name           string
		returnedPods   int
		wantHasMore    bool
		wantNextOffset int32
	}{
		{name: "lookahead pod means more results", returnedPods: 8, wantHasMore: true, wantNextOffset: 27},
		{name: "exact page has no more results", returnedPods: 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
					"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
				})
			})
			mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "8", r.URL.Query().Get("limit"))
				assert.Equal(t, "20", r.URL.Query().Get("offset"))
				pods := make([]map[string]any, tc.returnedPods)
				for i := range pods {
					pods[i] = map[string]any{
						"name": fmt.Sprintf("t1-pod-%d", 20+i), "id": validPodID,
						"workloadId": validUUID, "clusterId": "c-1", "k8sPhase": "Running",
						"createdAt": "2026-05-17T15:00:00Z", "k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
						"updatedAt": "2026-05-17T15:00:00Z", "tenantId": 1, "containers": []any{},
					}
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"next": int32(20 + tc.returnedPods),
					"pods": pods,
				})
			})

			metricsCalled := false
			mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
				metricsCalled = true
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
			})

			out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{
				WorkloadID: validUUID,
				Limit:      7,
				Offset:     20,
			})
			require.NoError(t, err)
			require.Len(t, out.Pods, 7)
			assert.Equal(t, tc.wantHasMore, out.HasMore)
			assert.Equal(t, tc.wantNextOffset, out.NextOffset)
			assert.Nil(t, out.Pods[0].Measurements)
			assert.False(t, metricsCalled, "per-pod metrics endpoint must not be called unless includeMetrics is true")
		})
	}
}

func TestGetWorkloadPodsRun_InferenceUsesInferenceVariant(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "i1",
			"type":           "Inference",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				{
					"name":              "i1-pod-0",
					"priorityClassName": "inf",
					"id":                validPodID,
					"workloadId":        validUUID,
					"clusterId":         "c-1",
					"revisionId":        "11111111-2222-3333-4444-555555555555",
					"k8sPhase":          "Running",
					"createdAt":         "2026-05-17T15:00:00Z",
					"k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
					"updatedAt":         "2026-05-17T15:00:00Z",
					"tenantId":          1,
					"containers":        []any{},
				},
			},
		})
	})

	nonInferenceCalled := false
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		nonInferenceCalled = true
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
	})

	var gotInferenceMetricTypes string
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotInferenceMetricTypes = r.URL.Query().Get("metricType")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "THROUGHPUT", "values": []map[string]any{{"value": "10", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "LATENCY", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})

	out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID, IncludeMetrics: true})
	require.NoError(t, err)

	assert.Equal(t, "Inference", out.Type)
	require.Len(t, out.Pods, 1)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", out.Pods[0].RevisionID, "revisionId surfaced for inference pods")
	require.NotNil(t, out.Pods[0].Measurements)
	require.Len(t, *out.Pods[0].Measurements, 2)

	assert.False(t, nonInferenceCalled, "non-inference per-pod endpoint must not be called for inference workloads")
	for _, want := range []string{"THROUGHPUT", "LATENCY"} {
		assert.Contains(t, gotInferenceMetricTypes, want)
	}
}

func TestGetWorkloadPodsRun_PrimaryWorkloadNotFoundMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadPodsRun_PrimaryWorkloadAuthMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
	})

	_, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func TestGetWorkloadPodsRun_PodsListForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no read permission on pods"})
	})

	_, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetWorkloadPodsRun_PodMetricsFailureKeepsPodWithNullMeasurements(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				{
					"name":              "t1-pod-0",
					"priorityClassName": "train",
					"id":                validPodID,
					"workloadId":        validUUID,
					"clusterId":         "c-1",
					"k8sPhase":          "Running",
					"createdAt":         "2026-05-17T15:00:00Z",
					"k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
					"updatedAt":         "2026-05-17T15:00:00Z",
					"tenantId":          1,
					"containers":        []any{},
				},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getWorkloadPodsRun(ctx, client, getWorkloadPodsInput{WorkloadID: validUUID, IncludeMetrics: true})
	require.NoError(t, err, "per-pod metrics failure must not fail the tool")
	require.Len(t, out.Pods, 1)
	assert.Nil(t, out.Pods[0].Measurements, "measurements must be nil when the metrics call failed")
	require.Len(t, out.Pods[0].MetricWarnings, 1)
	assert.Contains(t, out.Pods[0].MetricWarnings[0], "resource metrics")

	got := buf.String()
	for _, want := range []string{
		`"level":"warn"`,
		`"workload_id":"` + validUUID + `"`,
		`"pod_id":"` + validPodID + `"`,
		`per-pod metrics fetch failed, returning measurements=null`,
	} {
		assert.Contains(t, got, want)
	}
}

func TestGetWorkloadPodsRun_EmptyPodsList(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Training",
			"phase":          "Pending",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})

	out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.NotNil(t, out.Pods, "pods must be a non-nil empty slice")
	assert.Len(t, out.Pods, 0)
}

func TestAttachPodMeasurements_AdvancedFailureKeepsBasicMetrics(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var starts, ends []string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		starts = append(starts, r.URL.Query().Get("start"))
		ends = append(ends, r.URL.Query().Get("end"))
		if strings.Contains(r.URL.Query().Get("metricType"), "GPU_SM_ACTIVITY_PER_GPU") {
			runaiclienttest.WriteJSON(w, http.StatusBadRequest, map[string]any{"message": "Advanced GPU metrics are not enabled"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{{
				"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}},
			}},
		})
	})

	start := time.Date(2026, 5, 17, 14, 50, 0, 0, time.UTC)
	end := start.Add(10 * time.Minute)
	entry := podOutput{}
	attachPodMeasurements(context.Background(), client, validUUID, validPodID, false, true, start, end, &entry)

	require.NotNil(t, entry.Measurements)
	require.Len(t, *entry.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", (*entry.Measurements)[0].Type)
	require.Len(t, entry.MetricWarnings, 1)
	assert.Contains(t, entry.MetricWarnings[0], "advanced GPU metrics")
	require.Len(t, starts, 2)
	assert.Equal(t, starts[0], starts[1])
	assert.Equal(t, ends[0], ends[1])
}

func TestAttachPodMeasurements_DefaultSkipsAdvancedMetrics(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var calls int
	var gotMetricTypes string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		calls++
		gotMetricTypes = r.URL.Query().Get("metricType")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
	})

	end := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	entry := podOutput{}
	attachPodMeasurements(context.Background(), client, validUUID, validPodID, false, false, end.Add(-10*time.Minute), end, &entry)

	assert.Equal(t, 1, calls, "includeMetrics without includeAdvancedMetrics makes one baseline request")
	assert.Contains(t, gotMetricTypes, "GPU_UTILIZATION")
	assert.NotContains(t, gotMetricTypes, "GPU_SM_ACTIVITY_PER_GPU")
	require.NotNil(t, entry.Measurements)
	assert.Empty(t, entry.MetricWarnings)
}

func TestAttachPodMeasurements_MultiStatusNamesUnavailableMetrics(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("metricType"), "GPU_SM_ACTIVITY_PER_GPU") {
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "CPU_USAGE_CORES"},
			},
		})
	})

	end := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	entry := podOutput{}
	attachPodMeasurements(context.Background(), client, validUUID, validPodID, false, false, end.Add(-10*time.Minute), end, &entry)

	require.NotNil(t, entry.Measurements)
	require.Len(t, *entry.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", (*entry.Measurements)[0].Type)
	require.Len(t, entry.MetricWarnings, 1)
	assert.Contains(t, entry.MetricWarnings[0], "resource metrics were partially retrieved")
	assert.Contains(t, entry.MetricWarnings[0], "CPU_USAGE_CORES")
}

func TestGetWorkloadPodsRun_MaxPageUsesOneRecordLookahead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		hasMore bool
	}{
		{name: "exactly 500 pods", hasMore: false},
		{name: "more than 500 pods", hasMore: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
					"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
				})
			})

			pod := map[string]any{
				"name": "t1-pod", "id": validPodID, "workloadId": validUUID, "clusterId": "c-1",
				"k8sPhase": "Running", "createdAt": "2026-05-17T15:00:00Z", "k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
				"updatedAt": "2026-05-17T15:00:00Z", "tenantId": 1, "containers": []any{},
			}
			page := make([]map[string]any, podsMaxLimit)
			for i := range page {
				page[i] = pod
			}
			requests := 0
			mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods", func(w http.ResponseWriter, r *http.Request) {
				requests++
				switch r.URL.Query().Get("offset") {
				case "0":
					assert.Equal(t, "500", r.URL.Query().Get("limit"))
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": page, "next": int32(500)})
				case "500":
					assert.Equal(t, "1", r.URL.Query().Get("limit"))
					lookahead := []map[string]any{}
					if tc.hasMore {
						lookahead = append(lookahead, pod)
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": lookahead})
				default:
					t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
				}
			})

			out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID, Limit: podsMaxLimit})
			require.NoError(t, err)
			require.Len(t, out.Pods, int(podsMaxLimit))
			assert.Equal(t, tc.hasMore, out.HasMore)
			if tc.hasMore {
				assert.Equal(t, podsMaxLimit, out.NextOffset)
			} else {
				assert.Zero(t, out.NextOffset)
			}
			assert.Equal(t, 2, requests)
		})
	}
}
