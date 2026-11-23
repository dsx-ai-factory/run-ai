/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func GetDepartmentNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	permittedScopes *openapi.PermittedScopes,
	search string,
) (*[]models.DepartmentNodePoolResources, error) {
	var permittedScopesQuery *string = nil
	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			permittedScopesQuery = ptr.To(getPermittedScopesForDepartmentResources(*permittedScopes))
		}
	}

	nodePoolResources := &[]models.DepartmentNodePoolResources{}
	query, err := getNodePoolResources(ctx, tenantId, filterBy, sortBy, sortOrder, offset, limit, permittedScopesQuery, search, models.DepartmentNodePoolResourcesView)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department node pool resources, err: %v", err)
		return nil, err
	}

	if err := query.Find(&nodePoolResources); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department node pool resources: %v", err)
		return nil, err
	}
	return nodePoolResources, nil
}

func GetProjectNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	permittedScopes *openapi.PermittedScopes,
	search string,
) (*[]models.ProjectNodePoolResources, error) {
	var permittedScopesQuery *string = nil
	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			permittedScopesQuery = ptr.To(getPermittedScopesForProjectResources(*permittedScopes))
		}
	}

	nodePoolResources := &[]models.ProjectNodePoolResources{}
	query, err := getNodePoolResources(ctx, tenantId, filterBy, sortBy, sortOrder, offset, limit, permittedScopesQuery, search, models.ProjectNodePoolResourcesView)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project node pool resources, err: %v", err)
		return nil, err
	}

	if err := query.Find(&nodePoolResources); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project node pool resources: %v", err)
		return nil, err
	}
	return nodePoolResources, nil
}

func getNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	permittedScopesQuery *string,
	search string,
	viewName string,
) (orm.DB, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(viewName).
		Where("tenant_id = ?", tenantId)

	if permittedScopesQuery != nil {
		query = query.Where(*permittedScopesQuery)
	}

	query = applyParameterizedFilters(query, filterBy)
	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	if search != "" {
		query = query.Where("org_unit_name like ?", "%"+search+"%")
	}
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	return query, nil
}

func CountProjectNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	search string,
	permittedScopes *openapi.PermittedScopes,
) (*int64, error) {
	var permittedScopesQuery *string = nil
	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			permittedScopesQuery = ptr.To(getPermittedScopesForProjectResources(*permittedScopes))
		}
	}
	return countNodePoolResources(ctx, tenantId, filterBy, search, permittedScopesQuery, models.ProjectNodePoolResourcesView)
}

func CountDepartmentNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	search string,
	permittedScopes *openapi.PermittedScopes,
) (*int64, error) {
	var permittedScopesQuery *string = nil
	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			permittedScopesQuery = ptr.To(getPermittedScopesForDepartmentResources(*permittedScopes))
		}
	}
	return countNodePoolResources(ctx, tenantId, filterBy, search, permittedScopesQuery, models.DepartmentNodePoolResourcesView)
}

func countNodePoolResources(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	search string,
	permittedScopesQuery *string,
	viewName string,
) (*int64, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(viewName).
		Where("tenant_id = ?", tenantId)

	if permittedScopesQuery != nil {
		query = query.Where(*permittedScopesQuery)
	}

	query = applyParameterizedFilters(query, filterBy)

	if search != "" {
		query = query.Where("org_unit_name like ?", "%"+search+"%")
	}

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get node pool resources count from view %s: %s", viewName, err)
		return nil, err
	}
	return &count, nil
}

func SumProjectNodePoolResourcesGpuWeights(ctx context.Context, tenantId string, nodePoolId string,
	filterBy []queryapi.Filter, nodePoolAdjustableWeights bool,
	permittedScopes *openapi.PermittedScopes,
) (*float64, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.ProjectNodePoolResourcesView).
		Where("tenant_id = ?", tenantId).
		Where("node_pool_id = ?", nodePoolId)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjectResources(*permittedScopes))
		}
	}

	query = applyParameterizedFilters(query, filterBy)

	if nodePoolAdjustableWeights {
		query = query.Select("COALESCE(SUM(gpu_over_quota_weight), 0)")
	} else {
		// if AdjustableWeights is off, we sum the gpu_deserved instead as weights_sum
		query = query.Select("COALESCE(SUM(CASE WHEN gpu_deserved >= 0 THEN gpu_deserved ELSE 0 END), 0)")
	}

	var sum float64
	if _, err := query.Scan(&sum); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project node pool <%s> resources gpu weights sum: %v", nodePoolId, err)
		return nil, err
	}
	return &sum, nil
}

type DepartmentNodePoolResourcesSums struct {
	DeservedSum float64
	WeightsSum  float64
}

