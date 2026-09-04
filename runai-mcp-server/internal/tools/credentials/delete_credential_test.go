/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestDeleteCredentialIgnoreMissingOnlySuppressesNotFound(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/asset/credentials/generic-secret/"+testCredentialID,
			func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "missing", http.StatusNotFound)
			})

		out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
			Action: "delete", CredentialType: kindGenericSecret,
			CredentialID: testCredentialID, IgnoreMissing: true, Confirm: true,
		})
		require.NoError(t, err)
		assert.False(t, out.Applied)
		require.Len(t, out.Warnings, 1)
	})

	t.Run("upstream_error", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/asset/credentials/generic-secret/"+testCredentialID,
			func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "failure", http.StatusInternalServerError)
			})

		_, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
			Action: "delete", CredentialType: kindGenericSecret,
			CredentialID: testCredentialID, IgnoreMissing: true, Confirm: true,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	})
}

func TestDeleteCredentialIgnoreMissingHandlesDeleteRace(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	path := "/api/v1/asset/credentials/generic-secret/" + testCredentialID
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
				testCredentialID, "raced", kindGenericSecret, false, "cluster", ""))
			return
		}
		http.Error(w, "already deleted", http.StatusNotFound)
	})

	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "delete", CredentialType: kindGenericSecret,
		CredentialID: testCredentialID, IgnoreMissing: true, Confirm: true,
	})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.Diff)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "nothing to delete")
}

func TestDeleteCredentialRequiresConfirmationBeforeMutation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	path := "/api/v1/asset/credentials/generic-secret/" + testCredentialID
	var deletes int32
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			atomic.AddInt32(&deletes, 1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, credentialResponse(
			testCredentialID, "mine", kindGenericSecret, true, "tenant", ""))
	})

	out, err := manageCredentialsRun(context.Background(), client, manageCredentialsInput{
		Action: "delete", CredentialType: kindGenericSecret,
		CredentialID: testCredentialID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Equal(t, int32(0), atomic.LoadInt32(&deletes))
	assert.NotEmpty(t, out.Before)
	assert.NotEmpty(t, out.Diff)
	assert.True(t, out.Private)
}

func TestDeleteCredentialAllSupportedTypesAndPrivacyModes(t *testing.T) {
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
			var deletes int
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					deletes++
					w.WriteHeader(http.StatusNoContent)
					return
				}
				scope := "tenant"
				response := credentialResponse(
					testCredentialID, tc.name, tc.kind, tc.personal, scope, "")
				if !tc.personal {
					meta := response["meta"].(map[string]any)
					meta["scope"] = "cluster"
					meta["clusterId"] = testClusterID
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, response)
			})

			in := manageCredentialsInput{
				Action: "delete", CredentialType: tc.kind,
				CredentialID: testCredentialID, Confirm: true,
			}
			require.NoError(t, in.Validate())
			out, err := manageCredentialsRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.True(t, out.Applied)
			assert.Equal(t, 1, deletes)
			assert.Equal(t, tc.personal, out.Private)
		})
	}
}
