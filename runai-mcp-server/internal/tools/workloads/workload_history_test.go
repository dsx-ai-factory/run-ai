/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const (
	validRevisionA           = "11111111-2222-3333-4444-555555555555"
	validRevisionB           = "11111111-2222-3333-4444-aaaaaaaaaaaa"
	validForeignWorkloadUUID = "99999999-8888-7777-6666-555555555555"
)

func TestGetWorkloadHistoryInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadHistoryInput
		wantErr bool
	}{
		{"empty workloadId", getWorkloadHistoryInput{}, true},
		{"non-uuid workloadId", getWorkloadHistoryInput{WorkloadID: "wl-1"}, true},
		{"valid minimal", getWorkloadHistoryInput{WorkloadID: validUUID}, false},
		{"valid with limit+offset", getWorkloadHistoryInput{WorkloadID: validUUID, Limit: 10, Offset: 20}, false},
		{"limit at max", getWorkloadHistoryInput{WorkloadID: validUUID, Limit: historyMaxLimit}, false},
		{"limit over max", getWorkloadHistoryInput{WorkloadID: validUUID, Limit: historyMaxLimit + 1}, true},
		{"negative limit", getWorkloadHistoryInput{WorkloadID: validUUID, Limit: -1}, true},
		{"negative offset", getWorkloadHistoryInput{WorkloadID: validUUID, Offset: -1}, true},
		{"valid revisionId", getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA}, false},
		{"non-uuid revisionId", getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: "rev-1"}, true},
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

func TestGetWorkloadHistoryRun_NonInferenceReturnsRecordsOnly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var gotLimit, gotOffset string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		gotOffset = r.URL.Query().Get("offset")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"records": []map[string]any{
				{
					"meta": map[string]any{"creationTimestamp": "2026-05-17T15:00:00Z", "type": "PhaseUpdate"},
					"spec": map[string]any{"phaseUpdate": map[string]any{"phase": "Running", "phaseMessage": "scheduled"}},
				},
				{
					"meta": map[string]any{"creationTimestamp": "2026-05-17T15:01:00Z", "type": "Event"},
					"spec": map[string]any{"event": map[string]any{"type": "Normal", "reason": "Started", "message": "pod up", "clusterId": "c-1"}},
				},
			},
			"next": 20,
		})
	})
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

	revisionsCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		revisionsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err)

	assert.Equal(t, validUUID, out.WorkloadID)
	assert.Equal(t, "Training", out.Type)
	require.Len(t, out.Records, 2)

	assert.Equal(t, "PhaseUpdate", out.Records[0].Meta.Type)
	require.NotNil(t, out.Records[0].Spec.PhaseUpdate)
	assert.Equal(t, "Running", out.Records[0].Spec.PhaseUpdate.Phase)
	assert.Equal(t, "scheduled", out.Records[0].Spec.PhaseUpdate.PhaseMessage)
	assert.Nil(t, out.Records[0].Spec.Event, "phaseUpdate records do not carry event")

	assert.Equal(t, "Event", out.Records[1].Meta.Type)
	require.NotNil(t, out.Records[1].Spec.Event)
	assert.Equal(t, "Normal", out.Records[1].Spec.Event.Type)
	assert.Equal(t, "Started", out.Records[1].Spec.Event.Reason)

	require.NotNil(t, out.Next)
	assert.Equal(t, int32(20), *out.Next, "next forwarded verbatim for pagination round-trip")
	assert.Nil(t, out.Revisions, "non-inference workloads do not include revisions[]")

	assert.False(t, revisionsCalled.Load(), "revisions endpoint must not be called for non-inference workloads")
	assert.Equal(t, "20", gotLimit, "default limit is 20")
	assert.Equal(t, "0", gotOffset, "default offset is 0")
}

func TestGetWorkloadHistoryRun_LimitAndOffsetForwarded(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var gotLimit, gotOffset string
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		gotOffset = r.URL.Query().Get("offset")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Workspace", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, Limit: 50, Offset: 100})
	require.NoError(t, err)
	assert.Equal(t, "50", gotLimit)
	assert.Equal(t, "100", gotOffset)
}

