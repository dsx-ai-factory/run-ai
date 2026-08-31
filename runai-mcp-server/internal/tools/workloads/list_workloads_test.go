/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	wlclient "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func TestListWorkloadsInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   listWorkloadsInput
		wantErr string
	}{
		{name: "empty input is valid"},
		{name: "all filters", input: listWorkloadsInput{
			Name: "trainer", ProjectID: "4500002", ClusterID: validUUID, Phase: "Pending", Type: "Training",
			NodePoolName: "gpu-a", CreatedAfter: "2026-07-05T10:00:00Z", PhaseUpdatedAfter: "2026-07-06T10:00:00Z",
			DeletedOnly: true, DeletedAfter: "2026-07-07T10:00:00Z",
			Limit: 100, Offset: 200,
		}},
		{name: "invalid cluster", input: listWorkloadsInput{ClusterID: "cluster-a"}, wantErr: "clusterId must be a UUID"},
		{name: "invalid phase casing", input: listWorkloadsInput{Phase: "pending"}, wantErr: "phase must be one of"},
		{name: "GPU idle with running phase", input: listWorkloadsInput{GPUIdleOnly: true, Phase: "Running"}},
		{name: "GPU idle with conflicting phase", input: listWorkloadsInput{GPUIdleOnly: true, Phase: "Pending"}, wantErr: "phase must be Running"},
		{name: "GPU idle with deleted only", input: listWorkloadsInput{GPUIdleOnly: true, DeletedOnly: true}, wantErr: "gpuIdleOnly cannot be combined"},
		{name: "GPU idle with deleted after", input: listWorkloadsInput{GPUIdleOnly: true, DeletedAfter: "2026-07-01T00:00:00Z"}, wantErr: "gpuIdleOnly cannot be combined"},
		{name: "GPU idle with deletedAt API filter", input: listWorkloadsInput{GPUIdleOnly: true, FilterBy: []apiWorkloadFilter{{Field: "deletedAt", Operator: ">=", Value: "2026-07-01T00:00:00Z"}}}, wantErr: "gpuIdleOnly cannot be combined"},
		{name: "allocated GPU comparison", input: listWorkloadsInput{AllocatedGPU: &gpuComparison{Operator: ">", Value: 2}}},
		{name: "allocated GPU invalid operator", input: listWorkloadsInput{AllocatedGPU: &gpuComparison{Operator: "gt", Value: 2}}, wantErr: "allocatedGpu.operator"},
		{name: "allocated GPU negative value", input: listWorkloadsInput{AllocatedGPU: &gpuComparison{Operator: ">", Value: -1}}, wantErr: "allocatedGpu.value"},
		{name: "requested GPU with project scope", input: listWorkloadsInput{ProjectID: "4500002", RequestedGPU: &gpuComparison{Operator: ">=", Value: 2}}},
		{name: "requested GPU with cluster scope", input: listWorkloadsInput{ClusterID: validUUID, RequestedGPU: &gpuComparison{Operator: "<=", Value: 1}}},
		{name: "requested GPU without scope", input: listWorkloadsInput{RequestedGPU: &gpuComparison{Operator: ">", Value: 2}}, wantErr: "requires projectId or clusterId"},
		{name: "invalid created timestamp", input: listWorkloadsInput{CreatedAfter: "yesterday"}, wantErr: "createdAfter must be RFC3339"},
		{name: "invalid phase timestamp", input: listWorkloadsInput{PhaseUpdatedAfter: "recently"}, wantErr: "phaseUpdatedAfter must be RFC3339"},
		{name: "invalid deleted timestamp", input: listWorkloadsInput{DeletedAfter: "six hours ago"}, wantErr: "deletedAfter must be RFC3339"},
		{name: "blank search", input: listWorkloadsInput{Search: "   "}, wantErr: "search must not be blank"},
		{name: "invalid sort field", input: listWorkloadsInput{SortBy: "unknown"}, wantErr: "sortBy must be one of"},
		{name: "invalid sort order", input: listWorkloadsInput{SortOrder: "ascending"}, wantErr: "sortOrder must be one of"},
		{name: "valid API string filter", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "name", Operator: "=@", Value: "train"}}}},
		{name: "valid API timestamp filter", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "createdAt", Operator: ">", Value: "2026-07-01T00:00:00Z"}}}},
		{name: "invalid API filter field", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "unknown", Operator: "==", Value: "x"}}}, wantErr: "filterBy[0].field must be one of"},
		{name: "invalid API string operator", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "name", Operator: ">", Value: "x"}}}, wantErr: "filterBy[0].operator"},
		{name: "invalid API timestamp", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "deletedAt", Operator: ">=", Value: "yesterday"}}}, wantErr: "filterBy[0].value must be RFC3339"},
		{name: "invalid API integer", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "priority", Operator: ">=", Value: "1.5"}}}, wantErr: "filterBy[0].value must be a non-negative 32-bit integer"},
		{name: "invalid API number", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "idleAllocatedGpus", Operator: ">=", Value: "NaN"}}}, wantErr: "filterBy[0].value must be a finite non-negative number"},
		{name: "negative API number", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "allocatedGPU", Operator: ">=", Value: "-1"}}}, wantErr: "filterBy[0].value must be a finite non-negative number"},
		{name: "API filter injection", input: listWorkloadsInput{FilterBy: []apiWorkloadFilter{{Field: "name", Operator: "==", Value: "trainer,phase==Failed"}}}, wantErr: "filterBy[0] must not contain commas"},
		{name: "valid string comparison", input: listWorkloadsInput{ProjectName: &workloadFilterComparison{Operator: "=@", Value: "research"}}},
		{name: "valid ordered string comparison", input: listWorkloadsInput{ProjectName: &workloadFilterComparison{Operator: ">=", Value: "research"}}},
		{name: "invalid strict string operator", input: listWorkloadsInput{ProjectName: &workloadFilterComparison{Operator: ">", Value: "research"}}, wantErr: "projectName.operator"},
		{name: "invalid integer value", input: listWorkloadsInput{Priority: &workloadFilterComparison{Operator: ">=", Value: "40.5"}}, wantErr: "priority.value must be a non-negative 32-bit integer"},
		{name: "integer overflow", input: listWorkloadsInput{TotalRunningTimeSeconds: &workloadFilterComparison{Operator: ">=", Value: "2147483648"}}, wantErr: "totalRunningTimeSeconds.value must be a non-negative 32-bit integer"},
		{name: "invalid timestamp value", input: listWorkloadsInput{CompletedAt: &workloadFilterComparison{Operator: ">=", Value: "yesterday"}}, wantErr: "completedAt.value must be RFC3339"},
		{name: "valid AI application string ID", input: listWorkloadsInput{AIApplicationID: &workloadFilterComparison{Operator: "=@", Value: "app-1"}}},
		{name: "comparison filter injection", input: listWorkloadsInput{DepartmentName: &workloadFilterComparison{Operator: "==", Value: "eng,phase==Failed"}}, wantErr: "departmentName must not contain commas"},
		{name: "limit too large", input: listWorkloadsInput{Limit: 201}, wantErr: "limit must be between"},
		{name: "negative offset", input: listWorkloadsInput{Offset: -1}, wantErr: "offset must be non-negative"},
		{name: "blank name", input: listWorkloadsInput{Name: "   "}, wantErr: "name must not be blank"},
		{name: "filter injection", input: listWorkloadsInput{ProjectID: "1,phase==Failed"}, wantErr: "projectId must not contain commas"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestServerWorkloadComparisonFiltersPreserveStrictComparison(t *testing.T) {
	assert.Equal(t,
		[]string{"priority>=40", "priority!=40"},
		serverWorkloadComparisonFilters("priority", workloadFilterComparison{Operator: ">", Value: "40"}),
	)
	assert.Equal(t,
		[]string{"completedAt<=2026-07-01T00:00:00Z", "completedAt!=2026-07-01T00:00:00Z"},
		serverWorkloadComparisonFilters("completedAt", workloadFilterComparison{Operator: "<", Value: "2026-07-01T00:00:00Z"}),
	)
}

func TestListWorkloadsFilterLimitCountsExpandedStrictComparisons(t *testing.T) {
	eightFiltersAndStrict := listWorkloadsInput{
		Name: "w", ProjectID: "p", ClusterID: validUUID, Phase: "Running", Type: "Training", NodePoolName: "default",
		CreatedAfter: "2026-07-01T00:00:00Z", PhaseUpdatedAfter: "2026-07-01T00:00:00Z",
		Priority: &workloadFilterComparison{Operator: ">", Value: "40"},
	}
	require.Len(t, listWorkloadsFilters(eightFiltersAndStrict), 10)
	assert.NoError(t, eightFiltersAndStrict.Validate())

	elevenFilters := eightFiltersAndStrict
	elevenFilters.Category = &workloadFilterComparison{Operator: "==", Value: "Train"}
	require.Len(t, listWorkloadsFilters(elevenFilters), 11)
	err := elevenFilters.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum of 10")
}

func TestListWorkloadsFilterLimitCountsExpandedStructuredFilters(t *testing.T) {
	in := listWorkloadsInput{FilterBy: make([]apiWorkloadFilter, 0, 10)}
	for i := 0; i < 9; i++ {
		in.FilterBy = append(in.FilterBy, apiWorkloadFilter{
			Field: "name", Operator: "!=", Value: fmt.Sprintf("workload-%d", i),
		})
	}
	in.FilterBy = append(in.FilterBy, apiWorkloadFilter{
		Field: "allocatedGPU", Operator: ">", Value: "2",
	})
	require.Len(t, listWorkloadsFilters(in), 11)
	err := in.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum of 10")
}

func TestListWorkloadsAdditionalFilterExpressions(t *testing.T) {
	in := listWorkloadsInput{
		ProjectName:             &workloadFilterComparison{Operator: "==", Value: "p1"},
		DepartmentID:            &workloadFilterComparison{Operator: "!=", Value: "d1"},
		DepartmentName:          &workloadFilterComparison{Operator: "=@", Value: "eng"},
		SubmittedBy:             &workloadFilterComparison{Operator: "=^", Value: "tal"},
		CompletedAt:             &workloadFilterComparison{Operator: ">=", Value: "2026-07-01T00:00:00Z"},
		DistributedFramework:    &workloadFilterComparison{Operator: "==", Value: "PyTorch"},
		Category:                &workloadFilterComparison{Operator: "!@", Value: "test"},
		TotalPendingTimeSeconds: &workloadFilterComparison{Operator: ">=", Value: "60"},
		TotalRunningTimeSeconds: &workloadFilterComparison{Operator: "<=", Value: "3600"},
		Priority:                &workloadFilterComparison{Operator: "==", Value: "50"},
		PriorityClassName:       &workloadFilterComparison{Operator: "=$", Value: "high"},
		GuaranteedRuntimeEndsAt: &workloadFilterComparison{Operator: "<=", Value: "2026-07-31T00:00:00Z"},
		AIApplicationID:         &workloadFilterComparison{Operator: "!=", Value: validUUID},
		AIApplicationName:       &workloadFilterComparison{Operator: "=@", Value: "nemo"},
	}
	assert.ElementsMatch(t, []string{
		"projectName==p1", "departmentId!=d1", "departmentName=@eng", "submittedBy=^tal",
		"completedAt>=2026-07-01T00:00:00Z", "distributedFramework==PyTorch", "category!@test",
		"totalPendingTimeSeconds>=60", "totalRunningTimeSeconds<=3600", "priority==50",
		"priorityClassName=$high", "guaranteedRuntimeEndsAt<=2026-07-31T00:00:00Z",
		"aiApplicationId!=" + validUUID, "aiApplicationName=@nemo",
	}, listWorkloadsFilters(in))
}

func TestListWorkloadsRunPassesSearchAndSort(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "trainer", r.URL.Query().Get("search"))
		assert.Equal(t, "createdAt", r.URL.Query().Get("sortBy"))
		assert.Equal(t, "asc", r.URL.Query().Get("sortOrder"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	_, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		Search: "trainer", SortBy: "createdAt", SortOrder: "asc",
	})
	require.NoError(t, err)
}

