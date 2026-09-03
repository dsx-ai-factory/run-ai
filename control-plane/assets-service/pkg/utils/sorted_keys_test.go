/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validate sorting maps", func() {
	Context("no values at all", func() {
		It("should return empty array", func() {
			empty := map[string]string{}
			result := SortedMapKeys[string](empty)
			Expect(len(result)).To(Equal(0))
		})
	})

	Context("with values", func() {
		It("should return sorted", func() {
			empty := map[string]string{
				"zzz": "1",
				"xxx": "2",
				"yyy": "3",
			}
			result := SortedMapKeys[string](empty)
			Expect(result).To(Equal([]string{"xxx", "yyy", "zzz"}))
		})
	})
})
