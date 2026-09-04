/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"context"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/describe"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FindByName", func() {
	var (
		mockCtrl   *gomock.Controller
		mockLister *cpmocks.MockEnvironmentDescribeLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockEnvironmentDescribeLister(mockCtrl)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("when environment exists", func() {
		It("should find environment by name with project filter", func() {
			env := newTestEnvironmentForList("test-env", "env-id-1", "Project", "ubuntu:20.04")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-env", &projectID, nil, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-env"))
			Expect(result.Meta.Id).To(Equal("env-id-1"))
		})

		It("should find environment by name with cluster filter", func() {
			env := newTestEnvironmentForList("test-env", "env-id-1", "Cluster", "ubuntu:20.04")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			clusterID := "cluster-123"
			result, err := describe.FindByName(ctx, mockLister, "test-env", nil, &clusterID, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-env"))
		})

		It("should find environment by name with workload type filter", func() {
			env := newTestEnvironmentForList("test-env", "env-id-1", "Project", "ubuntu:20.04")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-env", &projectID, nil, "Workspace")

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("when environment does not exist", func() {
		It("should return error when environment not found", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "nonexistent-env", &projectID, nil, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(result).To(BeNil())
		})

		It("should return workload type specific error when not found with workload filter", func() {
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-env", &projectID, nil, "Workspace")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not support"))
			Expect(result).To(BeNil())
		})
	})

	Context("when list fails", func() {
		It("should return error when list API fails", func() {
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("list error"))

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-env", &projectID, nil, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to search for environment"))
			Expect(result).To(BeNil())
		})
	})

	Context("when multiple environments match", func() {
		It("should return the first matching environment", func() {
			env1 := newTestEnvironmentForList("test-env", "env-id-1", "Cluster", "ubuntu:20.04")
			env2 := newTestEnvironmentForList("test-env", "env-id-2", "Project", "pytorch:latest")
			listResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env1, env2},
			}
			mockLister.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(listResponse, nil)

			projectID := int32(123)
			result, err := describe.FindByName(ctx, mockLister, "test-env", &projectID, nil, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Id).To(Equal("env-id-1"))
		})
	})
})

// Helper function to create test environment assets
func newTestEnvironmentForList(name, id, scope, image string) openapi.EnvironmentAsset {
	spec := openapi.EnvironmentAssetSpec{}
	if image != "" {
		spec.SetImage(image)
	}

	return openapi.EnvironmentAsset{
		Meta: openapi.AssetMeta{
			Id:    id,
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}
