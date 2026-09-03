/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbtables

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
)

const (
	dataVolumeColumnClusterId    = "cluster_id"
	dataVolumeColumnProjectId    = "project_id"
	dataVolumeColumnDepartmentId = "department_id"
	DataVolumeColumnTenantId     = "tenant_id"
	DataVolumeColumnId           = "id"
	DataVolumeColumnInternal     = "internal"

	SharedScopesColumnId = "datavolume_id"
	scopeSystem          = "system"
	ScopeTenant          = "tenant"
)

// ScopeIdColumnName returns the column name of the scope id for a given scope type
func ScopeIdColumnName(scopeType api.ScopeType) (string, error) {
	switch scopeType {
	case ScopeTenant:
		return DataVolumeColumnTenantId, nil
	case api.SCOPETYPE_CLUSTER:
		return dataVolumeColumnClusterId, nil
	case api.SCOPETYPE_DEPARTMENT:
		return dataVolumeColumnDepartmentId, nil
	case api.SCOPETYPE_PROJECT:
		return dataVolumeColumnProjectId, nil
	case scopeSystem:
		return "", fmt.Errorf("system scope does not have a scope id column")
	}
	return "", fmt.Errorf("unknown scope type: %s", scopeType)
}
