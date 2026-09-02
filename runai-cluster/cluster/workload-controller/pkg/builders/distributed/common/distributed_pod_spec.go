/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	mnnvlbuilder "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	nodepoolbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/nodepool"
	featuregate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/feature-gate"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	workloadcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	common2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/security"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const failedNodeAffinityMessage = "failed to create node affinity for workload %s: %w"

func (db *DistributedBuilder) buildPodSpec(ctx context.Context, podSpec *corev1.PodSpec, spec *dp.DistributedSpec,
	replicaType workloadcommon.ReplicaType, mnnvlMode mnnvl.Mode) (*corev1.PodSpec, error) {

	log := logf.FromContext(ctx)
	if db.ExisingPodsSpec != nil {
		if existingPodSpec, ok := db.ExisingPodsSpec[replicaType]; ok && existingPodSpec != nil {
			return existingPodSpec, nil
		}
	}

	gitSyncInitContainers, err := workloadbuilders.GetGitSyncInitContainersWithError(ctx, &spec.ContainerFields, &spec.StorageFields, false)
	if err != nil {
		return nil, err
	}
	podSpec.InitContainers = gitSyncInitContainers

	podSpec.Containers = []corev1.Container{}

	envs := make([]corev1.EnvVar, 0)

	for _, name := range flds.SortedDictionaryMapKeys(spec.ContainerFields.Environment.Deref()) {
		value := (spec.ContainerFields.Environment.Deref())[name]
		if !ptr.Deref(value.Deleted, false) {
			envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, name, value.Value, &spec.ContainerFields, db.ObjectNaming, db.GetChildName(), common.KindDistributedWorkloadInitial))
		}
	}

	//  add home directory env var, if needed to set up runai default home directory
	if workloadbuilders.NeedToSetupRunaiHome(&spec.ContainerFields, false) {
		envs = append(envs, corev1.EnvVar{
			Name:  constants.HomeDirEnvName,
			Value: constants.RunaiDefaultHomeDir,
		})
	}

	command := workloadbuilders.PatchTemplateVariables(spec.Command.Deref(), &spec.ContainerFields, db.ObjectNaming, ctx)
	args := workloadbuilders.PatchTemplateVariables(spec.Arguments.Deref(), &spec.ContainerFields, db.ObjectNaming, ctx)

	commandVerbs, err := workloadbuilders.ParseCommandLine(command)
	if err != nil {
		return nil, err
	}
	commandArgs, err := workloadbuilders.ParseCommandLine(args)
	if err != nil {
		return nil, err
	}
	// supplemental groups, if specified, should go to the pod security context
	err = common2.GetPodSecurityContext(&spec.ContainerFields, false, podSpec)
	if err != nil {
		return nil, err
	}
	sc := common2.GetSecurityContext(&spec.ContainerFields, &spec.AccessControlFields, false, &envs)

	jobResources, resourcesEnvVars, err := workloadbuilders.GetResourcesReqAndLimits(ctx, &spec.ResourceAllocationFields)
	if err != nil {
		return nil, err
	}

	envs = append(envs, resourcesEnvVars...)

	ports, err := workloadbuilders.GetPorts(ctx, &spec.NetworkingFields, false)
	if err != nil {
		return nil, err
	}

	err = workloadbuilders.IdentifyPvcsVolumeMode(ctx, spec.Pvcs, db.ObjectNaming.Namespace, db.K8sClient)
	if err != nil {
		return nil, err
	}

	volumeMounts, err := workloadbuilders.GetVolumeMounts(ctx, &spec.StorageFields, &spec.ContainerFields, false, true)
	if err != nil {
		return nil, err
	}

	container := corev1.Container{
		Image:           spec.Image.Deref(),
		ImagePullPolicy: corev1.PullPolicy(spec.ImagePullPolicy.DerefOrDefault(constants.PullPolicyIfNotPresent)),
		Name:            db.ContainerName,
		SecurityContext: sc,
		Command:         commandVerbs,
		Args:            commandArgs,
		Resources:       *jobResources,
		WorkingDir:      spec.WorkingDir.Deref(),
		VolumeMounts:    volumeMounts,
		VolumeDevices:   workloadbuilders.GetVolumeDevices(ctx, &spec.StorageFields),
		Env:             envs,
		Ports:           ports,
		Stdin:           spec.Stdin.Deref(),
		TTY:             spec.Tty.Deref(),
	}

	err = workloadbuilders.CreateProbes(ctx, &spec.ContainerFields, &container, true)
	if err != nil {
		return nil, err
	}

	topologyEnabled := db.GetTopologyState(ctx).Enabled()
	if mnnvlMode.IsEnabled() {
		err = db.SupportMNNVL(ctx, podSpec, spec, mnnvlMode, &container, topologyEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to support %s workload %s: %w ", mnnvl.UserFacingFeatureName, db.MergedWorkload.Name, err)
		}
	} else if !topologyEnabled {
		// Use the original node affinity logic when MNNVL is not enabled
		log.V(4).Info("Using nodepool node affinity")
		err = nodepoolbuilders.CreateNodeAffinityForNodePools(
			ctx,
			db.K8sClient,
			spec.NodePools.Deref(),
			spec.NodePool.Deref(),
			&podSpec.Affinity,
		)
		if err != nil {
			return nil, fmt.Errorf(failedNodeAffinityMessage, db.MergedWorkload.Name, err)
		}
	} else {
		log.V(4).Info("Using network topology")
	}

	log.V(4).Info("adding user pod affinity")
	err = workloadbuilders.CreatePodAffinity(
		spec.PodAffinity.Deref(),
		spec.PodAffinityTopology.Deref(),
		spec.PodAffinitySchedulingRule.Deref(),
		&podSpec.Affinity,
		db.GetChildName(),
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod affinity for workload %s: %w", db.GetChildName(), err)
	}

	// if the user provided node affinity, we need to merge it with the one we created
	workloadbuilders.MergeNodeAffinity(&spec.NodeAffinityFields, &podSpec.Affinity)

	podSpec.Containers = append(podSpec.Containers, container)

	// Add S3 goofs containers if needed
	s3container, err := workloadbuilders.GetS3containerWithError(ctx, &spec.WorkloadFields)
	if err != nil {
		return nil, err
	}

	if s3container != nil {
		featureGateUtil := featuregate.GetFeatureGateUtilInstance()

		if featureGateUtil.FeatureGateSidecarContainersEnable {
			podSpec.InitContainers = append(podSpec.InitContainers, s3container...)
		} else {
			podSpec.Containers = append(podSpec.Containers, s3container...)
		}
	}

	podSpec.SchedulerName = constants.RunaiScheduler

	podSpec.HostIPC = spec.HostIpc.Deref()
	podSpec.HostNetwork = spec.HostNetwork.Deref()
	if podSpec.HostNetwork {
		podSpec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	if spec.NodeType.Deref() != "" {
		podSpec.NodeSelector = map[string]string{
			constants.NodeTypeKey: spec.NodeType.Deref(),
		}
	}

	err = workloadbuilders.SetPriorityClassName(ctx, db.K8sClient, spec.PriorityClassName, podSpec)
	if err != nil {
		return nil, err
	}

	podSpec.Volumes, err = workloadbuilders.GetVolumes(ctx, &spec.StorageFields, &spec.ContainerFields, &spec.NetworkingFields, db.K8sClient, db.ObjectNaming.Namespace, db.GetChildName(),
		common.GetPvcNameSuffix(true, replicaType != workloadcommon.Worker), false, true, false)
	if err != nil {
		return nil, err
	}

	err = workloadbuilders.CreateEphemeralVolumes(spec.Pvcs, &podSpec.Volumes, db.GetChildName())
	if err != nil {
		return nil, err
	}

	podSpec.Tolerations, err = workloadbuilders.GetTolerations(ctx, &spec.ContainerFields)
	if err != nil {
		return nil, err
	}

	restartPolicy, err := workloadbuilders.GetRestartPolicy(spec.RestartPolicy)
	if err != nil {
		return nil, err
	}
	if restartPolicy != nil {
		podSpec.RestartPolicy = *restartPolicy
	}

	err = workloadbuilders.BuildImagePullSecrets(ctx, spec.ImagePullSecrets, db.K8sClient, db.ObjectNaming.Namespace, &podSpec.ImagePullSecrets, db.ObjectNaming.Name, common.KindDistributedWorkloadInitial)
	if err != nil {
		return nil, err
	}

	return podSpec, nil
}

func (db *DistributedBuilder) SupportMNNVL(ctx context.Context, podSpec *corev1.PodSpec, spec *dp.DistributedSpec, mnnvlMode mnnvl.Mode, container *corev1.Container, topologyEnabled bool) error {
	workloadID := topowner.GetRunaiWorkloadId(db.MergedWorkload.UID, db.MergedWorkload.Annotations)

	nodepoolList, err := nodepoolbuilders.GetNodePools(ctx, db.K8sClient, spec.NodePools.Deref(), spec.NodePool.Deref())
	if err != nil {
		return err
	}

	return mnnvlbuilder.ConfigurePodForMNNVL(ctx, common.KindDistributedWorkloadInitial, db.MergedWorkload.Name, workloadID, mnnvlMode, nodepoolList, podSpec, container, topologyEnabled)
}
