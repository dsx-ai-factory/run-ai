/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

const (
	AccessRulesTableName     = "access_rules"
	UniqueViolationErrorCode = "23505"
	GroupsCacheHashField     = "groups"

	MigrationStepCreate = "create"
	MigrationStepDelete = "delete"

	errDetailsResourceNotFound = "requested resource does not exist"
)

const systemAdminDeleteQuery = `WITH admins_to_delete AS (
				SELECT access_rules.id FROM access_rules join roles on access_rules.role_id = roles.id WHERE access_rules.id in ? and role_id = ? and access_rules.tenant_id = ?
			),
			admins_remaining AS (
				SELECT * FROM access_rules join roles on access_rules.role_id = roles.id WHERE access_rules.id NOT IN ? and role_id = ? and access_rules.tenant_id = ? and access_rules.deleted_at is null
			)
			SELECT id AS ids_to_delete FROM admins_to_delete
			WHERE NOT EXISTS (SELECT 1 FROM admins_remaining);`

func DbCreateAccessRule(
	ctx context.Context,
	tenantId uint,
	subjectId string,
	scopeName string,
	clusterUuid string,
	usePhase bool,
	accessRuleRequest api.AccessRuleCreationFields,
) (*models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create access rule", nil)

	accessRuleSpec := CreationRequestToAccessRuleDao(accessRuleRequest)
	accessRuleSpec.TenantId = tenantId
	accessRuleSpec.CreatedBy = subjectId
	accessRuleSpec.ScopeName = scopeName
	accessRuleSpec.ClusterUuid = clusterUuid
	accessRuleSpec.UpdatedAt = time.Now()

	if usePhase {
		accessRuleSpec.Phase = string(api.ACCESSRULEPHASE_CREATING)
	}

	tx := runaidb.DB().WithContext(ctx)

	_, err := tx.
		Omit(clause.Associations).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "subject_id"}, {Name: "role_id"}, {Name: "scope_id"}, {Name: "scope_type"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": accessRuleSpec.UpdatedAt}),
			TargetWhere: clause.Where{
				Exprs: []clause.Expression{
					clause.Expr{SQL: "deleted_at IS NULL", Vars: []interface{}{}},
				},
			},
		}).Create(&accessRuleSpec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to create access rule of subject %s, scope %s and role id %d: error= %v",
			accessRuleSpec.SubjectId, accessRuleSpec.ScopeName, accessRuleSpec.RoleID, err)
		return nil, dbErr.WithDetails(err)
	}

	if accessRuleSpec.SubjectType == api.SUBJECTTYPEEXTENDED_GROUP {
		cache.Del[[]string](ctx, groupsCacheKey(tenantId))
	}

	role, apiErr := dbGetRoleById(ctx, fmt.Sprintf("%d", tenantId), accessRuleSpec.RoleID, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	accessRuleSpec.Role = *role

	return &accessRuleSpec, nil
}

func DbDeleteAccessRule(ctx context.Context, tenantId uint, accessRuleId uint) *api.Error {
	return DbBatchDeleteAccessRules(ctx, tenantId, []uint{accessRuleId})
}

func DbBatchDeleteAccessRules(ctx context.Context, tenantId uint, accessRuleIds []uint) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete access rules with ids: %v", accessRuleIds), nil)

	var rowsAffected int64
	var rowsMismatch bool
	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		var err error
		rowsAffected, err = tx.
			Where("tenant_id = ?", tenantId).
			Where("id IN ?", accessRuleIds).
			Delete(&models.AccessRule{})
		if err != nil {
			return err
		}
		if rowsAffected != int64(len(accessRuleIds)) {
			rowsMismatch = true
			return fmt.Errorf("expected to delete %d access rules, deleted %d", len(accessRuleIds), rowsAffected)
		}
		return nil
	})
	if err == nil {
		return nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete access rules from db: %v", err)
	if rowsMismatch {
		if rowsAffected == 0 {
			dbErr.Code = http.StatusNotFound
			dbErr.Details = errDetailsResourceNotFound
		}
		return dbErr
	}
	return dbErr.WithDetails(err)
}

