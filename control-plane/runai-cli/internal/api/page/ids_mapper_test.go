/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("idsMapper", func() {
	Context("IDs mapper functionality tests", func() {
		It("limit is zero", func() {
			im := NewIdsMapper(0)
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(0))
			im.ShouldAdd("1")
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(1))
		})

		It("limit is negative", func() {
			im := NewIdsMapper(-1)
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(0))
			im.ShouldAdd("1")
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(1))
		})

		It("limit is 1", func() {
			im := NewIdsMapper(1)
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(0))
			im.ShouldAdd("1")
			Expect(im.ReachedLimit()).To(BeTrue())
			Expect(im.GetLen()).To(Equal(1))
		})

		It("limit is 4 - stop at limit", func() {
			im := NewIdsMapper(4)
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(0))
			for _, index := range []string{"1", "2", "3"} {
				im.ShouldAdd(index)
				Expect(im.ReachedLimit()).To(BeFalse())
			}
			im.ShouldAdd("4")
			Expect(im.ReachedLimit()).To(BeTrue())
		})
	})

	Context("IDs mapper duplicate tests", func() {
		It("limit is 2 - duplicate is not counted", func() {
			im := NewIdsMapper(2)
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(0))
			im.ShouldAdd("1")
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(1))
			im.ShouldAdd("1")
			Expect(im.ReachedLimit()).To(BeFalse())
			Expect(im.GetLen()).To(Equal(1))

			im.ShouldAdd("2")
			Expect(im.ReachedLimit()).To(BeTrue())
			Expect(im.GetLen()).To(Equal(2))
		})
	})
})
