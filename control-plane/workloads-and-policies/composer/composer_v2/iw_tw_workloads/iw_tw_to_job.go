/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	trackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	JobVersion = "batch/v1"
	JobType    = "Job"
)

var validJobRestartPolicies = datastructs.NewSet(
	string(workloadsApi.RESTARTPOLICY_NEVER),
	string(workloadsApi.RESTARTPOLICY_ON_FAILURE),
)

func getLargeShmRequest(compute *workloadsApi.WorkspaceSpecSpecAllOfCompute) *bool {
	if compute == nil {
		return nil
	}
	return compute.LargeShmRequest.Get()
}

func isJupyterWorkload(src *workloadsApi.SupersetSpec) bool {
	for _, exposedUrl := range src.ExposedUrls {
		if convutils.Deref(exposedUrl.ToolType.Get(), "") == composer.JupyterToolType {
			return true
		}
	}
	for _, port := range src.Ports {
		if convutils.Deref(port.ToolType.Get(), "") == composer.JupyterToolType {
			return true
		}
	}
	return false
}

type BatchJobTarget struct {
	Job    v1.Job
	Config openapi.WorkloadV2Configuration
}

func NewMinimalBatchJobTarget() *BatchJobTarget {
	var result BatchJobTarget

	podSpec := &result.Job.Spec.Template.Spec
	podSpec.Containers = append(podSpec.Containers, corev1.Container{})

	return &result
}

type InterTrainToJobComposer struct {
	fc          *composer.FieldsComposer
	ReplicaType *workloadsApi.ReplicaType

	ActualNodePools []clusterApi.Nodepool

	AssetsMapping        []workloadsApi.AssetsMapping
	WorkloadRequirements composer.WorkloadRequirements
	DataVolumeClaimNames map[string]string // maps data volume ID to PVC claim name
}

func NewInterTrainJobComposer(
	workloadType string,
	workloadName string,
	projectName string,
	namespace string,
	requester string,
	assetsMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
	actualNodePools []clusterApi.Nodepool,
) *InterTrainToJobComposer {
	result := InterTrainToJobComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      workloadType,
			WorkloadName:      workloadName,
			Namespace:         namespace,
			SubmittedBy:       requester,
			ProjectName:       projectName,
			ResourcePrefix:    composer.NetworkingResourcePrefixJob,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		AssetsMapping:   assetsMapping,
		ActualNodePools: actualNodePools,
	}

	return &result
}