// DbGetDeprecatedAccessRulesForMigration returns up to `limit` live access rules for (tenantId, roleId).
// No ORDER BY — Postgres will early-terminate once `limit` matching rows are found, avoiding a full
// scan + sort. The result order is unspecified, which is fine: migrated rules are soft-deleted, so
// subsequent calls naturally see a shrinking set and drain the deprecated role over repeated calls
// without any explicit position tracking. System-scope rules are excluded defensively — this
// endpoint is tenant-scoped and shouldn't touch system-level access rules even if one is somehow
// attached to the deprecated role.
func DbGetDeprecatedAccessRulesForMigration(ctx context.Context, tenantId uint, roleId uint, limit int) ([]models.AccessRule, *api.Error) {
	var rules []models.AccessRule
	err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantId).
		Where("role_id = ?", roleId).
		Where("scope_type != ?", api.SCOPETYPE_SYSTEM).
		Limit(limit).
		Find(&rules)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch deprecated access rules for migration (role %d): %v", roleId, err)
		return nil, api.NewError(http.StatusInternalServerError, "Failed to fetch access rules for migration", nil).WithDetails(err)
	}
	return rules, nil
}

// DbMigrateSingleAccessRule migrates one access rule from an old (deprecated) role to a new role
// inside a single transaction. If a live access rule already exists for (subject, newRoleId, scope),
// no new rule is created and the pre-existing rule is returned with preExisted=true. The old rule is
// soft-deleted. If any step fails, the transaction is rolled back (old rule remains alive).
//
// Return values:
//   - migratedAccessRule: the live new-role rule (either freshly created or pre-existing). nil on error.
//   - preExisted: true iff migratedAccessRule was pre-existing (not created by this call).
//   - step: on error, MigrationStepCreate if create/find failed, MigrationStepDelete if delete failed. Empty on success.
//   - apiErr: the error, if any.
func DbMigrateSingleAccessRule(ctx context.Context, tenantId uint, oldRule *models.AccessRule, newRoleId uint, subjectId string) (migratedAccessRule *models.AccessRule, preExisted bool, step string, apiErr *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to migrate access rule %d", oldRule.ID), nil)

	// Try to create the new access rule first, as a single "INSERT ... ON CONFLICT DO NOTHING"
	// statement. Happy-path cost: 1 SQL round-trip. On conflict with an existing live rule for
	// (subject, new role, scope), the insert is a no-op (RowsAffected == 0) and the transaction
	// stays alive — we fetch the pre-existing rule afterwards. This avoids the "SELECT-then-INSERT"
	// pattern, which pays for 2 round-trips on every call.
	candidate := models.AccessRule{
		SubjectId:   oldRule.SubjectId,
		SubjectType: oldRule.SubjectType,
		RoleID:      newRoleId,
		ScopeId:     oldRule.ScopeId,
		ScopeType:   oldRule.ScopeType,
		ScopeName:   oldRule.ScopeName,
		TenantId:    tenantId,
		ClusterUuid: oldRule.ClusterUuid,
	}
	candidate.CreatedBy = subjectId
	step = MigrationStepCreate
	var deleteRowsAffected int64
	var deleteRowsMismatch bool
	transactionErr := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		rowsAffected, err := tx.
			Omit(clause.Associations).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "subject_id"}, {Name: "role_id"}, {Name: "scope_id"}, {Name: "scope_type"}},
				TargetWhere: clause.Where{
					Exprs: []clause.Expression{clause.Expr{SQL: "deleted_at IS NULL", Vars: []interface{}{}}},
				},
				DoNothing: true,
			}).
			Create(&candidate)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create new access rule while migrating rule %d: %v", oldRule.ID, err)
			return err
		}

		if rowsAffected == 0 {
			var existing models.AccessRule
			if err := tx.
				Where("tenant_id = ?", tenantId).
				Where("subject_id = ?", oldRule.SubjectId).
				Where("role_id = ?", newRoleId).
				Where("scope_id = ?", oldRule.ScopeId).
				Where("scope_type = ?", oldRule.ScopeType).
				First(&existing); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to look up pre-existing new-role access rule while migrating rule %d: %v", oldRule.ID, err)
				return err
			}
			preExisted = true
			migratedAccessRule = &existing
		} else {
			preExisted = false
			migratedAccessRule = &candidate
		}

		step = MigrationStepDelete
		deleteRowsAffected, err = tx.
			Where("tenant_id = ?", tenantId).
			Where("id = ?", oldRule.ID).
			Delete(&models.AccessRule{})
		if err != nil {
			return err
		}
		if deleteRowsAffected != 1 {
			deleteRowsMismatch = true
			return fmt.Errorf("expected to delete access rule %d, deleted %d rows", oldRule.ID, deleteRowsAffected)
		}
		return nil
	})
	if transactionErr != nil {
		if deleteRowsMismatch {
			if deleteRowsAffected == 0 {
				dbErr.Code = http.StatusNotFound
				dbErr.Details = errDetailsResourceNotFound
			}
			return nil, false, step, dbErr
		}
		return nil, false, step, dbErr.WithDetails(transactionErr)
	}

	return migratedAccessRule, preExisted, "", nil
}

