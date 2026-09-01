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

// the user is allowed to see only system scope formatter or tenant scoped formatter from his own tenant
func filterFormattersByTenant(q orm.DB, ctx context.Context) orm.DB {
	tenantID := auth.TenantIdOf(ctx)

	query := fmt.Sprintf("(%s = ?) OR (%s = ? AND %s = ?)",
		dbtables.MappingFormattersColumnScopeType,
		dbtables.MappingFormattersColumnScopeType,
		dbtables.MappingFormattersColumnTenantId)

	return q.Where(query, string(api.MAPPINGSCOPE_SYSTEM), string(api.MAPPINGSCOPE_TENANT), tenantID)
}

func CreateFormatter(ctx context.Context, row *dbmodel.MappingFormatters) *api.Error {
	if row.Fields == nil {
		row.Fields = map[string]api.FormatterField{}
	}

	if _, err := runaidb.DB().WithContext(ctx).Table(dbtables.MappingFormattersTable).Create(row); err != nil {
		dbErr := database.Error(database.CREATE, database.MappingFormatters).WithDetails(err)
		logging.Errorf("failed to create formatter: %s: %v", dbErr.Message, err)
		return dbErr
	}

	return nil
}

func GetFormatterByName(ctx context.Context, db orm.DB, name string) (*dbmodel.MappingFormatters, *api.Error) {
	query := dbmodel.MappingFormatters{
		Name: name,
	}

	var row dbmodel.MappingFormatters

	res := db.WithContext(ctx).Model(&query).Where(query)

	res = filterFormattersByTenant(res, ctx)

	if err := res.First(&row); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("formatter named %s not found", name), nil)
		}
		dbErr := database.Error(database.GET, database.MappingFormatters).WithDetails(err)
		logging.Errorf("failed to get formatter: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return &row, nil
}

// GetFormattersByNames returns visible formatters by a list of names
func GetFormattersByNames(ctx context.Context, names []string) (map[string]dbmodel.MappingFormatters, *api.Error) {
	out := make(map[string]dbmodel.MappingFormatters)
	if len(names) == 0 {
		return out, nil
	}

	var rows []dbmodel.MappingFormatters

	res := runaidb.DB().WithContext(ctx).Model(&dbmodel.MappingFormatters{}).
		Where(fmt.Sprintf("%s IN (?)", dbtables.MappingFormattersColumnName), names)

	res = filterFormattersByTenant(res, ctx)

	if err := res.Find(&rows); err != nil {
		dbErr := database.Error(database.LIST, database.MappingFormatters).WithDetails(err)
		logging.Errorf("failed to list formatters: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		out[row.Name] = row
	}

	return out, nil
}

func ListFormatters(ctx context.Context, offset, limit int) ([]dbmodel.MappingFormatters, *api.Error) {
	var rows []dbmodel.MappingFormatters

	res := runaidb.DB().WithContext(ctx).Model(&dbmodel.MappingFormatters{})

	res = filterFormattersByTenant(res, ctx)

	if err := res.Order(dbtables.MappingFormattersColumnName).
		Offset(offset).
		Limit(limit).
		Find(&rows); err != nil {
		dbErr := database.Error(database.LIST, database.MappingFormatters).WithDetails(err)
		logging.Errorf("failed to list formatters: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return rows, nil
}

func SaveFormatter(ctx context.Context, db orm.DB, row *dbmodel.MappingFormatters) *api.Error {
	if row.Fields == nil {
		row.Fields = map[string]api.FormatterField{}
	}

	// we do not allow changing the name, scope or tenant ID
	query := dbmodel.MappingFormatters{
		ID:        row.ID,
		Name:      row.Name,
		ScopeType: row.ScopeType,
		TenantId:  row.TenantId,
	}

	rowsAffected, err := db.WithContext(ctx).Model(&query).Where(query).Save(row)
	if err != nil {
		dbErr := database.Error(database.UPDATE, database.MappingFormatters).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to update formatter: %s: %v", dbErr.Message, err)
		return dbErr
	}
	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("formatter %s/%s is not found within scope %s/%d", row.Name, row.ID, row.ScopeType, row.TenantId), nil)
	}

	return nil
}

func DeleteFormatterByName(ctx context.Context, db orm.DB, name string) *api.Error {
	query := dbmodel.MappingFormatters{
		Name: name,
	}

	res := db.WithContext(ctx).Model(&query).Where(query)

	res = filterFormattersByTenant(res, ctx)

	rowsAffected, err := res.Delete(&dbmodel.MappingFormatters{})
	if err != nil {
		dbErr := database.Error(database.DELETE, database.MappingFormatters).WithDetailsFromDbResult(rowsAffected, err)
		logging.Errorf("failed to delete formatter: %s: %v", dbErr.Message, err)
		return dbErr
	}

	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound, fmt.Sprintf("formatter %s is not found", name), nil)
	}

	return nil
}
