/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const validClusterID = "10859ae9-93e2-4fbf-97fa-c586bd721336"

func TestListNodePoolsInput_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      listNodePoolsInput
		wantErr bool
	}{
		{"empty ok", listNodePoolsInput{}, false},
		{"valid filter and bounds", listNodePoolsInput{ClusterID: validClusterID, Limit: 50, Offset: 10}, false},
		{"negative limit", listNodePoolsInput{Limit: -1}, true},
		{"limit over max", listNodePoolsInput{Limit: common.MaxLimit + 1}, true},
		{"negative offset", listNodePoolsInput{Offset: -1}, true},
		{"comma in clusterId", listNodePoolsInput{ClusterID: "a,b"}, true},
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

func TestListNodePoolsRun_MapsAndSorts(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{
				{"id": "np-2", "name": "gpu-b", "clusterId": "c-1", "isDefault": false, "phase": "Ready", "networkTopologyName": "topo-x"},
				{"id": "np-1", "name": "gpu-a", "clusterId": "c-1", "isDefault": true, "phase": "Empty"},
			},
		})
	})

	out, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	require.NoError(t, err)
	require.Len(t, out.NodePools, 2)

	// Deterministic sort: clusterId then name, so gpu-a precedes gpu-b.
	assert.Equal(t, "np-1", out.NodePools[0].NodePoolID)
	assert.Equal(t, "gpu-a", out.NodePools[0].NodePoolName)
	assert.Equal(t, "c-1", out.NodePools[0].ClusterID)
	assert.True(t, out.NodePools[0].IsDefault)
	assert.Equal(t, "Empty", out.NodePools[0].Phase)
	assert.Nil(t, out.NodePools[0].NetworkTopologyName, "null topology maps to nil pointer")

	assert.Equal(t, "gpu-b", out.NodePools[1].NodePoolName)
	assert.Equal(t, "Ready", out.NodePools[1].Phase)
	require.NotNil(t, out.NodePools[1].NetworkTopologyName)
	assert.Equal(t, "topo-x", *out.NodePools[1].NetworkTopologyName)

	assert.False(t, out.HasMore)
	assert.Zero(t, out.NextOffset)
}

func TestListNodePoolsRun_FilterByClusterID(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotFilter string
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": []map[string]any{}})
	})

	out, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{ClusterID: validClusterID})
	require.NoError(t, err)
	assert.Equal(t, "clusterId=="+validClusterID, gotFilter, "clusterId becomes a filterBy predicate")
	require.NotNil(t, out.NodePools, "empty result is a non-nil array")
	assert.Len(t, out.NodePools, 0)
}

func TestListNodePoolsRun_NoFilterWhenClusterIDEmpty(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var sawFilter bool
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, r *http.Request) {
		sawFilter = r.URL.Query().Has("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": []map[string]any{}})
	})

	_, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	require.NoError(t, err)
	assert.False(t, sawFilter, "no filterBy sent when clusterId is omitted")
}

func TestListNodePoolsRun_Pagination(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{
				{"id": "np-1", "name": "a", "clusterId": "c-1"},
				{"id": "np-2", "name": "b", "clusterId": "c-1"},
				{"id": "np-3", "name": "c", "clusterId": "c-1"},
			},
		})
	})

	first, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{Limit: 2, Offset: 0})
	require.NoError(t, err)
	require.Len(t, first.NodePools, 2)
	assert.Equal(t, "a", first.NodePools[0].NodePoolName)
	assert.Equal(t, "b", first.NodePools[1].NodePoolName)
	assert.True(t, first.HasMore)
	assert.Equal(t, int32(2), first.NextOffset)

	second, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{Limit: 2, Offset: 2})
	require.NoError(t, err)
	require.Len(t, second.NodePools, 1)
	assert.Equal(t, "c", second.NodePools[0].NodePoolName)
	assert.False(t, second.HasMore)
	assert.Zero(t, second.NextOffset)
}

func TestListNodePoolsRun_ForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})

	_, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestListNodePoolsRun_UpstreamErrorMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusInternalServerError, map[string]any{"message": "boom"})
	})

	_, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

// Always returns a full upstream page (200), so fetchAllNodePools keeps paging
// until it trips the maxNodePools cap and fails loudly rather than truncating.
func TestListNodePoolsRun_CapHit(t *testing.T) {
	full := make([]map[string]any, 200)
	for i := range full {
		full[i] = map[string]any{"id": "np", "name": "n", "clusterId": "c-1"}
	}
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": full})
	})

	_, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	assert.ErrorIs(t, err, runaimcp.ErrUpstream)
}

// Pools from different clusters must order by clusterId first (not name), so
// c-1/zzz precedes c-2/aaa. Also asserts an absent phase maps to "".
func TestListNodePoolsRun_SortsByClusterIDThenName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/node-pools", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{
				{"id": "np-a", "name": "aaa", "clusterId": "c-2", "phase": "Ready"},
				{"id": "np-z", "name": "zzz", "clusterId": "c-1"},
			},
		})
	})

	out, err := listNodePoolsRun(context.Background(), client, listNodePoolsInput{})
	require.NoError(t, err)
	require.Len(t, out.NodePools, 2)
	assert.Equal(t, "c-1", out.NodePools[0].ClusterID, "lower clusterId sorts first regardless of name")
	assert.Equal(t, "zzz", out.NodePools[0].NodePoolName)
	assert.Equal(t, "", out.NodePools[0].Phase, "absent upstream phase maps to empty string")
	assert.Equal(t, "c-2", out.NodePools[1].ClusterID)
}
