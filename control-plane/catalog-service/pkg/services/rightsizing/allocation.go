/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import "errors"

// WholeGPUCutoffPercent is a provisional product policy, pending further
// research. It is not runtime headroom and must never make an oversized
// requirement appear to fit.
const WholeGPUCutoffPercent int64 = 90

var ErrDoesNotFitSingleGPU = errors.New("minimum memory exceeds single-GPU capacity")

type Allocation struct {
	MinimumMemoryBytes   int64
	AllocatedMemoryBytes int64
	WholeGPU             bool
}

// SingleGPUAllocation applies the cutoff to an already resolved, compatible
// single-device requirement. Both inputs must use the same byte units. Hardware
// compatibility, usable capacity, and multi-GPU estimates are separate concerns.
func SingleGPUAllocation(minimumBytes, capacityBytes int64) (Allocation, error) {
	if minimumBytes <= 0 || capacityBytes <= 0 {
		return Allocation{}, errors.New("minimum memory and GPU capacity must be positive")
	}
	if minimumBytes > capacityBytes {
		return Allocation{}, ErrDoesNotFitSingleGPU
	}
	// ceil(capacity * percent / 100), without multiplying the full capacity
	// (which can overflow) or rounding a floating-point ratio at the boundary.
	threshold := capacityBytes/100*WholeGPUCutoffPercent +
		(capacityBytes%100*WholeGPUCutoffPercent+99)/100
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
