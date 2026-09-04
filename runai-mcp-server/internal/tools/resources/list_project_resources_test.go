/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	orgunit "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	cellsPath       = "/api/v1/org-unit/projects/node-pool-resources"
	departmentsPath = "/api/v1/org-unit/departments"
)

func TestListProjectResourcesInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      listProjectResourcesInput
		wantErr string
	}{
		{name: "happy_project_id", in: listProjectResourcesInput{ProjectID: "proj-1"}},
		{name: "happy_project_name", in: listProjectResourcesInput{ProjectName: "team-alpha"}},
		{name: "happy_department_id", in: listProjectResourcesInput{DepartmentID: "dept-1"}},
		{name: "happy_department_name", in: listProjectResourcesInput{DepartmentName: "engineering"}},
		{name: "happy_node_pool_id", in: listProjectResourcesInput{NodePoolID: "np-1"}},
		{name: "happy_node_pool_name", in: listProjectResourcesInput{NodePoolName: "h100"}},
		{name: "happy_combined", in: listProjectResourcesInput{ProjectID: "proj-1", NodePoolName: "h100"}},
		{
			name:    "reject_project_id_and_name_together",
			in:      listProjectResourcesInput{ProjectID: "proj-1", ProjectName: "team-alpha"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "reject_department_id_and_name_together",
			in:      listProjectResourcesInput{DepartmentID: "dept-1", DepartmentName: "engineering"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "reject_node_pool_id_and_name_together",
			in:      listProjectResourcesInput{NodePoolID: "np-1", NodePoolName: "h100"},
			wantErr: "mutually exclusive",
		},
		{name: "happy_no_filter_tenant_wide", in: listProjectResourcesInput{}},
		{name: "happy_cluster_only", in: listProjectResourcesInput{ClusterID: "c-1"}},
		{
			name:    "reject_negative_limit",
			in:      listProjectResourcesInput{ProjectID: "proj-1", Limit: -1},
			wantErr: "limit must be between",
		},
		{
			name:    "reject_limit_over_max",
			in:      listProjectResourcesInput{ProjectID: "proj-1", Limit: 201},
			wantErr: "limit must be between",
		},
		{
			name:    "reject_negative_offset",
			in:      listProjectResourcesInput{ProjectID: "proj-1", Offset: -1},
			wantErr: "offset must be non-negative",
		},
		{
			name:    "reject_comma_in_name",
			in:      listProjectResourcesInput{ProjectName: "team-a,team-b"},
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

func TestListProjectResourcesRun_HappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var capturedFilters []string
	var capturedLimit, capturedOffset string
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		capturedFilters = parseFilterBy(q)
		capturedLimit = q.Get("limit")
		capturedOffset = q.Get("offset")
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-alpha", "dept-1", "np-1", "h100",
				&cellResources{
					GPUDeserved:        f(8),
					GPULimit:           f(16),
					GPUOverQuotaWeight: f(2),
					CPUMilliDeserved:   f(2000),
					CPUMilliLimit:      f(4000),
					CPUOQW:             f(1),
					MemoryDeserved:     f(16384),
					MemoryLimit:        f(32768),
					MemoryOQW:          f(1),
					MemoryUnits:        orgunit.UNITS_MIB,
				},
				&cellAllocated{GPU: f(6), CPUMilli: f(1500), MemoryMB: f(8192)}),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ProjectID: "proj-1",
	})
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"orgUnitId==proj-1"}, capturedFilters)
	// Upstream is paged at a fixed internal size; user-facing limit/offset are
	// applied locally over the scope-filtered set.
	assert.Equal(t, "200", capturedLimit)
	assert.Equal(t, "0", capturedOffset)

	require.Len(t, out.Resources, 1)
	c := out.Resources[0]
	assert.Equal(t, "proj-1", c.ProjectID)
	assert.Equal(t, "team-alpha", c.ProjectName)
	require.NotNil(t, c.DepartmentID)
	assert.Equal(t, "dept-1", *c.DepartmentID)
	assert.Equal(t, "np-1", c.NodePoolID)
	assert.Equal(t, "h100", c.NodePoolName)
	assert.Equal(t, "MediumLow", c.Rank)

	// GPU
	require.NotNil(t, c.GPU.Deserved)
	assert.Equal(t, 8.0, *c.GPU.Deserved)
	require.NotNil(t, c.GPU.Limit)
	assert.Equal(t, 16.0, *c.GPU.Limit)
	require.NotNil(t, c.GPU.OverQuotaWeight)
	assert.Equal(t, 2.0, *c.GPU.OverQuotaWeight)
	require.NotNil(t, c.GPU.Allocated)
	assert.Equal(t, 6.0, *c.GPU.Allocated)

	// CPU (millicores → cores)
	require.NotNil(t, c.CPUCores.Deserved)
	assert.Equal(t, 2.0, *c.CPUCores.Deserved)
	require.NotNil(t, c.CPUCores.Limit)
	assert.Equal(t, 4.0, *c.CPUCores.Limit)
	require.NotNil(t, c.CPUCores.OverQuotaWeight)
	assert.Equal(t, 1.0, *c.CPUCores.OverQuotaWeight)
	require.NotNil(t, c.CPUCores.Allocated)
	assert.Equal(t, 1.5, *c.CPUCores.Allocated)

	// Memory (MiB → GiB for config; MB → GiB for allocated)
	require.NotNil(t, c.MemoryGiB.Deserved)
	assert.InDelta(t, 16.0, *c.MemoryGiB.Deserved, 0.001)
	require.NotNil(t, c.MemoryGiB.Limit)
	assert.InDelta(t, 32.0, *c.MemoryGiB.Limit, 0.001)
	require.NotNil(t, c.MemoryGiB.OverQuotaWeight)
	assert.Equal(t, 1.0, *c.MemoryGiB.OverQuotaWeight)
	require.NotNil(t, c.MemoryGiB.Allocated)
	assert.InDelta(t, 7.629, *c.MemoryGiB.Allocated, 0.01)

	assert.False(t, out.HasMore)
	assert.Zero(t, out.NextOffset)
}

