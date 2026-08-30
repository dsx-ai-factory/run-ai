package utils

import (
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infw2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PatchWorkloadPolicy(workload client.Object, patches map[api.SubresourceComponent][]byte) error {
	switch workloadType := workload.(type) {
	case *twv2.TrainingWorkload:
		train := workloadType
		return patchTrainingWorkload(train, patches)
	case *iwv2.InteractiveWorkload:
		interactive := workloadType
		return patchInteractiveWorkload(interactive, patches)
	case *dwv2.DistributedWorkload:
		distributed := workloadType
		return patchDistributedWorkload(distributed, patches)
	case *infw2.InferenceWorkload:
		inference := workloadType
		return patchInferenceWorkload(inference, patches)
	default:
		return fmt.Errorf("unknown workload type %v", workload)

	}
}

func patchDistributedWorkload(distributed *dwv2.DistributedWorkload, patches map[api.SubresourceComponent][]byte) error {
	if workerSpec, ok := patches[api.WorkerSpecSubResource]; ok {

		err := json.Unmarshal(workerSpec, &distributed.Spec)
		if err != nil {
			return err
		}
	}
	if masterSpec, ok := patches[api.MasterSpecSubResource]; ok {
		if distributed.MasterSpec == nil {
			distributed.MasterSpec = &v2alpha1.DistributedMasterSpec{}
		}
		err := json.Unmarshal(masterSpec, &distributed.MasterSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

func patchInteractiveWorkload(interactive *iwv2.InteractiveWorkload, patches map[api.SubresourceComponent][]byte) error {
	if spec, ok := patches[api.SpecSubResource]; ok {
		err := json.Unmarshal(spec, &interactive.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func patchTrainingWorkload(workload *twv2.TrainingWorkload, patches map[api.SubresourceComponent][]byte) error {
	if spec, ok := patches[api.SpecSubResource]; ok {
		err := json.Unmarshal(spec, &workload.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func patchInferenceWorkload(workload *infw2.InferenceWorkload, patches map[api.SubresourceComponent][]byte) error {
	if spec, ok := patches[api.SpecSubResource]; ok {
		err := json.Unmarshal(spec, &workload.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}