func DbGetAccessRulesByIds(
	ctx context.Context,
	tenantId uint,
	accessRulesIds []string,
) (*[]models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get access rules", nil)

	accessRules := &[]models.AccessRule{}
	resultErr := runaidb.DB().WithContext(ctx).
		Preload("Role.PermissionSets").
		Preload("Role.PermissionSets.Permissions").
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.id IN ?", accessRulesIds).
		Find(accessRules)
	if resultErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get access rules from db: %v", resultErr)
		return nil, dbErr.WithDetails(resultErr)
	}
	return accessRules, nil
}

func updateQueryByPermittedScopes(db orm.DB, permittedScopes openapi.PermittedScopes) orm.DB {
	if permittedScopes.System != nil && *permittedScopes.System {
		return db
	}
	if len(permittedScopes.Tenants) == 0 && len(permittedScopes.Clusters) == 0 &&
		len(permittedScopes.Departments) == 0 && len(permittedScopes.Projects) == 0 {
		// No permitted scopes so return query that yields no rows
		return db.Where("1 = 0")
	}

	// Build OR conditions properly
	var conditions []string
	var args []interface{}

	if len(permittedScopes.Tenants) > 0 {
		conditions = append(conditions, "access_rules.tenant_id IN ?")
		args = append(args, permittedScopes.Tenants)
	}
	if len(permittedScopes.Clusters) > 0 {
		conditions = append(conditions, "access_rules.cluster_uuid IN ?")
		args = append(args, permittedScopes.Clusters)
	}
	if len(permittedScopes.Projects) > 0 {
		conditions = append(conditions, "access_rules.scope_type = ? AND access_rules.scope_id IN ?")
		args = append(args, string(api.SCOPETYPE_PROJECT), permittedScopes.Projects)
	}
	if len(permittedScopes.Departments) > 0 {
		conditions = append(conditions, "access_rules.scope_type = ? AND access_rules.scope_id IN ?")
		args = append(args, string(api.SCOPETYPE_DEPARTMENT), permittedScopes.Departments)
	}

	if len(conditions) == 0 {
		return db.Where("1 = 0")
	}

	orClause := "(" + strings.Join(conditions, " OR ") + ")"
	return db.Where(orClause, args...)
}

func DbGetAccessRules(
	ctx context.Context,
	tenantId uint,
	permittedScopes *openapi.PermittedScopes,
	subjectType string,
	subjectIds []string,
	limit int32,
	offset int32,
	lastUpdated string,
	includeDeleted bool,
	clusterId string,
	scopeId string,
	sortOrder string,
	sortBy string,
	filterBy []queryapi.Filter,
	search string,
) (*[]models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get access rules", nil)

	accessRules := &[]models.AccessRule{}
	db := runaidb.DB().WithContext(ctx).Preload(clause.Associations)
	query := db.
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("roles.internal_only = false").
		Where("access_rules.tenant_id = ?", tenantId)

	if subjectType != "" {
		query = query.Where("LOWER(subject_type) = LOWER(?)", subjectType)
	}
	if subjectIds != nil {
		query = query.Where("subject_id IN ?", subjectIds)
	}

	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		permittedScopesQuery := updateQueryByPermittedScopes(db, *permittedScopes)
		query = query.Where(permittedScopesQuery)
	}

	if lastUpdated != "" && includeDeleted {
		query = query.Where("access_rules.updated_at >= ? OR (access_rules.deleted_at IS NOT NULL AND access_rules.deleted_at >= ?)", lastUpdated, lastUpdated)
		query = query.Unscoped()
	} else {
		if lastUpdated != "" {
			query = query.Where("access_rules.updated_at >= ?", lastUpdated)
		}
		if includeDeleted {
			query = query.Unscoped()
		}
	}
	if clusterId != "" {
		query = query.Where("cluster_uuid IN ?", []string{clusterId, ""})
	}

	if scopeId != "" {
		query = query.Where("access_rules.scope_id = ?", scopeId)
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if sortBy == "" && (offset != 0 || limit != 0) {
		// for backwards compatibility default sort is by createdAt
		sortBy = "access_rules.created_at"
	}
	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("access_rules.id DESC")

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "access_rules.search_vector", "roles.search_vector"))
	}
	err := query.
		Find(accessRules)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get access rules from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}
	return accessRules, nil
}

