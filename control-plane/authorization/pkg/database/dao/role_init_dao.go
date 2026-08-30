package dao

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"

	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
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
			if config.Get[*authorization_config.General]().UsePermissionSets {
				existingRole.Description = roleRequest.Description
				existingRole.UpdatedAt = roleRequest.UpdatedAt
				existingRole.Deprecated = roleRequest.Deprecated
				if roleRequest.RoleType != "" {
					existingRole.RoleType = roleRequest.RoleType
				}
				if err := upsertPermissionSetAndPermissions(ctx, tx, roleRequest, existingRole); err != nil {
					return err
				}
				_, err := tx.Omit(fmt.Sprintf("%s.*", PermissionSetsField)).Save(existingRole)
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save role %s. err: %+v", roleRequest.Name, err)
				}
				return err
			}

			existingRole.Permissions = roleRequest.Permissions
			existingRole.Description = roleRequest.Description
			existingRole.CreatedBy = roleRequest.CreatedBy
			existingRole.UpdatedAt = roleRequest.UpdatedAt
			if roleRequest.RoleType != "" {
				existingRole.RoleType = roleRequest.RoleType
			}
			if _, err := tx.
				Table(RolesTable).
				Where("id = ?", existingRole.ID).
				Omit(PermissionSetsField).
				Save(&existingRole); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update role %s. err: %+v", existingRole.Name, err)
				return err
			}
			// Batch upsert all permissions in one SQL statement
			if len(existingRole.Permissions) > 0 {
				for _, permission := range existingRole.Permissions {
					permission.RoleID = existingRole.ID
				}
				if _, err := tx.
					Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "role_id"}, {Name: "resource_type"}},
						DoUpdates: clause.AssignmentColumns([]string{"create", "read", "update", "delete", "sync"}),
					}).
					Create(&existingRole.Permissions); err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save permissions for role %s. err: %+v", roleRequest.Name, err)
					return err
				}
			}

			// Delete permissions that are no longer needed
			permissionsToDelete := findPermissionsToDelete(api.AllowedResourceTypeEnumValues, existingRole.Permissions)
			for _, permission := range permissionsToDelete {
				if _, err := tx.
					Where("role_id = ?", permission.RoleID).
					Where("resource_type = ?", permission.ResourceType).
					Delete(&permission); err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete permission of role id %d on resource type %s. err: %+v", permission.RoleID, permission.ResourceType, err)
					return err
				}
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
		return existingRole.ID, nil
	} else {
		tx := runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true})
		var createErr error
		if config.Get[*authorization_config.General]().UsePermissionSets {
			// create permissionSet
			rolePermissionsFromPermissions(roleRequest.Permissions, nil)
			rolePermissions := rolePermissionsFromPermissions(roleRequest.Permissions, nil)
			permissionSet := &models.PermissionSet{
				Name:        getInternalPermissionSetNameFromRole(roleRequest.Name, nil),
				Internal:    true,
				Permissions: rolePermissions,
			}
			roleRequest.PermissionSets = &[]models.PermissionSet{*permissionSet}
			roleRequest.Permissions = nil
			_, createErr = tx.
				Session(&gorm.Session{FullSaveAssociations: true}).
				Create(&roleRequest)
		} else {
			// Create new role
			_, createErr = tx.
				Omit(PermissionSetsField).
				Session(&gorm.Session{FullSaveAssociations: true}).
				Create(&roleRequest)
		}
		if createErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create role %s. err: %+v", roleRequest.Name, createErr)
			return 0, createErr
		}
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
