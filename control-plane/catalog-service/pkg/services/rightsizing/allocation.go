/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"

var ErrDoesNotFitSingleGPU = api.NewValidationError("minimum memory exceeds single-GPU capacity")

type Allocation struct {
	MinimumMemoryBytes   int64
	AllocatedMemoryBytes int64
	WholeGPU             bool
}

// SingleGPUAllocation applies the cutoff to a resolved single-device memory
// requirement. Both inputs must use the same byte units. Profile eligibility
// and parallel-profile selection are separate concerns.
// The cutoff decides whether to share the GPU; the minimum is used as published.
func SingleGPUAllocation(minimumBytes, capacityBytes, cutoffPercent int64) (Allocation, *api.Error) {
	if cutoffPercent < 1 || cutoffPercent > 100 {
		return Allocation{}, api.NewValidationError("whole-GPU sharing cutoff must be between 1 and 100 percent")
	}
	if minimumBytes <= 0 || capacityBytes <= 0 {
		return Allocation{}, api.NewValidationError("minimum memory and GPU capacity must be positive")
	}
	if minimumBytes > capacityBytes {
		return Allocation{}, ErrDoesNotFitSingleGPU
	}
	// ceil(capacity * percent / 100), without multiplying the full capacity
	// (which can overflow) or rounding a floating-point ratio at the boundary.
	threshold := capacityBytes/100*cutoffPercent +
		(capacityBytes%100*cutoffPercent+99)/100
	allocation := Allocation{
		MinimumMemoryBytes:   minimumBytes,
		AllocatedMemoryBytes: minimumBytes,
		WholeGPU:             minimumBytes >= threshold,
	}
	if allocation.WholeGPU {
		allocation.AllocatedMemoryBytes = capacityBytes
	}
	return allocation, nil
}

// allocateProfile constructs the allocation for a profile that already fits.
// Parallel profiles use their published GPU count and full device capacity.
// Only single-GPU profiles use the sharing cutoff; it changes allocation size,
// never the published minimum or the selected profile.
func allocateProfile(selected *profileCandidate, capacityBytes, cutoffPercent int64) (*api.Recommendation, *api.Error) {
	if selected.gpuCount > 1 {
		return &api.Recommendation{
			AllocationType:          api.RECOMMENDATIONALLOCATIONTYPE_GPU_COUNT,
			GpuCount:                selected.gpuCount,
			GpuMemoryPerDeviceBytes: capacityBytes,
		}, nil
	}
	allocation, err := SingleGPUAllocation(selected.minimumMemoryPerDeviceBytes, capacityBytes, cutoffPercent)
	if err != nil {
		return nil, err
	}
	allocationType := api.RECOMMENDATIONALLOCATIONTYPE_GPU_MEMORY
	if allocation.WholeGPU {
		allocationType = api.RECOMMENDATIONALLOCATIONTYPE_GPU_COUNT
	}
	return &api.Recommendation{
		AllocationType:          allocationType,
		GpuCount:                1,
		GpuMemoryPerDeviceBytes: allocation.AllocatedMemoryBytes,
	}, nil
}
