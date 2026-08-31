/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const deptCellsPath = "/api/v1/org-unit/departments/node-pool-resources"

// deptCell reuses fullCell's Resources construction, dropping the project-only
// parent and quotaStatus that DepartmentNodePoolResources does not carry.
func deptCell(deptID, deptName, npID, npName string, res *cellResources) orgunit.DepartmentNodePoolResources {
	pc := fullCell(deptID, deptName, "", npID, npName, res, nil)
	return orgunit.DepartmentNodePoolResources{
		Id:           pc.Id,
		OrgUnitId:    pc.OrgUnitId,
		OrgUnitName:  pc.OrgUnitName,
		ClusterId:    pc.ClusterId,
		NodePoolId:   pc.NodePoolId,
		NodePoolName: pc.NodePoolName,
		Rank:         pc.Rank,
		Resources:    pc.Resources,
	}
}

func writeDeptResponse(t *testing.T, w http.ResponseWriter, cells []orgunit.DepartmentNodePoolResources, next *int32) {
	t.Helper()
	if cells == nil {
		cells = []orgunit.DepartmentNodePoolResources{}
	}
	resp := orgunit.GetDepartmentNodePoolResources200Response{NodePoolResources: cells, Next: next}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("encode dept response: %v", err)
	}
}

func TestListDepartmentResourcesInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      listDepartmentResourcesInput
		wantErr string
	}{
		{name: "happy_department_id", in: listDepartmentResourcesInput{DepartmentID: "d-1"}},
		{name: "happy_department_name", in: listDepartmentResourcesInput{DepartmentName: "engineering"}},
		{name: "happy_node_pool_id", in: listDepartmentResourcesInput{NodePoolID: "np-1"}},
		{name: "happy_combined", in: listDepartmentResourcesInput{DepartmentID: "d-1", NodePoolName: "h100"}},
		{
			name:    "reject_department_id_and_name_together",
			in:      listDepartmentResourcesInput{DepartmentID: "d-1", DepartmentName: "eng"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "reject_node_pool_id_and_name_together",
			in:      listDepartmentResourcesInput{NodePoolID: "np-1", NodePoolName: "h100"},
			wantErr: "mutually exclusive",
		},
		{name: "happy_no_filter_tenant_wide", in: listDepartmentResourcesInput{}},
		{name: "happy_cluster_only", in: listDepartmentResourcesInput{ClusterID: "c-1"}},
		{
			name:    "reject_limit_over_max",
			in:      listDepartmentResourcesInput{DepartmentID: "d-1", Limit: 201},
			wantErr: "limit must be between",
		},
		{
			name:    "reject_comma_in_name",
			in:      listDepartmentResourcesInput{DepartmentName: "a,b"},
			wantErr: "must not contain commas",
		},
	}
	for _, tc := range tests {
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

func TestListDepartmentResourcesRun_HappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var capturedFilters []string
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, r *http.Request) {
		capturedFilters = parseFilterBy(r.URL.Query())
		writeDeptResponse(t, w, []orgunit.DepartmentNodePoolResources{
			deptCell("d-1", "engineering", "np-1", "h100", &cellResources{
				GPUDeserved:        f(8),
				GPULimit:           f(16),
				GPUOverQuotaWeight: f(2),
				CPUMilliDeserved:   f(2000),
				MemoryDeserved:     f(16384),
				MemoryUnits:        orgunit.UNITS_MIB,
			}),
		}, nil)
	})

	out, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentID: "d-1"})
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"orgUnitId==d-1"}, capturedFilters)
	require.Len(t, out.Resources, 1)
	c := out.Resources[0]
	assert.Equal(t, "d-1", c.DepartmentID)
	assert.Equal(t, "engineering", c.DepartmentName)
	assert.Equal(t, "np-1", c.NodePoolID)
	assert.Equal(t, "h100", c.NodePoolName)

	require.NotNil(t, c.GPU.Deserved)
	assert.Equal(t, float64(8), *c.GPU.Deserved)
	require.NotNil(t, c.CPUCores.Deserved)
	assert.Equal(t, float64(2), *c.CPUCores.Deserved, "2000 millicores normalized to 2 cores")
	require.NotNil(t, c.MemoryGiB.Deserved)
	assert.Equal(t, float64(16), *c.MemoryGiB.Deserved, "16384 MiB normalized to 16 GiB")
	assert.False(t, out.HasMore)
}

func TestListDepartmentResourcesRun_FilterTranslation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var captured []string
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, r *http.Request) {
		captured = parseFilterBy(r.URL.Query())
		writeDeptResponse(t, w, nil, nil)
	})

	_, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{
		ClusterID:      "c-9",
		DepartmentName: "eng",
		NodePoolName:   "h100",
	})
	require.NoError(t, err)
	assert.ElementsMatch(t,
		[]string{"clusterId==c-9", "orgUnitName==eng", "nodePoolName==h100"},
		captured,
		"departmentName maps to orgUnitName; no parentId predicate (departments have no parent)",
	)
}

func TestListDepartmentResourcesRun_FiltersUnassigned(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, _ *http.Request) {
		writeDeptResponse(t, w, []orgunit.DepartmentNodePoolResources{
			deptCell("d-1", "eng", "np-1", "h100", &cellResources{GPUDeserved: f(4)}),
			deptCell("d-1", "eng", "np-2", "a100", &cellResources{GPUDeserved: f(0)}),
		}, nil)
	})

	out, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentID: "d-1"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1, "the deserved==0 cell is dropped as unassigned")
	assert.Equal(t, "np-1", out.Resources[0].NodePoolID)
}

// A CPU-only department quota is a real assignment; it must not be dropped
// just because GPU quota is unset.
func TestListDepartmentResourcesRun_KeepsCPUOnlyQuotaCell(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, _ *http.Request) {
		writeDeptResponse(t, w, []orgunit.DepartmentNodePoolResources{
			deptCell("d-1", "eng", "np-1", "h100", &cellResources{CPUMilliDeserved: f(8000)}),
		}, nil)
	})

	out, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentID: "d-1"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	assert.Nil(t, out.Resources[0].GPU.Deserved)
	require.NotNil(t, out.Resources[0].CPUCores.Deserved)
	assert.Equal(t, 8.0, *out.Resources[0].CPUCores.Deserved)
}

func TestListDepartmentResourcesRun_AmbiguousDepartment(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, _ *http.Request) {
		writeDeptResponse(t, w, []orgunit.DepartmentNodePoolResources{
			deptCell("d-1", "eng", "np-1", "h100", &cellResources{GPUDeserved: f(4)}),
			deptCell("d-2", "eng", "np-1", "h100", &cellResources{GPUDeserved: f(4)}),
		}, nil)
	})

	_, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentName: "eng"})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_department")
}

func TestListDepartmentResourcesRun_ForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentID: "d-1"})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestListDepartmentResourcesRun_UpstreamErrorMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(deptCellsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := listDepartmentResourcesRun(context.Background(), client, listDepartmentResourcesInput{DepartmentID: "d-1"})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}
