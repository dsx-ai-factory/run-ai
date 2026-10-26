/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/actionhistory"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

// Log history write failures without failing the accepted workload action.
func recordWorkloadAction(ctx context.Context, workload actionhistory.WorkloadToLog, action api.WorkloadActionType, acceptedAt time.Time) {
	record, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, action, workload, acceptedAt)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to record accepted workload %s action for %s: %v", action, workload.ID, apiErr)
		return
	}
	if apiErr := dao.DbInsertWorkloadActions(ctx, []models.WorkloadActionHistory{*record}); apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to record accepted workload %s action for %s: %v", action, workload.ID, apiErr)
	}
}

func recordWorkloadStateAction(ctx context.Context, workload *models.WorkloadView, desiredPhase api.WorkloadDesiredPhase, acceptedAt time.Time) {
	var action api.WorkloadActionType
	switch desiredPhase {
	case api.WORKLOADDESIREDPHASE_DELETED:
		action = api.WORKLOADACTIONTYPE_DELETE
	case api.WORKLOADDESIREDPHASE_STOPPED:
		if !supportsV1RunStateHistory(workload.Type) {
			return
		}
		action = api.WORKLOADACTIONTYPE_SUSPEND
	case api.WORKLOADDESIREDPHASE_RUNNING:
		if !supportsV1RunStateHistory(workload.Type) {
			return
		}
		action = api.WORKLOADACTIONTYPE_RESUME
	default:
		return
	}
	recordWorkloadAction(ctx, actionhistory.WorkloadToLog{ID: workload.ID, ClusterID: workload.ClusterID, ProjectID: workload.ProjectID, DepartmentID: workload.DepartmentID, Name: workload.Name}, action, acceptedAt)
}

func supportsV1RunStateHistory(workloadType string) bool {
	return workloadType == Workspace || workloadType == Training || workloadType == Distributed
}

func recordWorkloadResponseAction(ctx context.Context, response workloadsApi.SupersetWorkload, action api.WorkloadActionType, acceptedAt time.Time) {
	workloadID, err := uuid.Parse(response.WorkloadId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to record accepted workload %s: invalid workload ID: %v", action, err)
		return
	}
	clusterID, err := uuid.Parse(response.ClusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to record accepted workload %s for %s: invalid cluster ID: %v", action, workloadID, err)
		return
	}
	recordWorkloadAction(ctx, actionhistory.WorkloadToLog{
		ID: workloadID, ClusterID: clusterID,
		ProjectID: response.ProjectId, DepartmentID: convutils.Deref(response.DepartmentId, ""), Name: response.Name,
	}, action, acceptedAt)
}

func recordWorkloadCreate(ctx context.Context, params *createWorkloadParams) {
	recordWorkloadResponseAction(ctx, params.Response, api.WORKLOADACTIONTYPE_CREATE, params.AcceptedAt)
}
