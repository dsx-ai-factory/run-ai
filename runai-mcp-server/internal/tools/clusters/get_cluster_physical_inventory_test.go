/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	cpiClusterUUID = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	cpiNodesPath   = "/api/v1/nodes"
	telemetryPath  = "/api/v1/nodes/telemetry"
)

// Unconfigured telemetryType → 500, so missing expectations fail loudly.
func nodeTelemetryHandler(t *testing.T, byType map[string][]map[string]any) (http.HandlerFunc, *int32) {
	t.Helper()
	var calls int32
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		typ := r.URL.Query().Get("telemetryType")
		values, ok := byType[typ]
		if !ok {
			http.Error(w, "test handler has no telemetryType="+typ, http.StatusInternalServerError)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"type":      typ,
			"timestamp": "2026-06-09T12:00:00Z",
			"values":    values,
		})
	}, &calls
}

func phaseGroup(pool string) []map[string]any {
	return []map[string]any{{"key": "Nodepool", "value": pool}}
}

func onePool(pool string, value string) []map[string]any {
	return []map[string]any{{"value": value, "groups": phaseGroup(pool)}}
}

func node(name, pool string, gpuType string, gpuCount int) map[string]any {
	n := map[string]any{
		"name":        name,
		"nodePool":    pool,
		"status":      "Ready",
		"clusterUuid": cpiClusterUUID,
		"createdAt":   "2026-06-09T12:00:00Z",
		"updatedAt":   "2026-06-09T12:00:00Z",
	}
	if gpuType != "" {
		n["gpuInfo"] = map[string]any{"gpuType": gpuType, "gpuCount": gpuCount}
	}
	return n
}

func TestGetClusterPhysicalInventoryInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getClusterPhysicalInventoryInput
		wantErr string
	}{
		{name: "happy", in: getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID}},
		{name: "scoped_to_pool", in: getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID, NodePoolName: "default"}},
		{name: "missing_cluster", in: getClusterPhysicalInventoryInput{}, wantErr: "clusterId"},
		{name: "bad_uuid", in: getClusterPhysicalInventoryInput{ClusterID: "not-a-uuid"}, wantErr: "UUID"},
		{name: "comma_in_node_pool", in: getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID, NodePoolName: "a,b"}, wantErr: "nodePoolName"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestGetClusterPhysicalInventoryRun_HappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	// Values chosen so total ≠ allocated + free — free is a primitive, not a derivation.
	tHandler, _ := nodeTelemetryHandler(t, map[string][]map[string]any{
		"TOTAL_GPUS":                 {{"value": "16", "groups": phaseGroup("default")}, {"value": "8", "groups": phaseGroup("gpu-a")}},
		"READY_GPUS":                 {{"value": "16", "groups": phaseGroup("default")}, {"value": "8", "groups": phaseGroup("gpu-a")}},
		"ALLOCATED_GPUS":             {{"value": "4", "groups": phaseGroup("default")}, {"value": "2", "groups": phaseGroup("gpu-a")}},
		"FREE_GPUS":                  {{"value": "8", "groups": phaseGroup("default")}, {"value": "6", "groups": phaseGroup("gpu-a")}},
		"TOTAL_GPU_NODES":            {{"value": "1", "groups": phaseGroup("default")}, {"value": "1", "groups": phaseGroup("gpu-a")}},
		"TOTAL_CPU_CORES":            {{"value": "32", "groups": phaseGroup("default")}, {"value": "16", "groups": phaseGroup("gpu-a")}},
		"ALLOCATED_CPU_CORES":        {{"value": "8", "groups": phaseGroup("default")}, {"value": "4", "groups": phaseGroup("gpu-a")}},
		"TOTAL_CPU_MEMORY_BYTES":     {{"value": fmt.Sprintf("%d", 64*1024*1024*1024), "groups": phaseGroup("default")}},
		"ALLOCATED_CPU_MEMORY_BYTES": {{"value": fmt.Sprintf("%d", 16*1024*1024*1024), "groups": phaseGroup("default")}},
	})
	mux.HandleFunc(telemetryPath, tHandler)
	mux.HandleFunc(cpiNodesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{
			node("n1", "default", "Tesla-V100", 16),
			node("n2", "gpu-a", "NVIDIA-A100", 8),
			node("sys-1", "default", "", 0), // CPU node, gpuInfo null
		}})
	})

	out, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID})
	require.NoError(t, err)
	assert.Equal(t, cpiClusterUUID, out.ClusterID)
	assert.Nil(t, out.NodePool)
	assert.Equal(t, "2026-06-09T12:00:00Z", out.AsOf)

	require.Len(t, out.ByNodePool, 2)
	assert.Equal(t, "default", out.ByNodePool[0].NodePool, "sorted alphabetically")
	assert.Equal(t, 1, out.ByNodePool[0].GPUNodes)
	assert.Equal(t, 16, out.ByNodePool[0].GPUsTotal)
	assert.Equal(t, 16, out.ByNodePool[0].GPUsAllocatable)
	assert.Equal(t, 4, out.ByNodePool[0].GPUsAllocated)
	assert.Equal(t, 8, out.ByNodePool[0].GPUsFree, "free is a telemetry primitive, not total-allocated")
	assert.Equal(t, float64(32), out.ByNodePool[0].CPUCoresTotal)
	assert.Equal(t, float64(8), out.ByNodePool[0].CPUCoresAllocated)
	assert.Equal(t, float64(64), out.ByNodePool[0].MemoryGiBTotal, "64 GiB exactly when bytes = 64*2^30")
	assert.Equal(t, float64(16), out.ByNodePool[0].MemoryGiBAllocated)

	assert.Equal(t, "gpu-a", out.ByNodePool[1].NodePool)
	assert.Equal(t, 8, out.ByNodePool[1].GPUsTotal)
	assert.Equal(t, 6, out.ByNodePool[1].GPUsFree)
	assert.Equal(t, float64(0), out.ByNodePool[1].MemoryGiBTotal, "no memory data for gpu-a → zero")

	assert.Equal(t, 2, out.Totals.GPUNodes)
	assert.Equal(t, 24, out.Totals.GPUsTotal)
	assert.Equal(t, 6, out.Totals.GPUsAllocated)
	assert.Equal(t, 14, out.Totals.GPUsFree)
	assert.Equal(t, float64(48), out.Totals.CPUCoresTotal)

	require.Len(t, out.ByGPUModel, 2)
	assert.Equal(t, "NVIDIA-A100", out.ByGPUModel[0].GPUModel, "sorted alphabetically")
	assert.Equal(t, 1, out.ByGPUModel[0].Nodes)
	assert.Equal(t, 8, out.ByGPUModel[0].GPUsTotal)
	assert.Equal(t, "Tesla-V100", out.ByGPUModel[1].GPUModel)
	assert.Equal(t, 16, out.ByGPUModel[1].GPUsTotal)
}

func TestGetClusterPhysicalInventoryRun_ScopedToNodePool(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var seenPoolParams []string
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		seenPoolParams = append(seenPoolParams, r.URL.Query().Get("nodepoolName"))
		typ := r.URL.Query().Get("telemetryType")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"type": typ, "timestamp": "2026-06-09T12:00:00Z",
			"values": onePool("gpu-a", "1"),
		})
	})
	var seenNodeFilter string
	mux.HandleFunc(cpiNodesPath, func(w http.ResponseWriter, r *http.Request) {
		seenNodeFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{node("n1", "gpu-a", "Tesla-V100", 1)}})
	})

	out, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID, NodePoolName: "gpu-a"})
	require.NoError(t, err)
	require.NotNil(t, out.NodePool)
	assert.Equal(t, "gpu-a", *out.NodePool)
	require.Len(t, seenPoolParams, len(capacityMetrics), "one nodepoolName forward per telemetry call")
	for _, p := range seenPoolParams {
		assert.Equal(t, "gpu-a", p)
	}
	assert.Equal(t, "clusterUuid=="+cpiClusterUUID+",nodePool==gpu-a", seenNodeFilter, "node list filtered to the cluster and scoped pool")
}

