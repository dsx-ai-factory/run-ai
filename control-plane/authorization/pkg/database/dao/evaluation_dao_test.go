/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/cache"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	commonCache "github.com/run-ai/runai/runai-common-packages/cache/v2"
	cacheTesting "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T { return &v }

// --- filterPendingDeletionRules ---

func TestFilterPendingDeletionRules(t *testing.T) {
	cacheTesting.EnableCache()
	t.Cleanup(func() {
		commonCache.ClearCache()
		commonCache.Close()
	})

	ctx := context.Background()
	activeCluster := "active-cluster-uuid"
	pendingCluster := "pending-cluster-uuid"

	cache.MarkClusterPendingDeletion(ctx, pendingCluster)

	t.Run("keeps non-cluster rules regardless of pending deletion state", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "tenant-1"},
			{ScopeType: api.SCOPETYPE_DEPARTMENT, ScopeId: "dept-1"},
			{ScopeType: api.SCOPETYPE_PROJECT, ScopeId: "project-1"},
			{ScopeType: api.SCOPETYPE_SYSTEM, ScopeId: "system"},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER)})
		assert.Len(t, result, 4)
	})

	t.Run("keeps cluster rule when cluster is not pending deletion", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: activeCluster},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER)})
		assert.Len(t, result, 1)
		assert.Equal(t, activeCluster, result[0].ScopeId)
	})

	t.Run("filters out cluster rule when cluster is pending deletion", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: pendingCluster},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER)})
		assert.Empty(t, result)
	})

	t.Run("filters only pending clusters while keeping active clusters and non-cluster rules", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: activeCluster},
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: pendingCluster},
			{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "tenant-1"},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER)})
		assert.Len(t, result, 2)
		scopeIds := []string{result[0].ScopeId, result[1].ScopeId}
		assert.Contains(t, scopeIds, activeCluster)
		assert.Contains(t, scopeIds, "tenant-1")
		assert.NotContains(t, scopeIds, pendingCluster)
	})

	t.Run("bypasses filter entirely for CLUSTER_CONFIG resource type", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: pendingCluster},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER_CONFIG)})
		assert.Len(t, result, 1)
	})

	t.Run("handles nil ResourceType (applies filter normally)", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: pendingCluster},
			{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "tenant-1"},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: nil})
		assert.Len(t, result, 1)
		assert.Equal(t, "tenant-1", result[0].ScopeId)
	})

	t.Run("returns empty slice when all cluster rules are pending deletion", func(t *testing.T) {
		rules := []models.AccessRule{
			{ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: pendingCluster},
		}
		result := filterAccessRulesWhenClusterPendingDelete(ctx, rules, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_PROJECT)})
		assert.Empty(t, result)
	})

	t.Run("empty input returns empty output", func(t *testing.T) {
		result := filterAccessRulesWhenClusterPendingDelete(ctx, []models.AccessRule{}, api.BackendEvaluationRequest{ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER)})
		assert.Empty(t, result)
	})
}

// --- getScopesClause ---

