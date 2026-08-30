package dinf_to_lws

import (
	"context"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
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

		container := &target.Spec.Containers[0]
		var sizingInfo composer.GpuSizingInfo
		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &container.Resources.Requests, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dinfComposer.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &container.Resources.Limits, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dinfComposer.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &container.Resources.Requests, composer.StringToK8sResourceList, composer.ResourceMemory),
			dinfComposer.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &container.Resources.Limits, composer.StringToK8sResourceList, composer.ResourceMemory),
			dinfComposer.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuPortionRequest.Get(), src.GpuPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), composer.WholeGpusToK8sResources,
				&container.Resources, nil, &sizingInfo),
			dinfComposer.fc.ComposeExtendedResources(itemized.ExtendedResourceNickPrefix, src.ExtendedResources, &container.Resources, composer.ExtendedResourcesToK8sResources, nil),
		)

		if sizingInfo.MemoryLimit != nil || sizingInfo.PortionLimit != nil {
			dinfComposer.WorkloadRequirements.GpuOptimization = true
		}

		runner.Batch(
			dinfComposer.fc.ComposeMemoryLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			dinfComposer.fc.ComposePortionLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuPortionLimitEnvVar),
			dinfComposer.fc.ComposeMemoryReqAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuMemoryAnnotation),
			dinfComposer.fc.ComposePortionReqAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionAnnotation),
			dinfComposer.fc.ComposeNumDevicesAnnotation(&sizingInfo, &target.Annotations, composer.StringToMapString, composer.GpuFractionDevicesAnnotation),
		)
		return runner.Error()
	}
}