func TestWorkloadSortNormalizesMCPFieldAliases(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  string
	}{
		{input: "nodePoolName", want: "nodepool"},
		{input: "allocatedGpu", want: "allocatedGPU"},
		{input: "projectName", want: "projectName"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			assert.NoError(t, (listWorkloadsInput{SortBy: tt.input}).Validate())
			got, order := workloadSort(listWorkloadsInput{SortBy: tt.input})
			assert.Equal(t, tt.want, got)
			assert.Equal(t, "desc", order)
		})
	}
}

func TestWorkloadSortSchemaOptionsMatchValidation(t *testing.T) {
	want := strings.Split(workloadSortFieldOptions, ", ")
	got := make([]string, 0, len(workloadSortFields))
	for field := range workloadSortFields {
		got = append(got, field)
	}
	assert.ElementsMatch(t, want, got)
}

func TestWorkloadFilterFieldsMatchAPI(t *testing.T) {
	want := make([]string, 0, len(wlclient.AllowedWorkloadSortFilterFieldsEnumValues))
	for _, field := range wlclient.AllowedWorkloadSortFilterFieldsEnumValues {
		want = append(want, string(field))
	}
	got := make([]string, 0, len(workloadFilterKinds))
	for field := range workloadFilterKinds {
		got = append(got, field)
	}
	assert.ElementsMatch(t, want, got)
	assert.ElementsMatch(t, strings.Split(workloadFilterFieldOptions, ", "), got)
}

