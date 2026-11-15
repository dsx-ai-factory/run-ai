/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// mappingGvkIdExpr names a mapping row the way its dependency rows identify it: group/kind/version.
var mappingGvkIdExpr = fmt.Sprintf("%s || '/' || %s || '/' || %s",
	dbtables.MappingSelectorsColumnGvkGroup, dbtables.MappingSelectorsColumnGvkKind, dbtables.MappingSelectorsColumnGvkVersion)

// SystemComponent is a stored system-scoped component, identified the way its dependency rows identify it.
type SystemComponent struct {
	ComponentType dbmodel.ComponentType
	ComponentId   string
	Description   *string
}

// Dependency returns the component as its dependency rows reference it.
func (c SystemComponent) Dependency() dbmodel.EntitiesDependency {
	return dbmodel.EntitiesDependency{
		ComponentType: c.ComponentType,
		ComponentId:   c.ComponentId,
	}
}

// ListSystemComponentsByNames returns the system-scoped components of componentType among names, keyed by name.
// Formatters and field groups are named by their name, mappings by their group/kind/version GVK. Names without a
// system component are absent. Only the identifying columns are read.
func ListSystemComponentsByNames(_ context.Context, db orm.DB, componentType dbmodel.ComponentType, names []string,
) (map[string]SystemComponent, *api.Error) {
	out := make(map[string]SystemComponent, len(names))
	if len(names) == 0 {
		return out, nil
	}

	var table, nameExpr, scopeColumn string
	var entity database.DbEntity
	switch componentType {
	case dbmodel.ComponentTypeFormatter:
		table, nameExpr, scopeColumn = dbtables.MappingFormattersTable, dbtables.MappingFormattersColumnName, dbtables.MappingFormattersColumnScopeType
		entity = database.MappingFormatters
	case dbmodel.ComponentTypeFieldGroup:
		table, nameExpr, scopeColumn = dbtables.MappingFieldGroupsTable, dbtables.MappingFieldGroupsColumnName, dbtables.MappingFieldGroupsColumnScopeType
		entity = database.MappingFieldGroups
	case dbmodel.ComponentTypeMapping:
		table, nameExpr, scopeColumn = dbtables.MappingSelectorsTable, mappingGvkIdExpr, dbtables.MappingSelectorsColumnScopeType
		entity = database.MappingSelectors
	default:
		return nil, api.NewValidationError(fmt.Sprintf("unknown component type %q", componentType))
	}

	var rows []struct {
		Name        string
		Id          uuid.UUID
		Description *string
	}

	_, err := db.Table(table).
		Select(fmt.Sprintf("%s AS name, id, description", nameExpr)).
		Where(fmt.Sprintf("%s = ?", scopeColumn), api.MAPPINGSCOPE_SYSTEM).
		Where(fmt.Sprintf("%s IN ?", nameExpr), names).
		Scan(&rows)
	if err != nil {
		dbErr := database.Error(database.LIST, entity).WithDetails(err)
		logging.Errorf("failed to list system %s components: %s: %v", componentType, dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		componentId := row.Id.String()
		if componentType == dbmodel.ComponentTypeMapping {
			componentId = row.Name
		}
		out[row.Name] = SystemComponent{ComponentType: componentType, ComponentId: componentId, Description: row.Description}
	}

	return out, nil
}

// GetSystemFormattersByNames returns the system-scoped formatter rows among names, keyed by name.
func GetSystemFormattersByNames(_ context.Context, db orm.DB, names []string) (map[string]dbmodel.MappingFormatters, *api.Error) {
	out := make(map[string]dbmodel.MappingFormatters, len(names))
	if len(names) == 0 {
		return out, nil
	}

	var rows []dbmodel.MappingFormatters

	err := db.Model(&dbmodel.MappingFormatters{}).
		Where(fmt.Sprintf("%s = ?", dbtables.MappingFormattersColumnScopeType), api.MAPPINGSCOPE_SYSTEM).
		Where(fmt.Sprintf("%s IN ?", dbtables.MappingFormattersColumnName), names).
		Find(&rows)
	if err != nil {
		dbErr := database.Error(database.LIST, database.MappingFormatters).WithDetails(err)
		logging.Errorf("failed to list system formatters: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		out[row.Name] = row
	}

	return out, nil
}

// GetSystemFieldGroupsByNames returns the system-scoped field group rows among names, keyed by name.
func GetSystemFieldGroupsByNames(_ context.Context, db orm.DB, names []string) (map[string]dbmodel.MappingFieldGroups, *api.Error) {
	out := make(map[string]dbmodel.MappingFieldGroups, len(names))
	if len(names) == 0 {
		return out, nil
	}

	var rows []dbmodel.MappingFieldGroups

	err := db.Model(&dbmodel.MappingFieldGroups{}).
		Where(fmt.Sprintf("%s = ?", dbtables.MappingFieldGroupsColumnScopeType), api.MAPPINGSCOPE_SYSTEM).
		Where(fmt.Sprintf("%s IN ?", dbtables.MappingFieldGroupsColumnName), names).
		Find(&rows)
	if err != nil {
		dbErr := database.Error(database.LIST, database.MappingFieldGroups).WithDetails(err)
		logging.Errorf("failed to list system field groups: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		out[row.Name] = row
	}

	return out, nil
}

// GetSystemMappingsByGvkIds returns the system-scoped mapping rows among gvkIds (group/kind/version), keyed by GVK id.
func GetSystemMappingsByGvkIds(_ context.Context, db orm.DB, gvkIds []string) (map[string]dbmodel.MappingSelectors, *api.Error) {
	out := make(map[string]dbmodel.MappingSelectors, len(gvkIds))
	if len(gvkIds) == 0 {
		return out, nil
	}

	var rows []dbmodel.MappingSelectors

	err := db.Model(&dbmodel.MappingSelectors{}).
		Where(fmt.Sprintf("%s = ?", dbtables.MappingSelectorsColumnScopeType), api.MAPPINGSCOPE_SYSTEM).
		Where(fmt.Sprintf("%s IN ?", mappingGvkIdExpr), gvkIds).
		Find(&rows)
	if err != nil {
		dbErr := database.Error(database.LIST, database.MappingSelectors).WithDetails(err)
		logging.Errorf("failed to list system mappings: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	for _, row := range rows {
		out[string(makeGvkKey(row.GvkGroup, row.GvkKind, row.GvkVersion))] = row
	}

	return out, nil
}
