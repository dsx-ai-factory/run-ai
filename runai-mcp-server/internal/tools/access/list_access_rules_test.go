/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
)

// A NotReady rule is what the console renders as "issues found". The tool must
// surface the backend status so the assistant explains it from real data instead
// of fabricating a reason.
func TestListAccessRules_SurfacesStatus(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"accessRules": []map[string]any{
				{
					"id": 7, "subjectId": "chaoyu222@run.ai", "subjectType": "user",
					"roleId": 2, "roleName": "L1 researcher", "scopeId": "p1", "scopeType": "project",
					"scopeName": "proj", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
					"tenantId": 1, "createdBy": "admin",
					"status": map[string]any{
						"phase": "NotReady", "message": "cluster not connected", "details": "sync pending",
					},
				},
				{
					"id": 8, "subjectId": "chaoyu222@run.ai", "subjectType": "user",
					"roleId": 2, "roleName": "L1 researcher", "scopeId": "p2", "scopeType": "project",
					"scopeName": "proj2", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
					"tenantId": 1, "createdBy": "admin",
				},
			},
			"totalRecords": 2,
		})
	})

	out, err := listAccessRulesRun(context.Background(), client, listAccessRulesInput{SubjectID: "chaoyu222@run.ai"})
	require.NoError(t, err)
	require.Len(t, out.AccessRules, 2)

	require.NotNil(t, out.AccessRules[0].Status, "NotReady rule must carry its status")
	assert.Equal(t, "NotReady", out.AccessRules[0].Status.Phase)
	assert.Equal(t, "cluster not connected", out.AccessRules[0].Status.Message)
	assert.Equal(t, "sync pending", out.AccessRules[0].Status.Details)

	assert.Nil(t, out.AccessRules[1].Status, "a rule with no backend status stays nil, not an invented one")
}
