/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	mnnvlbuilder "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/mnnvl"
	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"
	"time"

	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

type LwsBuilder struct {
	leaderWorkerSet *lwsv1.LeaderWorkerSet
	k8sClient       client.Client
	elasticEnabled  bool
}

func NewLwsBuilder(k8sClient client.Client, elasticEnabled bool) LwsBuilder {
	return LwsBuilder{
		leaderWorkerSet: nil,
		k8sClient:       k8sClient,
		elasticEnabled:  elasticEnabled,
	}
}

func (lb LwsBuilder) GetPreview(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) (client.Object, error) {
	return &lwsv1.LeaderWorkerSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", constants.LwsGroup, constants.LwsVersion),
			Kind:       constants.LwsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetUniqueLwsName(workload),
			Namespace: workload.Namespace,
		},
	}, nil
}

func (lb LwsBuilder) Build(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) (*lwsv1.LeaderWorkerSet, error) {
	// init the leader worker set. we always init the leader template, even that in lws it's optional
	lb.leaderWorkerSet = &lwsv1.LeaderWorkerSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", constants.LwsGroup, constants.LwsVersion),
			Kind:       constants.LwsKind,
		},
	}
	lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate = &corev1.PodTemplateSpec{}

	// build the metadata
	err := lb.buildObjectMeta(ctx, workload)
	if err != nil {
		return nil, err
	}

	// build the spec
	err = lb.buildSpec(ctx, workload)
	if err != nil {
		return nil, err
	}

	// set the owner reference
	if owner != nil {
		err = controller_utils.SetControllerReference(owner, lb.leaderWorkerSet, lb.k8sClient.Scheme())
		if err != nil {
			return nil, fmt.Errorf("failed to set owner reference: %w", err)
		}
	}

	return lb.leaderWorkerSet, nil
}

func (lb LwsBuilder) buildObjectMeta(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	// handle lws meta
	err := buildGlobalObjectMeta(ctx, workload, &lb.leaderWorkerSet.ObjectMeta)
	if err != nil {
		return fmt.Errorf("failed to build lws object meta: %w", err)
	}

	// handle leader meta
	err = buildEntityObjectMeta(ctx, workload, &workload.Spec, &workload.LeaderSpec,
		&lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.ObjectMeta)
	if err != nil {
		return fmt.Errorf("failed to build leader object meta: %w", err)
	}

	// handle worker meta
	err = buildEntityObjectMeta(ctx, workload, &workload.Spec, GetDistributedInferenceWorkerSpec(workload),
		&lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.ObjectMeta)
	if err != nil {
		return fmt.Errorf("failed to build worker object meta: %w", err)
	}

	return nil
}

// buildGlobalObjectMeta builds the metadata for the given object.
// it handles the common metadata fields for the lws object.
func buildGlobalObjectMeta(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, om *metav1.ObjectMeta) error {
	if om.Annotations == nil {
		om.Annotations = map[string]string{}
	}

	if om.Labels == nil {
		om.Labels = map[string]string{}
	}

	om.Name = GetUniqueLwsName(workload)
	om.Namespace = workload.Namespace
	om.CreationTimestamp = metav1.Time{Time: time.Now()}

	om.Labels[constants.LabelWorkloadKind] = workload.Kind
	om.Labels[constants.LabelWorkloadName] = workload.Name
	om.Labels[constants.LabelRelease] = workload.Name
	om.Labels[constants.RunaiProjectLabel] = ctx.Value(workloadproject.ContextProjectName).(string)
	om.Labels[commonconstants.TopOwnerUidLabel] = topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
	if value := workload.Spec.Preemptibility.Deref(); value != "" {
		om.Labels[constants.LabelPreemptibilityName] = value
	}

	om.Annotations[constants.AnnotationWorkloadApiVersion] = workload.APIVersion

	return nil
}

