/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"maps"
	"reflect"
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

const (
	DeletionErrorLogPrefix = "error on tenant deletion."
	ServiceName            = "org-unit-service"
)

func logError(ctx context.Context, tenantId string, assetName string, err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s tenantId: <%s>, service: <%s>, asset: <%s>, error: %s", DeletionErrorLogPrefix, tenantId, ServiceName, assetName, err)
}

func GetAllTenantProjects(ctx context.Context, tenantId string) ([]*models.Project, error) {
	var projects []*models.Project
	projectsList, err := GetProjects(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, "", nil, true, models.SlimVerbosity)
	if err != nil {
		return nil, err
	}
	for _, project := range *projectsList {
		projects = append(projects, &project)
	}
	return projects, nil
}

func GetAllTenantDepartments(ctx context.Context, tenantId string) ([]*models.Department, error) {
	var departments []*models.Department
	departmentsList, err := GetDepartments(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, nil, true, models.SlimVerbosity)
	if err != nil {
		return nil, err
	}
	for _, department := range *departmentsList {
		departments = append(departments, &department)
	}

	return departments, nil
}

func GetAllTenantNodeTypes(ctx context.Context, tenantId string) ([]*models.NodeType, error) {
	var nodeTypes []*models.NodeType
	nodeTypesList := &[]models.NodeType{}
	db := runaidb.DB().WithContext(ctx).Table(models.NodeTypeTable).Unscoped()
	if err := db.Where(&models.NodeType{TenantId: tenantId}).Find(nodeTypesList); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", err)
		return nil, err
	}
	for _, nodeType := range *nodeTypesList {
		nodeTypes = append(nodeTypes, &nodeType)
	}

	return nodeTypes, nil
}

func getProjectsClusterIds(projects []*models.Project) []string {
	clusterIds := map[string]bool{}
	for _, project := range projects {
		clusterIds[project.ClusterId.String()] = true
	}
	return slices.Collect(maps.Keys(clusterIds))
}

func getDepartmentsClusterIds(departments []*models.Department) []string {
	clusterIds := map[string]bool{}
	for _, department := range departments {
		clusterIds[department.ClusterId.String()] = true
	}
	return slices.Collect(maps.Keys(clusterIds))
}

func invalidateCaches(ctx context.Context, tenantId string, clusterIds []string) {
	for _, clusterId := range clusterIds {
		InvalidateAllCache(ctx, tenantId, clusterId)
	}
}

func HardDeleteTenantNodeTypes(ctx context.Context, tenantId string) error {
	nodeTypes, err := GetAllTenantNodeTypes(ctx, tenantId)
	if err != nil {
		logError(ctx, tenantId, "NodeType", err)
		return err
	}

	if len(nodeTypes) == 0 {
		return nil
	}
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	var nodeTypesIds []int64
	for _, nodeType := range nodeTypes {
		nodeTypesIds = append(nodeTypesIds, nodeType.ID)
	}

	err = hardDeleteTenantNodeTypeWorkload[models.OrgUnitTrainNodeType](ctx, tx, models.OrgUnitNodeTypeTrainTable, tenantId, nodeTypesIds)
	if err != nil {
		return err
	}
	err = hardDeleteTenantNodeTypeWorkload[models.OrgUnitWorkspaceNodeType](ctx, tx, models.OrgUnitNodeTypeWorkspaceTable, tenantId, nodeTypesIds)
	if err != nil {
		return err
	}
	if _, err := tx.Table(models.NodeTypeTable).Where(nodeTypesIds).Unscoped().Delete(&[]models.NodeType{}); err != nil {
		logError(ctx, tenantId, "NodeType", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		logError(ctx, tenantId, "NodeType", err)
		return err
	}
	committed = true
	return nil
}

func hardDeleteTenantNodeTypeWorkload[T any](ctx context.Context, tx orm.DB, tableName string, tenantId string, nodeTypesIds []int64) error {
	if len(nodeTypesIds) == 0 {
		return nil
	}
	modelType := reflect.TypeFor[T]()
	assetName := modelType.String()
	if _, err := tx.Table(tableName).Where("node_type_id IN ?", nodeTypesIds).Unscoped().Delete(&[]T{}); err != nil {
		logError(ctx, tenantId, assetName, err)
		return err
	}
	return nil
}

func HardDeleteTenantProjects(ctx context.Context, tenantId string, projects []*models.Project) error {
	if len(projects) == 0 {
		return nil
	}

	clusterIds := getProjectsClusterIds(projects)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	var resourcesIds []uuid.UUID
	var orgUnitsIds []uuid.UUID
	var projectsIds []int64
	for _, project := range projects {
		projectsIds = append(projectsIds, project.ID)
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID)
		for _, projResources := range project.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, projResources.ID)
		}
	}

	if len(projectsIds) > 0 {
		if _, err := tx.Table(models.ProjectTable).Where(projectsIds).Unscoped().Delete(&[]models.Project{}); err != nil {
			logError(ctx, tenantId, "Projects", err)
			return err
		}
	}

	if len(resourcesIds) > 0 {
		if _, err := tx.Table(models.ResourcesTable).Where(resourcesIds).Unscoped().Delete(&[]models.Resources{}); err != nil {
			logError(ctx, tenantId, "Resources", err)
			return err
		}
	}

	for _, orgUnitsId := range orgUnitsIds {
		if _, err := tx.Table(models.OrgUnitDataTable).Where("id = ?", orgUnitsId).Unscoped().Delete(&[]models.OrgUnitData{}); err != nil {
			logError(ctx, tenantId, "OrgUnitData", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logError(ctx, tenantId, "Projects", err)
		return err
	}
	committed = true
	invalidateCaches(ctx, tenantId, clusterIds)

	return nil
}

func HardDeleteTenantDepartments(ctx context.Context, tenantId string, departments []*models.Department) error {
	if len(departments) == 0 {
		return nil
	}
	clusterIds := getDepartmentsClusterIds(departments)
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	var resourcesIds []uuid.UUID
	var orgUnitsIds []uuid.UUID
	var departmentsIds []int64
	for _, department := range departments {
		departmentsIds = append(departmentsIds, department.ID)
		orgUnitsIds = append(orgUnitsIds, department.OrgUnitData.ID)
		for _, deptResources := range department.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, deptResources.ID)
		}
	}
	if _, err := tx.Table(models.ResourcesTable).Where(resourcesIds).Unscoped().Delete(&models.Resources{}); err != nil {
		logError(ctx, tenantId, "Resources", err)
		return err
	}

	if _, err := tx.Where(departmentsIds).Unscoped().Delete(&models.Department{}); err != nil {
		logError(ctx, tenantId, "Departments", err)
		return err
	}

	if _, err := tx.Where(orgUnitsIds).Unscoped().Delete(&models.OrgUnitData{}); err != nil {
		logError(ctx, tenantId, "OrgUnitData", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logError(ctx, tenantId, "Departments", err)
		return err
	}
	committed = true
	invalidateCaches(ctx, tenantId, clusterIds)

	return nil
}

func HardDeleteOvertimeData(ctx context.Context, tenantId string, projects []*models.Project, departments []*models.Department) error {
	var orgUnitsIds []uuid.UUID
	for _, department := range departments {
		orgUnitsIds = append(orgUnitsIds, department.OrgUnitData.ID)
	}
	for _, project := range projects {
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID)
	}
	if _, err := runaidb.DB().WithContext(ctx).Where("org_unit_data_id in ?", orgUnitsIds).Unscoped().Delete(&models.OvertimeData{}); err != nil {
		logError(ctx, tenantId, "OvertimeData", err)
		return err
	}
	return nil
}
