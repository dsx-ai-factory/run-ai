/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	telemetry "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"
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
	ProjectsCacheHashField       = "projects"
	ProjectsSyncCacheHashField   = "projectsSync"
	MaxConcurrentDescendantFetch = 5
)

func CreateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	if _, err := tx.Omit(orgUnitNodeTypeOmitFields()...).Create(project); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed project creation: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit project creation: %v", err)
		return nil, err
	}
	committed = true
	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project, nil
}

func CreateProjects(ctx context.Context, projects []*models.Project) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Omit(orgUnitNodeTypeOmitFields()...).
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(projects, 20); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed projects creation: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit projects creation: %v", err)
		return err
	}

	committed = true
	return nil
}

func GetProjectsByParent(ctx context.Context, parentId int64, tenantId string, verbosity models.Verbosity) (*[]models.Project, error) {
	projects := &[]models.Project{}
	query := runaidb.DB().WithContext(ctx)
	if verbosity == models.MetaDataVerbosity {
		query = query.Table(models.ProjectTable)
		if err := query.Where(&models.Project{ParentID: &parentId, TenantId: tenantId}).Find(&projects); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project: %v", err)
			return nil, err
		}
		return projects, nil
	}

	query = query.Table(models.ProjectViewName)
	projectsViews := []models.ProjectView{}
	if err := query.Where(&models.ProjectView{ParentID: &parentId, TenantId: tenantId}).Find(&projectsViews); err != nil {
		logging.Errorf("Failed to get project: %v", err)
		return nil, err
	}
	return ptr.To(viewResultToProjects(projectsViews)), nil
}

// GetProjectsIdsByDepartmentIds - only populates id and parent_id column for faster fetch
func GetProjectsIdsByDepartmentIds(ctx context.Context, tenantId string, departmentIds []int64) ([]models.Project, error) {
	projects := []models.Project{}
	query := runaidb.DB().WithContext(ctx)

	query = query.Table(models.ProjectTable)
	if _, err := query.Where(&models.Project{}).Where("tenant_id = ?", tenantId).Where("parent_id IN ?", departmentIds).Where("deleted_at IS NULL").Select("id", "parent_id").Scan(&projects); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get projects Ids by department IDs: %v", err)
		return nil, err
	}
	return projects, nil
}

func GetProjectById(ctx context.Context, projectId int64, tenantId string, verbosity models.Verbosity) (*models.Project, error) {
	project := &models.Project{
		IntBaseModel: models.IntBaseModel{ID: projectId},
		TenantId:     tenantId,
	}
	return getProject(ctx, project, verbosity)
}

func GetProjectByIdWithPermittedScopes(ctx context.Context, projectId int64, tenantId string, permittedScopes *openapi.PermittedScopes) (*models.Project, error) {
	projectView := models.ProjectView{
		IntBaseModel: models.IntBaseModel{ID: projectId},
		TenantId:     tenantId,
	}
	query := runaidb.DB().WithContext(ctx).
		Table(models.ProjectViewName).
		Where(projectView)

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjects(*permittedScopes))
		}
	}

	if err := query.First(&projectView); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.Errorf("Failed get project: %v", err)
		}
		return nil, err
	}
	return ptr.To(viewResultToProject(projectView)), nil
}

