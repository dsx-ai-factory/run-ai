/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const disabledRolesTable = "disabled_roles"

func DbSetRoleDisabledPerTenant(ctx context.Context, tenantId string, roleId uint) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to set role %d to disabled", roleId), nil)

	disabledRole := models.DisabledRole{
		RoleId:    int64(roleId),
		ScopeType: models.ROLE_SCOPE_TENANT,
		ScopeId:   tenantId,
	}
	db := runaidb.DB().WithContext(ctx)
	_, err := db.
		Table(disabledRolesTable).
		Where(disabledRole).FirstOrCreate(&disabledRole)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to set role %d to disabled. error: %v", roleId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbSetRoleEnabledPerTenant(ctx context.Context, tenantId string, roleId uint) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to set role %d to enabled", roleId), nil)

	db := runaidb.DB().WithContext(ctx)
	_, err := db.
		Table(disabledRolesTable).
		Where("role_id = ?", roleId).
		Where("scope_type = ?", models.ROLE_SCOPE_TENANT).
		Where("scope_id = ?", tenantId).
		Delete(&models.DisabledRole{})

	if err != nil && !ErrRecordNotFound(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error trying to set role %d to enabled. error: %v", roleId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbGetRoleDisabledPerTenant(ctx context.Context, tx orm.DB, tenantId string) (map[int64]models.DisabledRole, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get disabled roles", nil)

	drs := []models.DisabledRole{}
	err := tx.
		Where("scope_type = ? AND scope_id = ?", models.ROLE_SCOPE_TENANT, tenantId).
		Find(&drs)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get disabled roles from db. error: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	mapRoleToDR := map[int64]models.DisabledRole{}
	for _, dr := range drs {
		mapRoleToDR[dr.RoleId] = dr
	}

	return mapRoleToDR, nil
}

func DbGetRoleDisabledById(ctx context.Context, tx orm.DB, tenantId string, roleId uint) (models.DisabledRole, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get disabled role", nil)

	dr := models.DisabledRole{}
	err := tx.
		Where("scope_type = ? AND scope_id = ?", models.ROLE_SCOPE_TENANT, tenantId).
		Where("role_id = ?", roleId).
		First(&dr)
	if err != nil && !ErrRecordNotFound(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get disabled role from db. error: %v", err)
		return models.DisabledRole{}, dbErr.WithDetails(err)
	}

	return dr, nil
}
