// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"math"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

type profileCandidate struct {
	profile      *ngcmetadata.Profile
	match        hardwareMatch
	count        int64
	minimumBytes int64
}

func selectProfile(profiles []ngcmetadata.Profile, pciDevice string, capacity, nodeCount int64) *profileCandidate {
	var chosen *profileCandidate
	for i := range profiles {
		profile := &profiles[i]
		count, valid := profileDeviceCount(profile)
		if !valid || count > nodeCount || profile.ID == "" || profile.MinimumVRAMPerDeviceGB == nil {
			continue
		}
		minimum := *profile.MinimumVRAMPerDeviceGB
		if math.IsNaN(minimum) || math.IsInf(minimum, 0) || minimum <= 0 || minimum >= float64(math.MaxInt64)/1e9 {
			continue
		}
		// NGC minima are per device in decimal GB. Never divide a TP1 floor
		// across GPUs or infer another parallel layout from a memory deficit.
		minimumBytes := int64(math.Ceil(minimum * 1e9))
		match := matchHardware(*profile, pciDevice)
		if minimumBytes > capacity || match == noHardwareMatch {
			continue
		}
		candidate := &profileCandidate{
			profile:      profile,
			match:        match,
			count:        count,
			minimumBytes: minimumBytes,
		}
		if chosen == nil || candidate.precedes(chosen) {
			chosen = candidate
		}
	}
	return chosen
}

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

func (c *profileCandidate) precedes(other *profileCandidate) bool {
	// Retain the fitting single-device path before considering parallel
	// layouts. Within each path prefer hardware evidence, then fewer GPUs.
	if (c.count == 1) != (other.count == 1) {
		return c.count == 1
	}
	if c.match != other.match {
		return c.match < other.match
	}
	if c.count != other.count {
		return c.count < other.count
	}
	if *c.profile.MinimumVRAMPerDeviceGB != *other.profile.MinimumVRAMPerDeviceGB {
		return *c.profile.MinimumVRAMPerDeviceGB < *other.profile.MinimumVRAMPerDeviceGB
	}
	return c.profile.ID < other.profile.ID
}
