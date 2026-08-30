package field_groups

import (
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

// +k8s:deepcopy-gen=true
type RunaiJobSpecificRules struct {

	// Indicates whether the job should be terminated, by the system, after it has been preempted.
	// Default to false.
	TerminateAfterPreemption *flds.BoolField `json:"terminateAfterPreemption,omitempty"`

	// Specifies the number of retries before marking a workload as failed.
	// Defaults to 6
	BackoffLimit *flds.IntField `json:"backoffLimit,omitempty"`
}
