package pod_controller_mutators

import (
	"k8s.io/api/admission/v1beta1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func getDestinationNamespace(podController *types.PodControllerCommon, request *v1beta1.AdmissionRequest) string {
	if podController.Namespace != "" {
		return podController.Namespace
	}
	return request.Namespace
}
