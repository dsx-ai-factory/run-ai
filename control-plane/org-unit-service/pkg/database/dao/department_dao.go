/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const (
	defaultDepartmentName         = "default"
	DepartmentsCacheHashField     = "departments"
	DepartmentsV1CacheHashField   = "departmentsV1"
	DepartmentsSyncCacheHashField = "departmentsSync"
)

func CreateDepartment(ctx context.Context, department *models.Department) (*models.Department, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Omit(orgUnitNodeTypeOmitFields()...).Create(department); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed create department: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit department creation: %v", err)
		return nil, err
	}
	committed = true
	InvalidateAllCache(ctx, department.TenantId, department.ClusterId.String())

	return department, nil
}

func CreateDepartments(ctx context.Context, departments []*models.Department) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Omit(orgUnitNodeTypeOmitFields()...).
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(departments, 20); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed create departments: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit departments creation: %v", err)
		return err
	}
	committed = true
	return nil
}

func HasUnlimitedGpuDepartment(ctx context.Context, clusterId string, tenantId string) (bool, error) {
	count := int64(0)

	query := runaidb.DB().WithContext(ctx).Table(models.DepartmentTable).
		Joins("LEFT JOIN resources ON department.org_unit_data_id = resources.org_unit_data_id").
		Where("resources.gpu_deserved = -1").
		Where("department.deleted_at is null").
		Where("resources.deleted_at is null").
		Where("department.cluster_id = ?", clusterId).
		Where("department.tenant_id = ?", tenantId)
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to check unlimited department: %v", err)
		return false, err
	}

	return count > 0, nil
}

func UpdateDepartment(ctx context.Context, department *models.Department) (*models.Department, error) {
	return UpdateDepartmentWithChildren(ctx, department, nil)
}

func UpdateDepartmentWithChildren(ctx context.Context, department *models.Department, projects *[]models.Project) (*models.Department, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	err := tx.Actual().Model(&models.OrgUnitData{BaseModel: dbmodels.BaseModel{ID: department.OrgUnitDataID}}).Association(TrainNodeType).Clear()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update department: %v", err)
		return nil, err
	}
	err = tx.Actual().Model(&models.OrgUnitData{BaseModel: dbmodels.BaseModel{ID: department.OrgUnitDataID}}).Association(WorkspaceNodeType).Clear()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update department: %v", err)
		return nil, err
	}

	departmentWhere := models.Department{
		IntBaseModel: models.IntBaseModel{ID: department.ID},
		TenantId:     department.TenantId,
		Name:         department.Name,
	}
	if _, err := tx.Table(models.DepartmentTable).Where(&departmentWhere).
		Omit(orgUnitNodeTypeOmitFields()...).
		Updates(department); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed department update: %v", err)
		return nil, err
	}

	if projects != nil && len(*projects) > 0 {
		now := time.Now()
		for _, project := range *projects {
			orgUnitData := project.OrgUnitData.EffectiveData
			updates := map[string]interface{}{
				"interactive_job_time_limit_effective":                   orgUnitData.InteractiveJobTimeLimitEffective,
				"interactive_job_time_limit_effective_origin":            orgUnitData.InteractiveJobTimeLimitEffectiveOrigin,
				"interactive_job_preempt_idle_duration_effective":        orgUnitData.InteractiveJobPreemptIdleDurationEffective,
				"interactive_job_preempt_idle_duration_effective_origin": orgUnitData.InteractiveJobPreemptIdleDurationEffectiveOrigin,
				"interactive_job_max_idle_duration_effective":            orgUnitData.InteractiveJobMaxIdleDurationEffective,
				"interactive_job_max_idle_duration_effective_origin":     orgUnitData.InteractiveJobMaxIdleDurationEffectiveOrigin,
				"training_job_time_limit_effective":                      orgUnitData.TrainingJobTimeLimitEffective,
				"training_job_time_limit_effective_origin":               orgUnitData.TrainingJobTimeLimitEffectiveOrigin,
				"training_job_max_idle_duration_effective":               orgUnitData.TrainingJobMaxIdleDurationEffective,
				"training_job_max_idle_duration_effective_origin":        orgUnitData.TrainingJobMaxIdleDurationEffectiveOrigin,
				"default_node_pool_list_effective":                       orgUnitData.DefaultNodePoolListEffective,
				"default_node_pool_list_effective_origin":                orgUnitData.DefaultNodePoolListEffectiveOrigin,
				"training_node_types_effective_origin":                   orgUnitData.TrainingNodeTypesEffectiveOrigin,
				"workspace_node_types_effective_origin":                  orgUnitData.WorkspaceNodeTypesEffectiveOrigin,
				"updated_at":                                             now,
			}
			if _, err := tx.Table(models.OrgUnitDataTable).Where("id = ?", project.OrgUnitDataID).Updates(updates); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed department children update: %v", err)
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit department update: %v", err)
		return nil, err
	}
	committed = true
	InvalidateAllCache(ctx, department.TenantId, department.ClusterId.String())

	return department, nil
}