// buildEntityObjectMeta builds the metadata for the given object.
// it handles the entity-specific metadata fields.
func buildEntityObjectMeta(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	generalSpec *dinfp.DistributedInferenceSpec, entitySpec *dinfp.DistributedInferenceEntitySpec, om *metav1.ObjectMeta) error {
	// start by resetting the ObjectMeta, as otherwise it comes from previous initializations with a non-nil creation timestamp,
	// which causes kube api warnings in the logs
	*om = metav1.ObjectMeta{}

	om.Annotations = map[string]string{}
	om.Annotations = utils.ConcatStringMaps(om.Annotations, workload.Annotations)
	// This annotation is changed on each agent operation on the workload, copying it to the LWS and eventually to the pods will lead to unnecessary restart
	delete(om.Annotations, constants.AnnotationsBackendWorkloadCounter)

	om.Labels = map[string]string{}
	om.Labels = utils.ConcatStringMaps(om.Labels, workload.Labels)

	om.Labels[commonconstants.TopOwnerUidLabel] = topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)

	// set the resource allocation annotations and the user-provided annotations and labels
	resourceAllocationAnnotations, err := createResourceAllocationAnnotationsForDistributedInference(ctx,
		&entitySpec.ResourceAllocationFields, &generalSpec.DistributedInferenceSchedulerFields)
	if err != nil {
		return err
	}
	om.Annotations = utils.ConcatStringMaps(om.Annotations, resourceAllocationAnnotations)

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(om, &entitySpec.ContainerFields)

	return nil
}

func (lb LwsBuilder) buildSpec(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	// handle LeaderTemplate - always taken from LeaderSpec, which is mandatory
	// we always set the leader template, even that in lws it's optional (and taken from the worker template if not set)
	err := lb.buildEntitySpec(ctx, workload.GetChildName(), workload.Namespace, LwsEntityNameLeader,
		workload, &workload.LeaderSpec, lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate)
	if err != nil {
		return fmt.Errorf("failed to build leader spec: %w", err)
	}

	// handle WorkerTemplate
	workerSpec := GetDistributedInferenceWorkerSpec(workload)
	err = lb.buildEntitySpec(ctx, workload.GetChildName(), workload.Namespace, LwsEntityNameWorker,
		workload, workerSpec, &lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate)
	if err != nil {
		return fmt.Errorf("failed to build worker spec: %w", err)
	}

	// handle Size
	if workload.Spec.Workers != nil {
		lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.Size = ptr.To(int32(workload.Spec.Workers.Deref() + 1))
	}

	// handle Replicas
	if workload.Spec.Replicas != nil {
		lb.leaderWorkerSet.Spec.Replicas = ptr.To(int32(workload.Spec.Replicas.Deref()))
	}

	// handle startup policy
	startupPolicy := workload.Spec.StartupPolicy.Deref()
	if startupPolicy == "" {
		// lws doesn't allow empty startup policy, so we set the default one
		startupPolicy = constants.DefaultDistributedInferenceStartupPolicy
	} else if startupPolicy != constants.DistributedInferenceStartupPolicyLeaderCreated && startupPolicy != constants.DistributedInferenceStartupPolicyLeaderReady {
		return fmt.Errorf("invalid startup policy: %s. Valid options are: %s, %s",
			startupPolicy, constants.DistributedInferenceStartupPolicyLeaderCreated, constants.DistributedInferenceStartupPolicyLeaderReady)
	}
	lb.leaderWorkerSet.Spec.StartupPolicy = lwsv1.StartupPolicyType(startupPolicy)

	// handle restart policy
	restartPolicy := workload.Spec.RestartPolicy.Deref()
	if restartPolicy != "" {
		if restartPolicy != constants.DistributedInferenceRestartPolicyRecreateGroupOnPodRestart && restartPolicy != constants.DistributedInferenceRestartPolicyNone {
			return fmt.Errorf("invalid restart policy: %s. Valid options are: %s, %s",
				restartPolicy, constants.DistributedInferenceRestartPolicyRecreateGroupOnPodRestart, constants.DistributedInferenceRestartPolicyNone)
		}
		lb.leaderWorkerSet.Spec.LeaderWorkerTemplate.RestartPolicy = lwsv1.RestartPolicyType(restartPolicy)
	}

	// handle rollout strategy. we currently don't have this in our api so we're using RollingUpdate as default,
	// which is the default (and only) strategy in the lws api
	lb.leaderWorkerSet.Spec.RolloutStrategy.Type = lwsv1.RollingUpdateStrategyType

	return nil
}

