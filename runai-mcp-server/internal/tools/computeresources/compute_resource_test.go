/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package computeresources

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	testComputeResourceID      = "1051ad5d-1a5a-4bb8-a0d1-c17d622d7fe8"
	testOtherComputeResourceID = "52f87537-2e01-42d7-9c66-362636419afb"
	testComputeClusterID       = "8a529bf5-d2aa-47bf-a523-bb36427d0875"
	testFingerprintPrefix      = "sha256:"
	testFingerprint            = testFingerprintPrefix + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func TestManageComputeResourceInputValidate(t *testing.T) {
	validCreate := map[string]any{
		"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
		"spec": map[string]any{"gpuDevicesRequest": float64(1)},
	}
	validUpdate := map[string]any{
		"meta": map[string]any{"name": "gpu-small"},
		"spec": map[string]any{"gpuDevicesRequest": float64(1)},
	}
	tests := []struct {
		name    string
		in      manageComputeResourceInput
		wantErr string
	}{
		{name: "create", in: manageComputeResourceInput{Action: "create", Request: validCreate}},
		{name: "update", in: manageComputeResourceInput{Action: "update", ComputeResourceID: testComputeResourceID, Request: validUpdate, ExpectedFingerprint: testFingerprint}},
		{name: "delete", in: manageComputeResourceInput{Action: "delete", ComputeResourceID: testComputeResourceID}},
		{name: "invalid_action", in: manageComputeResourceInput{Action: "edit"}, wantErr: "action must be one of"},
		{name: "create_with_id", in: manageComputeResourceInput{Action: "create", ComputeResourceID: testComputeResourceID, Request: validCreate}, wantErr: "must not be set"},
		{name: "create_system_scope", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "system", "scope": "system"}, "spec": map[string]any{},
		}}, wantErr: "system is not supported"},
		{name: "create_cluster_without_id", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "cluster", "scope": "cluster"}, "spec": map[string]any{},
		}}, wantErr: "clusterId"},
		{name: "create_null_spec", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"}, "spec": nil,
		}}},
		{name: "create_tenant_with_null_scope_ids", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{
				"name": "gpu-small", "scope": "tenant", "clusterId": nil, "departmentId": nil, "projectId": nil,
			},
			"spec": map[string]any{},
		}}},
		{name: "update_without_fingerprint", in: manageComputeResourceInput{Action: "update", ComputeResourceID: testComputeResourceID, Request: validUpdate}, wantErr: "expectedFingerprint is required"},
		{name: "update_moves_scope", in: manageComputeResourceInput{Action: "update", ComputeResourceID: testComputeResourceID, ExpectedFingerprint: testFingerprint, Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"}, "spec": map[string]any{},
		}}, wantErr: "create-only"},
		{name: "update_null_spec", in: manageComputeResourceInput{Action: "update", ComputeResourceID: testComputeResourceID, ExpectedFingerprint: testFingerprint, Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small"}, "spec": nil,
		}}, wantErr: "must be an object for update"},
		{name: "delete_with_request", in: manageComputeResourceInput{Action: "delete", ComputeResourceID: testComputeResourceID, Request: validUpdate}, wantErr: "must not be set"},
		{name: "unknown_spec_field", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
			"spec": map[string]any{"futureField": true},
		}}, wantErr: "unknown field"},
		{name: "tenant_with_zero_project", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant", "projectId": float64(0)}, "spec": map[string]any{},
		}}, wantErr: "must not set"},
		{name: "project_with_ignored_cluster", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "project", "projectId": int32(42), "clusterId": ""}, "spec": map[string]any{},
		}}, wantErr: "must not set clusterId"},
		{name: "private_compute", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant", "private": true}, "spec": map[string]any{},
		}}, wantErr: "private=true is not supported"},
		{name: "workload_types", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant", "workloadSupportedTypes": map[string]any{}}, "spec": map[string]any{},
		}}, wantErr: "workloadSupportedTypes is not supported"},
		{name: "empty_description", in: manageComputeResourceInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant", "description": ""}, "spec": map[string]any{},
		}}, wantErr: "between 1 and 250"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.in.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestListComputeResourcesInputValidate(t *testing.T) {
	projectID := int32(42)
	tests := []struct {
		name    string
		in      listComputeResourcesInput
		wantErr string
	}{
		{name: "all"},
		{name: "all_filters", in: listComputeResourcesInput{
			ComputeResourceID: testComputeResourceID, Name: "gpu-small",
			ClusterID: testComputeClusterID, DepartmentID: "department-a", ProjectID: &projectID,
			IncludeDescendants: true, Limit: 10, Offset: 5,
		}},
		{name: "bad_id", in: listComputeResourcesInput{ComputeResourceID: "bad"}, wantErr: "UUID"},
		{name: "system_scope", in: listComputeResourcesInput{Scope: "system"}, wantErr: "scope must be one of"},
		{name: "bad_scope", in: listComputeResourcesInput{Scope: "user"}, wantErr: "scope must be one of"},
		{name: "bad_cluster", in: listComputeResourcesInput{ClusterID: "bad"}, wantErr: "UUID"},
		{name: "bad_project", in: listComputeResourcesInput{ProjectID: int32Ptr(0)}, wantErr: "positive"},
		{name: "bad_limit", in: listComputeResourcesInput{Limit: 201}, wantErr: "limit"},
		{name: "bad_offset", in: listComputeResourcesInput{Offset: -1}, wantErr: "offset"},
		{name: "descendants_with_scope", in: listComputeResourcesInput{
			Scope: "cluster", ClusterID: testComputeClusterID, IncludeDescendants: true,
		}, wantErr: "mutually exclusive"},
		{name: "descendants_without_id", in: listComputeResourcesInput{IncludeDescendants: true}, wantErr: "requires at least one"},
		{name: "blank_department", in: listComputeResourcesInput{DepartmentID: "   "}, wantErr: "required"},
		{name: "direct_query_metacharacters", in: listComputeResourcesInput{
			Name: "gpu,small==1", DepartmentID: "dept,one==1",
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.in.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestListComputeResourcesReturnsUpdateSnapshotsAndPaginates(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		query := r.URL.Query()
		assert.Empty(t, query.Get("name"))
		assert.Empty(t, query.Get("scope"))
		assert.Equal(t, testComputeClusterID, query.Get("clusterId"))
		assert.Equal(t, "department-a", query.Get("departmentId"))
		assert.Equal(t, "42", query.Get("projectId"))
		assert.Equal(t, "true", query.Get("includeDescendants"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "gpu-a", "project", map[string]any{"gpuDevicesRequest": 1}),
			computeResourceResponse(testOtherComputeResourceID, "gpu-b", "project", map[string]any{"cpuCoreRequest": 2}),
		}})
	})

	projectID := int32(42)
	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{
		ClusterID:    testComputeClusterID,
		DepartmentID: "department-a", ProjectID: &projectID, IncludeDescendants: true, Limit: 1,
	})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	assert.Equal(t, int32(2), out.TotalCount)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(1), out.NextOffset)
	item := out.ComputeResources[0]
	assert.Equal(t, testComputeResourceID, item.ComputeResourceID)
	assert.Equal(t, "gpu-a", item.Name)
	assert.Equal(t, "project", item.Scope)
	assert.Equal(t, "gpu-a", item.UpdateRequest["meta"].(map[string]any)["name"])
	assert.Equal(t, float64(1), item.UpdateRequest["spec"].(map[string]any)["gpuDevicesRequest"])
	assert.True(t, len(item.ExpectedFingerprint) > len(testFingerprintPrefix))
}