func GetClusterIdByDepartment(ctx context.Context, departmentId int64, tenantId string) (string, error) {
	var clusterId string
	if err := runaidb.DB().WithContext(ctx).Model(&models.Department{}).Select("cluster_id").Where("id = ?", departmentId).Where("tenant_id = ?", tenantId).First(&clusterId); err != nil {
		return "", err
	}
	return clusterId, nil
}

func GetDepartmentById(ctx context.Context, departmentId int64, tenantId string, verbosity models.Verbosity, permittedScopes *openapi.PermittedScopes) (*models.Department, error) {
	department := &models.Department{
		IntBaseModel: models.IntBaseModel{ID: departmentId},
		TenantId:     tenantId,
	}
	return getDepartment(ctx, department, verbosity, tenantId, permittedScopes)
}

func GetDepartmentByIdAndCluster(ctx context.Context, departmentId int64, tenantId string, clusterId uuid.UUID, verbosity models.Verbosity) (*models.Department, error) {
	department := &models.Department{
		IntBaseModel: models.IntBaseModel{ID: departmentId},
		ClusterId:    clusterId,
		TenantId:     tenantId,
	}
	return getDepartment(ctx, department, verbosity, tenantId, nil)
}

func GetDepartmentsByCluster(ctx context.Context, tenantId string, clusterId uuid.UUID, permittedScopes *openapi.PermittedScopes, verbosity models.Verbosity) (*[]models.Department, error) {
	cacheHashField := DepartmentsCacheHashField
	if verbosity == models.VerboseVerbosity {
		cacheHashField = DepartmentsV1CacheHashField
	}
	GetDepartmentsByClusterInner := func() (*[]models.Department, error) {
		departments := &[]models.Department{}

		department := &models.Department{
			ClusterId: clusterId,
			TenantId:  tenantId,
		}
		db := runaidb.DB().WithContext(ctx)

		if permittedScopes != nil {
			tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
			if !tenantPermitted {
				db = db.Where(getPermittedScopesForDepartments(*permittedScopes))
			}
		}

		var err error
		if verbosity == models.MetaDataVerbosity {
			err = db.Table(models.DepartmentTable).Where(&department).Find(&departments)
		} else {
			departmentsViews := []models.DepartmentView{}
			err = db.Table(models.DepartmentViewName).Where(&department).Find(&departmentsViews)
			departments = ptr.To(viewResultToDepartments(departmentsViews))
		}
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments in verbosity %v: %v", verbosity, err)
			return nil, err
		}

		if verbosity == models.VerboseVerbosity {
			projects, err := GetProjectsByCluster(ctx, tenantId, clusterId, nil, models.SlimVerbosity)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments: %v", err)
				return nil, err
			}

			departments = MapProjectsToDepartments(departments, projects)
		}

		return departments, nil
	}

	var departments *[]models.Department
	var err error
	if IsEligibleForCachedData(permittedScopes, tenantId, clusterId.String()) {
		departments, err := cache.Get[[]models.Department](ctx, GetCacheKey(cacheHashField, tenantId, clusterId.String()))
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get cached departments. err %+v", err)
		}
		if departments == nil {
			departments, err = GetDepartmentsByClusterInner()
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments in cluster %s. err %+v", clusterId, err)
				return nil, err
			}
			if err := cache.Set[[]models.Department](ctx, GetCacheKey(cacheHashField, tenantId, clusterId.String()), *departments); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set cached departments. err %+v", err)
			}
		}
	}

	if departments == nil {
		departments, err = GetDepartmentsByClusterInner()
	}

	return departments, err
}