func (lb LwsBuilder) buildEntitySpec(ctx context.Context, workloadName string, workloadNamespace string, entityName string,
	workload *dinfw.DistributedInferenceWorkload, entitySpec *dinfp.DistributedInferenceEntitySpec, template *corev1.PodTemplateSpec) error {
	generalSpec := &workload.Spec

	objNaming := &metav1.ObjectMeta{
		Name: workloadName,
	}

	mnnvlBuilder := NewMnnvlBuilder(lb.k8sClient, lb.elasticEnabled)
	mnnvlEnabled := mnnvlBuilder.IsMnnvlRequired(workload) && gpu_validator.HasRequestedGpus(&entitySpec.ResourceAllocationFields)
	topologyEnabled := topology.GetState(ctx, workload.Annotations).Enabled()

	// scheduler name
	template.Spec.SchedulerName = constants.RunaiScheduler

	// containers
	containers, initContainers, err := lb.createContainers(ctx, entityName, generalSpec, entitySpec, objNaming, mnnvlEnabled)
	if err != nil {
		return err
	}

	template.Spec.Containers = containers
	template.Spec.InitContainers = initContainers

	// volumes (involves auth as well, so using the serving port builder)
	networkingFields := field_groups.NetworkingFields{}
	authRequired := false
	if entityName == LwsEntityNameLeader {
		servingPortBuilder := NewServingPortBuilder(lb.k8sClient, false)
		authRequired = servingPortBuilder.IsAuthRequired(generalSpec.ServingPorts)
	}
	volumes, err := workloadbuilders.GetVolumes(ctx, &entitySpec.StorageFields, &entitySpec.ContainerFields, &networkingFields,
		lb.k8sClient, workloadNamespace, workloadName, entityName, false, true, authRequired)
	if err != nil {
		return err
	}

	template.Spec.Volumes = volumes

	// tolerations
	tolerations, err := workloadbuilders.GetTolerations(ctx, &entitySpec.ContainerFields)
	if err != nil {
		return err
	}

	template.Spec.Tolerations = tolerations

	// mnnvl & node affinity for node pools
	// if mnnvl is enabled, node affinity is taken care of by the mnnvl builder
	// mnnvl is configured for the entity only if it has gpus requested
	if mnnvlEnabled {
		workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)

		nodepoolList, err := nodepoolbuilders.GetNodePools(ctx, lb.k8sClient, generalSpec.NodePools.Deref(), "")
		if err != nil {
			return err
		}

		// handles both node affinity, pod affinity and resource claim
		// container is nil as it's handled in createContainers
		err = mnnvlbuilder.ConfigurePodForMNNVL(ctx, DistributedInferenceWorkloadInitial, workloadName,
			workloadId, mnnvltypes.Required, nodepoolList, &template.Spec, nil, topologyEnabled)
		if err != nil {
			return err
		}
	} else if !topologyEnabled {
		// no mnnvl, so we create node affinity for node pools
		err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, lb.k8sClient, generalSpec.NodePools.Deref(), "", &template.Spec.Affinity)
		if err != nil {
			return err
		}
	}

	// user-provided node affinity
	workloadbuilders.MergeNodeAffinity(&entitySpec.NodeAffinityFields, &template.Spec.Affinity)

	// pod affinity
	err = workloadbuilders.CreatePodAffinity(entitySpec.PodAffinity.Deref(), entitySpec.PodAffinityTopology.Deref(),
		entitySpec.PodAffinitySchedulingRule.Deref(), &template.Spec.Affinity, workloadName, false)
	if err != nil {
		return err
	}

	// node type
	if entitySpec.NodeType.Deref() != "" {
		template.Spec.NodeSelector = map[string]string{
			constants.NodeTypeKey: entitySpec.NodeType.Deref(),
		}
	}

	err = workloadbuilders.SetPriorityClassName(ctx, lb.k8sClient, generalSpec.PriorityClassName, &template.Spec)
	if err != nil {
		return err
	}

	err = workloadbuilders.BuildImagePullSecrets(ctx, entitySpec.ImagePullSecrets, lb.k8sClient, workloadNamespace, &template.Spec.ImagePullSecrets, workloadName, DistributedInferenceWorkloadInitial)
	if err != nil {
		return err
	}

	// pod security context
	template.Spec.SecurityContext = createPodSecurityContext(&entitySpec.ContainerFields)

	template.Spec.HostIPC = entitySpec.HostIpc.Deref()
	template.Spec.HostNetwork = entitySpec.HostNetwork.Deref()
	if template.Spec.HostNetwork {
		template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	// restart policy
	if restartPolicy := entitySpec.ContainerFields.RestartPolicy.Deref(); restartPolicy != "" {
		template.Spec.RestartPolicy = corev1.RestartPolicy(restartPolicy)
	}

	return nil
}

func (lb LwsBuilder) createEnvironmentVariables(ctx context.Context, spec *dinfp.DistributedInferenceEntitySpec, objNaming *metav1.ObjectMeta) []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)
	for _, name := range field_types.SortedDictionaryMapKeys(spec.Environment.Deref()) {
		value := (spec.Environment.Deref())[name]
		if !ptr.Deref(value.Deleted, false) {
			envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, name, value.Value, &spec.ContainerFields, objNaming, objNaming.Name, DistributedInferenceWorkloadInitial))
		}
	}

	if workloadbuilders.NeedToSetupRunaiHome(&spec.ContainerFields, false) {
		envs = append(envs, corev1.EnvVar{
			Name:  constants.HomeDirEnvName,
			Value: constants.RunaiDefaultHomeDir,
		})
	}

	return envs
}

