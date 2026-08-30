package kartaparser

import (
	"context"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	riresource "github.com/run-ai/karta/pkg/resource"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	engineResources "github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EWConverter struct {
	cachedReader client.Reader
	draParser    *dra.DRAParser
}

func NewEWConverter(cachedClient client.Client, draParser *dra.DRAParser) *EWConverter {
	return &EWConverter{
		cachedReader: cachedClient,
		draParser:    draParser,
	}
}

func (h *EWConverter) ConvertComponentToExternalWorkloadSpec(ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind,
	riComponent *riresource.Component, isRoot bool, externalWorkload *v1alpha1.ExternalWorkload) error {

	if riComponent.HasPodDefinition() {
		extractedInstances, err := riComponent.GetExtractedInstances(ctx)
		if err != nil {
			return err
		}

		for _, extractedInstance := range extractedInstances {
			h.fillComponentSpecAndReplicasFields(ctx, topOwnerObjectMeta, topOwnerGroupKind, riComponent.Name(), extractedInstance, externalWorkload)
		}
	}

	if isRoot {
		status, err := riComponent.GetStatus(ctx)
		if err != nil {
			return err
		}
		if status == nil {
			status = &riresource.Status{}
		}
		h.fillStatus(status, externalWorkload)

	} else if externalWorkload.Status.ObservedStatus.Phase == "" {
		externalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWUnknownPhase
	}

	return nil
}

func (h *EWConverter) fillComponentSpecAndReplicasFields(
	ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind,
	componentName string, extractedInstance riresource.ExtractedInstance, externalWorkload *v1alpha1.ExternalWorkload,
) {
	previousResources := *externalWorkload.Spec.RequestedResources.DeepCopy()
	externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{}

	h.fillExternalWorkloadDataPodSpecValues(ctx, topOwnerObjectMeta, topOwnerGroupKind, componentName, extractedInstance, externalWorkload)

	requestedResources := h.fillReplicasAndMultiplyResources(extractedInstance, externalWorkload)
	externalWorkload.Spec.RequestedResources = engineResources.SumResourceRequirements(previousResources, requestedResources)
}

func (h *EWConverter) fillExternalWorkloadDataPodSpecValues(
	ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind,
	componentName string, summary riresource.ExtractedInstance, externalWorkload *v1alpha1.ExternalWorkload,
) {
	podTemplate := createPodTemplate(summary, externalWorkload)

	var specName *string = nil
	if componentName != "" {
		specName = ptr.To(componentName)
	}
	commonparsers.FillExternalWorkloadDataFromPodTemplateSpecIfEmpty(ctx, topOwnerObjectMeta, topOwnerGroupKind,
		podTemplate, specName, h.cachedReader, h.draParser, externalWorkload)
}

func createPodTemplate(summary riresource.ExtractedInstance, externalWorkload *v1alpha1.ExternalWorkload) *corev1.PodTemplateSpec {
	podTemplate := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}

	if summary.PodTemplateSpec != nil {
		podTemplate = summary.PodTemplateSpec.DeepCopy()
	} else if summary.PodSpec != nil {
		podTemplate.Spec = *summary.PodSpec.DeepCopy()
	}

	if summary.Metadata != nil {
		if summary.Metadata.Labels != nil {
			extendMap(summary.Metadata.Labels, podTemplate.Labels)
		}
		if summary.Metadata.Annotations != nil {
			extendMap(summary.Metadata.Annotations, podTemplate.Annotations)
		}
	}

	if summary.FragmentedPodSpec != nil {
		frag := summary.FragmentedPodSpec
		if frag.Annotations != nil {
			extendMap(frag.Annotations, podTemplate.Annotations)
		}
		if frag.Labels != nil {
			extendMap(frag.Labels, podTemplate.Labels)
		}
		if (frag.Containers != nil || frag.Container != nil) && len(podTemplate.Spec.Containers) == 0 {
			containers := getContainers(frag.Containers, frag.Container)
			podTemplate.Spec.Containers = containers
		}
		if frag.NodeAffinity != nil && podTemplate.Spec.Affinity == nil {
			podTemplate.Spec.Affinity = &corev1.Affinity{NodeAffinity: frag.NodeAffinity}
		}
		if frag.PriorityClassName != "" && podTemplate.Spec.PriorityClassName == "" {
			podTemplate.Spec.PriorityClassName = frag.PriorityClassName
		}
		if externalWorkload.Spec.ImageName == "" && frag.Image != "" {
			externalWorkload.Spec.ImageName = frag.Image
		}
		if len(frag.ResourceClaims) > 0 && len(podTemplate.Spec.ResourceClaims) == 0 {
			podTemplate.Spec.ResourceClaims = frag.ResourceClaims
		}
		if frag.Resources != nil {
			if len(podTemplate.Spec.Containers) == 0 {
				podTemplate.Spec.Containers = []corev1.Container{}
			}
			podTemplate.Spec.Containers = append(podTemplate.Spec.Containers, corev1.Container{Resources: *frag.Resources.DeepCopy()})
		}
	}
	return podTemplate
}