func CountDepartments(ctx context.Context, tenantId string, permittedScopes *openapi.PermittedScopes, filterBy []queryapi.Filter, verbosity models.Verbosity) (int64, error) {
	query := runaidb.DB().WithContext(ctx)

	query = verbosityBeforeFilterHook(verbosity, query).Where("deleted_at IS NULL").Where("tenant_id = ?", tenantId)

	query = applyParameterizedFilters(query, filterBy)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count departments: %v", err)
		return 0, err
	}

	return count, nil
}

func GetDepartmentsRanking(ctx context.Context, tenantId string, clusterId string, permittedScopes *openapi.PermittedScopes,
	columnNames RankingColumnNames, orderBy string, orderType api.RankingOrderType, limit int32,
) ([]api.DepartmentGetRanking, error) {
	results, err := GetOrgUnitsRanking(ctx, tenantId, clusterId, permittedScopes,
		columnNames, orderBy, orderType, limit, true, "")
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to query departments ranking. clusterId: %s, orderBy: %s, orderType: %s, limit: %d, error: %v",
			clusterId, orderBy, orderType, limit, err,
		)
		return nil, err
	}

	var departmentResults []api.DepartmentGetRanking
	for _, result := range results {
		departmentResults = append(departmentResults, api.DepartmentGetRanking{
			Name:                   result.Name,
			Id:                     result.Id,
			GpuQuota:               result.GpuQuota,
			AllocationValue:        result.AllocationValue,
			UtilizationValue:       result.UtilizationValue,
			MemoryUtilizationValue: result.MemoryUtilizationValue,
		})
	}

	return departmentResults, nil
}

func MapProjectsToDepartments(departments *[]models.Department, projects []*models.Project) *[]models.Department {
	projectsMap := map[int64][]models.Project{}

	for _, project := range projects {
		if project == nil || project.ParentID == nil {
			continue
		}

		if _, exists := projectsMap[*project.ParentID]; !exists {
			projectsMap[*project.ParentID] = []models.Project{}
		}

		projectsMap[*project.ParentID] = append(projectsMap[*project.ParentID], *project)
	}

	newList := &[]models.Department{}
	for _, department := range *departments {
		if projects, exists := projectsMap[department.ID]; exists {
			department.ProjectChildren = projects
		}

		*newList = append(*newList, department)
	}

	return newList
}

func getDepartment(ctx context.Context, department *models.Department, verbosity models.Verbosity, tenantId string, permittedScopes *openapi.PermittedScopes) (*models.Department, error) {
	query := runaidb.DB().WithContext(ctx)

	query = verbosityBeforeFilterHook(verbosity, query)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	query = query.Where(department)

	return verbosityAfterFilterHookSingle(ctx, verbosity, query)
}

func mapToVerboseVerbosity(ctx context.Context, department *models.Department) (*models.Department, error) {
	projects, err := GetProjectsByParent(ctx, department.ID, department.TenantId, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get projects: %v", err)
		return nil, err
	}

	department.ProjectChildren = *projects

	return department, nil
}

func GetDepartmentsByParent(ctx context.Context, parentId int64, tenantId string) (*[]models.Department, error) {
	departments := &[]models.Department{}
	if err := runaidb.DB().WithContext(ctx).Table(models.DepartmentTable).Where(&models.Department{ParentID: &parentId, TenantId: tenantId}).Find(&departments); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department: %v", err)
		return nil, err
	}
	return departments, nil
}

func GetDepartments(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	permittedScopes *openapi.PermittedScopes,
	includeDeleted bool,
	verbosity models.Verbosity,
) (*[]models.Department, error) {
	db := runaidb.DB()
	if includeDeleted {
		db = db.Unscoped()
	}
	query := db.WithContext(ctx)
	department := models.Department{
		TenantId: tenantId,
	}

	query = verbosityBeforeFilterHook(verbosity, query)

	query = query.Where(&department)

	query = applyParameterizedFilters(query, filterBy)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	return verbosityAfterFilterHookSlice(ctx, verbosity, query)
}

