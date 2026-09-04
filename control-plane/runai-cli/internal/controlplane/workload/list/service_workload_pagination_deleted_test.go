/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workload Paginator with Deleted Flag", func() {
	var (
		ctrl     *gomock.Controller
		mockpage *cpworkloadmocks.MockPaginator
		service  *Pager
		fb       *filter.Builder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockpage = cpworkloadmocks.NewMockPaginator(ctrl)
		service = NewPager(mockpage)
		fb = filter.NewBuilder()
	})

	Context("When deleted flag is false", func() {
		It("should call List with deleted=false", func() {
			response := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "1", Name: "active-workload-1", CreatedAt: time.Unix(10, 0)},
					{Id: "2", Name: "active-workload-2", CreatedAt: time.Unix(20, 0)},
				},
			}
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), false).Return(response, nil)

			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(2))
			Expect(workloadList[0].Name).To(Equal("active-workload-1"))
			Expect(workloadList[1].Name).To(Equal("active-workload-2"))
		})

		It("should handle pagination with deleted=false", func() {
			response1 := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "1", Name: "active-1", CreatedAt: time.Unix(10, 0)},
					{Id: "2", Name: "active-2", CreatedAt: time.Unix(20, 0)},
					{Id: "3", Name: "active-3", CreatedAt: time.Unix(30, 0)},
				},
			}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response1, nil)

			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(next).ToNot(BeNil())
			Expect(workloadList).To(HaveLen(2))
		})
	})

	Context("When deleted flag is true", func() {
		It("should call List with deleted=true", func() {
			response := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "1", Name: "deleted-workload-1", CreatedAt: time.Unix(10, 0)},
					{Id: "2", Name: "deleted-workload-2", CreatedAt: time.Unix(20, 0)},
				},
			}
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), true).Return(response, nil)

			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(2))
			Expect(workloadList[0].Name).To(Equal("deleted-workload-1"))
			Expect(workloadList[1].Name).To(Equal("deleted-workload-2"))
		})

		It("should handle empty response for deleted workloads", func() {
			response := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{},
			}
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), true).Return(response, nil)

			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(0))
		})

		It("should handle multi-page requests for deleted workloads", func() {
			response1 := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "1", Name: "deleted-1", CreatedAt: time.Unix(10, 0)},
					{Id: "2", Name: "deleted-2", CreatedAt: time.Unix(20, 0)},
					{Id: "3", Name: "deleted-3", CreatedAt: time.Unix(30, 0)},
				},
			}
			response2 := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "4", Name: "deleted-4", CreatedAt: time.Unix(40, 0)},
					{Id: "5", Name: "deleted-5", CreatedAt: time.Unix(50, 0)},
				},
			}

			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), true).Return(response1, nil)
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), gomock.Any(), true).Return(response2, nil)

			workloadList, next, err := service.multiPages(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(next).To(BeNil())
			Expect(workloadList).To(HaveLen(4))
		})
	})

	Context("Mixed scenarios", func() {
		It("should handle switching between deleted and non-deleted queries", func() {
			activeResponse := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "1", Name: "active-1", CreatedAt: time.Unix(10, 0)},
				},
			}
			deletedResponse := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{Id: "2", Name: "deleted-1", CreatedAt: time.Unix(20, 0)},
				},
			}

			// First query for active workloads
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), false).Return(activeResponse, nil)
			activeList, _, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(activeList).To(HaveLen(1))
			Expect(activeList[0].Name).To(Equal("active-1"))

			// Then query for deleted workloads
			mockpage.EXPECT().List(gomock.Any(), int32(11), int32(0), fb.StringList(), true).Return(deletedResponse, nil)
			deletedList, _, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 10}, fb, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(deletedList).To(HaveLen(1))
			Expect(deletedList[0].Name).To(Equal("deleted-1"))
		})
	})
})
