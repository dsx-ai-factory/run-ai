/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	cihClusterUUID = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	cihNodesPath   = "/api/v1/nodes"
	cihCountPath   = "/api/v1/clusters/" + cihClusterUUID + "/nodes/count"
)

// healthNode is a small shorthand for building a fake Node payload. Pass
// gpuType="" for a CPU/system node.
func healthNode(name, pool, status, gpuType string, gpuCount int, conds []map[string]any, taints []map[string]any) map[string]any {
	n := map[string]any{
		"name":        name,
		"nodePool":    pool,
		"status":      status,
		"clusterUuid": cihClusterUUID,
		"createdAt":   "2026-06-10T12:00:00Z",
		"updatedAt":   "2026-06-10T12:00:00Z",
	}
	if gpuType != "" {
		n["gpuInfo"] = map[string]any{"gpuType": gpuType, "gpuCount": gpuCount}
	}
	if conds != nil {
		n["conditions"] = conds
	}
	if taints != nil {
		n["taints"] = taints
	}
	return n
}

func TestGetClusterInfrastructureHealthInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      getClusterInfrastructureHealthInput
		wantErr string
	}{
		{name: "happy", in: getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID}},
		{name: "scoped_to_pool", in: getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID, NodePoolName: "default"}},
		{name: "missing_cluster", in: getClusterInfrastructureHealthInput{}, wantErr: "clusterId"},
		{name: "bad_uuid", in: getClusterInfrastructureHealthInput{ClusterID: "not-a-uuid"}, wantErr: "UUID"},
		{name: "comma_in_node_pool", in: getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID, NodePoolName: "a,b"}, wantErr: "nodePoolName"},
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

// happy path: cluster of 10 nodes total, 2 unhealthy across 2 pools.
// Exercises totals, byNodePool roll-up, GPU sum, and per-node detail mapping.
func TestGetClusterInfrastructureHealthRun_HappyPath(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var gotCountFilter, gotListFilter string
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, r *http.Request) {
		gotCountFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 10})
	})
	mux.HandleFunc(cihNodesPath, func(w http.ResponseWriter, r *http.Request) {
		gotListFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{
			healthNode("node-b", "gpu-a", "NotReady", "Tesla-V100", 8,
				[]map[string]any{{"type": "DiskPressure", "reason": "NoSpaceAvailable", "message": "no disk"}},
				[]map[string]any{{"key": "runai", "value": "drain", "effect": "NoExecute"}},
			),
			healthNode("node-a", "gpu-a", "Unknown", "Tesla-V100", 8, nil, nil),
			healthNode("node-c", "default", "NotReady", "", 0,
				[]map[string]any{{"type": "MemoryPressure", "reason": "NodeOOM"}}, // no message → omitted
				nil,
			),
		}})
	})

	out, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID})
	require.NoError(t, err)

	// Filter forwarding: count gets no filter (cluster-wide); list gets status!=Ready.
	assert.Empty(t, gotCountFilter, "no filter on the count when cluster-wide")
	assert.Equal(t, "status!=Ready,clusterUuid=="+cihClusterUUID, gotListFilter)

	// Totals.
	assert.Equal(t, cihClusterUUID, out.ClusterID)
	assert.Nil(t, out.NodePool)
	assert.Equal(t, 10, out.Totals.NodesTotal)
	assert.Equal(t, 3, out.Totals.NodesUnhealthy)
	assert.Equal(t, 16, out.Totals.GPUsUnschedulable, "8+8 from two GPU nodes (CPU node carries no GPUs)")

	// byNodePool roll-up: 2 rows, sorted alphabetically.
	require.Len(t, out.ByNodePool, 2)
	assert.Equal(t, "default", out.ByNodePool[0].NodePool)
	assert.Equal(t, 1, out.ByNodePool[0].NodesUnhealthy)
	assert.Equal(t, 0, out.ByNodePool[0].GPUsUnschedulable)
	assert.Equal(t, "gpu-a", out.ByNodePool[1].NodePool)
	assert.Equal(t, 2, out.ByNodePool[1].NodesUnhealthy)
	assert.Equal(t, 16, out.ByNodePool[1].GPUsUnschedulable)

	// unhealthyNodes: sorted by pool then name.
	require.Len(t, out.UnhealthyNodes, 3)
	assert.Equal(t, "node-c", out.UnhealthyNodes[0].Name)
	assert.Equal(t, "default", out.UnhealthyNodes[0].NodePool)
	assert.Equal(t, "NotReady", out.UnhealthyNodes[0].Status)
	assert.Nil(t, out.UnhealthyNodes[0].GPUs, "CPU node has no gpus block")
	require.Len(t, out.UnhealthyNodes[0].Conditions, 1)
	assert.Equal(t, "MemoryPressure", out.UnhealthyNodes[0].Conditions[0].Type)
	assert.Empty(t, out.UnhealthyNodes[0].Conditions[0].Message, "null message → empty/omitted")
	assert.Empty(t, out.UnhealthyNodes[0].Taints)

	assert.Equal(t, "node-a", out.UnhealthyNodes[1].Name)
	assert.Equal(t, "Unknown", out.UnhealthyNodes[1].Status)
	require.NotNil(t, out.UnhealthyNodes[1].GPUs)
	assert.Equal(t, "Tesla-V100", out.UnhealthyNodes[1].GPUs.Model)
	assert.Equal(t, 8, out.UnhealthyNodes[1].GPUs.Count)

	assert.Equal(t, "node-b", out.UnhealthyNodes[2].Name)
	require.Len(t, out.UnhealthyNodes[2].Conditions, 1)
	assert.Equal(t, "no disk", out.UnhealthyNodes[2].Conditions[0].Message)
	require.Len(t, out.UnhealthyNodes[2].Taints, 1)
	assert.Equal(t, "drain", out.UnhealthyNodes[2].Taints[0].Value)
	assert.Equal(t, "NoExecute", out.UnhealthyNodes[2].Taints[0].Effect)
}

