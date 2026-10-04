/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

type RoleDaoInitializer struct {
}

// DbUpsertRole + DbUpsertLegacyRole Is used in the services initialization only, to reconcile the roles and permissions tables to the static roles definitions.
func (r *RoleDaoInitializer) DbUpsertRole(ctx context.Context, roleRequest models.Role) (uint, error) {
	roleRequest.CreatedBy = RunAiRoleCreator
	roleRequest.NoPermissionSet = ptr.To(false)

	// we don't use on conflict since permission sets needs special handling.
	existingRole, err := r.getRoleByName(ctx, roleRequest.Name, roleRequest.ScopeId, roleRequest.ScopeType)
	if err != nil && !ErrRecordNotFound(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to retrieve role %s. err: %+v", roleRequest.Name, err)
		return 0, err
	}

	if existingRole == nil {
		return createNonExistingRole(ctx, roleRequest)
	}

	roleRequest.ID = existingRole.ID
	err = runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Transaction(func(tx orm.DB) error {
		// there is a role with that name - update the role and the associated permission sets
		if err := tx.Actual().Model(existingRole).Association(PermissionSetsField).Clear(); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clear permission sets: %+v", err)
			return err
		}
		_, err := tx.
			Where("id = ?", roleRequest.ID).
			Where("name = ?", roleRequest.Name).
			Omit(fmt.Sprintf("%s.*", PermissionSetsField)).
			Updates(&roleRequest)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update role %s. err: %+v", roleRequest.Name, err)
		}
		return err
	})
	if err != nil {
		return 0, err
	}
	deletePermissionsWithoutSet(ctx, roleRequest)
	return roleRequest.ID, nil
}

func (r *RoleDaoInitializer) DbUpsertLegacyRole(ctx context.Context, roleRequest models.Role) (uint, error) {
	roleRequest.CreatedBy = RunAiRoleCreator

	// GORM does not support the deletion of deleted association in upsert, so in order to reconcile also the deleted permissions, we need to find them explicitly if they exist.
	existingRole, err := r.getRoleByName(ctx, roleRequest.Name, roleRequest.ScopeId, roleRequest.ScopeType)
	if err != nil && !ErrRecordNotFound(err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to retrieve role %s. err: %+v", roleRequest.Name, err)
		return 0, err
	}

	// We cannot use ON CONFLICT with the COALESCE index because PostgreSQL's ON CONFLICT
	// requires referencing actual columns, not expressions.
	// So we explicitly check if the role exists and update or create accordingly.
	if existingRole != nil {
		err = runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Transaction(func(tx orm.DB) error {
			// Update existing role - preserve immutable/system fields from existing role
			existingRole.Description = roleRequest.Description
			existingRole.UpdatedAt = roleRequest.UpdatedAt
			existingRole.Deprecated = roleRequest.Deprecated
			if err := upsertPermissionSetAndPermissions(ctx, tx, roleRequest, existingRole); err != nil {
				return err
			}
			_, err := tx.Omit(fmt.Sprintf("%s.*", PermissionSetsField)).Save(existingRole)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save role %s. err: %+v", roleRequest.Name, err)
			}
			return err
		})
		if err != nil {
			return 0, err
		}
		return existingRole.ID, nil
	}

	permissionSet := &models.PermissionSet{
		Name:        getInternalPermissionSetNameFromRole(roleRequest.Name, nil),
		Internal:    true,
		Permissions: rolePermissionsFromPermissions(roleRequest.Permissions, nil),
	}
	roleRequest.PermissionSets = &[]models.PermissionSet{*permissionSet}
	roleRequest.Permissions = nil

	_, createErr := runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Create(&roleRequest)
	if createErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create role %s. err: %+v", roleRequest.Name, createErr)
		return 0, createErr
	}

	return roleRequest.ID, nil
}

func (r *RoleDaoInitializer) getRoleByName(ctx context.Context, roleName string, scopeId *string, scopeType models.RoleScope) (*models.Role, error) {
	role := &models.Role{}
	query := runaidb.DB().
		Preload("Permissions").
		Preload("PermissionSets").
		Where("name = ?", roleName).
		Where("scope_type = ?", scopeType)

	if scopeId != nil {
		query = query.Where("scope_id = ?", *scopeId)
	}
	err := query.First(role)
	if err != nil {
		if !ErrRecordNotFound(err) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get role with name %s, error: %s", roleName, err.Error())
		}
		return nil, err
	}

	return role, nil
}
