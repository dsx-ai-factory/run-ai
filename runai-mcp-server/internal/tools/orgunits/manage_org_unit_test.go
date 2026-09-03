/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	mouClusterUUID  = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	permScopesPath  = "/api/v1/authorization/permitted-scopes"
	nodePoolsPath   = "/api/v1/node-pools"
	projectsPath    = "/api/v1/org-unit/projects"
	projectPath     = "/api/v1/org-unit/projects/proj-1"
	departmentsPath = "/api/v1/org-unit/departments"
	departmentPath  = "/api/v1/org-unit/departments/dept-1"
)

func permScopesHandler(grant ...string) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, a := range grant {
		allowed[a] = true
	}
	scope := func(action string) map[string]any {
		if allowed[action] {
			return map[string]any{"clusters": []string{mouClusterUUID}}
		}
		return map[string]any{"clusters": []string{}}
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"create": scope("create"), "read": scope("read"),
			"update": scope("update"), "delete": scope("delete"),
		})
	}
}

func nodePoolResolveHandler(t *testing.T, name, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{{"id": id, "name": name, "clusterId": mouClusterUUID}},
		})
	}
}

func fptr(v float64) *float64 { return &v }
func sptr(v string) *string   { return &v }
func i32(v int32) *int32      { return &v }

func TestManageOrgUnitInput_Validate(t *testing.T) {
	gpuQuota := []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(2)}}
	cases := []struct {
		name    string
		in      manageOrgUnitInput
		wantErr string
	}{
		{name: "create_project_happy", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: gpuQuota}}},
		{name: "update_happy", in: manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1", Spec: &orgUnitSpec{}}},
		{name: "delete_happy", in: manageOrgUnitInput{Action: "delete", OrgType: "department", Name: "d"}},
		{name: "bad_action", in: manageOrgUnitInput{Action: "frob", OrgType: "project"}, wantErr: "action must be one of"},
		{name: "bad_orgtype", in: manageOrgUnitInput{Action: "create", OrgType: "team"}, wantErr: "orgType must be one of"},
		{name: "id_and_name", in: manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: "x", Name: "y"}, wantErr: "mutually exclusive"},
		{name: "create_no_name", in: manageOrgUnitInput{Action: "create", OrgType: "project", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{}}, wantErr: "name is required"},
		{name: "create_no_cluster", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", Spec: &orgUnitSpec{}}, wantErr: "clusterId is required"},
		{name: "create_no_spec", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID}, wantErr: "spec is required"},
		{name: "update_no_target", in: manageOrgUnitInput{Action: "update", OrgType: "project", Spec: &orgUnitSpec{}}, wantErr: "one of orgUnitId or name"},
		{name: "dept_rejects_parent", in: manageOrgUnitInput{Action: "create", OrgType: "department", Name: "d", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{ParentID: sptr("x")}}, wantErr: "parentId is only valid for orgType=project"},
		{name: "dept_rejects_enforce", in: manageOrgUnitInput{Action: "create", OrgType: "department", Name: "d", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{EnforceRunaiScheduler: new(bool)}}, wantErr: "enforceRunaiScheduler is only valid"},
		{name: "update_rejects_parent", in: manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1", Spec: &orgUnitSpec{ParentID: sptr("x")}}, wantErr: "parentId cannot be changed"},
		{name: "quota_no_nodepool", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: []orgUnitQuota{{GPUDeserved: fptr(1)}}}}, wantErr: "nodePoolName is required"},
		{name: "quota_no_gpu_on_create", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default"}}}}, wantErr: "gpuDeserved is required on create"},
		{name: "negative_scheduling_rule", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: gpuQuota, SchedulingRules: &orgUnitSchedulingRules{TrainingJobTimeLimitSeconds: i32(-5)}}}, wantErr: "must not be negative"},
		{name: "delete_with_spec", in: manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: "proj-1", Spec: &orgUnitSpec{}}, wantErr: "spec must not be set for delete"},
		{name: "negative_gpu_deserved", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(-1)}}}}, wantErr: "gpuDeserved must not be negative"},
		{name: "overquota_weight_out_of_range", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(1), GPUOverQuotaWeight: fptr(50)}}}}, wantErr: "gpuOverQuotaWeight must be between 1 and 20"},
		{name: "ignore_missing_on_create", in: manageOrgUnitInput{Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID, Spec: &orgUnitSpec{}, IgnoreMissing: true}, wantErr: "ignoreMissing is only valid"},
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

