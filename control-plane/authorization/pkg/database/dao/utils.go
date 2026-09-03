/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreationRequestToAccessRuleDao(request api.AccessRuleCreationFields) models.AccessRule {
	return models.AccessRule{
		SubjectId:   request.SubjectId,
		SubjectType: api.SubjectTypeExtended(request.SubjectType),
		RoleID:      uint(request.RoleId),
		ScopeId:     request.ScopeId,
		ScopeType:   request.ScopeType,
	}
}

func IsGlobalResource(resourceType api.ResourceType) bool {
	return !isClusterScopedResourceType(resourceType) && !isOrgType(resourceType) && !hasSystemScope(resourceType)
}

func hasSystemScope(resourceType api.ResourceType) bool {
	switch resourceType {
	case api.RESOURCETYPE_ROLES, api.RESOURCETYPE_SECURITY_SETTINGS, api.RESOURCETYPE_NOTIFICATION_SETTINGS, api.RESOURCETYPE_ACCESS_RULES, api.RESOURCETYPE_WORKLOAD_INTEGRATION_METRICS:
		return true
	default:
		return false
	}
}

// isClusterScopedResourceType
// checks if the resource type is cluster scoped, meaning it is not specific to a project or department.
func isClusterScopedResourceType(resourceType api.ResourceType) bool {
	switch resourceType {
	case api.RESOURCETYPE_NODES, api.RESOURCETYPE_NODEPOOLS, api.RESOURCETYPE_CLUSTER, api.RESOURCETYPE_NODEPOOLS_MINIMAL, api.RESOURCETYPE_CLUSTERS_MINIMAL,
		api.RESOURCETYPE_TENANT, api.RESOURCETYPE_WORKLOAD_PROPERTIES, api.RESOURCETYPE_NETWORK_TOPOLOGIES:
		return true
	default:
		return false
	}
}

func isPropagatingPermissionResourceType(resourceType api.ResourceType) bool {
	switch resourceType {
	case api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION:
		return true
	default:
		return false
	}
}

func isOrgType(resourceType api.ResourceType) bool {
	switch resourceType {
	case api.RESOURCETYPE_WORKLOADS,
		api.RESOURCETYPE_APPLICATIONS,
		api.RESOURCETYPE_TRAININGS,
		api.RESOURCETYPE_INFERENCES,
		api.RESOURCETYPE_WORKSPACES,
		api.RESOURCETYPE_PROJECT,
		api.RESOURCETYPE_DEPARTMENT,
		api.RESOURCETYPE_CLUSTER,
		api.RESOURCETYPE_CLUSTERS_MINIMAL,
		api.RESOURCETYPE_TENANT,
		api.RESOURCETYPE_DATAVOLUMES,
		api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION,
		api.RESOURCETYPE_PVC_ASSETS:
		return true
	default:
		return false
	}
}

func ErrRecordNotFound(err error) bool {
	return errors.Is(err, orm.ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

// isPermissionSetNameConflict detects unique-name conflicts under both raw Postgres errors and
// runaidb's TranslateError, which maps unique violations to orm.ErrDuplicatedKey.
func isPermissionSetNameConflict(err error) bool {
	if errors.Is(err, orm.ErrDuplicatedKey) {
		return true
	}
	var pqErr *pgconn.PgError
	return errors.As(err, &pqErr) && pqErr.Code == runaihttp.ErrDuplicateKey && pqErr.ConstraintName == permissionNameKey
}
