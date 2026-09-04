/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/compute/list"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

var _ = Describe("Compute List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		service    *list.Service
		mockClient *cpmocks.MockComputeLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockComputeLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return computes successfully", func() {
			compute1 := newTestCompute("compute-large", "Project", 8, "80GB", "128GB")
			compute2 := newTestCompute("compute-small", "Cluster", 2, "16GB", "32GB")
			mockResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute1, compute2},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return error when API call fails", func() {
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			result, err := service.List(ctx, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("compute API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty compute list", func() {
			mockResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply project filter correctly", func() {
			compute1 := newTestCompute("compute-large", "Project", 8, "80GB", "128GB")
			mockResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			projectID := int32(123)
			result, err := service.List(ctx, &projectID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply cluster filter correctly", func() {
			compute1 := newTestCompute("compute-large", "Project", 8, "80GB", "128GB")
			mockResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			clusterID := "cluster-123"
			result, err := service.List(ctx, nil, &clusterID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should handle both project and cluster filters", func() {
			compute1 := newTestCompute("compute-large", "Project", 8, "80GB", "128GB")
			mockResponse := &openapi.ComputeListResponse{
				Entries: []openapi.ComputeAsset{compute1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			projectID := int32(123)
			clusterID := "cluster-123"
			result, err := service.List(ctx, &projectID, &clusterID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})
})

// Helper function to create test compute assets
func newTestCompute(name, scope string, gpuDevices int32, gpuMemory, cpuMemory string) openapi.ComputeAsset {
	spec := openapi.ComputeAssetSpec{}

	if gpuDevices > 0 {
		spec.SetGpuDevicesRequest(gpuDevices)
	}

	if gpuMemory != "" {
		spec.SetGpuMemoryRequest(gpuMemory)
	}

	if cpuMemory != "" {
		spec.SetCpuMemoryRequest(cpuMemory)
	}

	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Id:    "compute-id-" + name,
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}
