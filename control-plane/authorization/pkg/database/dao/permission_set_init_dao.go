/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

const (
	permissionNameKey = "permission_sets_name_key"
)

type PermissionSetDaoInitializer struct {
}

// DbUpsertPermissionSet Is used in the services initialization only, to reconcile the permission sets and permissions tables to the static permission sets definitions.
func (ps *PermissionSetDaoInitializer) DbUpsertPermissionSet(ctx context.Context, psRequest models.PermissionSet) (uuid.UUID, error) {
	psRequest.CreatedBy = RunAiRoleCreator

	tx := runaidb.DB().WithContext(ctx)
	_, dbError := tx.Create(&psRequest)
	if dbError != nil {
		if isPermissionSetNameConflict(dbError) {
			// on conflict will not work with uuid since permissions will not be created with correct permission_set_id (gorm limitation with uuid)
			// so we use check for isPermissionSetNameConflict
			err := ps.updatePermissionSet(ctx, tx, &psRequest)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update permission set %s. err: %+v", psRequest.Name, err)
				return uuid.Nil, err
			}
			// handle permissions (new + old)
			err = ps.addNewPermissions(tx, psRequest.Permissions, psRequest.ID)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to add permission on init of permission set name %s. err: %+v", psRequest.Name, err)
				return uuid.Nil, err
			}
			err = ps.deleteOldPermissions(tx, psRequest.Permissions, psRequest.ID)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete permission on init of permission set name %s. err: %+v", psRequest.Name, err)
				return uuid.Nil, err
			}
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert permission set on init %s. err: %+v", psRequest.Name, dbError)
			return uuid.Nil, dbError
		}
	}
	return psRequest.ID, nil
}

func (ps *PermissionSetDaoInitializer) updatePermissionSet(ctx context.Context, tx orm.DB, psRequest *models.PermissionSet) error {
	existingPermissionSet, err := ps.getPermissionSetByName(ctx, psRequest.Name)
	if err != nil {
		return err
	}
	psRequest.ID = existingPermissionSet.ID
	_, err = tx.Model(psRequest).Updates(map[string]interface{}{
		"updated_at":  psRequest.UpdatedAt,
		"description": psRequest.Description,
		"internal":    psRequest.Internal,
	})
	return err
}

func (ps *PermissionSetDaoInitializer) addNewPermissions(tx orm.DB, permissions []models.RolePermission, permissionSetId uuid.UUID) error {
	for i := range permissions {
		permissions[i].PermissionSetID = permissionSetId
		_, err := tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "permission_set_id"}, {Name: "resource_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"create", "update", "delete", "read", "sync"}),
			}).
			Create(&permissions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PermissionSetDaoInitializer) deleteOldPermissions(tx orm.DB, existingPermissions []models.RolePermission, permissionSetId uuid.UUID) error {
	permissionsToDelete := []models.RolePermission{}
	allowedPermissionsMap := map[string]bool{}
	existingPermissionsMap := map[string]bool{}

	for _, requestedResourceType := range api.AllowedResourceTypeEnumValues {
		allowedPermissionsMap[string(requestedResourceType)] = true
	}
	for _, permission := range existingPermissions {
		existingPermissionsMap[(string(permission.ResourceType))] = true
	}
	for _, currentPermission := range existingPermissions {
		if _, found := allowedPermissionsMap[string(currentPermission.ResourceType)]; !found {
			permissionsToDelete = append(permissionsToDelete, models.RolePermission{
				PermissionSetID: permissionSetId,
				ResourceType:    currentPermission.ResourceType,
			})
		}
	}
	for _, resourceType := range api.AllowedResourceTypeEnumValues {
		if _, found := existingPermissionsMap[string(resourceType)]; !found {
			permissionsToDelete = append(permissionsToDelete, models.RolePermission{
				PermissionSetID: permissionSetId,
				ResourceType:    resourceType,
			})
		}
	}

	for _, permission := range permissionsToDelete {
		_, err := tx.
			Where("permission_set_id = ?", permission.PermissionSetID).
			Where("resource_type = ?", permission.ResourceType).
			Delete(&permission)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PermissionSetDaoInitializer) getPermissionSetByName(ctx context.Context, psName string) (*models.PermissionSet, error) {
	permissionSet := &models.PermissionSet{}
	err := runaidb.DB().
		Preload("Permissions").
		Where("name = ?", psName).
		First(permissionSet)
	if err != nil {
		if !ErrRecordNotFound(err) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get permission set on init with name %s, error: %s", psName, err.Error())
		}
		return nil, err
	}

	return permissionSet, nil
}
