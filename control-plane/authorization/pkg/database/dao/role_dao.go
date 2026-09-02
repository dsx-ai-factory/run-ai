/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const (
	RunAiRoleCreator            = "Run:ai"
	RolesTable                  = "roles"
	ExclusionViolationErrorCode = "23P01"
	PermissionSetsField         = "PermissionSets"
	InternalPermissionSetSuffix = "internal-ps"
)

func createNonExistingRole(ctx context.Context, roleRequest models.Role) (uint, error) {
	_, err := runaidb.DB().
		Omit(fmt.Sprintf("%s.*", PermissionSetsField)).
		Create(&roleRequest)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create role %s. err: %+v", roleRequest.Name, err)
		return 0, err
	}
	return roleRequest.ID, nil
}

func deletePermissionsWithoutSet(ctx context.Context, roleRequest models.Role) {
	// if the role was migrated from legacy - remove its permissions - do not fail the whole process if fails
	_, err := runaidb.DB().WithContext(ctx).Where("role_id = ?", roleRequest.ID).Delete(&models.Permission{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete old role %s permissions. error: %+v", roleRequest.Name, err)
	}
}

func getOrCreatePermissionSet(tx orm.DB, roleRequest models.Role, existingRole *models.Role) (*models.PermissionSet, error) {
	var permissionSet *models.PermissionSet
	var err error
	hasPermissionSet := existingRole.PermissionSets != nil && len(*existingRole.PermissionSets) != 0
	if hasPermissionSet {
		permissionSet = &(*existingRole.PermissionSets)[0]
	} else {
		var tenantId *string = nil
		if existingRole.ScopeType == models.ROLE_SCOPE_TENANT && existingRole.TenantId != nil {
			tenantId = existingRole.TenantId
		}
		permissionSet, err = createInternalPermissionSet(tx, roleRequest.Name, []models.RolePermission{}, tenantId)
		if err != nil {
			return nil, err
		}
		existingRole.PermissionSets = &[]models.PermissionSet{*permissionSet}
	}
	return permissionSet, nil
}

func upsertPermissionSetAndPermissions(ctx context.Context, tx orm.DB, roleRequest models.Role, existingRole *models.Role) error {
	permissionSet, err := getOrCreatePermissionSet(tx, roleRequest, existingRole)
	if err != nil {
		return err
	}
	permissions := rolePermissionsFromPermissions(roleRequest.Permissions, &(permissionSet.ID))
	if len(permissions) > 0 {
		_, err := tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "permission_set_id"}, {Name: "resource_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"create", "read", "update", "delete", "sync"}),
			}).
			Create(&permissions)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to save permissions for role %s. err: %+v", roleRequest.Name, err)
			return err
		}
	}
	// Delete permissions that are no longer needed
	permissionsToDelete := findRolePermissionsToDelete(api.AllowedResourceTypeEnumValues, permissions, permissionSet.ID)
	for _, permission := range permissionsToDelete {
		_, err := tx.
			Where("permission_set_id = ?", permission.PermissionSetID).
			Where("resource_type = ?", permission.ResourceType).
			Delete(&permission)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete permission of role id %d on resource type %s. err: %+v", existingRole.ID, permission.ResourceType, err)
			return err
		}
	}

	return nil
}

func getInternalPermissionSetNameFromRole(roleName string, tenantId *string) string {
	if tenantId != nil {
		return fmt.Sprintf("%s-%s-%s", *tenantId, roleName, InternalPermissionSetSuffix)
	}
	return fmt.Sprintf("%s-%s", roleName, InternalPermissionSetSuffix)
}

func createInternalPermissionSet(tx orm.DB, roleName string, permissions []models.RolePermission, tenantId *string) (*models.PermissionSet, error) {
	permissionSet := &models.PermissionSet{
		Name:        getInternalPermissionSetNameFromRole(roleName, tenantId),
		Internal:    true,
		Permissions: permissions,
	}
	_, err := tx.Create(permissionSet)
	return permissionSet, err
}

