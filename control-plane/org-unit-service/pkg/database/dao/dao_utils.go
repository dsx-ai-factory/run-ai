/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"golang.org/x/exp/slices"

	"github.com/google/uuid"
)

func applyParameterizedFilters(query orm.DB, filters []queryapi.Filter) orm.DB {
	return query.WithEnricher(querydbenrich.Filters(filters))
}

func applyTSSearchParam(query orm.DB, search, column string) orm.DB {
	return query.WithEnricher(dbenrich.TSSearch(search, column))
}

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

func orgUnitNodeTypeOmitFields() []string {
	return []string{
		fmt.Sprintf("%s.%s.*", OrgUnitData, TrainNodeType),
		fmt.Sprintf("%s.%s.*", OrgUnitData, WorkspaceNodeType),
		fmt.Sprintf("%s.%s.*", OrgUnitData, TrainNodeTypeEffective),
		fmt.Sprintf("%s.%s.*", OrgUnitData, WorkspaceNodeTypeEffective),
	}
}

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

func ListItemsForQuery(items []string) string {
	if len(items) == 0 {
		return "()"
	}

	return "( '" + strings.Join(items, "', '") + "')"
}

func deleteOrgUnitDataInTx(tx orm.DB, orgUnitDataID uuid.UUID) error {
	if _, err := tx.Where("org_unit_data_id = ?", orgUnitDataID).Delete(&models.Resources{}); err != nil {
		return err
	}
	if _, err := tx.Where("org_unit_data_id = ?", orgUnitDataID).Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{}); err != nil {
		return err
	}
	if _, err := tx.Where("org_unit_data_id = ?", orgUnitDataID).Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{}); err != nil {
		return err
	}
	if _, err := tx.Where("id = ?", orgUnitDataID).Delete(&models.OrgUnitData{}); err != nil {
		return err
	}
	if _, err := tx.Where("org_unit_data_id = ?", orgUnitDataID).Delete(&models.OvertimeData{}); err != nil {
		return err
	}
	return nil
}

func deleteQuotaStatusInTx(tx orm.DB, quotaStatusID uuid.UUID) error {
	if _, err := tx.Where("id = ?", quotaStatusID).Delete(&models.QuotaStatus{}); err != nil {
		return err
	}
	return nil
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
		BaseModel: dbmodels.BaseModel{
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
