// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"math"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

type profileCandidate struct {
	profile                     *ngcmetadata.Profile
	gpuCount                    int64
	minimumMemoryPerDeviceBytes int64
}

// selectFittingProfile ranks only profiles whose published memory and TP × PP
// requirements fit this hardware. Allocation never relaxes those requirements.
func selectFittingProfile(profiles []ngcmetadata.Profile, capacityBytes, maxDevicesPerNode int64) *profileCandidate {
	var selected *profileCandidate
	for i := range profiles {
		candidate := fittingCandidate(&profiles[i], capacityBytes, maxDevicesPerNode)
		if candidate == nil {
			continue
		}
		if selected == nil || candidate.preferredTo(selected) {
			selected = candidate
		}
	}
	return selected
}

// fittingCandidate establishes eligibility and per-device/single-node fit.
// Nil means the profile lacks usable evidence or does not fit this hardware.
func fittingCandidate(profile *ngcmetadata.Profile, capacityBytes, maxDevicesPerNode int64) *profileCandidate {
	// An explicit hardware target needs identity validation, which is outside
	// this API version. Tested-device lists are informational, not allowlists.
	if profile.GPU != "" || profile.GPUDevice != "" {
		return nil
	}
	gpuCount, valid := profileDeviceCount(profile)
	if !valid || gpuCount > maxDevicesPerNode || profile.ID == "" || profile.MinimumVRAMPerDeviceGB == nil {
		return nil
	}
	minimum := *profile.MinimumVRAMPerDeviceGB
	if math.IsNaN(minimum) || math.IsInf(minimum, 0) || minimum <= 0 || minimum >= float64(math.MaxInt64)/1e9 {
		return nil
	}
	// NGC minima are per device in decimal GB. Never divide a TP1 floor
	// across GPUs or infer another parallel layout from a memory deficit.
	minimumBytes := int64(math.Ceil(minimum * 1e9))
	if minimumBytes > capacityBytes {
		return nil
	}
	return &profileCandidate{
		profile:                     profile,
		gpuCount:                    gpuCount,
		minimumMemoryPerDeviceBytes: minimumBytes,
	}
}

// TP and PP are the published tensor- and pipeline-parallel sizes. Their
// product is the GPU count for this exact profile; missing sizes cannot
// establish a layout and must not default to one.
func profileDeviceCount(profile *ngcmetadata.Profile) (int64, bool) {
	if profile.TensorParallelSize == nil || profile.PipelineParallelSize == nil {
		return 0, false
	}
	tp, pp := int64(*profile.TensorParallelSize), int64(*profile.PipelineParallelSize)
	if tp <= 0 || pp <= 0 || tp > math.MaxInt64/pp {
		return 0, false
	}
	count := tp * pp
	if profile.MaxGPUCount != nil && count > *profile.MaxGPUCount {
		return 0, false
	}
	return count, true
}

func (c *profileCandidate) preferredTo(other *profileCandidate) bool {
	// Prefer fewer GPUs, including a fitting single-GPU profile, then a lower
	// per-device minimum and stable profile ID. A smaller minimum can require
	// more GPUs, so memory alone must not decide the parallel layout.
	if c.gpuCount != other.gpuCount {
		return c.gpuCount < other.gpuCount
	}
	if *c.profile.MinimumVRAMPerDeviceGB != *other.profile.MinimumVRAMPerDeviceGB {
		return *c.profile.MinimumVRAMPerDeviceGB < *other.profile.MinimumVRAMPerDeviceGB
	}
	return c.profile.ID < other.profile.ID
}
