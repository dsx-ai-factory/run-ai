/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test step rule", func() {
	Context("non numeric value", func() {
		It("should return with error", func() {
			sr := StepRule{
				minVal: "aaa",
				step:   "5",
			}
			err := sr.Validate("111")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("must match the regular expression"))
		})
	})

	Context("non numeric value", func() {
		It("should return with error", func() {
			sr := StepRule{
				minVal: "3",
				step:   "xxx",
			}
			err := sr.Validate("111")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("must match the regular expression"))
		})
	})

	Context("non numeric value", func() {
		It("should return with error", func() {
			sr := StepRule{
				minVal: "3",
				step:   "2",
			}
			err := sr.Validate("zzz")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("must match the regular expression"))
		})
	})

	Context("valid range", func() {
		It("should succeed", func() {
			sr := StepRule{
				minVal: "3.3",
				step:   "6.6",
			}
			err := sr.Validate("9.9")
			Expect(err).To(BeNil())
		})
	})

})
