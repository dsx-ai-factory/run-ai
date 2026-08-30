package mnnvl

import (
	"context"
	"errors"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	affinityutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/affinity"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	kartautils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/karta"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
)

// ErrMissingLabelPath is returned when the plugin stamps the workload-name label but
// a Karta FragmentedPodSpec has no labelsPath to write it back through.
var ErrMissingLabelPath = errors.New("MNNVL Plugin using Karta with FragmentedPodSpec requires " +
	"labelsPath in order to stamp the workload-name label")

type mnnvlPodSpecFields struct {
	NodeAffinity   *corev1.NodeAffinity
	PodAffinity    *corev1.PodAffinity
	ResourceClaims *[]corev1.PodResourceClaim
	Container      *corev1.Container
}

// MutateResource applies MNNVL to the resource's GPU pods: the DRA resource claim,
// node affinity, the pod affinity, and (unless disabled) the workload-name label.
func (p *Plugin) MutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	config := extractMnnvlConfig(runner)
	if config == nil || config.Mode != dromosv1alpha1.MnnvlModeRequired {
		return nil
	}

	if err := validateKartaSupportMnnvl(karta, !p.opts.DisableWorkloadLabelStamp); err != nil {
		return state.AddPluginIssue(ctx, p, resource, fmt.Errorf("given karta %s does not support MNNVL: %w", karta.Name, err), true, "", issue.Network)
	}

	if len(config.GpuNetworkAccelerationLabelKeys) == 0 {
		return state.AddPluginIssue(ctx, p, resource, errors.New("GPU network acceleration label keys are not set"), true, "", issue.Compute)
	}

	podAffinity := p.opts.PodAffinityProvider(runner, resource)

	mutate := func(instance kartaresource.ExtractedInstance) error {
		p.mutateInstance(runner, config.GpuNetworkAccelerationLabelKeys, instance, podAffinity)
		return nil
	}
	if err := kartautils.MutatePodInstances(ctx, karta, resource, mutate); err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "", issue.Internal)
	}

	return nil
}

// defaultPodAffinity is the plugin's built-in PodAffinityProvider, installed by
// NewPlugin when the consumer supplies none: group the workload's GPU pods by
// WorkloadNameLabel.
func defaultPodAffinity(runner *dromosv1alpha1.WorkloadRunner, _ *unstructured.Unstructured) *corev1.PodAffinity {
	config := extractMnnvlConfig(runner)
	if config == nil {
		return nil
	}
	return BuildGroupingAffinity(WorkloadNameLabel, runner.Name, config.GpuNetworkAccelerationLabelKeys)
}

func validateKartaSupportMnnvl(karta *kartav1alpha1.Karta, stampWorkloadLabel bool) error {
	hasSpecDefinition, err := validateComponentDefinition(&karta.Spec.StructureDefinition.RootComponent, stampWorkloadLabel)
	if err != nil {
		return fmt.Errorf("root component %s: %w", karta.Spec.StructureDefinition.RootComponent.Name, err)
	}

	for _, componentDef := range karta.Spec.StructureDefinition.ChildComponents {
		componentHasSpecDefinition, err := validateComponentDefinition(&componentDef, stampWorkloadLabel)
		if err != nil {
			return fmt.Errorf("component %s: %w", componentDef.Name, err)
		}
		hasSpecDefinition = hasSpecDefinition || componentHasSpecDefinition
	}
	if !hasSpecDefinition {
		return errors.New("karta does not have any component with a Spec definition")
	}
	return nil
}

func validateComponentDefinition(def *kartav1alpha1.ComponentDefinition, stampWorkloadLabel bool) (hasSpecDefinition bool, error error) {
	if def.SpecDefinition == nil {
		return false, nil
	}

	if def.SpecDefinition.PodTemplateSpecPath != nil || def.SpecDefinition.PodSpecPath != nil {
		return true, nil
	}

	fragDef := def.SpecDefinition.FragmentedPodSpecDefinition
	if fragDef == nil {
		// PodTemplateSpec or PodSpec is used, so they have all the necessary fields.
		return false, nil
	}
	hasContainersPath := ptr.Deref(fragDef.ContainersPath, "") != "" || ptr.Deref(fragDef.ContainerPath, "") != ""
	hasResourceClaimsPath := ptr.Deref(fragDef.ResourceClaimsPath, "") != ""

	if !hasContainersPath && !hasResourceClaimsPath {
		return false, errors.New("MNNVL Plugin using Karta with FragmentedPodSpec requires " +
			"containersPath or containerPath or resourceClaimsPath in order to set the resource claim")
	}

	if ptr.Deref(fragDef.NodeAffinityPath, "") == "" {
		return false, errors.New("MNNVL Plugin using Karta with FragmentedPodSpec requires " +
			"nodeAffinityPath in order to set the node affinity")
	}

	if ptr.Deref(fragDef.PodAffinityPath, "") == "" {
		return false, errors.New("MNNVL Plugin using Karta with FragmentedPodSpec requires " +
			"podAffinityPath in order to set the pod affinity")
	}

	if stampWorkloadLabel && ptr.Deref(fragDef.LabelsPath, "") == "" {
		return false, ErrMissingLabelPath
	}

	return true, nil
}

