/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"golang.org/x/exp/slices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	OrgUnitData                = "OrgUnitData"
	Resources                  = "Resources"
	DepartmentChildren         = "DepartmentChildren"
	ProjectChildren            = "ProjectChildren"
	TrainNodeType              = "TrainingNodeTypes"
	WorkspaceNodeType          = "WorkspaceNodeTypes"
	TrainNodeTypeEffective     = "TrainingNodeTypesEffective"
	WorkspaceNodeTypeEffective = "WorkspaceNodeTypesEffective"
	CacheKeyPrefix             = "orgunitservice"
	CacheKeyFormat             = "%s-%s-%s"
)

func IsEligibleForCachedData(permittedScopes *openapi.PermittedScopes, tenantId string, clusterId string) bool {
	if permittedScopes == nil || (permittedScopes.System != nil && *permittedScopes.System) {
		return true
	}

	if slices.Contains(permittedScopes.Tenants, tenantId) || slices.Contains(permittedScopes.Clusters, clusterId) {
		return true
	}

	return false
}

func GetCacheKey(field string, tenantId string, clusterId string) string {
	return fmt.Sprintf(CacheKeyFormat, field, tenantId, clusterId)
}

func deleteOrgUnitDataInTx(tx *gorm.DB, orgUnitDataID uuid.UUID) (*gorm.DB, error) {
	result := tx.Where("org_unit_data_id = ?", orgUnitDataID).Delete(&models.Resources{})
	if result.Error != nil {
		return nil, result.Error
	}
	result = tx.Where("org_unit_data_id = ?", orgUnitDataID).Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{})
	if result.Error != nil {
		return nil, result.Error
	}
	result = tx.Where("org_unit_data_id = ?", orgUnitDataID).Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{})
	if result.Error != nil {
		return nil, result.Error
	}
	result = tx.Where("id = ?", orgUnitDataID).Delete(&models.OrgUnitData{})
	if result.Error != nil {
		return nil, result.Error
	}
	result = tx.Where("org_unit_data_id = ?", orgUnitDataID).Delete(&models.OvertimeData{})
	if result.Error != nil {
		return nil, result.Error
	}
	return result, nil
}

func deleteQuotaStatusInTx(tx *gorm.DB, quotaStatusID uuid.UUID) (*gorm.DB, error) {
	result := tx.Where("id = ?", quotaStatusID).Delete(&models.QuotaStatus{})
	if result.Error != nil {
		return nil, result.Error
	}
	return result, nil
}

func IsTenantPermitted(tenantId string, permittedScopes *openapi.PermittedScopes) bool {
	return isScopePermitted(tenantId, permittedScopes.Tenants)
}

func IsClusterPermitted(clusterId string, permittedScopes *openapi.PermittedScopes) bool {
	return isScopePermitted(clusterId, permittedScopes.Clusters)
}

func isScopePermitted(scopeId string, permittedScopes []string) bool {
	for _, permittedScope := range permittedScopes {
		if permittedScope == scopeId {
			return true
		}
	}
	return false
}

