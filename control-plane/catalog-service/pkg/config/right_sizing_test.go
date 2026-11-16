// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package config_test

import (
	catalogconfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services/rightsizing"
	sharedconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sharing cutoff configuration changes allocation", func() {
	for _, tc := range []struct {
		name, value string
		whole       bool
	}{
		{"default 90 percent", "", false}, {"override to 80 percent", "80", true},
	} {
		It(tc.name, func() {
			GinkgoT().Setenv("WHOLE_GPU_CUTOFF_PERCENT", tc.value)
			sharedconfig.FlushAll()
			DeferCleanup(sharedconfig.FlushAll)
			{
				err := sharedconfig.Register[*catalogconfig.Services]()
				Expect(err).To(BeNil())
			}
			cutoff := sharedconfig.Get[*catalogconfig.Services]().WholeGPUCutoffPercent
			result, err := rightsizing.SingleGPUAllocation(85, 100, cutoff)
			Expect(err).To(BeNil())
			Expect(result.WholeGPU).To(Equal(tc.whole))
		})
	}
})

var _ = Describe("sharing cutoff configuration rejects invalid percentages", func() {
	for _, value := range []string{"0", "-1", "101"} {
		It(value, func() {
			GinkgoT().Setenv("WHOLE_GPU_CUTOFF_PERCENT", value)
			sharedconfig.FlushAll()
			DeferCleanup(sharedconfig.FlushAll)
			{
				err := sharedconfig.Register[*catalogconfig.Services]()
				Expect(err).NotTo(BeNil())
			}
		})
	}
})
