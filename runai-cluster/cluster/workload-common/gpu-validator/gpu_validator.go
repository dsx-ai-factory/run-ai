/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_validator

import (
	"context"
	"errors"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"k8s.io/utils/ptr"
)

const WholeGPUWithFractionalGPUError = "gpu does not support whole number with fractional part"

// GpuValidator validates ResourceAllocationFields
type GpuValidator interface {
	Validate(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error
	ValidateWorkloadGPU(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error
	ValidateWorkloadAnyGPU(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error
	DetermineGpuDevicesAndFraction(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) (int, *string, error)
	ValidateLimits(resourceFields *field_groups.ResourceAllocationFields, gpuFraction *string) error
}

// gpuValidator implements GpuValidator with validation logic
type gpuValidator struct{}

// NewGpuValidator creates a new GpuValidator
func NewGpuValidator() GpuValidator {
	return &gpuValidator{}
}

// GpuValidator implementation
func (self *gpuValidator) Validate(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error {
	return self.ValidateWorkloadGPU(ctx, resourceFields)
}

// ValidateWorkloadGPU implementation
func (self *gpuValidator) ValidateWorkloadGPU(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error {
	//
	//   common validations for any kind of GPU
	//
	return self.ValidateWorkloadAnyGPU(ctx, resourceFields)
}

// ValidateWorkloadAnyGPU validates the following valid cases:
//   - gpuDevices = 0 (or not provided)
//     and in this case no resource parameters can be specified
//   - gpuDevices = 1,2,3,4...
//     and in this case either gpu (between 0 and 1), gpuMemory must be specified,
//     and it is possible to also specify gpuLimit (also between 0 and 1) or gpuMemoryLimit, accordingly
//
// for backward compatibility, it is possible to specify gpu/memory only, with no gpuDevices. in this case
//   - if gpu = 0, 1 ,2, 3, 4... it is synonymous to gpuDevices = 0, 1, 2, 3, 4...
//   - if 0 < gpu < 1 (=fractional), then it is synonymous to gpuDevices = 1, gpu = portion
//   - if memory specified, assume 1 device with that amount of memory
func (self *gpuValidator) ValidateWorkloadAnyGPU(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) error {

	gpuGiven := resourceFields.Gpu.IsValueInitialized()
	gpuMemoryGiven := resourceFields.GpuMemory.IsValueInitialized()

	//
	//   first, let's sort ot the backward compatibility issue
	//
	gpuDevices, gpuFraction, err := self.DetermineGpuDevicesAndFraction(ctx, resourceFields)
	if err != nil {
		return err
	}

	if gpuFraction != nil {
		if gpuMemoryGiven {
			return errors.New("only one of the fields: gpu (fractional), gpuMemory should be provided. it is not permitted to specify both")
		}
	}

	//
	//   if gpuDevices == 1, either gpu or gpuMemory can be specified, if none specified it is assumed gpu = 1
	//   if gpuDevices == 0 or gpuDevices > 1, memory is not allowed
	//
	if gpuDevices == 1 {
		if !gpuGiven && !gpuMemoryGiven {
			return nil
		}
	}

	//backward compatibility without gpuDevices
	if workloadbuilders.IsWholeWithFractionalGpu(ctx, resourceFields.Gpu.Deref()) {
		return errors.New(WholeGPUWithFractionalGPUError)
	}

	if gpuDevices == 0 {
		if gpuMemoryGiven {
			return errors.New("gpuMemory can be provided only when requesting one or more GPU(s).")
		}
	}

	if bool2int(gpuMemoryGiven)+bool2int(gpuFraction != nil) > 1 {
		return errors.New("only one of the fields: gpu (fractional), gpuMemory should be provided. it is not permitted to specify both")
	}

	//
	//   validate gpu limit or memory limit
	//
	err = self.ValidateLimits(resourceFields, gpuFraction)
	if err != nil {
		return err
	}

	return nil
}

func (self *gpuValidator) DetermineGpuDevicesAndFraction(ctx context.Context,
	resourceFields *field_groups.ResourceAllocationFields) (int, *string, error) {

	gpuSpecified := resourceFields.Gpu.IsValueInitialized()
	gpuDevicesSpecified := resourceFields.GpuDevices.IsValueInitialized()

	if !gpuDevicesSpecified {
		if gpuSpecified {
			// only gpu specified, so either 1 gpu with specified gpu portion, or whole number (integer) or gpu devices
			gpuValue := resourceFields.Gpu.DerefStr()
			if workloadbuilders.IsFractionalGpu(ctx, gpuValue) {
				return 1, ptr.To(gpuValue), nil // 1 gpu, fractional
			} else {
				return int(resourceFields.Gpu.Deref()), nil, nil // whole number of gpus
			}
		}

		// for backward compatibility, if gpuDevices is not specified but memory specified,
		// assume gpuDevices is 1.
		if resourceFields.GpuMemory.IsValueInitialized() {
			return 1, nil, nil
		}

		return 0, nil, nil
	}

	// ok, so gpuDevices is specified. this is the "new" way of specifying gpu cores
	gpuDevices := resourceFields.GpuDevices.Deref()

	if gpuSpecified {
		gpuValue := resourceFields.Gpu.Deref()
		if gpuDevices == 0 && gpuValue != 0 {
			return gpuDevices, nil, errors.New("when the value of 'gpuDevices' is 0, 'gpu' should not be specified")
		}
		gpuValueStr := resourceFields.Gpu.DerefStr()
		if workloadbuilders.IsFractionalGpu(ctx, gpuValueStr) {
			// combination of gpuDevices + fractional gpu
			return gpuDevices, &gpuValueStr, nil
		}

		if gpuDevices >= 1 && gpuValue > 1 {
			return gpuDevices, nil, errors.New("when the value of 'gpuDevices' 1 and above, the value of 'gpu' must reside between 0 and 1")
		}
	}

	return gpuDevices, nil, nil
}

// ValidateLimits validates the gpu or memory limits. note that we only validate the limit which is relevant
// to this workload, meaning: if gpu is provided -> gpuLimit is relevant, gpuMemory is provided -> gpuMemoryLimit is relevant
func (self *gpuValidator) ValidateLimits(resourceFields *field_groups.ResourceAllocationFields, gpuFraction *string) error {
	gpuLimit := resourceFields.GpuLimit.Deref()
	if gpuLimit > 1 {
		return errors.New("gpuLimit must be no greater than 1")
	}

	if resourceFields.Gpu.IsValueInitialized() && resourceFields.GpuLimit.IsValueInitialized() {
		if gpuFraction == nil {
			return errors.New("gpuLimit can only be provided with fractional GPU (0 < gpu < 1)")
		}

		gpu := resourceFields.Gpu.Deref()
		if gpu > gpuLimit {
			return errors.New("the value of gpuLimit must be no less than gpu")
		}
	}

	if resourceFields.GpuMemory.IsValueInitialized() && resourceFields.GpuMemoryLimit.IsValueInitialized() {
		gpuMemoryRequest, errMemParse := resourceFields.GpuMemory.DerefQuantity(0.0)
		if errMemParse != nil {
			return errMemParse
		}

		gpuMemoryLimit, errMemLimitParse := resourceFields.GpuMemoryLimit.DerefQuantity(0.0)
		if errMemLimitParse != nil {
			return errMemLimitParse
		}

		if gpuMemoryRequest.Value() > gpuMemoryLimit.Value() {
			return errors.New("the value of gpuMemoryLimit must be no less than gpuMemory")
		}
	}

	return nil
}

func bool2int(val bool) int {
	if val {
		return 1
	}
	return 0
}