func TestListComputeResourcesEnforcesExactNameLocally(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "gpu-copy", "tenant", map[string]any{}),
			computeResourceResponse(testOtherComputeResourceID, "gpu", "tenant", map[string]any{}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{Name: "gpu"})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	assert.Equal(t, testOtherComputeResourceID, out.ComputeResources[0].ComputeResourceID)
}

func TestListComputeResourcesEnforcesExactNonTenantScopeLocally(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "tenant", "tenant", map[string]any{}),
			computeResourceResponse(testOtherComputeResourceID, "cluster", "cluster", map[string]any{}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{Scope: "cluster"})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	assert.Equal(t, testOtherComputeResourceID, out.ComputeResources[0].ComputeResourceID)
}

func TestListComputeResourcesRejectsUnknownResponseScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "gpu", "future", map[string]any{}),
		}})
	})

	_, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestListComputeResourcesPreparesOnlyRequestedPage(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		invalidOffPageResource := computeResourceResponse(
			testOtherComputeResourceID, "invalid-off-page", "tenant", map[string]any{},
		)
		invalidOffPageResource["spec"] = nil
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "gpu-a", "tenant", map[string]any{}),
			invalidOffPageResource,
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{Limit: 1})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	assert.Equal(t, testComputeResourceID, out.ComputeResources[0].ComputeResourceID)
	assert.Equal(t, int32(2), out.TotalCount)
	assert.True(t, out.HasMore)
}