func DeleteClusterProjects(ctx context.Context, tenantId string, clusterId uuid.UUID) error {
	projects, err := GetProjectsByCluster(ctx, tenantId, clusterId, nil, models.SlimVerbosity)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project: %v", err)
		return err
	}

	if len(projects) == 0 {
		return nil
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	resourcesIds := []uuid.UUID{}
	orgUnitsIds := []uuid.UUID{}
	projectsIds := []int64{}
	for _, project := range projects {
		projectsIds = append(projectsIds, project.ID)
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID)
		for _, projResources := range project.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, projResources.ID)
		}
	}

	if len(resourcesIds) > 0 {
		if _, err := tx.Table(models.ResourcesTable).Where(resourcesIds).Delete(&[]models.Resources{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project resources: %v", err)
			return err
		}
	}

	if len(orgUnitsIds) > 0 {
		if _, err := tx.Where("org_unit_data_id in ?", orgUnitsIds).Delete(&models.OvertimeData{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete projects OrgUnitData: %v", err)
			return err
		}
		if _, err := tx.Table(models.OrgUnitDataTable).Where(orgUnitsIds).Delete(&[]models.OrgUnitData{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete projects OrgUnitData: %v", err)
			return err
		}
	}

	if len(projectsIds) > 0 {
		if _, err := tx.Table(models.ProjectTable).Where(projectsIds).Delete(&[]models.Project{}); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete projects: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete projects: %v", err)
		return err
	}
	committed = true
	InvalidateAllCache(ctx, tenantId, clusterId.String())

	return nil
}

func GetProjectByIdAndCluster(ctx context.Context, projectId int64, tenantId string, clusterId uuid.UUID, verbosity models.Verbosity) (*models.Project, error) {
	project := &models.Project{
		IntBaseModel: models.IntBaseModel{ID: projectId},
		TenantId:     tenantId,
		ClusterId:    clusterId,
	}
	return getProject(ctx, project, verbosity)
}

func GetProjectsByCluster(ctx context.Context, tenantId string, clusterId uuid.UUID, permittedScopes *openapi.PermittedScopes, verbosity models.Verbosity) ([]*models.Project, error) {
	GetProjectsByClusterInner := func() ([]*models.Project, error) {
		projects := []*models.Project{}
		projectsList, err := GetProjects(ctx, tenantId, []queryapi.Filter{
			{
				Field:    "cluster_id",
				Operator: "=",
				Values:   []string{clusterId.String()},
			},
		}, "", "", 0, 0, "", permittedScopes, false, verbosity)
		if err != nil {
			return nil, err
		}
		for _, project := range *projectsList {
			projects = append(projects, &project)
		}
		return projects, nil
	}

	var projects []*models.Project
	var err error
	if IsEligibleForCachedData(permittedScopes, tenantId, clusterId.String()) {
		projectsPointer, err := cache.Get[[]*models.Project](ctx, GetCacheKey(ProjectsCacheHashField, tenantId, clusterId.String()))
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get cached projects. err %+v", err)
		}
		if projectsPointer != nil {
			projects = *projectsPointer
		}
		if projects == nil {
			projects, err = GetProjectsByClusterInner()
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get projects in cluster %s. err %+v", clusterId, err)
				return nil, err
			}
			if err := cache.Set[[]*models.Project](ctx, GetCacheKey(ProjectsCacheHashField, tenantId, clusterId.String()), projects); err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set cached projects. err %+v", err)
			}
		}
	}

	if projects == nil {
		projects, err = GetProjectsByClusterInner()
	}

	return projects, err
}

func GetClusterIdByProject(ctx context.Context, projectId int64, tenantId string) (string, error) {
	var clusterId string
	if err := runaidb.DB().WithContext(ctx).Model(&models.Project{}).Select("cluster_id").Where("id = ?", projectId).Where("tenant_id = ?", tenantId).First(&clusterId); err != nil {
		return "", err
	}
	return clusterId, nil
}

func getProject(ctx context.Context, project *models.Project, verbosity models.Verbosity) (*models.Project, error) {
	query := runaidb.DB().WithContext(ctx)
	if verbosity == models.MetaDataVerbosity {
		query = query.Table(models.ProjectTable)
		if err := query.Where(project).First(&project); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logging.Errorf("Failed get project: %v", err)
			}
			return nil, err
		}
		return project, nil
	}
	query = query.Table(models.ProjectViewName)
	projectsView := models.ProjectView{}
	if err := query.Where(project).First(&projectsView); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project: %v", err)
		}
		return nil, err
	}
	projectResult := viewResultToProject(projectsView)
	return &projectResult, nil
}