func TestEveryAPIWorkloadFilterFieldValidates(t *testing.T) {
	valueByKind := map[string]string{
		"string":    "value",
		"timestamp": "2026-07-01T00:00:00Z",
		"integer":   "1",
		"number":    "1.5",
	}
	for field, kind := range workloadFilterKinds {
		t.Run(field, func(t *testing.T) {
			in := listWorkloadsInput{FilterBy: []apiWorkloadFilter{{
				Field: field, Operator: ">=", Value: valueByKind[kind],
			}}}
			assert.NoError(t, in.Validate())
		})
	}
}

func TestServerGPUFiltersPreserveComparison(t *testing.T) {
	tests := []struct {
		name       string
		comparison gpuComparison
		want       []string
	}{
		{name: "more than", comparison: gpuComparison{Operator: ">", Value: 2}, want: []string{"allocatedGPU>=2", "allocatedGPU!=2"}},
		{name: "at least", comparison: gpuComparison{Operator: ">=", Value: 2}, want: []string{"allocatedGPU>=2"}},
		{name: "less than", comparison: gpuComparison{Operator: "<", Value: 1.5}, want: []string{"allocatedGPU<=1.5", "allocatedGPU!=1.5"}},
		{name: "at most", comparison: gpuComparison{Operator: "<=", Value: 1.5}, want: []string{"allocatedGPU<=1.5"}},
		{name: "equals", comparison: gpuComparison{Operator: "==", Value: 4}, want: []string{"allocatedGPU==4"}},
		{name: "not zero excludes null", comparison: gpuComparison{Operator: "!=", Value: 0}, want: []string{"allocatedGPU>=0", "allocatedGPU!=0"}},
		{name: "not nonzero excludes null", comparison: gpuComparison{Operator: "!=", Value: 2}, want: []string{"allocatedGPU>=0", "allocatedGPU!=2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, serverGPUFilters("allocatedGPU", tt.comparison))
		})
	}
}