func TestListComputeResourcesFiltersByIDLocally(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "gpu-a", "tenant", map[string]any{}),
			computeResourceResponse(testOtherComputeResourceID, "gpu-b", "tenant", map[string]any{}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{
		ComputeResourceID: testOtherComputeResourceID,
	})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	assert.Equal(t, testOtherComputeResourceID, out.ComputeResources[0].ComputeResourceID)
	assert.Equal(t, int32(1), out.TotalCount)
}

func TestListComputeResourcesPreservesInheritedSystemPresetsForTenantScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "tenant", r.URL.Query().Get("scope"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "preset", "system", map[string]any{}),
			computeResourceResponse(testOtherComputeResourceID, "tenant-owned", "tenant", map[string]any{}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{Scope: "tenant"})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 2)
	assert.Equal(t, testComputeResourceID, out.ComputeResources[0].ComputeResourceID)
	assert.Equal(t, testOtherComputeResourceID, out.ComputeResources[1].ComputeResourceID)
	assert.Equal(t, int32(2), out.TotalCount)
}

func TestListComputeResourcesPreservesDirectQueryValues(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gpu,small==1", r.URL.Query().Get("name"))
		assert.Equal(t, "dept,one==1", r.URL.Query().Get("departmentId"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{
		Name: "gpu,small==1", DepartmentID: "dept,one==1",
	})
	require.NoError(t, err)
	assert.Empty(t, out.ComputeResources)
}

func TestListComputeResourcesKeepsNewerSpecsReadable(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "future", "tenant", map[string]any{"futureField": true}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	item := out.ComputeResources[0]
	assert.Equal(t, true, item.ComputeResource["spec"].(map[string]any)["futureField"])
	assert.Nil(t, item.UpdateRequest)
	assert.Empty(t, item.ExpectedFingerprint)
	assert.Contains(t, item.UpdateBlockedReason, "futureField")
}

func TestListComputeResourcesKeepsNewerMetadataReadable(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	resource := computeResourceResponse(testComputeResourceID, "future-meta", "tenant", map[string]any{})
	resource["meta"].(map[string]any)["futureField"] = true
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{resource}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	item := out.ComputeResources[0]
	assert.Equal(t, true, item.ComputeResource["meta"].(map[string]any)["futureField"])
	assert.Nil(t, item.UpdateRequest)
	assert.Empty(t, item.ExpectedFingerprint)
	assert.Contains(t, item.UpdateBlockedReason, "meta.futureField")
}

func TestListComputeResourcesKeepsMalformedSpecsReadable(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	resourceWithoutSpec := computeResourceResponse(testComputeResourceID, "missing-spec", "tenant", map[string]any{})
	delete(resourceWithoutSpec, "spec")
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{resourceWithoutSpec}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	item := out.ComputeResources[0]
	assert.Equal(t, resourceWithoutSpec, item.ComputeResource)
	assert.Nil(t, item.UpdateRequest)
	assert.Empty(t, item.ExpectedFingerprint)
	assert.Contains(t, item.UpdateBlockedReason, "spec must be an object")
}

func TestListComputeResourcesMarksSystemPresetsReadOnly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			computeResourceResponse(testComputeResourceID, "preset", "system", map[string]any{}),
		}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.NoError(t, err)
	require.Len(t, out.ComputeResources, 1)
	item := out.ComputeResources[0]
	assert.Nil(t, item.UpdateRequest)
	assert.Empty(t, item.ExpectedFingerprint)
	assert.Contains(t, item.UpdateBlockedReason, "read-only")
}

func TestListComputeResourcesMapsForbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestListComputeResourcesReturnsNonNilEmptyArray(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{}})
	})

	out, err := listComputeResourcesRun(context.Background(), client, listComputeResourcesInput{})
	require.NoError(t, err)
	assert.NotNil(t, out.ComputeResources)
	assert.Empty(t, out.ComputeResources)
	assert.Zero(t, out.TotalCount)
	assert.False(t, out.HasMore)
}

func TestCreateComputeResourceSendsCompleteRequest(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var body map[string]any
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		response := computeResourceResponse(testComputeResourceID, "gpu-small", "cluster", map[string]any{
			"gpuDevicesRequest": 1, "largeShmRequest": true,
		})
		response["meta"].(map[string]any)["description"] = "one GPU"
		runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
	})

	in := manageComputeResourceInput{Action: "create", Request: map[string]any{
		"meta": map[string]any{
			"name": "gpu-small", "description": "one GPU", "scope": "cluster", "clusterId": testComputeClusterID,
		},
		"spec": map[string]any{"gpuDevicesRequest": float64(1), "largeShmRequest": true},
	}}
	require.NoError(t, in.Validate())
	out, err := manageComputeResourceRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, testComputeResourceID, out.ComputeResourceID)
	assert.Empty(t, out.Before)
	assert.Equal(t, "gpu-small", body["meta"].(map[string]any)["name"])
	assert.Equal(t, testComputeClusterID, body["meta"].(map[string]any)["clusterId"])
	assert.NotEmpty(t, out.Diff)
}

func TestCreateComputeResourceSupportsNullSpec(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			return
		}
		assert.Contains(t, body, "spec")
		assert.Nil(t, body["spec"])
		runaiclienttest.WriteJSON(w, http.StatusAccepted,
			computeResourceResponse(testComputeResourceID, "gpu-empty", "tenant", map[string]any{}))
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "create",
		Request: map[string]any{
			"meta": map[string]any{"name": "gpu-empty", "scope": "tenant"},
			"spec": nil,
		},
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, testComputeResourceID, out.ComputeResourceID)
}

func TestCreateComputeResourceMapsConflict(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusConflict, map[string]any{"message": "already exists"})
	})

	_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "create",
		Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
			"spec": map[string]any{},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConflict)
}

func TestCreateComputeResourceRejectsWrongResponseName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusAccepted,
			computeResourceResponse(testComputeResourceID, "different", "tenant", map[string]any{}))
	})

	_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "create",
		Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
			"spec": map[string]any{},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), `returned name "different"`)
}

func TestCreateComputeResourceRejectsWrongResponseScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusAccepted,
			computeResourceResponse(testComputeResourceID, "gpu-small", "cluster", map[string]any{}))
	})

	_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "create",
		Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
			"spec": map[string]any{},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "returned scope")
}

func TestCreateComputeResourceRejectsUnexpectedDescription(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute", func(w http.ResponseWriter, _ *http.Request) {
		response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
		response["meta"].(map[string]any)["description"] = "unexpected"
		runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
	})

	_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "create",
		Request: map[string]any{
			"meta": map[string]any{"name": "gpu-small", "scope": "tenant"},
			"spec": map[string]any{},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "description does not match")
}

func TestUpdateComputeResourceUsesFingerprintAndRechecksBeforePut(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{"cpuCoreRequest": 1})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	request["spec"] = map[string]any{"cpuCoreRequest": float64(2)}
	var getCalls atomic.Int32
	var putCalls atomic.Int32
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCalls.Add(1)
			runaiclienttest.WriteJSON(w, http.StatusOK, current)
		case http.MethodPut:
			putCalls.Add(1)
			var body map[string]any
			if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
				return
			}
			assert.Equal(t, float64(2), body["spec"].(map[string]any)["cpuCoreRequest"])
			runaiclienttest.WriteJSON(w, http.StatusAccepted,
				computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", body["spec"].(map[string]any)))
		default:
			assert.Fail(t, "unexpected method", r.Method)
		}
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, int32(2), getCalls.Load())
	assert.Equal(t, int32(1), putCalls.Load())
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.After)
	assert.NotEmpty(t, out.Diff)
}

