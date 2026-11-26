/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	config "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
)

type WorkflowsService struct {
	WorkloadsDao         *dao.WorkloadsDao
	TenantSettingsClient clients.TenantSettingsClient
}

func NewWorkflowsService(tenantSettingClient clients.TenantSettingsClient) api.WorkflowsAPIServicer {
	return &WorkflowsService{
		WorkloadsDao:         dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()),
		TenantSettingsClient: tenantSettingClient,
	}
}

func (ws *WorkflowsService) CreateWorkflow(ctx context.Context, request api.WorkflowCreationRequest) (api.ImplResponse, error) {
	spec, jsonErr := gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	workflowSpec, specErr := loadWorkflowSpec(spec)
	if specErr != nil {
		return api.Response(int(specErr.Code), specErr), nil
	}

	createWorkloadParams := createWorkloadParams{
		WorkloadType:         Workflow,
		UseGivenNameAsPrefix: request.UseGivenNameAsPrefix,

		Spec: spec,

		TemplateId: nil, // template is not yet supported for workflow submission

		AssetIds: nil, // asset ids are not yet supported for workflow submission

		Response: workloadApi.SupersetWorkload{
			RequestedName: request.Name,
			ClusterId:     request.ClusterId,
			ProjectId:     request.ProjectId,
		},
	}

	if createErr := createWorkflow(ctx, &createWorkloadParams, workflowSpec); createErr != nil {
		return api.Response(int(createErr.Code), createErr), nil
	}

	recordWorkloadCreate(ctx, &createWorkloadParams)
	return api.Response(http.StatusAccepted, newWorkflowResponse(&createWorkloadParams.Response, workflowSpec)), nil
}

func (ws *WorkflowsService) GetWorkflow(context.Context, string) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNotImplemented}, nil
}

func (ws *WorkflowsService) DeleteWorkflow(context.Context, string) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNotImplemented}, nil
}
