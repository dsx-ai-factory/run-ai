/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"net/http"
	"strings"
	"testing"

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

	var gotMetricTypes string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/pods/"+validPodID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotMetricTypes = r.URL.Query().Get("metricType")
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

	out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
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

	assert.Equal(t, "20", gotLimit, "pods fetched with limit=20")
	assert.Equal(t, "0", gotOffset, "pods fetched with offset=0")
	for _, want := range []string{"GPU_UTILIZATION", "GPU_MEMORY_USAGE_BYTES", "CPU_USAGE_CORES", "CPU_MEMORY_USAGE_BYTES"} {
		assert.Contains(t, gotMetricTypes, want, "non-inference pod metrics request includes %s", want)
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

	out, err := getWorkloadPodsRun(context.Background(), client, getWorkloadPodsInput{WorkloadID: validUUID})
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

	out, err := getWorkloadPodsRun(ctx, client, getWorkloadPodsInput{WorkloadID: validUUID})
	require.NoError(t, err, "per-pod metrics failure must not fail the tool")
	require.Len(t, out.Pods, 1)
	assert.Nil(t, out.Pods[0].Measurements, "measurements must be nil when the metrics call failed")

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