func SumDepartmentNodePoolResources(ctx context.Context, tenantId string, nodePoolId string,
	filterBy []queryapi.Filter, nodePoolAdjustableWeights bool,
	permittedScopes *openapi.PermittedScopes,
) (*DepartmentNodePoolResourcesSums, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.DepartmentNodePoolResourcesView).
		Where("tenant_id = ?", tenantId).
		Where("node_pool_id = ?", nodePoolId)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartmentResources(*permittedScopes))
		}
	}

	query = applyParameterizedFilters(query, filterBy)

	if nodePoolAdjustableWeights {
		query = query.Select(`
			CASE 
				WHEN COUNT(CASE WHEN gpu_deserved IS NULL OR gpu_deserved < 0 THEN 1 END) > 0 
				THEN -1 
				ELSE COALESCE(SUM(gpu_deserved), 0) 
			END AS deserved_sum,
			COALESCE(SUM(gpu_over_quota_weight), 0) AS weights_sum
		`)
	} else {
		// if AdjustableWeights is off, we sum the gpu_deserved instead as weights_sum
		// but deserved_sum returns -1 if any value is unlimited
		query = query.Select(`
			CASE 
				WHEN COUNT(CASE WHEN gpu_deserved IS NULL OR gpu_deserved < 0 THEN 1 END) > 0 
				THEN -1 
				ELSE COALESCE(SUM(gpu_deserved), 0) 
			END AS deserved_sum,
			COALESCE(SUM(CASE WHEN gpu_deserved >= 0 THEN gpu_deserved ELSE 0 END), 0) AS weights_sum
		`)
	}

	var sums DepartmentNodePoolResourcesSums
	if _, err := query.Scan(&sums); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department node pool <%s> resources sums: %v", nodePoolId, err)
		return nil, err
	}

	return &sums, nil
}

func getPermittedScopesForProjectResources(permittedScopes openapi.PermittedScopes) string {
	return GetPermittedScopesForProjectsWithColumn(permittedScopes, "org_unit_id")
}

func getPermittedScopesForDepartmentResources(permittedScopes openapi.PermittedScopes) string {
	return getPermittedScopesForDepartmentsWithColumn(permittedScopes, "org_unit_id")
}

func GetDepartmentNodePoolResourcesPerDepartmentNodePool(
	ctx context.Context, tenantId string, departmentId, nodePoolId string, permittedScopes *openapi.PermittedScopes,
) (*models.DepartmentNodePoolResources, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.DepartmentNodePoolResourcesView).
		Where("tenant_id = ?", tenantId).
		Where("org_unit_id = ?", departmentId).
		Where("node_pool_id = ?", nodePoolId)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartmentResources(*permittedScopes))
		}
	}

	nodePoolResources := &models.DepartmentNodePoolResources{}
	if err := query.First(nodePoolResources); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department node pool resources for department %s and node pool %s: %v",
			departmentId, nodePoolId, err)
		return nil, err
	}
	return nodePoolResources, nil
}

func GetProjectNodePoolResourcesPerDepartmentNodePool(
	ctx context.Context, tenantId string, departmentId, nodePoolId string,
) (*models.ProjectNodePoolResources, error) {
	nodePoolResources := &models.ProjectNodePoolResources{}

	err := runaidb.DB().WithContext(ctx).
		Table(models.ProjectNodePoolResourcesView).
		Where("tenant_id = ?", tenantId).
		Where("org_unit_id = ?", departmentId).
		Where("node_pool_id = ?", nodePoolId).
		First(nodePoolResources)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project node pool resources for project node %s and node pool %s: %v",
			departmentId, nodePoolId, err)
		return nil, err
	}
	return nodePoolResources, nil
}

func PatchDepartmentNodePoolResources(ctx context.Context,
	tenantId, clusterId, departmentId, nodePoolId, resourcesId string,
	subjectId *string,
	updates map[string]interface{},
) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// Update the concrete resources row for the specific node pool.
	if _, err := tx.Table(models.Resources{}.TableName()).Where("id = ?", resourcesId).Where("node_pool_id = ?", nodePoolId).Updates(updates); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to patch node pool resources: %v", err)
		return err
	}

	// Update the owning department's audit fields
	departmentUpdates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if subjectId != nil {
		departmentUpdates["updated_by"] = *subjectId
	}

	if _, err := tx.Table(models.DepartmentTable).Where("tenant_id = ?", tenantId).Where("cluster_id = ?", clusterId).Where("id = ?", departmentId).Updates(departmentUpdates); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update department metadata while patching node pool resources: %v", err)
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit PatchDepartmentNodePoolResources transaction: %v", commitErr)
		return commitErr
	}
	committed = true

	InvalidateAllCache(ctx, tenantId, clusterId)

	return nil
}