func GetProjectByNameAndCluster(ctx context.Context, name string, clusterId uuid.UUID, tenantId string, verbosity models.Verbosity) (*models.Project, error) {
	project := &models.Project{
		Name:      name,
		ClusterId: clusterId,
		TenantId:  tenantId,
	}
	return getProject(ctx, project, verbosity)
}

func GetProjects(
	ctx context.Context,
	tenantId string,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	search string,
	permittedScopes *openapi.PermittedScopes,
	includeDeleted bool,
	verbosity models.Verbosity,
) (*[]models.Project, error) {
	projects := &[]models.Project{}
	db := runaidb.DB()
	if includeDeleted {
		db = db.Unscoped()
	}
	query := db.WithContext(ctx)
	if verbosity == models.MetaDataVerbosity {
		query = query.Table(models.ProjectTable)
	} else {
		query = query.Table(models.ProjectViewName)
	}

	query = query.Where(&models.Project{
		TenantId: tenantId,
	})

	query = applyParameterizedFilters(query, filterBy)

	if search != "" {
		query = applyTSSearchParam(query, search, "search_vector")
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjects(*permittedScopes))
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

	if verbosity == models.MetaDataVerbosity {
		if err := query.Find(&projects); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get projects, err: %v", err)
			return nil, err
		}

		return projects, nil
	}
	projectsViews := []models.ProjectView{}
	if err := query.Find(&projectsViews); err != nil {
		logging.Errorf("Failed to get projects, err: %v", err)
		return nil, err
	}
	return ptr.To(viewResultToProjects(projectsViews)), nil
}

func GetProjectsByIds(ctx context.Context, tenantId string, clusterId *uuid.UUID, ids []string, permittedScopes *openapi.PermittedScopes) ([]*models.Project, error) {
	query := runaidb.DB().WithContext(ctx).Table(models.ProjectViewName)

	query = query.Where(&models.Project{
		TenantId: tenantId,
	}).Where("id in ?", ids)

	if clusterId != nil {
		query = query.Where("cluster_id = ?", *clusterId)
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjects(*permittedScopes))
		}
	}

	projectsViews := []models.ProjectView{}
	if err := query.Find(&projectsViews); err != nil {
		logging.Errorf("Failed to get projects, err: %v", err)
		return nil, err
	}
	projects := []*models.Project{}
	for _, result := range projectsViews {
		project := viewResultToProject(result)
		projects = append(projects, &project)
	}
	return projects, nil
}

func GetSyncProjects(ctx context.Context, tenantId string, clusterId uuid.UUID, updatedAfter time.Time, offset int32, limit int32, permittedScopes *openapi.PermittedScopes) ([]*models.Project, error) {
	if updatedAfter.IsZero() {
		// in this case we need to invalidate cache since it holds data from certain timestamp
		InvalidateAllCache(ctx, tenantId, clusterId.String())
	}
	GetSyncProjectsInner := func() ([]*models.ProjectSyncView, error) {
		projects := []*models.ProjectSyncView{}
		projectsList, err := GetSyncProjectsView(ctx, tenantId, clusterId, updatedAfter, 0, 0, permittedScopes)
		if err != nil {
			return nil, err
		}
		for i := range *projectsList {
			project := (*projectsList)[i]
			projects = append(projects, &project)
		}
		return projects, nil
	}
	var projects []*models.ProjectSyncView
	var err error

	projectsSyncPointer, err := cache.Get[[]*models.ProjectSyncView](ctx, GetCacheKey(ProjectsSyncCacheHashField, tenantId, clusterId.String()))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get cached project sync views. err %+v", err)
	}
	if projectsSyncPointer != nil {
		projects = *projectsSyncPointer
	}
	if projects == nil {
		projects, err = GetSyncProjectsInner()
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get project sync views in cluster %s. err %+v", clusterId, err)
			return nil, err
		}
		if err := cache.Set[[]*models.ProjectSyncView](ctx, GetCacheKey(ProjectsSyncCacheHashField, tenantId, clusterId.String()), projects); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set cached project sync views. err %+v", err)
		}
	}

	if offset != 0 {
		projects = projects[offset:]
	}
	if limit != 0 && len(projects) > int(limit) {
		projects = projects[:limit]
	}

	return syncViewResultToProjects(projects), nil
}

