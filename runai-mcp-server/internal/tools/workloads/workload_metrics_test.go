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
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestGetWorkloadMetricsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadMetricsInput
		wantErr bool
	}{
		{"empty workloadId", getWorkloadMetricsInput{}, true},
		{"non-uuid workloadId", getWorkloadMetricsInput{WorkloadID: "wl-1"}, true},
		{"valid no time", getWorkloadMetricsInput{WorkloadID: validUUID}, false},
		{"valid with full window", getWorkloadMetricsInput{WorkloadID: validUUID, Start: "2026-05-17T14:00:00Z", End: "2026-05-17T15:00:00Z", NumberOfSamples: 50}, false},
		{"start not RFC3339", getWorkloadMetricsInput{WorkloadID: validUUID, Start: "not-a-time"}, true},
		{"end not RFC3339", getWorkloadMetricsInput{WorkloadID: validUUID, End: "not-a-time"}, true},
		{"end before start", getWorkloadMetricsInput{WorkloadID: validUUID, Start: "2026-05-17T15:00:00Z", End: "2026-05-17T14:00:00Z"}, true},
		{"end equals start", getWorkloadMetricsInput{WorkloadID: validUUID, Start: "2026-05-17T15:00:00Z", End: "2026-05-17T15:00:00Z"}, true},
		{"negative samples", getWorkloadMetricsInput{WorkloadID: validUUID, NumberOfSamples: -1}, true},
		{"samples over max", getWorkloadMetricsInput{WorkloadID: validUUID, NumberOfSamples: 1001}, true},
		{"samples at max", getWorkloadMetricsInput{WorkloadID: validUUID, NumberOfSamples: 1000}, false},
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

func TestGetWorkloadMetricsRun_TrainingHappyPath(t *testing.T) {
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
			"additionalFields": map[string]any{
				"nimWorkload": "true",
			},
		})
	})

	var resourceTypes, nimTypes, gotStart, gotEnd, gotSamples string
	var resourceCalls int
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		mt := r.URL.Query().Get("metricType")
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		gotSamples = r.URL.Query().Get("numberOfSamples")
		if strings.Contains(mt, "NIM_") {
			nimTypes = mt
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"measurements": []map[string]any{
					{"type": "NIM_NUM_REQUESTS_RUNNING", "values": []map[string]any{{"value": "3", "timestamp": "2026-05-17T15:00:00Z"}}},
				},
			})
			return
		}
		resourceCalls++
		resourceTypes = mt
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{
					"type":   "GPU_UTILIZATION",
					"labels": map[string]string{"gpu": "0"},
					"values": []map[string]any{
						{"value": "0.42", "timestamp": "2026-05-17T14:00:00Z"},
						{"value": "0.50", "timestamp": "2026-05-17T15:00:00Z"},
					},
				},
				{
					"type":   "CPU_USAGE_CORES",
					"values": []map[string]any{{"value": "0.7", "timestamp": "2026-05-17T15:00:00Z"}},
				},
			},
		})
	})

	in := getWorkloadMetricsInput{
		WorkloadID:      validUUID,
		Start:           "2026-05-17T14:00:00Z",
		End:             "2026-05-17T15:00:00Z",
		NumberOfSamples: 50,
	}
	out, err := getWorkloadMetricsRun(context.Background(), client, in)
	require.NoError(t, err)

	assert.Equal(t, validUUID, out.WorkloadID)
	assert.Equal(t, "Training", out.Type)
	require.Len(t, out.Measurements, 3, "resource (2) + NIM (1) measurements concatenated")
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	assert.Equal(t, "CPU_USAGE_CORES", out.Measurements[1].Type)
	assert.Equal(t, "NIM_NUM_REQUESTS_RUNNING", out.Measurements[2].Type, "NIM measurements appended after resource")
	assert.Empty(t, out.Warnings)

	for _, want := range []string{"GPU_UTILIZATION", "GPU_MEMORY_USAGE_BYTES", "CPU_USAGE_CORES", "CPU_MEMORY_USAGE_BYTES", "NVLINK_BANDWIDTH_TOTAL"} {
		assert.Contains(t, resourceTypes, want, "combined resource batch includes %s", want)
	}
	assert.Equal(t, 1, resourceCalls, "supported NVLink uses one combined request")
	assert.NotContains(t, resourceTypes, "NIM_", "resource batch does not mix in NIM types")
	for _, want := range []string{"NIM_NUM_REQUESTS_RUNNING", "NIM_REQUEST_SUCCESS_TOTAL", "NIM_REQUEST_FAILURE_TOTAL", "NIM_GPU_CACHE_USAGE_PERC", "NIM_E2E_REQUEST_LATENCY_SECONDS"} {
		assert.Contains(t, nimTypes, want, "NIM batch includes %s", want)
	}
	assert.Equal(t, "50", gotSamples)
	startT, err := time.Parse(time.RFC3339, gotStart)
	require.NoError(t, err)
	endT, err := time.Parse(time.RFC3339, gotEnd)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-17T14:00:00Z", startT.UTC().Format(time.RFC3339))
	assert.Equal(t, "2026-05-17T15:00:00Z", endT.UTC().Format(time.RFC3339))
}

