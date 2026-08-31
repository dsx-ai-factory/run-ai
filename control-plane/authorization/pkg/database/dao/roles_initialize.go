/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	testing "github.com/run-ai/runai/control-plane/authorization/pkg/test"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	static_legacy_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const (
	SuperAdmins = "Super-Admins"
)

var (
	MapStaticRoleNameToRoleId = map[string]uint{}
	// roles which will be deprecated:
	LegacyRoles = []models.Role{
		static_legacy_roles.ResearchManager,
		static_legacy_roles.Researcher,
		static_legacy_roles.MLEngineer,
		static_legacy_roles.ResearcherL1,
		static_legacy_roles.EnvironmentAdmin,
		static_legacy_roles.CredentialsAdmin,
		static_legacy_roles.DataSourcesAdmin,
		static_legacy_roles.ComputeResourcesAdmin,
		static_legacy_roles.TemplatesAdmin,
		static_legacy_roles.DatavolumeAdmin,
	}
	// pre-defined roles which are constructed with permission-sets
	RolesMigrated = []models.Role{
		static_roles.SystemAdmin,
		static_roles.IntegrationMetricsCollector,
		static_roles.DepartmentAdminLegacy,
		static_roles.DepartmentAdminDgxcLegacy,
		static_roles.Editor,
		static_roles.Viewer,
		static_roles.DepartmentViewer,
		static_roles.CloudOperator,
		static_roles.ClusterApp,
		static_roles.InstallerApp,
	}
	// pre-defined roles which will be deployed starting 2.24
	RolesNew = []models.Role{
		static_roles.AiPractitioner,
		static_roles.DataAndStorageAdmin,
		static_roles.ProjectAdmin,
		static_roles.DepartmentAdmin,
		static_roles.DepartmentAdminDgxc,
		static_roles.CloudOperatorDgxc,
	}
	CspRoleChange = map[string]map[string]string{
		static_roles.SystemAdmin.Name: {
			static_permission_sets.ClusterEdit.Name:       static_permission_sets.ClusterRead.Name,
			static_permission_sets.ClusterConfigEdit.Name: static_permission_sets.ClusterConfigRead.Name,
		},
	}
)

func renameRoleIfLegacyNotExists(oldName, newName string) {
	var count int64
	runaidb.DB().Model(&models.Role{}).Where("name = ? AND scope_type = ?", newName, models.ROLE_SCOPE_SYSTEM).Count(&count)
	if count > 0 {
		return
	}
	if _, err := runaidb.DB().Model(&models.Role{}).
		Where("name = ? AND scope_type = ?", oldName, models.ROLE_SCOPE_SYSTEM).
		Update("name", newName); err != nil {
		logging.Warnf("Failed to rename '%s' to '%s': %v", oldName, newName, err)
	}
}

func migrateDeprecatedRoleNames() {
	renameRoleIfLegacyNotExists("Department administrator", "Department administrator (Legacy)")
	renameRoleIfLegacyNotExists("Department administrator-DGXC", "Department administrator-DGXC (Legacy)")
}

func InitStaticLegacyRoles() {
	migrateDeprecatedRoleNames()
	logging.Infof("Initializing static legacy roles")
	initializerDao := RoleDaoInitializer{}
	dgxcRoles := []models.Role{
		static_roles.ApplicationAdmin,
		static_roles.CloudSupport,
	}

	rolesForLegacyMigrate := LegacyRoles
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		rolesForLegacyMigrate = append(rolesForLegacyMigrate, RolesMigrated...)
	}
	for _, role := range rolesForLegacyMigrate {
		var added bool
		if config.Get[*authorization_config.Server]().IsCSP {
			role.Permissions = updateCSPPermissions(role.Permissions)
		}

		dgxcRoles, added = addToDGXCRolesIfNeeded(dgxcRoles, &role)
		if added {
			role.RoleType = models.NonDGXC
		} else {
			role.RoleType = models.ALL
		}

		id, err := initializerDao.DbUpsertLegacyRole(context.Background(), role)
		if err != nil {
			logging.Warnf("Error in initializing legacy role %s. err: %+v", role.Name, err)
		}
		MapStaticRoleNameToRoleId[role.Name] = id
	}

	if !config.Get[*authorization_config.General]().UsePermissionSets {
		for _, role := range dgxcRoles {
			id, err := initializerDao.DbUpsertLegacyRole(context.Background(), role)
			if err != nil {
				logging.Warnf("Error in initializing legacy role %s. err: %+v", role.Name, err)
			}
			MapStaticRoleNameToRoleId[role.Name] = id
		}
	}

	logging.Infof("Initializing static legacy roles completed")
}

func updatePermissionSetsWithIds(role models.Role) []models.PermissionSet {
	newPermissionSets := []models.PermissionSet{}
	for _, ps := range *role.PermissionSets {
		psId, found := MapStaticPermissionSetNameToId[ps.Name]
		if !found {
			return nil
		}
		newPermissionSets = append(newPermissionSets, models.PermissionSet{
			DeleteBaseModel: dbmodels.DeleteBaseModel{
				ID: psId,
			},
			Name: ps.Name,
		})
	}
	return newPermissionSets
}