func GetSyncProjectsView(ctx context.Context, tenantId string, clusterId uuid.UUID, updatedAfter time.Time, offset int32, limit int32, permittedScopes *openapi.PermittedScopes) (*[]models.ProjectSyncView, error) {
	// Use optimized SQL function instead of the view — filters by tenant/cluster
	// before aggregating resources, avoiding full-table scans. See migration 02250002.
	query := runaidb.DB().WithContext(ctx).Table(
		"(SELECT * FROM get_sync_projects(?, ?)) AS sync_view", tenantId, clusterId)

	if !updatedAfter.IsZero() {
		query = query.Where("calculated_updated_at > ? OR "+
			"status_last_updated_time is null OR "+
			"calculated_updated_at > status_last_updated_time OR "+
			"(deleted_at is not null AND deleted_at > ?)", updatedAfter, updatedAfter)
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjects(*permittedScopes))
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

	projectsViews := []models.ProjectSyncView{}
	if err := query.Find(&projectsViews); err != nil {
		logging.Errorf("Failed to get projects sync, err: %v", err)
		return nil, err
	}
	return ptr.To(projectsViews), nil
}

func UpdateProject(ctx context.Context, project *models.Project, clearTrainNodeTypes bool, clearWorkspaceNodeTypes bool) (*models.Project, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	err := UpdateProjectTx(ctx, tx, project, clearTrainNodeTypes, clearWorkspaceNodeTypes)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit project update: %v", err)
		return nil, err
	}
	committed = true
	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project, nil
}

func UpdateProjectTx(ctx context.Context, tx orm.DB, project *models.Project, clearTrainNodeTypes bool, clearWorkspaceNodeTypes bool) error {
	if clearTrainNodeTypes {
		err := tx.Actual().Model(&models.OrgUnitData{BaseModel: dbmodels.BaseModel{ID: project.OrgUnitDataID}}).Association(TrainNodeType).Clear()
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project: %v", err)
			return err
		}
	}
	if clearWorkspaceNodeTypes {
		err := tx.Actual().Model(&models.OrgUnitData{BaseModel: dbmodels.BaseModel{ID: project.OrgUnitDataID}}).Association(WorkspaceNodeType).Clear()
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project: %v", err)
			return err
		}
	}

	projectWhere := models.Project{
		IntBaseModel: models.IntBaseModel{ID: project.ID},
		TenantId:     project.TenantId,
		Name:         project.Name,
	}
	omit := append(orgUnitNodeTypeOmitFields(),
		"deleted_at", "deleted_by", "deletion_request", "deletion_type", "force_delete",
		"status_last_updated_time", "status_namespace", "conditions", "quota_status_id",
	)
	if _, err := tx.Table(models.ProjectTable).Where(&projectWhere).Select("*").Omit(omit...).Updates(project); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project: %v", err)
		return err
	}

	return nil
}