func TestGetWorkloadHistoryRun_NextNullWhenNoNextPage(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Nil(t, out.Next, "next is nil/absent when upstream omits it (no synthesized has_more)")
}

func TestGetWorkloadHistoryRun_InferenceReturnsRevisionsList(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"revisions": []map[string]any{
				{
					"id":           validRevisionA,
					"name":         "rev-a",
					"type":         "InferenceRevision",
					"workloadId":   validUUID,
					"tenantId":     1,
					"clusterId":    "c-1",
					"projectId":    "11",
					"departmentId": "22",
					"createdAt":    "2026-05-17T14:00:00Z",
					"phase":        "Running",
					"requestedPods": map[string]any{
						"min":    int32(1),
						"max":    int32(3),
						"number": int32(2),
					},
				},
				{
					"id":           validRevisionB,
					"name":         "rev-b",
					"type":         "InferenceRevision",
					"workloadId":   validUUID,
					"tenantId":     1,
					"clusterId":    "c-1",
					"projectId":    "11",
					"departmentId": "22",
					"createdAt":    "2026-05-17T13:00:00Z",
					"phase":        "Completed",
				},
			},
		})
	})

	podsCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		podsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err)

	assert.Equal(t, "Inference", out.Type)
	require.NotNil(t, out.Revisions, "revisions surfaced when lookup succeeded")
	revisions := *out.Revisions
	require.Len(t, revisions, 2, "both revisions surfaced without revisionId drill-in")

	assert.Equal(t, validRevisionA, revisions[0].ID)
	assert.Equal(t, "rev-a", revisions[0].Name)
	assert.Equal(t, "Running", revisions[0].Phase)
	require.NotNil(t, revisions[0].RequestedPods)
	require.NotNil(t, revisions[0].RequestedPods.Min)
	assert.Equal(t, int32(1), *revisions[0].RequestedPods.Min)
	require.NotNil(t, revisions[0].RequestedPods.Max)
	assert.Equal(t, int32(3), *revisions[0].RequestedPods.Max)
	assert.Nil(t, revisions[0].Pods, "pods only attached when matching revisionId input")
	assert.Nil(t, revisions[0].Events, "events only attached when matching revisionId input")

	assert.Equal(t, validRevisionB, revisions[1].ID)
	assert.Equal(t, "Completed", revisions[1].Phase)

	assert.False(t, podsCalled.Load(), "revision pods not fetched without revisionId input")
}

func TestGetWorkloadHistoryRun_InferenceWithRevisionIDAttachesDetail(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"revisions": []map[string]any{
				{"id": validRevisionA, "name": "rev-a", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T14:00:00Z", "phase": "Running"},
				{"id": validRevisionB, "name": "rev-b", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T13:00:00Z", "phase": "Completed"},
			},
		})
	})

	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				{
					"name":              "rev-a-pod-0",
					"priorityClassName": "inf",
					"id":                validPodID,
					"workloadId":        validUUID,
					"clusterId":         "c-1",
					"k8sPhase":          "Running",
					"createdAt":         "2026-05-17T14:01:00Z",
					"k8sPhaseUpdatedAt": "2026-05-17T14:01:00Z",
					"updatedAt":         "2026-05-17T14:01:00Z",
					"tenantId":          1,
					"containers":        []any{},
				},
			},
		})
	})

	var gotEventsLimit string
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/events", func(w http.ResponseWriter, r *http.Request) {
		gotEventsLimit = r.URL.Query().Get("limit")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"events": []map[string]any{
				{"type": "Normal", "reason": "Pulled", "message": "image pulled", "source": "kubelet", "createdAt": "2026-05-17T14:00:30Z"},
			},
		})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.NoError(t, err)
	require.NotNil(t, out.Revisions)
	revisions := *out.Revisions
	require.Len(t, revisions, 1, "with revisionId, output contains only the matched revision")
	assert.Equal(t, validRevisionA, revisions[0].ID)
	require.NotNil(t, revisions[0].Pods, "matching revision has pods attached")
	require.Len(t, *revisions[0].Pods, 1)
	assert.Equal(t, "rev-a-pod-0", (*revisions[0].Pods)[0].Name)
	assert.Nil(t, (*revisions[0].Pods)[0].Measurements, "revision pods do not carry measurements")
	require.NotNil(t, revisions[0].Events)
	require.Len(t, *revisions[0].Events, 1)
	assert.Equal(t, "Pulled", (*revisions[0].Events)[0].Reason)
	assert.Equal(t, "kubelet", (*revisions[0].Events)[0].Source, "revision events surface upstream source (parity with history-record events)")
	assert.Equal(t, "5", gotEventsLimit, "revision events fetched newest-5")
}

