/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"

	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func CreateInference(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, request api.InferenceCreationRequest) (api.ImplResponse, error) {
	err := services_utils.CheckKnativeAvailable(ctx, request.ClusterId, services_utils.KnativeFeatures{
		ScaleToZero: isScaleToZeroWorkload(request.Spec.Autoscaling),
		NodeType:    isNodeTypeConfigured(request.Spec.NodeType), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	spec, jsonErr := gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	createWorkloadParams := createWorkloadParams{
		WorkloadType:         Inference,
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

	createErr := createWorkload(ctx, tenantSettingsClient, &createWorkloadParams, Inference)
	if createErr != nil {
		createErr = redactSensitiveValuesFromResponse(createErr, sensitiveEnvironmentVariableValuesFromMap(spec))
		return api.Response(int(createErr.Code), createErr), nil
	}

	sanitizeWorkloadResponse(ctx, &createWorkloadParams.Response)

	recordWorkloadCreate(ctx, &createWorkloadParams)
	createWorkloadParams.Response.Spec = gpucompat.LegacyCompatibleWorkloadSpec(createWorkloadParams.Response.Spec)
	if createWorkloadParams.Response.MasterSpec.Get() != nil {
		createWorkloadParams.Response.MasterSpec.Set(gpucompat.LegacyCompatibleMasterSpec(createWorkloadParams.Response.MasterSpec.Get()))
	}
	return api.Response(http.StatusAccepted, createWorkloadParams.Response), nil
}

func isScaleToZeroWorkload(autoscaling *api.AutoScaling) bool {
	if autoscaling == nil {
		return false
	}

	return autoscaling.MinReplicas != nil && *autoscaling.MinReplicas == 0
}

func isNodeTypeConfigured(nodeType *string) bool {
	if nodeType == nil {
		return false
	}

	return *nodeType != ""
}
