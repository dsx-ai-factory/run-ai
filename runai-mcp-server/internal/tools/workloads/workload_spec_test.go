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
	"sync/atomic"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func TestGetWorkloadSpecInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getWorkloadSpecInput
		wantErr bool
	}{
		{"empty", getWorkloadSpecInput{WorkloadID: ""}, true},
		{"whitespace", getWorkloadSpecInput{WorkloadID: "   "}, true},
		{"not a uuid", getWorkloadSpecInput{WorkloadID: "wl-1"}, true},
		{"valid", getWorkloadSpecInput{WorkloadID: validUUID}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetWorkloadSpecRun_TrainingHappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             validUUID,
			"name":           "t1",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"departmentName": "default",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
			"allocatedResources": map[string]any{
				"gpu":       1.0,
				"gpuMemory": "8Gi",
				"cpu":       2.0,
				"cpuMemory": "4Gi",
			},
			"topology": map[string]any{
				"actualPlacement": map[string]any{"level": "rack", "value": "rack-1"},
			},
		})
	})

	var workspaceCalled, distributedCalled, inferenceCalled atomic.Bool
	mux.HandleFunc("/api/v1/workloads/workspaces/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		workspaceCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
	})
	mux.HandleFunc("/api/v1/workloads/distributed/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		distributedCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		inferenceCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
	})

	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name":          "t1",
			"requestedName": "t1",
			"workloadId":    validUUID,
			"projectId":     "11",
			"clusterId":     "c-1",
			"createdBy":     "tkattan@nvidia.com",
			"createdAt":     "2026-05-17T15:00:00Z",
			"desiredPhase":  "Running",
			"spec": map[string]any{
				"image":   "ubuntu:24.04",
				"command": "/bin/sleep",
				"args":    "3600",
				"environmentVariables": []map[string]any{
					{"name": "FOO", "value": "bar"},
				},
				"compute": map[string]any{
					"gpuDevicesRequest": 1,
					"cpuCoreRequest":    2.0,
					"cpuMemoryRequest":  "4Gi",
					"largeShmRequest":   false,
					"gpuRequestType":    "portion",
					"gpuPortionRequest": 1.0,
				},
				"storage":   map[string]any{"ephemeral": map[string]any{"sizeLimit": "10Gi"}},
				"nodePools": []string{"gpu-a"},
			},
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)

	assert.Equal(t, validUUID, out.WorkloadID)
	assert.Equal(t, "t1", out.Name, "name from generic workload endpoint")
	assert.Equal(t, "Training", out.Type, "type from generic workload endpoint")
	assert.Equal(t, "p1", out.ProjectName)
	assert.Equal(t, "default", out.DepartmentName)

	require.NotNil(t, out.Spec, "compact spec returned from kind-specific endpoint")
	assert.Equal(t, "ubuntu:24.04", out.Spec["image"])
	assert.NotContains(t, out.Spec, "command")
	assert.NotContains(t, out.Spec, "args")
	assert.NotContains(t, out.Spec, "environmentVariables")
	require.Contains(t, out.Spec, "compute")
	require.Contains(t, out.Spec, "storage")
	assert.Equal(t, []any{"gpu-a"}, out.Spec["nodePools"])
	compute, ok := out.Spec["compute"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "portion", compute["gpuRequestType"])

	require.NotNil(t, out.AllocatedResources, "allocatedResources forwarded from generic workload endpoint")
	assert.Equal(t, json.Number("1"), out.AllocatedResources["gpu"], "gojsonq decodes numbers as json.Number")
	assert.Equal(t, "8Gi", out.AllocatedResources["gpuMemory"])

	require.NotNil(t, out.Topology)
	require.NotNil(t, out.Topology.ActualPlacement)
	assert.Equal(t, "rack", out.Topology.ActualPlacement["level"])
	assert.Equal(t, "rack-1", out.Topology.ActualPlacement["value"])

	assert.Empty(t, out.PendingSchedulingMessages, "Running workload has no pending-scheduling messages")

	assert.False(t, workspaceCalled.Load(), "workspaces endpoint not called for Training")
	assert.False(t, distributedCalled.Load(), "distributed endpoint not called for Training")
	assert.False(t, inferenceCalled.Load(), "inferences endpoint not called for Training")
}

