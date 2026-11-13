// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	catalogconfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	sharedconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("metadata cache TTL configuration", func() {
	for _, tc := range []struct {
		name, value string
		want        int
		valid       bool
	}{
		{"default", "", 900, true}, {"override", "60", 60, true},
		{"zero", "0", 0, false}, {"negative", "-1", 0, false},
	} {
		It(tc.name, func() {
			GinkgoT().Setenv("NGC_METADATA_CACHE_TTL_SECONDS", tc.value)
			sharedconfig.FlushAll()
			DeferCleanup(sharedconfig.FlushAll)
			err := sharedconfig.Register[*catalogconfig.NGC]()
			if !tc.valid {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).To(BeNil())
			Expect(sharedconfig.Get[*catalogconfig.NGC]().MetadataCacheTTLSeconds).To(Equal(tc.want))
		})
	}
})
