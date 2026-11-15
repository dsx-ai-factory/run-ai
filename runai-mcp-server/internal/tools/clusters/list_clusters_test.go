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
)

func TestListClustersFiltersAndPagination(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		input                 listClustersInput
		filter, limit, offset string
	}{
		{name: "name filter and pagination", input: listClustersInput{Name: "shared", Limit: 2, Offset: 4}, filter: "name==shared", limit: "2", offset: "4"},
		{name: "default page", limit: "50", offset: "0"},
		{name: "maximum page size", input: listClustersInput{Limit: 500}, limit: "500", offset: "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/clusters/minimal", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tc.filter, r.URL.Query().Get("filterBy"))
				assert.Equal(t, tc.limit, r.URL.Query().Get("limit"))
				assert.Equal(t, tc.offset, r.URL.Query().Get("offset"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"clusters": []map[string]any{
					{"uuid": "c-2", "tenantId": 1, "name": "shared"},
					{"uuid": "c-1", "tenantId": 1, "name": "shared"},
				}})
			})
			require.NoError(t, tc.input.Validate())
			out, err := listClustersRun(context.Background(), client, tc.input)
			require.NoError(t, err)
			assert.Equal(t, []clusterListItem{
				{ClusterID: "c-2", Name: "shared"},
				{ClusterID: "c-1", Name: "shared"},
			}, out.Clusters, "preserve the API page and ordering")
		})
	}
}
