/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page_test

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/page"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Project Page Appender", func() {
	Context("Project", func() {
		It("Add nil page", func() {
			pa := page.NewAppender[openapi.Project](0)
			err := pa.AddPage(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeFalse())
			Expect(pa.GetList()).To(HaveLen(0))
		})

		It("Last item has no creation time", func() {
			projectList := []openapi.Project{
				{Id: "1"},
				{Id: "2"},
			}
			pa := page.NewAppender[openapi.Project](1)
			Expect(pa.ReachedLimit()).To(BeFalse())
			err := pa.AddPage(projectList)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("next token cannot be with 0 timestamp"))
			Expect(pa.ReachedLimit()).To(BeTrue())
			Expect(pa.GetList()).To(HaveLen(1))
		})

		It("No maxItems", func() {
			projectList := []openapi.Project{
				{Id: "1"},
			}
			pa := page.NewAppender[openapi.Project](0)
			Expect(pa.ReachedLimit()).To(BeFalse())
			err := pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeFalse())
			Expect(pa.GetList()).To(HaveLen(1))
		})

		It("No maxItems, duplicate workload", func() {
			projectList := []openapi.Project{
				{Id: "1"},
			}
			pa := page.NewAppender[openapi.Project](0)
			Expect(pa.ReachedLimit()).To(BeFalse())
			err := pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeFalse())
			Expect(pa.GetList()).To(HaveLen(1))

			err = pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeFalse())
			Expect(pa.GetList()).To(HaveLen(1))
		})

		It("Limit 2, incremental increase", func() {
			projectList := []openapi.Project{
				{Id: "1", CreatedAt: time.Unix(11, 0)},
			}
			pa := page.NewAppender[openapi.Project](2)
			Expect(pa.ReachedLimit()).To(BeFalse())
			err := pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeFalse())
			Expect(pa.GetList()).To(HaveLen(1))

			projectList[0].Id = "2"
			err = pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeTrue())
			Expect(pa.GetList()).To(HaveLen(2))

			projectList[0].Id = "3"
			err = pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeTrue())
			Expect(pa.GetList()).To(HaveLen(2))
		})

		It("Limit 2, over the max maxItems", func() {
			projectList := []openapi.Project{
				{Id: "1"},
				{Id: "2"},
				{Id: "3", CreatedAt: time.Unix(11, 0)},
				{Id: "4"},
			}
			pa := page.NewAppender[openapi.Project](2)
			Expect(pa.ReachedLimit()).To(BeFalse())
			err := pa.AddPage(projectList)
			Expect(err).ToNot(HaveOccurred())
			Expect(pa.ReachedLimit()).To(BeTrue())
			Expect(pa.GetList()).To(HaveLen(2))
		})
	})
})
