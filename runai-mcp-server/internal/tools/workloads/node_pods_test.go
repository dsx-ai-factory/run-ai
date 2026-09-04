/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
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

const (
	nodeName        = "node-a"
	clusterAUUID    = "10859ae9-93e2-4fbf-97fa-c586bd721336"
	clusterBUUID    = "20859ae9-93e2-4fbf-97fa-c586bd721336"
	nodeWorkloadID1 = "11111111-1111-1111-1111-111111111111"
	nodeWorkloadID2 = "22222222-2222-2222-2222-222222222222"
	nodePodID1      = "0f3d6a2b-58a3-4f54-9bba-71f0c0a0a9d1"
	nodePodID2      = "0f3d6a2b-58a3-4f54-9bba-71f0c0a0a9d2"
)

func TestGetNodePodsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getNodePodsInput
		wantErr bool
	}{
		{"empty nodeName", getNodePodsInput{}, true},
		{"whitespace nodeName", getNodePodsInput{NodeName: "   "}, true},
		{"nodeName with comma rejected", getNodePodsInput{NodeName: "node-a,node-b"}, true},
		{"valid nodeName, no cluster", getNodePodsInput{NodeName: nodeName}, false},
		{"valid nodeName + valid cluster", getNodePodsInput{NodeName: nodeName, ClusterUUID: clusterAUUID}, false},
		{"example invalid cluster UUID abc", getNodePodsInput{NodeName: "runai-cluster-system-0-3", ClusterUUID: "abc"}, true},
		{"invalid cluster UUID", getNodePodsInput{NodeName: nodeName, ClusterUUID: "not-a-uuid"}, true},
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

func nodePodFixture(name, podID, workloadID, workloadName, clusterID string) map[string]any {
	pod := map[string]any{
		"name":              name,
		"priorityClassName": "default",
		"id":                podID,
		"workloadId":        workloadID,
		"clusterId":         clusterID,
		"nodeName":          nodeName,
		"currentNodePool":   "pool-default",
		"k8sPhase":          "Running",
		"createdAt":         "2026-05-17T15:00:00Z",
		"k8sPhaseUpdatedAt": "2026-05-17T15:00:00Z",
		"updatedAt":         "2026-05-17T15:00:00Z",
		"tenantId":          1,
		"containers":        []any{},
	}
	if workloadName != "" {
		pod["workloadName"] = workloadName
	}
	return pod
}

func TestGetNodePodsRun_HappyPath_DerivesClusterFromPods(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var gotPodFilter, gotPodLimit string
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, r *http.Request) {
		gotPodFilter = r.URL.Query().Get("filterBy")
		gotPodLimit = r.URL.Query().Get("limit")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				nodePodFixture("pod-b", nodePodID2, nodeWorkloadID2, "wl-b", clusterAUUID),
				nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "", clusterAUUID), // missing workloadName/Type → eligible for enrichment.
			},
		})
	})

	var gotNodesLimit, gotNodesFilter string
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		gotNodesLimit = r.URL.Query().Get("limit")
		gotNodesFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{
					"name":        nodeName,
					"status":      "Ready",
					"nodePool":    "pool-default",
					"clusterUuid": clusterAUUID,
					"createdAt":   "2026-05-17T15:00:00Z",
					"updatedAt":   "2026-05-17T15:00:00Z",
					"gpuInfo":     map[string]any{"gpuType": "Tesla-V100", "gpuCount": 8},
				},
			},
		})
	})

	var enrichCalls atomic.Int32
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		enrichCalls.Add(1)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             nodeWorkloadID1,
			"name":           "wl-a-enriched",
			"type":           "Training",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID2, func(w http.ResponseWriter, _ *http.Request) {
		enrichCalls.Add(1)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id":             nodeWorkloadID2,
			"name":           "wl-b-enriched",
			"type":           "Workspace",
			"phase":          "Running",
			"projectName":    "p1",
			"createdAt":      "2026-05-17T15:00:00Z",
			"phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err)

	assert.Equal(t, "nodeName=="+nodeName, gotPodFilter, "pods are filtered by nodeName")
	assert.Equal(t, "500", gotPodLimit, "pods are fetched with limit=500")
	assert.Equal(t, "name=="+nodeName+",clusterUuid=="+clusterAUUID, gotNodesFilter, "nodes are filtered server-side by name and cluster")
	assert.Equal(t, "2", gotNodesLimit, "nodes are fetched with a tight limit since the filter narrows to one")

	assert.Equal(t, nodeName, out.Name, "name echoes the input nodeName")
	assert.Equal(t, "pool-default", out.NodePool)
	assert.Equal(t, "Ready", out.Status)
	require.NotNil(t, out.GpuInfo)
	assert.Equal(t, "Tesla-V100", out.GpuInfo.GpuType)
	assert.Equal(t, float64(8), out.GpuInfo.GpuCount)

	require.Len(t, out.Pods, 2, "both pods returned, sorted by name")
	assert.Equal(t, "pod-a", out.Pods[0].Name, "sorted alphabetically by pod name")
	assert.Equal(t, "pod-b", out.Pods[1].Name)

	assert.Equal(t, nodeWorkloadID1, out.Pods[0].WorkloadID)
	assert.Equal(t, "wl-a-enriched", out.Pods[0].WorkloadName, "missing upstream workloadName filled by enrichment")
	assert.Equal(t, "Training", out.Pods[0].WorkloadType, "workloadType always populated by enrichment")

	assert.Equal(t, nodeWorkloadID2, out.Pods[1].WorkloadID)
	assert.Equal(t, "wl-b", out.Pods[1].WorkloadName, "upstream workloadName preserved when present")
	assert.Equal(t, "Workspace", out.Pods[1].WorkloadType, "workloadType still populated by enrichment")

	assert.Nil(t, out.Pods[0].Measurements, "measurements are never populated by this tool")
	assert.Equal(t, int32(2), enrichCalls.Load(), "one enrichment call per unique workload id")
}

