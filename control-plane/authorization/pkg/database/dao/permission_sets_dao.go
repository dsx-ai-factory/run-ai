/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"net/http"
)

func DbGetPermissionSets(ctx context.Context, includeInternal bool) (*[]models.PermissionSet, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get permission sets", nil)

	tx := runaidb.DB().WithContext(ctx)
	permissionSets := &[]models.PermissionSet{}
	query := tx.Preload("Permissions")

	if !includeInternal {
		query = query.Where("internal = false")
	}

	err := query.
		Order("id DESC").
		Find(permissionSets)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get permission sets from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return permissionSets, nil
}

func DbGetPermissionSetsByIds(ctx context.Context, ids []string) (*[]models.PermissionSet, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get permission sets", nil)

	tx := runaidb.DB().WithContext(ctx)
	permissionSets := &[]models.PermissionSet{}
	err := tx.Preload("Permissions").
		Where("id in ?", ids).
		Find(permissionSets)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get permission sets from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return permissionSets, nil
}

func DbGetPermissionSetById(ctx context.Context, id uuid.UUID) (*models.PermissionSet, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get permission set", nil)

	permissionSet := &models.PermissionSet{}
	err := runaidb.DB().WithContext(ctx).
		Preload("Permissions").
		Where("internal = false").
		Where("id = ?", id).
		First(permissionSet)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get permission set from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return permissionSet, nil
}

func DbGetPermissionSetsByRoleId(ctx context.Context, roleId uint) (*[]models.PermissionSet, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get permission sets", nil)

	permissionSets := &[]models.PermissionSet{}

	err := runaidb.DB().WithContext(ctx).
		Table("permission_sets").
		Joins("JOIN role_to_permission_set rps ON rps.permission_set_id = permission_sets.id").
		Where("rps.role_id = ?", roleId).
		Preload("Permissions").
		Find(&permissionSets)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get permission sets from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return permissionSets, nil
}
