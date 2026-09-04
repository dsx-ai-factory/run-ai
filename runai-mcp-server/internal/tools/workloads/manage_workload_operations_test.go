/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestManageWorkloadLifecycleInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      manageWorkloadLifecycleInput
		wantErr bool
	}{
		{name: "missing action", in: manageWorkloadLifecycleInput{WorkloadID: validUUID}, wantErr: true},
		{name: "unknown action", in: manageWorkloadLifecycleInput{Action: "pause", WorkloadID: validUUID}, wantErr: true},
		{name: "missing workload id", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionStop}, wantErr: true},
		{name: "invalid workload id", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionDelete, WorkloadID: "job-1"}, wantErr: true},
		{name: "stop", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionStop, WorkloadID: validUUID}},
		{name: "restart", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionRestart, WorkloadID: validUUID}},
		{name: "update missing request", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID}, wantErr: true},
		{name: "update missing spec", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: map[string]any{"specification": map[string]any{}}}, wantErr: true},
		{name: "update null spec", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: map[string]any{"spec": nil}}, wantErr: true},
		{name: "update empty spec", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: map[string]any{"spec": map[string]any{}}}, wantErr: true},
		{name: "update unknown top-level field", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: map[string]any{"spec": map[string]any{"image": "image"}, "unknown": true}}, wantErr: true},
		{name: "update", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest()}},
		{name: "delete", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionDelete, WorkloadID: validUUID}},
		{name: "request with stop", in: manageWorkloadLifecycleInput{Action: workloadLifecycleActionStop, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest()}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestManageWorkloadLifecycleRun_ActionMapping(t *testing.T) {
	tests := []struct {
		action      string
		batchAction string
	}{
		{action: workloadLifecycleActionStop, batchAction: "suspend"},
		{action: workloadLifecycleActionRestart, batchAction: "resume"},
		{action: workloadLifecycleActionDelete, batchAction: "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditions(mux)

			mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				var body struct {
					IDs    []string `json:"ids"`
					Action string   `json:"action"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, []string{validUUID}, body.IDs)
				assert.Equal(t, tt.batchAction, body.Action)
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"id": "batch-1", "succeeded": []string{validUUID}, "failed": []any{},
				})
			})

			out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: tt.action, WorkloadID: validUUID, Confirm: true,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.action, out.Action)
			assert.Equal(t, validUUID, out.WorkloadID)
			assert.True(t, out.Applied)
		})
	}
}

func TestLifecycleBatchActionRejectsUnknownAction(t *testing.T) {
	action, err := lifecycleBatchAction("pause")

	assert.Empty(t, action)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "unsupported workload lifecycle action")
}

func TestManageWorkloadLifecycleRun_RequiresConfirmation(t *testing.T) {
	for _, action := range []string{
		workloadLifecycleActionStop,
		workloadLifecycleActionRestart,
		workloadLifecycleActionUpdate,
		workloadLifecycleActionDelete,
	} {
		t.Run(action, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			if action == workloadLifecycleActionUpdate {
				registerLifecyclePreconditionsForType(mux, workloadTypeInference)
			} else {
				registerLifecyclePreconditions(mux)
			}
			mutated := false
			mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
				mutated = true
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "unexpected"})
			})

			out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: action, WorkloadID: validUUID, UpdateRequest: updateRequestForAction(action),
			})
			assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
			assert.False(t, mutated)
			assert.Equal(t, validUUID, out.WorkloadID)
			if action == workloadLifecycleActionDelete {
				assert.Contains(t, err.Error(), "permanent")
			}
			if action == workloadLifecycleActionUpdate {
				assert.Contains(t, err.Error(), "image")
				assert.NotContains(t, err.Error(), "nvcr.io/test:2")
			}
			assert.NotContains(t, err.Error(), "()")
		})
	}
}

func TestManageWorkloadLifecycleRun_UpdateConfirmationDoesNotExposeSpecValues(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeInference)
	request := map[string]any{"spec": map[string]any{
		"image": "private.registry/image:secret-tag",
		"environmentVariables": []any{
			map[string]any{"name": "API_TOKEN", "value": "super-secret-value"},
		},
	}}

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: request,
	})
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.ErrorContains(t, err, "environmentVariables")
	assert.ErrorContains(t, err, "image")
	assert.NotContains(t, err.Error(), "super-secret-value")
	assert.NotContains(t, err.Error(), "secret-tag")
}

func TestManageWorkloadLifecycleRun_UpdatesInference(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeInference)

	var gotBody map[string]any
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{"workloadId": validUUID, "name": "inference"})
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, workloadLifecycleActionUpdate, out.Action)
	assert.Equal(t, "nvcr.io/test:2", gotBody["spec"].(map[string]any)["image"])
}

func TestManageWorkloadLifecycleRun_UpdatesDistributedInference(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeDistribInf)

	var gotBody map[string]any
	mux.HandleFunc("/api/v1/workloads/distributed-inferences/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{"workloadId": validUUID, "name": "distributed-inference"})
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: distributedInferenceUpdateRequest(), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, workloadLifecycleActionUpdate, out.Action)
	assert.Equal(t, float64(3), gotBody["spec"].(map[string]any)["replicas"])
}

func TestManageWorkloadLifecycleRun_RejectsUpdateForNonInference(t *testing.T) {
	for _, workloadType := range []string{workloadTypeTraining, workloadTypeWorkspace, workloadTypeDistributed} {
		t.Run(workloadType, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditionsForType(mux, workloadType)

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
			})
			assert.ErrorIs(t, err, runaimcp.ErrValidation)
			assert.ErrorContains(t, err, "supported only for native Run:ai Inference and DistributedInference")
			assert.ErrorContains(t, err, workloadType)
		})
	}
}

func TestManageWorkloadLifecycleRun_UpdatesNativeInferenceBackedByAnyWorkload(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForTypeAndSource(mux, workloadTypeInference, "WorkloadsV2", true)
	mutated := false
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		mutated = true
		assert.Equal(t, http.MethodPatch, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusAccepted, map[string]any{"workloadId": validUUID, "name": "inference"})
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, mutated)
	assert.True(t, out.Applied)
}

func TestManageWorkloadLifecycleRun_RoutesWorkloadsV2RunStateActions(t *testing.T) {
	for _, workloadType := range []string{workloadTypeInference, workloadTypeDistribInf} {
		for _, action := range []string{workloadLifecycleActionStop, workloadLifecycleActionRestart} {
			t.Run(workloadType+"_"+action, func(t *testing.T) {
				client, mux := runaiclienttest.NewServer(t)
				registerLifecyclePreconditionsForTypeAndSource(mux, workloadType, "WorkloadsV2", true)
				mux.HandleFunc("/api/v2/workloads/"+validUUID+"/runState", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPatch, r.Method)
					var body struct {
						Value string `json:"value"`
					}
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					if action == workloadLifecycleActionStop {
						assert.Equal(t, "suspend", body.Value)
					} else {
						assert.Equal(t, "resume", body.Value)
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
						"id": validUUID, "desiredPhase": "Running", "metadata": map[string]any{}, "manifest": map[string]any{},
					})
				})

				out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
					Action: action, WorkloadID: validUUID, Confirm: true,
				})
				require.NoError(t, err)
				assert.True(t, out.Applied)
			})
		}
	}
}

func TestManageWorkloadLifecycleRun_RejectsUnsupportedRunStateBeforeConfirmationOrMutation(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForTypeAndSource(mux, workloadTypeInference, "WorkloadsV2", false)
	mutated := false
	mux.HandleFunc("/api/v2/workloads/"+validUUID+"/runState", func(w http.ResponseWriter, _ *http.Request) {
		mutated = true
		w.WriteHeader(http.StatusOK)
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID,
	})
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.ErrorContains(t, err, "not supported for Inference")
	assert.False(t, mutated)
}

func TestManageWorkloadLifecycleRun_MapsWorkloadsV2RunStateErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, wantError: runaimcp.ErrValidation},
		{name: "forbidden", statusCode: http.StatusForbidden, wantError: runaimcp.ErrForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantError: runaimcp.ErrNotFound},
		{name: "conflict", statusCode: http.StatusConflict, wantError: runaimcp.ErrConflict},
		{name: "server error", statusCode: http.StatusInternalServerError, wantError: runaimcp.ErrUpstream},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditionsForTypeAndSource(mux, workloadTypeInference, "WorkloadsV2", true)
			mux.HandleFunc("/api/v2/workloads/"+validUUID+"/runState", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, tt.statusCode, map[string]any{"message": tt.name})
			})

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
			})
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestManageWorkloadLifecycleRun_AcceptsSparseWorkloadsV2RunStateResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForTypeAndSource(mux, workloadTypeInference, "WorkloadsV2", true)
	mux.HandleFunc("/api/v2/workloads/"+validUUID+"/runState", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
}

func TestManageWorkloadLifecycleRun_AcceptedButInvalidWorkloadsV2RunStateResponseHasUnknownOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForTypeAndSource(mux, workloadTypeInference, "WorkloadsV2", true)
	mux.HandleFunc("/api/v2/workloads/"+validUUID+"/runState", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"invalid"`))
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect the workload before retrying")
}

func TestLifecycleRunStateRejectsUnknownAction(t *testing.T) {
	runState, err := lifecycleRunState("pause")

	assert.Empty(t, runState)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestManageWorkloadLifecycleRun_RejectsRequestForWrongInferenceType(t *testing.T) {
	tests := []struct {
		name         string
		workloadType string
		request      map[string]any
	}{
		{name: "standard inference rejects replicas", workloadType: workloadTypeInference, request: distributedInferenceUpdateRequest()},
		{name: "distributed inference rejects image", workloadType: workloadTypeDistribInf, request: inferenceUpdateRequest()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditionsForType(mux, tt.workloadType)

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: tt.request,
			})
			assert.ErrorIs(t, err, runaimcp.ErrValidation)
			assert.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
			assert.ErrorContains(t, err, "unknown field")
		})
	}
}