func TestGetWorkloadSpecRun_RecoversSpecFromSuccessfulGeneratedClientDecodeFailure(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "w5", "type": "Training", "phase": "Running",
			"projectName": "p2", "createdAt": "2026-07-09T10:00:00Z", "phaseUpdatedAt": "2026-07-09T10:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "w5", "requestedName": "w5", "workloadId": validUUID,
			"projectId": "4500002", "clusterId": "c-1", "createdBy": "u",
			"createdAt": "2026-07-09T10:00:00Z",
			// V2 snapshot responses can currently contain an empty value even though
			// the generated enum accepts only Running, Stopped, or Deleted.
			"desiredPhase": "",
			"spec": map[string]any{
				"image":   "ubuntu",
				"command": "sleep infinity",
				"compute": map[string]any{"gpuDevicesRequest": 1},
			},
		})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())
	out, err := getWorkloadSpecRun(ctx, client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.NotNil(t, out.Spec, "a decode error in response metadata must not hide a valid spec")
	assert.Equal(t, "ubuntu", out.Spec["image"])
	assert.NotContains(t, out.Spec, "command", "raw recovery still uses the compact summary by default")
	compute, ok := out.Spec["compute"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, json.Number("1"), compute["gpuDevicesRequest"])
	assert.Contains(t, buf.String(), "recovered spec from raw successful response")
}

func TestGetWorkloadSpecRun_WorkspaceDispatch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "ws1", "type": "Workspace", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/workspaces/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "", "requestedName": "ws1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
			"spec": map[string]any{"image": "jupyter:latest"},
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "Workspace", out.Type)
	assert.Equal(t, "ws1", out.Name, "falls back to generic workload name when kind-specific name is empty")
	require.NotNil(t, out.Spec)
	assert.Equal(t, "jupyter:latest", out.Spec["image"])
}

func TestGetWorkloadSpecRun_DistributedDispatch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "d1", "type": "Distributed", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/distributed/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "d1", "requestedName": "d1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
			"distributedFramework":   "PyTorch",
			"spec":                   map[string]any{"image": "pytorch/pytorch:2.6.0"},
			"masterSpecSameAsWorker": false,
			"masterSpec":             map[string]any{"image": "pytorch/master:2.6.0"},
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "Distributed", out.Type)
	require.NotNil(t, out.Spec)
	assert.Equal(t, "pytorch/pytorch:2.6.0", out.Spec["image"])
	assert.Equal(t, false, out.Spec["masterSpecSameAsWorker"])
	masterSpec, ok := out.Spec["masterSpec"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pytorch/master:2.6.0", masterSpec["image"])
}

func TestGetWorkloadSpecRun_InferenceDispatch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "i1", "type": "Inference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "i1", "requestedName": "i1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
			"spec": map[string]any{
				"image": "nim:1.0.0",
				"servingConfiguration": map[string]any{
					"path":         "/v1",
					"endpointPath": "/health",
				},
			},
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "Inference", out.Type)
	require.NotNil(t, out.Spec)
	assert.Equal(t, "nim:1.0.0", out.Spec["image"])
	assert.Contains(t, out.Spec, "servingConfiguration")
}

func TestGetWorkloadSpecRun_PreservesUnknownFullSpecFieldsAndRemovesSessionTokens(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "t1", "requestedName": "t1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
			"spec": map[string]any{
				"image":           "ubuntu:24.04",
				"futureSpecField": map[string]any{"enabled": true},
				"storage": map[string]any{
					"s3": []map[string]any{
						{
							"name": "models", "bucket": "model-bucket", "path": "/models",
							"accessKeySecret": "model-storage", "secretKeyOfAccessKeyId": "access-key-id",
							"secretKeyOfSecretKey": "secret-access-key", "exclude": false,
							"awsSessionToken":   "must-never-be-returned",
							"futureNestedField": map[string]any{"AWSSessionToken": "also-must-never-be-returned", "region": "us-east-1"},
						},
					},
				},
			},
		})
	})

	compact, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.NotContains(t, compact.Spec, "futureSpecField", "compact mode still limits non-essential fields")
	compactJSON, err := json.Marshal(compact.Spec)
	require.NoError(t, err)
	assert.NotContains(t, string(compactJSON), "awsSessionToken")
	assert.NotContains(t, string(compactJSON), "AWSSessionToken")
	assert.Contains(t, string(compactJSON), "model-bucket", "useful non-secret storage fields remain")
	assert.Contains(t, string(compactJSON), "model-storage")

	full, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID, IncludeFullSpec: true})
	require.NoError(t, err)
	futureField, ok := full.Spec["futureSpecField"].(map[string]any)
	require.True(t, ok, "raw successful response preserves fields newer than the generated client")
	assert.Equal(t, true, futureField["enabled"])
	fullJSON, err := json.Marshal(full.Spec)
	require.NoError(t, err)
	assert.NotContains(t, string(fullJSON), "must-never-be-returned")
	assert.NotContains(t, string(fullJSON), "also-must-never-be-returned")
	assert.Contains(t, string(fullJSON), "futureNestedField", "unknown non-secret nested fields remain in full mode")
	assert.Contains(t, string(fullJSON), "us-east-1")
}

