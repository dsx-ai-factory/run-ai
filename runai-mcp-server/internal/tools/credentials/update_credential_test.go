/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func TestUpdateCredentialSharedDescription(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	path := "/api/v1/asset/credentials/access-key/" + testCredentialID
	var putBody map[string]any
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
				testCredentialID, "s3", kindAccessKey, false, "tenant", "old"))
		case http.MethodPut:
			require.NoError(t, json.NewDecoder(r.Body).Decode(&putBody))
			runaiclienttest.WriteJSON(w, http.StatusAccepted, credentialResponse(
				testCredentialID, "s3", kindAccessKey, false, "tenant", "new"))
		}
	})

	description := "new"
	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "update", CredentialType: kindAccessKey,
		CredentialID: testCredentialID, Description: &description, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.False(t, out.Private)
	assert.Equal(t, "old", out.Before["description"])
	assert.Equal(t, "new", out.After["description"])
	assert.Contains(t, out.Diff, "description")
	assert.Equal(t, "new", putBody["meta"].(map[string]any)["description"])
}

func TestUpdateCredentialRequiresConfirmationBeforeMutation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	path := "/api/v1/asset/credentials/generic-secret/" + testCredentialID
	var puts int32
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			atomic.AddInt32(&puts, 1)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
			testCredentialID, "mine", kindGenericSecret, true, "tenant", "old"))
	})

	description := "new"
	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "update", CredentialType: kindGenericSecret,
		CredentialID: testCredentialID, Description: &description,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Equal(t, int32(0), atomic.LoadInt32(&puts))
	assert.Equal(t, "old", out.Before["description"])
	assert.Equal(t, "new", out.After["description"])
	assert.Contains(t, out.Diff, "description")
}

func TestUpdateCredentialClearsDescription(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	path := "/api/v1/asset/credentials/password/" + testCredentialID
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
				testCredentialID, "registry-login", kindPassword, false, "tenant", "old"))
			return
		}

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "", body["meta"].(map[string]any)["description"])
		runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
			testCredentialID, "registry-login", kindPassword, false, "tenant", ""))
	})

	empty := ""
	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "update", CredentialType: kindPassword,
		CredentialID: testCredentialID, Description: &empty, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "old", out.Before["description"])
	assert.NotContains(t, out.After, "description")
	assert.Contains(t, out.Diff, "description")
}

func TestUpdateCredentialAllSupportedTypesAndPrivacyModes(t *testing.T) {
	cases := []struct {
		name     string
		kind     string
		path     string
		personal bool
	}{
		{"shared_access_key", kindAccessKey, "access-key", false},
		{"shared_password", kindPassword, "password", false},
		{"shared_docker_registry", kindDockerRegistry, "docker-registry", false},
		{"shared_generic_secret", kindGenericSecret, "generic-secret", false},
		{"shared_ngc_api_key", kindNGCAPIKey, "ngc-api-key", false},
		{"personal_docker_registry", kindDockerRegistry, "docker-registry", true},
		{"personal_generic_secret", kindGenericSecret, "generic-secret", true},
		{"personal_ngc_api_key", kindNGCAPIKey, "ngc-api-key", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			path := "/api/v1/asset/credentials/" + tc.path + "/" + testCredentialID
			var puts int
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				scope := "tenant"
				response := credentialResponse(
					testCredentialID, tc.name, tc.kind, tc.personal, scope, "old")
				if !tc.personal {
					meta := response["meta"].(map[string]any)
					meta["scope"] = "cluster"
					meta["clusterId"] = testClusterID
				}
				if r.Method == http.MethodPut {
					puts++
					response["meta"].(map[string]any)["description"] = "new"
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, response)
			})

			description := "new"
			in := manageCredentialsInput{
				Action: "update", CredentialType: tc.kind,
				CredentialID: testCredentialID, Description: &description, Confirm: true,
			}
			require.NoError(t, in.Validate())
			out, err := manageCredentialsRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.True(t, out.Applied)
			assert.Equal(t, 1, puts)
			assert.Equal(t, "new", out.After["description"])
			assert.Equal(t, tc.personal, out.Private)
		})
	}
}