// When the caller provides clusterUuid, it must match the cluster of the pods on the node
func TestGetNodePodsRun_InputClusterMismatchesPods_HardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterBUUID)},
		})
	})

	var calledNodes bool
	mux.HandleFunc("/api/v1/nodes", func(_ http.ResponseWriter, _ *http.Request) { calledNodes = true })

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName, ClusterUUID: clusterAUUID})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "wrong clusterUuid is a validation-class hard error")
	assert.Contains(t, err.Error(), "does not match the cluster of pods", "error message points the LLM at the mismatch")
	assert.Contains(t, err.Error(), clusterAUUID, "error names the input cluster")
	assert.Contains(t, err.Error(), clusterBUUID, "error names the actual pod cluster")
	assert.False(t, calledNodes, "no node-metadata lookup happens when the caller's cluster is wrong")
}

func TestGetNodePodsRun_InputClusterMatchesPods_Success(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID)},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": nodeName, "status": "Ready", "nodePool": "pool-from-A", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl-a", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName, ClusterUUID: clusterAUUID})
	require.NoError(t, err)
	assert.Equal(t, "pool-from-A", out.NodePool)
	require.Len(t, out.Pods, 1)
}

// Pods past the 10-pod cap must not feed the ambiguity check or enrichment — otherwise a
// hidden cluster-B pod would hard-fail a result the LLM never sees.
func TestGetNodePodsRun_AmbiguityIgnoresPodsBeyondCap(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		pods := make([]map[string]any, 0, nodePodsCap+1)
		for i := 1; i <= nodePodsCap; i++ {
			pods = append(pods, nodePodFixture(fmt.Sprintf("pod-%02d", i), validPodIDFor(i), nodeWorkloadID1, "wl", clusterAUUID))
		}
		pods = append(pods, nodePodFixture("pod-99", validPodIDFor(99), nodeWorkloadID2, "wl-b", clusterBUUID))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": pods})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": nodeName, "status": "Ready", "nodePool": "pool-default", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})
	// Cluster-B and workload-2 endpoints intentionally NOT registered — the dropped pod must
	// drive neither cluster resolution nor enrichment.

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err, "dropped cluster-B pod must not turn this into a multi-cluster error")
	require.Len(t, out.Pods, nodePodsCap)
	assert.Equal(t, "pod-01", out.Pods[0].Name)
	assert.Equal(t, "pod-10", out.Pods[nodePodsCap-1].Name)
	assert.Equal(t, "pool-default", out.NodePool, "cluster resolved from visible (cluster-A) pods")
}

