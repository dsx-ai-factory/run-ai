/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

// WorkloadRequirements are all sort of requirements that the workload service needs to
// validate before allowing the workload to be submitted.
type WorkloadRequirements struct {
	// workload uses GPU memory limit, this requires the GPU optimization tenant setting
	// to be enabled
	GpuOptimization bool
}

func (wr WorkloadRequirements) Or(anotherReq WorkloadRequirements) WorkloadRequirements {
	return WorkloadRequirements{
		GpuOptimization: wr.GpuOptimization || anotherReq.GpuOptimization,
	}
}
