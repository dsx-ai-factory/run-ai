/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	corev1 "k8s.io/api/core/v1"
)

func (dinfComposer *DistributedInferenceToLWSComposer) processComputeFields(
	ctx context.Context,
	src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
	target *corev1.PodTemplateSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}
		canonical, err := gpucompat.NormalizeLegacyWorkloadCompute(src)
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, err.Error())
		}
		src = canonical

		container := &target.Spec.Containers[0]
		var sizingInfo composer.GpuSizingInfo
		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &container.Resources.Requests, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dinfComposer.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &container.Resources.Limits, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dinfComposer.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &container.Resources.Requests, composer.StringToK8sResourceList, composer.ResourceMemory),
			dinfComposer.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &container.Resources.Limits, composer.StringToK8sResourceList, composer.ResourceMemory),
			dinfComposer.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuMemoryPortionRequest.Get(), src.GpuMemoryPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), composer.WholeGpusToK8sResources,
				&container.Resources, nil, &sizingInfo),
			dinfComposer.fc.ComposeExtendedResources(itemized.ExtendedResourceNickPrefix, src.ExtendedResources, &container.Resources, composer.ExtendedResourcesToK8sResources, nil),
		)

		if sizingInfo.MemoryLimit != nil || sizingInfo.PortionLimit != nil {
			dinfComposer.WorkloadRequirements.GpuOptimization = true
		}

		runner.Batch(
			// GPU memory/portion limit is written either as the legacy env var (< 3.0
			// clusters) or as the per-container CUDA/KAI annotations (>= 3.0). The
			// two pairs are mutually exclusive at the compose helper level.
			dinfComposer.fc.ComposeGpuComputeMode(composer.EnumToString(src.GpuComputeMode.Get()), &sizingInfo, container.Name, &target.Annotations, composer.StringToMapString),
			dinfComposer.fc.ComposeMemoryLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			dinfComposer.fc.ComposePortionLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuPortionLimitEnvVar),
			dinfComposer.fc.ComposeMemoryLimitAnnotation(&sizingInfo, container.Name, &target.Annotations, composer.StringToMapString),
			dinfComposer.fc.ComposePortionLimitAnnotation(&sizingInfo, container.Name, &target.Annotations, composer.StringToMapString),
			dinfComposer.fc.ComposeMemoryReqAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuMemoryAnnotation),
			dinfComposer.fc.ComposePortionReqAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionAnnotation),
			dinfComposer.fc.ComposeNumDevicesAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionDevicesAnnotation),
		)
		return runner.Error()
	}
}