func TestGetClusterPhysicalInventoryRun_SkipsUnparseableAndNegative(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		typ := r.URL.Query().Get("telemetryType")
		var values []map[string]any
		switch typ {
		case "TOTAL_GPUS":
			values = []map[string]any{
				{"value": "NaN", "groups": phaseGroup("bad-nan")},
				{"value": "+Inf", "groups": phaseGroup("bad-inf")},
				{"value": "-3", "groups": phaseGroup("bad-neg")},
				{"value": "wat", "groups": phaseGroup("bad-str")},
				{"value": "16", "groups": phaseGroup("default")},
			}
		default:
			values = onePool("default", "0")
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"type": typ, "timestamp": "2026-06-09T12:00:00Z", "values": values})
	})
	mux.HandleFunc(cpiNodesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{}})
	})

	out, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID})
	require.NoError(t, err)
	require.Len(t, out.ByNodePool, 1, "only the 'default' pool survives")
	assert.Equal(t, "default", out.ByNodePool[0].NodePool)
	assert.Equal(t, 16, out.ByNodePool[0].GPUsTotal)
}

func TestGetClusterPhysicalInventoryRun_NodeListFailsDegradesGracefully(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		typ := r.URL.Query().Get("telemetryType")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"type": typ, "timestamp": "2026-06-09T12:00:00Z",
			"values": onePool("default", "1"),
		})
	})
	mux.HandleFunc(cpiNodesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"message": "boom"})
	})

	out, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID})
	require.NoError(t, err, "telemetry succeeded → tool returns degraded, not error")
	assert.NotEmpty(t, out.ByNodePool, "capacity returned even when node list fails")
	assert.Empty(t, out.ByGPUModel, "byGpuModel omitted when node list errors")
}

func TestGetClusterPhysicalInventoryRun_TelemetryForbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})
	_, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetClusterPhysicalInventoryRun_NodeListPaginates(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(telemetryPath, func(w http.ResponseWriter, r *http.Request) {
		typ := r.URL.Query().Get("telemetryType")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"type": typ, "timestamp": "2026-06-09T12:00:00Z", "values": onePool("default", "0"),
		})
	})

	// First page full + short tail → loop terminates on the short page.
	mux.HandleFunc(cpiNodesPath, func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		var nodes []map[string]any
		switch offset {
		case "", "0":
			for i := int32(0); i < common.UpstreamPageSize; i++ {
				nodes = append(nodes, node(fmt.Sprintf("n%d", i), "default", "Tesla-V100", 1))
			}
		case fmt.Sprintf("%d", common.UpstreamPageSize):
			nodes = []map[string]any{node("last", "default", "NVIDIA-A100", 2)}
		default:
			http.Error(w, "unexpected offset "+offset, http.StatusInternalServerError)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
	})

	out, err := getClusterPhysicalInventoryRun(context.Background(), client, getClusterPhysicalInventoryInput{ClusterID: cpiClusterUUID})
	require.NoError(t, err)
	require.Len(t, out.ByGPUModel, 2)
	assert.Equal(t, "NVIDIA-A100", out.ByGPUModel[0].GPUModel)
	assert.Equal(t, 1, out.ByGPUModel[0].Nodes)
	assert.Equal(t, 2, out.ByGPUModel[0].GPUsTotal)
	assert.Equal(t, "Tesla-V100", out.ByGPUModel[1].GPUModel)
	assert.Equal(t, int(common.UpstreamPageSize), out.ByGPUModel[1].Nodes)
}
