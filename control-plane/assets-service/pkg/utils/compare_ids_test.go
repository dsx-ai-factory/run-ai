/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validate assets comparison", func() {
	Context("simple compute gpu requests", func() {
		It("create two empty lists", func() {
			existing := []string{}
			requested := []string{}
			result := CompareIds(existing, requested)
			Expect(len(result)).To(Equal(0))
		})

		It("only existing", func() {
			existing := []string{"e1", "e2"}
			requested := []string{}
			expected := RequestedVsExistingMap{
				"e1": RequestedVsExisting{Existing: true},
				"e2": RequestedVsExisting{Existing: true},
			}
			result := CompareIds(existing, requested)
			Expect(result).To(Equal(expected))
		})

		It("only requested", func() {
			existing := []string{}
			requested := []string{"r1", "r2"}
			expected := RequestedVsExistingMap{
				"r1": RequestedVsExisting{Requested: true},
				"r2": RequestedVsExisting{Requested: true},
			}
			result := CompareIds(existing, requested)
			Expect(result).To(Equal(expected))
		})

		It("mixed lists", func() {
			existing := []string{"e1", "both1", "e2", "both2"}
			requested := []string{"r1", "both2", "both1", "r2"}
			expected := RequestedVsExistingMap{
				"e1":    RequestedVsExisting{Existing: true},
				"e2":    RequestedVsExisting{Existing: true},
				"r1":    RequestedVsExisting{Requested: true},
				"r2":    RequestedVsExisting{Requested: true},
				"both1": RequestedVsExisting{Existing: true, Requested: true},
				"both2": RequestedVsExisting{Existing: true, Requested: true},
			}
			result := CompareIds(existing, requested)
			Expect(result).To(Equal(expected))
		})
	})
})
