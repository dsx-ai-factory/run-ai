/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	apicommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	builderscommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KnativeBuilder struct {
	knativeService *servingv1.Service
	k8sClient      client.Client
}

func (kb KnativeBuilder) GetPreview(_ context.Context, workload *jobs.JobWorkload) ([]apicommon.Resource, error) {
	return []apicommon.Resource{{
		Kind: metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeServiceKind,
		},
		Name:      workload.GetChildName(),
		Namespace: workload.Namespace,
	}}, nil
}

func NewKnativeBuilder(k8sClient client.Client) builderscommon.ComputeBuilder {
	return KnativeBuilder{
		knativeService: &servingv1.Service{},
		k8sClient:      k8sClient,
	}
}

func (kb KnativeBuilder) Build(ctx context.Context, jWorkload *jobs.JobWorkload, options *builderscommon.ComputeBuilderOptions) (*builderscommon.ComputeResources, error) {
	if options == nil {
		options = &builderscommon.ComputeBuilderOptions{}
	}

	kb.knativeService = &servingv1.Service{}
	workload := infv2.InferenceWorkload{}

	// Copy the fields from the job workload to the inference workload
	workload.TypeMeta = jWorkload.TypeMeta
	workload.TypeMeta.Kind = constants.KindInferenceWorkload

	// Copy the metadata from the job workload to the inference workload
	jWorkload.ObjectMeta.DeepCopyInto(&workload.ObjectMeta)

	// Copy the spec from the job workload to the inference workload
	jWorkload.Spec.WorkloadFields.DeepCopyInto(&workload.Spec.WorkloadFields)
	jWorkload.Spec.ResourceAllocationFields.DeepCopyInto(&workload.Spec.ResourceAllocationFields)
	jWorkload.Spec.RunaiSchedulerFields.DeepCopyInto(&workload.Spec.RunaiSchedulerFields)
	jWorkload.Spec.InferenceFields.DeepCopyInto(&workload.Spec.InferenceFields)
	jWorkload.Spec.SchedulingFields.DeepCopyInto(&workload.Spec.SchedulingFields)

	// Copy the status from the job workload to the inference workload
	err := kb.buildObjectMeta(ctx, &workload)
	if err != nil {
		return nil, err
	}

	// Build the spec
	if err = kb.buildSpec(ctx, &workload); err != nil {
		return nil, err
	}

	// Set the owner reference
	if options.Owner != nil {
		err = controllerutils.SetControllerReference(options.Owner, kb.knativeService, kb.k8sClient.Scheme())
		if err != nil {
			return nil, err
		}
	}
	return builderscommon.NewComputeResourcesFromJob(kb.knativeService), nil
}

func (kb KnativeBuilder) buildObjectMeta(ctx context.Context, workload *infv2.InferenceWorkload) error {
	om := &kb.knativeService.ObjectMeta

	if om.Annotations == nil {
		om.Annotations = map[string]string{}
	}

	if om.Labels == nil {
		om.Labels = map[string]string{}
	}

	om.Name = workload.GetChildName()

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(om, &workload.Spec.ContainerFields)

	om.Namespace = workload.Namespace
	om.CreationTimestamp = metav1.Time{Time: time.Now()}

	om.Labels[constants.LabelWorkloadKind] = workload.Kind
	om.Labels[constants.LabelWorkloadName] = workload.Name
	projectName, err := workloadproject.GetProjectFromContext(ctx)
	if err != nil {
		return err
	}
	om.Labels[constants.RunaiProjectLabel] = projectName
	if workload.Spec.RoutingAndAuthorizationFields.IsPrivateServiceUrl.Deref() {
		om.Labels[constants.LabelKnativeServiceUrlVisibility] = "cluster-local"
	}

	om.Annotations[constants.AnnotationWorkloadApiVersion] = workload.APIVersion
	return nil
}

