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
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"golang.org/x/exp/slices"
)

const deleteAll = "1=1"

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

	rowsAffected, err := runaidb.DB().WithContext(ctx).
		Table(models.WorkloadsQueueTableName).
		Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
			models.WorkloadsQueueColumnId,
			models.WorkloadsQueueColumnUpdateCount,
			models.WorkloadsQueueColumnTenantId),
			record.ID, expectedUpdateCount, record.TenantId).
		Updates(updates)

	if err != nil {
		dbErr := api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("Failed to update %s", record.DebugInfo()), nil)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	if rowsAffected == 0 {
		return api.NewError(http.StatusConflict,
			"workload was modified by a concurrent request; retry", nil)
	}

	return nil
}

func DbUpsertWorkload(ctx context.Context, qRecord *models.WorkloadsQueueRecord) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to upsert %s", qRecord.DebugInfo()), nil)

	_, err := runaidb.DB().WithContext(ctx).Table(models.WorkloadsQueueTableName).Save(&qRecord)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbGetWorkload(ctx context.Context, workloadId string, permittedScopes *authorization.PermittedScopes) (*models.WorkloadsQueueRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get by id: %s", workloadId), nil)

	var record models.WorkloadsQueueRecord

	filters := models.WorkloadsQueueRecord{
		ID: workloadId,
	}

	tx := runaidb.DB().WithContext(ctx).Model(record).Unscoped().Where(&filters)
	tx = addPermittedScopesQueries(tx, ctx, permittedScopes)

	if err := tx.First(&record); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	return &record, nil
}

func DbListByIdsToMap(ctx context.Context, ids []string, permittedScopes *authorization.PermittedScopes) (map[string]*models.WorkloadsQueueRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to get by ids", nil)

	var records []models.WorkloadsQueueRecord

	tx := runaidb.DB().WithContext(ctx).Model(&models.WorkloadsQueueRecord{}).
		Where(fmt.Sprintf("%s in ?", models.WorkloadsQueueColumnId), ids)
	tx = addPermittedScopesQueries(tx, ctx, permittedScopes)

	if _, err := tx.Scan(&records); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
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

	_, err := runaidb.DB().WithContext(ctx).
		Model(models.WorkloadsQueueRecord{}).
		Where(&filters).
		Where(fmt.Sprintf("%s > ?", models.WorkloadsQueueColumnUpdateCount), afterUpdate).
		Limit(bulkSize).
		Order(models.WorkloadsQueueColumnUpdateCount).
		Scan(&records)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	return records, nil
}

func addPermittedScopesQueries(tx orm.DB, ctx context.Context, permittedScopes *authorization.PermittedScopes) orm.DB {
	tenantId := auth.TenantIdOf(ctx)

	// no matter which permissions the tenant has, we always limit the search by tenant
	tx = tx.Where("tenant_id = ?", tenantId)

	if permittedScopes == nil {
		return tx
	}

	if convutils.BoolIsTrue(permittedScopes.System) {
		return tx // allowed to see entire system
	}

	tenantIdStr := strconv.Itoa(int(tenantId))

	if slices.Contains(permittedScopes.Tenants, tenantIdStr) {
		return tx // allowed to see entire tenant
	}

	orExpression := "(false"

	if len(permittedScopes.Clusters) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+listItemsForQuery(permittedScopes.Clusters), models.WorkloadsQueueColumnClusterId)
	}

	if len(permittedScopes.Departments) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+listItemsForQuery(permittedScopes.Departments), models.WorkloadsQueueColumnDepartmentId)
	}

	if len(permittedScopes.Projects) > 0 {
		orExpression += fmt.Sprintf(" or %s IN "+listItemsForQuery(permittedScopes.Projects), models.WorkloadsQueueColumnProjectId)
	}

	orExpression += ")"

	return tx.Where(orExpression)
}

func listItemsForQuery(ids []string) string {
	base := "("
	for _, id := range ids {
		base += fmt.Sprintf(" '%s',", id)
	}
	return strings.TrimSuffix(base, ",") + ")"
}

// CleanWorkloadsOfScope is triggered by the cleanup pubsub notifications.
func CleanWorkloadsOfScope(ctx context.Context, tenantId int32, clusterId *string, departmentId *string, projectId *string) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to cleanup records for tenant %d", tenantId), nil)

	tx := runaidb.DB().WithContext(ctx).
		Table(models.WorkloadsQueueTableName).
		Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnTenantId), tenantId)

	if clusterId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnClusterId), clusterId)
	}

	if departmentId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnDepartmentId), departmentId)
	}

	if projectId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.WorkloadsQueueColumnProjectId), projectId)
	}

	if _, err := tx.Unscoped().Delete(deleteAll); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	return nil
}