func TestManageOrgUnit_CreateProjectHappy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-default"))

	var body map[string]any
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &body))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "proj-new", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "ML",
			"resources": []map[string]any{{"nodePool": map[string]any{"id": "np-default", "name": "default"}, "gpu": map[string]any{"deserved": 4.0}}},
		})
	})

	in := manageOrgUnitInput{
		Action: "create", OrgType: "project", Name: "ml-team", ClusterID: mouClusterUUID,
		Spec: &orgUnitSpec{Description: sptr("ML"), Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(4)}}},
	}
	in.Confirm = true
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "proj-new", out.OrgUnitID)
	assert.Empty(t, out.Before)
	assert.Equal(t, "ml-team", out.After["name"])

	resources := body["resources"].([]any)
	require.Len(t, resources, 1)
	r0 := resources[0].(map[string]any)
	assert.Equal(t, "np-default", r0["nodePool"].(map[string]any)["id"])
	assert.Equal(t, 4.0, r0["gpu"].(map[string]any)["deserved"])
}

// CPU cores and memory GiB at the boundary must reach upstream as millicores and MiB,
// and schedulingRules must pass through.
func TestManageOrgUnit_CreateConvertsUnitsAndSchedulingRules(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-default"))

	var body map[string]any
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &body))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "proj-new", "name": "p", "clusterId": mouClusterUUID, "parentId": "",
			"resources": []map[string]any{{"nodePool": map[string]any{"id": "np-default"}, "gpu": map[string]any{"deserved": 1.0}}},
		})
	})

	in := manageOrgUnitInput{
		Action: "create", OrgType: "project", Name: "p", ClusterID: mouClusterUUID,
		Spec: &orgUnitSpec{
			Quota:           []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(1), CPUDeserved: fptr(2), MemoryDeserved: fptr(4)}},
			SchedulingRules: &orgUnitSchedulingRules{TrainingJobTimeLimitSeconds: i32(3600)},
		},
	}
	in.Confirm = true
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)

	r0 := body["resources"].([]any)[0].(map[string]any)
	assert.Equal(t, 2000.0, r0["cpu"].(map[string]any)["deserved"], "2 cores -> 2000 millicores")
	mem := r0["memory"].(map[string]any)
	assert.Equal(t, 4096.0, mem["deserved"], "4 GiB -> 4096 MiB")
	assert.Equal(t, "Mib", mem["units"])
	assert.Equal(t, 3600.0, body["schedulingRules"].(map[string]any)["trainingJobTimeLimitSeconds"])
}

func TestManageOrgUnit_CreateProjectDryRun(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-default"))
	mux.HandleFunc(projectsPath, func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("dry-run must not POST to the projects endpoint")
	})

	in := manageOrgUnitInput{
		Action: "create", OrgType: "project", Name: "ml-team", ClusterID: mouClusterUUID, DryRun: true,
		Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(4)}}},
	}
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.True(t, out.DryRun)
	assert.Equal(t, "ml-team", out.After["name"])
}

// A description-only update must re-send the project's existing per-node-pool quota,
// because the PUT replaces the whole resources array.
func TestManageOrgUnit_UpdateCarriesQuotaForward(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))

	var putBody map[string]any
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "old",
				"enforceRunaiScheduler": true,
				"resources":             []map[string]any{{"nodePool": map[string]any{"id": "np-a", "name": "a"}, "gpu": map[string]any{"deserved": 8.0}}},
			})
		case http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(raw, &putBody))
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "new",
				"enforceRunaiScheduler": true,
				"resources":             []map[string]any{{"nodePool": map[string]any{"id": "np-a", "name": "a"}, "gpu": map[string]any{"deserved": 8.0}}},
			})
		}
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1", Spec: &orgUnitSpec{Description: sptr("new")}}
	in.Confirm = true
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, map[string]any{"before": "old", "after": "new"}, out.Diff["description"])

	resources := putBody["resources"].([]any)
	require.Len(t, resources, 1, "existing quota must be carried forward")
	r0 := resources[0].(map[string]any)
	assert.Equal(t, "np-a", r0["nodePool"].(map[string]any)["id"])
	assert.Equal(t, 8.0, r0["gpu"].(map[string]any)["deserved"])
}

