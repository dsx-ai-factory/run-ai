/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {
	Context("Events Client", func() {

		DescribeTable("sort order validation",
			func(sortOrder string, expectedErr string) {
				eventsClient, err := NewBackendWorkloadEventsClient("url", "token", sortOrder)
				if expectedErr == "" {
					Expect(err).NotTo(HaveOccurred())
					Expect(eventsClient).NotTo(BeNil())
				} else {
					Expect(err.Error()).To(ContainSubstring(expectedErr))
					Expect(eventsClient).To(BeNil())
				}
			},
			Entry("asc", "asc", ""),
			Entry("desc", "desc", ""),
			Entry("error1", "ascending", "unable to parse sort order 'ascending'"),
			Entry("error2", "descending", "unable to parse sort order 'descending'"),
			Entry("error3", "bla", "unable to parse sort order 'bla'"),
			Entry("error4", "", "unable to parse sort order ''"),
		)

		It("sort order options", func() {
			sortOptions := GetEventSortOrderOptions()
			Expect(sortOptions).To(HaveLen(2))
			Expect(sortOptions).To(ContainElement("asc"))
			Expect(sortOptions).To(ContainElement("desc"))
		})
	})
})