// mutateInstance stamps the workload-name label on a GPU pod definition (when
// stamping is enabled and the pod shape can carry it) and writes the MNNVL
// resource claim, node affinity, and pod affinity. Non-GPU instances are skipped.
func (p *Plugin) mutateInstance(
	runner *dromosv1alpha1.WorkloadRunner,
	gpuNetworkAccelerationLabelKeys []string,
	instance kartaresource.ExtractedInstance,
	podAffinity *corev1.PodAffinity,
) {
	if !kartautils.HasGPU(instance) {
		return
	}

	if !p.opts.DisableWorkloadLabelStamp && kartautils.CanStampPodLabels(instance) {
		stampWorkloadName(instance, runner.Name)
	}

	switch {
	case instance.PodTemplateSpec != nil:
		mutatePodSpec(&instance.PodTemplateSpec.Spec, runner, gpuNetworkAccelerationLabelKeys, podAffinity)
	case instance.PodSpec != nil:
		mutatePodSpec(instance.PodSpec, runner, gpuNetworkAccelerationLabelKeys, podAffinity)
	case instance.FragmentedPodSpec != nil:
		mutateFragmentedPodSpec(instance.FragmentedPodSpec, runner, gpuNetworkAccelerationLabelKeys, podAffinity)
	}
}

func mutatePodSpec(podSpec *corev1.PodSpec, runner *dromosv1alpha1.WorkloadRunner, gpuNetworkAccelerationLabelKeys []string, podAffinity *corev1.PodAffinity) {
	container := getContainerWithGPUResources(podSpec.Containers...)
	if container == nil {
		return
	}
	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{}
	}
	if podAffinity != nil && podSpec.Affinity.PodAffinity == nil {
		podSpec.Affinity.PodAffinity = &corev1.PodAffinity{}
	}
	if podSpec.Affinity.NodeAffinity == nil {
		podSpec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	mutatePod(&mnnvlPodSpecFields{
		NodeAffinity:   podSpec.Affinity.NodeAffinity,
		PodAffinity:    podSpec.Affinity.PodAffinity,
		ResourceClaims: &podSpec.ResourceClaims,
		Container:      container,
	}, runner, gpuNetworkAccelerationLabelKeys, podAffinity)
}

func mutateFragmentedPodSpec(frag *kartaresource.FragmentedPodSpec, runner *dromosv1alpha1.WorkloadRunner, gpuNetworkAccelerationLabelKeys []string, podAffinity *corev1.PodAffinity) {
	if podAffinity != nil && frag.PodAffinity == nil {
		frag.PodAffinity = &corev1.PodAffinity{}
	}
	if frag.NodeAffinity == nil {
		frag.NodeAffinity = &corev1.NodeAffinity{}
	}

	mutatePod(&mnnvlPodSpecFields{
		NodeAffinity:   frag.NodeAffinity,
		PodAffinity:    frag.PodAffinity,
		ResourceClaims: &frag.ResourceClaims,
		Container:      getContainerWithGPUResources(frag.Containers...),
	}, runner, gpuNetworkAccelerationLabelKeys, podAffinity)
}

func mutatePod(fields *mnnvlPodSpecFields, runner *dromosv1alpha1.WorkloadRunner, gpuNetworkAccelerationLabelKeys []string, podAffinity *corev1.PodAffinity) {
	addResourceClaimTemplateToPod(workloadRunnerInitials, runner.Name, fields.ResourceClaims, fields.Container)
	if podAffinity != nil {
		mergePodAffinity(fields.PodAffinity, podAffinity)
	}
	addNodeAffinity(fields.NodeAffinity, gpuNetworkAccelerationLabelKeys)
}

func addNodeAffinity(nodeAffinity *corev1.NodeAffinity, gpuNetworkAccelerationLabelKeys []string) {
	expressions := make([]corev1.NodeSelectorRequirement, 0, len(gpuNetworkAccelerationLabelKeys))
	for _, labelKey := range gpuNetworkAccelerationLabelKeys {
		expressions = append(expressions, corev1.NodeSelectorRequirement{
			Key:      labelKey,
			Operator: corev1.NodeSelectorOpExists,
		})
	}
	nodeSelector := &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchExpressions: expressions,
			},
		},
	}

	nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = affinityutils.MergeRequired(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, nodeSelector)
}

func getContainerWithGPUResources(containers ...corev1.Container) *corev1.Container {
	_, index, _ := lo.FindIndexOf(containers, func(container corev1.Container) bool {
		return hasGPUResources(container.Resources)
	})
	if index == -1 {
		// We should mutate only pods with GPU resources
		return nil
	}
	return &containers[index]
}

// stampWorkloadName sets WorkloadNameLabel on the instance's pod template. Caller
// gates on CanStampPodLabels (a bare PodSpec has no pod-label target).
func stampWorkloadName(instance kartaresource.ExtractedInstance, workloadName string) {
	switch {
	case instance.PodTemplateSpec != nil:
		if instance.PodTemplateSpec.Labels == nil {
			instance.PodTemplateSpec.Labels = map[string]string{}
		}
		instance.PodTemplateSpec.Labels[WorkloadNameLabel] = workloadName
	case instance.FragmentedPodSpec != nil:
		if instance.FragmentedPodSpec.Labels == nil {
			instance.FragmentedPodSpec.Labels = map[string]string{}
		}
		instance.FragmentedPodSpec.Labels[WorkloadNameLabel] = workloadName
	}
}
