/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	config "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
)

type WorkflowsService struct {
	WorkloadsDao *dao.WorkloadsDao
}

func NewWorkflowsService() api.WorkflowsAPIServicer {
	return &WorkflowsService{
		WorkloadsDao: dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()),
	}
}

func (ws *WorkflowsService) CreateWorkflow(context.Context, api.WorkflowCreationRequest) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNotImplemented}, nil
}

func (ws *WorkflowsService) GetWorkflow(context.Context, string) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNotImplemented}, nil
}

func (ws *WorkflowsService) DeleteWorkflow(context.Context, string) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNotImplemented}, nil
}