func DbGetSubjectPermissions(ctx context.Context, tenantId uint, subjectId string, groups []string) (*[]models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get access rules", nil)

	accessRules := &[]models.AccessRule{}
	db := runaidb.DB().WithContext(ctx)
	resultErr := db.Preload("Role.PermissionSets").
		Preload("Role.PermissionSets.Permissions").
		Where("tenant_id = ?", tenantId).
		Where(getSubjectClause(db, groups, subjectId)).
		Find(accessRules)
	if resultErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get subject %s access rules from db: %v", subjectId, resultErr)
		return nil, dbErr.WithDetails(resultErr)
	}
	return accessRules, nil
}

func DbGetGroupNames(ctx context.Context, tenantId uint) ([]string, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get group names", nil)
	GetGroupNamesInner := func() (*[]string, error) {
		accessRules := []models.AccessRule{}
		query := runaidb.DB().WithContext(ctx).
			Select("DISTINCT subject_id").
			Where("tenant_id = ?", tenantId).
			Where("LOWER(subject_type) = LOWER(?)", api.SUBJECTTYPE_GROUP)

		err := query.
			Find(&accessRules)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get access rules from db: %v", err)
			return nil, err
		}

		names := []string{}
		for _, accessRule := range accessRules {
			names = append(names, accessRule.SubjectId)
		}

		return &names, nil
	}
	names, err := cache.Get[[]string](ctx, groupsCacheKey(tenantId))
	if err == nil && names != nil {
		return *names, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss. Group names for tenant %d are not in cache, calculating it synchronously.", tenantId)
	names, err = GetGroupNamesInner()
	if err != nil || names == nil {
		return nil, dbErr.WithDetails(err)
	}
	if err = cache.Set[[]string](ctx, groupsCacheKey(tenantId), *names); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set group names in cache: %v", err)
	}

	return *names, nil
}

func DbGetAccessRulesByScopes(ctx context.Context, roles []string, tenantId uint, clusterId string, departmentIds []string, projectIds []string) (*[]models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get access rules", nil)
	accessRules := &[]models.AccessRule{}
	db := runaidb.DB().WithContext(ctx)

	query := db.Preload(clause.Associations).
		Where("access_rules.tenant_id = ?", tenantId)

	query, apiErr := filterByRoles(ctx, tenantId, query, roles)
	if apiErr != nil {
		return nil, apiErr
	}

	query = query.Where(updateScopeFilters(db, strconv.FormatUint(uint64(tenantId), 10), clusterId, departmentIds, projectIds))

	err := query.Find(accessRules)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get access rules from db: %v", err)
		return nil, dbErr.WithDetails(err)
	}
	return accessRules, nil
}

func filterByRoles(ctx context.Context, tenantId uint, query orm.DB, roleNames []string) (orm.DB, *api.Error) {
	if len(roleNames) > 0 {
		roles, err := DbGetRoles(ctx, fmt.Sprintf("%d", tenantId), 0, 0, "", "", nil, "", false)
		if err != nil {
			return nil, err
		}

		var roleIds []string
		for _, name := range roleNames {
			value, ok := getRoleId(name, roles)
			if ok {
				roleIds = append(roleIds, strconv.FormatUint(uint64(value), 10))
			}
		}

		if len(roleIds) > 0 {
			query = query.Where("role_id IN ?", roleIds)
		} else {
			query = query.Where("1 = 0")
		}
	}

	return query, nil
}

func getRoleId(roleName string, roles *[]models.Role) (uint, bool) {
	for _, role := range *roles {
		if role.Name == roleName {
			return role.ID, true
		}
	}

	return 0, false
}