func (itc *InterTrainToJobComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *InterTrainToJobComposer {
	itc.fc.StorageClassConfiguration = config
	return itc
}

func (itc *InterTrainToJobComposer) WithClusterStorageConfig(config composer.ClusterStorageConfig) *InterTrainToJobComposer {
	itc.fc.ClusterStorageConfig = config
	return itc
}

func (itc *InterTrainToJobComposer) WithDataVolumeClaimNames(claimNames map[string]string) *InterTrainToJobComposer {
	itc.DataVolumeClaimNames = claimNames
	return itc
}

// Compose function takes care to compose the common fields to all the workload types.
func (itc *InterTrainToJobComposer) Compose(ctx context.Context, src *workloadsApi.SupersetSpec,
	target *BatchJobTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var jupyterContainerPort int32
		var exposedUrlInfo *composer.ExposedUrlInfo
		var obsoletePremptibility bool

		if itc.fc.WorkloadType != string(workloadsApi.WORKLOADTYPE_TRAINING) && itc.fc.WorkloadType != string(workloadsApi.WORKLOADTYPE_WORKSPACE) {
			return workloadsApi.NewError(http.StatusNotImplemented, "this composer supports only workspace and training")
		}

		jobMeta := &target.Job.ObjectMeta
		jobSpec := &target.Job.Spec
		podMeta := &jobSpec.Template.ObjectMeta
		podSpec := &jobSpec.Template.Spec
		container := &podSpec.Containers[0]

		envVarsUsingCreds := make([]composer.EnvVarInfo, 0)

		runner := &composer.StepRunner{}

		runner.Batch(
			// mark the workload as managed by runai control plane
			itc.fc.ComposeFullySaasInd(&jobMeta.Labels, composer.BoolToMapString, composer.ControlPlaneManagedLabel),

			// mark the workload as converted-from-native. some plugins may need to know this
			itc.fc.CopyBoolean(convutils.Ref(true), &jobMeta.Labels, composer.BoolToMapString, composer.ConvertedFromNativeLabel),
			itc.fc.CopyString(convutils.Ref(itc.fc.SubmittedBy), &jobMeta.Annotations, composer.StringToMapString, composer.SubmittedByAnnotation),

			// assign the workload to the runai scheduler
			itc.fc.CopyString(convutils.Ref(composer.RunaiScheduler), &podSpec.SchedulerName, composer.StringToString, nil),

			itc.fc.CopyString(convutils.Ref(JobType), &target.Job.Kind, composer.StringToString, nil),
			itc.fc.CopyString(convutils.Ref(JobVersion), &target.Job.APIVersion, composer.StringToString, nil),

			itc.fc.CopyString(&itc.fc.Namespace, &jobMeta.Namespace, composer.StringToString, nil),
			itc.fc.CopyString(&itc.fc.WorkloadName, &jobMeta.Name, composer.StringToString, nil),
			itc.fc.CopyString(&itc.fc.WorkloadName, &container.Name, composer.StringToString, nil),

			itc.fc.ComposeNameAndTypeLabels(itc.fc.WorkloadName, itc.fc.WorkloadType, JobType, itc.fc.ProjectName, &jobMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeNameAndTypeLabels(itc.fc.WorkloadName, itc.fc.WorkloadType, JobType, itc.fc.ProjectName, &podMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeLabels(src.Labels, &jobMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeAnnotations(src.Annotations, &jobMeta.Annotations, composer.MapToMapString, nil),
			itc.fc.ComposeAnnotations(src.Annotations, &podMeta.Annotations, composer.MapToMapString, nil),
			// preserve the native pod "user" annotation contract: the native RunaiJob builder
			// stamps annotations["user"] on the pod template from spec.username. the redirected
			// batch/v1 job bypasses that builder, so set it here from the same requester identity.
			itc.fc.CopyString(convutils.Ref(itc.fc.SubmittedBy), &podMeta.Annotations, composer.StringToMapString, composer.UserAnnotation),

			itc.fc.ComposeImage(src.Image.Get(), &container.Image, composer.StringToString, nil),
			itc.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &container.ImagePullPolicy, composer.StringToString, nil),
			itc.fc.CopyString(src.Command.Get(), &container.Command, composer.UnixStrToStringArray, nil),
			itc.fc.CopyString(src.Args.Get(), &container.Args, composer.UnixStrToStringArray, nil),
			itc.fc.ComposeHomeDirEnvVar(src.CreateHomeDir.Get(), &container.Env, composer.EnvVarsToK8sEnv, nil),
			itc.fc.CopyBoolean(src.Tty.Get(), &container.TTY, composer.BoolToBool, nil),
			itc.fc.CopyBoolean(src.Stdin.Get(), &container.Stdin, composer.BoolToBool, nil),

			itc.fc.ComposeExposedUrls(replicaBasedItemPrefix(itemized.ExposedUrlPrefix, nil),
				src.ExposedUrls, &target.Config, composer_v2.UrlsToWorkloadRunner, nil, &jupyterContainerPort, &exposedUrlInfo,
				false /*forceSuffix*/),
			itc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer_v2.UrlsToK8sInitContainers, nil),
			itc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer.UrlsToK8sContainers, nil),

			itc.fc.ComposePorts(replicaBasedItemPrefix(itemized.PortPrefix, nil),
				src.Ports, &target.Config, composer_v2.PortsToWorkloadRunner, nil),
		)

		var uidGid composer.UidGid
		runner.Run(itc.processSecurityFields(ctx, src.Security.Get(), isJupyterWorkload(src), target, &uidGid))

		rc := composer_v2.ExpandTemplateVarsInCommandAndArgs(container, itc.fc.ProjectName, itc.fc.WorkloadName)
		if rc != nil {
			return rc
		}

		runaiEnvVars := composer.RunaiEnvVars(itc.fc.ProjectName, itc.fc.WorkloadName, int(jupyterContainerPort), &uidGid)

		runner.Batch(
			itc.fc.ComposeEnvironmentVariables(itc.AssetsMapping, runaiEnvVars, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),
			itc.fc.ComposeEnvironmentVariables(itc.AssetsMapping, src.EnvironmentVariables, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),

			itc.fc.CopyString(src.WorkingDir.Get(), &container.WorkingDir, composer.StringToString, nil),
		)

		runner.Batch(
			itc.fc.CopyString(src.NodeType.Get(), &podSpec.NodeSelector, composer.StringToMapString, "run.ai/node-type"), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
			itc.fc.CopyString(src.Category.Get(), &jobMeta.Labels, composer.StringToMapString, composer.WorkloadCategoryLabel),

			itc.fc.ComposePriorityClass(ctx, src.PriorityClass.Get(), &podSpec.PriorityClassName, composer.StringToString,
				&obsoletePremptibility, composer.BoolToBool, nil),
			itc.fc.ComposePreemptibility(ctx, src.Preemptibility.Get(), &podMeta.Labels, composer.PreemptibilityToK8sLabel, nil),

			itc.composeBackoffLimit(src.BackoffLimit.Get(), &jobSpec.BackoffLimit, composer.IntToInt, nil),
			itc.fc.CopyInt(src.Completions.Get(), &jobSpec.Completions, composer.NonZeroIntToInt, nil),
			itc.fc.ComposeParallelism(src.Parallelism.Get(), src.Completions.Get(), &jobSpec.Parallelism, composer.NonZeroIntToInt, nil),
			itc.composeRestartPolicy(src.RestartPolicy.Get(), &podSpec.RestartPolicy),
			itc.fc.ComposeTerminationGracePeriod(ctx, src.TerminationGracePeriodSeconds.Get(), &podSpec.TerminationGracePeriodSeconds, composer.NonZeroIntToInt, nil),
			itc.fc.CopyInt(src.AutoDeletionTimeAfterCompletionSeconds.Get(), &jobMeta.Annotations, composer.NonZeroIntToMapString, composer.AutoDeleteTimeoutAnnotation),
		)

		ipsTarget := composer_v2.PersonalCredsTarget{
			Config:  &target.Config,
			PodSpec: podSpec,
		}
		ipsCookie := composer_v2.PersonalCredsCookie{
			WorkloadName: itc.fc.WorkloadName,
		}
		runner.Run(itc.fc.ComposeImagePullSecrets(itc.AssetsMapping, src.ImagePullSecrets, &ipsTarget, composer_v2.ArrayToImagePullSecretsSetter, &ipsCookie))

		runner.Batch(
			itc.processComputeFields(ctx, src.Compute.Get(), target),
			itc.processStorageFields(ctx, src.Storage.Get(), uidGid, src.CreateHomeDir.Get(), getLargeShmRequest(src.Compute.Get()), target),
			itc.processAffinityFields(ctx, src, target),
		)

		targetForEnvVarsUsingCreds := composer_v2.EnvVarUsingCredTarget{
			Config:  &target.Config,
			EnvVars: &container.Env,
		}
		cookieForEnvVarsUsingCreds := composer_v2.PersonalCredsCookie{
			WorkloadName: itc.fc.WorkloadName,
		}

		runner.Run(itc.fc.ComposeEnvVarsUsingCreds(envVarsUsingCreds, &targetForEnvVarsUsingCreds, composer_v2.EnvVarsUsingCredSetter, &cookieForEnvVarsUsingCreds))

		runner.Run(itc.fc.ComposeTolerations(replicaBasedItemPrefix(itemized.TolerationPrefix, nil), src.Tolerations, &podSpec.Tolerations, composer.MapToK8sTolerations, nil))

		if convutils.HasValue(src.Probes) {
			runner.Run(itc.fc.ComposeReadinessProbe(src.Probes.Get().Readiness.Get(), &container.ReadinessProbe, composer.ProbeToK8s, nil))
		}

		return runner.Error()
	}
}