func TestGetWorkloadSpecRun_DistributedInferenceDispatch(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "di1", "type": "DistributedInference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	kindCalled := atomic.Bool{}
	mux.HandleFunc("/api/v1/workloads/distributed-inferences/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		kindCalled.Store(true)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "di1", "requestedName": "di1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
			"spec": map[string]any{
				"replicas":      2,
				"workers":       4,
				"priorityClass": "high",
			},
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.True(t, kindCalled.Load(), "distributed-inferences kind endpoint called for DistributedInference workloads")
	assert.Equal(t, "DistributedInference", out.Type)
	require.NotNil(t, out.Spec, "spec attached for DistributedInference workloads")
	assert.Contains(t, out.Spec, "replicas")
	assert.Contains(t, out.Spec, "workers")
	assert.Contains(t, out.Spec, "priorityClass")
}

func TestGetWorkloadSpecRun_PendingSchedulingMessagesForwarded(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Pending",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
			"pendingSchedulingMessages": []map[string]any{
				{"nodePool": "default", "phaseReason": "NonPreemptibleOverQuota", "reason": "MaxNodePoolResource", "orgType": "PROJECT", "userMessage": "project gpu quota exhausted"},
				{"nodePool": "gpu-a", "phaseReason": "OverLimit", "reason": "MaxNodePoolResource"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "t1", "requestedName": "t1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	require.Len(t, out.PendingSchedulingMessages, 2, "pendingSchedulingMessages forwarded verbatim")
	assert.Equal(t, "default", out.PendingSchedulingMessages[0].NodePool)
	assert.Equal(t, "NonPreemptibleOverQuota", out.PendingSchedulingMessages[0].PhaseReason)
	assert.Equal(t, "MaxNodePoolResource", out.PendingSchedulingMessages[0].Reason)
	assert.Equal(t, "PROJECT", out.PendingSchedulingMessages[0].OrgType)
	assert.Equal(t, "project gpu quota exhausted", out.PendingSchedulingMessages[0].UserMessage)
	assert.Empty(t, out.PendingSchedulingMessages[1].OrgType, "omitempty when upstream omits orgType")
	assert.Empty(t, out.PendingSchedulingMessages[1].UserMessage)
}

func TestGetWorkloadSpecRun_KindSpecificFailureOmitsSpec(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
			"allocatedResources": map[string]any{"gpu": 1.0, "cpu": 2.0},
		})
	})
	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "boom",
			"spec":    map[string]any{"image": "must-not-be-recovered"},
		})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getWorkloadSpecRun(ctx, client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err, "kind-specific failure must not fail the tool")
	assert.Nil(t, out.Spec, "spec omitted when kind-specific call failed")
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "Training workload spec could not be retrieved")
	assert.Equal(t, "Training", out.Type, "type still present")
	assert.Equal(t, "t1", out.Name, "name still falls back to top-level workload's name")
	require.NotNil(t, out.AllocatedResources, "runtime fields still present when only the kind-specific endpoint failed")
	assert.Contains(t, buf.String(), "kind-specific spec lookup failed")
}

func TestGetWorkloadSpecRun_KindSpecificForbiddenStillForwardsRuntime(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "ws1", "type": "Workspace", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/workspaces/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no read access"})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err, "kind-specific 403 must not fail the tool")
	assert.Equal(t, "Workspace", out.Type)
	assert.Nil(t, out.Spec)
}

func TestGetWorkloadSpecRun_NonNativeTypeOmitsSpecWithoutCall(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "di1", "type": "DistributedInference", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Equal(t, "DistributedInference", out.Type, "non-native type still echoed")
	assert.Nil(t, out.Spec, "no kind-specific endpoint, no spec")
	assert.Equal(t, "di1", out.Name)
}

