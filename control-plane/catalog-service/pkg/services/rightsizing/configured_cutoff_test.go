// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package rightsizing

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"testing"
)

func TestSingleGPUAllocationUsesConfiguredSharingCutoff(t *testing.T) {
	for _, tc := range []struct {
		name            string
		minimum, cutoff int64
		whole           bool
	}{
		{"below configured cutoff", 79, 80, false}, {"at configured cutoff", 80, 80, true},
		{"above configured cutoff", 85, 80, true}, {"same minimum below default cutoff", 85, 90, false},
		{"100 percent permits sharing below capacity", 99, 100, false}, {"100 percent exact fit", 100, 100, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SingleGPUAllocation(tc.minimum, 100, tc.cutoff)
			if err != nil || got.WholeGPU != tc.whole {
				t.Fatalf("allocation=%+v error=%v", got, err)
			}
		})
	}
	for _, cutoff := range []int64{-1, 0, 101} {
		if _, err := SingleGPUAllocation(20, 100, cutoff); err == nil {
			t.Errorf("accepted invalid cutoff %d", cutoff)
		}
	}
}

func TestInvalidSharingPolicyDoesNotBecomeEstimatedGPUCount(t *testing.T) {
	capacity := int64(100_000_000_000)
	result := recommendForHardware(gpuinventory.Hardware{Type: "unknown", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{sizingProfile("generic", 20)}, 0)
	if result.Status != "no_recommendation" || result.Recommendation != nil || result.Reason == "" {
		t.Fatalf("invalid sharing policy produced allocation: %+v", result)
	}
}