func TestUpdateComputeResourceRequiresConfirmationBeforePut(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	request["meta"].(map[string]any)["name"] = "gpu-renamed"
	var putCalls atomic.Int32
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putCalls.Add(1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Zero(t, putCalls.Load())
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.Diff)
}

func TestUpdateComputeResourceRejectsStaleFingerprint(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{"cpuCoreRequest": 2})
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request:             map[string]any{"meta": map[string]any{"name": "gpu-small"}, "spec": map[string]any{}},
		ExpectedFingerprint: testFingerprint, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConflict)
	assert.Contains(t, err.Error(), "changed since")
}

func TestUpdateComputeResourceRejectsWrongResponseID(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			runaiclienttest.WriteJSON(w, http.StatusAccepted, computeResourceResponse(
				testOtherComputeResourceID, "wrong", "tenant", map[string]any{}))
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	_, err = manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), testOtherComputeResourceID)
}

func TestUpdateComputeResourceRejectsWrongResponseName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	request["meta"].(map[string]any)["name"] = "gpu-renamed"
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			runaiclienttest.WriteJSON(w, http.StatusAccepted, computeResourceResponse(
				testComputeResourceID, "different", "tenant", map[string]any{}))
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	_, err = manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), `returned name "different"`)
}

func TestUpdateComputeResourceRejectsChangedResponseScope(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			runaiclienttest.WriteJSON(w, http.StatusAccepted, computeResourceResponse(
				testComputeResourceID, "gpu-small", "cluster", map[string]any{}))
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	_, err = manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "changed scope")
}

func TestUpdateComputeResourceRejectsIgnoredSpecChange(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{"cpuCoreRequest": 1})
	request, fingerprint, err := computeResourceUpdateSnapshot(current)
	require.NoError(t, err)
	request["spec"] = map[string]any{"cpuCoreRequest": float64(2)}
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			runaiclienttest.WriteJSON(w, http.StatusAccepted, current)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	_, err = manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "update", ComputeResourceID: testComputeResourceID,
		Request: request, ExpectedFingerprint: fingerprint, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "spec does not match")
}

func TestGetComputeResourceAcceptsCanonicalizedUUID(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/compute/"+strings.ToUpper(testComputeResourceID), func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, computeResourceResponse(
			testComputeResourceID, "gpu-small", "tenant", map[string]any{}))
	})

	target, err := getComputeResource(context.Background(), client, strings.ToUpper(testComputeResourceID), true)
	require.NoError(t, err)
	assert.Equal(t, "gpu-small", target.name)
}

func TestDeleteComputeResourceIgnoreMissingOnlySuppressesNotFound(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, _ *http.Request) {
			runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "missing"})
		})

		out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
			Action: "delete", ComputeResourceID: testComputeResourceID, IgnoreMissing: true, Confirm: true,
		})
		require.NoError(t, err)
		assert.False(t, out.Applied)
		require.Len(t, out.Warnings, 1)
		assert.Contains(t, out.Warnings[0], "nothing to delete")
	})

	t.Run("forbidden", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, _ *http.Request) {
			runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
		})

		_, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
			Action: "delete", ComputeResourceID: testComputeResourceID, IgnoreMissing: true, Confirm: true,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	})
}

func TestDeleteComputeResourceRequiresConfirmationBeforeDelete(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	var deleteCalls atomic.Int32
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalls.Add(1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "delete", ComputeResourceID: testComputeResourceID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Zero(t, deleteCalls.Load())
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.Diff)
}

