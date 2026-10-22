/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"errors"
	"math"
	"testing"
)

func TestSingleGPUAllocationUsesNinetyPercentCutoff(t *testing.T) {
	for _, tt := range []struct {
		name     string
		minimum  int64
		capacity int64
		whole    bool
	}{
		{"below cutoff", 71_999_999_999, 80_000_000_000, false},
		{"exactly cutoff", 72_000_000_000, 80_000_000_000, true},
		{"above cutoff", 74_000_000_000, 80_000_000_000, true},
		{"exact fit", 80_000_000_000, 80_000_000_000, true},
		{"fractional byte cutoff rounds up", 90, 101, false},
		{"first byte above fractional cutoff", 91, 101, true},
		{"large capacity does not overflow", math.MaxInt64, math.MaxInt64, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SingleGPUAllocation(tt.minimum, tt.capacity)
			if err != nil {
				t.Fatal(err)
			}
			if got.WholeGPU != tt.whole || got.MinimumMemoryBytes != tt.minimum {
				t.Fatalf("unexpected allocation: %+v", got)
			}
			wantMemory := tt.minimum
			if tt.whole {
				wantMemory = tt.capacity
			}
			if got.AllocatedMemoryBytes != wantMemory {
				t.Fatalf("allocated memory = %d; want %d", got.AllocatedMemoryBytes, wantMemory)
			}
		})
	}
}

func TestSingleGPUAllocationNeverRoundsOversizedRequirementDown(t *testing.T) {
	_, err := SingleGPUAllocation(80_000_000_001, 80_000_000_000)
	if !errors.Is(err, ErrDoesNotFitSingleGPU) {
		t.Fatalf("oversized requirement must not become one whole GPU: %v", err)
	}
}

func TestSingleGPUAllocationRejectsMissingOrInvalidMemory(t *testing.T) {
	for _, tt := range []struct{ minimum, capacity int64 }{
		{0, 80}, {-1, 80}, {32, 0}, {32, -1},
	} {
		if _, err := SingleGPUAllocation(tt.minimum, tt.capacity); err == nil {
			t.Errorf("accepted minimum %d capacity %d", tt.minimum, tt.capacity)
		}
	}
}
