package pod_mutators

import (
	"log"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
)

func (pm *PodMutator) mutateProjectNameLabel(pod *v1.Pod, request *v1beta1.AdmissionRequest) error {
	_, found := pod.Labels["project"]
	if found {
		return nil
	}

	namespace := getDestinationNamespace(pod, request)
	projectName := utils.GetDesiredProjectName(pm.client, namespace, pod.Spec.Containers)

	if projectName == "" {
		log.Printf("Warning: Could not identify the runai project related to this pod")
		return nil
	}

	patch.AddLabelToObjectMeta(&pod.ObjectMeta, "project", projectName)
	return nil
}
