/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

func DbGetSyncStatus(ctx context.Context, workloadId string) (*models.SyncStatusRecord, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get sync status for: %s", workloadId), nil)

	var record models.SyncStatusRecord

	filters := models.SyncStatusRecord{
		WorkloadId: workloadId,
		TenantId:   auth.TenantIdOf(ctx),
	}

	if err := runaidb.DB().WithContext(ctx).Model(record).Unscoped().Where(&filters).First(&record); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetails(err)
	}

	return &record, nil
}

func DbUpdateSentToAgent(ctx context.Context, clusterId string, updates api.WorkloadsV2Updates, workloadsScope map[string]models.WorkloadsQueueRecord) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update for cluster %s", clusterId), nil)

	tenantId := auth.TenantIdOf(ctx)

	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, update := range updates.Workloads {
		record := models.SyncStatusRecord{
			WorkloadId: update.Id,

			TenantId:  tenantId,
			ClusterId: clusterId,

			Deleting: update.DesiredPhase == api.DESIREDPHASE_DELETED,

			SentToAgent: update.UpdateCount,
		}

		workload, found := workloadsScope[update.Id]
		if !found {
			// workload not found in map this actually cannot happen, because the caller builds them
			// both based on same workloads, but to be on the safe side we log it. nothing much will
			// happen, the department and project id will be missing, but they are there only for cleanup
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("details of workload %s not found", update.Id)
		} else {
			record.DepartmentId = workload.DepartmentId
			record.ProjectId = workload.ProjectId
		}

		rowsAffected, err := tx.Model(&record).Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{
						Name: models.SyncStatusColumnWorkloadId,
					},
				},
				DoUpdates: clause.AssignmentColumns(
					[]string{
						models.SyncStatusColumnSentToAgent,
						models.SyncStatusColumnDeleting,
					},
				),
			},
		).Create(&record)

		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("update process by cluster of %s to %d: %s: %v", update.Id, update.UpdateCount, dbErr.Message, err)
		} else if rowsAffected == 0 {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("update process by cluster of %s to %d: update ignored, probably already deleted", update.Id, update.UpdateCount)
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	committed = true

	return nil
}

// CleanSyncStatusOfScope is triggered by the cleanup pubsub notifications.
func CleanSyncStatusOfScope(ctx context.Context, tenantId int32, clusterId *string, departmentId *string, projectId *string) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to cleanup status info record for tenant %d", tenantId), nil)

	tx := runaidb.DB().WithContext(ctx).
		Table(models.SyncStatusTableName).
		Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnTenantId), tenantId)

	if clusterId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnClusterId), clusterId)
	}

	if departmentId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnDepartmentId), departmentId)
	}

	if projectId != nil {
		tx = tx.Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnProjectId), projectId)
	}

	if _, err := tx.Unscoped().Delete(deleteAll); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	return nil
}
