/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbtables

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

const (
	DeleteThisRecord = "1 = 1"

	ColumnLastValue = "last_value"

	DeletedAtColumn = "deleted_at"
	CreatedAtColumn = "created_at"
	CreatedByColumn = "created_by"
	UpdatedAtColumn = "updated_at"
	UpdatedByColumn = "updated_by"
	DeletedByColumn = "deleted_by"

	PoliciesTable                  = "policies"
	PoliciesColumnId               = "id"
	PoliciesColumnScope            = "scope"
	PoliciesColumnTenantId         = "tenant_id"
	PoliciesColumnClusterId        = "cluster_id"
	PoliciesColumnDepartmentId     = "department_id"
	PoliciesColumnProjectId        = "project_id"
	PoliciesColumnName             = "name"
	PoliciesColumnDefaults         = "defaults_v2"
	PoliciesColumnRules            = "rules_v2"
	PoliciesColumnImposedAssets    = "imposed_assets"
	PoliciesColumnWorkloadType     = "workload_type"
	PoliciesColumnGvkWorkload      = "gvk_workload"
	PoliciesColumnValidationStatus = "validation_status"
	PoliciesColumnValidationIssues = "validation_issues"

	EffectivePoliciesTable              = "effective_policies"
	EffectivePoliciesColumnPolicyId     = "policy_id"
	EffectivePoliciesColumnTenantId     = "tenant_id"
	EffectivePoliciesColumnUpdateCount  = "update_count"
	EffectivePoliciesColumnForcesUpdate = "forced_update"

	EffectivePoliciesColumnDefaultFromOrder = "default_from_order"
	EffectivePoliciesColumnValidationError  = "validation_error"

	SequenceEffectivePolicyUpdateCounter = "effective_policies_update_count"

	ApprovedDeletionsTable           = "approved_deletions"
	ApprovedDeletionsColumnClusterId = "cluster_id"

	MappingFormattersTable           = "mapping_formatters"
	MappingFormattersColumnName      = "name"
	MappingFormattersColumnScopeType = "scope_type"
	MappingFormattersColumnTenantId  = "tenant_id"

	MappingFieldGroupsTable           = "mapping_field_groups"
	MappingFieldGroupsColumnName      = "name"
	MappingFieldGroupsColumnScopeType = "scope_type"
	MappingFieldGroupsColumnTenantId  = "tenant_id"

	MappingSelectorsTable            = "mapping_selectors"
	MappingSelectorsColumnGvkGroup   = "gvk_group"
	MappingSelectorsColumnGvkKind    = "gvk_kind"
	MappingSelectorsColumnGvkVersion = "gvk_version"
	MappingSelectorsColumnScopeType  = "scope_type"
	MappingSelectorsColumnTenantId   = "tenant_id"

	MappingLockingTable          = "mapping_locking"
	MappingLockingColumnTenantId = "tenant_id"
	MappingLockingColumnLockedAt = "locked_at"

	EntitiesCacheTable            = "entities_cache"
	EntitiesCacheColumnId         = "id"
	EntitiesCacheColumnTenantId   = "tenant_id"
	EntitiesCacheColumnEntityType = "entity_type"
	EntitiesCacheColumnEntityId   = "entity_id"
	EntitiesCacheColumnCachedInfo = "cached_info"

	EntitiesDependencyTable               = "entities_dependency"
	EntitiesDependencyColumnCacheId       = "cache_id"
	EntitiesDependencyColumnTenantId      = "tenant_id"
	EntitiesDependencyColumnComponentType = "component_type"
	EntitiesDependencyColumnComponentId   = "component_id"
)

// ScopeIdColumnName returns the column name of the scope id for a given scope type
func ScopeIdColumnName(scopeType api.ScopeType) (string, error) {
	switch scopeType {
	case api.SCOPETYPE_TENANT:
		return PoliciesColumnTenantId, nil
	case api.SCOPETYPE_CLUSTER:
		return PoliciesColumnClusterId, nil
	case api.SCOPETYPE_DEPARTMENT:
		return PoliciesColumnDepartmentId, nil
	case api.SCOPETYPE_PROJECT:
		return PoliciesColumnProjectId, nil
	case api.SCOPETYPE_SYSTEM:
		return "", fmt.Errorf("system scope does not have a scope id column")
	}
	return "", fmt.Errorf("unknown scope type: %s", scopeType)
}
