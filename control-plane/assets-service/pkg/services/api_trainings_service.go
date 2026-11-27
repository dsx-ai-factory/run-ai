/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

// TrainingsApiService is a service that implements the logic for the TrainingsAPIServicer
type TrainingsApiService struct{}

func NewTrainingsApiService() *TrainingsApiService {
	return &TrainingsApiService{}
}

// ListTrainings is a placeholder, currently implemented by the "old" backend, so it should never arrive here
func (s *TrainingsApiService) ListTrainings(context.Context /*clusterIdParam WAIT_FOR_BEING_DELETED old backend*/, string /*nodeIdParam*/, string /*filterParam*/, string,
	/*sortByParam*/ string /*sortDirectionParam*/, string /*fromParam*/, int32 /*limitParam*/, int32,
) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *TrainingsApiService) UpdateTrainingStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *TrainingsApiService) SyncTrainingStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *TrainingsApiService) DeleteTrainingById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}

func (s *TrainingsApiService) UpdateDistributedStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *TrainingsApiService) SyncDistributedStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *TrainingsApiService) DeleteDistributedById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}
