/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	testEnvironmentID      = "1051ad5d-1a5a-4bb8-a0d1-c17d622d7fe8"
	testOtherEnvironmentID = "52f87537-2e01-42d7-9c66-362636419afb"
	testEnvironmentCluster = "8a529bf5-d2aa-47bf-a523-bb36427d0875"
)

func TestManageEnvironmentInputValidate(t *testing.T) {
	create := environmentCreateRequest("pytorch", "tenant", map[string]any{"training": true}, map[string]any{"image": "pytorch:latest"})
	update := map[string]any{"meta": map[string]any{"name": "pytorch"}, "spec": map[string]any{"image": "pytorch:2"}}
	updateCount := int64(3)
	tests := []struct {
		name    string
		in      manageEnvironmentInput
		wantErr string
	}{
		{name: "create", in: manageEnvironmentInput{Action: "create", Request: create}},
		{name: "update", in: manageEnvironmentInput{Action: "update", EnvironmentID: testEnvironmentID, Request: update, ExpectedUpdateCount: &updateCount}},
		{name: "delete", in: manageEnvironmentInput{Action: "delete", EnvironmentID: testEnvironmentID}},
		{name: "invalid_action", in: manageEnvironmentInput{Action: "edit"}, wantErr: "action must be one of"},
		{name: "create_with_id", in: manageEnvironmentInput{Action: "create", EnvironmentID: testEnvironmentID, Request: create}, wantErr: "must not be set"},
		{name: "create_without_request", in: manageEnvironmentInput{Action: "create"}, wantErr: "request is required"},
		{name: "create_without_meta", in: manageEnvironmentInput{Action: "create", Request: map[string]any{"spec": map[string]any{}}}, wantErr: "request.meta is required"},
		{name: "create_without_name", in: manageEnvironmentInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"scope": "tenant", "workloadSupportedTypes": map[string]any{"workspace": true}}, "spec": map[string]any{},
		}}, wantErr: "request.meta.name is required"},
		{name: "create_without_scope", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("pytorch", "", map[string]any{"workspace": true}, map[string]any{})}, wantErr: "request.meta.scope is required"},
		{name: "create_without_workload_types", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("pytorch", "tenant", nil, map[string]any{})}, wantErr: "workloadSupportedTypes"},
		{name: "create_without_spec", in: manageEnvironmentInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "pytorch", "scope": "tenant", "workloadSupportedTypes": map[string]any{"workspace": true}},
		}}, wantErr: "request.spec is required"},
		{name: "create_system", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("system", "system", map[string]any{"workspace": true}, map[string]any{})}, wantErr: "scope must be one of"},
		{name: "create_private", in: manageEnvironmentInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "private", "scope": "tenant", "private": true, "workloadSupportedTypes": map[string]any{"workspace": true}}, "spec": map[string]any{},
		}}, wantErr: "private=true is not supported"},
		{name: "create_cluster_without_id", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("cluster", "cluster", map[string]any{"workspace": true}, map[string]any{})}, wantErr: "clusterId"},
		{name: "create_department_without_id", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("department", "department", map[string]any{"workspace": true}, map[string]any{})}, wantErr: "departmentId"},
		{name: "create_project_without_id", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("project", "project", map[string]any{"workspace": true}, map[string]any{})}, wantErr: "projectId"},
		{name: "create_project_id_overflow", in: manageEnvironmentInput{Action: "create", Request: map[string]any{
			"meta": map[string]any{"name": "project", "scope": "project", "projectId": int64(1 << 31), "workloadSupportedTypes": map[string]any{"workspace": true}}, "spec": map[string]any{},
		}}, wantErr: "positive 32-bit integer"},
		{name: "unknown_spec_field", in: manageEnvironmentInput{Action: "create", Request: environmentCreateRequest("future", "tenant", map[string]any{"workspace": true}, map[string]any{"futureField": true})}, wantErr: "unknown field"},
		{name: "update_without_count", in: manageEnvironmentInput{Action: "update", EnvironmentID: testEnvironmentID, Request: update}, wantErr: "expectedUpdateCount is required"},
		{name: "update_create_fields", in: manageEnvironmentInput{Action: "update", EnvironmentID: testEnvironmentID, ExpectedUpdateCount: &updateCount, Request: map[string]any{
			"meta": map[string]any{"name": "pytorch", "scope": "tenant"}, "spec": map[string]any{},
		}}, wantErr: "unknown field"},
		{name: "delete_with_request", in: manageEnvironmentInput{Action: "delete", EnvironmentID: testEnvironmentID, Request: update}, wantErr: "must not be set"},
		{name: "ignore_missing_on_update", in: manageEnvironmentInput{Action: "update", EnvironmentID: testEnvironmentID, Request: update, ExpectedUpdateCount: &updateCount, IgnoreMissing: true}, wantErr: "only valid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.in.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestListEnvironmentsByIDUsesGetEndpoint(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK,
			environmentResponse(testEnvironmentID, "python", "tenant", 2, map[string]any{"workspace": true}, map[string]any{"image": "python:3"}))
	})

	out, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{EnvironmentID: testEnvironmentID})
	require.NoError(t, err)
	require.Len(t, out.Environments, 1)
	assert.Equal(t, testEnvironmentID, out.Environments[0].EnvironmentID)
	assert.Equal(t, int32(1), out.TotalCount)
}