func InitStaticRoles() {
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		return
	}
	migrateDeprecatedRoleNames()
	dgxcRoles := []models.Role{
		static_roles.ApplicationAdmin,
		static_roles.CloudSupport,
	}
	logging.Infof("Initializing static roles")
	initializerDao := RoleDaoInitializer{}
	InitPermissionSetsMap(&MapStaticPermissionSetNameToId)
	rolesWithPermissionSets := append(RolesMigrated, RolesNew...)
	for _, role := range rolesWithPermissionSets {
		if config.Get[*authorization_config.Server]().IsCSP {
			changeMap, exists := CspRoleChange[role.Name]
			if exists && len(changeMap) > 0 {
				permissionSets := getCspPermissionSets(role.PermissionSets, changeMap)
				role.PermissionSets = &permissionSets
			}
		}
		newPermissionSet := updatePermissionSetsWithIds(role)
		role.PermissionSets = &newPermissionSet

		id, err := initializerDao.DbUpsertRole(context.Background(), role)
		if err != nil {
			logging.Warnf("Error in initializing role %s. err: %+v", role.Name, err)
		}
		MapStaticRoleNameToRoleId[role.Name] = id
	}
	for _, role := range dgxcRoles {
		newPermissionSet := updatePermissionSetsWithIds(role)
		role.PermissionSets = &newPermissionSet

		id, err := initializerDao.DbUpsertRole(context.Background(), role)
		if err != nil {
			logging.Warnf("Error in initializing dgxc role %s. err: %+v", role.Name, err)
		}
		MapStaticRoleNameToRoleId[role.Name] = id
	}

	logging.Infof("Initializing static roles completed")
}

func getCspPermissionSets(original *[]models.PermissionSet, changeMap map[string]string) []models.PermissionSet {
	newPermissionSets := []models.PermissionSet{}
	for i := range *original {
		newPsName, exists := changeMap[(*original)[i].Name]
		if exists && newPsName != "" {
			newPermissionSets = append(newPermissionSets, models.PermissionSet{
				Name: newPsName,
			})
		} else {
			newPermissionSets = append(newPermissionSets, (*original)[i])
		}
	}
	return newPermissionSets
}

func addToDGXCRolesIfNeeded(roles []models.Role, role *models.Role) ([]models.Role, bool) {
	role.RoleType = models.ALL
	dgxcRole := models.Role{
		ScopeType: models.ROLE_SCOPE_SYSTEM,
	}
	for _, permission := range role.Permissions {
		if permission.ReadOnlyForDGXCManaged && (permission.Create || permission.Update || permission.Delete) {
			role.RoleType = models.NonDGXC
			dgxcRole.Permissions = append(dgxcRole.Permissions, &models.Permission{
				ResourceType: permission.ResourceType,
				Read:         permission.Read,
			})
		} else {
			dgxcRole.Permissions = append(dgxcRole.Permissions, permission)
		}
	}

	// Marking the dgxcRole as DGXC and adding it to the roles to create since the base role had at least one permission that should be duplicated
	if role.RoleType == models.NonDGXC {
		dgxcRole.Name = fmt.Sprintf("%s%s", role.Name, models.DGXCManagedPostfix)
		dgxcRole.Description = role.Description
		dgxcRole.RoleType = models.DGXC
		dgxcRole.InternalOnly = role.InternalOnly
		dgxcRole.Enabled = role.Enabled

		if role.ShowForDGXC {
			dgxcRole.InternalOnly = false
		}

		roles = append(roles, dgxcRole)
	}

	return roles, role.RoleType == models.NonDGXC
}

func updateCSPPermissions(permissions []*models.Permission) []*models.Permission {
	filtered := []*models.Permission{}
	for _, permission := range permissions {
		if permission.ReadOnlyForCSP {
			filtered = append(filtered, &models.Permission{
				ResourceType: permission.ResourceType,
				Read:         true,
			})
		} else {
			filtered = append(filtered, permission)
		}
	}
	return filtered
}

func InitAccessRulesForTests(tenantId uint, managementTenantId uint) (*models.AccessRule, error) {
	var accessRuleSpec *models.AccessRule
	err := runaidb.DB().Transaction(func(tx orm.DB) error {
		accessRuleSpec = &models.AccessRule{
			SubjectId:   SuperAdmins,
			SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
			RoleID:      MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name],
			ScopeId:     fmt.Sprintf("%d", tenantId),
			ScopeType:   api.SCOPETYPE_TENANT,
			TenantId:    tenantId,
		}
		if _, err := tx.Create(accessRuleSpec); err != nil {
			return err
		}

		// create a system scope permission for tests
		accessRuleSpec = &models.AccessRule{
			SubjectId:   SuperAdmins,
			SubjectType: api.SUBJECTTYPEEXTENDED_GROUP,
			RoleID:      MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name],
			ScopeId:     "",
			ScopeType:   api.SCOPETYPE_SYSTEM,
			TenantId:    managementTenantId,
		}
		_, err := tx.Create(accessRuleSpec)
		return err
	})
	if err != nil {
		logging.Errorf("Failed to create initial access rules: %v", err)
		panic(err)
	}

	return accessRuleSpec, nil
}

func InitRolesMap(rolesMap *map[string]uint) {
	logging.Infof("Initializing roles map")
	roles, _ := DbGetRoles(context.Background(), fmt.Sprintf("%d", testing.TestTenant), testing.TestTenant == testing.DGXCTestTenant, 0, 0, "", "", nil, "", false)

	for _, role := range *roles {
		(*rolesMap)[role.Name] = role.ID
	}
}