func GetSyncDepartments(ctx context.Context, tenantId string, clusterId uuid.UUID, updatedAfter time.Time, offset int32, limit int32, permittedScopes *openapi.PermittedScopes) (*[]models.Department, error) {
	if updatedAfter.IsZero() {
		// in this case we need to invalidate cache since it holds data from certain timestamp
		InvalidateAllCache(ctx, tenantId, clusterId.String())
	}
	GetSyncDepartmentsInner := func() ([]*models.DepartmentSyncView, error) {
		departments := []*models.DepartmentSyncView{}
		departmentsList, err := GetSyncDepartmentsView(ctx, tenantId, clusterId, updatedAfter, 0, 0, permittedScopes)
		if err != nil {
			return nil, err
		}
		for i := range *departmentsList {
			department := (*departmentsList)[i]
			departments = append(departments, &department)
		}
		return departments, nil
	}
	var departments []*models.DepartmentSyncView
	var err error
	departmentsSyncPointer, err := cache.Get[[]*models.DepartmentSyncView](ctx, GetCacheKey(DepartmentsSyncCacheHashField, tenantId, clusterId.String()))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get cached departments sync views. err %+v", err)
	}
	if departmentsSyncPointer != nil {
		departments = *departmentsSyncPointer
	}
	if departments == nil {
		departments, err = GetSyncDepartmentsInner()
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments sync views in cluster %s. err %+v", clusterId, err)
			return nil, err
		}
		if err := cache.Set[[]*models.DepartmentSyncView](ctx, GetCacheKey(DepartmentsSyncCacheHashField, tenantId, clusterId.String()), departments); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set cached departments sync views. err %+v", err)
		}
	}

	if offset != 0 {
		departments = departments[offset:]
	}
	if limit != 0 && len(departments) > int(limit) {
		departments = departments[:limit]
	}

	return ptr.To(syncViewResultToDepartments(departments)), nil
}

func GetSyncDepartmentsView(
	ctx context.Context,
	tenantId string,
	clusterId uuid.UUID,
	updatedAfter time.Time,
	offset int32,
	limit int32,
	permittedScopes *openapi.PermittedScopes,
) (*[]models.DepartmentSyncView, error) {
	// Use optimized SQL function instead of the view — filters by tenant/cluster
	// before aggregating resources, avoiding full-table scans. See migration 02250002.
	query := runaidb.DB().WithContext(ctx).Table(
		"(SELECT * FROM get_sync_departments(?, ?)) AS sync_view", tenantId, clusterId)

	if !updatedAfter.IsZero() {
		query = query.Where("calculated_updated_at > ? OR "+
			"status_last_updated_time is null OR "+
			"calculated_updated_at > status_last_updated_time OR "+
			"(deleted_at is not null AND deleted_at > ?)", updatedAfter, updatedAfter)
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	query = query.Order("status_last_updated_time desc").Order("calculated_updated_at").Order("id")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	query = query.Unscoped()

	departmentsViews := []models.DepartmentSyncView{}
	if err := query.Find(&departmentsViews); err != nil {
		logging.Errorf("Failed to get departments sync, err: %v", err)
		return nil, err
	}
	return ptr.To(departmentsViews), nil
}

func GetDepartmentByClusterAndIds(ctx context.Context, tenantId string, clusterId uuid.UUID, ids []string, permittedScopes *openapi.PermittedScopes) (*[]models.Department, error) {
	query := runaidb.DB().WithContext(ctx).Table(models.DepartmentViewName)

	query = query.Where(&models.Department{
		TenantId:  tenantId,
		ClusterId: clusterId,
	}).Where("id in ?", ids)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	return verbosityAfterFilterHookSlice(ctx, models.SlimVerbosity, query)
}

func GetDepartmentsByIds(ctx context.Context, tenantId string, ids []int64, permittedScopes *openapi.PermittedScopes) (*[]models.Department, error) {
	departments := &[]models.Department{}
	query := runaidb.DB().WithContext(ctx).
		Table(models.DepartmentTable).
		Where("tenant_id = ?", tenantId).
		Where("id in ?", ids)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForDepartments(*permittedScopes))
		}
	}

	if err := query.Find(&departments); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department: %v", err)
		return nil, err
	}
	return departments, nil
}

func DeleteDepartment(ctx context.Context, departmentId int64, tenantId string, deletedBy *string, defaultDepartmentName string) (string, error) {
	department, err := GetDepartmentById(ctx, departmentId, tenantId, models.MetaDataVerbosity, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department: %v", err)
		return "", err
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	if department.Name == defaultDepartmentName {
		return "", errors.New("cannot delete default department")
	}
	if deletedBy != nil {
		department.DeletedBy = *deletedBy
	}
	if err := deleteDepartmentOnly(ctx, department, tx); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department: %v", err)
		return "", err
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete department: %v", err)
		return "", err
	}
	committed = true
	InvalidateAllCache(ctx, department.TenantId, department.ClusterId.String())

	return department.Name, nil
}