// A description-only update must re-send the project's existing nodeTypes: the MCP doesn't
// expose them, and the PUT replaces the whole object, so omitting them would wipe them.
func TestManageOrgUnit_UpdateCarriesNodeTypesForward(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))

	var putBody map[string]any
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "old",
				"resources": []map[string]any{},
				"nodeTypes": map[string]any{"training": []string{"nt-1"}, "workspace": []string{"nt-2"}, "names": map[string]string{"nt-1": "gpu-a"}},
			})
		case http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(raw, &putBody))
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "new", "resources": []map[string]any{},
			})
		}
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1", Confirm: true, Spec: &orgUnitSpec{Description: sptr("new")}}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)

	nodeTypes, ok := putBody["nodeTypes"].(map[string]any)
	require.True(t, ok, "existing nodeTypes must be carried forward in the PUT")
	assert.Equal(t, []any{"nt-1"}, nodeTypes["training"])
	assert.Equal(t, []any{"nt-2"}, nodeTypes["workspace"])
}

// Same carry-forward for departments.
func TestManageOrgUnit_DepartmentUpdateCarriesNodeTypesForward(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))

	var putBody map[string]any
	mux.HandleFunc(departmentPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "dept-1", "name": "research", "clusterId": mouClusterUUID, "description": "old",
				"totalResources": map[string]any{}, "resources": []map[string]any{},
				"nodeTypes": map[string]any{"training": []string{"nt-9"}},
			})
		case http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(raw, &putBody))
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "dept-1", "name": "research", "clusterId": mouClusterUUID, "description": "new",
				"totalResources": map[string]any{}, "resources": []map[string]any{},
			})
		}
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "department", OrgUnitID: "dept-1", Confirm: true, Spec: &orgUnitSpec{Description: sptr("new")}}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)

	nodeTypes, ok := putBody["nodeTypes"].(map[string]any)
	require.True(t, ok, "existing nodeTypes must be carried forward in the PUT")
	assert.Equal(t, []any{"nt-9"}, nodeTypes["training"])
}

// A quota-only department update must re-send the existing description: the PUT replaces
// the whole object, so omitting it would wipe it.
func TestManageOrgUnit_DepartmentUpdateCarriesDescriptionForward(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-a"))

	var putBody map[string]any
	mux.HandleFunc(departmentPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "dept-1", "name": "research", "clusterId": mouClusterUUID, "description": "keep me",
				"totalResources":   map[string]any{},
				"defaultNodePools": []string{"default"},
				"resources":        []map[string]any{{"nodePool": map[string]any{"id": "np-a", "name": "default"}, "gpu": map[string]any{"deserved": 4.0}}},
			})
		case http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(raw, &putBody))
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "dept-1", "name": "research", "clusterId": mouClusterUUID, "description": "keep me",
				"totalResources": map[string]any{}, "resources": []map[string]any{},
			})
		}
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "department", OrgUnitID: "dept-1",
		Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(2)}}}}
	in.Confirm = true
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.Equal(t, "keep me", putBody["description"], "description must be carried forward on a quota-only update")
	assert.Equal(t, []any{"default"}, putBody["defaultNodePools"], "defaultNodePools must be carried forward")
}

func TestManageOrgUnit_DeleteDepartmentBlockedByChildren(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(departmentPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "dept-1", "name": "research", "clusterId": mouClusterUUID, "description": "", "totalResources": map[string]any{},
			"resources": []map[string]any{},
			"children":  []map[string]any{{"id": "proj-x", "name": "x"}},
		})
	})

	in := manageOrgUnitInput{Action: "delete", OrgType: "department", OrgUnitID: "dept-1", Confirm: true}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConflict)
	assert.Contains(t, err.Error(), "subordinate project")
	assert.Contains(t, err.Error(), "delete them first")
	assert.NotContains(t, err.Error(), "move them")
}

func TestManageOrgUnit_DeleteProjectRequiresConfirm(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "", "resources": []map[string]any{},
		})
	})

	in := manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: "proj-1"}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "workloads", "confirmation must warn about the workload cascade")
}

func TestManageOrgUnit_DeleteProjectConfirmed(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	deleted := false
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "ml-team", "clusterId": mouClusterUUID, "parentId": "", "description": "", "resources": []map[string]any{},
			})
		case http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusOK)
		}
	})

	in := manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: "proj-1", Confirm: true}
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.True(t, deleted)
	assert.Contains(t, strings.Join(out.Warnings, " "), "workloads", "result must note the workload cascade")
}

func TestManageOrgUnit_DeleteIgnoreMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "gone"})
	})

	in := manageOrgUnitInput{Action: "delete", OrgType: "project", OrgUnitID: "proj-1", IgnoreMissing: true, Confirm: true}
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.NotEmpty(t, out.Warnings)
}

