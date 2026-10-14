/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

const (
	LWSNotAvailableMsg = "Distributed inference can't be created because LeaderWorkerSet (LWS) is not installed on the selected cluster. Contact your administrator."
)

func CreateDistributedInference(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, request api.DistributedInferenceCreationRequest) (api.ImplResponse, error) {
	// check if a cluster version supports distributed inference create
	if err := services_utils.ValidateClusterSupportsCapability(ctx, request.ClusterId, checker.MinVersionCreateDistributedInference); err != nil {
		return api.Response(http.StatusBadRequest, err.Message), nil
	}

	// check if LWS is available
	err := workloadutils.CheckLWSAvailable(ctx, request.ClusterId, LWSNotAvailableMsg)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	spec, jsonErr := gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	createWorkloadParams := createWorkloadParams{
		WorkloadType:         DistributedInference,
		UseGivenNameAsPrefix: request.UseGivenNameAsPrefix,

		Spec: spec,

		TemplateId: request.TemplateId,

		Response: workloadApi.SupersetWorkload{
			RequestedName: request.Name,
			ClusterId:     request.ClusterId,
			ProjectId:     request.ProjectId,
		},
	}

	if request.Assets != nil {
		createWorkloadParams.AssetIds = request.Assets.Worker
		createWorkloadParams.LeaderAssetIds = request.Assets.Leader
	}

	createErr := createWorkload(ctx, tenantSettingsClient, &createWorkloadParams, DistributedInference)
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