func TestManageWorkloadLifecycleRun_RejectsMalformedTypedValuesBeforeConfirmation(t *testing.T) {
	tests := []struct {
		name         string
		workloadType string
		request      map[string]any
	}{
		{name: "inference image must be a string", workloadType: workloadTypeInference, request: map[string]any{"spec": map[string]any{"image": 42}}},
		{name: "distributed replicas must be an integer", workloadType: workloadTypeDistribInf, request: map[string]any{"spec": map[string]any{"replicas": 1.5}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditionsForType(mux, tt.workloadType)

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: tt.request,
			})
			assert.ErrorIs(t, err, runaimcp.ErrValidation)
			assert.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
		})
	}
}

func TestManageWorkloadLifecycleRun_MapsInferenceUpdateErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, wantError: runaimcp.ErrValidation},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantError: runaimcp.ErrAuth},
		{name: "forbidden", statusCode: http.StatusForbidden, wantError: runaimcp.ErrForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantError: runaimcp.ErrNotFound},
		{name: "conflict", statusCode: http.StatusConflict, wantError: runaimcp.ErrConflict},
		{name: "throttled", statusCode: http.StatusTooManyRequests, wantError: runaimcp.ErrUpstreamThrottled},
		{name: "server error", statusCode: http.StatusInternalServerError, wantError: runaimcp.ErrUpstream},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditionsForType(mux, workloadTypeInference)
			mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, tt.statusCode, map[string]any{"message": tt.name})
			})

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
			})
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestManageWorkloadLifecycleRun_AcceptsSparseInferenceUpdateResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeInference)
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusAccepted, nil)
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
}

