/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
)

func TestCreateCredentialPersonalGenericSecret(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotBody map[string]any
	mux.HandleFunc("/api/v1/asset/credentials/generic-secret", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		runaiclienttest.WriteJSON(w, http.StatusAccepted, credentialResponse(
			testCredentialID, "mine", kindGenericSecret, true, "tenant", "personal secret"))
	})

	description := "personal secret"
	in := manageCredentialsInput{
		Action: "create", Private: true, CredentialType: kindGenericSecret,
		Name: "mine", Description: &description,
		Spec: &credentialSpec{KeyValuePairs: []credentialKeyValue{{Key: "TOKEN", Value: "super-secret"}}},
	}
	out, err := manageCredentialsRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, testCredentialID, out.CredentialID)
	assert.NotContains(t, out.After, "spec")
	assert.NotContains(t, out.After, "secretFields")
	assert.NotContains(t, out.After, "secretFieldCount")

	meta := gotBody["meta"].(map[string]any)
	assert.Equal(t, true, meta["private"])
	assert.Equal(t, "tenant", meta["scope"])
	bodyJSON, err := json.Marshal(gotBody)
	require.NoError(t, err)
	assert.Contains(t, string(bodyJSON), "super-secret")
}

func TestCreateCredentialDepartmentSendsOnlyDepartmentID(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotBody map[string]any
	mux.HandleFunc("/api/v1/asset/credentials/access-key", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		response := credentialResponse(testCredentialID, "s3", kindAccessKey, false, "department", "")
		meta := response["meta"].(map[string]any)
		meta["clusterId"] = testClusterID
		meta["departmentId"] = "department-a"
		runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
	})

	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "create", CredentialType: kindAccessKey,
		Name: "s3", Scope: "department", DepartmentID: "department-a",
		Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	meta := gotBody["meta"].(map[string]any)
	assert.NotContains(t, meta, "clusterId")
	assert.Equal(t, "department-a", meta["departmentId"])
}

func TestCreateCredentialAllSupportedTypesAndPrivacyModes(t *testing.T) {
	cases := []struct {
		name       string
		kind       string
		path       string
		personal   bool
		spec       *credentialSpec
		wantFields []string
	}{
		{"shared_access_key", kindAccessKey, "access-key", false,
			&credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			[]string{"accessKeyId", "secretAccessKey"}},
		{"shared_password", kindPassword, "password", false,
			&credentialSpec{User: "user", Password: "secret"},
			[]string{"user", "password"}},
		{"shared_docker_registry", kindDockerRegistry, "docker-registry", false,
			&credentialSpec{User: "user", Password: "secret", URL: "registry.example.com"},
			[]string{"user", "password", "url"}},
		{"shared_generic_secret", kindGenericSecret, "generic-secret", false,
			&credentialSpec{KeyValuePairs: []credentialKeyValue{{Key: "TOKEN", Value: "secret"}}},
			[]string{"keyValuePairs"}},
		{"shared_ngc_api_key", kindNGCAPIKey, "ngc-api-key", false,
			&credentialSpec{NGCAPIKey: "secret"},
			[]string{"ngc_api_key"}},
		{"personal_docker_registry", kindDockerRegistry, "docker-registry", true,
			&credentialSpec{User: "user", Password: "secret", URL: "registry.example.com"},
			[]string{"user", "password", "url"}},
		{"personal_generic_secret", kindGenericSecret, "generic-secret", true,
			&credentialSpec{KeyValuePairs: []credentialKeyValue{{Key: "TOKEN", Value: "secret"}}},
			[]string{"keyValuePairs"}},
		{"personal_ngc_api_key", kindNGCAPIKey, "ngc-api-key", true,
			&credentialSpec{NGCAPIKey: "secret"},
			[]string{"ngc_api_key"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			var body map[string]any
			mux.HandleFunc("/api/v1/asset/credentials/"+tc.path, func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				scope := "tenant"
				response := credentialResponse(testCredentialID, tc.name, tc.kind, tc.personal, scope, "")
				if !tc.personal {
					scope = "cluster"
					meta := response["meta"].(map[string]any)
					meta["scope"] = scope
					meta["clusterId"] = testClusterID
				}
				runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
			})

			in := manageCredentialsInput{
				Action: "create", CredentialType: tc.kind, Name: tc.name, Scope: "cluster",
				ClusterID: testClusterID, Spec: tc.spec,
			}
			if tc.personal {
				in.Private = true
				in.Scope = ""
				in.ClusterID = ""
			}
			require.NoError(t, in.Validate())
			out, err := manageCredentialsRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.True(t, out.Applied)
			assert.Equal(t, tc.personal, out.Private)
			meta := body["meta"].(map[string]any)
			assert.Equal(t, tc.personal, meta["private"])
			spec := body["spec"].(map[string]any)
			for _, field := range tc.wantFields {
				assert.Contains(t, spec, field)
			}
		})
	}
}

func TestCreateCredentialImportsExistingSecretForSupportedTypes(t *testing.T) {
	tests := []struct {
		kind string
		path string
	}{
		{kindAccessKey, "access-key"},
		{kindPassword, "password"},
		{kindDockerRegistry, "docker-registry"},
		{kindGenericSecret, "generic-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/asset/credentials/"+tt.path, func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, map[string]any{"existingSecretName": "existing-secret"}, body["spec"])
				runaiclienttest.WriteJSON(w, http.StatusCreated, credentialResponse(
					testCredentialID, "imported", tt.kind, false, "cluster", ""))
			})

			in := manageCredentialsInput{
				Action: "create", CredentialType: tt.kind, Name: "imported",
				Scope: "cluster", ClusterID: testClusterID,
				Spec: &credentialSpec{ExistingSecretName: "existing-secret"},
			}
			require.NoError(t, in.Validate())
			out, err := manageCredentialsRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.True(t, out.Applied)
			assert.Equal(t, testCredentialID, out.CredentialID)
		})
	}
}