func TestManageOrgUnit_ForbiddenViaProbe(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("read")) // create not granted
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-default"))

	in := manageOrgUnitInput{
		Action: "create", OrgType: "project", Name: "ml-team", ClusterID: mouClusterUUID,
		Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(1)}}},
	}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestManageOrgUnitToolsLint(t *testing.T) {
	errs := runaimcp.Lint(func(r *runaimcp.Registry) {
		registerManageOrgUnit(r, nil)
	})
	if len(errs) > 0 {
		t.Fatalf("tool lint errors: %v", errs)
	}
}

// Adding a node pool not already in the org-unit's quota during update has nothing to
// carry forward, so gpuDeserved must be supplied — buildResources enforces this (Validate
// only requires it on create).
func TestManageOrgUnit_UpdateNewNodePoolRequiresGpuDeserved(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": []map[string]any{
			{"id": "np-a", "name": "a", "clusterId": mouClusterUUID},
			{"id": "np-b", "name": "newpool", "clusterId": mouClusterUUID},
		}})
	})
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "proj-1", "name": "p", "clusterId": mouClusterUUID, "parentId": "",
			"resources": []map[string]any{{"nodePool": map[string]any{"id": "np-a", "name": "a"}, "gpu": map[string]any{"deserved": 2.0}}},
		})
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1",
		Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "newpool", GPULimit: fptr(4)}}}}
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gpuDeserved is required when adding a node pool")
}

// Changing gpuDeserved on update must not carry the prior overQuotaWeight forward (it would
// break the "weight equals deserved or empty" rule when over-quota is off). Pins applyQuota.
func TestManageOrgUnit_UpdateDropsStaleOverQuotaWeight(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath, nodePoolResolveHandler(t, "default", "np-default"))

	var putBody map[string]any
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "p", "clusterId": mouClusterUUID, "parentId": "",
				"resources": []map[string]any{{"nodePool": map[string]any{"id": "np-default", "name": "default"}, "gpu": map[string]any{"deserved": 2.0, "overQuotaWeight": 5.0}}},
			})
		case http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(raw, &putBody))
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "proj-1", "name": "p", "clusterId": mouClusterUUID, "parentId": "",
				"resources": []map[string]any{{"nodePool": map[string]any{"id": "np-default", "name": "default"}, "gpu": map[string]any{"deserved": 4.0}}},
			})
		}
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1",
		Spec: &orgUnitSpec{Quota: []orgUnitQuota{{NodePoolName: "default", GPUDeserved: fptr(4)}}}}
	in.Confirm = true
	_, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	gpu := putBody["resources"].([]any)[0].(map[string]any)["gpu"].(map[string]any)
	assert.EqualValues(t, 4, gpu["deserved"])
	_, hasWeight := gpu["overQuotaWeight"]
	assert.False(t, hasWeight, "stale overQuotaWeight must not be carried forward when deserved changes")
}

func TestManageOrgUnit_DeleteByNameIgnoreMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"projects": []map[string]any{}})
	})

	in := manageOrgUnitInput{Action: "delete", OrgType: "project", Name: "ghost", ClusterID: mouClusterUUID, IgnoreMissing: true, Confirm: true}
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.NotEmpty(t, out.Warnings)
}

func TestManageOrgUnit_UpdateDryRun(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(projectPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			t.Fatal("dry-run must not PUT to the project endpoint")
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "proj-1", "name": "p", "clusterId": mouClusterUUID, "parentId": "",
			"resources":       []map[string]any{{"nodePool": map[string]any{"id": "np-a", "name": "a"}, "gpu": map[string]any{"deserved": 2.0}}},
			"schedulingRules": map[string]any{"trainingJobTimeLimitSeconds": 3600},
		})
	})

	in := manageOrgUnitInput{Action: "update", OrgType: "project", OrgUnitID: "proj-1", DryRun: true,
		Spec: &orgUnitSpec{SchedulingRules: &orgUnitSchedulingRules{InteractiveJobMaxIdleDurationSeconds: i32(900)}}}
	out, err := manageOrgUnitRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.True(t, out.DryRun)
	sr := out.After["schedulingRules"].(map[string]any)
	assert.EqualValues(t, 900, sr["interactiveJobMaxIdleDurationSeconds"], "caller's change applied in preview")
	assert.EqualValues(t, 3600, sr["trainingJobTimeLimitSeconds"], "existing rule preserved in preview")
}
