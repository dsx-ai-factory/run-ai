/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

func GetPermissionSets(ctx context.Context, tenantId uint, subjectId string, groups []string) ([]*models.RolePermission, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get subject %s permissions", subjectId), nil)

	rolePermissions := make([]*models.RolePermission, 0)
	db := runaidb.DB().WithContext(ctx)
	err := db.
		Table("role_permissions").
		Select("resource_type, "+
			"bool_or(role_permissions.create) AS create, "+
			"bool_or(role_permissions.read) AS read, "+
			"bool_or(role_permissions.update) AS update, "+
			"bool_or(role_permissions.delete) AS delete, "+
			"bool_or(role_permissions.sync) AS sync").
		Joins("INNER JOIN role_to_permission_set ON role_to_permission_set.permission_set_id = role_permissions.permission_set_id").
		Joins("INNER JOIN access_rules ON role_to_permission_set.role_id = access_rules.role_id").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.deleted_at IS NULL").
		Where(getSubjectClause(db, groups, subjectId)).
		Group("resource_type").
		Find(&rolePermissions)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to fetch user %s permissions. err: %+v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}

	return rolePermissions, nil
}