func TestListEnvironmentsByIDReturnsEmptyForMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "missing"})
	})

	out, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{EnvironmentID: testEnvironmentID})
	require.NoError(t, err)
	assert.Empty(t, out.Environments)
	assert.Zero(t, out.TotalCount)
}

func TestListEnvironmentsPreservesUnknownFieldsAndBlocksUpdate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	response := environmentResponse(testEnvironmentID, "future", "tenant", 2, map[string]any{"workspace": true}, map[string]any{
		"image": "python:3", "futureField": true,
	})
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{response}})
	})

	out, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{})
	require.NoError(t, err)
	require.Len(t, out.Environments, 1)
	assert.Equal(t, true, out.Environments[0].Environment["spec"].(map[string]any)["futureField"])
	assert.Nil(t, out.Environments[0].UpdateRequest)
	assert.Contains(t, out.Environments[0].UpdateBlockedReason, "cannot safely preserve")
}

func TestListEnvironmentsInputValidate(t *testing.T) {
	projectID := int32(42)
	tests := []struct {
		name    string
		in      listEnvironmentsInput
		wantErr string
	}{
		{name: "all"},
		{name: "all_filters", in: listEnvironmentsInput{EnvironmentID: testEnvironmentID, ClusterID: testEnvironmentCluster, ProjectID: &projectID, WorkloadType: "Distributed", DistributedFramework: "MPI", IncludeDescendants: true}},
		{name: "bad_id", in: listEnvironmentsInput{EnvironmentID: "bad"}, wantErr: "UUID"},
		{name: "system_scope", in: listEnvironmentsInput{Scope: "system"}, wantErr: "scope must be one of"},
		{name: "bad_workload_type", in: listEnvironmentsInput{WorkloadType: "Job"}, wantErr: "workloadType"},
		{name: "bad_framework", in: listEnvironmentsInput{DistributedFramework: "Ray"}, wantErr: "distributedFramework must be one of"},
		{name: "framework_with_training", in: listEnvironmentsInput{WorkloadType: "Training", DistributedFramework: "MPI"}, wantErr: "requires workloadType=Distributed"},
		{name: "descendants_with_scope", in: listEnvironmentsInput{Scope: "cluster", ClusterID: testEnvironmentCluster, IncludeDescendants: true}, wantErr: "mutually exclusive"},
		{name: "bad_limit", in: listEnvironmentsInput{Limit: 201}, wantErr: "limit"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.in.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestListEnvironmentsFrameworkImpliesDistributed(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("isDistributed"))
		assert.Equal(t, "MPI", r.URL.Query().Get("distributedFramework"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{}})
	})

	_, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{DistributedFramework: "MPI"})
	require.NoError(t, err)
}

