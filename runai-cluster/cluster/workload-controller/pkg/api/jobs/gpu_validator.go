/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

// this file should be in workload-controller/pkg/webhook/validators/gpu_validator.go,
// but it is in workload-controller/pkg/api/jobs/gpu_validator.go bc currently cycle import
// after refactor workload webhooks it will be moved to the right place
import (
	"context"

	gpu_validator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/gpu-validator"
)

// WorkloadGpuValidator validates JobWorkload by calling validation on workload.Spec and workload.MasterSpec
type WorkloadGpuValidator interface {
	Validate(ctx context.Context, workload *JobWorkload) error
}

// workloadGpuValidator implements WorkloadGpuValidator
type workloadGpuValidator struct {
	gpu_validator.GpuValidator
}

// NewWorkloadGpuValidator creates a new WorkloadGpuValidator
func NewWorkloadGpuValidator() WorkloadGpuValidator {
	return &workloadGpuValidator{GpuValidator: gpu_validator.NewGpuValidator()}
}

// WorkloadGpuValidator implementation
func (self *workloadGpuValidator) Validate(ctx context.Context, workload *JobWorkload) error {
	err := self.ValidateWorkloadGPU(ctx, &workload.Spec.ResourceAllocationFields)
	if err != nil {
		return err
	}
	if workload.MasterSpec != nil {
		err = self.ValidateWorkloadGPU(ctx, &workload.MasterSpec.ResourceAllocationFields)
		if err != nil {
			return err
		}
	}
	return nil
}