func (kb KnativeBuilder) buildSpec(ctx context.Context, workload *infv2.InferenceWorkload) error {
	resourceRequirements, resourcesEnvVars, err := kb.createResourceRequirements(ctx, workload)
	if err != nil {
		return err
	}

	objectNaming := &metav1.ObjectMeta{
		Name: workload.Name,
	}

	// Command
	commandStr := workloadbuilders.PatchTemplateVariables(workload.Spec.Command.Deref(), &workload.Spec.ContainerFields, objectNaming, ctx)
	command, err := workloadbuilders.ParseCommandLine(commandStr)
	if err != nil {
		return err
	}

	// Arguments
	argsStr := workloadbuilders.PatchTemplateVariables(workload.Spec.Arguments.Deref(), &workload.Spec.ContainerFields, objectNaming, ctx)
	args, err := workloadbuilders.ParseCommandLine(argsStr)
	if err != nil {
		return err
	}

	tolerations, err := workloadbuilders.GetTolerations(ctx, &workload.Spec.ContainerFields)
	if err != nil {
		return err
	}

	podSpecVolumes, bypassedVolumes, err := kb.getCategorizedVolumes(ctx, workload)
	if err != nil {
		return err
	}

	volumesAnnotationValue, mountsAnnotationValue, err := createPatchAnnotationsValues(bypassedVolumes)
	if err != nil {
		return err
	}

	revisionTemplateAnnotations, err := createRevisionTemplateAnnotations(ctx, workload.Spec)
	if err != nil {
		return err
	}

	if volumesAnnotationValue != "" && mountsAnnotationValue != "" {
		revisionTemplateAnnotations[constants.ValidationBypassAnnotationVolumeKey] = volumesAnnotationValue
		revisionTemplateAnnotations[constants.ValidationBypassAnnotationVolumeMountKey] = mountsAnnotationValue
	}

	revisionTemplateLabels, err := createRevisionTemplateLabels(ctx, kb.k8sClient, workload)
	if err != nil {
		return err
	}

	// Copy NIM label from workload Spec.Labels (DictionaryField)
	CopyNimLabel(workload.Spec.Labels, revisionTemplateLabels)

	envs := kb.createEnvironmentVariables(ctx, workload)
	envs = append(envs, resourcesEnvVars...)

	sc := createContainerSecurityContext(&workload.Spec.ContainerFields, nil)
	psc := createPodSecurityContext(&workload.Spec.ContainerFields)

	containers, err := kb.createContainers(workload, resourceRequirements, envs, command, args, sc, podSpecVolumes.Mounts)
	if err != nil {
		return err
	}

	podSpec := corev1.PodSpec{
		EnableServiceLinks: ptr.To(false),
		SchedulerName:      constants.RunaiScheduler,
		Containers:         containers,
		SecurityContext:    psc,
		Volumes:            podSpecVolumes.Volumes,
		Tolerations:        tolerations,
	}

	// Only add NodeSelector if NodeType is provided
	if workload.Spec.NodeType.Deref() != "" {
		podSpec.NodeSelector = map[string]string{
			constants.NodeTypeKey: workload.Spec.NodeType.Deref(),
		}
	}

	var timeoutSeconds *int64
	workloadTimeoutSeconds := workload.Spec.RequestTimeoutSeconds.Deref()
	if workloadTimeoutSeconds != 0 {
		timeoutSeconds = ptr.To(int64(workloadTimeoutSeconds))
	}

	kb.knativeService.Spec = servingv1.ServiceSpec{
		ConfigurationSpec: servingv1.ConfigurationSpec{
			Template: servingv1.RevisionTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: revisionTemplateAnnotations,
					Labels:      revisionTemplateLabels,
				},
				Spec: servingv1.RevisionSpec{
					TimeoutSeconds:       timeoutSeconds,
					PodSpec:              podSpec,
					ContainerConcurrency: getConcurrencyHardLimit(workload.Spec.ConcurrencyHardLimit),
				},
			},
		},
	}

	err = nodepoolbuilders.CreateNodeAffinityForNodePools(
		ctx,
		kb.k8sClient,
		workload.Spec.NodePools.Deref(),
		workload.Spec.NodePool.Deref(),
		&kb.knativeService.Spec.ConfigurationSpec.Template.Spec.PodSpec.Affinity,
	)
	if err != nil {
		return err
	}

	// if the user provided node affinity, we need to merge it with the one we created
	workloadbuilders.MergeNodeAffinity(&workload.Spec.NodeAffinityFields, &kb.knativeService.Spec.ConfigurationSpec.Template.Spec.PodSpec.Affinity)

	err = workloadbuilders.CreatePodAffinity(
		workload.Spec.PodAffinity.Deref(),
		workload.Spec.PodAffinityTopology.Deref(),
		workload.Spec.PodAffinitySchedulingRule.Deref(),
		&kb.knativeService.Spec.ConfigurationSpec.Template.Spec.PodSpec.Affinity,
		workload.GetChildName(),
		false,
	)
	if err != nil {
		return err
	}

	if err := createGitSyncInitContainer(ctx, &workload.Spec.ContainerFields, &workload.Spec.StorageFields, &kb.knativeService.Spec.Template.Spec); err != nil {
		return err
	}

	servingPort, hasServingPortAuth, containerWithPortIndex, err := handleNetworkingFields(
		ctx, workload.GetChildName(), &workload.Spec.NetworkingFields,
		&workload.Spec.RoutingAndAuthorizationFields,
		&kb.knativeService.Spec.ConfigurationSpec.Template.Spec.PodSpec,
		workload.Kind, &workload.ObjectMeta,
		workload.Spec.RequestTimeoutSeconds.Deref(),
	)
	if err != nil {
		return err
	}

	// using containerWithPortIndex in order to create the probes on the correct container:
	// the single container where port is declared, either the user container or the serving port
	// authorization container, if such.
	err = createProbes(ctx, &kb.knativeService.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[containerWithPortIndex], &workload.Spec.ContainerFields, servingPort, hasServingPortAuth)
	if err != nil {
		return err
	}

	workloadKindInitial, err := builderscommon.GetWorkloadInitials(workload.Kind)
	if err != nil {
		return err
	}
	err = workloadbuilders.BuildImagePullSecrets(
		ctx, workload.Spec.ImagePullSecrets, kb.k8sClient, workload.Namespace,
		&kb.knativeService.Spec.ConfigurationSpec.Template.Spec.ImagePullSecrets,
		workload.Name, string(workloadKindInitial),
	)
	if err != nil {
		return err
	}

	return nil
}

