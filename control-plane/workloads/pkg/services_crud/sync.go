/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"net/http"
	"regexp"
)

var (
	workloadNotFoundPattern = "Workload \".*\" not found"
	workloadNotFoundRegexp  = regexp.MustCompile(workloadNotFoundPattern)

	workloadV2NotFoundPattern = "\".*\" not found"
	workloadV2NotFoundRegexp  = regexp.MustCompile(workloadV2NotFoundPattern)
)

type SyncService struct{}

func NewSyncService() *SyncService {
	return &SyncService{}
}

func (*SyncService) ReportSyncIssues(ctx context.Context, clusterId string, report api.ClusterSyncReport) (api.ImplResponse, error) {
	ProcessClusterReport(ctx, clusterId, &report)
	return services.NoContentResponse(), nil
}

func ListUpdates(ctx context.Context, workloadType string, afterUpdate int64, clusterId string,
	syncConfig *config.WorkloadsSyncConfig, syncReport *api.ClusterSyncReport,
) (api.ImplResponse, error) {
	ProcessClusterReport(ctx, clusterId, syncReport)

	tenantId, parseErr := services_utils.ParseTenantId(ctx)
	if parseErr != nil {
		return api.Response(int(parseErr.Code), parseErr), nil
	}

	updates, dbErr := dao_crud.DbListUpdates(ctx, tenantId, workloadType, clusterId, afterUpdate, syncConfig.BulkSize)
	if dbErr != nil {
		return api.Response(int(dbErr.Code), dbErr), nil
	}

	syncResponse := api.WorkloadsSyncResponse{
		RequestInfo: api.RequestInfo{
			AfterUpdate: afterUpdate,
		},
	}

	if len(updates) != 0 {
		syncResponse.ResponseInfo.LowestUpdate = updates[0].UpdateCount
		syncResponse.ResponseInfo.HighestUpdate = updates[len(updates)-1].UpdateCount
	}

	for _, update := range updates {
		syncResponse.Workloads = append(syncResponse.Workloads, api.WorkloadsSyncInfo{
			UpdateCount: update.UpdateCount,
			Workload:    WorkloadSyncUpdateFromDbRecord(&update),
		})
	}

	macCount, countErr := dao_crud.GetHighestUpdateCount(ctx)
	if countErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get highest update count: %s", countErr.Message)
	} else {
		syncResponse.SystemInfo.HighestUpdate = macCount
	}

	cleanupErr := dao_crud.CleanupObsoleteWorkloadUpdates(ctx, tenantId, workloadType, clusterId, afterUpdate)
	if cleanupErr != nil {
		return api.Response(int(cleanupErr.Code), cleanupErr), nil
	}

	return api.Response(http.StatusOK, syncResponse), nil
}

func ProcessClusterReport(ctx context.Context, clusterId string, syncReport *api.ClusterSyncReport) {
	if syncReport != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("processing %d sync reports for cluster %s", len(syncReport.Updates), clusterId)

		for _, update := range syncReport.Updates {

			if update.Message != nil {
				if workloadNotFoundRegexp.MatchString(*update.Message) || workloadV2NotFoundRegexp.MatchString(*update.Message) {
					actualPhase, err := services_integration.GetPhaseOfWorkload(ctx, &clusterId, update.Id)
					if err != nil {
						logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get phase of workload %s: %v", update.Id, err)
					} else {
						if actualPhase == api.PHASE_FAILED || actualPhase == api.PHASE_DELETING {
							logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("workload %s was reported missing from cluster %s, removing from DB", update.Id, clusterId)

							_, err := services_integration.DbDeleteWorkload(ctx, update.Id)
							if err != nil {
								logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to remove workload %s from DB: %s", update.Id, err)
							} else {
								logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleted workload %s from cluster", update.Id)
							}
						} else {
							logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("workload %s was reported missing from cluster %s, but has phase %s", update.Id, clusterId, actualPhase)
						}
					}
				} else {
					err := services_integration.UpdateWorkloadFailure(ctx, update.Id, *update.Message, update.UpdateAttemptTime)
					if err != nil {
						logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to notify failure of update %d for workload %s (%s): %s",
							update.UpdateCount, update.Id, *update.Message, err.Message)
					} else {
						logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("update status of workload %s from cluster %s: %s", update.Id, clusterId, *update.Message)
					}
				}
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("skip report for workload %s from cluster %s: no message", update.Id, clusterId)
			}
		}
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("missing sync report of cluster %s", clusterId)
	}
}