func DeleteDepartmentWithChildren(ctx context.Context, department *models.Department, deletedBy *string, defaultDepartmentName string) error {
	if department.Name == defaultDepartmentName {
		return errors.New("cannot delete default department")
	}
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	if deletedBy != nil {
		department.DeletedBy = *deletedBy
	}
	if err := deleteDepartmentOnly(ctx, department, tx); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department: %v", err)
		return err
	}

	projects := department.ProjectChildren
	if len(projects) > 0 {
		if err := deleteDepartmentChildren(ctx, projects, tx); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department's projects: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete department: %v", err)
		return err
	}
	committed = true

	InvalidateAllCache(ctx, department.TenantId, department.ClusterId.String())
	return nil
}

func deleteDepartmentOnly(ctx context.Context, department *models.Department, tx orm.DB) error {
	if _, err := tx.Where("org_unit_data_id = ?", department.OrgUnitDataID).Delete(&models.Resources{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department resources: %v", err)
		return err
	}
	if _, err := tx.Where("id = ?", department.OrgUnitDataID).Delete(&models.OrgUnitData{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department OrgUnitData: %v", err)
		return err
	}
	if _, err := tx.Where("org_unit_data_id = ?", department.OrgUnitDataID).Delete(&models.OvertimeData{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department OvertimeData: %v", err)
		return err
	}
	if _, err := tx.Delete(&department, department.ID); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete department: %v", err)
		return err
	}
	return nil
}

func deleteDepartmentChildren(ctx context.Context, projects []models.Project, tx orm.DB) error {
	resourcesIds := []string{}
	orgUnitsIds := []string{}
	projectsIds := []string{}
	for _, project := range projects {
		projectsIds = append(projectsIds, strconv.FormatInt(project.ID, 10))
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID.String())
		for _, projResources := range project.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, projResources.ID.String())
		}
	}

	if len(resourcesIds) > 0 {
		if _, err := tx.Where("id IN ?", resourcesIds).Delete(&[]models.Resources{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project resources: %v", err)
			return err
		}
	}

	if len(orgUnitsIds) > 0 {
		if _, err := tx.Where("id IN ?", orgUnitsIds).Delete(&[]models.OrgUnitData{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete projects OrgUnitData: %v", err)
			return err
		}
	}

	if len(projectsIds) > 0 {
		if _, err := tx.Where("id IN ?", projectsIds).Delete(&[]models.Project{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete projects: %v", err)
			return err
		}
	}
	return nil
}

func DeleteClusterDepartments(ctx context.Context, tenantId string, clusterId uuid.UUID) error {
	departments, err := GetDepartmentsByCluster(ctx, tenantId, clusterId, nil, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get departments: %v", err)
		return err
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	resources := []models.Resources{}
	orgUnits := []models.OrgUnitData{}
	resourcesIds := []uuid.UUID{}
	orgUnitsIds := []uuid.UUID{}
	departmentsIds := []int64{}
	for _, department := range *departments {
		departmentsIds = append(departmentsIds, department.ID)
		orgUnitsIds = append(orgUnitsIds, department.OrgUnitData.ID)
		orgUnits = append(orgUnits, department.OrgUnitData)
		resources = append(resources, department.OrgUnitData.Resources...)
		for _, deptResources := range department.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, deptResources.ID)
		}
	}
	if _, err := tx.Table(models.ResourcesTable).Where(resourcesIds).Delete(&resources); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete departments resources: %v", err)
		return err
	}
	if _, err := tx.Where("org_unit_data_id in ?", orgUnitsIds).Delete(&models.OvertimeData{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete departments OrgUnitData: %v", err)
		return err
	}
	if _, err := tx.Where(orgUnitsIds).Delete(&orgUnits); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete departments OrgUnitData: %v", err)
		return err
	}

	if _, err := tx.Where(departmentsIds).Delete(departments); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete departments: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete cluster departments: %v", err)
		return err
	}
	committed = true
	InvalidateAllCache(ctx, tenantId, clusterId.String())

	return nil
}