func TestGetWorkloadHistoryRun_InferenceRevisionIDFallback(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"revisions": []map[string]any{
				{"id": validRevisionB, "name": "rev-b", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T13:00:00Z", "phase": "Completed"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validRevisionA, "name": "rev-a-deleted", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T10:00:00Z", "phase": "Stopped",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"events": []any{}})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.NoError(t, err)
	require.NotNil(t, out.Revisions)
	revisions := *out.Revisions
	require.Len(t, revisions, 1, "with revisionId, output contains only the matched revision (fallback in this case)")
	assert.Equal(t, validRevisionA, revisions[0].ID)
	assert.Equal(t, "rev-a-deleted", revisions[0].Name, "fallback resolves to the single-revision endpoint")
	require.NotNil(t, revisions[0].Pods, "fallback revision still gets pods (non-nil pointer renders as [] for empty fetch)")
	assert.Empty(t, *revisions[0].Pods)
	require.NotNil(t, revisions[0].Events, "fallback revision still gets events (non-nil pointer renders as [] for empty fetch)")
	assert.Empty(t, *revisions[0].Events)
}

func TestGetWorkloadHistoryRun_WorkloadTypeLookupFailureKeepsRecordsOmitsType(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"records": []map[string]any{
				{"meta": map[string]any{"creationTimestamp": "2026-05-17T15:00:00Z", "type": "PhaseUpdate"}, "spec": map[string]any{"phaseUpdate": map[string]any{"phase": "Running", "phaseMessage": ""}}},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getWorkloadHistoryRun(ctx, client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err, "secondary type lookup failure must not fail the tool")
	assert.Empty(t, out.Type, "type omitted when workload-type lookup failed")
	require.Len(t, out.Records, 1, "records still surfaced")
	assert.Contains(t, buf.String(), "workload type lookup failed")
}

func TestGetWorkloadHistoryRun_RevisionsLookupFailureOmitsRevisions(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getWorkloadHistoryRun(ctx, client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err, "revisions lookup failure must not fail the tool")
	assert.Nil(t, out.Revisions, "revisions[] omitted when revisions lookup failed")
	assert.Contains(t, buf.String(), "revisions lookup failed")
}

func TestGetWorkloadHistoryRun_RevisionPodsFailureOmitsPodsOnly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"revisions": []map[string]any{
				{"id": validRevisionA, "name": "rev-a", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T14:00:00Z", "phase": "Running"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"events": []map[string]any{
				{"type": "Normal", "reason": "Pulled", "message": "ok", "createdAt": "2026-05-17T14:00:30Z"},
			},
		})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.NoError(t, err)
	require.NotNil(t, out.Revisions)
	revisions := *out.Revisions
	require.Len(t, revisions, 1)
	assert.Nil(t, revisions[0].Pods, "revision.pods omitted when revision pods call failed")
	require.NotNil(t, revisions[0].Events, "revision.events kept independently")
	assert.Len(t, *revisions[0].Events, 1)
}

