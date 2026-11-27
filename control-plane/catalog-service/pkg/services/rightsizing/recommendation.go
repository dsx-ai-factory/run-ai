/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

// recommendForHardware owns one complete result: check the hardware evidence,
// select a fitting profile, record its published requirements, then allocate.
// Unavailable evidence or invalid allocation policy leaves a reason instead.
func recommendForHardware(gpu gpuinventory.Hardware, profiles []ngcmetadata.Profile, cutoffPercent int64) api.RecommendationResult {
	result := api.RecommendationResult{
		GpuType: gpu.Type,
		Status:  api.RECOMMENDATIONSTATUS_NO_RECOMMENDATION,
	}
	if gpu.MemoryBytes == nil || *gpu.MemoryBytes <= 0 {
		result.Reason = "GPU memory capacity is not reported by the cluster"
		return result
	}
	result.GpuMemoryCapacityBytes = *gpu.MemoryBytes
	// Older inventory responses can omit this count. A pool total cannot
	// replace it, so leave this entry without a recommendation when it is unknown.
	if gpu.MaxDevicesPerNode <= 0 {
		result.Reason = "Matching single-node GPU count is not reported by the cluster"
		return result
	}
	selected := selectFittingProfile(profiles, *gpu.MemoryBytes, gpu.MaxDevicesPerNode)
	if selected == nil {
		result.Reason = "No GPU-agnostic published profile with explicit TP/PP and a positive minimum fits per-device memory and the matching GPU count on one node"
		return result
	}
	// Preserve the selected profile's evidence even if allocation policy fails.
	profile := selected.profile
	result.MinimumGpuMemoryPerDeviceGb = *profile.MinimumVRAMPerDeviceGB
	result.ProfileId = profile.ID
	result.Precision = profile.Precision
	result.InferenceBackend = profile.InferenceBackend
	result.TensorParallelSize = int64(*profile.TensorParallelSize)
	result.PipelineParallelSize = int64(*profile.PipelineParallelSize)
	allocation, err := allocateProfile(selected, *gpu.MemoryBytes, cutoffPercent)
	if err != nil {
		result.Reason = "GPU allocation sharing policy is invalid"
		return result
	}
	result.Status = api.RECOMMENDATIONSTATUS_RECOMMENDED
	result.Recommendation = allocation
	return result
}
