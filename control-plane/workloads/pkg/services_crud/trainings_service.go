/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
)

type TrainingsService struct {
	Conf                 *config.ServerConfig
	SyncConf             *config.WorkloadsSyncConfig
	tenantSettingsClient clients.TenantSettingsClient
}

func NewTrainingsService(tenantSettingsClient clients.TenantSettingsClient) api.TrainingsAPIServicer {
	return &TrainingsService{
		Conf:                 runai_config.Get[*config.ServerConfig](),
		SyncConf:             runai_config.Get[*config.WorkloadsSyncConfig](),
		tenantSettingsClient: tenantSettingsClient,
	}
}

func (s *TrainingsService) CreateTraining(ctx context.Context, request api.TrainingCreationRequest) (api.ImplResponse, error) {
	return CreateTraining(ctx, s.tenantSettingsClient, request)
}

func (s *TrainingsService) DeleteTraining(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	res, err := ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_DELETED)
	if res.Code >= 200 && res.Code < 300 {
		res.Body = nil // When deleting a single workload, we do not return a response body for 2xx status codes.
	}
	return res, err
}

func (s *TrainingsService) GetTraining(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return GetWorkloadById(ctx, Training, workloadId)
}

func (s *TrainingsService) ResumeTraining(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_RUNNING)
}

func (s *TrainingsService) SuspendTraining(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_STOPPED)
}

func (s *TrainingsService) ListTrainingsUpdates(ctx context.Context, afterUpdate int64, clusterId string) (api.ImplResponse, error) {
	return ListUpdates(ctx, Training, afterUpdate, clusterId, s.SyncConf, nil)
}

func (s *TrainingsService) RequestTrainingsUpdates(ctx context.Context, afterUpdate int64, clusterId string, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return ListUpdates(ctx, Training, afterUpdate, clusterId, s.SyncConf, &syncReport)
}