func TestListEnvironmentsReturnsUpdateDataAndPaginates(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("isDistributed"))
		assert.Equal(t, "MPI", r.URL.Query().Get("distributedFramework"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			environmentResponse(testEnvironmentID, "mpi", "tenant", 3, map[string]any{"distributed": true}, map[string]any{"image": "mpi:1"}),
			environmentResponse(testOtherEnvironmentID, "mpi-2", "tenant", 4, map[string]any{"distributed": true}, map[string]any{"image": "mpi:2"}),
		}})
	})

	out, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{
		WorkloadType: "Distributed", DistributedFramework: "MPI", Limit: 1,
	})
	require.NoError(t, err)
	require.Len(t, out.Environments, 1)
	assert.Equal(t, int32(2), out.TotalCount)
	assert.True(t, out.HasMore)
	assert.Equal(t, int64(3), *out.Environments[0].ExpectedUpdateCount)
	assert.Equal(t, "mpi:1", out.Environments[0].UpdateRequest["spec"].(map[string]any)["image"])
}

func TestListEnvironmentsFiltersExactlyAndKeepsSystemReadOnly(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			environmentResponse(testEnvironmentID, "python-copy", "system", 1, map[string]any{"workspace": true}, map[string]any{}),
			environmentResponse(testOtherEnvironmentID, "python", "tenant", 2, map[string]any{"workspace": true}, map[string]any{}),
		}})
	})

	out, err := listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{Scope: "tenant"})
	require.NoError(t, err)
	require.Len(t, out.Environments, 2)
	assert.Contains(t, out.Environments[0].UpdateBlockedReason, "read-only")

	out, err = listEnvironmentsRun(context.Background(), client, listEnvironmentsInput{Name: "python"})
	require.NoError(t, err)
	require.Len(t, out.Environments, 1)
	assert.Equal(t, testOtherEnvironmentID, out.Environments[0].EnvironmentID)
}

func TestManageEnvironmentCreate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusAccepted,
			environmentResponse(testEnvironmentID, "pytorch", "tenant", 1, map[string]any{"training": true}, map[string]any{"image": "pytorch:2"}))
	})

	out, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "create", Request: environmentCreateRequest("pytorch", "tenant", map[string]any{"training": true}, map[string]any{"image": "pytorch:2"}),
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, testEnvironmentID, out.EnvironmentID)
	assert.Equal(t, "pytorch:2", out.Environment["spec"].(map[string]any)["image"])
}

func TestManageEnvironmentRejectsMismatchedCreateResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusAccepted,
			environmentResponse(testEnvironmentID, "different", "tenant", 1, map[string]any{"training": true}, map[string]any{}))
	})

	_, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "create", Request: environmentCreateRequest("expected", "tenant", map[string]any{"training": true}, map[string]any{}),
	})
	require.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestManageEnvironmentUpdateRereadsAndReplaces(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	current := environmentResponse(testEnvironmentID, "python", "tenant", 3, map[string]any{"workspace": true}, map[string]any{"image": "python:1"})
	var gets atomic.Int32
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			gets.Add(1)
			runaiclienttest.WriteJSON(w, http.StatusOK, current)
		case http.MethodPut:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "python:2", body["spec"].(map[string]any)["image"])
			runaiclienttest.WriteJSON(w, http.StatusAccepted,
				environmentResponse(testEnvironmentID, "python", "tenant", 4, map[string]any{"workspace": true}, map[string]any{"image": "python:2"}))
		}
	})

	updateCount := int64(3)
	out, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "update", EnvironmentID: testEnvironmentID, ExpectedUpdateCount: &updateCount, Confirm: true,
		Request: map[string]any{"meta": map[string]any{"name": "python"}, "spec": map[string]any{"image": "python:2"}},
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, int32(2), gets.Load())
	assert.Equal(t, "python:2", out.Environment["spec"].(map[string]any)["image"])
}

