package dao

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
)

func DbApplyClusterReport(ctx context.Context, clusterId string, clusterReports *api.ClusterSyncReport) *api.Error {
	if len(clusterReports.Updates) == 0 {
		return nil
	}

	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update for cluster %s", clusterId), nil)

	tenantId := auth.TenantIdOf(ctx)

	tx := dao_utils.DB().WithContext(ctx).Begin()

	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

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

		result := tx.Model(&filters).Where(&filters).Unscoped().
			UpdateColumn(models.SyncStatusColumnProcessedByAgent, report.UpdateCount).
			UpdateColumn(models.SyncStatusColumnSyncMessage, report.Message)

		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%d: update count %s workload %s: %v", report.UpdateCount, report.Id, dbErr.Message, result.Error)
		} else if result.RowsAffected == 0 {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("%d: update count %s workload: cluster report ignored, probably newer update is pending", report.UpdateCount, report.Id)
		} else {
			allIds = append(allIds, report.Id)
		}
	}

	// find and delete records that their deletion was processed by the cluster
	deletedFromCluster := make([]models.SyncStatusRecord, 0, len(allIds))

	deletedIds := make([]string, 0, len(allIds))

	result := tx.Table(models.SyncStatusTableName).
		Where(fmt.Sprintf("%s in ?", models.SyncStatusColumnWorkloadId), allIds).
		Where(fmt.Sprintf("%s = ?", models.SyncStatusColumnDeleting), true).
		Where(fmt.Sprintf("%s = %s", models.SyncStatusColumnSentToAgent, models.SyncStatusColumnProcessedByAgent)).
		Scan(&deletedFromCluster)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to locate records for deletion %s: %v", dbErr.Message, result.Error)
	} else {
		//
		//   log those records that the cluster could not physically delete. we will still remove
		//   them from the control plane, but we need information about it in the log
		//
		for _, deletedFromCluster := range deletedFromCluster {
			if deletedFromCluster.Deleting && convutils.IsStringHasValue(deletedFromCluster.SyncMessage) {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("agent failed to delete workload %s: %s: %v, but we will remove it from control plane",
					deletedFromCluster.WorkloadId, dbErr.Message, result.Error)
			}

			deletedIds = append(deletedIds, deletedFromCluster.WorkloadId)
		}
	}

	//
	//   delete the records that completed deletion processing from both tables
	//
	if len(deletedFromCluster) > 0 {
		result = tx.Table(models.SyncStatusTableName).Debug().
			Where(fmt.Sprintf("%s in ?", models.SyncStatusColumnWorkloadId), deletedIds).
			Delete(dao_utils.All)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete sync status records: %s: %v", dbErr.Message, result.Error)
		}

		result = tx.Table(models.WorkloadsQueueTableName).
			Where(fmt.Sprintf("%s in ?", models.WorkloadsQueueColumnId), deletedIds).
			Delete(dao_utils.All)
		if result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workloads from queue: %s: %v", dbErr.Message, result.Error)
		}
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	committed = true

	return nil
}
