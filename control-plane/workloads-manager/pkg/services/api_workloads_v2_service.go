/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type WorkloadsV2Service struct {
}

func NewWorkloadsV2Service() api.WorkloadsV2APIServicer {
	return &WorkloadsV2Service{}
}

var impl = NewWorkloadsV2Impl()

func (w WorkloadsV2Service) CreateWorkloadV2(ctx context.Context, request api.WorkloadV2CreateRequest, dryRun bool) (api.ImplResponse, error) {
	payload, errors := impl.CreateWorkloadV2WithDryRun(ctx, request, CreateWorkloadV2Options{
		DryRun:                  dryRun,
		CheckYamlSubmissionFlag: true,
		TranslatedFromV1:        false,
		RecordAction:            recordAction,
	})
	if len(errors) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create workload with %d errors", len(errors))
		return api.Response(int(errors[0].Code), errors), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

func (w WorkloadsV2Service) CreateWorkloadInternal(ctx context.Context, request api.WorkloadV2CreateRequest, dryRun bool) (api.ImplResponse, error) {
	payload, errors := impl.CreateWorkloadV2WithDryRun(ctx, request, CreateWorkloadV2Options{
		DryRun:                  dryRun,
		CheckYamlSubmissionFlag: false,
		TranslatedFromV1:        true,
	})
	if len(errors) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create internal workload with %d errors", len(errors))
		return api.Response(int(errors[0].Code), errors), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

func (w WorkloadsV2Service) GetWorkloadV2ById(ctx context.Context, id string) (api.ImplResponse, error) {
	payload, err := impl.GetWorkloadV2ById(ctx, id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to obtain workload %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, payload), nil
}

func (w WorkloadsV2Service) DeleteWorkloadV2(ctx context.Context, id string) (api.ImplResponse, error) {
	payload, err := impl.DeleteWorkloadV2(ctx, id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workload %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

// SetWorkloadRunState - Set the run state of a workload. [Experimental]
func (w WorkloadsV2Service) SetWorkloadRunState(ctx context.Context, workloadV2Id string, workloadRunStateRequest api.WorkloadRunStateRequest) (api.ImplResponse, error) {
	payload, err := impl.SetWorkloadRunState(ctx, workloadV2Id, workloadRunStateRequest)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to set run state for workload %s: %v", workloadV2Id, err.Message)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

func (w WorkloadsV2Service) GetWorkloadV2Updates(ctx context.Context, clusterId string, syncAfter int64, limit int32,
	clusterReport api.ClusterSyncReport,
) (api.ImplResponse, error) {
	payload, err := impl.GetWorkloadV2Updates(ctx, clusterId, syncAfter, limit, clusterReport)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workload %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err.Message), nil
	}

	return api.Response(http.StatusOK, payload), nil
}

func (w WorkloadsV2Service) BatchWorkloads(ctx context.Context, batch api.WorkloadsV2Batch) (api.ImplResponse, error) {
	payload, err := impl.ProcessBatch(ctx, batch)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to proces workload batch %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err.Message), nil
	}

	return api.Response(http.StatusOK, payload), nil
}

func (w WorkloadsV2Service) UpdateWorkloadV2(ctx context.Context, id string, request api.WorkloadV2UpdateRequest) (api.ImplResponse, error) {
	payload, err := impl.UpdateWorkloadV2(ctx, id, request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update workload %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

func (w WorkloadsV2Service) UpdateWorkloadInternal(ctx context.Context, id string, request api.WorkloadV2UpdateRequest) (api.ImplResponse, error) {
	payload, err := impl.UpdateWorkloadInternal(ctx, id, request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update internal workload %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}