func TestManageComputeResourceRejectsSystemPresetMutationsBeforeConfirmation(t *testing.T) {
	for _, action := range []string{"update", "delete"} {
		t.Run(action, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			var mutationCalls atomic.Int32
			mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					mutationCalls.Add(1)
				}
				runaiclienttest.WriteJSON(w, http.StatusOK,
					computeResourceResponse(testComputeResourceID, "preset", "system", map[string]any{}))
			})

			in := manageComputeResourceInput{
				Action: action, ComputeResourceID: testComputeResourceID, Confirm: true,
			}
			if action == "update" {
				in.Request = map[string]any{"meta": map[string]any{"name": "preset"}, "spec": map[string]any{}}
				in.ExpectedFingerprint = testFingerprint
			}

			out, err := manageComputeResourceRun(context.Background(), client, in)
			require.Error(t, err)
			assert.ErrorIs(t, err, runaimcp.ErrForbidden)
			assert.Zero(t, mutationCalls.Load())
			assert.Equal(t, "preset", out.Name)
			assert.NotEmpty(t, out.Before)
		})
	}
}

func TestDeleteComputeResourceConfirmed(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, current)
		case http.MethodDelete:
			runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{"code": 202, "message": "accepted"})
		default:
			assert.Fail(t, "unexpected method", r.Method)
		}
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "delete", ComputeResourceID: testComputeResourceID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.NotEmpty(t, out.Before)
	assert.Empty(t, out.After)
}

func TestDeleteComputeResourceDoesNotRequireUpdateSnapshot(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	resourceWithoutSpec := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	delete(resourceWithoutSpec, "spec")
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, resourceWithoutSpec)
		case http.MethodDelete:
			runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{"code": 202, "message": "accepted"})
		default:
			assert.Fail(t, "unexpected method", r.Method)
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "delete", ComputeResourceID: testComputeResourceID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, resourceWithoutSpec, out.Before)
}

func TestDeleteComputeResourceIgnoreMissingHandlesDeleteRace(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	mux.HandleFunc("/api/v1/asset/compute/"+testComputeResourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "gone"})
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, current)
	})

	out, err := manageComputeResourceRun(context.Background(), client, manageComputeResourceInput{
		Action: "delete", ComputeResourceID: testComputeResourceID, IgnoreMissing: true, Confirm: true,
	})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.Diff)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "nothing to delete")
}

func TestComputeResourceUpdateSnapshotPreservesUnknownSpecFields(t *testing.T) {
	response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{
		"futureField": true,
	})

	request, fingerprint, err := computeResourceUpdateSnapshot(response)
	require.NoError(t, err)
	assert.Equal(t, true, request["spec"].(map[string]any)["futureField"])
	assert.NotEmpty(t, fingerprint)
}

func TestComputeResourceUpdateSnapshotAllowsKnownReadOnlyFields(t *testing.T) {
	response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	response["usedBy"] = map[string]any{"templates": []any{}}
	response["usageTimes"] = map[string]any{"lastUsedAt": nil}
	response["compliance"] = map[string]any{"compliant": true}
	response["meta"].(map[string]any)["updateCount"] = float64(3)

	request, fingerprint, err := computeResourceUpdateSnapshot(response)
	require.NoError(t, err)
	assert.NotNil(t, request)
	assert.NotEmpty(t, fingerprint)
}

func TestComputeResourceUpdateSnapshotRejectsUnknownResponseFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
		set   func(map[string]any)
	}{
		{name: "top_level", field: "futureField", set: func(response map[string]any) {
			response["futureField"] = true
		}},
		{name: "metadata", field: "meta.futureField", set: func(response map[string]any) {
			response["meta"].(map[string]any)["futureField"] = true
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
			test.set(response)

			_, _, err := computeResourceUpdateSnapshot(response)
			require.Error(t, err)
			assert.ErrorIs(t, err, runaimcp.ErrConflict)
			assert.Contains(t, err.Error(), test.field)
		})
	}
}

func TestComputeResourceUpdateSnapshotRejectsMissingSpec(t *testing.T) {
	response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	delete(response, "spec")

	_, _, err := computeResourceUpdateSnapshot(response)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestComputeResourceIdentityRejectsMalformedID(t *testing.T) {
	response := computeResourceResponse("not-a-uuid", "gpu-small", "tenant", map[string]any{})

	_, _, _, err := computeResourceIdentity(response)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "meta.id must be a UUID")
}