func (lb LwsBuilder) createResourceRequirements(ctx context.Context, spec *dinfp.DistributedInferenceEntitySpec) (*corev1.ResourceRequirements, []corev1.EnvVar, error) {
	resourceRequirements, resourcesEnvVars, err := workloadbuilders.GetResourcesReqAndLimits(ctx, &spec.ResourceAllocationFields)
	if err != nil {
		return nil, nil, err
	}

	if resourceRequirements == nil {
		resourceRequirements = &corev1.ResourceRequirements{}
	}

	return resourceRequirements, resourcesEnvVars, nil
}

func (lb LwsBuilder) createContainers(ctx context.Context, entityName string,
	generalSpec *dinfp.DistributedInferenceSpec, spec *dinfp.DistributedInferenceEntitySpec,
	objNaming *metav1.ObjectMeta, mnnvlEnabled bool) ([]corev1.Container, []corev1.Container, error) {
	// env vars
	envVars := lb.createEnvironmentVariables(ctx, spec, objNaming)

	// resource and resource-related env vars
	resourceRequirements, resourcesEnvVars, err := lb.createResourceRequirements(ctx, spec)
	if err != nil {
		return nil, nil, err
	}

	envVars = append(envVars, resourcesEnvVars...)

	// command
	commandStr := workloadbuilders.PatchTemplateVariables(spec.Command.Deref(), &spec.ContainerFields, objNaming, ctx)
	command, err := workloadbuilders.ParseCommandLine(commandStr)
	if err != nil {
		return nil, nil, err
	}

	// arguments
	argsStr := workloadbuilders.PatchTemplateVariables(spec.Arguments.Deref(), &spec.ContainerFields, objNaming, ctx)
	args, err := workloadbuilders.ParseCommandLine(argsStr)
	if err != nil {
		return nil, nil, err
	}

	// security context
	securityContext := createContainerSecurityContext(&spec.ContainerFields, spec.Privileged)

	// ports and auth init containers - leader only
	var ports []corev1.ContainerPort
	var containers []corev1.Container
	var initContainers []corev1.Container
	if entityName == LwsEntityNameLeader {
		servingPortBuilder := NewServingPortBuilder(lb.k8sClient, false)
		ports, err = servingPortBuilder.GetPorts(generalSpec.ServingPorts)
		if err != nil {
			return nil, nil, err
		}

		authContainers, authInitContainers, err := servingPortBuilder.GetAuthContainers(ctx, objNaming.Name, generalSpec.ServingPorts)
		if err != nil {
			return nil, nil, err
		}

		containers = append(containers, authContainers...)
		initContainers = append(initContainers, authInitContainers...)
	}

	// volume mounts
	volumeMounts, err := workloadbuilders.GetVolumeMounts(ctx, &spec.StorageFields, &spec.ContainerFields, false, false)
	if err != nil {
		return nil, nil, err
	}

	// initialize the container
	container := corev1.Container{
		Name:            entityName,
		Image:           spec.ContainerFields.Image.Deref(),
		ImagePullPolicy: corev1.PullPolicy(spec.ContainerFields.ImagePullPolicy.Deref()),
		WorkingDir:      spec.ContainerFields.WorkingDir.Deref(),
		Resources:       *resourceRequirements,
		Command:         command,
		Args:            args,
		Env:             envVars,
		SecurityContext: securityContext,
		Ports:           ports,
		Stdin:           spec.ContainerFields.Stdin.Deref(),
		TTY:             spec.ContainerFields.Tty.Deref(),
		VolumeMounts:    volumeMounts,
	}

	// probes
	err = workloadbuilders.CreateProbes(ctx, &spec.ContainerFields, &container, false)
	if err != nil {
		return nil, nil, err
	}
	// for the leader pod, if no readiness probe is provided, we create a default one for the serving port
	if entityName == LwsEntityNameLeader && container.ReadinessProbe == nil && len(ports) != 0 {
		err = workloadbuilders.CreateTcpReadinessProbe(ports, &container)
		if err != nil {
			return nil, nil, err
		}
	}

	// mnnvl
	if mnnvlEnabled {
		mnnvlbuilder.ConfigureContainerForMNNVL(&container)
	}

	containers = append(containers, container)

	return containers, initContainers, nil
}
