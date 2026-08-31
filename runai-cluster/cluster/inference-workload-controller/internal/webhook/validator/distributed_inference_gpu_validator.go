/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"

	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
)

// DistributedInferenceGpuValidator validates DistributedInferenceWorkload by calling validation on workload.LeaderSpec and workload.WorkerSpec
type DistributedInferenceGpuValidator interface {
	Validate(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error
}

// distributedInferenceGpuValidator implements DistributedInferenceGpuValidator
type distributedInferenceGpuValidator struct {
	gpu_validator.GpuValidator
}

// NewDistributedInferenceGpuValidator creates a new DistributedInferenceGpuValidator
func NewDistributedInferenceGpuValidator() DistributedInferenceGpuValidator {
	return &distributedInferenceGpuValidator{GpuValidator: gpu_validator.NewGpuValidator()}
}

// DistributedInferenceGpuValidator implementation
func (self *distributedInferenceGpuValidator) Validate(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	err := self.ValidateWorkloadGPU(ctx, &workload.LeaderSpec.ResourceAllocationFields)
	if err != nil {
		return err
	}

	if workload.WorkerSpec != nil {
		err = self.ValidateWorkloadGPU(ctx, &workload.WorkerSpec.ResourceAllocationFields)
		if err != nil {
			return err
		}
	}

	return nil
}
