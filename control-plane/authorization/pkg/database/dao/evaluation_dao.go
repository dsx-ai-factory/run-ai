/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/cache"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func EvaluatePermission(ctx context.Context, tenantId uint, subjectId string, groups []string, request api.BackendEvaluationRequest, forceCalculatePermittedScopes bool) (*api.PermittedScopes, *api.Error) {
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		return EvaluatePermissionLegacy(ctx, tenantId, subjectId, groups, request, forceCalculatePermittedScopes)
	}
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to evaluate subject %s permissions", subjectId), nil)
	isReadOnClusterScopedResource := request.ResourceType != nil &&
		request.RequestAction != nil &&
		*request.RequestAction == api.ACTION_READ &&
		(isClusterScopedResourceType(*request.ResourceType) || isPropagatingPermissionResourceType(*request.ResourceType))
	isReadOnTenants := request.ResourceType != nil &&
		request.RequestAction != nil &&
		*request.RequestAction == api.ACTION_READ &&
		*request.ResourceType == api.RESOURCETYPE_TENANT

	isGlobal := request.ResourceType == nil || IsGlobalResource(*request.ResourceType)
	accessRules := make([]models.AccessRule, 0)
	db := runaidb.DB().WithContext(ctx)
	query := db.
		Table("access_rules").
		Select("scope_type, scope_id, cluster_uuid, tenant_id").
		Joins("JOIN role_to_permission_set ON role_to_permission_set.role_id = access_rules.role_id").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_set_id = role_to_permission_set.permission_set_id").
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		Where(getSubjectClause(db, groups, subjectId))

	if !isGlobal && request.ResourceType != nil && request.ResourceId != nil && *request.ResourceType != api.RESOURCETYPE_ACCESS_RULES {
		ancestors, _, err := cache.GetOrgHierarchyManager().GetAncestors(ctx, *request.ResourceId, *request.ResourceType)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error response from the ancestors service, usually that means that the resource is not found")
			return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}

		scopesClause, scopesArgs := getScopesClause(ancestors, request, isReadOnClusterScopedResource, isReadOnTenants)
		query = query.Where(scopesClause, scopesArgs...)
	}

	if request.ResourceType != nil {
		query = query.
			Where("role_permissions.resource_type = ?", *request.ResourceType)
	}

	if request.RequestAction != nil {
		switch *request.RequestAction {
		case api.ACTION_READ:
			query = query.Where("role_permissions.read = true")
		case api.ACTION_CREATE:
			query = query.Where("role_permissions.create = true")
		case api.ACTION_UPDATE:
			query = query.Where("role_permissions.update = true")
		case api.ACTION_DELETE:
			query = query.Where("role_permissions.delete = true")
		case api.ACTION_SYNC:
			query = query.Where("role_permissions.sync = true")
		}
	}

	err := query.
		Find(&accessRules)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch user %s permissions. err: %+v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}

	if len(accessRules) == 0 {
		return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	accessRules = filterAccessRulesWhenClusterPendingDelete(ctx, accessRules, request)
	if len(accessRules) == 0 {
		return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	if isGlobal && !forceCalculatePermittedScopes {
		// Global resource with at least one access rule that allow the requested action, approved with no need of scopes.
		return nil, nil
	}

	permittedScopes := api.PermittedScopes{
		System:      false,
		Tenants:     []string{},
		Clusters:    []string{},
		Departments: []string{},
		Projects:    []string{},
	}
	for _, accessRule := range accessRules {
		switch accessRule.ScopeType {
		case api.SCOPETYPE_SYSTEM:
			permittedScopes.System = true
		case api.SCOPETYPE_TENANT:
			permittedScopes.Tenants = append(permittedScopes.Tenants, accessRule.ScopeId)
		case api.SCOPETYPE_CLUSTER:
			permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ScopeId)
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
		case api.SCOPETYPE_DEPARTMENT:
			if isReadOnClusterScopedResource && accessRule.ClusterUuid != "" {
				permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ClusterUuid)
			}
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
			permittedScopes.Departments = append(permittedScopes.Departments, accessRule.ScopeId)
		case api.SCOPETYPE_PROJECT:
			if isReadOnClusterScopedResource && accessRule.ClusterUuid != "" {
				permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ClusterUuid)
			}
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
			permittedScopes.Projects = append(permittedScopes.Projects, accessRule.ScopeId)
		}
	}

	permittedScopes.Tenants = RemoveDuplicates(permittedScopes.Tenants)
	permittedScopes.Clusters = RemoveDuplicates(permittedScopes.Clusters)
	permittedScopes.Departments = RemoveDuplicates(permittedScopes.Departments)
	permittedScopes.Projects = RemoveDuplicates(permittedScopes.Projects)

	return &permittedScopes, nil
}

