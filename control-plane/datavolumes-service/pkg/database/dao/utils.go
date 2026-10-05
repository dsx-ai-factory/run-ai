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
