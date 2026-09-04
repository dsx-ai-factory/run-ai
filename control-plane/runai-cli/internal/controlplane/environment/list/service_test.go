/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/list"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		service    *list.Service
		mockClient *cpmocks.MockEnvironmentLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockEnvironmentLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return environments successfully", func() {
			env1 := newTestEnvironment("env-1", "Cluster", "ubuntu:20.04")
			env2 := newTestEnvironment("env-2", "Project", "pytorch:latest")
			mockResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env1, env2},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return error when API call fails", func() {
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			result, err := service.List(ctx, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("environment API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty environment list", func() {
			mockResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply project filter correctly", func() {
			env1 := newTestEnvironment("env-1", "Cluster", "ubuntu:20.04")
			mockResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			projectID := int32(123)
			result, err := service.List(ctx, &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply cluster filter correctly", func() {
			env1 := newTestEnvironment("env-1", "Cluster", "ubuntu:20.04")
			mockResponse := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{env1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			clusterID := "cluster-123"
			result, err := service.List(ctx, nil, &clusterID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})

})

// Helper function to create test environment assets
func newTestEnvironment(name, scope, image string) openapi.EnvironmentAsset {
	spec := openapi.EnvironmentAssetSpec{}
	if image != "" {
		spec.SetImage(image)
	}

	return openapi.EnvironmentAsset{
		Meta: openapi.AssetMeta{
			Id:    "env-id-" + name,
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}
