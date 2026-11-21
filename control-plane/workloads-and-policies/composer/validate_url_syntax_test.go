/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateUrlSyntax", func() {
	DescribeTable("host validation",
		func(url string, valid bool) {
			err := ValidateUrlSyntax(convutils.Ref(url), false)
			if valid {
				Expect(err).To(BeNil())
			} else {
				Expect(err).NotTo(BeNil())
			}
		},
		Entry("domain", "http://example.com/path", true),
		Entry("domain with port", "http://example.com:8080", true),
		Entry("ipv4 with port", "http://10.0.0.1:8080", true),
		Entry("ipv6 literal", "http://[2600:1f18::1]", true),
		Entry("ipv6 literal with port", "http://[2600:1f18::1]:8080/path", true),
		Entry("single label host", "http://localhost", false),
	)
})
