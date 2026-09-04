/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"strconv"

	"github.com/google/uuid"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

func sharedScopeApiToModel(sharedScopeBase api.SharedScopeBase, datavolumeId uuid.UUID,
	tenantId uint, subjectId string,
) models.SharedScope {
	return models.SharedScope{
		DatavolumeId: datavolumeId,
		ScopeType:    sharedScopeBase.ScopeType,
		ScopeId:      sharedScopeBase.ScopeId,
		TenantId:     tenantId,
		CreatedBy:    subjectId,
	}
}

func sharedScopeModelToBase(sharedScope models.SharedScope) api.SharedScopeBase {
	return api.SharedScopeBase{
		ScopeType: sharedScope.ScopeType,
		ScopeId:   sharedScope.ScopeId,
	}
}

func sharedScopeModelToApi(sharedScope models.SharedScope) api.SharedScope {
	return api.SharedScope{
		ScopeType: sharedScope.ScopeType,
		ScopeId:   sharedScope.ScopeId,
		ScopeName: sharedScope.ScopeName,
		CreatedAt: sharedScope.CreatedAt,
		CreatedBy: sharedScope.CreatedBy,
	}
}

func sharedScopesModelToApi(sharedScopes []models.SharedScope) []api.SharedScope {
	return convutils.ConvertSliceTypes(sharedScopes, sharedScopeModelToApi)
}

func sharedScopesBasesToMap(sharedScopeBases []api.SharedScopeBase) map[api.SharedScopeBase]bool {
	result := make(map[api.SharedScopeBase]bool)
	for _, sharedScope := range sharedScopeBases {
		result[sharedScope] = true
	}
	return result
}

func sharedScopesToMap(sharedScopes []models.SharedScope) map[api.SharedScopeBase]models.SharedScope {
	result := make(map[api.SharedScopeBase]models.SharedScope)
	for _, sharedScope := range sharedScopes {
		result[sharedScopeModelToBase(sharedScope)] = sharedScope
	}
	return result
}

func filterOutNonPermittedSharedScopes(tenantId uint, clusterId uuid.UUID,
	sharedScopes []models.SharedScope, permittedScopes *authorization_api.PermittedScopes,
) []models.SharedScope {
	if permittedScopes == nil ||
		(permittedScopes.System != nil && *permittedScopes.System) {
		return sharedScopes
	}

	permittedTenants := convertSliceToMap(permittedScopes.Tenants)
	tenantIdStr := strconv.FormatUint(uint64(tenantId), 10)
	if _, tenantPermitted := permittedTenants[tenantIdStr]; tenantPermitted {
		return sharedScopes
	}

	permittedClusters := convertSliceToMap(permittedScopes.Clusters)
	if _, clusterPermitted := permittedClusters[clusterId.String()]; clusterPermitted {
		return sharedScopes
	}

	permittedProjects := convertSliceToMap(permittedScopes.Projects)
	permittedDepartments := convertSliceToMap(permittedScopes.Departments)

	result := make([]models.SharedScope, 0)
	for _, sharedScope := range sharedScopes {
		if _, projectPermitted := permittedProjects[sharedScope.ProjectId]; projectPermitted {
			result = append(result, sharedScope)
			continue
		}

		if _, departmentPermitted := permittedDepartments[sharedScope.DepartmentId]; departmentPermitted {
			result = append(result, sharedScope)
		}
	}
	return result
}

func areAllScopesPermitted(tenantId uint, clusterId uuid.UUID,
	sharedScopes []models.SharedScope, permittedScopes *authorization_api.PermittedScopes,
) bool {
	permittedSharedScopes := filterOutNonPermittedSharedScopes(tenantId, clusterId, sharedScopes, permittedScopes)
	return len(permittedSharedScopes) == len(sharedScopes)
}
