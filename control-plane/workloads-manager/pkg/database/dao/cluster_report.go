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
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func DbApplyClusterReport(ctx context.Context, clusterId string, clusterReports *api.ClusterSyncReport) *api.Error {
	if len(clusterReports.Updates) == 0 {
		return nil
	}

	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update for cluster %s", clusterId), nil)

	tenantId := auth.TenantIdOf(ctx)

	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	allIds := make([]string, 0, len(clusterReports.Updates))

	for _, report := range clusterReports.Updates {
		//
		//   update the status with the information that the cluster reported
		//
		filters := models.SyncStatusRecord{
			WorkloadId:  report.Id,
			TenantId:    tenantId,
			ClusterId:   clusterId,
			SentToAgent: report.UpdateCount,
		}

		rowsAffected, err := tx.Model(&filters).Where(&filters).Unscoped().
			UpdateColumns(map[string]interface{}{
				models.SyncStatusColumnProcessedByAgent: report.UpdateCount,
				models.SyncStatusColumnSyncMessage:      report.Message,
			})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%d: update count %s workload %s: %v", report.UpdateCount, report.Id, dbErr.Message, err)
		} else if rowsAffected == 0 {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("%d: update count %s workload: cluster report ignored, probably newer update is pending", report.UpdateCount, report.Id)
		} else {
			allIds = append(allIds, report.Id)
		}
	}

	// find and delete records that their deletion was processed by the cluster
	deletedFromCluster := make([]models.SyncStatusRecord, 0, len(allIds))

	deletedIds := make([]string, 0, len(allIds))

	_, err := tx.Table(models.SyncStatusTableName).
		Where(fmt.Sprintf("%s in ?", models.SyncStatusColumnWorkloadId), allIds).
		Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnDeleting), true).
		Where(fmt.Sprintf("%s = %s", models.SyncStatusColumnSentToAgent, models.SyncStatusColumnProcessedByAgent)).
		Scan(&deletedFromCluster)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to locate records for deletion %s: %v", dbErr.Message, err)
	} else {
		//
		//   log those records that the cluster could not physically delete. we will still remove
		//   them from the control plane, but we need information about it in the log
		//
		for _, deletedFromCluster := range deletedFromCluster {
			if deletedFromCluster.Deleting && convutils.IsStringHasValue(deletedFromCluster.SyncMessage) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("agent failed to delete workload %s: %s: %v, but we will remove it from control plane",
					deletedFromCluster.WorkloadId, dbErr.Message, err)
			}

			deletedIds = append(deletedIds, deletedFromCluster.WorkloadId)
		}
	}

	//
	//   delete the records that completed deletion processing from both tables
	//
	if len(deletedFromCluster) > 0 {
		if _, err := tx.Table(models.SyncStatusTableName).
			Where(fmt.Sprintf("%s in ?", models.SyncStatusColumnWorkloadId), deletedIds).
			Delete(deleteAll); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete sync status records: %s: %v", dbErr.Message, err)
		}

		if _, err := tx.Table(models.WorkloadsQueueTableName).
			Where(fmt.Sprintf("%s in ?", models.WorkloadsQueueColumnId), deletedIds).
			Delete(deleteAll); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workloads from queue: %s: %v", dbErr.Message, err)
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return dbErr.WithDetails(err)
	}

	committed = true

	return nil
}
