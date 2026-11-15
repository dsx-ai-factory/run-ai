/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func TestGetPolicyMapping(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		fields     []policyMappingFieldRef
		response   map[string]any
	}{
		{
			name: "list available fields", path: "/api/v1/policies/mappings/minimal",
			response: map[string]any{"fields": map[string]any{"image": map[string]any{}}, "specSelectors": map[string]any{"allContainers": "Every container"}},
		},
		{
			name: "get details for selected fields", path: "/api/v1/policies/mappings/field-info",
			fields:   []policyMappingFieldRef{{FieldName: "image", SpecSelector: "allContainers"}},
			response: map[string]any{"breakdown": map[string]any{"fields": map[string]any{"image": []any{}}, "specSelectors": map[string]any{}}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, map[string]any{"group": "apps", "version": "v1", "kind": "Deployment"}, body["workloadType"])
				if len(tc.fields) > 0 {
					assert.Equal(t, []any{map[string]any{"fieldName": "image", "specSelector": "allContainers"}}, body["fieldsRef"])
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, tc.response)
			})

			in := getWorkloadPolicyMappingInput{
				WorkloadType: common.GVK{Group: "apps", Version: "v1", Kind: "Deployment"},
				FieldsRef:    tc.fields,
			}
			require.NoError(t, in.Validate())
			out, err := getWorkloadPolicyMappingRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.Equal(t, tc.response, out.Mapping)
		})
	}
}
