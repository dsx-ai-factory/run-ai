/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	trackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

var validRestartPolicies = datastructs.NewSet(
	string(workloadsApi.RESTARTPOLICY_ALWAYS),
	string(workloadsApi.RESTARTPOLICY_NEVER),
	string(workloadsApi.RESTARTPOLICY_ON_FAILURE),
)

var validRestartPoliciesDinf = datastructs.NewSet(
	string(workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
	string(workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_NONE),
)

type JobSpecComposer struct {
	fc          *composer.FieldsComposer
	ReplicaType *workloadsApi.ReplicaType

	AssetsMapping        []workloadsApi.AssetsMapping
	WorkloadRequirements composer.WorkloadRequirements
}

func NewSupersetToJobSpecComposer(
	workloadType string,
	workloadName string,
	requester string,
	replicaType *workloadsApi.ReplicaType,
	assetsMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
) *JobSpecComposer {
	result := JobSpecComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      workloadType,
			WorkloadName:      workloadName,
			SubmittedBy:       requester,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		ReplicaType:   replicaType,
		AssetsMapping: assetsMapping,
	}

	// in runaijob we need to normalize the serving port auth type based on cluster supporting
	// the new set of auth type values
	result.fc.Options.NormalizeServingPortAuthType = true

	return &result
}