func GetDefaultDepartment(ctx context.Context, tenantId string, clusterId uuid.UUID, permittedScopes *openapi.PermittedScopes) (*models.Department, error) {
	department := &models.Department{
		ClusterId: clusterId,
		TenantId:  tenantId,
		Name:      defaultDepartmentName,
	}
	return getDepartment(ctx, department, models.SlimVerbosity, tenantId, permittedScopes)
}

func getPermittedScopesForDepartments(permittedScopes openapi.PermittedScopes) string {
	return getPermittedScopesForDepartmentsWithColumn(permittedScopes, "id")
}

func getPermittedScopesForDepartmentsWithColumn(permittedScopes openapi.PermittedScopes, idColumn string) string {
	orClause := "(false"
	if len(permittedScopes.Clusters) > 0 {
		orClause = orClause + " OR cluster_id IN " + ListItemsForQuery(permittedScopes.Clusters)
	}
	if len(permittedScopes.Departments) > 0 {
		orClause = orClause + " OR " + idColumn + " IN " + ListItemsForQuery(permittedScopes.Departments)
	}
	orClause += ")"
	return orClause
}

func extendedViewResultToDepartments(viewResult []models.DepartmentExtendedView) []models.Department {
	departments := []models.Department{}
	for _, result := range viewResult {
		departments = append(departments, extendedViewResultToDepartment(result))
	}
	return departments
}

func viewResultToDepartments(viewResult []models.DepartmentView) []models.Department {
	departments := []models.Department{}
	for _, result := range viewResult {
		departments = append(departments, viewResultToDepartment(result))
	}
	return departments
}

func syncViewResultToDepartments(viewResult []*models.DepartmentSyncView) []models.Department {
	departments := []models.Department{}
	for i := range viewResult {
		result := viewResult[i]
		departments = append(departments, syncViewResultToDepartment(result))
	}
	return departments
}

func viewResultToDepartment(result models.DepartmentView) models.Department {
	return models.Department{
		IntBaseModel:  result.IntBaseModel,
		TenantId:      result.TenantId,
		ClusterId:     result.ClusterId,
		Name:          result.Name,
		Description:   result.Description,
		ParentID:      result.ParentID,
		OrgUnitDataID: result.OrgUnitDataID,
		OrgUnitData:   orgUnitDataFromView(result.OrgUnitDataView, result.OrgUnitDataID),
	}
}

func syncViewResultToDepartment(result *models.DepartmentSyncView) models.Department {
	department := models.Department{
		IntBaseModel:  result.IntBaseModel,
		TenantId:      result.TenantId,
		ClusterId:     result.ClusterId,
		Name:          result.Name,
		ParentID:      result.ParentID,
		OrgUnitDataID: result.OrgUnitDataID,
		OrgUnitData: models.OrgUnitData{
			BaseModel: dbmodels.BaseModel{
				ID: result.OrgUnitDataID,
			},
			Resources: models.ConvertRawToResources(result.ResourcesJson),
		},
	}

	department.UpdatedAt = result.CalculatedUpdatedAt

	return department
}

func extendedViewResultToDepartment(result models.DepartmentExtendedView) models.Department {
	return models.Department{
		IntBaseModel:  result.IntBaseModel,
		TenantId:      result.TenantId,
		ClusterId:     result.ClusterId,
		Name:          result.Name,
		Description:   result.Description,
		ParentID:      result.ParentID,
		OrgUnitDataID: result.OrgUnitDataID,
		OrgUnitData:   orgUnitDataFromView(result.OrgUnitDataView, result.OrgUnitDataID),
		DepartmentStatus: models.DepartmentStatus{QuotaStatus: models.QuotaStatus{
			GpuAllocated:                  result.GpuAllocated,
			GpuAllocatedNonPreemptible:    result.GpuAllocatedNonPreemptible,
			GpuRequested:                  result.GpuRequested,
			CpuAllocated:                  result.CpuAllocated,
			CpuAllocatedNonPreemptible:    result.CpuAllocatedNonPreemptible,
			CpuRequested:                  result.CpuRequested,
			MemoryAllocated:               result.MemoryAllocated,
			MemoryAllocatedNonPreemptible: result.MemoryAllocatedNonPreemptible,
			MemoryRequested:               result.MemoryRequested,
		}},
	}
}