func TestListProjectResourcesRun_AllFiltersPassThrough(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var captured []string
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		captured = parseFilterBy(r.URL.Query())
		writeCellsResponse(t, w, nil, nil)
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ClusterID:    "cluster-1",
		ProjectName:  "team-alpha",
		DepartmentID: "dept-1",
		NodePoolName: "h100",
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"clusterId==cluster-1",
		"orgUnitName==team-alpha",
		"parentId==dept-1",
		"nodePoolName==h100",
	}, captured)
}

func TestListProjectResourcesRun_DepartmentNameResolution(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var deptFilters, cellFilters []string
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		deptFilters = parseFilterBy(r.URL.Query())
		writeDepartmentsResponse(t, w, []orgunit.DepartmentForList{
			{Id: "dept-eng-uuid", Name: "engineering", ClusterId: "cluster-1"},
		})
	})
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		cellFilters = parseFilterBy(r.URL.Query())
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-alpha", "dept-eng-uuid", "np-1", "h100",
				&cellResources{GPUDeserved: f(4)}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ClusterID:      "cluster-1",
		DepartmentName: "engineering",
	})
	require.NoError(t, err)

	// Departments call was invoked and scoped to cluster + name
	assert.Contains(t, deptFilters, "name==engineering")
	assert.Contains(t, deptFilters, "clusterId==cluster-1")

	// Cells call used the resolved ID
	assert.Contains(t, cellFilters, "parentId==dept-eng-uuid")
	assert.Contains(t, cellFilters, "clusterId==cluster-1")

	require.Len(t, out.Resources, 1)
	assert.Equal(t, "proj-1", out.Resources[0].ProjectID)
}