func updateScopeFilters(db orm.DB, tenantId string, clusterId string, departmentIds []string, projectIds []string) orm.DB {
	orClause := db.Where("access_rules.scope_id = ? and access_rules.scope_type = ?", tenantId, api.SCOPETYPE_TENANT)
	if clusterId != "" {
		orClause = orClause.Or(db.Where("access_rules.scope_id = ? and access_rules.scope_type = ?", clusterId, api.SCOPETYPE_CLUSTER))
	}
	if departmentIds != nil {
		orClause = orClause.Or(db.Where("access_rules.scope_type = ? and access_rules.scope_id IN ?", api.SCOPETYPE_DEPARTMENT, departmentIds))
	}
	if projectIds != nil {
		orClause = orClause.Or(db.Where("access_rules.scope_type = ? and access_rules.scope_id IN ?", api.SCOPETYPE_PROJECT, projectIds))
	}
	return orClause
}

func DbGetAccessRuleById(ctx context.Context, accessRuleId uint, tenantId uint,
	permittedScopes *openapi.PermittedScopes, includeInternal bool, includeDeleted bool,
) (*models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get access rule", nil)

	accessRule := &models.AccessRule{}
	db := runaidb.DB().WithContext(ctx)
	if includeDeleted {
		db = db.Unscoped()
	}
	query := db.Preload("Role.PermissionSets").
		Preload("Role.PermissionSets.Permissions").
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("roles.internal_only IN (false, ?)", includeInternal).
		Where("access_rules.tenant_id = ?", tenantId)
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		permittedScopesQuery := updateQueryByPermittedScopes(db, *permittedScopes)
		query = query.Where(permittedScopesQuery)
	}

	err := query.First(accessRule, accessRuleId)
	if err != nil {
		if !ErrRecordNotFound(err) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get access rule with id %d. error: %v", accessRuleId, err)
		}
		return nil, dbErr.WithDetails(err)
	}
	return accessRule, nil
}

func DbCountSystemAdminAccessRules(ctx context.Context, tenantId uint, excludeIds []string) (int64, error) {
	roleId := MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name]
	var accessRulesNumber int64
	err := runaidb.DB().Model(&models.AccessRule{}).WithContext(ctx).Preload(clause.Associations).
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("roles.id = ?", roleId).
		Where("access_rules.id not in ?", excludeIds).
		Where("access_rules.deleted_at is null").
		Count(&accessRulesNumber)
	return accessRulesNumber, err
}

func DbGetIdsForLastSystemAdminDeletion(ctx context.Context, tenantId string, idsToDelete []string) ([]string, error) {
	roleId := fmt.Sprint(MapStaticRoleNameToRoleId[static_roles.SystemAdmin.Name])

	lastSystemAdminAccessRules := []string{}
	_, err := runaidb.DB().WithContext(ctx).
		Raw(systemAdminDeleteQuery, idsToDelete, roleId, tenantId, idsToDelete, roleId, tenantId).
		Scan(&lastSystemAdminAccessRules)

	return lastSystemAdminAccessRules, err
}

func DbCreateInternalUserAccessRule(
	ctx context.Context,
	request api.InternalCreateAccessRuleRequest,
	tenantId uint,
	roleId uint,
	scopeType api.ScopeType,
	scopeId string,
	scopeName string,
	clusterUuid string,
) (*models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create internal user access rule", nil)

	accessRuleSpec := models.AccessRule{
		Common: models.Common{
			CreatedBy: RunAiRoleCreator,
		},
		SubjectId:   request.SubjectId,
		SubjectType: api.SubjectTypeExtended(request.SubjectType),
		RoleID:      roleId,
		ScopeId:     scopeId,
		ScopeType:   scopeType,
		TenantId:    tenantId,
		ScopeName:   scopeName,
		ClusterUuid: clusterUuid,
	}

	tx := runaidb.DB().WithContext(ctx)
	_, err := tx.Omit(clause.Associations).Create(&accessRuleSpec)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, api.NewError(http.StatusConflict, "internal access rule already exists", nil)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to create internal access rule of subject %s, scope %s and role id %d: error= %v",
			accessRuleSpec.SubjectId, accessRuleSpec.ScopeName, accessRuleSpec.RoleID, err)
		return nil, dbErr.WithDetails(err)
	}

	role, apiErr := dbGetRoleById(ctx, fmt.Sprintf("%d", tenantId), accessRuleSpec.RoleID, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	accessRuleSpec.Role = *role

	return &accessRuleSpec, nil
}

