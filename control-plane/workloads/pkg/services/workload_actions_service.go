/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"
)

func (s *EventsService) GetWorkloadActionsHistory(ctx context.Context, workloadId string, offset, limit int32) (api.ImplResponse, error) {
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