func TestSummarizeWorkloadSpecIncludesNestedRoleResourcesAndPlacement(t *testing.T) {
	spec := map[string]any{
		"command":                   "omit me",
		"distributedFramework":      "MPI",
		"cleanPodPolicy":            "Running",
		"mpiLauncherCreationPolicy": "AtStartup",
		"masterSpecSameAsWorker":    false,
		"masterSpec": map[string]any{
			"image": "master:1", "compute": map[string]any{"gpuDevicesRequest": json.Number("1")},
			"nodePools": []any{"gpu-a"}, "nodeType": "gpu-node", "restartPolicy": "Never", "args": "omit me",
		},
		"leader": map[string]any{
			"image": "leader:1", "storage": map[string]any{"pvcSources": []any{map[string]any{"claimName": "models"}}},
			"servingPort": map[string]any{"port": json.Number("8080")}, "startupPolicy": "LeaderCreated",
		},
		"worker": map[string]any{
			"image": "worker:1", "compute": map[string]any{"cpuCoreRequest": json.Number("2")},
			"nodeAffinityRequired": map[string]any{"nodeSelectorTerms": []any{}},
			"environmentVariables": []any{map[string]any{"name": "SECRET", "value": "omit me"}},
		},
		"replicas": json.Number("2"), "workers": json.Number("4"), "slotsPerWorker": json.Number("2"),
		"minReplicas": json.Number("1"), "maxReplicas": json.Number("8"),
		"autoscaling":  map[string]any{"metric": "concurrency"},
		"masterAssets": map[string]any{"environment": "omit me"},
	}

	summary := summarizeWorkloadSpec(spec)
	assert.Contains(t, summary, "masterSpec")
	assert.Equal(t, "MPI", summary["distributedFramework"])
	assert.Equal(t, "Running", summary["cleanPodPolicy"])
	assert.Equal(t, "AtStartup", summary["mpiLauncherCreationPolicy"])
	assert.Equal(t, false, summary["masterSpecSameAsWorker"])
	assert.Contains(t, summary, "leader")
	assert.Contains(t, summary, "worker")
	assert.Equal(t, json.Number("2"), summary["replicas"])
	assert.Equal(t, json.Number("4"), summary["workers"])
	assert.Equal(t, json.Number("1"), summary["minReplicas"])
	assert.Equal(t, json.Number("8"), summary["maxReplicas"])
	assert.Equal(t, json.Number("2"), summary["slotsPerWorker"])
	assert.Contains(t, summary, "autoscaling")
	assert.NotContains(t, summary, "command")
	assert.NotContains(t, summary, "masterAssets")
	assert.NotContains(t, summary["masterSpec"].(map[string]any), "args")
	assert.NotContains(t, summary["worker"].(map[string]any), "environmentVariables")
}

func TestGetWorkloadSpecRun_AllocatedResourcesOmittedWhenAbsent(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": validUUID, "name": "t1", "type": "Training", "phase": "Pending",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/trainings/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"name": "t1", "requestedName": "t1", "workloadId": validUUID, "projectId": "11", "clusterId": "c-1",
			"createdBy": "u", "createdAt": "2026-05-17T15:00:00Z", "desiredPhase": "Running",
		})
	})

	out, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.NoError(t, err)
	assert.Nil(t, out.AllocatedResources, "allocatedResources omitted when upstream returned no allocation")
	assert.Nil(t, out.Topology, "topology omitted when actualPlacement absent")
}

func TestGetWorkloadSpecRun_PrimaryNotFoundMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestGetWorkloadSpecRun_PrimaryAuthMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusUnauthorized, map[string]any{"message": "unauthorized"})
	})

	_, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func TestGetWorkloadSpecRun_PrimaryForbiddenMaskedAsNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/"+validUUID, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "Insufficient permissions"})
	})

	_, err := getWorkloadSpecRun(context.Background(), client, getWorkloadSpecInput{WorkloadID: validUUID})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound, "CP's 403 masking surfaces as not_found")
	assert.NotErrorIs(t, err, runaimcp.ErrForbidden, "must not also categorize as forbidden")
	assert.NotContains(t, err.Error(), "undefined response type", "generated-client placeholder must not leak")
	assert.NotContains(t, err.Error(), "forbidden", "forbidden prefix must not appear when masked as not_found")
}