func (itc *InterTrainToJobComposer) processComputeFields(ctx context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
	target *BatchJobTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		canonical, aliasErr := gpucompat.NormalizeLegacyWorkloadCompute(src)
		if aliasErr != nil {
			return workloadsApi.NewError(http.StatusBadRequest, aliasErr.Error())
		}
		src = canonical

		podMeta := &target.Job.Spec.Template.ObjectMeta
		podSpec := &target.Job.Spec.Template.Spec
		container := &podSpec.Containers[0]

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

			itc.fc.ComposeExtendedResources(replicaBasedItemPrefix(itemized.ExtendedResourceNickPrefix, nil), src.ExtendedResources,
				&container.Resources, composer.ExtendedResourcesToK8sResources, nil),
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

func (itc *InterTrainToJobComposer) processSecurityFields(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfSecurity,
	isJupyter bool, target *BatchJobTarget, uidGid *composer.UidGid,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		podSpec := &target.Job.Spec.Template.Spec
		container := &podSpec.Containers[0]

		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
		}

		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}

		runner := &composer.StepRunner{}

		runner.Batch(
			itc.fc.CopyBoolean(src.HostIpc.Get(), &podSpec.HostIPC, composer.BoolToBool, nil),
			itc.fc.CopyBoolean(src.HostNetwork.Get(), &podSpec.HostNetwork, composer.BoolToBool, nil),
			itc.fc.ComposeDnsPolicy(src.HostNetwork.Get(), &podSpec.DNSPolicy, composer.StringToString, nil),
			itc.fc.CopyBoolean(src.Privileged.Get(), &container.SecurityContext.Privileged, composer.BoolToBool, nil),
			itc.fc.CopyBoolean(src.ReadOnlyRootFilesystem.Get(), &container.SecurityContext.ReadOnlyRootFilesystem, composer.BoolToBool, nil),
			itc.fc.ComposeAllowPrivilegeEscalation(src.AllowPrivilegeEscalation.Get(), src.Privileged.Get(), &container.SecurityContext.AllowPrivilegeEscalation, composer.BoolToBool, nil),
			itc.fc.ComposeRunAsNonRoot(src.RunAsNonRoot.Get(), src.Privileged.Get(), &container.SecurityContext.RunAsNonRoot, composer.BoolToBool, nil),
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

			if !isJupyter || !workloadconfig.JupyterUserIdentity() {
				runner.Run(itc.fc.CopyInt(src.RunAsGid.Get(), &podSpec.SecurityContext.FSGroup, composer.IntToInt, nil))
			}

			uidGid.Uid = src.RunAsUid.Get()
			uidGid.Gid = src.RunAsGid.Get()
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

func (itc *InterTrainToJobComposer) processStorageFields(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfStorage,
	uidGid composer.UidGid, createHomeDir *bool, largeShmRequest *bool, target *BatchJobTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil && convutils.BoolIsFalse(createHomeDir) && convutils.BoolIsFalse(largeShmRequest) {
			return nil
		}

		var mountPropagation workloadsApi.HostPathMountPropagation

		podSpec := &target.Job.Spec.Template.Spec
		container := &podSpec.Containers[0]

		volumesAndMounts := composer.VolumesAndMounts{
			Config: &target.Config,
		}

		runner := &composer.StepRunner{}

		if src != nil {
			runner.Batch(
				itc.fc.ComposeHostPath(replicaBasedItemPrefix(itemized.HostPathPrefix, nil), src.HostPath, &volumesAndMounts, composer.MapHostPathToK8sVol, nil, &mountPropagation),
				itc.fc.ComposeNfs(replicaBasedItemPrefix(itemized.NfsPrefix, nil), src.Nfs, &volumesAndMounts, composer.MapNfsToK8sVol, nil),
				itc.fc.ComposeConfigMap(replicaBasedItemPrefix(itemized.ConfigMapPrefix, nil), src.ConfigMapVolume, &volumesAndMounts, composer.MapConfigMapToK8sVol, nil),
				itc.fc.ComposeEmptyDir(replicaBasedItemPrefix(itemized.EmptyDirPrefix, nil), src.EmptyDirVolume, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
				itc.fc.ComposeSecretVolume(replicaBasedItemPrefix(itemized.SecretStoragePrefix, nil), src.SecretVolume, &volumesAndMounts, composer.MapSecretToK8sVol, nil),

				itc.fc.ComposeGit(replicaBasedItemPrefix(itemized.GitPrefix, nil), src.Git, podSpec, composer.MapGitToWorkloadK8s, nil),
				itc.fc.ComposePvc(replicaBasedItemPrefix(itemized.PvcPrefix, nil), src.Pvc, &volumesAndMounts, composer_v2.MapPvcToWorkloadRunner, nil),
				itc.fc.ComposeS3(replicaBasedItemPrefix(itemized.S3Prefix, nil), src.S3, uidGid, podSpec, composer_v2.MapS3ToWorkloadK8s, nil),
				itc.fc.ComposeDataVolume(replicaBasedItemPrefix(itemized.DataVolumePrefix, nil), src.DataVolume, &volumesAndMounts, composer.MapDataVolumesToK8s, itc.DataVolumeClaimNames),
			)
		}

		runner.Batch(
			itc.fc.ComposeHomeDirectory(createHomeDir, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			itc.fc.ComposeLargeShm(largeShmRequest, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			itc.fc.CopyBoolean(itc.fc.Options.CleanupCreatedPvcsOnCompletion, volumesAndMounts.Config, composer_v2.BoolToPvcsCleanupConfig, nil),
		)

		if len(volumesAndMounts.VolumeMount) != 0 {
			container.VolumeMounts = append(container.VolumeMounts, volumesAndMounts.VolumeMount...)
		}

		if len(volumesAndMounts.Volume) != 0 {
			podSpec.Volumes = append(podSpec.Volumes, volumesAndMounts.Volume...)
		}

		return runner.Error()
	}
}

func (itc *InterTrainToJobComposer) processAffinityFields(_ context.Context, src *workloadsApi.SupersetSpec,
	target *BatchJobTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		runner := &composer.StepRunner{}

		podSpec := &target.Job.Spec.Template.Spec

		if podSpec.Affinity == nil {
			podSpec.Affinity = &corev1.Affinity{}
		}

		nodePoolsInfo := &composer.NodePoolsInfo{
			RequestedNodePools: src.NodePools,
			ActualNodePools:    itc.ActualNodePools,
		}

		runner.Batch(
			itc.fc.ComposeNodeAffinity(src.NodeAffinityRequired.Get(), &podSpec.Affinity.NodeAffinity, composer.NodeAffinityRequiredToK8s, nil),
			itc.fc.ComposePodAffinity(src.PodAffinity.Get(), &podSpec.Affinity.PodAffinity, composer.PodAffinityToK8s, itc.fc.WorkloadName),
			itc.fc.ComposeNodePools(nodePoolsInfo, &podSpec.Affinity.NodeAffinity, composer.NodePoolsToK8s, nil),
		)

		if composer.IsEmpty(*podSpec.Affinity) {
			podSpec.Affinity = nil
		}

		return runner.Error()
	}
}

// For interactive and training we do not have replicas, so just leave the prefix as is.
func replicaBasedItemPrefix(prefix string, _ *workloadsApi.ReplicaType) string {
	return prefix
}

const (
	backoffLimitDefaultWorkspace = int64(0)
	backoffLimitDefaultTraining  = int64(6)
)

func (itc *InterTrainToJobComposer) composeRestartPolicy(src *string,
	target *corev1.RestartPolicy,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// Jobs require an explicit restart policy, and only support Never/OnFailure.
		restartPolicy := convutils.Deref(src, string(workloadsApi.RESTARTPOLICY_NEVER))
		if restartPolicy == string(workloadsApi.RESTARTPOLICY_ALWAYS) {
			restartPolicy = string(workloadsApi.RESTARTPOLICY_ON_FAILURE)
		}

		return itc.fc.ComposeRestartPolicy(validJobRestartPolicies, &restartPolicy, target, composer.StringToString, nil)()
	}
}

func (itc *InterTrainToJobComposer) composeBackoffLimit(src *int64,
	target any, setter composer.FnIntSetter[int64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var val int64

		if src != nil {
			val = *src
		} else {
			switch itc.fc.WorkloadType {
			case composer.Workspace, composer.Interactive:
				val = backoffLimitDefaultWorkspace
			case composer.Training:
				val = backoffLimitDefaultTraining
			default:
				return nil
			}
		}

		return setter(&val, target, cookie)
	}
}
