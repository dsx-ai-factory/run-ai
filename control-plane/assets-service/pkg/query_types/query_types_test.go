/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package query_types

import (
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Types Suite")
}

var _ = Describe("Query Types Tests", func() {
	Describe("crud test", func() {
		BeforeEach(func() {
		})

		Context("invalid scope value", func() {
			It("list show both of them", func() {
				_, err := NewQueryParams("", "", "invalid", false, 0, "", "", nil)
				Expect(err).NotTo(BeNil())
				Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
				Expect(err.Message).To(ContainSubstring("invalid value provided for scope"))
			})
		})

		Context("empty query", func() {
			It("should return valid query params", func() {
				query, err := NewQueryParams("", "", "", false, 0, "", "", nil)
				Expect(err).To(BeNil())
				Expect(query.Name).To(BeNil())
				Expect(query.ProjectId).To(BeNil())
				Expect(query.Scope).To(BeNil())
				Expect(query.AssetIds).To(BeNil())
			})
		})

		Context("valid combination", func() {
			It("should return valid query params", func() {
				name := "aName"
				assets := "550e8400-e29b-41d4-a716-446655440000,ce7ed135-b751-4a55-8e03-3180be899e4c"

				query, err := NewQueryParams(assets, name, string(api.SCOPE_SYSTEM), false, 0, "", "", nil)
				Expect(err).To(BeNil())
				Expect(query.Name).ToNot(BeNil())
				Expect(*query.Name).To(Equal(name))
				Expect(query.ProjectId).To(BeNil())
				Expect(query.DepartmentId).To(BeNil())
				Expect(query.ClusterId).To(BeNil())
				Expect(*query.Scope).To(Equal(api.SCOPE_SYSTEM))
				Expect(*query.AssetIds).To(Equal([]string{"550e8400-e29b-41d4-a716-446655440000", "ce7ed135-b751-4a55-8e03-3180be899e4c"}))

				query, err = NewQueryParams(assets, name, string(api.SCOPE_TENANT), false, 0, "", "", nil)
				Expect(err).To(BeNil())
				Expect(query.Name).ToNot(BeNil())
				Expect(*query.Name).To(Equal(name))
				Expect(query.ProjectId).To(BeNil())
				Expect(query.DepartmentId).To(BeNil())
				Expect(query.ClusterId).To(BeNil())
				Expect(*query.Scope).To(Equal(api.SCOPE_TENANT))
				Expect(*query.AssetIds).To(Equal([]string{"550e8400-e29b-41d4-a716-446655440000", "ce7ed135-b751-4a55-8e03-3180be899e4c"}))

				query, err = NewQueryParams(assets, name, string(api.SCOPE_CLUSTER), false, 0, "", "123", nil)
				Expect(err).To(BeNil())
				Expect(query.Name).ToNot(BeNil())
				Expect(*query.Name).To(Equal(name))
				Expect(query.ProjectId).To(BeNil())
				Expect(query.DepartmentId).To(BeNil())
				Expect(*query.ClusterId).To(Equal("123"))
				Expect(*query.Scope).To(Equal(api.SCOPE_CLUSTER))
				Expect(*query.AssetIds).To(Equal([]string{"550e8400-e29b-41d4-a716-446655440000", "ce7ed135-b751-4a55-8e03-3180be899e4c"}))

				query, err = NewQueryParams(assets, name, string(api.SCOPE_PROJECT), false, 0, "123", "", nil)
				Expect(err).To(BeNil())
				Expect(query.Name).ToNot(BeNil())
				Expect(*query.Name).To(Equal(name))
				Expect(query.ProjectId).To(BeNil())
				Expect(query.DepartmentId).NotTo(BeNil())
				Expect(*query.DepartmentId).To(Equal("123"))
				Expect(query.ClusterId).To(BeNil())
				Expect(*query.Scope).To(Equal(api.SCOPE_PROJECT))
				Expect(*query.AssetIds).To(Equal([]string{"550e8400-e29b-41d4-a716-446655440000", "ce7ed135-b751-4a55-8e03-3180be899e4c"}))
			})
		})

		Context("valid query type", func() {
			name := "aName"

			It("valid training", func() {
				query, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					IsTraining: true,
				})
				Expect(err).To(BeNil())
				Expect(query.DistributedQueryParams.IsTraining).To(Equal(true))
			})

			It("invalid training", func() {
				query, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					IsInference: true,
				})
				Expect(err).To(BeNil())
				Expect(query.DistributedQueryParams.IsInference).To(Equal(true))
			})

			It("valid distributed params", func() {
				query, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					DistributedFramework: "PyTorch",
					IsDistributed:        true,
				})
				Expect(err).To(BeNil())
				Expect(query.DistributedQueryParams.IsDistributed).To(Equal(true))
				Expect(query.DistributedQueryParams.DistributedFramework).To(Equal("PyTorch"))
			})

			It("valid workspace", func() {
				query, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					IsWorkspace: true,
				})
				Expect(err).To(BeNil())
				Expect(query.DistributedQueryParams.IsWorkspace).To(Equal(true))
			})

			It("valid distributed inference", func() {
				query, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					IsDistributedInference: true,
				})
				Expect(err).To(BeNil())
				Expect(query.DistributedQueryParams.IsDistributedInference).To(Equal(true))
			})
		})

		Context("invalid query type", func() {
			name := "aName"

			It("invalid distributedFramework", func() {
				_, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					DistributedFramework: "xyz",
					IsTraining:           true,
				})
				Expect(err).ToNot(BeNil())
				Expect(err.Message).To(Equal("Invalid query parameters - cannot use distributedFramework without isDistributed set to true"))
			})

			It("multi types not supported", func() {
				_, err := NewQueryParams("", name, string(api.SCOPE_PROJECT), false, 0, "", "", &DistributedQueryParams{
					DistributedFramework: "xyz",
					IsDistributed:        true,
				})
				Expect(err).ToNot(BeNil())
				Expect(err.Message).To(Equal("Invalid query parameters (distributedFramework)"))
			})
		})
	})

	Describe("testing load kind", func() {
		Context("no kind given", func() {
			It("should return nil", func() {
				kind, err := LoadKindParameter("")
				Expect(err).To(BeNil())
				Expect(kind).To(BeNil())
			})
		})

		Context("incorrect kind given", func() {
			It("should return error", func() {
				_, err := LoadKindParameter("xxxyyyzzz")
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
