/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func TestCountWorkloads(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("precise"))
		assert.ElementsMatch(t, []string{"phase==Running", "projectId==4500002"}, strings.Split(r.URL.Query().Get("filterBy"), ","))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"count": 7})
	})

	count, err := countWorkloads(context.Background(), client, []string{"phase==Running", "projectId==4500002"}, "", false)
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}

func TestCountWorkloadsRejectsInvalidResponses(t *testing.T) {
	for _, tt := range []struct {
		name       string
		statusCode int
		response   any
	}{
		{name: "nil response", statusCode: http.StatusOK, response: nil},
		{name: "negative count", statusCode: http.StatusOK, response: map[string]any{"count": -1}},
		{name: "upstream failure", statusCode: http.StatusInternalServerError, response: map[string]any{"message": "failed"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/workloads/count", func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, tt.statusCode, tt.response)
			})

			_, err := countWorkloads(context.Background(), client, nil, "", false)
			require.Error(t, err)
			assert.ErrorIs(t, err, runaimcp.ErrUpstream)
		})
	}
}
