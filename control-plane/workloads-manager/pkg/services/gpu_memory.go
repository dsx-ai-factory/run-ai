/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"

func legacyCompatibleNimSpec(source api.NimServiceSpec) api.NimServiceSpec {
	return projectNimGPUMemory(source, true)
}

func nimSpecForDB(source api.NimServiceSpec) api.NimServiceSpec {
	return projectNimGPUMemory(source, false)
}

// Project validated fields on a copy so response aliases do not leak into the DB.
func projectNimGPUMemory(source api.NimServiceSpec, response bool) api.NimServiceSpec {
	if source.Compute == nil {
		return source
	}
	compute := *source.Compute
	if compute.GpuMemoryPortionRequest != nil {
		compute.GpuPortionRequest = compute.GpuMemoryPortionRequest
	}
	if compute.GpuMemoryPortionLimit != nil {
		compute.GpuPortionLimit = compute.GpuMemoryPortionLimit
	}
	compute.GpuPortionRequest = copyGPUMemoryValue(compute.GpuPortionRequest)
	compute.GpuPortionLimit = copyGPUMemoryValue(compute.GpuPortionLimit)
	compute.GpuMemoryPortionRequest, compute.GpuMemoryPortionLimit = nil, nil
	if response {
		if compute.GpuPortionRequest == nil || *compute.GpuPortionRequest <= 1 {
			compute.GpuMemoryPortionRequest = copyGPUMemoryValue(compute.GpuPortionRequest)
		}
		compute.GpuMemoryPortionLimit = copyGPUMemoryValue(compute.GpuPortionLimit)
	}
	source.Compute = &compute
	return source
}

func copyGPUMemoryValue(source *float64) *float64 {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}
