/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aiapplications

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	testClusterID     = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	testProjectID     = "4500002"
	testApplicationID = "4bfbe3a8-6659-44a3-914f-b82952b0d99b"
)

func validSubmitInput() submitAIApplicationInput {
	return submitAIApplicationInput{
		ClusterID: testClusterID,
		ProjectID: testProjectID,
		Name:      "nemo-app",
		Spec: aiApplicationSpec{ChartRef: aiApplicationChartRef{
			ChartURL: "https://charts.example.com/nemo-1.0.0.tgz",
		}},
	}
}

func TestSubmitAIApplicationInputValidate(t *testing.T) {
	encodedValues := "cmVwbGljYUNvdW50OiAyCg=="
	tests := []struct {
		name    string
		mutate  func(*submitAIApplicationInput)
		wantErr string
	}{
		{name: "chart URL"},
		{name: "HTTP repository", mutate: func(in *submitAIApplicationInput) {
			in.Spec.ChartRef = aiApplicationChartRef{Repo: "http://charts.example.com", ChartName: "nemo", ChartVersion: "1.0.0"}
		}},
		{name: "OCI repository", mutate: func(in *submitAIApplicationInput) {
			in.Spec.ChartRef = aiApplicationChartRef{Repo: "oci://registry.example.com/charts", ChartName: "nemo", ChartVersion: "1.0.0"}
		}},
		{name: "chart ID", mutate: func(in *submitAIApplicationInput) {
			in.Spec.ChartRef = aiApplicationChartRef{ChartID: testApplicationID}
		}},
		{name: "values base64", mutate: func(in *submitAIApplicationInput) { in.Spec.ValuesBase64 = &encodedValues }},
		{name: "shared chart credential", mutate: func(in *submitAIApplicationInput) {
			in.Spec.Auth = &aiApplicationAuth{AssetID: testApplicationID}
		}},
		{name: "personal chart credential", mutate: func(in *submitAIApplicationInput) {
			in.Spec.Auth = &aiApplicationAuth{AssetID: testApplicationID, UserCredential: true}
		}},
		{name: "missing source", mutate: func(in *submitAIApplicationInput) { in.Spec.ChartRef = aiApplicationChartRef{} }, wantErr: "exactly one source"},
		{name: "multiple sources", mutate: func(in *submitAIApplicationInput) { in.Spec.ChartRef.ChartID = testApplicationID }, wantErr: "exactly one source"},
		{name: "partial repository", mutate: func(in *submitAIApplicationInput) {
			in.Spec.ChartRef = aiApplicationChartRef{Repo: "https://charts.example.com"}
		}, wantErr: "must all be set"},
		{name: "values conflict", mutate: func(in *submitAIApplicationInput) {
			in.Spec.Values = map[string]any{"replicas": 2}
			in.Spec.ValuesBase64 = &encodedValues
		}, wantErr: "mutually exclusive"},
		{name: "invalid base64", mutate: func(in *submitAIApplicationInput) { invalid := "%%%"; in.Spec.ValuesBase64 = &invalid }, wantErr: "valid base64"},
		{name: "missing auth asset", mutate: func(in *submitAIApplicationInput) {
			in.Spec.Auth = &aiApplicationAuth{UserCredential: true}
		}, wantErr: "assetId is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validSubmitInput()
			if tt.mutate != nil {
				tt.mutate(&in)
			}
			err := in.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestListAIApplications(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "25", r.URL.Query().Get("limit"))
		assert.Equal(t, "10", r.URL.Query().Get("offset"))
		assert.Equal(t, "clusterId=="+testClusterID+",projectId=="+testProjectID+",name==nemo-app", r.URL.Query().Get("filterBy"))
		assert.Equal(t, "createdAt", r.URL.Query().Get("sortBy"))
		assert.Equal(t, "desc", r.URL.Query().Get("sortOrder"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"next": 35,
			"applications": []map[string]any{{
				"id": testApplicationID, "name": "nemo-app", "phase": "Running",
				"actionStatus": "deployed", "actionMessage": "application is ready",
				"projectId": testProjectID, "projectName": "team-a", "createdAt": "2026-08-20T09:00:00Z",
			}},
		})
	})

	out, err := listAIApplicationsRun(context.Background(), client, listAIApplicationsInput{
		ClusterID: testClusterID, ProjectID: testProjectID, Name: "nemo-app",
		SortBy: "createdAt", SortOrder: "desc", Limit: 25, Offset: 10,
	})
	require.NoError(t, err)
	require.Len(t, out.Applications, 1)
	assert.Equal(t, testApplicationID, out.Applications[0].ID)
	assert.Equal(t, "Running", out.Applications[0].Phase)
	assert.Equal(t, "deployed", out.Applications[0].ActionStatus)
	assert.Equal(t, "application is ready", out.Applications[0].ActionMessage)
	assert.Equal(t, "2026-08-20T09:00:00Z", out.Applications[0].CreatedAt)
	assert.True(t, out.HasMore)
	assert.Equal(t, int32(35), out.NextOffset)
}