func (jsw *JobSpecComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *JobSpecComposer {
	jsw.fc.StorageClassConfiguration = config
	return jsw
}

// Compose function takes care to compose the common fields to all the workload types.
func (jsw *JobSpecComposer) Compose(ctx context.Context, src *workloadsApi.SupersetSpec,
	target *jobs.JobSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var jupyterContainerPort int32

		runner := &composer.StepRunner{}

		runner.Batch(
			jsw.fc.ComposeFullySaasInd(&target.Labels, boolToDictField, composer.ControlPlaneManagedLabel),
			jsw.fc.CopyString(convutils.Ref(jsw.fc.SubmittedBy), &target.UserName, stringToStringField, nil),
			jsw.fc.ComposeImage(src.Image.Get(), &target.Image, stringToStringField, nil),
			jsw.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &target.ImagePullPolicy, stringToStringField, nil),
			jsw.fc.ComposeImagePullSecrets(jsw.AssetsMapping, src.ImagePullSecrets, &target.ImagePullSecrets, arrayToStringField, nil),
			jsw.fc.CopyString(src.Command.Get(), &target.Command, nonEmptyStringToStringField, nil),
			jsw.fc.CopyString(src.Args.Get(), &target.Arguments, nonEmptyStringToStringField, nil),
			jsw.composeEnvironmentVariables(jsw.AssetsMapping, src.EnvironmentVariables, &target.Environment, envVarsToDictField, nil),
			jsw.fc.ComposeLabels(src.Labels, &target.Labels, mapToDictField, nil),
			jsw.fc.ComposeAnnotations(src.Annotations, &target.Annotations, mapToDictField, nil),
			jsw.fc.CopyString(src.WorkingDir.Get(), &target.WorkingDir, stringToStringField, nil),
			jsw.fc.CopyBoolean(src.CreateHomeDir.Get(), &target.CreateHomeDir, boolToBoolField, nil),
			jsw.fc.CopyBoolean(src.Tty.Get(), &target.Tty, boolToBoolField, nil),
			jsw.fc.CopyBoolean(src.Stdin.Get(), &target.Stdin, boolToBoolField, nil),
			jsw.fc.ComposeExposedUrls(replicaBasedItemPrefix(itemized.ExposedUrlPrefix, jsw.ReplicaType),
				src.ExposedUrls, &target.ExposedUrls, mapToExposedUrls, nil, &jupyterContainerPort, nil,
				false /*forceSuffix*/),
			jsw.InjectJupyterPortEnv(&jupyterContainerPort, target),
			jsw.fc.ComposePorts(replicaBasedItemPrefix(itemized.PortPrefix, jsw.ReplicaType),
				src.Ports, &target.Ports, mapToPortField, nil),
			jsw.fc.CopyString(src.NodeType.Get(), &target.NodeType, stringToStringField, nil), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
			jsw.fc.CopyArray(src.NodePools, &target.NodePools, arrayToStringField, nil),
			jsw.fc.CopyString(src.Category.Get(), &target.Labels, stringToDictField, composer.WorkloadCategoryLabel),
			jsw.fc.ComposePriorityClass(ctx, src.PriorityClass.Get(),
				&target.PriorityClassName, stringToStringField,
				// nolint:staticcheck
				&target.Preemptible, boolToBoolField, nil),
			jsw.fc.ComposePreemptibility(ctx, src.Preemptibility.Get(), &target.Preemptibility, stringToStringField, nil),
			jsw.fc.ComposeAutoDeletionAfterTermination(src.AutoDeletionTimeAfterCompletionSeconds.Get(), &target.AutoDeletionTimeAfterCompletionSeconds, intToIntField, nil),
			jsw.fc.CopyBoolean(src.TerminateAfterPreemption.Get(), &target.TerminateAfterPreemption, boolToBoolField, nil),
			jsw.fc.CopyInt(src.BackoffLimit.Get(), &target.BackoffLimit, nonZeroIntToIntField, nil),
			jsw.fc.CopyInt(src.Completions.Get(), &target.Completions, nonZeroIntToIntField, nil),
			jsw.fc.ComposeParallelism(src.Parallelism.Get(), src.Completions.Get(), &target.Parallelism, nonZeroIntToIntField, nil),
			jsw.fc.ComposeTerminationGracePeriod(ctx, src.TerminationGracePeriodSeconds.Get(), &target.TerminationGracePeriodSeconds, nonZeroIntToIntField, nil),
			jsw.ComposeCompute(ctx, src.Compute.Get(), target),
		)

		var uidGid composer.UidGid
		runner.Run(jsw.ComposeSecurity(ctx, src.Security.Get(), target, &uidGid))

		runner.Run(jsw.ComposeStorage(ctx, src.Storage.Get(), target, uidGid))

		if jsw.fc.WorkloadType == string(assetsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
			runner.Run(jsw.fc.ComposeRestartPolicy(validRestartPoliciesDinf, src.RestartPolicy.Get(), &target.RestartPolicy, stringToStringField, nil))
		} else {
			runner.Run(jsw.fc.ComposeRestartPolicy(validRestartPolicies, src.RestartPolicy.Get(), &target.RestartPolicy, stringToStringField, nil))
		}

		var portAuthInfo composer.PortAuthInfo

		if jsw.fc.WorkloadType == string(workloadsApi.WORKLOADTYPE_INFERENCE) {
			runner.Run(jsw.fc.ComposeServingPort(jsw.ReplicaType, src.ServingPort.Get(), &target.Ports, servingPortToPortField, nil, &portAuthInfo))

			if src.ServingPort.Get() != nil {
				runner.Batch(
					jsw.fc.CopyBoolean(src.ServingPort.Get().ClusterLocalAccessOnly.Get(), &target.IsPrivateServiceUrl, boolToBoolField, nil),
					jsw.fc.ComposeServingPortAuthType(src.ServingPort.Get().AuthorizationType.Get(), &portAuthInfo, &target.ServingPortAuthorizationType, stringToStringField, nil),
					jsw.fc.CopyArray(portAuthInfo.AuthorizedUsers, &target.ServingPortAuthorizedUsers, arrayToStringField, ","),
					jsw.fc.CopyArray(portAuthInfo.AuthorizedGroups, &target.ServingPortAuthorizedGroups, arrayToStringField, ","),
				)
			}
		}

		if jsw.fc.WorkloadType == string(assetsApi.WORKLOADTYPE_INFERENCE) {
			runner.Run(jsw.fc.ComposeAutoScaling(src.Autoscaling.Get(), &target.AutoScalingFields, autoScalingSetter, nil))
		}

		if convutils.HasValue(src.ServingConfiguration) {
			runner.Run(jsw.fc.CopyInt(src.ServingConfiguration.Get().InitializationTimeoutSeconds.Get(), &target.InitializationTimeoutSeconds, intToIntField, nil))
			runner.Run(jsw.fc.CopyInt(src.ServingConfiguration.Get().RequestTimeoutSeconds.Get(), &target.RequestTimeoutSeconds, intToIntField, nil))
		}

		runner.Run(jsw.fc.ComposeNodeAffinity(src.NodeAffinityRequired.Get(), &target.NodeAffinityRequired, nodeAffinitySetter, nil))

		if convutils.HasValue(src.PodAffinity) {
			runner.Batch(
				jsw.fc.CopyBoolean(convutils.Ref(true), &target.PodAffinity, boolToBoolField, nil),
				jsw.fc.ComposePodAffinityKey(src.PodAffinity.Get().Key.Get(), &target.PodAffinityTopology, stringToStringField, nil),
				jsw.fc.ComposePodAffinityType(src.PodAffinity.Get().Type.Get(), &target.PodAffinitySchedulingRule, stringToStringField, nil),
			)
		}

		runner.Run(jsw.fc.ComposeTolerations(replicaBasedItemPrefix(itemized.TolerationPrefix, jsw.ReplicaType), src.Tolerations, &target.Tolerations, mapToTolerationSetter, nil))

		if convutils.HasValue(src.Probes) {
			runner.Run(jsw.fc.ComposeReadinessProbe(src.Probes.Get().Readiness.Get(), &target.ReadinessProbe, probeSetter, nil))
		}

		return runner.Error()
	}
}

