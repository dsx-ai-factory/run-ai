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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const telemetryPath = "/api/v1/workloads/telemetry"

func phaseVal(phase, count string) map[string]any {
	return map[string]any{
		"value":  count,
		"groups": []map[string]any{{"key": "Phase", "value": phase}},
	}
}

func telemetryBody(values ...map[string]any) map[string]any {
	return map[string]any{
		"type":      "WORKLOADS_COUNT",
		"timestamp": "2026-06-08T12:00:00Z",
		"values":    values,
	}
}

func TestGetWorkloadsSummaryInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadsSummaryInput
		wantErr string
	}{
		{name: "empty_is_platform_wide", in: getWorkloadsSummaryInput{}},
		{name: "cluster_only", in: getWorkloadsSummaryInput{ClusterID: validUUID}},
		{name: "org_pair", in: getWorkloadsSummaryInput{OrgType: "project", OrgName: "alpha"}},
		{name: "cluster_and_org_and_pool", in: getWorkloadsSummaryInput{ClusterID: validUUID, OrgType: "department", OrgName: "research", NodePoolName: "gpu"}},
		{name: "GPU idle platform wide", in: getWorkloadsSummaryInput{GPUIdleOnly: true}},
		{name: "bad_cluster_uuid", in: getWorkloadsSummaryInput{ClusterID: "not-a-uuid"}, wantErr: "clusterId"},
		{name: "org_type_without_name", in: getWorkloadsSummaryInput{OrgType: "project"}, wantErr: "together"},
		{name: "org_name_without_type", in: getWorkloadsSummaryInput{OrgName: "alpha"}, wantErr: "together"},
		{name: "bad_org_type", in: getWorkloadsSummaryInput{OrgType: "team", OrgName: "alpha"}, wantErr: "orgType must be"},
		{name: "comma_in_org_name", in: getWorkloadsSummaryInput{OrgType: "project", OrgName: "a,b"}, wantErr: "orgName"},
		{name: "node pool filter injection", in: getWorkloadsSummaryInput{NodePoolName: "gpu,phase==Failed"}, wantErr: "nodePoolName"},
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

func TestGetWorkloadsSummaryRun_BucketsAndByPhase(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotCluster, gotGroupBy string
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		gotCluster = r.URL.Query().Get("clusterId")
		gotGroupBy = r.URL.Query().Get("groupBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(
			phaseVal("Running", "5"),
			phaseVal("Pending", "2"),
			phaseVal("Initializing", "1"), // folds into pending
			phaseVal("Resuming", "1"),     // folds into pending
			phaseVal("Failed", "3"),
			phaseVal("Degraded", "1"), // folds into failed
			phaseVal("Completed", "4"),
			phaseVal("Stopped", "2"),
			phaseVal("Stopping", "1"), // folds into stopped
			phaseVal("Deleting", "1"), // folds into other
			phaseVal("Unknown", "1"),  // folds into other
		))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{ClusterID: validUUID})
	require.NoError(t, err)

	assert.Equal(t, validUUID, gotCluster, "clusterId forwarded")
	assert.Equal(t, "Phase", gotGroupBy, "grouped by Phase (capitalized server-side)")
	assert.Equal(t, validUUID, out.Scope.ClusterID)
	assert.Equal(t, "2026-06-08T12:00:00Z", out.AsOf)

	assert.Equal(t, 5, out.Counts.Running)
	assert.Equal(t, 4, out.Counts.Pending, "Pending+Initializing+Resuming")
	assert.Equal(t, 4, out.Counts.Failed, "Failed+Degraded")
	assert.Equal(t, 4, out.Counts.Completed)
	assert.Equal(t, 3, out.Counts.Stopped, "Stopped+Stopping")
	assert.Equal(t, 2, out.Counts.Other, "Deleting+Unknown")
	assert.Equal(t, 22, out.Total, "every phase summed")

	c := out.Counts
	assert.Equal(t, out.Total, c.Running+c.Pending+c.Failed+c.Completed+c.Stopped+c.Other, "buckets reconcile to total")

	assert.Equal(t, 5, out.ByPhase["Running"])
	assert.Equal(t, 1, out.ByPhase["Degraded"])
	assert.NotContains(t, out.ByPhase, "Terminating", "zero phases omitted")
}

func TestGetWorkloadsSummaryRun_PlatformWide(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.Query().Get("clusterId"), "no scope params when platform-wide")
		assert.Empty(t, r.URL.Query().Get("projectId"))
		assert.Empty(t, r.URL.Query().Get("departmentId"))
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(phaseVal("Running", "7")))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{})
	require.NoError(t, err)
	assert.Empty(t, out.Scope.ClusterID)
	assert.Empty(t, out.Scope.OrgUnitID)
	assert.Equal(t, 7, out.Counts.Running)
	assert.Equal(t, 7, out.Total)
}

func TestGetWorkloadsSummaryRun_GPUIdleCountUsesPreciseCountEndpoint(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("precise"))
		filters := strings.Split(r.URL.Query().Get("filterBy"), ",")
		assert.ElementsMatch(t, []string{
			"phase==Running",
			"idleAllocatedGpus!=0",
			"clusterId==" + validUUID,
			"nodepool==gpu-a",
		}, filters)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{
		ClusterID: validUUID, NodePoolName: "gpu-a", GPUIdleOnly: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Scope.GPUIdleOnly)
	assert.Equal(t, validUUID, out.Scope.ClusterID)
	assert.Equal(t, "gpu-a", out.Scope.NodePool)
	assert.NotEmpty(t, out.AsOf)
	assert.Equal(t, 3, out.Total)
	assert.Equal(t, 3, out.Counts.Running)
	assert.Equal(t, map[string]int{"Running": 3}, out.ByPhase)
}