func TestListProjectResourcesRun_DepartmentNameNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		writeDepartmentsResponse(t, w, []orgunit.DepartmentForList{})
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		DepartmentName: "ghost",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestListProjectResourcesRun_DepartmentNameAmbiguous(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		writeDepartmentsResponse(t, w, []orgunit.DepartmentForList{
			{Id: "dept-a", Name: "engineering", ClusterId: "c-1"},
			{Id: "dept-b", Name: "engineering", ClusterId: "c-2"},
		})
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		DepartmentName: "engineering",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_org_unit")
	assert.Contains(t, err.Error(), "dept-a")
	assert.Contains(t, err.Error(), "dept-b")
}

func TestListProjectResourcesRun_DepartmentLookupSkippedWhenNameAbsent(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var deptCalled bool
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		deptCalled = true
		writeDepartmentsResponse(t, w, nil)
	})
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, nil, nil)
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ProjectID: "proj-1",
	})
	require.NoError(t, err)
	assert.False(t, deptCalled, "departments lookup must not fire when departmentName is empty")
}

func TestListProjectResourcesRun_EmptyCells(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, nil, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		NodePoolID: "np-empty",
	})
	require.NoError(t, err)
	require.NotNil(t, out.Resources)
	assert.Empty(t, out.Resources)
}

func TestListProjectResourcesRun_Pagination(t *testing.T) {
	// Pagination is local over the scope-filtered set: 3 assigned cells, limit 2.
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("p1", "team-a", "dept-1", "np-1", "default", &cellResources{GPUDeserved: f(4)}, nil),
			fullCell("p2", "team-b", "dept-1", "np-1", "default", &cellResources{GPUDeserved: f(4)}, nil),
			fullCell("p3", "team-c", "dept-1", "np-1", "default", &cellResources{GPUDeserved: f(4)}, nil),
		}, nil)
	})

	// Page 1: offset 0, limit 2 → 2 cells, more remain.
	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		NodePoolName: "default",
		Limit:        2,
	})
	require.NoError(t, err)
	require.Len(t, out.Resources, 2)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(2), out.NextOffset)

	// Page 2: offset 2, limit 2 → final cell, no more.
	out2, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		NodePoolName: "default",
		Limit:        2,
		Offset:       2,
	})
	require.NoError(t, err)
	require.Len(t, out2.Resources, 1)
	assert.False(t, out2.HasMore)
	assert.Zero(t, out2.NextOffset)
}

func TestListProjectResourcesRun_AmbiguousProjectName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-alpha", "dept-1", "np-1", "h100",
				&cellResources{GPUDeserved: f(4)}, nil),
			fullCell("proj-2", "team-alpha", "dept-2", "np-1", "h100",
				&cellResources{GPUDeserved: f(2)}, nil),
		}, nil)
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ProjectName: "team-alpha",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_project")
}

func TestListProjectResourcesRun_AmbiguousNodePoolName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-alpha", "dept-1", "np-1", "h100",
				&cellResources{GPUDeserved: f(4)}, nil),
			fullCell("proj-1", "team-alpha", "dept-1", "np-2", "h100",
				&cellResources{GPUDeserved: f(4)}, nil),
		}, nil)
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		NodePoolName: "h100",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_node_pool")
}

func TestListProjectResourcesRun_SameProjectAcrossNodePoolsNotAmbiguous(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-alpha", "dept-1", "np-1", "h100",
				&cellResources{GPUDeserved: f(4)}, nil),
			fullCell("proj-1", "team-alpha", "dept-1", "np-2", "a100",
				&cellResources{GPUDeserved: f(2)}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ProjectName: "team-alpha",
	})
	require.NoError(t, err)
	assert.Len(t, out.Resources, 2)
}

func TestListProjectResourcesRun_HTTPErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantTarget error
	}{
		{name: "not_found_404", status: http.StatusNotFound, wantTarget: runaimcp.ErrNotFound},
		{name: "forbidden_403", status: http.StatusForbidden, wantTarget: runaimcp.ErrForbidden},
		{name: "auth_401", status: http.StatusUnauthorized, wantTarget: runaimcp.ErrAuth},
		{name: "throttled_429", status: http.StatusTooManyRequests, wantTarget: runaimcp.ErrUpstreamThrottled},
		{name: "upstream_500", status: http.StatusInternalServerError, wantTarget: runaimcp.ErrUpstream},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":"boom"}`))
			})

			_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
				ProjectID: "proj-1",
			})
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantTarget)
		})
	}
}

