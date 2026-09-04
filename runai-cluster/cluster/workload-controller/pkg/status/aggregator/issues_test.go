/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Issues", func() {
	var (
		wi WorkloadIssues
	)

	BeforeEach(func() {
		wi = NewWorkloadIssues()
	})

	It("WorkloadIssues.Set sorts the groups", func() {
		wiToSet := NewWorkloadIssues()
		wiToSet.hasCriticalCompute = true
		wiToSet.hasIssues = true
		wiToSet.groupsWithCritical = []string{"c", "a", "ab", "b", "ca"}

		wi.Set(wiToSet)
		Expect(wi.hasIssues).To(Equal(wiToSet.hasIssues))
		Expect(wi.hasCritical).To(Equal(wiToSet.hasCritical))
		Expect(wi.hasCriticalCompute).To(Equal(wiToSet.hasCriticalCompute))

		Expect(wi.groupsWithCritical).To(HaveLen(len(wiToSet.groupsWithCritical)))

		for i := 0; i+1 < len(wi.groupsWithCritical); i++ {
			first := wi.groupsWithCritical[i]
			second := wi.groupsWithCritical[i+1]

			Expect(strings.Compare(first, second) <= 0).To(BeTrue())
		}
	})
})