func TestListWorkloadsRunFiltersMapsAndPaginates(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "2", r.URL.Query().Get("limit"))
		assert.Equal(t, "4", r.URL.Query().Get("offset"))
		assert.Equal(t, "phaseUpdatedAt", r.URL.Query().Get("sortBy"))
		assert.Equal(t, "desc", r.URL.Query().Get("sortOrder"))
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))

		filters := strings.Split(r.URL.Query().Get("filterBy"), ",")
		assert.ElementsMatch(t, []string{
			"name==trainer",
			"projectId==4500002",
			"clusterId==" + validUUID,
			"phase==Pending",
			"type==Training",
			"nodepool==gpu-a",
			"createdAt>=2026-07-05T10:00:00Z",
			"phaseUpdatedAt>=2026-07-06T10:00:00Z",
			"deletedAt>=2026-07-07T10:00:00Z",
		}, filters)

		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"next": 6,
			"workloads": []map[string]any{{
				"id":                 validUUID,
				"name":               "trainer",
				"type":               "Training",
				"phase":              "Pending",
				"projectId":          "4500002",
				"projectName":        "research",
				"clusterId":          validUUID,
				"createdAt":          "2026-07-05T10:00:00+03:00",
				"phaseUpdatedAt":     "2026-07-06T10:00:00+03:00",
				"deletedAt":          "2026-07-07T10:00:00+03:00",
				"requestedNodePools": []string{"gpu-a"},
				"currentNodePools":   []string{},
				"priority":           50,
				"priorityClassName":  "medium",
				"preemptible":        true,
			}},
		})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		Name: "trainer", ProjectID: "4500002", ClusterID: validUUID, Phase: "Pending", Type: "Training",
		NodePoolName: "gpu-a", CreatedAfter: "2026-07-05T10:00:00Z", PhaseUpdatedAfter: "2026-07-06T10:00:00Z",
		DeletedOnly: true, DeletedAfter: "2026-07-07T10:00:00Z",
		Limit: 2, Offset: 4,
	})
	require.NoError(t, err)
	require.Len(t, out.Workloads, 1)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(6), out.NextOffset)

	workload := out.Workloads[0]
	assert.Equal(t, validUUID, workload.ID)
	assert.Equal(t, "trainer", workload.Name)
	assert.Equal(t, "Training", workload.Type)
	assert.Equal(t, "Pending", workload.Phase)
	assert.Equal(t, "4500002", workload.ProjectID)
	assert.Equal(t, "research", workload.ProjectName)
	assert.Equal(t, validUUID, workload.ClusterID)
	assert.Equal(t, "2026-07-05T07:00:00Z", workload.CreatedAt)
	assert.Equal(t, "2026-07-06T07:00:00Z", workload.PhaseUpdatedAt)
	assert.Equal(t, "2026-07-07T07:00:00Z", workload.DeletedAt)
	assert.Equal(t, []string{"gpu-a"}, workload.RequestedNodePools)
	assert.Empty(t, workload.CurrentNodePools)
	require.NotNil(t, workload.Priority)
	assert.Equal(t, int32(50), *workload.Priority)
	assert.Equal(t, "medium", workload.PriorityClassName)
	require.NotNil(t, workload.Preemptible)
	assert.True(t, *workload.Preemptible)
	assert.Nil(t, workload.AllocatedGPU)
	assert.Nil(t, workload.RequestedGPU)
}

