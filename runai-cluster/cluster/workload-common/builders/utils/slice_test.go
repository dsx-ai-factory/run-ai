/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slice tests", func() {
	Context("AreEquivalent", func() {
		DescribeTable("string",
			func(a, b []string, expected bool) {
				Expect(AreEquivalent(a, b)).To(Equal(expected))
			},
			Entry("same order", []string{"apple", "banana", "cherry"}, []string{"apple", "banana", "cherry"}, true),
			Entry("different order", []string{"apple", "banana", "cherry"}, []string{"cherry", "banana", "apple"}, true),
			Entry("different", []string{"apple", "banana", "cherry"}, []string{"cherry", "banana"}, false),
			Entry("empty", []string{}, []string{}, true),
			Entry("empty/not empty", []string{}, []string{"bla"}, false),
			Entry("nil", nil, nil, true),
			Entry("empty/nil", []string{}, nil, true),
		)
	})
})