func (jsw *JobSpecComposer) ComposeCompute(ctx context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
	target *jobs.JobSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		var sizingInfo composer.GpuSizingInfo

		runner := &composer.StepRunner{}

		runner.Batch(
			jsw.fc.CopyBoolean(src.LargeShmRequest.Get(), &target.LargeShm, boolToBoolField, nil),
			jsw.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &target.Cpu, nonZeroFloatToNumberField, nil),
			jsw.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &target.CpuLimit, nonZeroFloatToNumberField, nil),
			jsw.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &target.Memory, nonEmptyStringToQuantityField, nil),
			jsw.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &target.MemoryLimit, nonEmptyStringToQuantityField, nil),
			jsw.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuPortionRequest.Get(), src.GpuPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), gpuSizingInfoFieldsSetter,
				&target.ResourceAllocationFields, nil, &sizingInfo),
			jsw.fc.ComposeExtendedResources(replicaBasedItemPrefix(itemized.ExtendedResourceNickPrefix, jsw.ReplicaType), src.ExtendedResources,
				&target.ExtendedResources, extendedResourcesFieldSetter, nil),
		)

		if sizingInfo.MemoryLimit != nil || sizingInfo.PortionLimit != nil {
			jsw.WorkloadRequirements.GpuOptimization = true
		}

		return runner.Error()
	}
}

func (jsw *JobSpecComposer) ComposeSecurity(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfSecurity,
	target *jobs.JobSpec, uidGid *composer.UidGid,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		runner := &composer.StepRunner{}

		runner.Batch(
			jsw.fc.CopyBoolean(src.HostIpc.Get(), &target.HostIpc, boolToBoolField, nil),
			jsw.fc.CopyBoolean(src.HostNetwork.Get(), &target.HostNetwork, boolToBoolField, nil),
			jsw.fc.CopyBoolean(src.Privileged.Get(), &target.Privileged, boolToBoolField, nil),
			jsw.fc.CopyBoolean(src.ReadOnlyRootFilesystem.Get(), &target.ReadOnlyRootFilesystem, boolToBoolField, nil),
			jsw.fc.ComposeAllowPrivilegeEscalation(src.AllowPrivilegeEscalation.Get(), src.Privileged.Get(), &target.AllowPrivilegeEscalation, boolToBoolField, nil),
			jsw.fc.ComposeRunAsNonRoot(src.RunAsNonRoot.Get(), src.Privileged.Get(), &target.RunAsNonRoot, boolToBoolField, nil),
			jsw.fc.CopyInt(src.RunAsUid.Get(), &target.RunAsUid, intToIntField, nil),
			jsw.fc.CopyInt(src.RunAsGid.Get(), &target.RunAsGid, intToIntField, nil),
			jsw.fc.ComposeSupplementalGroups(src.SupplementalGroups.Get(), &target.SupplementalGroups, stringToStringField, nil),
			jsw.fc.CopyString(composer.EnumToString(src.SeccompProfileType.Get()), &target.SeccompProfileType, stringToStringField, nil),
			jsw.fc.ComposeRunAsUser(composer.EnumToString(src.UidGidSource.Get()), &target.RunAsUser, boolToBoolField, nil),
			jsw.fc.ComposeCapabilities(src.Capabilities, &target.Capabilities, mapToCapabilitiesField, nil),
		)

		uidGid.Uid = src.RunAsUid.Get()
		uidGid.Gid = src.RunAsGid.Get()

		return runner.Error()
	}
}

