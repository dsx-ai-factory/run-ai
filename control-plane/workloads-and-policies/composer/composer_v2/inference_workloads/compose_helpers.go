/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/init_containers"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// inferenceServingURL is the upstream URL the serving proxy forwards to. It is
// the workload's public URL when it is externally exposed, or an in-cluster DNS
// name when ClusterLocalAccessOnly is true.
func inferenceServingURL(config composer.ClusterStorageConfig, projectName, workloadName, namespace string, clusterLocal bool) (string, *workloadsApi.Error) {
	if clusterLocal {
		return fmt.Sprintf("http://%s.%s.svc.cluster.local/", workloadName, namespace), nil
	}
	parsed, err := composer.ComposeAutoExposedURL(
		config.ClusterDomain(),
		projectName,
		workloadName,
		"",
		config.SubdomainSupportEnabled(),
	)
	if err != nil {
		return "", workloadsApi.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to compose inference serving URL: %v", err))
	}
	return parsed.String(), nil
}

func (itc *InferenceToKnativeComposer) processComputeFields(
	ctx context.Context,
	src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
	target *KnativeServiceTarget,
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

		container := &target.Service.Spec.Template.Spec.Containers[0]
		podMeta := &target.Service.Spec.Template.ObjectMeta
		var sizingInfo composer.GpuSizingInfo

		runner := &composer.StepRunner{}
		runner.Batch(
			itc.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &container.Resources.Requests, composer.FloatToK8sResourceList, composer.ResourceCpu),
			itc.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &container.Resources.Limits, composer.FloatToK8sResourceList, composer.ResourceCpu),
			itc.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &container.Resources.Requests, composer.StringToK8sResourceList, composer.ResourceMemory),
			itc.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &container.Resources.Limits, composer.StringToK8sResourceList, composer.ResourceMemory),

			itc.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuMemoryPortionRequest.Get(), src.GpuMemoryPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), composer.WholeGpusToK8sResources,
				&container.Resources, nil, &sizingInfo),
		)

		if sizingInfo.MemoryLimit != nil || sizingInfo.PortionLimit != nil {
			itc.WorkloadRequirements.GpuOptimization = true
		}

		runner.Batch(
			// GPU memory/portion limit is written either as the legacy env var (< 3.0
			// clusters) or as the per-container CUDA/KAI annotations (>= 3.0). The
			// two pairs are mutually exclusive at the compose helper level.
			itc.fc.ComposeGpuComputeMode(composer.EnumToString(src.GpuComputeMode.Get()), &sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			itc.fc.ComposeMemoryLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			itc.fc.ComposePortionLimitEnv(&sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuPortionLimitEnvVar),
			itc.fc.ComposeMemoryLimitAnnotation(&sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			itc.fc.ComposePortionLimitAnnotation(&sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			itc.fc.ComposeMemoryReqAnnotation(&sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuMemoryAnnotation),
			itc.fc.ComposePortionReqAnnotation(&sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuFractionAnnotation),
			itc.fc.ComposeNumDevicesAnnotation(&sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuFractionDevicesAnnotation),
		)

		return runner.Error()
	}
}

func (itc *InferenceToKnativeComposer) processSecurityFields(
	_ context.Context,
	src *workloadsApi.InferenceSpecSpecAllOfSecurity,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		podSpec := &target.Service.Spec.Template.Spec.PodSpec
		container := &target.Service.Spec.Template.Spec.Containers[0]

		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
		}
		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}

		// we do not have privileged and allowPrivilegedEscalation in inference spec we want to force its behavior
		// to be non-privileged
		forcedNonPrivileged := convutils.Ref(false)

		runner := &composer.StepRunner{}
		runner.Batch(
			itc.fc.CopyBoolean(src.HostIpc.Get(), &podSpec.HostIPC, composer.BoolToBool, nil),
			itc.fc.CopyBoolean(src.HostNetwork.Get(), &podSpec.HostNetwork, composer.BoolToBool, nil),
			itc.fc.ComposeDnsPolicy(src.HostNetwork.Get(), &podSpec.DNSPolicy, composer.StringToString, nil),
			itc.fc.CopyBoolean(src.ReadOnlyRootFilesystem.Get(), &container.SecurityContext.ReadOnlyRootFilesystem, composer.BoolToBool, nil),
			itc.fc.ComposeRunAsNonRoot(src.RunAsNonRoot.Get(), forcedNonPrivileged, &container.SecurityContext.RunAsNonRoot, composer.BoolToBool, nil),
			itc.fc.ComposeCapabilities(src.Capabilities, &container.SecurityContext.Capabilities, composer.MapToK8sCapabilities, nil),
			itc.fc.ComposeSeccompProfile(src.SeccompProfileType.Get(), nil, &container.SecurityContext.SeccompProfile, composer.StringToSeccompProfile, nil),
		)

		var runAsUser bool
		runner.Run(
			itc.fc.ComposeRunAsUser(composer.EnumToString(src.UidGidSource.Get()), &runAsUser, composer.BoolToBool, nil),
		)
		if runAsUser {
			runner.Batch(
				itc.fc.CopyInt(src.RunAsUid.Get(), &container.SecurityContext.RunAsUser, composer.IntToInt, nil),
				itc.fc.CopyInt(src.RunAsGid.Get(), &container.SecurityContext.RunAsGroup, composer.IntToInt, nil),
				itc.fc.ComposeSupplementalGroups(src.SupplementalGroups.Get(), &podSpec.SecurityContext.SupplementalGroups, composer.StringToSupplementalGroups, nil),
			)
		}

		if composer.IsEmpty(*podSpec.SecurityContext) {
			podSpec.SecurityContext = nil
		}
		if composer.IsEmpty(*container.SecurityContext) {
			container.SecurityContext = nil
		}

		return runner.Error()
	}
}

