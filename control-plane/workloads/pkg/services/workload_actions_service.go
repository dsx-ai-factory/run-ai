/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/actionhistory"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"
)

type WorkloadActionsService struct{}

func NewWorkloadActionsService() api.WorkloadActionsAPIServicer {
	return &WorkloadActionsService{}
}

func (s *WorkloadActionsService) InternalCreateWorkloadActions(ctx context.Context, request api.CreateWorkloadActionsRequest) (api.ImplResponse, error) {
	records := make([]models.WorkloadActionHistory, 0, len(request.Actions))
	for _, input := range request.Actions {
		recordID, err := uuid.Parse(input.Id)
		if err != nil {
			return NewErrorResponse(api.NewValidationError("action Id must be a valid uuid")), nil
		}
		workloadID, err := uuid.Parse(input.WorkloadId)
		if err != nil {
			return NewErrorResponse(api.NewValidationError("workload Id must be a valid uuid")), nil
		}
		clusterID, err := uuid.Parse(input.ClusterId)
		if err != nil {
			return NewErrorResponse(api.NewValidationError("cluster Id must be a valid uuid")), nil
		}
		record, apiErr := actionhistory.NewWorkloadActionRecord(ctx, recordID, input.Action, actionhistory.WorkloadToLog{
			ID: workloadID, ClusterID: clusterID, Name: input.WorkloadName,
		}, input.AcceptedAt)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}
		records = append(records, *record)
	}
	if apiErr := dao.DbInsertRegisteredWorkloadActions(ctx, records); apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *WorkloadActionsService) GetWorkloadActionsHistory(ctx context.Context, workloadId string, offset, limit int32) (api.ImplResponse, error) {
	id, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewValidationError("workload Id must be a valid uuid")), nil
	}
	if offset < 0 || limit < 0 {
		return NewErrorResponse(api.NewValidationError("offset and limit must not be negative")), nil
	}
	limit = setDefaultLimit(limit)
	records, hasNext, dbErr := dao.DbGetWorkloadActions(ctx, id, offset, limit)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}
	response := buildWorkloadActionsResponse(records, offset, hasNext)
	return api.Response(http.StatusOK, response), nil
}

func buildWorkloadActionsResponse(records []models.WorkloadActionHistory, offset int32, hasNext bool) api.GetWorkloadActionsResponse {
	response := api.GetWorkloadActionsResponse{Actions: []api.WorkloadActionRecord{}}
	for _, record := range records {
		response.Actions = append(response.Actions, api.WorkloadActionRecord{
			Id:           record.ID.String(),
			WorkloadId:   record.WorkloadID.String(),
			WorkloadName: record.WorkloadName,
			Subject:      record.Subject,
			SubjectType:  record.SubjectType,
			Action:       api.WorkloadActionType(record.Action),
			AcceptedAt:   record.AcceptedAt,
		})
	}
	if hasNext {
		response.Next = offset + int32(len(records))
	}
	return response
}