func rolePermissionsFromPermissions(permissions []*models.Permission, permissionSetId *uuid.UUID) []models.RolePermission {
	rolePermissions := []models.RolePermission{}
	for _, permission := range permissions {
		rolePermission := models.RolePermission{
			ResourceType:           permission.ResourceType,
			Create:                 permission.Create,
			Read:                   permission.Read,
			Update:                 permission.Update,
			Delete:                 permission.Delete,
			Sync:                   permission.Sync,
			ReadOnlyForDGXCManaged: permission.ReadOnlyForDGXCManaged,
			ReadOnlyForCSP:         permission.ReadOnlyForCSP,
		}
		if permissionSetId != nil {
			rolePermission.PermissionSetID = *permissionSetId
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}
	return rolePermissions
}

func findPermissionsToDelete(requestedResourceTypes []api.ResourceType, currentPermissions []*models.Permission) []models.Permission {
	permissionsToDelete := []models.Permission{}
	permissionsMap := map[string]bool{}

	for _, requestedResourceType := range requestedResourceTypes {
		permissionsMap[string(requestedResourceType)] = true
	}
	for _, currentPermission := range currentPermissions {
		if _, found := permissionsMap[string(currentPermission.ResourceType)]; !found {
			permissionsToDelete = append(permissionsToDelete, models.Permission{
				RoleID:       currentPermission.RoleID,
				ResourceType: currentPermission.ResourceType,
			})
		}
	}

	return permissionsToDelete
}

func findRolePermissionsToDelete(resourceTypes []api.ResourceType, rolePermissions []models.RolePermission, permissionSetId uuid.UUID) []models.RolePermission {
	permissionsToDelete := []models.RolePermission{}
	permissionsMap := map[string]bool{}

	for _, rolePermission := range rolePermissions {
		permissionsMap[string(rolePermission.ResourceType)] = true
	}
	for _, resourceType := range resourceTypes {
		if _, found := permissionsMap[string(resourceType)]; !found {
			permissionsToDelete = append(permissionsToDelete, models.RolePermission{
				PermissionSetID: permissionSetId,
				ResourceType:    resourceType,
			})
		}
	}

	return permissionsToDelete
}

func DbUpdateRole(ctx context.Context, tenantId string, roleId int32, roleSpec models.Role) (*models.Role, *api.Error) {
	if config.Get[*authorization_config.General]().UsePermissionSets {
		return DbUpdateRolePS(ctx, tenantId, roleId, roleSpec)
	}
	return DbUpdateRoleNoPS(ctx, tenantId, roleId, roleSpec)
}

func DbUpdateRoleNoPS(ctx context.Context, tenantId string, roleId int32, roleSpec models.Role) (*models.Role, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to update role", nil)

	existingRole, err := dbGetRoleById(ctx, tenantId, uint(roleId), nil, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to retrieve role %d. err: %+v", roleId, err)
		return nil, err
	}

	setStaticData(&roleSpec, existingRole)

	transactionErr := runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Transaction(func(tx orm.DB) error {
		if _, err := tx.
			Where("tenant_id = ?", tenantId).
			Where("id = ?", roleId).
			Save(&roleSpec); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update role %d: %v", roleId, err)
			return err
		}

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
	if transactionErr != nil {
		return nil, dbErr.WithDetails(transactionErr)
	}

	return &roleSpec, nil
}

func getRoleForUpdateAndValidate(ctx context.Context, tenantId string, roleId int32, roleSpec models.Role) (*models.Role, *api.Error) {
	existingRole, err := dbGetRoleById(ctx, tenantId, uint(roleId), nil, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to retrieve role %d. err: %+v", roleId, err)
		return nil, err
	}
	if !*existingRole.NoPermissionSet && len(roleSpec.Permissions) > 0 {
		return nil,
			api.NewError(http.StatusBadRequest,
				"Failed to update role %s. cannot update role created with permission sets to be with permissions", nil)
	}
	return existingRole, nil
}

func DbUpdateRolePS(ctx context.Context, tenantId string, roleId int32, roleSpec models.Role) (*models.Role, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to update role", nil)

	existingRole, err := getRoleForUpdateAndValidate(ctx, tenantId, roleId, roleSpec)
	if err != nil {
		return nil, err
	}

	transactionErr := runaidb.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Transaction(func(tx orm.DB) error {
		setStaticData(&roleSpec, existingRole)
		deleteInternalPermissionSet := false
		if *existingRole.NoPermissionSet {
			if roleSpec.PermissionSets != nil && len(*roleSpec.PermissionSets) > 0 {
				existingRole.NoPermissionSet = ptr.To(false)
				deleteInternalPermissionSet = true
			} else {
				permissionSet := (*existingRole.PermissionSets)[0]
				// update the internal permission set - since it's a "dummy" one we only change it
				updatedPermissionSet := models.PermissionSet{
					DeleteBaseModel: permissionSet.DeleteBaseModel,
					Name:            permissionSet.Name,
					Description:     permissionSet.Description,
					Internal:        permissionSet.Internal,
					Permissions:     rolePermissionsFromPermissions(roleSpec.Permissions, &(permissionSet.ID)),
				}
				if _, err := tx.
					Where("id = ?", permissionSet.ID).
					Save(&updatedPermissionSet); err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update permission sets for role %d: %v", roleId, err)
					return err
				}
				roleSpec.PermissionSets = &[]models.PermissionSet{updatedPermissionSet}
			}
		} else {
			// Update the role's permission sets association (replaces old with new)
			if err := tx.Actual().Model(existingRole).Association(PermissionSetsField).Clear(); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update permission sets for role %d: %v", roleId, err)
				return err
			}
		}

		if _, err := tx.
			Where("tenant_id = ?", tenantId).
			Where("id = ?", roleId).
			Omit(fmt.Sprintf("%s.*", PermissionSetsField)).
			Select("*").
			Updates(&roleSpec); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update role %d: %v", roleId, err)
			return err
		}

		if *roleSpec.NoPermissionSet {
			if err := deleteInternalPermissions(ctx, tx, &roleSpec); err != nil {
				return err
			}
		} else {
			if deleteInternalPermissionSet {
				// role was with internal permission set and now updated to be with external permission sets
				if err := deleteInternalPS(tx, tenantId, roleSpec.Name); err != nil {
					return err
				}
			}
			if err := tx.
				Preload("PermissionSets").
				Preload("PermissionSets.Permissions").
				First(&roleSpec, roleSpec.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if transactionErr != nil {
		return nil, dbErr.WithDetails(transactionErr)
	}

	return &roleSpec, nil
}

func deleteInternalPermissions(ctx context.Context, tx orm.DB, existingRole *models.Role) error {
	permissionsToDelete := findRolePermissionsToDelete(api.AllowedResourceTypeEnumValues, (*existingRole.PermissionSets)[0].Permissions, (*existingRole.PermissionSets)[0].ID)
	for _, permission := range permissionsToDelete {
		_, err := tx.
			Where("permission_set_id = ?", (*existingRole.PermissionSets)[0].ID).
			Where("resource_type = ?", permission.ResourceType).
			Delete(&permission)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete permission of role id %d on resource type %s. err: %+v", existingRole.ID, permission.ResourceType, err)
			return err
		}
	}
	return nil
}

func deleteInternalPS(tx orm.DB, tenantId string, roleName string) error {
	var permTenantId *string = nil
	if tenantId != "" {
		permTenantId = &tenantId
	}
	_, err := tx.
		Where("name = ?", getInternalPermissionSetNameFromRole(roleName, permTenantId)).
		Where("internal = true").
		Delete(&models.PermissionSet{})
	return err
}

func DbSetRoleEnabled(ctx context.Context, tenantId string, roleId int32, enabled bool) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to set role enabled", nil)

	tx := runaidb.DB().WithContext(ctx)

	_, err := tx.
		Table(RolesTable).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", roleId).
		Update("enabled", enabled)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set role %d enabled to %v: %v", roleId, enabled, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func setStaticData(roleSpec *models.Role, existingRole *models.Role) {
	roleSpec.ID = existingRole.ID
	roleSpec.CreatedBy = existingRole.CreatedBy
	roleSpec.CreatedAt = existingRole.CreatedAt
	roleSpec.TenantId = existingRole.TenantId
	roleSpec.InternalOnly = existingRole.InternalOnly
	roleSpec.Custom = existingRole.Custom
	roleSpec.ShowForDGXC = existingRole.ShowForDGXC
	roleSpec.RoleType = existingRole.RoleType
	roleSpec.NoPermissionSet = existingRole.NoPermissionSet
}

func DbCreateRole(ctx context.Context, subjectId string, roleSpec models.Role) (*models.Role, *api.Error) {
	var dbErr *api.Error

	roleSpec.CreatedBy = subjectId

	tx := runaidb.DB().WithContext(ctx)
	var resultErr error
	if config.Get[*authorization_config.General]().UsePermissionSets && len(roleSpec.Permissions) > 0 &&
		(roleSpec.PermissionSets == nil || len(*roleSpec.PermissionSets) == 0) {

		roleSpec.NoPermissionSet = ptr.To(true)
		rolePermissions := []models.RolePermission{}
		for _, permission := range roleSpec.Permissions {
			rolePermissions = append(rolePermissions, models.RolePermission{
				ResourceType: permission.ResourceType,
				Create:       permission.Create,
				Read:         permission.Read,
				Update:       permission.Update,
				Delete:       permission.Delete,
				Sync:         permission.Sync,
			})
		}
		var tenantId *string = nil
		if roleSpec.ScopeType == models.ROLE_SCOPE_TENANT && roleSpec.TenantId != nil {
			tenantId = roleSpec.TenantId
		}
		permissionSet := &models.PermissionSet{
			Name:        getInternalPermissionSetNameFromRole(roleSpec.Name, tenantId),
			Internal:    true,
			Permissions: rolePermissions,
		}
		roleSpec.PermissionSets = &[]models.PermissionSet{*permissionSet}
		roleSpec.Permissions = nil
		_, resultErr = tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&roleSpec)
	} else {
		roleSpec.NoPermissionSet = ptr.To(false)
		_, resultErr = tx.Omit(fmt.Sprintf("%s.*", PermissionSetsField)).Create(&roleSpec)
	}

	if resultErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create role %s: %v", roleSpec.Name, resultErr)

		// Check for constraint violation errors (unique/exclusion). runaidb TranslateError maps
		// unique violations (23505) to orm.ErrDuplicatedKey; exclusion (23P01) stays as PgError.
		if isDuplicatedKeyError(resultErr) {
			dbErr = api.NewError(http.StatusConflict, "Role with this name already exists", nil)
		} else {
			dbErr = api.NewError(http.StatusInternalServerError, "Failed to create role", nil)
		}
		return nil, dbErr.WithDetails(resultErr)
	}

	// Reload the role with permission sets and their permissions populated
	if config.Get[*authorization_config.General]().UsePermissionSets && !*roleSpec.NoPermissionSet {
		reloadErr := tx.
			Preload("PermissionSets").
			Preload("PermissionSets.Permissions").
			First(&roleSpec, roleSpec.ID)
		if reloadErr != nil {
			dbErr = api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to reload role %s after creation: %v", roleSpec.Name, reloadErr), nil)
			return nil, dbErr.WithDetails(reloadErr)
		}
	}
	return &roleSpec, nil
}