func (kb KnativeBuilder) createEnvironmentVariables(ctx context.Context, workload *infv2.InferenceWorkload) []corev1.EnvVar {

	objectNaming := &metav1.ObjectMeta{
		Name: workload.Name,
	}

	envs := make([]corev1.EnvVar, 0)
	for _, name := range field_types.SortedDictionaryMapKeys(workload.Spec.Environment.Deref()) {
		value := (workload.Spec.Environment.Deref())[name]
		if !ptr.Deref(value.Deleted, false) {
			envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, name, value.Value, &workload.Spec.ContainerFields, objectNaming, workload.GetChildName(), InferenceWorkloadInitial))
		}
	}

	if workloadbuilders.NeedToSetupRunaiHome(&workload.Spec.ContainerFields, false) {
		envs = append(envs, corev1.EnvVar{
			Name:  constants.HomeDirEnvName,
			Value: constants.RunaiDefaultHomeDir,
		})
	}

	return envs
}

func (kb KnativeBuilder) createResourceRequirements(ctx context.Context, workload *infv2.InferenceWorkload) (*corev1.ResourceRequirements, []corev1.EnvVar, error) {
	resourceRequirements, envVars, err := workloadbuilders.GetResourcesReqAndLimits(ctx, &workload.Spec.ResourceAllocationFields)
	if err != nil {
		return nil, nil, err
	}

	if resourceRequirements == nil {
		resourceRequirements = &corev1.ResourceRequirements{}
	}
	return resourceRequirements, envVars, nil
}

func (kb KnativeBuilder) createContainers(workload *infv2.InferenceWorkload, resourceRequirements *corev1.ResourceRequirements,
	envs []corev1.EnvVar, command, args []string, sc *corev1.SecurityContext, mounts []corev1.VolumeMount) ([]corev1.Container, error) {

	kbms := KnativeBuilderModelSwap{}
	if kbms.getModelSwapAnnotation(workload.Spec.Annotations) != "" {
		return kbms.createContainersForModelSwap(workload, resourceRequirements, envs, command, args, sc, mounts)
	}

	return []corev1.Container{
		{
			Image:           workload.Spec.Image.Deref(),
			ImagePullPolicy: corev1.PullPolicy(workload.Spec.ImagePullPolicy.Deref()),
			WorkingDir:      workload.Spec.WorkingDir.Deref(),
			Resources:       *resourceRequirements,
			Command:         command,
			Args:            args,
			SecurityContext: sc,
			VolumeMounts:    mounts,
			Env:             envs,
		},
	}, nil
}

func createProbes(ctx context.Context, container *corev1.Container, containerFields *field_groups.ContainerFields, servingPort *corev1.ContainerPort, hasServingPortAuth bool) error {
	// if the user didn't provide a readiness probe, and we have authorization on the serving port,
	// then we need to create a default tcp readiness probe, to replace the knative default one.
	// otherwise, the default knative readiness probe will be on the authorization port,
	// which effectively causes the workload to be considered ready immediately,
	// before the serving port is actually ready.

	if containerFields.ReadinessProbe == nil && hasServingPortAuth {
		err := workloadbuilders.CreateTcpReadinessProbe([]corev1.ContainerPort{*servingPort}, container)
		if err != nil {
			return err
		}
	} else {
		err := workloadbuilders.CreateProbes(ctx, containerFields, container, false)
		if err != nil {
			return err
		}
	}

	return nil
}
