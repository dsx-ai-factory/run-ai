package templates

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const (
	testTemplateID = "6fced230-c208-4b73-98a8-964fc1a32787"
	testClusterID  = "f49990e7-52ba-4519-a760-e241ef6b01b1"
)

func templateResponse(workloadType, name string) map[string]any {
	return templateResponseWithScope(workloadType, name, "tenant")
}

func templateResponseWithScope(workloadType, name, scopeType string) map[string]any {
	scopeID := "1001"
	if scopeType == "system" {
		scopeID = ""
	}
	return map[string]any{"meta": map[string]any{
		"id": testTemplateID, "name": name, "description": "description", "workloadType": workloadType,
		"scopeType": scopeType, "scopeId": scopeID, "createdBy": "test@example.com",
		"createdAt": "2026-08-26T10:00:00Z",
	}}
}

func TestManageTemplateCreateRoutesEveryWorkloadType(t *testing.T) {
	tests := []struct {
		workloadType string
		path         string
	}{
		{templateTypeTraining, "/api/v1/workload-templates/trainings"},
		{templateTypeWorkspace, "/api/v1/workload-templates/workspaces"},
		{templateTypeDistributed, "/api/v1/workload-templates/distributed"},
		{templateTypeInference, "/api/v1/workload-templates/inferences"},
		{templateTypeDistributedInference, "/api/v1/workload-templates/distributed-inferences"},
	}

	for _, tt := range tests {
		t.Run(tt.workloadType, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(tt.path, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "base", body["meta"].(map[string]any)["name"])
				runaiclienttest.WriteJSON(w, http.StatusAccepted, templateResponse(tt.workloadType, "base"))
			})

			out, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
				Action: "create", WorkloadType: tt.workloadType,
				Request: map[string]any{"meta": map[string]any{"name": "base", "scopeType": "tenant"}},
			})
			require.NoError(t, err)
			assert.True(t, out.Applied)
			assert.Equal(t, testTemplateID, out.TemplateID)
		})
	}
}

func TestManageTemplateCreateHandlesAPIContractEdges(t *testing.T) {
	t.Run("nil success response", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/workload-templates/trainings", func(w http.ResponseWriter, _ *http.Request) {
			runaiclienttest.WriteJSON(w, http.StatusOK, nil)
		})

		_, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
			Action: "create", WorkloadType: templateTypeTraining,
			Request: map[string]any{"meta": map[string]any{"name": "base", "scopeType": "tenant"}},
		})
		assert.ErrorIs(t, err, runaimcp.ErrUpstream)
		assert.ErrorContains(t, err, "nil response")
	})

	t.Run("mapped API error keeps body", func(t *testing.T) {
		client, mux := runaiclienttest.NewServer(t)
		mux.HandleFunc("/api/v1/workload-templates/trainings", func(w http.ResponseWriter, _ *http.Request) {
			runaiclienttest.WriteJSON(w, http.StatusConflict, "template already exists")
		})

		_, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
			Action: "create", WorkloadType: templateTypeTraining,
			Request: map[string]any{"meta": map[string]any{"name": "base", "scopeType": "tenant"}},
		})
		assert.ErrorIs(t, err, runaimcp.ErrConflict)
		assert.ErrorContains(t, err, "template already exists")
	})
}

func TestManageTemplateUpdateConfirmationAndPut(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, templateResponse(templateTypeTraining, "base"))
	})
	putCalls := 0
	mux.HandleFunc("/api/v1/workload-templates/trainings/"+testTemplateID, func(w http.ResponseWriter, r *http.Request) {
		putCalls++
		assert.Equal(t, http.MethodPut, r.Method)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "base-v2", body["meta"].(map[string]any)["name"])
		runaiclienttest.WriteJSON(w, http.StatusOK, templateResponse(templateTypeTraining, "base-v2"))
	})
	input := manageTemplateInput{
		Action: "update", WorkloadType: templateTypeTraining, TemplateID: testTemplateID,
		Request: map[string]any{"meta": map[string]any{"name": "base-v2"}},
	}

	out, err := manageTemplateRun(context.Background(), client, input)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.ErrorContains(t, err, "base-v2")
	assert.Equal(t, "base", out.Name)
	assert.Zero(t, putCalls)

	input.Confirm = true
	out, err = manageTemplateRun(context.Background(), client, input)
	require.NoError(t, err)
	assert.Equal(t, 1, putCalls)
	assert.True(t, out.Applied)
	assert.Equal(t, "base-v2", out.Name)
}

func TestManageTemplateSystemUpdateFailsBeforeConfirmation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK,
			templateResponseWithScope(templateTypeWorkspace, "system-base", "system"))
	})
	mutationCalled := false
	mux.HandleFunc("/api/v1/workload-templates/workspaces/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		mutationCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
		Action: "update", WorkloadType: templateTypeWorkspace, TemplateID: testTemplateID,
		Request: map[string]any{"meta": map[string]any{"name": "renamed"}},
	})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	assert.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.False(t, mutationCalled)
}

func TestManageTemplateRejectsWrongWorkloadTypeBeforeMutation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, templateResponse(templateTypeTraining, "base"))
	})
	mutationCalled := false
	mux.HandleFunc("/api/v1/workload-templates/workspaces/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
		mutationCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
		Action: "delete", WorkloadType: templateTypeWorkspace, TemplateID: testTemplateID, Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.ErrorContains(t, err, "has workloadType Training, not Workspace")
	assert.False(t, mutationCalled)
}

