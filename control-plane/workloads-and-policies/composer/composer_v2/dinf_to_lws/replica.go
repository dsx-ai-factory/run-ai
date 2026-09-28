/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
)

func (dinfComposer *DistributedInferenceToLWSComposer) processReplica(
	ctx context.Context,
	common *workloadsApi.DistributedInferenceSpecSpec,
	src *workloadsApi.DistributedInferenceLeaderWorkerSpec,
	target *corev1.PodTemplateSpec,
	role string,
	assetsMapping []workloadsApi.AssetsMapping,
	fullTarget *LeaderWorkerSetTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		podSpec := &target.Spec
		container := &podSpec.Containers[0]
		envVarsUsingCreds := make([]composer.EnvVarInfo, 0)
		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.CopyString(convutils.Ref(composer.RunaiSchedulerName), &podSpec.SchedulerName, composer.StringToString, nil),
			dinfComposer.fc.CopyString(convutils.Ref(role), &container.Name, composer.StringToString, nil),
			dinfComposer.fc.ComposeImage(src.Image.Get(), &container.Image, composer.StringToString, nil),
			dinfComposer.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &container.ImagePullPolicy, composer.StringToString, nil),
			dinfComposer.fc.CopyString(src.Command.Get(), &container.Command, composer.UnixStrToStringArray, nil),
			dinfComposer.fc.CopyString(src.Args.Get(), &container.Args, composer.UnixStrToStringArray, nil),
			dinfComposer.fc.CopyString(src.WorkingDir.Get(), &container.WorkingDir, composer.StringToString, nil),
			dinfComposer.fc.ComposeHomeDirEnvVar(src.CreateHomeDir.Get(), &container.Env, composer.EnvVarsToK8sEnv, nil),
			dinfComposer.fc.ComposeLabels(src.Labels, &target.Labels, composer.MapToMapString, nil),
			dinfComposer.fc.ComposeNameAndTypeLabels(dinfComposer.fc.WorkloadName, dinfComposer.fc.WorkloadType, LeaderWorkerSetKind, dinfComposer.fc.ProjectName, &target.Labels, composer.MapToMapString, nil),
			dinfComposer.fc.ComposeAnnotations(src.Annotations, &target.Annotations, composer.MapToMapString, nil),
			dinfComposer.fc.CopyString(common.Category.Get(), &target.Labels, composer.StringToMapString, composer.WorkloadCategoryLabel),
			dinfComposer.fc.CopyString(src.NodeType.Get(), &podSpec.NodeSelector, composer.StringToMapString, "run.ai/node-type"), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
			dinfComposer.fc.ComposePreemptibility(ctx, common.Preemptibility.Get(), &target.Labels, composer.PreemptibilityToK8sLabel, nil),
			dinfComposer.fc.ComposePriorityClass(ctx, common.PriorityClass.Get(), &podSpec.PriorityClassName, composer.StringToString, nil, nil, nil),
		)

		var uidGid composer.UidGid
		runner.Run(dinfComposer.processSecurityFields(src.Security.Get(), target, &uidGid))

		rc := composer_v2.ExpandTemplateVarsInCommandAndArgs(container, dinfComposer.fc.ProjectName, dinfComposer.fc.WorkloadName)
		if rc != nil {
			return rc
		}

		runaiEnvVars := composer.RunaiEnvVars(dinfComposer.fc.ProjectName, dinfComposer.fc.WorkloadName, 0, &uidGid)

		runner.Batch(
			dinfComposer.fc.ComposeEnvironmentVariables(assetsMapping, runaiEnvVars, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),
			dinfComposer.fc.ComposeEnvironmentVariables(assetsMapping, src.EnvironmentVariables, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),
		)

		credentialsTarget := composer_v2.PersonalCredsTarget{Config: &fullTarget.Config, PodSpec: podSpec}
		credentialsCookie := composer_v2.PersonalCredsCookie{WorkloadName: dinfComposer.fc.WorkloadName}
		runner.Run(dinfComposer.fc.ComposeImagePullSecrets(assetsMapping, src.ImagePullSecrets, &credentialsTarget, composer_v2.ArrayToImagePullSecretsSetter, &credentialsCookie))

		envCredentialsTarget := composer_v2.EnvVarUsingCredTarget{Config: &fullTarget.Config, EnvVars: &container.Env}
		runner.Run(dinfComposer.fc.ComposeEnvVarsUsingCreds(envVarsUsingCreds, &envCredentialsTarget, composer_v2.EnvVarsUsingCredSetter, &credentialsCookie))

		runner.Batch(
			dinfComposer.processComputeFields(ctx, src.Compute.Get(), target),
			dinfComposer.processStorageFields(src.Storage.Get(), src.CreateHomeDir.Get(), largeShmRequest(src.Compute.Get()), role, target, &fullTarget.Config),
			dinfComposer.processPlacementFields(common.NodePools, src, target),
			dinfComposer.fc.ComposeTolerations(replicaItemPrefix(itemized.TolerationPrefix, role), src.Tolerations, &podSpec.Tolerations, composer.MapToK8sTolerations, nil),
		)

		if src.Probes.Get() != nil {
			runner.Run(dinfComposer.fc.ComposeReadinessProbe(src.Probes.Get().Readiness.Get(), &container.ReadinessProbe, composer.ProbeToK8s, nil))
		}

		return runner.Error()
	}
}

func replicaItemPrefix(prefix, role string) string {
	return fmt.Sprintf("%s-%s", prefix, role)
}

func largeShmRequest(compute *workloadsApi.WorkspaceSpecSpecAllOfCompute) *bool {
	if compute == nil {
		return nil
	}
	return compute.LargeShmRequest.Get()
}

func (dinfComposer *DistributedInferenceToLWSComposer) processPlacementFields(
	nodePools []string,
	src *workloadsApi.DistributedInferenceLeaderWorkerSpec,
	target *corev1.PodTemplateSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if target.Spec.Affinity == nil {
			target.Spec.Affinity = &corev1.Affinity{}
		}

		nodePoolsInfo := &composer.NodePoolsInfo{
			RequestedNodePools: nodePools,
			ActualNodePools:    dinfComposer.ActualNodePools,
		}
		runner := &composer.StepRunner{}
		runner.Batch(
			dinfComposer.fc.ComposeNodeAffinity(src.NodeAffinityRequired.Get(), &target.Spec.Affinity.NodeAffinity, composer.NodeAffinityRequiredToK8s, nil),
			dinfComposer.fc.ComposePodAffinity(src.PodAffinity.Get(), &target.Spec.Affinity.PodAffinity, composer.PodAffinityToK8s, dinfComposer.fc.WorkloadName),
			dinfComposer.fc.ComposeNodePools(nodePoolsInfo, &target.Spec.Affinity.NodeAffinity, composer.NodePoolsToK8s, nil),
		)

		if composer.IsEmpty(*target.Spec.Affinity) {
			target.Spec.Affinity = nil
		}
		return runner.Error()
	}
}