func UpdateProjectStatus(ctx context.Context, tenantId string, clusterId uuid.UUID,
	project *models.Project, status models.ProjectStatus, resourcesStatuses map[string]models.ResourcesStatus,
) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	projectStatusUpdates := map[string]interface{}{
		"phase":                    status.Phase,
		"phase_message":            status.PhaseMessage,
		"status_namespace":         status.StatusNamespace,
		"conditions":               status.Conditions,
		"status_last_updated_time": status.StatusLastUpdatedTime,
	}
	if _, err := tx.Table(models.ProjectTable).
		Where("id = ?", project.ID).
		Where("tenant_id = ?", tenantId).
		Updates(projectStatusUpdates); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project status, error: %v", err)
		return err
	}

	quotaStatusUpdates := map[string]interface{}{
		"gpu_allocated":                    status.QuotaStatus.GpuAllocated,
		"cpu_allocated":                    status.QuotaStatus.CpuAllocated,
		"memory_allocated":                 status.QuotaStatus.MemoryAllocated,
		"gpu_allocated_non_preemptible":    status.QuotaStatus.GpuAllocatedNonPreemptible,
		"cpu_allocated_non_preemptible":    status.QuotaStatus.CpuAllocatedNonPreemptible,
		"memory_allocated_non_preemptible": status.QuotaStatus.MemoryAllocatedNonPreemptible,
		"gpu_requested":                    status.QuotaStatus.GpuRequested,
		"cpu_requested":                    status.QuotaStatus.CpuRequested,
		"memory_requested":                 status.QuotaStatus.MemoryRequested,
	}
	if _, err := tx.Table(models.QuotaStatusTable).Where("id = ?", project.QuotaStatusID).Updates(quotaStatusUpdates); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project's quota_status: %v", err)
		return err
	}
	for nodePool, statuses := range resourcesStatuses {
		if err := updateResourcesStatus(tx, *project, nodePool, statuses); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update project's node pool quota_status: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit project status update: %s", err)
		return err
	}
	committed = true
	InvalidateNonSyncCache(ctx, tenantId, clusterId.String())

	return nil
}

func updateResourcesStatus(
	tx orm.DB,
	p models.Project,
	nodePool string,
	statuses models.ResourcesStatus,
) error {
	resourceStatusUpdates := map[string]interface{}{
		"allocated_gpu":                    statuses.AllocatedGpu,
		"allocated_cpu":                    statuses.AllocatedCpu,
		"allocated_memory":                 statuses.AllocatedMemory,
		"allocated_non_preemptible_gpu":    statuses.AllocatedNonPreemptibleGpu,
		"allocated_non_preemptible_cpu":    statuses.AllocatedNonPreemptibleCpu,
		"allocated_non_preemptible_memory": statuses.AllocatedNonPreemptibleMemory,
		"requested_gpu":                    statuses.RequestedGpu,
		"requested_cpu":                    statuses.RequestedCpu,
		"requested_memory":                 statuses.RequestedMemory,
	}
	_, err := tx.Table(models.ResourcesTable).
		Where("org_unit_data_id = ?", p.OrgUnitDataID).
		Where("node_pool_name = ?", nodePool).
		Updates(resourceStatusUpdates)
	return err
}

func CountProjectByNamespace(ctx context.Context, clusterId uuid.UUID, namespace string) (int64, error) {
	var projectCount int64
	err := runaidb.DB().WithContext(ctx).
		Table(models.ProjectTable).
		Where("cluster_id = ?", clusterId.String()).
		Where("requested_namespace = ? OR status_namespace = ?", namespace, namespace).
		Where("deleted_at is null").
		Count(&projectCount)
	if err != nil {
		return projectCount, err
	}
	return projectCount, nil
}

func CountProjects(ctx context.Context, tenantId string, permittedScopes *openapi.PermittedScopes, filterBy []queryapi.Filter, search string, verbosity models.Verbosity) (int64, error) {
	query := runaidb.DB().WithContext(ctx)
	if verbosity == models.MetaDataVerbosity {
		query = query.Table(models.ProjectTable)
	} else {
		query = query.Table(models.ProjectViewName)
	}

	query = query.Where("deleted_at IS NULL").Where("tenant_id = ?", tenantId)

	query = applyParameterizedFilters(query, filterBy)

	if search != "" {
		query = applyTSSearchParam(query, search, "search_vector")
	}

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			query = query.Where(getPermittedScopesForProjects(*permittedScopes))
		}
	}

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count projects: %v", err)
		return 0, err
	}

	return count, nil
}

func GetProjectsRanking(ctx context.Context, tenantId string, clusterId string, permittedScopes *openapi.PermittedScopes,
	columnNames RankingColumnNames, orderBy string, orderType api.RankingOrderType, limit int32, departmentId string,
) ([]api.ProjectGetRanking, error) {
	results, err := GetOrgUnitsRanking(ctx, tenantId, clusterId, permittedScopes,
		columnNames, orderBy, orderType, limit, false, departmentId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to query projects ranking. clusterId: %s, orderBy: %s, orderType: %s, limit: %d, error: %v",
			clusterId, orderBy, orderType, limit, err,
		)
		return nil, err
	}

	return results, nil
}