func DbDeleteSubjectAccessRules(ctx context.Context, tenantId uint, subjectType string, subjectId string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete subject access rules of %s-%s", subjectType, subjectId), nil)

	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		_, err := tx.
			Where("tenant_id = ?", tenantId).
			Where("subject_type = ?", subjectType).
			Where("subject_id = ?", subjectId).
			Delete(&[]models.AccessRule{})
		return err
	})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete access rules from db: %v", err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbDeleteScopeAccessRules(ctx context.Context, tenantId uint, scopeType string, scopeId string, lastUpdated string, roles []string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete scope access rules of %s-%s", scopeType, scopeId), nil)

	var apiErr *api.Error
	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		query := tx.Where("tenant_id = ?", tenantId)
		if scopeType == string(api.SCOPETYPE_CLUSTER) {
			query = query.Where("cluster_uuid = ?", scopeId)
		} else {
			query = query.
				Where("access_rules.scope_type = ?", scopeType).
				Where("access_rules.scope_id = ?", scopeId)
		}
		if lastUpdated != "" {
			query = query.Where("access_rules.updated_at < ?", lastUpdated)
		}
		query, apiErr = filterByRoles(ctx, tenantId, query, roles)
		if apiErr != nil {
			return apiErr
		}
		_, err := query.Delete(&[]models.AccessRule{})
		return err
	})
	if apiErr != nil {
		return apiErr
	}
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete access rules from db: %v", err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbUpdateScopeAccessRules(ctx context.Context, tenantId uint, scopeType string, scopeId string, request api.UpdateScopeAccessRulesRequest) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update scope access rules of %s-%s", scopeType, scopeId), nil)

	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		_, err := tx.
			Model(&models.AccessRule{}).
			Where("tenant_id = ?", tenantId).
			Where("scope_type = ?", scopeType).
			Where("scope_id = ?", scopeId).
			Update("scope_name", request.ScopeName)
		return err
	})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update access rules in db: %v", err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbCountAccessRules(ctx context.Context, tenantId uint, deleted bool, filterBy []queryapi.Filter, search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count access rules", nil)

	tx := runaidb.DB().WithContext(ctx)

	query := tx.Table(AccessRulesTableName).Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("roles.internal_only = false").Where("access_rules.tenant_id = ?", tenantId)

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "access_rules.search_vector", "roles.search_vector"))
	}

	var count int64
	err := query.
		Count(&count)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count access rules, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func DbGetMeAccessRules(ctx context.Context, tenantId uint, subjectId string, groups []string, limit int32, offset int32, search string) (*[]models.AccessRule, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get subject %s access rules", subjectId), nil)

	accessRules := &[]models.AccessRule{}
	db := runaidb.DB().WithContext(ctx)
	query := db.Table(AccessRulesTableName).Preload(clause.Associations).
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("roles.internal_only = false").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.deleted_at is null").
		Where(getSubjectClause(db, groups, subjectId)).
		Order("id DESC")

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "access_rules.search_vector", "roles.search_vector"))
	}

	err := query.Find(accessRules)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get subject %s access rules from db: %v", subjectId, err)
		return nil, dbErr.WithDetails(err)
	}
	return accessRules, nil
}

func DbCountMeAccessRules(ctx context.Context, tenantId uint, subjectId string, groups []string, search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to count subject %s access rules", subjectId), nil)

	db := runaidb.DB().WithContext(ctx)
	query := db.Table(AccessRulesTableName).
		Joins("LEFT JOIN roles ON roles.id = access_rules.role_id").
		Where("roles.internal_only = false").
		Where("access_rules.tenant_id = ?", tenantId).
		Where("access_rules.deleted_at is null").
		Where(getSubjectClause(db, groups, subjectId))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "access_rules.search_vector", "roles.search_vector"))
	}

	var count int64
	err := query.
		Count(&count)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count subject %s access rules, err: %v", subjectId, err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func groupsCacheKey(tenantId uint) string {
	return fmt.Sprintf("%d-%s", tenantId, GroupsCacheHashField)
}