func (jsw *JobSpecComposer) InjectJupyterPortEnv(src *int32,
	target *jobs.JobSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		containerPort := convutils.Deref(src, 0)
		if containerPort == 0 {
			return nil
		}

		if target.Environment == nil {
			target.Environment = field_types.NewDictionaryField(map[string]string{})
		}

		envVars := target.Environment

		for envVar := range *envVars.Value {
			if envVar == composer.JupyterPortEnvVar {
				return nil // variable already provided
			}
		}

		(*target.Environment.Value)[composer.JupyterPortEnvVar] = field_types.DictionaryItemAndRules{
			Value: convutils.Ref(fmt.Sprintf("%d", containerPort)),
		}

		return nil
	}
}

func (jsw *JobSpecComposer) ComposeStorage(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfStorage,
	target *jobs.JobSpec, uidGid composer.UidGid,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		var mountPropagation workloadsApi.HostPathMountPropagation

		runner := &composer.StepRunner{}

		runner.Batch(
			jsw.fc.ComposeHostPath(replicaBasedItemPrefix(itemized.HostPathPrefix, jsw.ReplicaType), src.HostPath, &target.Volumes, mapToVolumesField, nil, &mountPropagation),
			jsw.fc.ComposeMountPropagation(&mountPropagation, &target.MountPropagation, boolToBoolField, nil),
			jsw.fc.ComposeNfs(replicaBasedItemPrefix(itemized.NfsPrefix, jsw.ReplicaType), src.Nfs, &target.Nfs, mapToNfsField, nil),
			jsw.fc.ComposeGit(replicaBasedItemPrefix(itemized.GitPrefix, jsw.ReplicaType), src.Git, &target.GitSync, mapToGitSyncField, nil),
			jsw.fc.ComposePvc(replicaBasedItemPrefix(itemized.PvcPrefix, jsw.ReplicaType), src.Pvc, &target.Pvcs, mapToPvcField, nil),
			jsw.fc.ComposeS3(replicaBasedItemPrefix(itemized.S3Prefix, jsw.ReplicaType), src.S3, uidGid, &target.S3buckets, mapToS3Field, nil),
			jsw.fc.ComposeConfigMap(replicaBasedItemPrefix(itemized.ConfigMapPrefix, jsw.ReplicaType), src.ConfigMapVolume, &target.ConfigMapVolumes, mapToConfigMapField, nil),
			jsw.fc.ComposeEmptyDir(replicaBasedItemPrefix(itemized.EmptyDirPrefix, jsw.ReplicaType), src.EmptyDirVolume, &target.EmptyDirVolumes, mapToEmptyDirField, nil),
			jsw.fc.ComposeDataVolume(replicaBasedItemPrefix(itemized.DataVolumePrefix, jsw.ReplicaType), src.DataVolume, &target.DataVolumes, mapToDataVolumeField, nil),
			jsw.fc.ComposeSecretVolume(replicaBasedItemPrefix(itemized.SecretStoragePrefix, jsw.ReplicaType), src.SecretVolume, &target.SecretVolumes, mapToSecretVolumeField, nil),
		)

		return runner.Error()
	}
}

func replicaBasedItemPrefix(prefix string, replicaType *workloadsApi.ReplicaType) string {
	if replicaType == nil {
		return prefix
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(string(*replicaType)))
}

func (jsw *JobSpecComposer) composeEnvironmentVariables(assetsMapping []workloadsApi.AssetsMapping, src []workloadsApi.EnvironmentVariable,
	target any, setter composer.FnEnvVarSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		completeList := make([]workloadsApi.EnvironmentVariable, 0, len(src)+2)

		// add runa:ai internal environment variables
		completeList = append(completeList, workloadsApi.EnvironmentVariable{
			Name:  *workloadsApi.NewNullableString(convutils.Ref("RUNAI_PROJECT")),
			Value: *workloadsApi.NewNullableString(convutils.Ref("${RUNAI_PROJECT}")),
		})
		completeList = append(completeList, workloadsApi.EnvironmentVariable{
			Name:  *workloadsApi.NewNullableString(convutils.Ref("RUNAI_JOB_NAME")),
			Value: *workloadsApi.NewNullableString(convutils.Ref("${RUNAI_JOB_NAME}")),
		})

		if len(src) > 0 {
			completeList = append(completeList, src...)
		}

		return jsw.fc.ComposeEnvironmentVariables(assetsMapping, completeList, target, setter, cookie, nil)()
	}
}
