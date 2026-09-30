/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/describe"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockLister struct {
	response *openapi.ListWorkloadTypes200Response
	err      error
}

func (m *mockLister) List(_ context.Context, _ []string, _ *bool) (*openapi.ListWorkloadTypes200Response, error) {
	return m.response, m.err
}

type mockGetter struct {
	response *openapi.WorkloadTypeConfig
	err      error
}

func (m *mockGetter) Get(_ context.Context, _ string) (*openapi.WorkloadTypeConfig, error) {
	return m.response, m.err
}

func strPtr(s string) *string {
	return &s
}

var _ = Describe("Service", func() {
	var (
		ctx    context.Context
		lister *mockLister
		getter *mockGetter
		svc    *describe.Service
	)

	BeforeEach(func() {
		ctx = context.Background()
		lister = &mockLister{}
		getter = &mockGetter{}
		svc = describe.NewService(lister, getter)
	})

	Describe("Describe", func() {
		Context("when the workload type exists", func() {
			BeforeEach(func() {
				lister.response = &openapi.ListWorkloadTypes200Response{
					Types: []openapi.WorkloadTypeConfig{
						{
							Name:           "Training",
							Id:             strPtr("wt-123"),
							CategoryName:   strPtr("batch"),
							PriorityName:   strPtr("high"),
							Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
							Group:          "run.ai",
						},
					},
				}
				getter.response = &openapi.WorkloadTypeConfig{
					Name:           "Training",
					Id:             strPtr("wt-123"),
					CategoryName:   strPtr("batch"),
					PriorityName:   strPtr("high"),
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
					Group:          "run.ai",
				}
			})

			It("should return the workload type", func() {
				result, err := svc.Describe(ctx, "Training")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Name).To(Equal("Training"))
				Expect(*result.Id).To(Equal("wt-123"))
			})
		})

		Context("when using case-insensitive name match", func() {
			BeforeEach(func() {
				lister.response = &openapi.ListWorkloadTypes200Response{
					Types: []openapi.WorkloadTypeConfig{
						{
							Name:           "Training",
							Id:             strPtr("wt-123"),
							Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
							Group:          "run.ai",
						},
					},
				}
				getter.response = &openapi.WorkloadTypeConfig{
					Name:           "Training",
					Id:             strPtr("wt-123"),
					Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
					Group:          "run.ai",
				}
			})

			It("should match regardless of case", func() {
				result, err := svc.Describe(ctx, "training")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Name).To(Equal("Training"))
			})
		})

		Context("when the workload type is not found", func() {
			BeforeEach(func() {
				lister.response = &openapi.ListWorkloadTypes200Response{
					Types: []openapi.WorkloadTypeConfig{
						{
							Name:           "Training",
							Id:             strPtr("wt-123"),
							Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
							Group:          "run.ai",
						},
					},
				}
			})

			It("should return a not found error", func() {
				result, err := svc.Describe(ctx, "NonExistent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(result).To(BeNil())
			})
		})

		Context("when the list API returns an error", func() {
			BeforeEach(func() {
				lister.err = errors.New("connection refused")
			})

			It("should propagate the error", func() {
				result, err := svc.Describe(ctx, "Training")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to list workload types"))
				Expect(result).To(BeNil())
			})
		})

		Context("when the get API returns an error", func() {
			BeforeEach(func() {
				lister.response = &openapi.ListWorkloadTypes200Response{
					Types: []openapi.WorkloadTypeConfig{
						{
							Name:           "Training",
							Id:             strPtr("wt-123"),
							Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
							Group:          "run.ai",
						},
					},
				}
				getter.err = errors.New("server error")
			})

			It("should propagate the error", func() {
				result, err := svc.Describe(ctx, "Training")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get workload type"))
				Expect(result).To(BeNil())
			})
		})

		Context("when the list is empty", func() {
			BeforeEach(func() {
				lister.response = &openapi.ListWorkloadTypes200Response{
					Types: []openapi.WorkloadTypeConfig{},
				}
			})

			It("should return a not found error", func() {
				result, err := svc.Describe(ctx, "Training")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(result).To(BeNil())
			})
		})
	})
})