func TestGetNodePodsRun_AmbiguousCluster_HardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID),
				nodePodFixture("pod-b", nodePodID2, nodeWorkloadID2, "wl-b", clusterBUUID),
			},
		})
	})

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "multi-cluster pods produce a validation-class hard error")
	assert.Contains(t, err.Error(), "multiple clusters", "error message points the LLM at clusterUuid disambiguation")
}

func TestGetNodePodsRun_NoPodsAndNoCluster_HardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})

	var calledNodes bool
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		calledNodes = true
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []any{}})
	})

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "no pods + no cluster is a validation-class hard error")
	assert.Contains(t, err.Error(), "no pods found")
	assert.Contains(t, err.Error(), "check the node name", "primary recovery: fix the typo")
	assert.Contains(t, err.Error(), "retry with clusterUuid", "secondary recovery: idle real node")
	assert.False(t, calledNodes, "listNodes must not be called when there is no cluster to call against")
}

func TestGetNodePodsRun_IdleNode_WithClusterUUID_Success(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": nodeName, "status": "Ready", "nodePool": "pool-default", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName, ClusterUUID: clusterAUUID})
	require.NoError(t, err, "idle node + valid cluster + node verified to exist must succeed")
	assert.Equal(t, "pool-default", out.NodePool)
	assert.Equal(t, "Ready", out.Status)
	assert.Len(t, out.Pods, 0, "idle node has no pods")
}

func TestGetNodePodsRun_NodeNotInCluster_HardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": []any{}})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []any{}})
	})

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName, ClusterUUID: clusterAUUID})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation, "node-not-found in caller-supplied cluster is validation-class")
	assert.Contains(t, err.Error(), "not found in cluster")
	assert.Contains(t, err.Error(), nodeName)
	assert.Contains(t, err.Error(), clusterAUUID)
}

func TestGetNodePodsRun_ListNodesForbidden_PodsStillReturned(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID)},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "no read permission on nodes"})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl-a-enriched", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getNodePodsRun(ctx, client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err, "listNodes failure must not fail the tool")
	require.Len(t, out.Pods, 1, "pod list survives a forbidden listNodes")
	assert.Empty(t, out.NodePool, "node metadata is dropped when listNodes fails")
	assert.Empty(t, out.Status)
	assert.Nil(t, out.GpuInfo)
	assert.Contains(t, buf.String(), "listNodes failed")
}

func TestGetNodePodsRun_NodeMissingFromListNodes_MetadataOmitted(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID)},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": "node-other", "status": "Ready", "nodePool": "pool-other", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl-a-enriched", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	var buf strings.Builder
	ctx := zerolog.New(&buf).WithContext(context.Background())

	out, err := getNodePodsRun(ctx, client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err, "missing node in listNodes is a soft fail — pods are still returned")
	require.Len(t, out.Pods, 1, "pod list survives a missing-node listNodes response")
	assert.Empty(t, out.NodePool, "node metadata is dropped when the name isn't in the response")
	assert.Empty(t, out.Status)
	assert.Nil(t, out.GpuInfo)
	assert.Contains(t, buf.String(), "node not in listNodes response", "soft fail is logged at warn")
}

func TestGetNodePodsRun_GpuTypeLegacyFallback(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID)},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{
					"name":        nodeName,
					"status":      "Ready",
					"nodePool":    "pool-default",
					"clusterUuid": clusterAUUID,
					"createdAt":   "2026-05-17T15:00:00Z",
					"updatedAt":   "2026-05-17T15:00:00Z",
					"gpuType":     "Tesla-V100", // legacy top-level field; modern gpuInfo missing.
				},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err)
	require.NotNil(t, out.GpuInfo, "legacy gpuType still surfaces gpuInfo")
	assert.Equal(t, "Tesla-V100", out.GpuInfo.GpuType, "falls back to nodes[].gpuType when gpuInfo.gpuType is empty")
	assert.Equal(t, float64(0), out.GpuInfo.GpuCount, "legacy schema carries no gpuCount")
}