func DbDeleteRole(ctx context.Context, tenantId string, roleId uint) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to delete role", nil)

	tx := runaidb.DB().WithContext(ctx)

	role, err := DbGetRoleById(ctx, tenantId, roleId, nil)
	if err != nil {
		return err
	}

	if config.Get[*authorization_config.General]().UsePermissionSets {
		clearErr := tx.Actual().Model(&role).Association(PermissionSetsField).Clear()
		if clearErr != nil {
			return api.NewError(http.StatusInternalServerError, "Failed to delete role", clearErr)
		}
	}
	rowsAffected, deleteErr := tx.
		Where("tenant_id = ?", tenantId).
		Delete(&models.Role{}, roleId)
	if deleteErr != nil || rowsAffected != 1 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete role from db: error= %v, rowsAffected= %d", deleteErr, rowsAffected)
		if deleteErr != nil {
			return dbErr.WithDetails(deleteErr)
		}
		if rowsAffected == 0 {
			dbErr.Code = http.StatusNotFound
			dbErr.Details = errDetailsResourceNotFound
		}
		return dbErr
	}
	if config.Get[*authorization_config.General]().UsePermissionSets && *role.NoPermissionSet {
		_, err := tx.
			Where("name = ?", getInternalPermissionSetNameFromRole(role.Name, &tenantId)).
			Where("internal = true").
			Delete(&models.PermissionSet{})
		if err != nil {
			return dbErr.WithDetails(err)
		}
	}

	return nil
}