func verbosityBeforeFilterHook(verbosity models.Verbosity, query orm.DB) orm.DB {
	switch verbosity {
	case models.MetaDataVerbosity:
		query = query.Table(models.DepartmentTable)
	case models.SlimVerbosity:
		query = query.Table(models.DepartmentViewName)
	case models.VerboseVerbosity:
		query = query.Table(models.DepartmentStatusViewName)
	}

	return query
}

func verbosityAfterFilterHookSingle(ctx context.Context, verbosity models.Verbosity, query orm.DB) (*models.Department, error) {
	switch verbosity {
	case models.MetaDataVerbosity:
		department := &models.Department{}
		if err := query.First(&department); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get department: %v", err)
			}
			return nil, err
		}
		return department, nil
	case models.SlimVerbosity:
		departmentView := models.DepartmentView{}
		if err := query.First(&departmentView); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get department: %v", err)
			}
			return nil, err
		}

		departmentResult := viewResultToDepartment(departmentView)
		return &departmentResult, nil
	case models.VerboseVerbosity:
		departmentView := models.DepartmentExtendedView{}
		if err := query.First(&departmentView); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get department: %v", err)
			}
			return nil, err
		}

		departmentResult := extendedViewResultToDepartment(departmentView)
		mapToVerboseVerbosity(ctx, &departmentResult)

		return &departmentResult, nil
	}

	return nil, fmt.Errorf("verbosity %v is not supported", verbosity)
}

func verbosityAfterFilterHookSlice(ctx context.Context, verbosity models.Verbosity, query orm.DB) (*[]models.Department, error) {
	switch verbosity {
	case models.MetaDataVerbosity:
		departments := &[]models.Department{}
		if err := query.Find(departments); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department: %v", err)
			return nil, err
		}
		return departments, nil
	case models.SlimVerbosity:
		departmentsViews := []models.DepartmentView{}
		if err := query.Find(&departmentsViews); err != nil {
			logging.Errorf("Failed to get department: %v", err)
			return nil, err
		}
		return ptr.To(viewResultToDepartments(departmentsViews)), nil
	case models.VerboseVerbosity:
		departmentsExtendedViews := []models.DepartmentExtendedView{}
		if err := query.Find(&departmentsExtendedViews); err != nil {
			logging.Errorf("Failed to get department: %v", err)
			return nil, err
		}
		return ptr.To(extendedViewResultToDepartments(departmentsExtendedViews)), nil
	}

	return nil, fmt.Errorf("verbosity %v is not supported", verbosity)
}

func UpdateDepartmentStatus(ctx context.Context, clusterId uuid.UUID, tenantId string, departmentId string, lastSyncTime *time.Time) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Table(models.DepartmentTable).Where("id = ?", departmentId).Where("tenant_id = ?", tenantId).Updates(map[string]interface{}{"status_last_updated_time": lastSyncTime}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update department status, error: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit department status update: %v", err)
		return err
	}
	committed = true
	InvalidateNonSyncCache(ctx, tenantId, clusterId.String())
	return nil
}

func GetDepartmentNamesByIds(ctx context.Context, departmentIds []string) ([]string, error) {
	departmentNames := []string{}
	if err := runaidb.DB().WithContext(ctx).Table(models.DepartmentTable).Where("id IN (?)", departmentIds).Pluck("name", &departmentNames); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get department names by ids: %v", err)
		return nil, err
	}
	return departmentNames, nil
}

func GetDeletedDepartments(ctx context.Context, tenantId string, deletedAfter time.Time) ([]models.Department, error) {
	departments := []models.Department{}
	if err := runaidb.DB().WithContext(ctx).Table(models.DepartmentTable).Unscoped().Where("tenant_id = ?", tenantId).Where("deleted_at is not null").Where("deleted_at > ?", deletedAfter).Find(&departments); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch departments: %v", err)
		return nil, err
	}

	return departments, nil
}

func GetDepartmentsByClusters(ctx context.Context, tenantId string, clusterIds []string) ([]models.Department, error) {
	departments := []models.Department{}
	if err := runaidb.DB().WithContext(ctx).Table(models.DepartmentTable).Where("tenant_id = ?", tenantId).Where("cluster_id in ?", clusterIds).Find(&departments); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch departments: %v", err)
		return nil, err
	}

	return departments, nil
}