func TestManageTemplateDeleteConfirmationWarnings(t *testing.T) {
	tests := []struct {
		name       string
		scope      string
		want       string
		notContain string
	}{
		{name: "tenant is permanent", scope: "tenant", want: "deleting a template is permanent"},
		{name: "system is tenant-local", scope: "system", want: "hides the preset for the current tenant", notContain: "permanent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workload-templates/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK,
					templateResponseWithScope(templateTypeWorkspace, "base", tt.scope))
			})
			deleteCalled := false
			mux.HandleFunc("/api/v1/workload-templates/workspaces/"+testTemplateID, func(w http.ResponseWriter, _ *http.Request) {
				deleteCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			_, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
				Action: "delete", WorkloadType: templateTypeWorkspace, TemplateID: testTemplateID,
			})
			assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
			assert.ErrorContains(t, err, tt.want)
			if tt.notContain != "" {
				assert.NotContains(t, err.Error(), tt.notContain)
			}
			assert.False(t, deleteCalled)
		})
	}
}

func TestManageTemplateDeleteUsesCanonicalAPIID(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	requestedID := strings.ToUpper(testTemplateID)
	mux.HandleFunc("/api/v1/workload-templates/"+requestedID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, templateResponse(templateTypeTraining, "base"))
	})
	mux.HandleFunc("/api/v1/workload-templates/trainings/"+testTemplateID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	out, err := manageTemplateRun(context.Background(), client, manageTemplateInput{
		Action: "delete", WorkloadType: templateTypeTraining, TemplateID: requestedID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, testTemplateID, out.TemplateID)
}

func TestManageTemplateInputValidation(t *testing.T) {
	validCreate := map[string]any{
		"meta": map[string]any{"name": "base", "scopeType": "cluster", "scopeId": testClusterID},
		"spec": map[string]any{"image": "nvcr.io/nvidia/pytorch:latest"},
	}
	validUpdate := map[string]any{
		"meta": map[string]any{"name": "base-v2"},
		"spec": map[string]any{"image": "nvcr.io/nvidia/pytorch:latest"},
	}
	tests := []struct {
		name    string
		input   manageTemplateInput
		wantErr string
	}{
		{name: "valid create", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: validCreate}},
		{name: "valid update", input: manageTemplateInput{Action: "update", WorkloadType: templateTypeTraining, TemplateID: testTemplateID, Request: validUpdate}},
		{name: "valid delete", input: manageTemplateInput{Action: "delete", WorkloadType: templateTypeTraining, TemplateID: testTemplateID}},
		{name: "unsupported action", input: manageTemplateInput{Action: "patch", WorkloadType: templateTypeTraining}, wantErr: "action must be one of"},
		{name: "unsupported workload type", input: manageTemplateInput{Action: "create", WorkloadType: "Job"}, wantErr: "workloadType must be one of"},
		{name: "create rejects id", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, TemplateID: testTemplateID, Request: validCreate}, wantErr: "must not be set for create"},
		{name: "update requires id", input: manageTemplateInput{Action: "update", WorkloadType: templateTypeTraining, Request: validUpdate}, wantErr: "templateId is required"},
		{name: "delete rejects request", input: manageTemplateInput{Action: "delete", WorkloadType: templateTypeTraining, TemplateID: testTemplateID, Request: validUpdate}, wantErr: "must not be set for delete"},
		{name: "system create unsupported", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: map[string]any{
			"meta": map[string]any{"name": "base", "scopeType": "system"},
		}}, wantErr: "system is not supported"},
		{name: "project id must fit API type", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: map[string]any{
			"meta": map[string]any{"name": "base", "scopeType": "project", "scopeId": "2147483648"},
		}}, wantErr: "positive 32-bit project ID"},
		{name: "description lower boundary", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: map[string]any{
			"meta": map[string]any{"name": "base", "description": "", "scopeType": "tenant"},
		}}, wantErr: "between 1 and 250 characters"},
		{name: "description upper boundary", input: manageTemplateInput{Action: "update", WorkloadType: templateTypeTraining, TemplateID: testTemplateID, Request: map[string]any{
			"meta": map[string]any{"name": "base", "description": strings.Repeat("x", 251)},
		}}, wantErr: "between 1 and 250 characters"},
		{name: "asset id must be UUID", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeWorkspace, Request: map[string]any{
			"meta": map[string]any{"name": "base", "scopeType": "tenant"}, "assets": map[string]any{"environment": "invalid"},
		}}, wantErr: "request.assets.environment must be a UUID"},
		{name: "nested datasource id must be UUID", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: map[string]any{
			"meta": map[string]any{"name": "base", "scopeType": "tenant"}, "assets": map[string]any{"datasources": []any{testTemplateID, "invalid"}},
		}}, wantErr: "request.assets.datasources[1] must be a UUID"},
		{name: "distributed master conflict", input: manageTemplateInput{Action: "update", WorkloadType: templateTypeDistributed, TemplateID: testTemplateID, Request: map[string]any{
			"meta": map[string]any{"name": "base"}, "masterSpecSameAsWorker": true, "master": map[string]any{},
		}}, wantErr: "master must not be set"},
		{name: "typed API rejects unknown nested field", input: manageTemplateInput{Action: "create", WorkloadType: templateTypeTraining, Request: map[string]any{
			"meta": map[string]any{"name": "base", "scopeType": "tenant"}, "spec": map[string]any{"imageTypo": "x"},
		}}, wantErr: "imageTypo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