func EvaluatePermissionLegacy(ctx context.Context, tenantId uint, subjectId string, groups []string, request api.BackendEvaluationRequest, forceCalculatePermittedScopes bool) (*api.PermittedScopes, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to evaluate subject %s permissions", subjectId), nil)
	isReadOnClusterScopedResource := request.ResourceType != nil &&
		request.RequestAction != nil &&
		*request.RequestAction == api.ACTION_READ &&
		(isClusterScopedResourceType(*request.ResourceType) || isPropagatingPermissionResourceType(*request.ResourceType))
	isReadOnTenants := request.ResourceType != nil &&
		request.RequestAction != nil &&
		*request.RequestAction == api.ACTION_READ &&
		*request.ResourceType == api.RESOURCETYPE_TENANT

	isGlobal := request.ResourceType == nil || IsGlobalResource(*request.ResourceType)
	accessRules := make([]models.AccessRule, 0)
	db := runaidb.DB().WithContext(ctx)
	query := db.
		Table("access_rules").
		Select("scope_type, scope_id, cluster_uuid, tenant_id").
		Joins("INNER JOIN permissions ON permissions.role_id = access_rules.role_id").
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		Where(getSubjectClause(db, groups, subjectId))

	if !isGlobal && request.ResourceType != nil && request.ResourceId != nil && *request.ResourceType != api.RESOURCETYPE_ACCESS_RULES {
		ancestors, _, err := cache.GetOrgHierarchyManager().GetAncestors(ctx, *request.ResourceId, *request.ResourceType)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error response from the ancestors service, usually that means that the resource is not found")
			return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
		}

		scopesClause, scopesArgs := getScopesClause(ancestors, request, isReadOnClusterScopedResource, isReadOnTenants)
		query = query.Where(scopesClause, scopesArgs...)
	}

	if request.ResourceType != nil {
		query = query.
			Where("permissions.resource_type = ?", *request.ResourceType)
	}

	if request.RequestAction != nil {
		switch *request.RequestAction {
		case api.ACTION_READ:
			query = query.Where("permissions.read = true")
		case api.ACTION_CREATE:
			query = query.Where("permissions.create = true")
		case api.ACTION_UPDATE:
			query = query.Where("permissions.update = true")
		case api.ACTION_DELETE:
			query = query.Where("permissions.delete = true")
		case api.ACTION_SYNC:
			query = query.Where("permissions.sync = true")
		}
	}

	err := query.
		Find(&accessRules)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch user %s permissions. err: %+v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}

	if len(accessRules) == 0 {
		return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	accessRules = filterAccessRulesWhenClusterPendingDelete(ctx, accessRules, request)
	if len(accessRules) == 0 {
		return nil, api.NewError(http.StatusForbidden, "Insufficient permissions", nil)
	}

	if isGlobal && !forceCalculatePermittedScopes {
		// Global resource with at least one access rule that allow the requested action, approved with no need of scopes.
		return nil, nil
	}

	permittedScopes := api.PermittedScopes{
		System:      false,
		Tenants:     []string{},
		Clusters:    []string{},
		Departments: []string{},
		Projects:    []string{},
	}
	for _, accessRule := range accessRules {
		switch accessRule.ScopeType {
		case api.SCOPETYPE_SYSTEM:
			permittedScopes.System = true
		case api.SCOPETYPE_TENANT:
			permittedScopes.Tenants = append(permittedScopes.Tenants, accessRule.ScopeId)
		case api.SCOPETYPE_CLUSTER:
			permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ScopeId)
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
		case api.SCOPETYPE_DEPARTMENT:
			if isReadOnClusterScopedResource && accessRule.ClusterUuid != "" {
				permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ClusterUuid)
			}
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
			permittedScopes.Departments = append(permittedScopes.Departments, accessRule.ScopeId)
		case api.SCOPETYPE_PROJECT:
			if isReadOnClusterScopedResource && accessRule.ClusterUuid != "" {
				permittedScopes.Clusters = append(permittedScopes.Clusters, accessRule.ClusterUuid)
			}
			if isReadOnTenants {
				permittedScopes.Tenants = append(permittedScopes.Tenants, fmt.Sprintf("%d", accessRule.TenantId))
			}
			permittedScopes.Projects = append(permittedScopes.Projects, accessRule.ScopeId)
		}
	}

	permittedScopes.Tenants = RemoveDuplicates(permittedScopes.Tenants)
	permittedScopes.Clusters = RemoveDuplicates(permittedScopes.Clusters)
	permittedScopes.Departments = RemoveDuplicates(permittedScopes.Departments)
	permittedScopes.Projects = RemoveDuplicates(permittedScopes.Projects)

	return &permittedScopes, nil
}