// On a fully-healthy cluster the filtered fetch returns zero nodes. Totals are
// still emitted (denominator from CountNodes); byNodePool and unhealthyNodes
// are empty slices, not nil — so callers iterating over them never trip.
func TestGetClusterInfrastructureHealthRun_AllHealthyEmptyShape(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 5})
	})
	mux.HandleFunc(cihNodesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{}})
	})

	out, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID})
	require.NoError(t, err)
	assert.Equal(t, 5, out.Totals.NodesTotal)
	assert.Equal(t, 0, out.Totals.NodesUnhealthy)
	assert.Equal(t, 0, out.Totals.GPUsUnschedulable)
	assert.NotNil(t, out.ByNodePool)
	assert.NotNil(t, out.UnhealthyNodes)
	assert.Empty(t, out.ByNodePool)
	assert.Empty(t, out.UnhealthyNodes)
}

// Scoped: both count and list must carry the nodePool== predicate; status
// filter still appears on the list.
func TestGetClusterInfrastructureHealthRun_Scoped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)

	var gotCountFilters, gotListFilters []string
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, r *http.Request) {
		gotCountFilters = parseFilters(r.URL.Query())
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	})
	mux.HandleFunc(cihNodesPath, func(w http.ResponseWriter, r *http.Request) {
		gotListFilters = parseFilters(r.URL.Query())
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": []map[string]any{
			healthNode("n1", "gpu-a", "NotReady", "Tesla-V100", 4, nil, nil),
		}})
	})

	out, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{
		ClusterID: cihClusterUUID, NodePoolName: "gpu-a",
	})
	require.NoError(t, err)
	require.NotNil(t, out.NodePool)
	assert.Equal(t, "gpu-a", *out.NodePool)
	assert.Equal(t, []string{"nodePool==gpu-a"}, gotCountFilters, "count scoped to the pool, no status filter")
	assert.Equal(t, []string{"status!=Ready", "clusterUuid==" + cihClusterUUID, "nodePool==gpu-a"}, gotListFilters, "list carries status, cluster, and pool filters")
	assert.Equal(t, 3, out.Totals.NodesTotal)
	assert.Equal(t, 1, out.Totals.NodesUnhealthy)
	assert.Equal(t, 4, out.Totals.GPUsUnschedulable)
}

// Pagination: first page returns a full unhealthyNodesPageSize, second page a
// short tail. The loop terminates on the short page.
func TestGetClusterInfrastructureHealthRun_Pagination(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": unhealthyNodesPageSize + 1})
	})
	mux.HandleFunc(cihNodesPath, func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		var nodes []map[string]any
		switch offset {
		case "", "0":
			for i := 0; i < unhealthyNodesPageSize; i++ {
				nodes = append(nodes, healthNode(fmt.Sprintf("p1-n%d", i), "default", "NotReady", "", 0, nil, nil))
			}
		case fmt.Sprintf("%d", unhealthyNodesPageSize):
			nodes = []map[string]any{healthNode("p2-tail", "default", "NotReady", "", 0, nil, nil)}
		default:
			http.Error(w, "unexpected offset "+offset, http.StatusInternalServerError)
			return
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
	})

	out, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID})
	require.NoError(t, err)
	assert.Equal(t, unhealthyNodesPageSize+1, out.Totals.NodesUnhealthy)
	assert.Len(t, out.UnhealthyNodes, unhealthyNodesPageSize+1)
}

func TestGetClusterInfrastructureHealthRun_CountForbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestGetClusterInfrastructureHealthRun_ListNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(cihCountPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 0})
	})
	mux.HandleFunc(cihNodesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "not found"})
	})

	_, err := getClusterInfrastructureHealthRun(context.Background(), client, getClusterInfrastructureHealthInput{ClusterID: cihClusterUUID})
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

// helper: filterBy is CSV-joined by the client; split it for clean assertions.
func parseFilters(q url.Values) []string {
	v := q.Get("filterBy")
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