func (h *EWConverter) fillReplicasAndMultiplyResources(extractedInstance riresource.ExtractedInstance, externalWorkload *v1alpha1.ExternalWorkload) corev1.ResourceRequirements {

	scale := ptr.Deref(extractedInstance.Scale, riresource.Scale{})
	sumReplicas(&externalWorkload.Spec.RequestedPods.Replicas, scale.Replicas)
	if externalWorkload.Spec.RequestedPods.Replicas == nil {
		externalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(1))
	}
	sumReplicas(&externalWorkload.Spec.RequestedPods.MinReplicas, scale.MinReplicas)
	sumReplicas(&externalWorkload.Spec.RequestedPods.MaxReplicas, scale.MaxReplicas)

	replicas := utils.DerefInt32(scale.Replicas, 0)

	if replicas == 0 {
		replicas = utils.DerefInt32(scale.MinReplicas, 1)
	}

	requestedResources := commonparsers.MultiplyResources(replicas, externalWorkload.Spec.RequestedResources)
	return requestedResources
}

func (h *EWConverter) fillStatus(riStatus *riresource.Status, externalWorkload *v1alpha1.ExternalWorkload) {
	if riStatus == nil {
		return
	}

	if riStatus.Phase != nil {
		externalWorkload.Status.ObservedStatus.K8sStatus = *riStatus.Phase
	}

	externalWorkload.Status.ObservedStatus.Phase = h.mapResourceStatusToEwPhase(riStatus.MatchedStatuses)

	conditions := make([]v1alpha1.ObservedCondition, 0, len(riStatus.Conditions))
	for _, cond := range riStatus.Conditions {
		status := ""
		if cond.Status != nil {
			status = *cond.Status
		}
		conditions = append(conditions, v1alpha1.ObservedCondition{
			Type:    cond.Type,
			Status:  parseConditionStatus(status),
			Message: cond.Message,
		})
	}
	externalWorkload.Status.ObservedStatus.Conditions = conditions
}

func (h *EWConverter) mapResourceStatusToEwPhase(statuses []kartav1alpha1.ResourceStatus) v1alpha1.ExternalWorkloadPhase {
	statusesSet := make(map[kartav1alpha1.ResourceStatus]bool)
	for _, status := range statuses {
		statusesSet[status] = true
	}
	if statusesSet[kartav1alpha1.RunningStatus] {
		return v1alpha1.ExWRunningPhase
	}
	if statusesSet[kartav1alpha1.FailedStatus] {
		return v1alpha1.ExWFailedPhase
	}
	if statusesSet[kartav1alpha1.CompletedStatus] {
		return v1alpha1.ExWCompletedPhase
	}
	if statusesSet[kartav1alpha1.InitializingStatus] {
		return v1alpha1.ExWInitializingPhase
	}
	return v1alpha1.ExWUnknownPhase
}

func (h *EWConverter) FillDefaultsAfterAllParsing(ctx context.Context, topOwnerObjectMeta metav1.ObjectMeta, topOwnerGroupKind schema.GroupKind, externalWorkload *v1alpha1.ExternalWorkload) {
	if externalWorkload.Spec.RequestedPods.Replicas == nil {
		externalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(1))
	}

	if externalWorkload.Spec.PriorityClassName == "" || externalWorkload.Spec.Preemptibility == "" {
		priorityClassName, preemptibility := commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedReader,
			topOwnerObjectMeta.Labels, topOwnerGroupKind, "")
		if externalWorkload.Spec.PriorityClassName == "" {
			externalWorkload.Spec.PriorityClassName = priorityClassName
		}
		if externalWorkload.Spec.Preemptibility == "" {
			externalWorkload.Spec.Preemptibility = preemptibility
		}
	}
}

func parseConditionStatus(status string) corev1.ConditionStatus {
	switch status {
	case string(corev1.ConditionTrue):
		return corev1.ConditionTrue
	case string(corev1.ConditionFalse):
		return corev1.ConditionFalse
	case string(corev1.ConditionUnknown):
		return corev1.ConditionUnknown
	default:
		return corev1.ConditionUnknown
	}
}

func getContainers(containers []corev1.Container, container *corev1.Container) []corev1.Container {
	if len(containers) > 0 {
		return containers
	}
	if container != nil {
		return []corev1.Container{*container.DeepCopy()}
	}
	return []corev1.Container{}
}

func sumReplicas(source **int32, toAdd *int32) {
	if source == nil || toAdd == nil {
		return
	}

	toAddR := *toAdd

	if *source == nil {
		*source = ptr.To(toAddR)
	} else {
		*source = ptr.To(utils.DerefInt32(*source, 0) + toAddR)
	}
}

func extendMap(from, to map[string]string) {
	if from == nil || to == nil {
		return
	}

	for key, value := range from {
		if _, ok := to[key]; !ok {
			to[key] = value
		}
	}
}
