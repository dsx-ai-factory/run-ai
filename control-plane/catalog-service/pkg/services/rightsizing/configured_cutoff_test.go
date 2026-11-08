// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package rightsizing

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

var _ = It("invalid sharing policy does not become estimated GPU count", func() {
	capacity := int64(100_000_000_000)
	result := recommendForHardware(gpuinventory.Hardware{Type: "unknown", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{sizingProfile("generic", 20)}, 0)
	Expect(string(result.Status)).To(Equal("no_recommendation"))
	Expect(result.Recommendation).To(BeNil())
	Expect(result.Reason).NotTo(BeEmpty())
})