func getPermittedScopesForProjects(permittedScopes openapi.PermittedScopes) string {
	return GetPermittedScopesForProjectsWithColumn(permittedScopes, "id")
}

func GetPermittedScopesForProjectsWithColumn(permittedScopes openapi.PermittedScopes, idColumn string) string {
	orClause := "(false"
	if len(permittedScopes.Clusters) > 0 {
		orClause = orClause + " OR cluster_id IN " + ListItemsForQuery(permittedScopes.Clusters)
	}
	if len(permittedScopes.Departments) > 0 {
		orClause = orClause + " OR parent_id IN " + ListItemsForQuery(permittedScopes.Departments)
	}
	if len(permittedScopes.Projects) > 0 {
		orClause = orClause + " OR " + idColumn + " IN " + ListItemsForQuery(permittedScopes.Projects)
	}
	orClause += ")"
	return orClause
}

func viewResultToProjects(viewResult []models.ProjectView) []models.Project {
	projects := []models.Project{}
	for _, result := range viewResult {
		projects = append(projects, viewResultToProject(result))
	}
	return projects
}

func syncViewResultToProjects(viewResult []*models.ProjectSyncView) []*models.Project {
	projects := []*models.Project{}
	for _, result := range viewResult {
		projects = append(projects, syncViewResultToProject(*result))
	}
	return projects
}

func viewResultToProject(result models.ProjectView) models.Project {
	return models.Project{
		IntBaseModel:          result.IntBaseModel,
		TenantId:              result.TenantId,
		ClusterId:             result.ClusterId,
		Name:                  result.Name,
		Description:           result.Description,
		RequestedNamespace:    result.RequestedNamespace,
		EnforceRunaiScheduler: result.EnforceRunaiScheduler,
		ParentID:              result.ParentID,
		ParentName:            result.ParentName,
		DeletionRequest:       result.DeletionRequest,
		DeletionType:          result.DeletionType,
		ForceDelete:           result.ForceDelete,
		OrgUnitDataID:         result.OrgUnitDataID,
		OrgUnitData:           orgUnitDataFromView(result.OrgUnitDataView, result.OrgUnitDataID),
		ProjectStatus: models.ProjectStatus{
			Phase:           result.Phase,
			PhaseMessage:    result.PhaseMessage,
			StatusNamespace: result.StatusNamespace,
			Conditions:      result.Conditions,
			QuotaStatusID:   result.QuotaStatusID,
			QuotaStatus:     result.QuotaStatus,
		},
	}
}

func syncViewResultToProject(result models.ProjectSyncView) *models.Project {
	orgUnitDataSyncView := result.OrgUnitDataSyncView
	project := &models.Project{
		IntBaseModel:          result.IntBaseModel,
		TenantId:              result.TenantId,
		ClusterId:             result.ClusterId,
		Name:                  result.Name,
		RequestedNamespace:    result.RequestedNamespace,
		EnforceRunaiScheduler: result.EnforceRunaiScheduler,
		ParentID:              result.ParentID,
		ParentName:            result.ParentName,
		DeletionRequest:       result.DeletionRequest,
		DeletionType:          result.DeletionType,
		ForceDelete:           result.ForceDelete,
		OrgUnitDataID:         result.OrgUnitDataID,
		OrgUnitData: models.OrgUnitData{
			BaseModel: dbmodels.BaseModel{
				ID: result.OrgUnitDataID,
			},
			Resources:                         models.ConvertRawToResources(orgUnitDataSyncView.ResourcesJson),
			InteractiveJobTimeLimit:           orgUnitDataSyncView.InteractiveJobTimeLimit,
			InteractiveJobMaxIdleDuration:     orgUnitDataSyncView.InteractiveJobMaxIdleDuration,
			InteractiveJobPreemptIdleDuration: orgUnitDataSyncView.InteractiveJobPreemptIdleDuration,
			TrainingJobTimeLimit:              orgUnitDataSyncView.TrainingJobTimeLimit,
			TrainingJobMaxIdleDuration:        orgUnitDataSyncView.TrainingJobMaxIdleDuration,
			DefaultNodePoolList:               orgUnitDataSyncView.DefaultNodePoolList,
			TrainingNodeTypes:                 models.ConvertRawToNodeTypes(orgUnitDataSyncView.TrainingNodeTypesView),
			WorkspaceNodeTypes:                models.ConvertRawToNodeTypes(orgUnitDataSyncView.WorkspaceNodeTypesView),
		},
	}

	project.UpdatedAt = result.CalculatedUpdatedAt

	return project
}