func TestManageWorkloadLifecycleRun_AcceptedButInvalidUpdateResponseHasUnknownOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeInference)
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"invalid"`))
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: inferenceUpdateRequest(), Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect the workload before retrying")
}

func TestManageWorkloadLifecycleRun_MapsDistributedInferenceUpdateError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditionsForType(mux, workloadTypeDistribInf)
	mux.HandleFunc("/api/v1/workloads/distributed-inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"message": "unavailable"})
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionUpdate, WorkloadID: validUUID, UpdateRequest: distributedInferenceUpdateRequest(), Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

func TestManageWorkloadLifecycleRun_BatchItemFailure(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, wantError: runaimcp.ErrValidation},
		{name: "not found", statusCode: http.StatusNotFound, wantError: runaimcp.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditions(mux)
			mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"id":     "batch-1",
					"failed": []map[string]any{{"id": validUUID, "code": tt.statusCode, "message": tt.name}},
				})
			})

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
			})
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestManageWorkloadLifecycleRun_BatchErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, wantError: runaimcp.ErrValidation},
		{name: "forbidden", statusCode: http.StatusForbidden, wantError: runaimcp.ErrForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantError: runaimcp.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			registerLifecyclePreconditions(mux)
			mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, tt.statusCode, map[string]any{"message": tt.name})
			})

			_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
				Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
			})
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestManageWorkloadLifecycleRun_RejectsNilBatchResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditions(mux)
	mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, nil)
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "nil response")
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect the workload before retrying")
}

func TestManageWorkloadLifecycleRun_AcceptedButInvalidBatchResponseHasUnknownOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditions(mux)
	mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"invalid"`))
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect the workload before retrying")
}

