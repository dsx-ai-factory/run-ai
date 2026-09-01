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

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

// the user is allowed to see only system scope field group or tenant scoped field group from his own tenant
func filterFieldGroupsByTenant(res orm.DB, ctx context.Context) orm.DB {
	tenantID := auth.TenantIdOf(ctx)
	query := filterFieldGroupsByTenantQuery()
	return res.Where(query, string(api.MAPPINGSCOPE_SYSTEM), string(api.MAPPINGSCOPE_TENANT), tenantID)
}

// compose the query for filtering field groups by system OR tenant scope of the caller's tenant
func filterFieldGroupsByTenantQuery() string {
	return fmt.Sprintf("((%s = ?) OR (%s = ? AND %s = ?))",
		dbtables.MappingFieldGroupsColumnScopeType,
		dbtables.MappingFieldGroupsColumnScopeType,
		dbtables.MappingFieldGroupsColumnTenantId)
}

func CreateFieldGroup(_ context.Context, tx orm.DB, row *dbmodel.MappingFieldGroups) *api.Error {
	if row.Fields == nil {
		row.Fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}
	if row.ItemSelectors == nil {
		row.ItemSelectors = map[string]api.ItemSelectorsWithScopeValue{}
	}

	if _, err := tx.Table(dbtables.MappingFieldGroupsTable).Create(row); err != nil {
		dbErr := database.Error(database.CREATE, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to create field group: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

// FieldGroupByScope returns a field group of a given name in both tenant and system scope.
// this allows us to have them both for merging them.
type FieldGroupByScope map[api.MappingScope]dbmodel.MappingFieldGroups

// FieldGroupsByNameAndScope is the same thing, just for lists. so we have for each name
// the field groups of both scopes.
type FieldGroupsByNameAndScope map[string]FieldGroupByScope

// GetFieldGroupByName returns every visible row for this name (system and/or tenant) as a map by scope
func GetFieldGroupByName(ctx context.Context, tx orm.DB, name string) (FieldGroupByScope, *api.Error) {
	query := dbmodel.MappingFieldGroups{
		Name: name,
	}

	var rows []dbmodel.MappingFieldGroups

	res := tx.Model(&query).Where(query)

	res = filterFieldGroupsByTenant(res, ctx)

	if err := res.Find(&rows); err != nil {
		dbErr := database.Error(database.GET, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to get field group: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}
	if len(rows) == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("field group named %s not found", name), nil)
	}

	asMap := FieldGroupByScope{}

	for _, row := range rows {
		asMap[row.ScopeType] = row
	}

	return asMap, nil
}

// GetFieldGroupByNameAndScope returns one row for name and scope visible to the current tenant (same
// visibility rules as GetFieldGroupByName). Tenant scope restricts to auth.TenantIdOf(ctx), matching
// DeleteFieldGroupByNameAndScope.
func GetFieldGroupByNameAndScope(ctx context.Context, tx orm.DB, name string, scope api.MappingScope) (*dbmodel.MappingFieldGroups, *api.Error) {
	query := dbmodel.MappingFieldGroups{
		Name:      name,
		ScopeType: scope,
	}
	if scope == api.MAPPINGSCOPE_TENANT {
		query.TenantId = auth.TenantIdOf(ctx)
	}

	var row dbmodel.MappingFieldGroups

	if err := tx.Model(&query).Where(query).Take(&row); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("%s scoped field group %s not found", scope, name), nil)
		}
		dbErr := database.Error(database.GET, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to get field group: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return &row, nil
}

// ListFieldGroupNamesPage returns distinct field group names visible to the tenant, sorted by name, after offset/limit.
// Pagination applies to logical names (taking into account system and tenant scopes)
// for example, if we have
//
//	system: a1, a3, a5
//	tenant: a1, a2, a6
//
// if the caller wants (offset 1, limit 3) we should return a2, a3, a5 as the unified set is (a1, a2, a3, a5, a6)
func ListFieldGroupNamesPage(ctx context.Context, tx orm.DB, offset, limit int) ([]string, *api.Error) {
	var names []string

	res := tx.WithContext(ctx).Model(&dbmodel.MappingFieldGroups{})

	res = filterFieldGroupsByTenant(res, ctx)

	if err := res.Distinct(dbtables.MappingFieldGroupsColumnName).
		Order(dbtables.MappingFieldGroupsColumnName).
		Offset(offset).
		Limit(limit).
		Pluck(dbtables.MappingFieldGroupsColumnName, &names); err != nil {
		dbErr := database.Error(database.LIST, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to list field group names: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return names, nil
}

// GetFieldGroupsByNames loads every visible row for the given names and groups them as map[name] -> map[scope] -> row.
func GetFieldGroupsByNames(ctx context.Context, tx orm.DB, names []string) (FieldGroupsByNameAndScope, *api.Error) {
	out := make(FieldGroupsByNameAndScope)
	if len(names) == 0 {
		return out, nil
	}

	var rows []dbmodel.MappingFieldGroups

	res := tx.Model(&dbmodel.MappingFieldGroups{})

	res = res.Where(fmt.Sprintf("%s IN (?)", dbtables.MappingFieldGroupsColumnName), names)

	res = filterFieldGroupsByTenant(res, ctx)

	if err := res.Find(&rows); err != nil {
		dbErr := database.Error(database.LIST, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to list field groups: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		if out[row.Name] == nil {
			out[row.Name] = make(map[api.MappingScope]dbmodel.MappingFieldGroups)
		}

		out[row.Name][row.ScopeType] = row
	}

	return out, nil
}

func SaveFieldGroup(_ context.Context, tx orm.DB, row *dbmodel.MappingFieldGroups) *api.Error {
	if row.Fields == nil {
		row.Fields = map[string]api.FieldGroupFieldsWithScopeValue{}
	}
	if row.ItemSelectors == nil {
		row.ItemSelectors = map[string]api.ItemSelectorsWithScopeValue{}
	}

	query := dbmodel.MappingFieldGroups{
		ID:        row.ID,
		Name:      row.Name,
		ScopeType: row.ScopeType,
		TenantId:  row.TenantId,
	}

	rowsAffected, err := tx.Model(&query).Where(query).Save(row)
	if err != nil {
		dbErr := database.Error(database.UPDATE, database.MappingFieldGroups).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to update field group: %s: %v", dbErr.Message, err)
		return dbErr
	}

	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("field group %s/%s is not found within scope %s/%d", row.Name, row.ID, row.ScopeType, row.TenantId), nil)
	}

	return nil
}

// DeleteFieldGroupByNameAndScope deletes at most one row for name and scope.
// Tenant scope adds TenantId so GORM includes it in WHERE (non-zero); system rows match on name + scope only
// (schema enforces tenant_id = 0 for system).
func DeleteFieldGroupByNameAndScope(ctx context.Context, tx orm.DB, name string, scope api.MappingScope) *api.Error {
	query := dbmodel.MappingFieldGroups{
		Name:      name,
		ScopeType: scope,
	}
	if scope == api.MAPPINGSCOPE_TENANT {
		query.TenantId = auth.TenantIdOf(ctx)
	}

	rowsAffected, err := tx.Model(&dbmodel.MappingFieldGroups{}).Where(query).Delete(&dbmodel.MappingFieldGroups{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingFieldGroups).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete field group: %s: %v", dbErr.Message, err)
		return dbErr
	}

	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("field group %s (scope %s) is not found", name, scope), nil)
	}
	return nil
}

// CheckIfFormatterBeingUsed reports whether any field group visible to the request tenant references
// the given formatter name in its fields map (jsonb values contain "formatter": "<name>").
func CheckIfFormatterBeingUsed(ctx context.Context, formatterName string) (bool, *api.Error) {
	tenantID := auth.TenantIdOf(ctx)
	sql := fmt.Sprintf(`
			SELECT COUNT(*) FROM %s mfg
			WHERE %s
			AND EXISTS (
			  SELECT 1 FROM jsonb_each(mfg.fields) AS e(k, v)
			  WHERE (v->>'formatter') = ?
			)`,
		dbtables.MappingFieldGroupsTable,
		filterFieldGroupsByTenantQuery(),
	)

	var count int64

	if _, err := runaidb.DB().WithContext(ctx).Raw(sql,
		string(api.MAPPINGSCOPE_SYSTEM),
		string(api.MAPPINGSCOPE_TENANT),
		tenantID,
		formatterName,
	).Scan(&count); err != nil {
		dbErr := database.Error(database.COUNT, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to count field groups: %s: %v", dbErr.Message, err)
		return false, dbErr
	}

	return count > 0, nil
}
