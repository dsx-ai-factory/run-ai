/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"errors"
	"math"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("single GPU allocation uses ninety percent cutoff", func() {
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
		It(tt.name, func() {
			got, err := SingleGPUAllocation(tt.minimum, tt.capacity, 90)
			Expect(err).To(BeNil())
			Expect(got.WholeGPU).To(Equal(tt.whole))
			Expect(got.MinimumMemoryBytes).To(Equal(tt.minimum))
			wantMemory := tt.minimum
			if tt.whole {
				wantMemory = tt.capacity
			}
			Expect(got.AllocatedMemoryBytes).To(Equal(wantMemory))
		})
	}
})

var _ = Describe("allocation failures use API validation errors", func() {
	for _, testCase := range []struct {
		name                      string
		minimum, capacity, cutoff int64
	}{
		{"invalid policy", 20, 80, 0},
		{"missing memory", 0, 80, 90},
		{"oversized requirement", 81, 80, 90},
	} {
		It(testCase.name, func() {
			_, err := SingleGPUAllocation(testCase.minimum, testCase.capacity, testCase.cutoff)
			var apiError *api.Error
			Expect(errors.As(err, &apiError)).To(BeTrue())
			Expect(apiError.Code).To(BeNumerically("==", http.StatusBadRequest))
			Expect(apiError.Message).NotTo(BeEmpty())
		})
	}
})

var _ = It("single GPU allocation never rounds oversized requirement down", func() {
	_, err := SingleGPUAllocation(80_000_000_001, 80_000_000_000, 90)
	Expect(errors.Is(err, ErrDoesNotFitSingleGPU)).To(BeTrue())
})

var _ = It("single GPU allocation rejects missing or invalid memory", func() {
	for _, tt := range []struct{ minimum, capacity int64 }{
		{0, 80}, {-1, 80}, {32, 0}, {32, -1},
	} {
		{
			_, err := SingleGPUAllocation(tt.minimum, tt.capacity, 90)
			Expect(err).NotTo(BeNil())
		}
	}
})

var _ = It("single GPU allocation uses configured sharing cutoff", func() {
	for _, tc := range []struct {
		name            string
		minimum, cutoff int64
		whole           bool
	}{
		{"below configured cutoff", 79, 80, false}, {"at configured cutoff", 80, 80, true},
		{"above configured cutoff", 85, 80, true}, {"same minimum below default cutoff", 85, 90, false},
		{"100 percent permits sharing below capacity", 99, 100, false}, {"100 percent exact fit", 100, 100, true},
	} {
		By(tc.name)
		func() {
			got, err := SingleGPUAllocation(tc.minimum, 100, tc.cutoff)
			Expect(err).To(BeNil())
			Expect(got.WholeGPU).To(Equal(tc.whole))
		}()
	}
	for _, cutoff := range []int64{-1, 0, 101} {
		{
			_, err := SingleGPUAllocation(20, 100, cutoff)
			Expect(err).NotTo(BeNil())
		}
	}
})