func orgUnitDataFromView(orgUnitDataView models.OrgUnitDataView, orgUnitDataId uuid.UUID) models.OrgUnitData {
	return models.OrgUnitData{
		BaseModel: dao_utils.BaseModel{
			ID: orgUnitDataId,
		},
		Resources:                         models.ConvertRawToResources(orgUnitDataView.ResourcesJson),
		TotalGpuQuota:                     orgUnitDataView.TotalGpuQuota,
		TotalCpuQuota:                     orgUnitDataView.TotalCpuQuota,
		TotalMemoryQuota:                  orgUnitDataView.TotalMemoryQuota,
		InteractiveJobTimeLimit:           orgUnitDataView.InteractiveJobTimeLimit,
		InteractiveJobMaxIdleDuration:     orgUnitDataView.InteractiveJobMaxIdleDuration,
		InteractiveJobPreemptIdleDuration: orgUnitDataView.InteractiveJobPreemptIdleDuration,
		TrainingJobTimeLimit:              orgUnitDataView.TrainingJobTimeLimit,
		TrainingJobMaxIdleDuration:        orgUnitDataView.TrainingJobMaxIdleDuration,
		DefaultNodePoolList:               orgUnitDataView.DefaultNodePoolList,
		Children:                          models.ConvertRawToChildren(orgUnitDataView.ChildrenView),
		TrainingNodeTypes:                 models.ConvertRawToNodeTypes(orgUnitDataView.TrainingNodeTypesView),
		WorkspaceNodeTypes:                models.ConvertRawToNodeTypes(orgUnitDataView.WorkspaceNodeTypesView),
		EffectiveData: models.EffectiveData{
			InteractiveJobTimeLimitEffective:                 orgUnitDataView.InteractiveJobTimeLimitEffective,
			InteractiveJobTimeLimitEffectiveOrigin:           orgUnitDataView.InteractiveJobTimeLimitEffectiveOrigin,
			InteractiveJobMaxIdleDurationEffective:           orgUnitDataView.InteractiveJobMaxIdleDurationEffective,
			InteractiveJobMaxIdleDurationEffectiveOrigin:     orgUnitDataView.InteractiveJobMaxIdleDurationEffectiveOrigin,
			InteractiveJobPreemptIdleDurationEffective:       orgUnitDataView.InteractiveJobPreemptIdleDurationEffective,
			InteractiveJobPreemptIdleDurationEffectiveOrigin: orgUnitDataView.InteractiveJobPreemptIdleDurationEffectiveOrigin,
			TrainingJobTimeLimitEffective:                    orgUnitDataView.TrainingJobTimeLimitEffective,
			TrainingJobTimeLimitEffectiveOrigin:              orgUnitDataView.TrainingJobTimeLimitEffectiveOrigin,
			TrainingJobMaxIdleDurationEffective:              orgUnitDataView.TrainingJobMaxIdleDurationEffective,
			TrainingJobMaxIdleDurationEffectiveOrigin:        orgUnitDataView.TrainingJobMaxIdleDurationEffectiveOrigin,
			DefaultNodePoolListEffective:                     orgUnitDataView.DefaultNodePoolListEffective,
			DefaultNodePoolListEffectiveOrigin:               orgUnitDataView.DefaultNodePoolListEffectiveOrigin,
			TrainingNodeTypesEffective:                       models.ConvertRawToNodeTypes(orgUnitDataView.TrainingNodeTypesViewEffective),
			TrainingNodeTypesEffectiveOrigin:                 orgUnitDataView.TrainingNodeTypesViewEffectiveOrigin,
			WorkspaceNodeTypesEffective:                      models.ConvertRawToNodeTypes(orgUnitDataView.WorkspaceNodeTypesViewEffective),
			WorkspaceNodeTypesEffectiveOrigin:                orgUnitDataView.WorkspaceNodeTypesViewEffectiveOrigin,
		},
		OvertimeData: models.OvertimeData{
			ID:                         orgUnitDataView.OvertimeDataID,
			AvgGpuAllocation24h:        orgUnitDataView.AvgGpuAllocation24h,
			AvgGpuUtilization24h:       orgUnitDataView.AvgGpuUtilization24h,
			AvgGpuMemoryUtilization24h: orgUnitDataView.AvgGpuMemoryUtilization24h,
			Avg24hRangeUpdatedAt:       orgUnitDataView.Avg24hRangeUpdatedAt,
			AvgGpuAllocation7d:         orgUnitDataView.AvgGpuAllocation7d,
			AvgGpuUtilization7d:        orgUnitDataView.AvgGpuUtilization7d,
			AvgGpuMemoryUtilization7d:  orgUnitDataView.AvgGpuMemoryUtilization7d,
			Avg7dRangeUpdatedAt:        orgUnitDataView.Avg7dRangeUpdatedAt,
			AvgGpuAllocation30d:        orgUnitDataView.AvgGpuAllocation30d,
			AvgGpuUtilization30d:       orgUnitDataView.AvgGpuUtilization30d,
			AvgGpuMemoryUtilization30d: orgUnitDataView.AvgGpuMemoryUtilization30d,
			Avg30dRangeUpdatedAt:       orgUnitDataView.Avg30dRangeUpdatedAt,
		},
	}
}

func InvalidateAllCache(ctx context.Context, tenantId string, clusterId string) {
	if err := cache.Del[[]models.Department](ctx, GetCacheKey(DepartmentsCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateAllCache] failed to delete key for Department")
	}
	if err := cache.Del[[]*models.Project](ctx, GetCacheKey(ProjectsCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateAllCache] failed to delete key for Project")
	}
	if err := cache.Del[[]*models.ProjectSyncView](ctx, GetCacheKey(ProjectsSyncCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateAllCache] failed to delete key for ProjectsSync")
	}
	if err := cache.Del[[]*models.DepartmentSyncView](ctx, GetCacheKey(DepartmentsSyncCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateAllCache] failed to delete key for DepartmentsSync")
	}
	if err := cache.Del[[]*models.DepartmentSyncView](ctx, GetCacheKey(DepartmentsV1CacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateAllCache] failed to delete key for DepartmentsV1")
	}
}

func InvalidateNonSyncCache(ctx context.Context, tenantId string, clusterId string) {
	if err := cache.Del[[]models.Department](ctx, GetCacheKey(DepartmentsCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateNonSyncCache] failed to delete key for Department")
	}
	if err := cache.Del[[]*models.Project](ctx, GetCacheKey(ProjectsCacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateNonSyncCache] failed to delete key for Project")
	}
	if err := cache.Del[[]*models.DepartmentSyncView](ctx, GetCacheKey(DepartmentsV1CacheHashField, tenantId, clusterId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[dao.InvalidateNonSyncCache] failed to delete key for DepartmentsV1")
	}
}
