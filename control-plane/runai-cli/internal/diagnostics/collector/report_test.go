/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package collector

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CollectionReporter", func() {
	Context("report", func() {
		It("no errors", func() {
			cr := CollectionReporter{
				reportByNamespace: make(map[string]*NamespaceReport),
			}
			cr.ReportNamespace("namespace-a", nil)
			cr.ReportContainer("namespace-a", "pod-1", "c-a1", false, nil)
			cr.ReportContainer("namespace-a", "pod-2", "c-a2", false, nil)
			cr.ReportContainer("namespace-a", "pod-3", "c-a3", false, nil)
			cr.ReportContainer("namespace-a", "pod-4", "c-a4", false, nil)
			cr.ReportContainer("namespace-a", "pod-4", "c-a5", true, nil)
			nr := cr.reportByNamespace["namespace-a"]
			Expect(nr.FailCount).To(Equal(0))
			Expect(nr.SuccessCount).To(Equal(5))
			Expect(nr.listErr).To(BeNil())
			report := cr.CreateReport()
			expected :=
				`Summary:

Containers [current, previous]:
 Namespace       Total       Succeeded   Failed     
────────────────────────────────────────────────────
 namespace-a     5           5           0          
`
			Expect(report).To(Equal(expected))
		})
		It("should report errors", func() {
			cr := NewCollectionReporter(true)
			cr.ReportNamespace("namespace-c", fmt.Errorf("failed to authenticate"))
			cr.ReportContainer("namespace-a", "pod-1", "c-a1", false, nil)
			cr.ReportContainer("namespace-a", "pod-2", "c-a2", false, nil)
			cr.ReportContainer("namespace-b", "pod-0", "c-b0", true, fmt.Errorf("OOM"))
			cr.ReportContainer("namespace-b", "pod-1", "c-b1", false, nil)
			cr.ReportContainer("namespace-b", "pod-2", "c-b2", false, fmt.Errorf("NPE"))
			cr.ReportContainer("namespace-b", "pod-3", "c-b3", true, nil)
			cr.ReportContainer("namespace-d", "pod-1", "c-c1", false, fmt.Errorf("NPE"))

			namespaceA := cr.reportByNamespace["namespace-a"]
			namespaceB := cr.reportByNamespace["namespace-b"]
			namespaceC := cr.reportByNamespace["namespace-c"]
			Expect(namespaceA.FailCount).To(Equal(0))
			Expect(namespaceA.SuccessCount).To(Equal(2))
			Expect(namespaceA.listErr).To(BeNil())
			Expect(namespaceB.FailCount).To(Equal(2))
			Expect(namespaceB.SuccessCount).To(Equal(2))
			Expect(namespaceB.listErr).To(BeNil())
			Expect(namespaceC.listErr).To(Equal(fmt.Errorf("failed to authenticate")))

			report := cr.CreateReport()
			expected :=
				`Summary:

Namespaces:
Failed to list namespace-c pods: [failed to authenticate]

Containers [current]:
 Namespace       Total       Succeeded   Failed     
────────────────────────────────────────────────────
 namespace-a     2           2           0          
 namespace-b     4           2           2          
 namespace-c     0           0           0          
 namespace-d     1           0           1          


namespace-b:
Failed to get logs of pod-0/c-b0 (previous): [OOM]
Failed to get logs of pod-2/c-b2: [NPE]

namespace-d:
Failed to get logs of pod-1/c-c1: [NPE]

`
			Expect(report).To(Equal(expected))
		})
	})
	Context("getPodsErrors", func() {
		const namespace = "namespace-a"

		It("no error", func() {
			cr := NewCollectionReporter(true)
			cr.ReportContainer(namespace, "pod-1", "container-a", false, nil)
			cr.ReportContainer(namespace, "pod-1", "container-a", true, nil)
			cr.ReportContainer(namespace, "pod-3", "container-b", false, nil)
			Expect(cr.getPodsErrors(namespace)).To(BeEmpty())
			Expect(cr.reportByNamespace[namespace].SuccessCount).To(Equal(3))
			Expect(cr.reportByNamespace[namespace].FailCount).To(Equal(0))
		})

		It("should have previous", func() {
			cr := NewCollectionReporter(true)
			cr.ReportContainer(namespace, "pod-1", "container-1", true, fmt.Errorf("NPE"))
			cr.ReportContainer(namespace, "pod-1", "container-1", false, nil)
			errs := cr.getPodsErrors(namespace)
			Expect(errs).To(Equal([]string{"Failed to get logs of pod-1/container-1 (previous): [NPE]"}))
			Expect(cr.reportByNamespace[namespace].SuccessCount).To(Equal(1))
			Expect(cr.reportByNamespace[namespace].FailCount).To(Equal(1))
		})

		It("shouldn't have previous", func() {
			cr := NewCollectionReporter(true)
			cr.ReportContainer(namespace, "pod-1", "container-1", false, fmt.Errorf("NPE-new"))
			cr.ReportContainer(namespace, "pod-1", "container-1", true, nil)
			errs := cr.getPodsErrors(namespace)
			Expect(errs).To(Equal([]string{"Failed to get logs of pod-1/container-1: [NPE-new]"}))
			Expect(cr.reportByNamespace[namespace].SuccessCount).To(Equal(1))
			Expect(cr.reportByNamespace[namespace].FailCount).To(Equal(1))
		})
	})
})
