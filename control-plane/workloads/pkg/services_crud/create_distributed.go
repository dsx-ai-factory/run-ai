/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"maps"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

func CreateDistributed(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, request api.DistributedCreationRequest) (api.ImplResponse, error) {
	createWorkloadParams := createWorkloadParams{
		WorkloadType:         Distributed,
		UseGivenNameAsPrefix: request.UseGivenNameAsPrefix,

		TemplateId: request.TemplateId,

		AssetIds:       request.Assets,
		MasterAssetIds: request.MasterAssets,

		Response: workloadApi.SupersetWorkload{
			RequestedName: request.Name,
			ClusterId:     request.ClusterId,
			ProjectId:     request.ProjectId,
		},
	}

	//
	//   perform basic validations on the spec
	//
	if request.Spec.DistributedFramework == nil {
		err := workloadApi.NewSubmissionErrorV2(http.StatusBadRequest, distributedFrameworkField+" must be provided for distributed workload")
		return api.Response(int(err.Code), err), nil
	}

	if err := services_utils.CheckDistributedFrameworkAvailable(ctx, request.ClusterId, string(*request.Spec.DistributedFramework)); err != nil {
		return api.Response(int(err.Code), err), nil
	}

	if convutils.BoolIsTrue(request.MasterSpecSameAsWorker) {
		if request.MasterSpec != nil {
			err := workloadApi.NewSubmissionErrorV2(http.StatusBadRequest, "masterSpec should not be provided when masterSpecSameAsMaster is requested")
			return api.Response(int(err.Code), err), nil
		}
		createWorkloadParams.Response.MasterSpecSameAsWorker.Set(convutils.Ref(true))
	} else {
		distributedFramework := workloadApi.DistributedFramework(*request.Spec.DistributedFramework)
		if request.MasterSpec == nil && !frameworksSupportingNoMaster.Has(distributedFramework) {
			err := workloadApi.NewSubmissionErrorV2(http.StatusBadRequest, "either provide a masterSpec or enable masterSpecSameAsMaster")
			return api.Response(int(err.Code), err), nil
		}
	}

	var jsonErr error

	//
	//   we store the payload in the same sections as they appear in the defaults, i.e.
	//   master and worker. this will allow us to merge with the policy easily
	//
	createWorkloadParams.Spec = map[string]interface{}{}

	createWorkloadParams.Spec[workerSection], jsonErr = gojsonq.LoadJson(request.Spec)
	if jsonErr != nil {
		return api.Response(http.StatusBadRequest, jsonErr), nil
	}

	if request.MasterSpec != nil {
		createWorkloadParams.Spec[masterSection], jsonErr = gojsonq.LoadJson(request.MasterSpec)
		if jsonErr != nil {
			return api.Response(http.StatusBadRequest, jsonErr), nil
		}
	} else if convutils.BoolIsTrue(request.MasterSpecSameAsWorker) {
		createWorkloadParams.Spec[masterSection] = maps.Clone(convutils.MapMe(createWorkloadParams.Spec[workerSection]))
	}

	createErr := createWorkload(ctx, tenantSettingsClient, &createWorkloadParams, Distributed)
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