func (itc *InferenceToKnativeComposer) processStorageFields(
	_ context.Context,
	src *workloadsApi.InferenceSpecSpecAllOfStorage,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		podSpec := &target.Service.Spec.Template.Spec.PodSpec
		container := &podSpec.Containers[0]
		volumesAndMounts := composer.VolumesAndMounts{Config: &target.Config}
		var mountPropagation workloadsApi.HostPathMountPropagation

		runner := &composer.StepRunner{}
		runner.Batch(
			itc.fc.ComposeHostPath(itemizedPrefix(itemized.HostPathPrefix), src.HostPath, &volumesAndMounts, composer.MapHostPathToK8sVol, nil, &mountPropagation),
			itc.fc.ComposePvc(itemizedPrefix(itemized.PvcPrefix), src.Pvc, &volumesAndMounts, composer_v2.MapPvcToWorkloadRunner, nil),
			itc.fc.ComposeNfs(itemizedPrefix(itemized.NfsPrefix), src.Nfs, &volumesAndMounts, composer.MapNfsToK8sVol, nil),
			itc.fc.ComposeConfigMap(itemizedPrefix(itemized.ConfigMapPrefix), src.ConfigMapVolume, &volumesAndMounts, composer.MapConfigMapToK8sVol, nil),
			itc.fc.ComposeEmptyDir(itemizedPrefix(itemized.EmptyDirPrefix), src.EmptyDirVolume, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			itc.fc.ComposeSecretVolume(itemizedPrefix(itemized.SecretStoragePrefix), src.SecretVolume, &volumesAndMounts, composer.MapSecretToK8sVol, nil),
		)
		runner.Batch(
			itc.fc.ComposeDataVolume(itemizedPrefix(itemized.DataVolumePrefix), src.DataVolume, &volumesAndMounts, composer.MapDataVolumesToK8s, itc.DataVolumeClaimNames),
			itc.fc.ComposeGit(itemizedPrefix(itemized.GitPrefix), src.Git, &target.Service.Spec.Template.Spec.PodSpec, composer.MapGitToWorkloadK8s, nil),
			itc.fc.CopyBoolean(itc.fc.Options.CleanupCreatedPvcsOnCompletion, volumesAndMounts.Config, composer_v2.BoolToPvcsCleanupConfig, nil),
		)

		container.VolumeMounts = append(container.VolumeMounts, volumesAndMounts.VolumeMount...)
		podSpec.Volumes = append(podSpec.Volumes, volumesAndMounts.Volume...)

		return runner.Error()
	}
}

func itemizedPrefix(p string) string { return p }

func (itc *InferenceToKnativeComposer) processNetworkingFields(
	src *workloadsApi.ServingPort,
	auth ServingPortAuthFields,
	engine inferenceEngine,
	exposedURLInfo *composer.ExposedUrlInfo,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if itc.fc.Options.WorkloadV2ServiceSpecLabelsSupported {
			metricsEndpoint := buildMetricsEndpoint(engine, itc.fc.WorkloadName, src)
			appendMetricsEndpoint(&target.Config, metricsEndpoint)
		}

		if src == nil {
			return nil
		}

		container := &target.Service.Spec.Template.Spec.Containers[0]

		runner := &composer.StepRunner{}
		runner.Batch(
			itc.fc.CopyInt(src.Container.Get(), container, int32ToKnativeContainerPorts, nil),
			itc.buildAuthProxySidecar(src, auth, exposedURLInfo, target),
		)
		return runner.Error()
	}
}