func TestManageWorkloadLifecycleRun_RequiresTargetOutcome(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditions(mux)
	mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "batch-1", "succeeded": []string{"another-workload"}, "failed": []any{},
		})
	})

	_, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionStop, WorkloadID: validUUID, Confirm: true,
	})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
	assert.ErrorContains(t, err, "neither success nor failure")
	assert.ErrorContains(t, err, "outcome is unknown")
	assert.ErrorContains(t, err, "inspect the workload before retrying")

}

func TestManageWorkloadLifecycleRun_IgnoresForeignFailureWhenTargetSucceeds(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	registerLifecyclePreconditions(mux)
	mux.HandleFunc("/api/v1/workloads/batch", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":        "batch-1",
			"succeeded": []string{validUUID},
			"failed":    []map[string]any{{"id": "another-workload", "code": http.StatusNotFound, "message": "gone"}},
		})
	})

	out, err := manageWorkloadLifecycleRun(context.Background(), client, manageWorkloadLifecycleInput{
		Action: workloadLifecycleActionRestart, WorkloadID: validUUID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
}

func TestApplyWorkloadOperationResultMatchesCanonicalizedUUID(t *testing.T) {
	out := manageWorkloadLifecycleOutput{Action: workloadLifecycleActionStop, WorkloadID: strings.ToUpper(validUUID)}
	resp := &workloadsapi.BatchResponse{Succeeded: []string{validUUID}}

	got, err := applyWorkloadLifecycleResult(strings.ToUpper(validUUID), workloadLifecycleActionStop, resp, out)
	require.NoError(t, err)
	assert.True(t, got.Applied)
}

func TestGetOperationWorkloadTrustsAPIResponse(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "00000000-0000-0000-0000-000000000001", "name": "wrong", "type": workloadTypeTraining,
		})
	})

	workload, err := getLifecycleWorkload(context.Background(), client, validUUID)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", workload.Id)
}

func registerLifecyclePreconditions(mux *http.ServeMux) {
	registerLifecyclePreconditionsForType(mux, workloadTypeTraining)
}

func registerLifecyclePreconditionsForType(mux *http.ServeMux, workloadType string) {
	registerLifecyclePreconditionsForTypeAndSource(mux, workloadType, "", true)
}

func registerLifecyclePreconditionsForTypeAndSource(mux *http.ServeMux, workloadType, sourceAPI string, suspendSupported bool) {
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		body := map[string]any{
			"id": validUUID, "name": "workload", "type": workloadType,
			"clusterId": submitClusterID, "projectId": submitProjectID, "departmentId": "department-1",
			"actionsSupport": map[string]any{"delete": true, "suspend": suspendSupported},
		}
		if sourceAPI != "" {
			body["sourceApi"] = sourceAPI
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, body)
	})
}

func inferenceUpdateRequest() map[string]any {
	return map[string]any{"spec": map[string]any{"image": "nvcr.io/test:2"}}
}

func distributedInferenceUpdateRequest() map[string]any {
	return map[string]any{"spec": map[string]any{"replicas": 3}}
}

func updateRequestForAction(action string) map[string]any {
	if action == workloadLifecycleActionUpdate {
		return inferenceUpdateRequest()
	}
	return nil
}