func TestGetWorkloadsSummaryRun_GPUIdleZeroCountOmitsZeroPhase(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 0})
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{GPUIdleOnly: true})
	require.NoError(t, err)
	assert.Zero(t, out.Total)
	assert.Zero(t, out.Counts.Running)
	assert.Empty(t, out.ByPhase, "zero phases are omitted consistently with the normal summary path")
}

func TestGetWorkloadsSummaryRun_GPUIdleCountForOrgUnit(t *testing.T) {
	tests := []struct {
		name         string
		orgType      string
		resolvePath  string
		responseKey  string
		orgID        string
		expectedExpr string
	}{
		{name: "project", orgType: "project", resolvePath: "/api/v1/org-unit/projects", responseKey: "projects", orgID: "proj-123", expectedExpr: "projectId==proj-123"},
		{name: "department", orgType: "department", resolvePath: "/api/v1/org-unit/departments", responseKey: "departments", orgID: "dept-123", expectedExpr: "departmentId==dept-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(tt.resolvePath, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					tt.responseKey: []map[string]any{{"id": tt.orgID, "clusterId": validUUID}},
				})
			})
			mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
				filters := strings.Split(r.URL.Query().Get("filterBy"), ",")
				assert.Contains(t, filters, tt.expectedExpr)
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 2})
			})

			out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{
				OrgType: tt.orgType, OrgName: "research", GPUIdleOnly: true,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.orgID, out.Scope.OrgUnitID)
			assert.Equal(t, 2, out.Total)
		})
	}
}

func TestGetWorkloadsSummaryRun_GPUIdleCountMapsUpstreamError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{GPUIdleOnly: true})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetWorkloadsSummaryRun_GPUIdleCountRejectsNegativeUpstreamCount(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": -1})
	})

	_, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{GPUIdleOnly: true})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestGetWorkloadsSummaryRun_ProjectScopeResolvesName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotProjectID string
	mux.HandleFunc("/api/v1/org-unit/projects", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{{"id": "proj-123", "clusterId": validUUID}},
		})
	})
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		gotProjectID = r.URL.Query().Get("projectId")
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(phaseVal("Running", "2")))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{OrgType: "project", OrgName: "alpha"})
	require.NoError(t, err)
	assert.Equal(t, "proj-123", gotProjectID, "resolved project id forwarded to telemetry")
	assert.Equal(t, "project", out.Scope.OrgType)
	assert.Equal(t, "alpha", out.Scope.OrgName)
	assert.Equal(t, "proj-123", out.Scope.OrgUnitID)
	assert.Equal(t, 2, out.Counts.Running)
}

func TestGetWorkloadsSummaryRun_DepartmentScopeUsesDepartmentId(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotDeptID string
	mux.HandleFunc("/api/v1/org-unit/departments", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{{"id": "dept-9", "clusterId": validUUID}},
		})
	})
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		gotDeptID = r.URL.Query().Get("departmentId")
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(phaseVal("Failed", "1")))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{OrgType: "department", OrgName: "research"})
	require.NoError(t, err)
	assert.Equal(t, "dept-9", gotDeptID, "department id forwarded to telemetry")
	assert.Equal(t, "dept-9", out.Scope.OrgUnitID)
	assert.Equal(t, 1, out.Counts.Failed)
}

func TestGetWorkloadsSummaryRun_SkipsBadValues(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(
			phaseVal("Running", "4"),
			phaseVal("Pending", "notanumber"),                          // unparseable → dropped
			phaseVal("Failed", "-3"),                                   // negative → dropped
			map[string]any{"value": "9", "groups": []map[string]any{}}, // no phase group → dropped
			phaseVal("WeirdNewPhase", "2"),                             // unknown phase → other bucket, kept in byPhase
		))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{ClusterID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, 4, out.Counts.Running)
	assert.Equal(t, 0, out.Counts.Pending, "unparseable dropped")
	assert.Equal(t, 0, out.Counts.Failed, "negative dropped")
	assert.Equal(t, 2, out.Counts.Other, "unknown phase folds to other")
	assert.Equal(t, 6, out.Total, "only valid values counted")
	assert.Equal(t, 2, out.ByPhase["WeirdNewPhase"])
}

func TestGetWorkloadsSummaryRun_AmbiguousOrgUnit(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				{"id": "p-1", "clusterId": "c-1"},
				{"id": "p-2", "clusterId": "c-2"},
			},
		})
	})

	_, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{OrgType: "project", OrgName: "dup"})
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestGetWorkloadsSummaryRun_OrgUnitNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/org-unit/projects", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"projects": []map[string]any{}})
	})

	_, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{OrgType: "project", OrgName: "ghost"})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadsSummaryRun_NodePoolForwarded(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotNodePool string
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		gotNodePool = r.URL.Query().Get("nodepoolName")
		runaiclienttest.WriteJSON(w, http.StatusOK, telemetryBody(phaseVal("Running", "1")))
	})

	out, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{
		ClusterID: validUUID, NodePoolName: "gpu-a",
	})
	require.NoError(t, err)
	assert.Equal(t, "gpu-a", gotNodePool, "nodepoolName forwarded to telemetry")
	assert.Equal(t, "gpu-a", out.Scope.NodePool)
}

func TestGetWorkloadsSummaryRun_ForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getWorkloadsSummaryRun(context.Background(), client, getWorkloadsSummaryInput{ClusterID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}
