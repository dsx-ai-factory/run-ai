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

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"

	authorization "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

// DbUpdateWorkloadOptimistic updates a workloads_queue record with optimistic concurrency control.
// Returns 409 if no row is affected (concurrent modification). Uses Updates() so GORM never INSERTs.
func DbUpdateWorkloadOptimistic(ctx context.Context, record *models.WorkloadsQueueRecord, expectedUpdateCount int64) *api.Error {
	updates := map[string]interface{}{
		models.WorkloadsQueueColumnRawManifest:       record.RawManifest,
		models.WorkloadsQueueColumnEffectiveManifest: record.EffectiveManifest,
		models.WorkloadsQueueColumnDesiredPhase:      record.DesiredPhase,
		models.WorkloadsQueueColumnUpdatedAt:         record.UpdatedAt,
		models.WorkloadsQueueColumnUpdatedBy:         record.UpdatedBy,
	}

	result := dao_utils.DB().WithContext(ctx).
		Table(models.WorkloadsQueueTableName).
		Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
			models.WorkloadsQueueColumnId,
			models.WorkloadsQueueColumnUpdateCount,
			models.WorkloadsQueueColumnTenantId),
			record.ID, expectedUpdateCount, record.TenantId).
		Updates(updates)

	if result.Error != nil {
		dbErr := api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("Failed to update %s", record.DebugInfo()), nil)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	if result.RowsAffected == 0 {
		return api.NewError(http.StatusConflict,
			"workload was modified by a concurrent request; retry", nil)
	}

	return nil
}

func DbUpsertWorkload(ctx context.Context, qRecord *models.WorkloadsQueueRecord) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to upsert %s", qRecord.DebugInfo()), nil)

	result := dao_utils.DB().WithContext(ctx).Table(models.WorkloadsQueueTableName).Save(&qRecord)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbGetWorkload(ctx context.Context, workloadId string, permittedScopes *authorization.PermittedScopes) (*models.WorkloadsQueueRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get by id: %s", workloadId), nil)

	var record models.WorkloadsQueueRecord

	filters := models.WorkloadsQueueRecord{
		ID: workloadId,
	}

	result := dao_utils.DB().WithContext(ctx).Model(record).Unscoped().Where(&filters)

	addPermittedScopesQueries(result, ctx, permittedScopes)

	result = result.First(&record)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &record, nil
}

func DbListByIdsToMap(ctx context.Context, ids []string, permittedScopes *authorization.PermittedScopes) (map[string]*models.WorkloadsQueueRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to get by ids", nil)

	var records []models.WorkloadsQueueRecord

	result := dao_utils.DB().WithContext(ctx).Model(&models.WorkloadsQueueRecord{}).
		Where(fmt.Sprintf("%s in ?", models.WorkloadsQueueColumnId), ids)

	addPermittedScopesQueries(result, ctx, permittedScopes)

	result = result.Scan(&records)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	asMap := make(map[string]*models.WorkloadsQueueRecord)

	for idx, record := range records {
		asMap[record.ID] = &records[idx]
	}

	return asMap, nil
}

func DbListUpdates(ctx context.Context, clusterId string, afterUpdate int64, bulkSize int) ([]models.WorkloadsQueueRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to list updates for cluster %s", clusterId), nil)

	var records []models.WorkloadsQueueRecord

	filters := models.WorkloadsQueueRecord{
		TenantId:  auth.TenantIdOf(ctx),
		ClusterId: clusterId,
	}

	result := dao_utils.DB().WithContext(ctx).
		Model(models.WorkloadsQueueRecord{}).
		Where(&filters).
		Where(fmt.Sprintf("%s > ?", models.WorkloadsQueueColumnUpdateCount), afterUpdate).
		Limit(bulkSize).
		Order(models.WorkloadsQueueColumnUpdateCount).
		Scan(&records)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return records, nil
}

func addPermittedScopesQueries(tx *gorm.DB, ctx context.Context, permittedScopes *authorization.PermittedScopes) {
	tenantId := auth.TenantIdOf(ctx)

	// no matter which permissions the tenant has, we always limit the search by tenant
	tx.Where("tenant_id = ?", tenantId)

	if permittedScopes == nil {
		return
	}

	if convutils.BoolIsTrue(permittedScopes.System) {
		return // allowed to see entire system
	}

	tenantIdStr := strconv.Itoa(int(tenantId))

	if slices.Contains(permittedScopes.Tenants, tenantIdStr) {
		return // allowed to see entire tenant
	}

	orExpression := "(false"

	if len(permittedScopes.Clusters) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+dao_utils.ListItemsForQuery(permittedScopes.Clusters), models.WorkloadsQueueColumnClusterId)
	}

	if len(permittedScopes.Departments) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+dao_utils.ListItemsForQuery(permittedScopes.Departments), models.WorkloadsQueueColumnDepartmentId)
	}

	if len(permittedScopes.Projects) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+dao_utils.ListItemsForQuery(permittedScopes.Projects), models.WorkloadsQueueColumnProjectId)
	}

	orExpression += ")"

	tx.Where(orExpression)
}

// CleanWorkloadsOfScope is triggered by the cleanup pubsub notifications.
func CleanWorkloadsOfScope(ctx context.Context, tenantId int32, clusterId *string, departmentId *string, projectId *string) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to cleanup records for tenant %d", tenantId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(models.WorkloadsQueueTableName).
		Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnTenantId), tenantId)

	if clusterId != nil {
		result = result.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnClusterId), clusterId)
	}

	if departmentId != nil {
		result = result.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnDepartmentId), departmentId)
	}

	if projectId != nil {
		result = result.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnProjectId), projectId)
	}

	result = result.Unscoped().Delete(dao_utils.All)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}