func TestListWorkloadsRunDeletedAfterSelectsDeletedCollection(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.Equal(t, "deletedAt>=2026-07-07T10:00:00Z", r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		DeletedAfter: "2026-07-07T10:00:00Z",
	})
	require.NoError(t, err)
	assert.Empty(t, out.Workloads)
}

func TestListWorkloadsRunPassesStructuredAPIFilters(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.ElementsMatch(t, []string{
			"name=@trainer",
			"deletedAt>=2026-07-13T06:00:00Z",
			"allocatedGPU>=2",
			"allocatedGPU!=2",
		}, strings.Split(r.URL.Query().Get("filterBy"), ","))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		FilterBy: []apiWorkloadFilter{
			{Field: "name", Operator: "=@", Value: "trainer"},
			{Field: "deletedAt", Operator: ">=", Value: "2026-07-13T06:00:00Z"},
			{Field: "allocatedGPU", Operator: ">", Value: "2"},
		},
	})
	require.NoError(t, err)
	assert.Empty(t, out.Workloads)
}

func TestListWorkloadsRunDeletedOnlyHasNoDeletedAtFilter(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.Empty(t, r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	_, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{DeletedOnly: true})
	require.NoError(t, err)
}

func TestListWorkloadsRunEmptyPageUsesDefaults(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Empty(t, r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{})
	require.NoError(t, err)
	assert.NotNil(t, out.Workloads)
	assert.Empty(t, out.Workloads)
	assert.False(t, out.HasMore)
	assert.Zero(t, out.NextOffset)
}

func TestListWorkloadsRunGPUIdleFilterAndMapping(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		filters := strings.Split(r.URL.Query().Get("filterBy"), ",")
		assert.ElementsMatch(t, []string{
			"phase==Running",
			"idleAllocatedGpus!=0",
		}, filters)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"workloads": []map[string]any{{
				"id":                validUUID,
				"name":              "idle-trainer",
				"type":              "Training",
				"phase":             "Running",
				"projectId":         "4500002",
				"projectName":       "research",
				"clusterId":         validUUID,
				"createdAt":         "2026-07-05T10:00:00Z",
				"phaseUpdatedAt":    "2026-07-06T10:00:00Z",
				"idleAllocatedGpus": 1.5,
			}},
		})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{GPUIdleOnly: true})
	require.NoError(t, err)
	require.Len(t, out.Workloads, 1)
	require.NotNil(t, out.Workloads[0].IdleAllocatedGpus)
	assert.Equal(t, float32(1.5), *out.Workloads[0].IdleAllocatedGpus)
}