func TestListAIApplicationsReturnsEmptyArray(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"applications": []any{}})
	})

	out, err := listAIApplicationsRun(context.Background(), client, listAIApplicationsInput{})
	require.NoError(t, err)
	assert.NotNil(t, out.Applications)
	assert.Empty(t, out.Applications)
}

func TestListAIApplicationsInputUsesAPILimit(t *testing.T) {
	require.NoError(t, (listAIApplicationsInput{Limit: 500}).Validate())
	assert.ErrorContains(t, (listAIApplicationsInput{Limit: 501}).Validate(), "between 0 and 500")
}

func TestSubmitAIApplicationCreatesAfterConfirmation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "nemo-app", body["name"])
		assert.Equal(t, testClusterID, body["clusterId"])
		assert.Equal(t, testProjectID, body["projectId"])
		spec := body["spec"].(map[string]any)
		auth := spec["auth"].(map[string]any)
		assert.Equal(t, true, auth["userCredential"])
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": testApplicationID, "name": "nemo-app", "projectId": testProjectID,
			"desiredPhase": "deployed", "phase": "Pending", "actionStatus": "creating",
		})
	})

	in := validSubmitInput()
	in.Confirm = true
	in.Spec.Values = map[string]any{"password": "do-not-return"}
	in.Spec.Auth = &aiApplicationAuth{AssetID: testApplicationID, UserCredential: true}
	out, err := submitAIApplicationRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.Equal(t, testApplicationID, out.ID)
	assert.Equal(t, "Pending", out.Phase)
	assert.Equal(t, "creating", out.ActionStatus)
	safeOutput, marshalErr := json.Marshal(out)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(safeOutput), "do-not-return")
}

func TestSubmitAIApplicationRequiresConfirmationBeforeCreate(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	createCalled := false
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, _ *http.Request) {
		createCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := submitAIApplicationRun(context.Background(), client, validSubmitInput())
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.False(t, createCalled)
}

func TestSubmitAIApplicationAcceptsNoContentSuccess(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	in := validSubmitInput()
	in.Confirm = true
	out, err := submitAIApplicationRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.Empty(t, out.ID)
}

func TestSubmitAIApplicationTreatsServerErrorAsUnknownOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/ai-applications", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "application created but failed to retrieve it"})
	})

	in := validSubmitInput()
	in.Confirm = true
	_, err := submitAIApplicationRun(context.Background(), client, in)
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "list_ai_applications before retrying")
}

func TestAIApplicationConfirmationDetailsOmitSensitiveInput(t *testing.T) {
	in := validSubmitInput()
	in.Spec.ChartRef.ChartURL = "https://user:password@charts.example.com/chart.tgz?token=secret"
	in.Spec.Values = map[string]any{"password": "helm-secret"}
	in.Spec.Auth = &aiApplicationAuth{AssetID: testApplicationID}

	details, err := json.Marshal(aiApplicationConfirmationDetails(in))
	require.NoError(t, err)
	assert.JSONEq(t, `{"chartSource":"url","clusterId":"`+testClusterID+`","projectId":"`+testProjectID+`"}`, string(details))
}
