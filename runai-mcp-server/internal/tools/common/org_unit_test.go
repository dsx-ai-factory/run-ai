/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

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

const (
	projectsPath    = "/api/v1/org-unit/projects"
	departmentsPath = "/api/v1/org-unit/departments"
)

func TestResolveOrgUnitID_ProjectHappy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotFilter string
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{{"id": "p-1", "clusterId": "c-1"}},
		})
	})

	id, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "alpha", "")
	require.NoError(t, err)
	assert.Equal(t, "p-1", id)
	assert.Equal(t, "name==alpha", gotFilter)
}

func TestResolveOrgUnitID_DepartmentNarrowedByCluster(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	var gotFilter string
	mux.HandleFunc(departmentsPath, func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("filterBy")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"departments": []map[string]any{{"id": "d-9", "clusterId": "c-1"}},
		})
	})

	id, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeDepartment, "research", "c-1")
	require.NoError(t, err)
	assert.Equal(t, "d-9", id)
	assert.Equal(t, "name==research,clusterId==c-1", gotFilter)
}

// A comma in the name would smuggle a second predicate into FilterBy.
// Guard sits inside the resolver so no caller can forget — no HTTP call.
func TestResolveOrgUnitID_RejectsCommaInName(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t)
	_, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "alpha,beta", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestResolveOrgUnitID_RejectsCommaInClusterID(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t)
	_, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "alpha", "c-1,c-2")
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestResolveOrgUnitID_UnknownOrgType(t *testing.T) {
	client, _ := runaiclienttest.NewServer(t)
	_, err := common.ResolveOrgUnitID(context.Background(), client, "team", "alpha", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestResolveOrgUnitID_NotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"projects": []map[string]any{}})
	})
	_, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "ghost", "")
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestResolveOrgUnitID_Ambiguous(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"projects": []map[string]any{
				{"id": "p-1", "clusterId": "c-1"},
				{"id": "p-2", "clusterId": "c-2"},
			},
		})
	})
	_, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "dup", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_org_unit")
	assert.Contains(t, err.Error(), "cluster c-1")
	assert.Contains(t, err.Error(), "cluster c-2")
}

func TestResolveOrgUnitID_ForbiddenMapped(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(projectsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusForbidden, map[string]any{"message": "forbidden"})
	})
	_, err := common.ResolveOrgUnitID(context.Background(), client, common.OrgTypeProject, "alpha", "")
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}
