package pod_mutators

import (
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
)

const (
	kubevirtObjAPIVersion         = "kubevirt.io/v1"
	kubevirtAntiNodeAffinityKey   = "kubevirt.io/schedulable"
	nodeTypeAffinityExpressionKey = "run.ai/type"
)

func (pm *PodMutator) mutateNodeAffinityForNodeType(pod *v1.Pod, request *v1beta1.AdmissionRequest) error {
	namespace := getDestinationNamespace(pod, request)
	projectName := utils.GetDesiredProjectName(pm.client, namespace, pod.Spec.Containers)

	if isRunaiPodJob(pod) {
		err := pm.setRunaiNodeAffinityForNodeType(pod, pod.Kind, projectName)
		if err != nil {
			return err
		}
	}

	pm.setKubevirtNodeAntiAffinity(pod)
	return nil
}

func (pm *PodMutator) setRunaiNodeAffinityForNodeType(pod *v1.Pod, _ string, projectName string) error {
	interactive := utils.IsInteractiveByPriority(pm.client.GetCachedReader(),
		pod.Labels, pod.Spec.PriorityClassName)
	project, err := utils.GetProjectResource(pm.client, projectName)
	if err != nil {
		return err
	}

	setNodeAffinityForNodeType(pod, project, interactive)

	return nil
}

func getPriorityClassName(pod *v1.Pod) string {
	priorityClassName, found := pod.Labels["priorityClassName"]
	if !found {
		return ""
	}
	return priorityClassName
}

func setNodeAffinityForNodeType(pod *v1.Pod, project *unstructured.Unstructured, interactive bool) {
	// The node types moved from Project.spec to an annotation with the kai.resources CRD.
	nodeAffinityTypes := utils.NodeAffinityTypesFromAnnotation(project, interactive)

	if len(nodeAffinityTypes) == 0 {
		return
	}

	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      nodeTypeAffinityExpressionKey,
		Operator: v1.NodeSelectorOpIn,
		Values:   nodeAffinityTypes,
	}
	patch.AddNodeAffinityToPodSpec(&pod.Spec, []v1.NodeSelectorRequirement{matchExpressionToAdd})
}

func (pm *PodMutator) setKubevirtNodeAntiAffinity(pod *v1.Pod) {
	if !pm.params.IsKubevirtEnabledCluster {
		return
	}

	ownderReferences := pod.ObjectMeta.OwnerReferences
	if len(ownderReferences) > 0 && ownderReferences[0].APIVersion == kubevirtObjAPIVersion {
		return
	}
	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      kubevirtAntiNodeAffinityKey,
		Operator: v1.NodeSelectorOpNotIn,
		Values:   []string{"true"},
	}
	patch.AddNodeAffinityToPodSpec(&pod.Spec, []v1.NodeSelectorRequirement{matchExpressionToAdd})
}
