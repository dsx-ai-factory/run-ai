package mnnvlutils

import (
	"fmt"
	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

func ValidateResources(w *jobs.JobWorkload) error {
	if !gpu_validator.HasRequestedGpus(&w.Spec.ResourceAllocationFields) && (w.MasterSpec == nil || !gpu_validator.HasRequestedGpus(&w.MasterSpec.ResourceAllocationFields)) {
		return fmt.Errorf("no GPU resources requested")
	}

	// Check if either spec or master spec is using node pools
	if !jobs.IsUsingNodePools(&w.Spec) && !jobs.IsUsingNodePools(w.MasterSpec) {
		return fmt.Errorf("empty node pools list")
	}

	return nil
}