func TestGetWorkloadMetricsRun_DefaultsAppliedWhenInputsOmitted(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Workspace",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	var gotStart, gotEnd, gotSamples string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		gotSamples = r.URL.Query().Get("numberOfSamples")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
	})

	before := time.Now().UTC()
	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	after := time.Now().UTC()
	require.NoError(t, err)
	assert.NotNil(t, out.Measurements)

	assert.Equal(t, "20", gotSamples, "numberOfSamples defaults to 20")
	endT, err := time.Parse(time.RFC3339, gotEnd)
	require.NoError(t, err)
	startT, err := time.Parse(time.RFC3339, gotStart)
	require.NoError(t, err)
	assert.True(t, !endT.Before(before.Truncate(time.Second)) && !endT.After(after.Add(time.Second)),
		"end defaults to ~now (got %s, before=%s, after=%s)", endT, before, after)
	assert.WithinDuration(t, endT.Add(-time.Hour), startT, time.Second, "start defaults to end - 1h")
}

func TestGetWorkloadMetricsRun_InferenceAppendsThroughputLatency(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("metricType"), "NIM_") {
			runaiclienttest.WriteJSON(w, http.StatusBadRequest, map[string]any{"message": "workload is not identified as NVIDIA NIM"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "CPU_USAGE_CORES", "values": []map[string]any{{"value": "0.7", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "THROUGHPUT", "values": []map[string]any{{"value": "10", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "LATENCY", "values": []map[string]any{{"value": "0.1", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Measurements, 4, "inference appends throughput+latency after primary measurements")
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	assert.Equal(t, "CPU_USAGE_CORES", out.Measurements[1].Type)
	assert.Equal(t, "THROUGHPUT", out.Measurements[2].Type, "inference measurements appended after primary")
	assert.Equal(t, "LATENCY", out.Measurements[3].Type)
}

func TestGetWorkloadMetricsRun_InferenceOneEndpointFailsReturnsAvailable(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("metricType"), "NIM_") {
			runaiclienttest.WriteJSON(w, http.StatusBadRequest, map[string]any{"message": "workload is not identified as NVIDIA NIM"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getWorkloadMetricsRun(ctx, client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err, "one inference-side failure must not fail the tool")
	require.Len(t, out.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "inference serving metrics")

	got := buf.String()
	assert.Contains(t, got, `serving metrics unavailable`)
}

func TestGetWorkloadMetricsRun_AllExpectedGroupsFailListsGroupsAndPreservesPrimaryError(t *testing.T) {
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
			"additionalFields": map[string]any{
				"nimWorkload": "true",
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("metricType"), "NIM_") {
			runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "NIM metrics unavailable"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "metrics not found"})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	_, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "all-fail path preserves the primary resource error classification")
	for _, group := range []string{"resource", "NVLink", "NIM", "inference serving"} {
		assert.ErrorContains(t, err, group)
	}
}

func TestGetWorkloadMetricsRun_InferenceGPUFailsInferenceOK(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var resourceCalls int

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "i1",
			"type":           "DistributedInference",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		resourceCalls++
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "THROUGHPUT", "values": []map[string]any{{"value": "10", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Measurements, 1)
	assert.Equal(t, "THROUGHPUT", out.Measurements[0].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "resource and NVLink metrics")
	assert.Equal(t, 1, resourceCalls, "a 500 response is not retried as though NVLink were unsupported")
}

func TestGetWorkloadMetricsRun_NonInferenceMetricsFailIsHardError(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no perm"})
	})

	_, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestIsNIMWorkload(t *testing.T) {
	for _, tc := range []struct {
		name string
		wl   *workloadsapi.WorkloadDetailed
		want bool
	}{
		{name: "nil workload", want: false},
		{name: "missing marker", wl: &workloadsapi.WorkloadDetailed{}, want: false},
		{name: "non-string marker", wl: &workloadsapi.WorkloadDetailed{AdditionalFields: map[string]any{"nimWorkload": true}}, want: false},
		{name: "false string marker", wl: &workloadsapi.WorkloadDetailed{AdditionalFields: map[string]any{"nimWorkload": "false"}}, want: false},
		{name: "true string marker", wl: &workloadsapi.WorkloadDetailed{AdditionalFields: map[string]any{"nimWorkload": "true"}}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isNIMWorkload(tc.wl))
		})
	}
}

func TestGetWorkloadMetricsRun_NIMFailureReturnsPartialWarning(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "nim-1",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
			"additionalFields": map[string]any{
				"nimWorkload": "true",
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("metricType"), "NIM_") {
			runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "NIM metrics backend unavailable"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err, "a failed NIM group must not hide available resource metrics")
	require.Len(t, out.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "NIM metrics")
}

func TestGetWorkloadMetricsRun_ResourceMultiStatusNamesUnavailableMetric(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var resourceCalls int

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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		resourceCalls++
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "GPU_MEMORY_USAGE_BYTES"},
			},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Measurements, 1, "the values-less failure placeholder must not be returned as a valid series")
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "resource and NVLink metrics")
	assert.Contains(t, out.Warnings[0], "GPU_MEMORY_USAGE_BYTES")
	assert.Equal(t, 1, resourceCalls, "a 207 response already describes partial data and is not retried")
}

func TestGetWorkloadMetricsRun_InferenceMultiStatusNamesUnavailableMetric(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{
				{"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}}},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{
				{"type": "THROUGHPUT", "values": []map[string]any{{"value": "10", "timestamp": "2026-05-17T15:00:00Z"}}},
				{"type": "LATENCY"},
			},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Measurements, 2)
	assert.Equal(t, "THROUGHPUT", out.Measurements[1].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "inference serving metrics")
	assert.Contains(t, out.Warnings[0], "LATENCY")
}

func TestGetWorkloadMetricsRun_InferenceCoreFailsServingEmptyKeepsMeasurementsNonNil(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var resourceCalls int

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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		resourceCalls++
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/metrics", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err, "one inference-side failure must not fail the tool")
	require.NotNil(t, out.Measurements, "measurements must be a non-nil empty slice (renders [] in JSON, not null) when the surviving side is empty")
	assert.Len(t, out.Measurements, 0)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "resource and NVLink metrics")
	assert.Equal(t, 1, resourceCalls)
}