func GetProjectResourcesTelemetry(
	ctx context.Context,
	quotaColumn string,
	tenantId string,
	permittedScopes *openapi.PermittedScopes,
	clusterId string,
	nodepoolId string,
	departmentId string,
	projectId string,
	groupBy [][]string,
) (*[]telemetry.TelemetryGroupValue, error) {
	telemetryGroupValues := []telemetry.TelemetryGroupValue{}

	selectPhrase := fmt.Sprintf("SUM(%s) as data", quotaColumn)
	for i, groups := range groupBy {
		selectPhrase = fmt.Sprintf("%s, %s as group%d_main_value", selectPhrase, groups[0], i+1)
		if len(groups) > 1 {
			selectPhrase = fmt.Sprintf("%s, %s as group%d_secondary_value", selectPhrase, groups[1], i+1)
		}
	}

	query := runaidb.DB().WithContext(ctx).
		Table(models.ProjectTable).
		Joins("JOIN resources on project.org_unit_data_id = resources.org_unit_data_id").
		Select(selectPhrase).
		Where("tenant_id = ?", tenantId).
		Where("project.deleted_at IS NULL").
		Where("resources.deleted_at IS NULL")

	if permittedScopes != nil {
		tenantPermitted := IsTenantPermitted(tenantId, permittedScopes)
		if !tenantPermitted {
			// qualify "id" because this query JOINs project with resources (both have an id column)
			query = query.Where(GetPermittedScopesForProjectsWithColumn(*permittedScopes, "project.id"))
		}
	}

	if clusterId != "" {
		query = query.Where("cluster_id = ?", clusterId)
	}
	if nodepoolId != "" {
		query = query.Where("node_pool_id = ?", nodepoolId)
	}
	if departmentId != "" {
		query = query.Where("parent_id = ?", departmentId)
	}
	if projectId != "" {
		query = query.Where("project.id = ?", projectId)
	}
	for _, groups := range groupBy {
		query = query.Group(groups[0])
		if len(groups) > 1 {
			query = query.Group(groups[1])
		}
	}
	if _, err := query.Scan(&telemetryGroupValues); err != nil {
		return &telemetryGroupValues, err
	}

	return &telemetryGroupValues, nil
}

func GetDeletedProjects(ctx context.Context, tenantId string, deletedAfter time.Time) ([]models.Project, error) {
	projects := []models.Project{}
	if err := runaidb.DB().WithContext(ctx).Table(models.ProjectTable).Unscoped().Where("tenant_id = ?", tenantId).Where("deleted_at is not null").Where("deleted_at > ?", deletedAfter).Find(&projects); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch projects: %v", err)
		return nil, err
	}

	return projects, nil
}

func GetProjectsByClusters(ctx context.Context, tenantId string, clusterIds []string) ([]models.Project, error) {
	projects := []models.Project{}
	if err := runaidb.DB().WithContext(ctx).Table(models.ProjectTable).Where("tenant_id = ?", tenantId).Where("cluster_id in ?", clusterIds).Find(&projects); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch projects: %v", err)
		return nil, err
	}

	return projects, nil
}