// buildAuthProxySidecar returns a batch step that installs an oauth2-proxy
// sidecar in front of the Knative user container when the serving port is not
// public. No-op for public serving ports. Must run after CopyInt has projected
// the user container's port (the CustomizeServingProxy hook reads and clears
// the container's Ports and ReadinessProbe).
func (itc *InferenceToKnativeComposer) buildAuthProxySidecar(
	src *workloadsApi.ServingPort,
	auth ServingPortAuthFields,
	exposedURLInfo *composer.ExposedUrlInfo,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if auth.IsPublic() {
			return nil
		}

		container := &target.Service.Spec.Template.Spec.Containers[0]
		userPort := convutils.Deref(src.Container.Get(), 0)
		config := itc.fc.ClusterStorageConfig
		if config == nil {
			config = composer.DefaultClusterStorageConfig()
		}

		servingURL, urlErr := inferenceServingURL(config,
			itc.fc.ProjectName, itc.fc.WorkloadName, itc.fc.Namespace,
			convutils.BoolIsTrue(src.ClusterLocalAccessOnly.Get()))
		if urlErr != nil {
			return urlErr
		}

		return init_containers.BuildOauth2ProxySidecar(
			&target.Service.Spec.Template.Spec.PodSpec,
			init_containers.ServingProxySpec{
				ContainerPort:                 userPort,
				ServingURL:                    servingURL,
				CookiePrefix:                  fmt.Sprintf("InferenceWorkload-%s-%s-0", itc.fc.Namespace, itc.fc.WorkloadName),
				AuthorizedUsers:               auth.AuthorizedUsers,
				AuthorizedGroups:              auth.AuthorizedGroups,
				UpstreamTimeoutSeconds:        convutils.Deref(target.Service.Spec.Template.Spec.TimeoutSeconds, 0),
				WorkloadName:                  itc.fc.WorkloadName,
				ProjectName:                   itc.fc.ProjectName,
				Config:                        config,
				ExposedURLInfo:                exposedURLInfo,
				OAuth2ProxyTemplatesSupported: itc.fc.Options.OAuth2ProxyTemplatesSupported,
				// Knative admits only one port-declaring spec.container:
				// hand the readiness probe to the serving proxy and clear
				// the user container's ports so it becomes a helper behind
				// the proxy.
				CustomizeServingProxy: func(_ *corev1.PodSpec, servingProxy *corev1.Container) *workloadsApi.Error {
					servingProxy.ReadinessProbe = container.ReadinessProbe
					container.ReadinessProbe = nil
					if servingProxy.ReadinessProbe == nil {
						servingProxy.ReadinessProbe = &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(userPort)},
						}}
					}
					container.Ports = nil
					return nil
				},
			},
		)
	}
}

// processPlacementFields — nodePools and priorityClass are intentionally skipped; see supportedPriorityClass[Inference].
func (itc *InferenceToKnativeComposer) processPlacementFields(
	ctx context.Context,
	src *workloadsApi.InferenceSpecSpec,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		podSpec := &target.Service.Spec.Template.Spec.PodSpec
		podMeta := &target.Service.Spec.Template.ObjectMeta

		if podSpec.Affinity == nil {
			podSpec.Affinity = &corev1.Affinity{}
		}

		runner := &composer.StepRunner{}
		runner.Batch(
			itc.fc.CopyString(src.NodeType.Get(), &podSpec.NodeSelector, composer.StringToMapString, "run.ai/node-type"), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
			itc.fc.ComposePreemptibility(ctx, src.Preemptibility.Get(), &podMeta.Labels, composer.PreemptibilityToK8sLabel, nil),
			itc.fc.ComposeNodeAffinity(src.NodeAffinityRequired.Get(), &podSpec.Affinity.NodeAffinity, composer.NodeAffinityRequiredToK8s, nil),
			itc.fc.ComposePodAffinity(src.PodAffinity.Get(), &podSpec.Affinity.PodAffinity, composer.PodAffinityToK8s, itc.fc.WorkloadName),
		)
		runner.Run(
			itc.fc.ComposeTolerations(itemized.TolerationPrefix, src.Tolerations, &podSpec.Tolerations, composer.MapToK8sTolerations, nil),
		)

		if composer.IsEmpty(*podSpec.Affinity) {
			podSpec.Affinity = nil
		}

		return runner.Error()
	}
}

func (itc *InferenceToKnativeComposer) processServingConfigFields(
	sc *workloadsApi.ServingConfiguration,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if sc == nil {
			return nil
		}

		podMeta := &target.Service.Spec.Template.ObjectMeta
		runner := &composer.StepRunner{}
		runner.Run(itc.fc.CopyInt(sc.InitializationTimeoutSeconds.Get(), podMeta, initTimeoutToProgressDeadlineAnnotation, nil))
		runner.Run(itc.fc.CopyInt(sc.RequestTimeoutSeconds.Get(), &target.Service.Spec.Template.Spec.TimeoutSeconds, composer.IntToInt[int64], nil))
		return runner.Error()
	}
}
