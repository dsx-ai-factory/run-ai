package runaistatus

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	commoninference "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/inference"
)

type InferenceServiceCondition = commoninference.ServiceCondition
type InferenceServiceState = commoninference.ServiceState
type InferenceInput = commoninference.Input

func ComputeInferenceStatus(input InferenceInput) (v1alpha1.ExternalWorkloadPhase, string) {
	result := commoninference.Calculate(input)
	if result.Phase == statuscontract.Ready {
		return v1alpha1.ExWRunningPhase, result.Message
	}
	return v1alpha1.ExternalWorkloadPhase(result.Phase), result.Message
}