func TestGetWorkloadHistoryRun_RevisionEventsFailureOmitsEventsOnly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"revisions": []map[string]any{
				{"id": validRevisionA, "name": "rev-a", "type": "InferenceRevision", "workloadId": validUUID, "tenantId": 1, "clusterId": "c-1", "projectId": "11", "departmentId": "22", "createdAt": "2026-05-17T14:00:00Z", "phase": "Running"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/events", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.NoError(t, err)
	require.NotNil(t, out.Revisions)
	revisions := *out.Revisions
	require.Len(t, revisions, 1)
	require.NotNil(t, revisions[0].Pods, "revision.pods independent of events")
	assert.Empty(t, *revisions[0].Pods)
	assert.Nil(t, revisions[0].Events, "revision.events omitted when events call failed")
}

func TestGetWorkloadHistoryRun_UnknownWorkloadIDHardFailsNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}, "next": nil})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "Insufficient permissions"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "CP's 403 masking surfaces as not_found")
	assert.NotErrorIs(t, err, runaimcp.ErrForbidden, "must not also categorize as forbidden")
	assert.NotContains(t, err.Error(), "undefined response type", "generated-client placeholder must not leak")
	assert.NotContains(t, err.Error(), "forbidden", "forbidden prefix must not appear when masked as not_found")
}

func TestGetWorkloadHistoryRun_PrimaryNotFoundMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadHistoryRun_PrimaryAuthMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func TestGetWorkloadHistoryRun_PrimaryForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no permission"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetWorkloadHistoryRun_InferenceZeroRevisionsRendersEmptySlice(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})

	out, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.NotNil(t, out.Revisions, "revisions present (non-nil pointer) when lookup succeeded with zero revisions")
	assert.Empty(t, *out.Revisions, "renders as [] rather than collapsing into the 'lookup failed' omission")
}

func TestGetWorkloadHistoryRun_FallbackForeignRevisionHardFails(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validRevisionA, "name": "rev-a-foreign", "type": "InferenceRevision",
			"workloadId": validForeignWorkloadUUID,
			"tenantId":   1, "clusterId": "c-1", "projectId": "11", "departmentId": "22",
			"createdAt": "2026-05-17T10:00:00Z", "phase": "Stopped",
		})
	})

	podsCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA+"/pods", func(w http.ResponseWriter, _ *http.Request) {
		podsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "foreign revisionId must surface as not_found")
	assert.NotContains(t, err.Error(), validForeignWorkloadUUID, "must not leak the foreign workload UUID — generic not_found only")
	assert.False(t, podsCalled.Load(), "pods/events not fetched once foreign revision detected")
}

func TestGetWorkloadHistoryRun_UnknownRevisionIDHardFails(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "revision not found"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "unknown revisionId surfaces as not_found")
}

func TestGetWorkloadHistoryRun_RevisionForbiddenMaskedAsNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "Insufficient permissions"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "CP's 403 masking surfaces as not_found")
	assert.NotErrorIs(t, err, runaimcp.ErrForbidden, "must not also categorize as forbidden")
	assert.NotContains(t, err.Error(), "undefined response type", "generated-client placeholder must not leak")
	assert.NotContains(t, err.Error(), "forbidden", "forbidden prefix must not appear when masked as not_found")
}

func TestGetWorkloadHistoryRun_RevisionIDOnNonInferenceHardFails(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	revisionsCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		revisionsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		revisionsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "revision not found"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "revisionId for non-inference workload surfaces as validation")
	assert.Contains(t, err.Error(), "Training", "error names the actual workload type so the LLM can correct course")
	assert.False(t, revisionsCalled.Load(), "no upstream revisions call when type is known and non-inference")
}

func TestGetWorkloadHistoryRun_RevisionIDWithUnknownTypeFallsBackToInferenceBranch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID+"/history", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"records": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "type lookup boom"})
	})

	revisionsCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID+"/revisions", func(w http.ResponseWriter, _ *http.Request) {
		revisionsCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"revisions": []any{}})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/revisions/"+validRevisionA, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "revision not found"})
	})

	_, err := getWorkloadHistoryRun(context.Background(), client, getWorkloadHistoryInput{WorkloadID: validUUID, RevisionID: validRevisionA})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "unknown type + unresolvable revisionId surfaces as not_found from the upstream lookup")
	assert.True(t, revisionsCalled.Load(), "inference branch ran because type lookup was inconclusive")
}