func TestGetWorkloadMetricsRun_PrimaryWorkloadNotFoundMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadMetricsRun_NVLinkUnsupportedKeepsBaselineMetrics(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var resourceCalls int
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		resourceCalls++
		if strings.Contains(r.URL.Query().Get("metricType"), "NVLINK_") {
			runaiclienttest.WriteJSON(w, http.StatusBadRequest, map[string]any{"message": "metric requires cluster version 2.23"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"measurements": []map[string]any{{
				"type": "GPU_UTILIZATION", "values": []map[string]any{{"value": "0.5", "timestamp": "2026-05-17T15:00:00Z"}},
			}},
		})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Measurements, 1)
	assert.Equal(t, "GPU_UTILIZATION", out.Measurements[0].Type)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "NVLink metrics")
	assert.Equal(t, 2, resourceCalls, "a validation failure retries once without NVLink")
}

func TestGetWorkloadMetricsRun_MapsNIMHistogram(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "nim-1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
			"additionalFields": map[string]any{"nimWorkload": "true"},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricTypes := r.URL.Query().Get("metricType")
		if strings.Contains(metricTypes, "NIM_") {
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"measurements": []any{},
				"histogram": []map[string]any{{
					"type": "NIM_E2E_REQUEST_LATENCY_SECONDS",
					"values": []map[string]any{{
						"timestamp": "2026-05-17T15:00:00Z",
						"data":      map[string]string{"0.1": "2", "+Inf": "3"},
					}},
				}},
			})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"measurements": []any{}})
	})

	out, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.Histograms, 1)
	assert.Equal(t, "NIM_E2E_REQUEST_LATENCY_SECONDS", out.Histograms[0].Type)
	require.Len(t, out.Histograms[0].Values, 1)
	assert.Equal(t, "2026-05-17T15:00:00Z", out.Histograms[0].Values[0].Timestamp)
	assert.Equal(t, "3", out.Histograms[0].Values[0].Data["+Inf"])
	assert.Empty(t, out.Warnings)
}

func TestGetWorkloadMetricsRun_AllPlaceholderMultiStatusIsFailure(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricType := "GPU_UTILIZATION"
		if strings.Contains(r.URL.Query().Get("metricType"), "NVLINK_") {
			metricType = "NVLINK_BANDWIDTH_TOTAL"
		}
		runaiclienttest.WriteJSON(w, http.StatusMultiStatus, map[string]any{
			"measurements": []map[string]any{{"type": metricType}},
		})
	})

	_, err := getWorkloadMetricsRun(context.Background(), client, getWorkloadMetricsInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "resource")
	assert.ErrorContains(t, err, "NVLink")
}
