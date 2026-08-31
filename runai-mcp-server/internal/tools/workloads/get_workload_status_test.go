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

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func TestGetWorkloadStatusInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadStatusInput
		wantErr bool
	}{
		{"empty", getWorkloadStatusInput{WorkloadID: ""}, true},
		{"whitespace", getWorkloadStatusInput{WorkloadID: "   "}, true},
		{"not a uuid", getWorkloadStatusInput{WorkloadID: "t1"}, true},
		{"valid", getWorkloadStatusInput{WorkloadID: validUUID}, false},
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

func TestGetWorkloadStatusRun_PendingWithWarningEvent(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotEventsLimit, gotEventsSortOrder string

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"clusterId":      "c-1",
			"type":           "Training",
			"phase":          "Pending",
			"phaseReason":    "OverLimit",
			"projectName":    "p1",
			"departmentName": "default",
			"createdAt":      "2026-05-17T15:47:12Z",
			"phaseUpdatedAt": "2026-05-17T15:47:17Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, r *http.Request) {
		gotEventsLimit = r.URL.Query().Get("limit")
		gotEventsSortOrder = r.URL.Query().Get("sortOrder")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"events": []map[string]any{
				{"type": "Warning", "reason": "Unschedulable", "message": "MaxNodePoolResource", "createdAt": "2026-05-17T15:47:21Z"},
				{"type": "Normal", "reason": "SuccessfulCreate", "message": "Created pod", "createdAt": "2026-05-17T15:47:17Z"},
			},
		})
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)

	assert.Equal(t, validUUID, out.ID)
	assert.Equal(t, "Training", out.Type, "type forwarded verbatim from GET /api/v1/workloads/{id}.type")
	assert.Equal(t, "Pending", out.Phase)
	assert.Equal(t, "OverLimit", out.PhaseReason)
	assert.Equal(t, "c-1", out.ClusterID)
	assert.Equal(t, "p1", out.ProjectName)
	assert.Equal(t, "default", out.DepartmentName)
	assert.Equal(t, "2026-05-17T15:47:12Z", out.CreatedAt)
	assert.Equal(t, "2026-05-17T15:47:17Z", out.PhaseUpdatedAt)

	require.NotNil(t, out.Events, "events should be non-nil on success")
	require.Len(t, *out.Events, 2)
	assert.Equal(t, "Warning", (*out.Events)[0].Type, "first event type forwarded verbatim")
	assert.Equal(t, "Unschedulable", (*out.Events)[0].Reason)

	assert.Equal(t, "5", gotEventsLimit, "events fetched with the tool-side recent limit")
	assert.Equal(t, "desc", gotEventsSortOrder, "events fetched newest-first")
}

func TestGetWorkloadStatusRun_RunningWithImmutableAttachedResourceFailure(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	const diagnostic = "failed to create/update attached resource: Resources \"shared-memory\" is invalid: spec: field is immutable"
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "workspace-with-issue",
			"type":           "Workspace",
			"phase":          "Running",
			"phaseMessage":   diagnostic,
			"k8sPhase":       "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T16:00:00Z",
			"conditions": []map[string]any{{
				"type":               "Failed",
				"status":             "True",
				"reason":             "AttachedResourceUpdateFailed",
				"message":            diagnostic,
				"lastTransitionTime": "2026-05-17T19:00:00+03:00",
			}},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"events": []any{}})
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "Running", out.Phase)
	assert.Equal(t, diagnostic, out.PhaseMessage)
	assert.Equal(t, "Running", out.K8sPhase)
	require.Len(t, out.Conditions, 1)
	condition := out.Conditions[0]
	assert.Equal(t, "Failed", condition.Type)
	assert.Equal(t, "True", condition.Status)
	assert.Equal(t, "AttachedResourceUpdateFailed", condition.Reason)
	assert.Equal(t, diagnostic, condition.Message)
	assert.Equal(t, "2026-05-17T16:00:00Z", condition.LastTransitionTime)
}

func TestGetWorkloadStatusRun_RunningHealthy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "ws",
			"type":           "Workspace",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T16:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"events": []map[string]any{
				{"type": "Normal", "reason": "Sync", "message": "ok", "createdAt": "2026-05-17T16:00:00Z"},
			},
		})
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "Workspace", out.Type, "type forwarded verbatim")
	assert.Equal(t, "Running", out.Phase)
	assert.Empty(t, out.PhaseMessage)
	assert.Empty(t, out.PhaseReason)
	assert.Empty(t, out.K8sPhase)
	assert.NotNil(t, out.Conditions, "conditions should always be an array")
	assert.Empty(t, out.Conditions)
	assert.Empty(t, out.DepartmentName, "omitempty: empty when upstream omits departmentName")
	require.NotNil(t, out.Events)
	assert.Len(t, *out.Events, 1)
}

func TestGetWorkloadStatusRun_EventsEmptyButCallSucceeded(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "i1",
			"type":           "Inference",
			"phase":          "Failed",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"events": []map[string]any{}})
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.NotNil(t, out.Events, "events should be non-nil ([]) when the call succeeded with zero events")
	assert.Len(t, *out.Events, 0)
}

func TestGetWorkloadStatusRun_NilPrimaryResponseIsUpstreamError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	_, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestGetWorkloadStatusRun_NilEventsResponseRendersNullEvents(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Nil(t, out.Events)
}

func TestGetWorkloadStatusRun_PrimaryNotFoundMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadStatusRun_PrimaryAuthMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
	})

	_, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func TestGetWorkloadStatusRun_EventsFailureKeepsPrimaryRendersNullEvents(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "upstream throttled"})
	})

	out, err := getWorkloadStatusRun(context.Background(), client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err, "events failure must not fail the tool")
	assert.Nil(t, out.Events, "events must be nil (renders JSON null) when the events call failed")
	assert.Equal(t, "Pending", out.Phase, "primary fields still populated")
}

func TestGetWorkloadStatusRun_EventsFailureLogsWarning(t *testing.T) {
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
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "upstream throttled"})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	_, err := getWorkloadStatusRun(ctx, client, getWorkloadStatusInput{WorkloadID: validUUID})
	require.NoError(t, err)
	got := buf.String()
	for _, want := range []string{
		`"level":"warn"`,
		`"workload_id":"` + validUUID + `"`,
		`events fetch failed, returning events=null`,
	} {
		assert.Contains(t, got, want, "expected log to contain %q; got %q", want, got)
	}
}
