/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func TestListCredentialsInputValidate(t *testing.T) {
	cases := []struct {
		name    string
		in      listCredentialsInput
		wantErr string
	}{
		{name: "all"},
		{name: "by_id", in: listCredentialsInput{CredentialID: testCredentialID}},
		{name: "personal", in: listCredentialsInput{Private: boolPtr(true)}},
		{name: "shared_project", in: listCredentialsInput{
			Private: boolPtr(false), Scope: "project", ClusterID: testClusterID,
		}},
		{name: "bad_type", in: listCredentialsInput{CredentialType: "token"}, wantErr: "credentialType"},
		{name: "personal_scope", in: listCredentialsInput{
			Private: boolPtr(true), Scope: "tenant",
		}, wantErr: "scope filters"},
		{name: "bad_cluster", in: listCredentialsInput{
			ClusterID: "not-a-uuid",
		}, wantErr: "UUID"},
		{name: "bad_id", in: listCredentialsInput{CredentialID: "not-a-uuid"}, wantErr: "UUID"},
		{name: "bad_limit", in: listCredentialsInput{Limit: 201}, wantErr: "limit"},
		{name: "bad_offset", in: listCredentialsInput{Offset: -1}, wantErr: "offset"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestListCredentialsReturnsPersonalCredentialMetadata(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/credentials", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "genericSecret", r.URL.Query().Get("type"))
		assert.Equal(t, "true", r.URL.Query().Get("privateOnly"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			credentialResponse(testCredentialID, "mine", kindGenericSecret, true, "tenant", "my token"),
		}})
	})

	out, err := listCredentialsRun(context.Background(), client, listCredentialsInput{
		CredentialType: kindGenericSecret,
		Private:        boolPtr(true),
	})
	require.NoError(t, err)
	require.Len(t, out.Credentials, 1)
	assert.Equal(t, testCredentialID, out.Credentials[0].CredentialID)
	assert.Equal(t, "mine", out.Credentials[0].Name)
	assert.True(t, out.Credentials[0].Private)
	require.NotNil(t, out.Credentials[0].Description)
	assert.Equal(t, "my token", *out.Credentials[0].Description)
	assert.Equal(t, int32(1), out.TotalCount)
}

func TestListCredentialsFiltersAndPaginatesInAPIOrder(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/credentials", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{
			credentialResponse(testThirdCredentialID, "c", kindAccessKey, false, "tenant", ""),
			credentialResponse(testCredentialID, "a", kindGenericSecret, false, "cluster", ""),
			credentialResponse(testOtherCredentialID, "b", kindPassword, false, "tenant", ""),
		}})
	})

	byID, err := listCredentialsRun(context.Background(), client, listCredentialsInput{
		CredentialID: testOtherCredentialID,
	})
	require.NoError(t, err)
	require.Len(t, byID.Credentials, 1)
	assert.Equal(t, testOtherCredentialID, byID.Credentials[0].CredentialID)
	assert.Equal(t, int32(1), byID.TotalCount)

	first, err := listCredentialsRun(context.Background(), client, listCredentialsInput{Limit: 2})
	require.NoError(t, err)
	require.Len(t, first.Credentials, 2)
	assert.Equal(t, "c", first.Credentials[0].Name)
	assert.Equal(t, "a", first.Credentials[1].Name)
	assert.Equal(t, int32(3), first.TotalCount)
	assert.True(t, first.HasMore)
	assert.Equal(t, int32(2), first.NextOffset)

	second, err := listCredentialsRun(context.Background(), client, listCredentialsInput{
		Limit: 2, Offset: first.NextOffset,
	})
	require.NoError(t, err)
	require.Len(t, second.Credentials, 1)
	assert.Equal(t, "b", second.Credentials[0].Name)
	assert.False(t, second.HasMore)
	assert.Zero(t, second.NextOffset)
}

func TestListCredentialsForwardsExactNameAndScopeFilters(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	projectID := int32(42)
	mux.HandleFunc("/api/v1/asset/credentials", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "shared-registry", query.Get("name"))
		assert.Equal(t, kindDockerRegistry, query.Get("type"))
		assert.Equal(t, "project", query.Get("scope"))
		assert.Equal(t, testClusterID, query.Get("clusterId"))
		assert.Equal(t, "department-a", query.Get("departmentId"))
		assert.Equal(t, "42", query.Get("projectId"))
		assert.Equal(t, "false", query.Get("includePrivate"))
		assert.Empty(t, query.Get("privateOnly"))

		response := credentialResponse(
			testCredentialID, "shared-registry", kindDockerRegistry, false, "project", "")
		meta := response["meta"].(map[string]any)
		meta["clusterId"] = testClusterID
		meta["departmentId"] = "department-a"
		meta["projectId"] = projectID
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{response}})
	})

	in := listCredentialsInput{
		Name: "shared-registry", CredentialType: kindDockerRegistry,
		Private: boolPtr(false), Scope: "project", ClusterID: testClusterID,
		DepartmentID: "department-a", ProjectID: &projectID,
	}
	require.NoError(t, in.Validate())
	out, err := listCredentialsRun(context.Background(), client, in)
	require.NoError(t, err)
	require.Len(t, out.Credentials, 1)
	assert.Equal(t, "shared-registry", out.Credentials[0].Name)
	assert.False(t, out.Credentials[0].Private)
	require.NotNil(t, out.Credentials[0].ProjectID)
	assert.Equal(t, projectID, *out.Credentials[0].ProjectID)
}

func TestListCredentialsNilResponseIsUpstreamError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/credentials", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := listCredentialsRun(context.Background(), client, listCredentialsInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestListCredentialsEmptyArray(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/asset/credentials", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{}})
	})

	out, err := listCredentialsRun(context.Background(), client, listCredentialsInput{})
	require.NoError(t, err)
	assert.NotNil(t, out.Credentials)
	assert.Empty(t, out.Credentials)
}
