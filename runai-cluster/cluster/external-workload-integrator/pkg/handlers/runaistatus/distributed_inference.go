package runaistatus

import (
	"sort"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	commondistributedinference "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/distributedinference"
)

type DistributedInferenceInput = commondistributedinference.Input

type DistributedInferenceResult struct {
	Phase         v1alpha1.ExternalWorkloadPhase
	Message       string
	Replicas      int32
	ReadyReplicas int32
}

func ComputeDistributedInferenceStatus(input DistributedInferenceInput) DistributedInferenceResult {
	input.MissingResources = append([]string(nil), input.MissingResources...)
	sort.Strings(input.MissingResources)
	result := commondistributedinference.Calculate(input)

	phase := v1alpha1.ExternalWorkloadPhase(result.Phase)
	if result.Phase == statuscontract.Ready {
		phase = v1alpha1.ExWRunningPhase
	}
	return DistributedInferenceResult{
		Phase:         phase,
		Message:       result.Message,
		Replicas:      result.Replicas,
		ReadyReplicas: result.ReadyReplicas,
	}
}