func getSubjectClause(db orm.DB, groups []string, subjectId string) orm.DB {
	orClause := db.Where(
		db.Where("subject_id = ?", subjectId).Where("subject_type != ?", api.SUBJECTTYPE_GROUP),
	).Or(db.Where("subject_type = ?", models.AllAuthenticatedUsers))

	if len(groups) == 0 || (len(groups) == 1 && groups[0] == "") {
		return orClause
	}

	return orClause.Or(
		db.Where("subject_id IN ?", groups).Where("subject_type = ?", api.SUBJECTTYPE_GROUP),
	)
}

// getScopesClause builds the scope filter as a parameterized SQL fragment.
// All user/ancestor supplied values are passed as bound arguments (never
// interpolated into the SQL string) to prevent SQL injection.
func getScopesClause(ancestors *authorization_utils.Ancestors, request api.BackendEvaluationRequest, isReadOnComputeResource bool, isReadOnTenants bool) (string, []interface{}) {
	clause := "(False "
	args := make([]interface{}, 0)

	resourceClause, resourceArgs := getResourceClause(*request.ResourceId, api.ScopeType(*request.ResourceType))
	clause += resourceClause
	args = append(args, resourceArgs...)

	clause += " OR scope_type = ?"
	args = append(args, api.SCOPETYPE_SYSTEM)

	if ancestors.TenantId != "" {
		c, a := getResourceClause(ancestors.TenantId, api.SCOPETYPE_TENANT)
		clause += c
		args = append(args, a...)
	}
	if ancestors.ClusterUuid != nil {
		c, a := getResourceClause(*ancestors.ClusterUuid, api.SCOPETYPE_CLUSTER)
		clause += c
		args = append(args, a...)
	}
	if ancestors.DepartmentId != nil {
		c, a := getResourceClause(*ancestors.DepartmentId, api.SCOPETYPE_DEPARTMENT)
		clause += c
		args = append(args, a...)
	}
	if ancestors.ProjectId != nil {
		c, a := getResourceClause(*ancestors.ProjectId, api.SCOPETYPE_PROJECT)
		clause += c
		args = append(args, a...)
	}

	if isReadOnComputeResource {
		if *request.ResourceType == api.RESOURCETYPE_CLUSTER || *request.ResourceType == api.RESOURCETYPE_CLUSTERS_MINIMAL {
			clause += " OR cluster_uuid = ?"
			args = append(args, *request.ResourceId)
		} else {
			if ancestors.ClusterUuid != nil {
				clause += " OR cluster_uuid = ?"
				args = append(args, *ancestors.ClusterUuid)
			}
		}
	}

	if isReadOnTenants {
		if *request.ResourceType == api.RESOURCETYPE_TENANT {
			clause += " OR tenant_id = ?"
			args = append(args, *request.ResourceId)
		} else {
			clause += " OR tenant_id = ?"
			args = append(args, ancestors.TenantId)
		}
	}

	clause += ")"
	return clause, args
}

func getResourceClause(resourceId string, scopeType api.ScopeType) (string, []interface{}) {
	return " OR (scope_id = ? AND scope_type = ?)", []interface{}{resourceId, scopeType}
}

// filterAccessRulesWhenClusterPendingDelete removes access rules whose scope type is cluster
// when that cluster is in a graceful deletion process (delete_requested_at set,
// deleted_at not yet set).
//
// CLUSTER_CONFIG requests are exempt: they need to reach the cluster even while
// it is in the graceful deletion process (e.g. to clean up the cluster's own access rules).
func filterAccessRulesWhenClusterPendingDelete(ctx context.Context, accessRules []models.AccessRule, request api.BackendEvaluationRequest) []models.AccessRule {
	if request.ResourceType != nil && *request.ResourceType == api.RESOURCETYPE_CLUSTER_CONFIG {
		return accessRules
	}
	filtered := make([]models.AccessRule, 0, len(accessRules))
	for _, accessRule := range accessRules {
		isPendingDeletion := accessRule.ScopeType == api.SCOPETYPE_CLUSTER && cache.IsClusterPendingDeletion(ctx, accessRule.ScopeId)
		if !isPendingDeletion {
			filtered = append(filtered, accessRule)
		}
	}
	return filtered
}

func RemoveDuplicates(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