func TestListWorkloadsRunAllocatedGPUUsesServerFilterAndMapsGPUCounts(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		filters := strings.Split(r.URL.Query().Get("filterBy"), ",")
		assert.ElementsMatch(t, []string{"allocatedGPU>=2", "allocatedGPU!=2"}, filters)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"workloads": []map[string]any{{
				"id":                         validUUID,
				"name":                       "gpu-trainer",
				"type":                       "Training",
				"phase":                      "Running",
				"projectId":                  "4500002",
				"projectName":                "research",
				"clusterId":                  validUUID,
				"createdAt":                  "2026-07-05T10:00:00Z",
				"phaseUpdatedAt":             "2026-07-06T10:00:00Z",
				"allocatedResources":         map[string]any{"gpu": 3.5},
				"workloadRequestedResources": map[string]any{"gpu": map[string]any{"request": 4.0, "limit": 4.0}},
			}},
		})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		AllocatedGPU: &gpuComparison{Operator: ">", Value: 2},
	})
	require.NoError(t, err)
	require.Len(t, out.Workloads, 1)
	require.NotNil(t, out.Workloads[0].AllocatedGPU)
	require.NotNil(t, out.Workloads[0].RequestedGPU)
	assert.Equal(t, 3.5, *out.Workloads[0].AllocatedGPU)
	assert.Equal(t, 4.0, *out.Workloads[0].RequestedGPU)
}

func TestListWorkloadsRunRequestedGPUFiltersScopedSetLocally(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("precise"))
		assert.Equal(t, "projectId==4500002", r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 4})
	})
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "200", r.URL.Query().Get("limit"))
		assert.Equal(t, "0", r.URL.Query().Get("offset"))
		assert.Equal(t, "projectId==4500002", r.URL.Query().Get("filterBy"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"workloads": []map[string]any{
				{
					"id": validUUID, "name": "one-gpu", "type": "Training", "phase": "Running",
					"projectId": "4500002", "projectName": "research", "clusterId": validUUID,
					"createdAt": "2026-07-05T10:00:00Z", "phaseUpdatedAt": "2026-07-06T10:00:00Z",
					"workloadRequestedResources": map[string]any{"gpu": map[string]any{"request": 1.0}},
				},
				{
					"id": "00000000-0000-0000-0000-000000000002", "name": "two-gpu", "type": "Training", "phase": "Running",
					"projectId": "4500002", "projectName": "research", "clusterId": validUUID,
					"createdAt": "2026-07-05T10:00:00Z", "phaseUpdatedAt": "2026-07-06T10:00:00Z",
					"workloadRequestedResources": map[string]any{"gpu": map[string]any{"request": 2.0}},
				},
				{
					"id": "00000000-0000-0000-0000-000000000003", "name": "three-gpu", "type": "Training", "phase": "Running",
					"projectId": "4500002", "projectName": "research", "clusterId": validUUID,
					"createdAt": "2026-07-05T10:00:00Z", "phaseUpdatedAt": "2026-07-06T10:00:00Z",
					"workloadRequestedResources": map[string]any{"gpu": map[string]any{"request": 3.0}},
				},
				{
					"id": "00000000-0000-0000-0000-000000000004", "name": "external-four-gpu", "type": "External", "phase": "Running",
					"projectId": "4500002", "projectName": "research", "clusterId": validUUID,
					"createdAt": "2026-07-05T10:00:00Z", "phaseUpdatedAt": "2026-07-06T10:00:00Z",
					"podsRequestedResources": map[string]any{"gpu": map[string]any{"request": 4.0}},
				},
			},
		})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		ProjectID:    "4500002",
		RequestedGPU: &gpuComparison{Operator: ">", Value: 2},
	})
	require.NoError(t, err)
	require.Len(t, out.Workloads, 2)
	assert.Equal(t, "three-gpu", out.Workloads[0].Name)
	require.NotNil(t, out.Workloads[0].RequestedGPU)
	assert.Equal(t, 3.0, *out.Workloads[0].RequestedGPU)
	assert.Equal(t, "external-four-gpu", out.Workloads[1].Name)
	require.NotNil(t, out.Workloads[1].RequestedGPU)
	assert.Equal(t, 4.0, *out.Workloads[1].RequestedGPU)
	assert.False(t, out.HasMore)
}