func DbGetRoles(
	ctx context.Context,
	tenantId string,
	isDGXC bool,
	limit int32,
	offset int32,
	sortOrder string,
	sortBy string,
	filterBy []queryapi.Filter,
	search string,
	includeInternal bool,
) (*[]models.Role, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get roles", nil)
	roleTypeFilter := getRoleTypeFilter(&isDGXC)

	tx := runaidb.DB().WithContext(ctx)
	roles := &[]models.Role{}
	var query orm.DB
	if config.Get[*authorization_config.General]().UsePermissionSets {
		query = tx.
			Preload("PermissionSets").
			Preload("PermissionSets.Permissions").
			Where("(scope_type = ?) OR (scope_type = ? AND scope_id = ?)", models.ROLE_SCOPE_SYSTEM, models.ROLE_SCOPE_TENANT, tenantId).
			Where("role_type <> ?", roleTypeFilter)
	} else {
		query = tx.
			Preload("Permissions").
			Where("(scope_type = ?) OR (scope_type = ? AND scope_id = ?)", models.ROLE_SCOPE_SYSTEM, models.ROLE_SCOPE_TENANT, tenantId).
			Where("role_type <> ?", roleTypeFilter)
	}
	query = query.WithEnricher(querydbenrich.Filters(filterBy))
	if !includeInternal {
		query = query.Where("internal_only = false")
	}

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id DESC")

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	err := query.
		Find(roles)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get roles from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	disabledRoles, apiErr := DbGetRoleDisabledPerTenant(ctx, tx, tenantId)
	if apiErr != nil {
		return nil, apiErr
	}

	for i := range *roles {
		if _, ok := disabledRoles[int64((*roles)[i].ID)]; ok {
			(*roles)[i].EffectiveEnabled = ptr.To(false)
		} else {
			(*roles)[i].EffectiveEnabled = ptr.To((*roles)[i].Enabled)
		}
		if config.Get[*authorization_config.General]().UsePermissionSets {
			(*roles)[i].Permissions = getAggregatedPermissions((*roles)[i].PermissionSets)
		}
	}

	return roles, nil
}