func TestGetNodePodsRun_ModernGpuTypeSkipsLegacyDecode(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "wl-a", clusterAUUID)},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{
					"name":        nodeName,
					"status":      "Ready",
					"nodePool":    "pool-default",
					"clusterUuid": clusterAUUID,
					"createdAt":   "2026-05-17T15:00:00Z",
					"updatedAt":   "2026-05-17T15:00:00Z",
					// Both shapes present — modern must win, legacy must NOT be consulted.
					"gpuInfo": map[string]any{"gpuType": "H100", "gpuCount": 8},
					"gpuType": "should-not-win-legacy",
				},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err)
	require.NotNil(t, out.GpuInfo)
	assert.Equal(t, "H100", out.GpuInfo.GpuType, "modern gpuInfo.gpuType wins over legacy gpuType")
	assert.Equal(t, float64(8), out.GpuInfo.GpuCount)
}

func TestGetNodePodsRun_PodCapAt10AfterSorting(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	pods := make([]map[string]any, 0, 12)
	// Insert in reverse-name order to prove the cap is applied AFTER sorting.
	for i := 12; i >= 1; i-- {
		pods = append(pods, nodePodFixture(fmt.Sprintf("pod-%02d", i), validPodIDFor(i), nodeWorkloadID1, "wl", clusterAUUID))
	}
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"pods": pods})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": nodeName, "status": "Ready", "nodePool": "pool-default", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})
	// Single shared workload id → enrichment fans out exactly once.
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID1, "name": "wl", "type": "Training", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err)
	require.Len(t, out.Pods, nodePodsCap, "pods are capped at 10 after sorting")
	assert.Equal(t, "pod-01", out.Pods[0].Name, "sorted ascending by name")
	assert.Equal(t, "pod-10", out.Pods[nodePodsCap-1].Name, "tail of the capped slice is the 10th name")
}

func validPodIDFor(i int) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", i)
}

func TestGetNodePodsRun_EnrichmentPartialFailureKeepsOtherPods(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"pods": []map[string]any{
				nodePodFixture("pod-a", nodePodID1, nodeWorkloadID1, "", clusterAUUID),
				nodePodFixture("pod-b", nodePodID2, nodeWorkloadID2, "", clusterAUUID),
			},
		})
	})
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodes": []map[string]any{
				{"name": nodeName, "status": "Ready", "nodePool": "pool-default", "clusterUuid": clusterAUUID, "createdAt": "2026-05-17T15:00:00Z", "updatedAt": "2026-05-17T15:00:00Z"},
			},
		})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID1, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})
	mux.HandleFunc("/api/v1/workloads/"+nodeWorkloadID2, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": nodeWorkloadID2, "name": "wl-b", "type": "Workspace", "phase": "Running",
			"projectName": "p1", "createdAt": "2026-05-17T15:00:00Z", "phaseUpdatedAt": "2026-05-17T15:00:00Z",
		})
	})

	out, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	require.NoError(t, err, "an enrichment failure must not fail the tool")
	require.Len(t, out.Pods, 2)

	assert.Equal(t, "pod-a", out.Pods[0].Name)
	assert.Empty(t, out.Pods[0].WorkloadName, "failed enrichment leaves workloadName as upstream provided (empty)")
	assert.Empty(t, out.Pods[0].WorkloadType, "failed enrichment leaves workloadType empty")

	assert.Equal(t, "pod-b", out.Pods[1].Name)
	assert.Equal(t, "wl-b", out.Pods[1].WorkloadName, "second enrichment still completed")
	assert.Equal(t, "Workspace", out.Pods[1].WorkloadType)
}

func TestGetNodePodsRun_ListPodsForbiddenIsHardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetNodePodsRun_ListPodsUpstreamErrorIsHardError(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/pods", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	_, err := getNodePodsRun(context.Background(), client, getNodePodsInput{NodeName: nodeName})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}