func TestListWorkloadsRunRequestedGPUPreservesDeletedScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	wantFilters := []string{"projectId==4500002", "deletedAt>=2026-07-07T10:00:00Z"}
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.Equal(t, "gpu", r.URL.Query().Get("search"))
		assert.ElementsMatch(t, wantFilters, strings.Split(r.URL.Query().Get("filterBy"), ","))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 0})
	})
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.Equal(t, "gpu", r.URL.Query().Get("search"))
		assert.Equal(t, "name", r.URL.Query().Get("sortBy"))
		assert.Equal(t, "asc", r.URL.Query().Get("sortOrder"))
		assert.ElementsMatch(t, wantFilters, strings.Split(r.URL.Query().Get("filterBy"), ","))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		ProjectID:    "4500002",
		DeletedAfter: "2026-07-07T10:00:00Z",
		Search:       "gpu",
		SortBy:       "name",
		SortOrder:    "asc",
		RequestedGPU: &gpuComparison{Operator: ">", Value: 1},
	})
	require.NoError(t, err)
	assert.Empty(t, out.Workloads)
}

func TestListWorkloadsRunRequestedGPUPreservesStructuredDeletedAtScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	wantFilters := []string{"projectId==4500002", "deletedAt>=2026-07-07T10:00:00Z"}
	assertRequest := func(r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("deleted"))
		assert.ElementsMatch(t, wantFilters, strings.Split(r.URL.Query().Get("filterBy"), ","))
	}
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
		assertRequest(r)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 0})
	})
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, r *http.Request) {
		assertRequest(r)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		ProjectID: "4500002",
		FilterBy: []apiWorkloadFilter{{
			Field: "deletedAt", Operator: ">=", Value: "2026-07-07T10:00:00Z",
		}},
		RequestedGPU: &gpuComparison{Operator: ">", Value: 1},
	})
	require.NoError(t, err)
	assert.Empty(t, out.Workloads)
}

func TestCompareGPUDoesNotTreatMissingValueAsZero(t *testing.T) {
	for _, operator := range []string{"<", "<=", "==", "!=", ">=", ">"} {
		t.Run(operator, func(t *testing.T) {
			assert.False(t, compareGPU(nil, gpuComparison{Operator: operator, Value: 0}))
		})
	}
}

func TestListWorkloadsRunRequestedGPURejectsScopeOver200BeforeListing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 201})
	})
	listCalled := false
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, _ *http.Request) {
		listCalled = true
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"workloads": []any{}})
	})

	_, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		ClusterID:    validUUID,
		RequestedGPU: &gpuComparison{Operator: ">=", Value: 2},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ask the user to narrow")
	assert.False(t, listCalled, "oversized scopes must be rejected before fetching workload rows")
}

func TestListWorkloadsRunRequestedGPUAcceptsExactly200(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	countCalls := 0
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
		countCalls++
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 200})
	})
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, _ *http.Request) {
		workloads := make([]map[string]any, 200)
		for i := range workloads {
			workloads[i] = map[string]any{
				"id":   fmt.Sprintf("00000000-0000-0000-0000-%012d", i),
				"name": fmt.Sprintf("gpu-%d", i), "type": "Training", "phase": "Running",
				"projectId": "4500002", "projectName": "research", "clusterId": validUUID,
				"createdAt": "2026-07-05T10:00:00Z", "phaseUpdatedAt": "2026-07-06T10:00:00Z",
				"workloadRequestedResources": map[string]any{"gpu": map[string]any{"request": 1.0}},
			}
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"next": 200, "workloads": workloads})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{
		ProjectID:    "4500002",
		RequestedGPU: &gpuComparison{Operator: ">=", Value: 1},
	})
	require.NoError(t, err)
	assert.Len(t, out.Workloads, 50)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(50), out.NextOffset)
	assert.Equal(t, 2, countCalls, "the full-page boundary is recounted to distinguish exactly 200 from a growing scope")
}

func TestListWorkloadsRunIgnoresNonAdvancingNextOffset(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"next":      20,
			"workloads": []any{},
		})
	})

	out, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{Offset: 20})
	require.NoError(t, err)
	assert.False(t, out.HasMore)
	assert.Zero(t, out.NextOffset)
}

func TestListWorkloadsRunMapsUpstreamError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "denied"})
	})

	_, err := listWorkloadsRun(context.Background(), client, listWorkloadsInput{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, runaimcp.ErrForbidden))
}