func getAggregatedPermissions(permissionSets *[]models.PermissionSet) []*models.Permission {
	resourceTypePermissionMap := map[api.ResourceType]models.Permission{}
	for _, permissionSet := range *permissionSets {
		for _, permission := range permissionSet.Permissions {
			resourceTypePerm, exists := resourceTypePermissionMap[permission.ResourceType]
			if !exists {
				resourceTypePerm = models.Permission{
					ResourceType: permission.ResourceType,
					Create:       false,
					Read:         false,
					Update:       false,
					Delete:       false,
					Sync:         false,
				}
			}
			if permission.Read {
				resourceTypePerm.Read = true
			}
			if permission.Create {
				resourceTypePerm.Create = true
			}
			if permission.Update {
				resourceTypePerm.Update = true
			}
			if permission.Delete {
				resourceTypePerm.Delete = true
			}
			if permission.Sync {
				resourceTypePerm.Sync = true
			}
			resourceTypePermissionMap[permission.ResourceType] = resourceTypePerm
		}
	}
	aggregatedPermissions := []*models.Permission{}
	for _, permission := range resourceTypePermissionMap {
		aggregatedPermissions = append(aggregatedPermissions, &permission)
	}
	return aggregatedPermissions
}

func DbGetRoleById(ctx context.Context, tenantId string, roleId uint, isDGXC *bool) (*models.Role, *api.Error) {
	return dbGetRoleById(ctx, tenantId, roleId, ptr.To(false), isDGXC)
}

