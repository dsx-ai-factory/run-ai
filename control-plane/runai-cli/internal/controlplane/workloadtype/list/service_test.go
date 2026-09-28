/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/list"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workload Type List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		service    *list.Service
		mockClient *cpmocks.MockWorkloadTypeLister
		ctx        context.Context
		fb         *filter.Builder
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockWorkloadTypeLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
		fb = filter.NewBuilder()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return workload types successfully", func() {
			wt1 := newTestWorkloadTypeForService("interactive", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
			wt2 := newTestWorkloadTypeForService("train", "Train", "High", openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)
			mockResponse := &openapi.ListWorkloadTypes200Response{
				Types: []openapi.WorkloadTypeConfig{wt1, wt2},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, "cpUrl", "token", fb, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return error when API call fails", func() {
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			result, err := service.List(ctx, "cpUrl", "token", fb, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("workload type API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty workload type list", func() {
			mockResponse := &openapi.ListWorkloadTypes200Response{
				Types: []openapi.WorkloadTypeConfig{},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, "cpUrl", "token", fb, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply external types only filter correctly", func() {
			wt1 := newTestWorkloadTypeWithSourceForService("external-wt", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE, "external")
			mockResponse := &openapi.ListWorkloadTypes200Response{
				Types: []openapi.WorkloadTypeConfig{wt1},
			}
			externalOnly := true
			mockClient.EXPECT().List(ctx, gomock.Any(), &externalOnly).Return(mockResponse, nil)

			result, err := service.List(ctx, "cpUrl", "token", fb, &externalOnly)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply filter by correctly", func() {
			wt1 := newTestWorkloadTypeForService("interactive", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
			mockResponse := &openapi.ListWorkloadTypes200Response{
				Types: []openapi.WorkloadTypeConfig{wt1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, "cpUrl", "token", fb, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return all types when no external filter is provided", func() {
			wt1 := newTestWorkloadTypeWithSourceForService("internal-wt", "Train", "High", openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE, "internal")
			wt2 := newTestWorkloadTypeWithSourceForService("external-wt", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE, "external")
			mockResponse := &openapi.ListWorkloadTypes200Response{
				Types: []openapi.WorkloadTypeConfig{wt1, wt2},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), nil).Return(mockResponse, nil)

			result, err := service.List(ctx, "cpUrl", "token", fb, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})
})

// Helper function to create test workload type configs
func newTestWorkloadTypeForService(name string, category string, priority string, preemptibility openapi.WorkloadTypePreemptibility) openapi.WorkloadTypeConfig {
	return openapi.WorkloadTypeConfig{
		Name:           name,
		CategoryName:   &category,
		PriorityName:   &priority,
		Preemptibility: preemptibility,
		Group:          "default",
		CategoryId:     "cat-id",
		PriorityId:     "pri-id",
	}
}

func newTestWorkloadTypeWithSourceForService(name string, category string, priority string, preemptibility openapi.WorkloadTypePreemptibility, source string) openapi.WorkloadTypeConfig {
	wt := newTestWorkloadTypeForService(name, category, priority, preemptibility)
	wt.Source = &source
	return wt
}