func TestListProjectResourcesRun_NullablesAndAbsentBlocks(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		// gpu has only deserved; cpu/memory blocks and gpu limit/oqw absent
		body := `{"nodePoolResources":[{
			"id":"r-1","orgUnitId":"proj-1","orgUnitName":"team-alpha",
			"parentId":"","clusterId":"c-1",
			"nodePoolId":"np-1","nodePoolName":"h100","rank":"MediumLow",
			"resources":{"gpu":{"deserved":2}}
		}]}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})

	// gpu.deserved=2 so the cell passes the assigned-only filter; cpu/memory
	// blocks and gpu limit/oqw/allocated are absent → must surface as null.
	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{
		ProjectID: "proj-1",
	})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	c := out.Resources[0]
	assert.Nil(t, c.DepartmentID, "empty parentId → null department")
	assert.Equal(t, "MediumLow", c.Rank)

	require.NotNil(t, c.GPU.Deserved)
	assert.Equal(t, 2.0, *c.GPU.Deserved)
	assert.Nil(t, c.GPU.Limit)
	assert.Nil(t, c.GPU.OverQuotaWeight)
	assert.Nil(t, c.GPU.Allocated)
	assert.Nil(t, c.CPUCores.Deserved)
	assert.Nil(t, c.CPUCores.Allocated)
	assert.Nil(t, c.MemoryGiB.Deserved)
	assert.Nil(t, c.MemoryGiB.Allocated)
}

func TestListProjectResourcesRun_FiltersUnassignedCells(t *testing.T) {
	// One assigned cell (gpu deserved 4) + one unassigned phantom cell (gpu
	// deserved 0, sentinels). Only the assigned cell is returned.
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-a", "dept-1", "np-1", "default",
				&cellResources{GPUDeserved: f(4), GPULimit: f(8), GPUOverQuotaWeight: f(2)}, nil),
			fullCell("proj-1", "team-a", "dept-1", "np-2", "np-2",
				&cellResources{GPUDeserved: f(0), GPULimit: f(-1), GPUOverQuotaWeight: f(0)}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{ProjectName: "team-a"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	assert.Equal(t, "default", out.Resources[0].NodePoolName)
	require.NotNil(t, out.Resources[0].GPU.Deserved)
	assert.Equal(t, 4.0, *out.Resources[0].GPU.Deserved)
}

// A CPU-only quota is a real assignment; it must not be dropped just because
// GPU quota is unset.
func TestListProjectResourcesRun_KeepsCPUOnlyQuotaCell(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-a", "dept-1", "np-1", "default",
				&cellResources{CPUMilliDeserved: f(4000)}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{ProjectName: "team-a"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	assert.Nil(t, out.Resources[0].GPU.Deserved)
	require.NotNil(t, out.Resources[0].CPUCores.Deserved)
	assert.Equal(t, 4.0, *out.Resources[0].CPUCores.Deserved)
}

// A project can hold live allocation with no configured quota at all (e.g.
// running over-quota). Dropping those cells made "who's using my GPUs?" return
// zero rows.
func TestListProjectResourcesRun_KeepsAllocationOnlyCell(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-a", "dept-1", "np-1", "default",
				nil, &cellAllocated{GPU: f(3)}),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{ProjectName: "team-a"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	assert.Nil(t, out.Resources[0].GPU.Deserved)
	require.NotNil(t, out.Resources[0].GPU.Allocated)
	assert.Equal(t, 3.0, *out.Resources[0].GPU.Allocated)
}

func TestListProjectResourcesRun_CapHitErrors(t *testing.T) {
	// Upstream always reports more pages (Next keeps advancing). Once the
	// internal cap is reached with data still pending, the tool must error
	// rather than silently return a partial set as complete.
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		offset := int32(0)
		if v := r.URL.Query().Get("offset"); v != "" {
			n, _ := strconv.Atoi(v)
			offset = int32(n)
		}
		cells := make([]orgunit.ProjectNodePoolResources, 200)
		for i := range cells {
			cells[i] = fullCell("p-1", "team-a", "dept-1", "np-1", "default",
				&cellResources{GPUDeserved: f(1)}, nil)
		}
		next := offset + 200
		writeCellsResponse(t, w, cells, &next)
	})

	_, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{DepartmentID: "dept-1"})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "maximum supported size")
}

func TestListProjectResourcesRun_SentinelNormalization(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-a", "dept-1", "np-1", "default",
				&cellResources{
					GPUDeserved: f(2), GPULimit: f(-1), GPUOverQuotaWeight: f(0),
					CPUMilliDeserved: f(-1), CPUOQW: f(-1),
				}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{ProjectName: "team-a"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	c := out.Resources[0]
	require.NotNil(t, c.GPU.Deserved)
	assert.Equal(t, 2.0, *c.GPU.Deserved)
	assert.Nil(t, c.GPU.Limit, "limit -1 → null (unlimited)")
	assert.Nil(t, c.GPU.OverQuotaWeight, "overQuotaWeight 0 → null")
	assert.Nil(t, c.CPUCores.Deserved, "negative cpu deserved → null")
	assert.Nil(t, c.CPUCores.OverQuotaWeight, "negative cpu oqw → null")
}

func TestListProjectResourcesRun_Rounding(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		// 8192 MB allocated → 7.62939... GiB → expect 7.629
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("proj-1", "team-a", "dept-1", "np-1", "default",
				&cellResources{GPUDeserved: f(2)},
				&cellAllocated{GPU: f(1), MemoryMB: f(8192)}),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{ProjectName: "team-a"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 1)
	require.NotNil(t, out.Resources[0].MemoryGiB.Allocated)
	assert.Equal(t, 7.629, *out.Resources[0].MemoryGiB.Allocated, "rounded to 3 decimals")
}

func TestListProjectResourcesRun_DeterministicSort(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cellsPath, func(w http.ResponseWriter, r *http.Request) {
		// upstream returns out of order
		writeCellsResponse(t, w, []orgunit.ProjectNodePoolResources{
			fullCell("p2", "team-c", "dept-1", "np-2", "np-2", &cellResources{GPUDeserved: f(1)}, nil),
			fullCell("p1", "team-a", "dept-1", "np-2", "zeta", &cellResources{GPUDeserved: f(1)}, nil),
			fullCell("p1", "team-a", "dept-1", "np-1", "alpha", &cellResources{GPUDeserved: f(1)}, nil),
		}, nil)
	})

	out, err := listProjectResourcesRun(context.Background(), client, listProjectResourcesInput{DepartmentID: "dept-1"})
	require.NoError(t, err)
	require.Len(t, out.Resources, 3)
	assert.Equal(t, "team-a", out.Resources[0].ProjectName)
	assert.Equal(t, "alpha", out.Resources[0].NodePoolName)
	assert.Equal(t, "team-a", out.Resources[1].ProjectName)
	assert.Equal(t, "zeta", out.Resources[1].NodePoolName)
	assert.Equal(t, "team-c", out.Resources[2].ProjectName)
}

// ---- helpers ----

type cellResources struct {
	GPUDeserved        *float64
	GPULimit           *float64
	GPUOverQuotaWeight *float64
	CPUMilliDeserved   *float64
	CPUMilliLimit      *float64
	CPUOQW             *float64
	MemoryDeserved     *float64
	MemoryLimit        *float64
	MemoryOQW          *float64
	MemoryUnits        orgunit.Units
}

type cellAllocated struct {
	GPU      *float64
	CPUMilli *float64
	MemoryMB *float64
}

func fullCell(projectID, projectName, parentID, npID, npName string,
	res *cellResources, alloc *cellAllocated,
) orgunit.ProjectNodePoolResources {
	c := orgunit.ProjectNodePoolResources{
		Id:           "r-" + projectID + "-" + npID,
		OrgUnitId:    projectID,
		OrgUnitName:  projectName,
		ParentId:     parentID,
		ClusterId:    "c-1",
		NodePoolId:   npID,
		NodePoolName: npName,
		Rank:         "MediumLow",
		Resources:    orgunit.ResourcesPerNodePool{},
	}
	if res != nil {
		if res.GPUDeserved != nil || res.GPULimit != nil || res.GPUOverQuotaWeight != nil {
			gpu := &orgunit.ResourcesPerNodePoolGpu{}
			if res.GPUDeserved != nil {
				gpu.Deserved = *orgunit.NewNullableFloat64(res.GPUDeserved)
			}
			if res.GPULimit != nil {
				gpu.Limit = *orgunit.NewNullableFloat64(res.GPULimit)
			}
			if res.GPUOverQuotaWeight != nil {
				gpu.OverQuotaWeight = *orgunit.NewNullableFloat64(res.GPUOverQuotaWeight)
			}
			c.Resources.Gpu = gpu
		}
		if res.CPUMilliDeserved != nil || res.CPUMilliLimit != nil || res.CPUOQW != nil {
			cpu := orgunit.ResourcesCpu{}
			if res.CPUMilliDeserved != nil {
				cpu.Deserved = *orgunit.NewNullableFloat64(res.CPUMilliDeserved)
			}
			if res.CPUMilliLimit != nil {
				cpu.Limit = *orgunit.NewNullableFloat64(res.CPUMilliLimit)
			}
			if res.CPUOQW != nil {
				cpu.OverQuotaWeight = *orgunit.NewNullableFloat64(res.CPUOQW)
			}
			c.Resources.Cpu = *orgunit.NewNullableResourcesCpu(&cpu)
		}
		if res.MemoryDeserved != nil || res.MemoryLimit != nil || res.MemoryOQW != nil {
			units := res.MemoryUnits
			if units == "" {
				units = orgunit.UNITS_MIB
			}
			mem := orgunit.ResourcesMemory{Units: units}
			if res.MemoryDeserved != nil {
				mem.Deserved = *orgunit.NewNullableFloat64(res.MemoryDeserved)
			}
			if res.MemoryLimit != nil {
				mem.Limit = *orgunit.NewNullableFloat64(res.MemoryLimit)
			}
			if res.MemoryOQW != nil {
				mem.OverQuotaWeight = *orgunit.NewNullableFloat64(res.MemoryOQW)
			}
			c.Resources.Memory = *orgunit.NewNullableResourcesMemory(&mem)
		}
	}
	if alloc != nil {
		qs := &orgunit.QuotaStatus{
			Allocated: &orgunit.QuotaStatusResource{
				Gpu:    alloc.GPU,
				Cpu:    alloc.CPUMilli,
				Memory: alloc.MemoryMB,
			},
		}
		c.QuotaStatus = qs
	}
	return c
}

func f(v float64) *float64 { return &v }

func writeCellsResponse(t *testing.T, w http.ResponseWriter, cells []orgunit.ProjectNodePoolResources, next *int32) {
	t.Helper()
	if cells == nil {
		cells = []orgunit.ProjectNodePoolResources{}
	}
	resp := orgunit.GetProjectNodePoolResources200Response{
		NodePoolResources: cells,
		Next:              next,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func writeDepartmentsResponse(t *testing.T, w http.ResponseWriter, depts []orgunit.DepartmentForList) {
	t.Helper()
	if depts == nil {
		depts = []orgunit.DepartmentForList{}
	}
	resp := orgunit.GetDepartments200Response{Departments: depts}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("encode departments response: %v", err)
	}
}

func parseFilterBy(q url.Values) []string {
	raw := q.Get("filterBy")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	sort.Strings(parts)
	return parts
}
