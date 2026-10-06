/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"fmt"
	"sort"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/google/uuid"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func constructDeleteQueryInCluster(tx orm.DB, tableName string, tenantId uint, clusterId string, from *uuid.UUID, to *uuid.UUID, excludingIds []string) orm.DB {
	deleteQuery := tx.Table(tableName).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("phase IN ?", []string{string(api.DATAVOLUMEPHASE_DELETING), string(api.DATAVOLUMEPHASE_CREATING)})

	if from != nil {
		deleteQuery = deleteQuery.
			Where("id >= ?", from)
	}

	if to != nil {
		deleteQuery = deleteQuery.
			Where("id <= ?", to)
	}

	if len(excludingIds) > 0 {
		deleteQuery = deleteQuery.
			Not("id IN ?", excludingIds)
	}

	return deleteQuery
}

func updateStatusQuery(query orm.DB, status models.DatavolumeStatus) (int64, error) {
	// this action is preformed from the cluster-sync, so we don't want to update the updated_at, updated_by fields;
	// and also - allow updating the status to "Deleting" only if it was "Deleting" already.
	if status.Phase == api.DATAVOLUMEPHASE_DELETING {
		query = query.Where("phase = ?", api.DATAVOLUMEPHASE_DELETING)
	}

	return query.Updates(map[string]interface{}{
		"phase": status.Phase, "phase_message": status.PhaseMessage, "conditions": status.Conditions, "datavolume_pvc_name": status.DatavolumePvcName, "datavolume_pv_name": status.DatavolumePvName,
	})
}

func sortScopesByIdAndType(sharedScopes []models.SharedScope) {
	sort.Slice(sharedScopes, func(i, j int) bool {
		return sharedScopeUniqueKey(sharedScopes[i]) < sharedScopeUniqueKey(sharedScopes[j])
	})
}

func sharedScopeUniqueKey(sharedScope models.SharedScope) string {
	return fmt.Sprintf("%s/%s", sharedScope.ScopeId, sharedScope.ScopeType)
}

func getUsableInProjectIdDbFilter(usableInProjectIdAncestors *authorization_utils.Ancestors) (string, error) {
	if usableInProjectIdAncestors == nil {
		return "", nil
	}

	if usableInProjectIdAncestors.ProjectId == nil ||
		usableInProjectIdAncestors.DepartmentId == nil ||
		usableInProjectIdAncestors.ClusterUuid == nil {
		return "", fmt.Errorf("usableInProjectIdAncestors must include projectId, departmentId and clusterUuid")
	}

	// this filter will be used when submitting workloads in a project, together with "phase=Ready" filter.
	return fmt.Sprintf("(project_id = '%s' OR"+
		"(scope_type = '%s' AND scope_id = '%s') OR"+
		"(scope_type = '%s' AND scope_id = '%s') OR"+
		"(scope_type = '%s' AND scope_id = '%s'))",
		*usableInProjectIdAncestors.ProjectId,
		string(api.SCOPETYPE_PROJECT), *usableInProjectIdAncestors.ProjectId,
		string(api.SCOPETYPE_DEPARTMENT), *usableInProjectIdAncestors.DepartmentId,
		string(api.SCOPETYPE_CLUSTER), *usableInProjectIdAncestors.ClusterUuid), nil
}

func getPermittedScopesDbFilter(permittedScopes *authorization_api.PermittedScopes) string {
	if permittedScopes == nil || (permittedScopes.System != nil && *permittedScopes.System) {
		return ""
	}

	return authorization_utils.GetPermittedScopesDbFilter(*permittedScopes)
}

// we are keeping this code here just in case, we don't need it right now.
// it is designed to maybe be used in the future when we will need to get all the datavolumes that are shared with the caller.

//func getPermittedScopesAndSharedWithDbFilter(permittedScopes, permittedScopesAncestors *authorization_api.PermittedScopes) string {
//	if permittedScopes == nil || (permittedScopes.System != nil && *permittedScopes.System) {
//		return ""
//	}
//
//	datavolumesFilter := authorization_utils.GetPermittedScopesDbFilter(*permittedScopes)
//	sharedScopesFilter := getPermittedScopesDbFilterForSharedScopes(*permittedScopes, models.DatavolumeViewSharedScopeFieldsPrefix)
//
//	if permittedScopesAncestors == nil {
//		return fmt.Sprintf("(%s OR %s)", datavolumesFilter, sharedScopesFilter)
//	}
//
//	sharedScopesAncestorsFilter := getPermittedScopesAncestorsDbFilter(*permittedScopesAncestors)
//	return fmt.Sprintf("(%s OR %s OR %s)", datavolumesFilter, sharedScopesFilter, sharedScopesAncestorsFilter)
//}

// permittedScopes - only the permitted scopes themselves.
// example case: permittedScope=department1; sharedScope=projectId1 => WHERE shared_scope_department_id=department1
// (using only the permitted scopes themselves).

//func getPermittedScopesDbFilterForSharedScopes(permittedScopes authorization_api.PermittedScopes, fieldPrefix string) string {
//	// tenant_id, cluster_id are always the same as the datavolume's tenant_id, cluster_id; so no need for extra filter for them.
//	orClause := "(false"
//
//	if len(permittedScopes.Departments) > 0 {
//		orClause = orClause + fmt.Sprintf(" OR %sdepartment_id IN %s", fieldPrefix, "('" + strings.Join(permittedScopes.Departments, "','") + "')")
//	}
//	if len(permittedScopes.Projects) > 0 {
//		orClause = orClause + fmt.Sprintf(" OR %sproject_id IN %s", fieldPrefix, "('" + strings.Join(permittedScopes.Projects, "','") + "')")
//	}
//
//	orClause += ")"
//
//	return orClause
//}

// permittedScopesAncestors - includes the ancestors of the permitted scopes + permitted scopes themselves.
// example case: permittedScope=projectId1; sharedScope=departmentId1 => WHERE shared_scope_type=DEPARTMENT AND shared_scope_id=departmentId1
// (using GET ancestors of original permittedScopes).

//func getPermittedScopesAncestorsDbFilter(permittedScopesAncestors authorization_api.PermittedScopes) string {
//	// tenant_id is always the same as the datavolume's tenant_id; so no need for extra filter here.
//	// also the permitted projects filter will be included as part of the "permittedScopes" filter (getPermittedScopesDbFilterForSharedScopes func).
//	orClause := "(false"
//
//	if len(permittedScopesAncestors.Clusters) > 0 {
//		orClause = orClause + fmt.Sprintf(" OR (scope_type = '%s' AND scope_id IN %s)", string(api.SCOPETYPE_CLUSTER), "('" + strings.Join(permittedScopesAncestors.Clusters, "','") + "')")
//	}
//	if len(permittedScopesAncestors.Departments) > 0 {
//		orClause = orClause + fmt.Sprintf(" OR (scope_type = '%s' AND scope_id IN %s)", string(api.SCOPETYPE_DEPARTMENT), "('" + strings.Join(permittedScopesAncestors.Departments, "','") + "')")
//	}
//
//	orClause += ")"
//
//	return orClause
//}