func TestManageEnvironmentUpdatePreservesCredentialReferences(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	credentialIDs := []string{
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		"33333333-3333-4333-8333-333333333333",
		"44444444-4444-4444-8444-444444444444",
		"55555555-5555-4555-8555-555555555555",
		"66666666-6666-4666-8666-666666666666",
		"77777777-7777-4777-8777-777777777777",
		"88888888-8888-4888-8888-888888888888",
	}
	variables := make([]any, 0, len(credentialIDs))
	for index, id := range credentialIDs {
		variables = append(variables, map[string]any{
			"name":       fmt.Sprintf("CREDENTIAL_%d", index),
			"credential": map[string]any{"assetId": id, "key": "value"},
		})
	}
	spec := map[string]any{"image": "python:3", "environmentVariables": variables}
	current := environmentResponse(testEnvironmentID, "python", "tenant", 3, map[string]any{"workspace": true}, spec)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, current)
		case http.MethodPut:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, variables, body["spec"].(map[string]any)["environmentVariables"])
			runaiclienttest.WriteJSON(w, http.StatusAccepted, current)
		}
	})

	updateCount := int64(3)
	_, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "update", EnvironmentID: testEnvironmentID, ExpectedUpdateCount: &updateCount, Confirm: true,
		Request: map[string]any{"meta": map[string]any{"name": "python"}, "spec": spec},
	})
	require.NoError(t, err)
}

func TestManageEnvironmentRejectsUnknownFieldsFromCurrentEnvironment(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	response := environmentResponse(testEnvironmentID, "future", "tenant", 3, map[string]any{"workspace": true}, map[string]any{
		"image": "python:3", "futureField": true,
	})
	var puts atomic.Int32
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			puts.Add(1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, response)
	})

	updateCount := int64(3)
	_, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "update", EnvironmentID: testEnvironmentID, ExpectedUpdateCount: &updateCount, Confirm: true,
		Request: map[string]any{"meta": map[string]any{"name": "future"}, "spec": map[string]any{"image": "python:4"}},
	})
	require.ErrorIs(t, err, runaimcp.ErrConflict)
	assert.Zero(t, puts.Load())
}

func TestManageEnvironmentRejectsStaleUpdate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK,
			environmentResponse(testEnvironmentID, "python", "tenant", 4, map[string]any{"workspace": true}, map[string]any{}))
	})

	updateCount := int64(3)
	_, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "update", EnvironmentID: testEnvironmentID, ExpectedUpdateCount: &updateCount, Confirm: true,
		Request: map[string]any{"meta": map[string]any{"name": "python"}, "spec": map[string]any{}},
	})
	require.ErrorIs(t, err, runaimcp.ErrConflict)
}

func TestDeleteEnvironmentIgnoreMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "missing"})
	})
	out, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "delete", EnvironmentID: testEnvironmentID, IgnoreMissing: true, Confirm: true,
	})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.Contains(t, out.Warnings[0], "nothing to delete")
}

func TestDeleteEnvironmentSuccess(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/environment/"+testEnvironmentID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			runaiclienttest.WriteJSON(w, http.StatusOK,
				environmentResponse(testEnvironmentID, "delete-me", "tenant", 1, map[string]any{"workspace": true}, map[string]any{}))
			return
		}
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusAccepted)
	})

	out, err := manageEnvironmentRun(context.Background(), client, manageEnvironmentInput{
		Action: "delete", EnvironmentID: testEnvironmentID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Nil(t, out.Environment)
}

func environmentCreateRequest(name, scope string, workloadTypes, spec map[string]any) map[string]any {
	meta := map[string]any{"name": name, "scope": scope}
	if workloadTypes != nil {
		meta["workloadSupportedTypes"] = workloadTypes
	}
	return map[string]any{"meta": meta, "spec": spec}
}

func environmentResponse(id, name, scope string, updateCount int64, workloadTypes, spec map[string]any) map[string]any {
	return map[string]any{
		"meta": map[string]any{
			"id": id, "name": name, "scope": scope, "kind": "environment",
			"updateCount": updateCount, "workloadSupportedTypes": workloadTypes,
		},
		"spec": spec,
	}
}