func DbGetRoleByIdWithInternal(ctx context.Context, tenantId string, roleId uint, isDGXC *bool) (*models.Role, *api.Error) {
	return dbGetRoleById(ctx, tenantId, roleId, nil, isDGXC)
}

func dbGetRoleById(ctx context.Context, tenantId string, roleId uint, internalOnly *bool, isDGXC *bool) (*models.Role, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get role", nil)
	roleTypeFilter := getRoleTypeFilter(isDGXC)

	tx := runaidb.DB().WithContext(ctx)
	role := &models.Role{}
	var query orm.DB
	if config.Get[*authorization_config.General]().UsePermissionSets {
		query = tx.
			Preload("PermissionSets").
			Preload("PermissionSets.Permissions").
			Where("(scope_type = ?) OR (scope_type = ? AND scope_id = ?)", models.ROLE_SCOPE_SYSTEM, models.ROLE_SCOPE_TENANT, tenantId).
			Where("role_type <> ?", roleTypeFilter)
	} else {
		query = tx.
			Preload("Permissions").
			Where("(scope_type = ?) OR (scope_type = ? AND scope_id = ?)", models.ROLE_SCOPE_SYSTEM, models.ROLE_SCOPE_TENANT, tenantId).
			Where("role_type <> ?", roleTypeFilter)
	}
	if internalOnly != nil {
		query = query.
			Where("internal_only = ?", *internalOnly)
	}

	err := query.First(role, roleId)
	if err != nil {
		if !ErrRecordNotFound(err) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to get role with id %d", roleId), err)
		}
		return nil, dbErr.WithDetails(err)
	}

	dr, apiErr := DbGetRoleDisabledById(ctx, tx, tenantId, roleId)
	if apiErr != nil {
		return nil, apiErr
	}

	if dr.RoleId == int64(roleId) {
		role.EffectiveEnabled = ptr.To(false)
	}
	if config.Get[*authorization_config.General]().UsePermissionSets {
		role.Permissions = getAggregatedPermissions(role.PermissionSets)
	}

	return role, nil
}

// Its a filter-out filter, since we have 3 types. All, DGXC and NonDGXC then we need to filter out a single type each time
func getRoleTypeFilter(isDGXC *bool) (roleTypeFilter string) {
	if isDGXC != nil {
		if *isDGXC {
			roleTypeFilter = string(models.NonDGXC)
		} else {
			roleTypeFilter = string(models.DGXC)
		}
	}

	return roleTypeFilter
}

func IsRoleInUse(ctx context.Context, roleId uint) (bool, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to check if role in use", nil)

	count := int64(0)
	err := runaidb.DB().WithContext(ctx).
		Table(AccessRulesTableName).
		Where("role_id = ?", roleId).
		Where("deleted_at IS NULL").
		Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to check if roleId %d in use", roleId), err)
		return false, dbErr.WithDetails(err)
	}

	return count > 0, nil
}

func IsRoleInScopeType(ctx context.Context, roleId uint, scopyType string) (bool, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to check if role in use", nil)

	count := int64(0)
	err := runaidb.DB().WithContext(ctx).
		Table(RolesTable).
		Where("id = ?", roleId).
		Where("scope_type = ?", scopyType).
		Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(fmt.Sprintf("Failed to check if roleId %d in use", roleId), err)
		return false, dbErr.WithDetails(err)
	}

	return count > 0, nil
}

func isDuplicatedKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.Is(err, orm.ErrDuplicatedKey) ||
		(errors.As(err, &pgErr) && pgErr.Code == ExclusionViolationErrorCode)
}