func TestGetScopesClause(t *testing.T) {
	resourceId := "resource-id"
	tenantId := "tenant-id"
	clusterUUID := "cluster-uuid"
	departmentId := "dept-uuid"
	projectId := "project-uuid"

	t.Run("uses bound placeholders and never interpolates values into the SQL", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		clause, args := getScopesClause(ancestors, request, false, false)
		assert.NotContains(t, clause, "'", "clause must not contain quoted/interpolated values")
		assert.Contains(t, clause, "scope_id = ?")
		assert.Contains(t, clause, "scope_type = ?")
		assert.Equal(t, strings.Count(clause, "?"), len(args), "every placeholder must have a matching bound argument")
	})

	t.Run("includes resource scope and system scope", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		clause, args := getScopesClause(ancestors, request, false, false)
		assert.Contains(t, clause, "scope_id = ?")
		assert.Contains(t, args, interface{}(resourceId))
		assert.Contains(t, args, interface{}(api.SCOPETYPE_SYSTEM))
	})

	t.Run("includes tenant ancestor scope", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		_, args := getScopesClause(ancestors, request, false, false)
		assert.Contains(t, args, interface{}(tenantId))
		assert.Contains(t, args, interface{}(api.SCOPETYPE_TENANT))
	})

	t.Run("includes cluster ancestor scope when present", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId, ClusterUuid: &clusterUUID}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		_, args := getScopesClause(ancestors, request, false, false)
		assert.Contains(t, args, interface{}(clusterUUID))
		assert.Contains(t, args, interface{}(api.SCOPETYPE_CLUSTER))
	})

	t.Run("includes department ancestor scope when present", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId, DepartmentId: &departmentId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		_, args := getScopesClause(ancestors, request, false, false)
		assert.Contains(t, args, interface{}(departmentId))
		assert.Contains(t, args, interface{}(api.SCOPETYPE_DEPARTMENT))
	})

	t.Run("includes project ancestor scope when present", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId, ProjectId: &projectId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		_, args := getScopesClause(ancestors, request, false, false)
		assert.Contains(t, args, interface{}(projectId))
		assert.Contains(t, args, interface{}(api.SCOPETYPE_PROJECT))
	})

	t.Run("isReadOnComputeResource: cluster resource type adds cluster_uuid filter", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &clusterUUID,
			ResourceType: ptrTo(api.RESOURCETYPE_CLUSTER),
		}
		clause, args := getScopesClause(ancestors, request, true, false)
		assert.Contains(t, clause, "cluster_uuid = ?")
		assert.Contains(t, args, interface{}(clusterUUID))
	})

	t.Run("isReadOnComputeResource: non-cluster resource type adds ancestor cluster_uuid filter", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId, ClusterUuid: &clusterUUID}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		clause, args := getScopesClause(ancestors, request, true, false)
		assert.Contains(t, clause, "cluster_uuid = ?")
		assert.Contains(t, args, interface{}(clusterUUID))
	})

	t.Run("isReadOnTenants: tenant resource type adds tenant_id filter", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_TENANT),
		}
		clause, args := getScopesClause(ancestors, request, false, true)
		assert.Contains(t, clause, "tenant_id = ?")
		assert.Contains(t, args, interface{}(resourceId))
	})

	t.Run("isReadOnTenants: non-tenant resource type uses ancestor tenant_id", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		clause, args := getScopesClause(ancestors, request, false, true)
		assert.Contains(t, clause, "tenant_id = ?")
		assert.Contains(t, args, interface{}(tenantId))
	})

	t.Run("malicious resource id is passed as a bound argument, not interpolated", func(t *testing.T) {
		malicious := "x' OR '1'='1' --"
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &malicious,
			ResourceType: ptrTo(api.RESOURCETYPE_TENANT),
		}
		clause, args := getScopesClause(ancestors, request, false, true)
		assert.NotContains(t, clause, malicious, "malicious payload must not appear in the SQL text")
		assert.NotContains(t, clause, "'", "clause must not contain any quotes")
		assert.Contains(t, args, interface{}(malicious), "malicious payload must be a bound argument")
	})

	t.Run("clause is wrapped in parentheses", func(t *testing.T) {
		ancestors := &authorization_utils.Ancestors{TenantId: tenantId}
		request := api.BackendEvaluationRequest{
			ResourceId:   &resourceId,
			ResourceType: ptrTo(api.RESOURCETYPE_PROJECT),
		}
		clause, _ := getScopesClause(ancestors, request, false, false)
		assert.True(t, clause[0] == '(', "clause should start with '('")
		assert.True(t, clause[len(clause)-1] == ')', "clause should end with ')'")
	})
}

// --- RemoveDuplicates ---

func TestRemoveDuplicates(t *testing.T) {
	t.Run("empty input returns nil", func(t *testing.T) {
		result := RemoveDuplicates([]string{})
		assert.Empty(t, result)
	})

	t.Run("no duplicates returns same elements", func(t *testing.T) {
		result := RemoveDuplicates([]string{"a", "b", "c"})
		assert.ElementsMatch(t, []string{"a", "b", "c"}, result)
	})

	t.Run("removes duplicate values", func(t *testing.T) {
		result := RemoveDuplicates([]string{"a", "b", "a", "c", "b"})
		assert.ElementsMatch(t, []string{"a", "b", "c"}, result)
	})

	t.Run("all same values deduplicates to one", func(t *testing.T) {
		result := RemoveDuplicates([]string{"x", "x", "x"})
		assert.Equal(t, []string{"x"}, result)
	})

	t.Run("preserves first occurrence order", func(t *testing.T) {
		result := RemoveDuplicates([]string{"c", "a", "b", "a", "c"})
		assert.Equal(t, []string{"c", "a", "b"}, result)
	})
}
