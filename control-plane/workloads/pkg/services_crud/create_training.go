/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func CreateTraining(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, request api.TrainingCreationRequest) (api.ImplResponse, error) {
	spec, jsonErr := gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	createWorkloadParams := createWorkloadParams{
		WorkloadType:         Training,
		UseGivenNameAsPrefix: request.UseGivenNameAsPrefix,

		Spec: spec,

		TemplateId: request.TemplateId,

		AssetIds: request.Assets,

		Response: workloadApi.SupersetWorkload{
			RequestedName: request.Name,
			ClusterId:     request.ClusterId,
			ProjectId:     request.ProjectId,
		},
	}

	createErr := createWorkload(ctx, tenantSettingsClient, &createWorkloadParams, Training)
	if createErr != nil {
		return api.Response(int(createErr.Code), createErr), nil
	}

	recordWorkloadCreate(ctx, &createWorkloadParams)
	createWorkloadParams.Response.Spec = gpucompat.LegacyCompatibleWorkloadSpec(createWorkloadParams.Response.Spec)
	if createWorkloadParams.Response.MasterSpec.Get() != nil {
		createWorkloadParams.Response.MasterSpec.Set(gpucompat.LegacyCompatibleMasterSpec(createWorkloadParams.Response.MasterSpec.Get()))
	}
	return api.Response(http.StatusAccepted, createWorkloadParams.Response), nil
}
