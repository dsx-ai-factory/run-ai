/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"strings"
	"testing"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/stretchr/testify/require"
)

// ---------- getUsableInProjectIdDbFilter ----------

func TestGetUsableInProjectIdDbFilter_Nil(t *testing.T) {
	filter, err := getUsableInProjectIdDbFilter(nil)
	require.NoError(t, err)
	require.Equal(t, "", filter)
}

func TestGetUsableInProjectIdDbFilter_MissingFields(t *testing.T) {
	_, err := getUsableInProjectIdDbFilter(&authorization_utils.Ancestors{TenantId: "1"})
	require.Error(t, err)
}

func TestGetUsableInProjectIdDbFilter_Full(t *testing.T) {
	filter, err := getUsableInProjectIdDbFilter(fullAncestors())
	require.NoError(t, err)
	require.Contains(t, filter, "project_id = 'proj-1'")
	require.Contains(t, filter, "scope_type = 'project' AND scope_id = 'proj-1'")
	require.Contains(t, filter, "scope_type = 'department' AND scope_id = 'dep-1'")
	require.Contains(t, filter, "scope_type = 'cluster' AND scope_id = 'cluster-uuid-1'")
}

// ---------- getPermittedScopesDbFilter ----------

func TestGetPermittedScopesDbFilter_Nil(t *testing.T) {
	require.Equal(t, "", getPermittedScopesDbFilter(nil))
}

func TestGetPermittedScopesDbFilter_System(t *testing.T) {
	system := true
	require.Equal(t, "", getPermittedScopesDbFilter(&authorization_api.PermittedScopes{System: &system}))
}

func TestGetPermittedScopesDbFilter_NonSystem(t *testing.T) {
	ps := &authorization_api.PermittedScopes{
		Projects:    []string{"proj-1"},
		Departments: []string{"dep-1"},
		Clusters:    []string{"cluster-1"},
		Tenants:     []string{"tenant-1"},
	}
	filter := getPermittedScopesDbFilter(ps)
	require.Contains(t, filter, "tenant_id IN")
	require.Contains(t, filter, "cluster_id IN")
	require.Contains(t, filter, "department_id IN")
	require.Contains(t, filter, "project_id IN")
}

// ---------- sortScopesByIdAndType / sharedScopeUniqueKey ----------

func TestSortScopesByIdAndType(t *testing.T) {
	scopes := []models.SharedScope{
		{ScopeId: "c", ScopeType: api.SCOPETYPE_PROJECT},
		{ScopeId: "a", ScopeType: api.SCOPETYPE_PROJECT},
		{ScopeId: "b", ScopeType: api.SCOPETYPE_DEPARTMENT},
	}
	sortScopesByIdAndType(scopes)
	require.Equal(t, "a", scopes[0].ScopeId)
	require.Equal(t, "b", scopes[1].ScopeId)
	require.Equal(t, "c", scopes[2].ScopeId)
}

func TestSortScopesByIdAndType_SameIdDifferentType(t *testing.T) {
	scopes := []models.SharedScope{
		{ScopeId: "x", ScopeType: api.SCOPETYPE_PROJECT},
		{ScopeId: "x", ScopeType: api.SCOPETYPE_DEPARTMENT},
	}
	sortScopesByIdAndType(scopes)
	// "x/department" < "x/project"
	require.Equal(t, api.SCOPETYPE_DEPARTMENT, scopes[0].ScopeType)
	require.Equal(t, api.SCOPETYPE_PROJECT, scopes[1].ScopeType)
}

func TestSharedScopeUniqueKey(t *testing.T) {
	key := sharedScopeUniqueKey(models.SharedScope{ScopeId: "id1", ScopeType: api.SCOPETYPE_PROJECT})
	require.True(t, strings.HasPrefix(key, "id1/"))
	require.Contains(t, key, string(api.SCOPETYPE_PROJECT))
}