func TestComputeResourceIdentityRejectsWrongKind(t *testing.T) {
	response := computeResourceResponse(testComputeResourceID, "gpu-small", "tenant", map[string]any{})
	response["meta"].(map[string]any)["kind"] = "environment"

	_, _, _, err := computeResourceIdentity(response)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.Contains(t, err.Error(), "kind must be")
}

func TestComputeResourceIdentityRejectsMalformedScopeMetadata(t *testing.T) {
	tests := []struct {
		name   string
		scope  string
		mutate func(map[string]any)
		want   string
	}{
		{"cluster_without_id", "cluster", func(meta map[string]any) { delete(meta, "clusterId") }, "clusterId"},
		{"department_without_id", "department", func(meta map[string]any) { delete(meta, "departmentId") }, "departmentId"},
		{"project_without_id", "project", func(meta map[string]any) { delete(meta, "projectId") }, "projectId"},
		{"private_compute", "tenant", func(meta map[string]any) { meta["private"] = true }, "must not be private"},
		{"empty_description", "tenant", func(meta map[string]any) { meta["description"] = "" }, "description must contain"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := computeResourceResponse(testComputeResourceID, "gpu-small", test.scope, map[string]any{})
			test.mutate(response["meta"].(map[string]any))

			_, _, _, err := computeResourceIdentity(response)
			require.Error(t, err)
			assert.ErrorIs(t, err, runaimcp.ErrUpstream)
			assert.Contains(t, err.Error(), test.want)
		})
	}
}

func TestReadComputeResourceResponse(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader(`{"name":"gpu"}`))}
		got, err := readComputeResourceResponse(resp, "read compute resource")
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"name": "gpu"}, got)
	})

	for _, test := range []struct {
		name string
		resp *http.Response
	}{
		{name: "nil_response"},
		{name: "nil_body", resp: &http.Response{}},
		{name: "invalid_json", resp: &http.Response{Body: io.NopCloser(strings.NewReader("{"))}},
		{name: "json_null", resp: &http.Response{Body: io.NopCloser(strings.NewReader("null"))}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := readComputeResourceResponse(test.resp, "read compute resource")
			require.Error(t, err)
			assert.ErrorIs(t, err, runaimcp.ErrUpstream)
		})
	}
}

func TestComputeResourceJSONFingerprint(t *testing.T) {
	first, err := computeResourceJSONFingerprint(map[string]any{"name": "gpu", "count": 2})
	require.NoError(t, err)
	second, err := computeResourceJSONFingerprint(map[string]any{"count": 2, "name": "gpu"})
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.True(t, isValidComputeResourceFingerprint(first))

	_, err = computeResourceJSONFingerprint(math.Inf(1))
	assert.Error(t, err)
}

func TestIsValidComputeResourceFingerprint(t *testing.T) {
	valid := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tests := []struct {
		value string
		want  bool
	}{
		{value: valid, want: true},
		{value: strings.TrimPrefix(valid, "sha256:")},
		{value: "sha256:" + strings.ToUpper(strings.TrimPrefix(valid, "sha256:"))},
		{value: "sha256:00"},
		{value: "sha256:zz23456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, isValidComputeResourceFingerprint(test.value))
	}
}

func computeResourceResponse(id, name, scope string, spec map[string]any) map[string]any {
	meta := map[string]any{
		"id": id, "name": name, "scope": scope, "kind": "compute",
		"createdBy": "creator@example.com", "createdAt": "2026-10-01T10:00:00Z",
		"updatedBy": "updater@example.com", "updatedAt": "2026-10-02T10:00:00Z",
	}
	switch scope {
	case "cluster":
		meta["clusterId"] = testComputeClusterID
	case "department":
		meta["departmentId"] = "department-a"
	case "project":
		meta["projectId"] = float64(42)
	}
	return map[string]any{
		"meta": meta,
		"spec": spec,
	}
}

func int32Ptr(value int32) *int32 {
	return &value
}
