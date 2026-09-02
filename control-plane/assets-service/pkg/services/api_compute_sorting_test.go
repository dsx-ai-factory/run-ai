/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

var _ = Describe("compute asset sort order", func() {
	createdAssets := make([]api.ComputeCreationRequest, 0)

	BeforeEach(func() {
		for i := 0; i < 9; i++ {
			createdAssets = append(createdAssets, buildRandomTenantCompute())
		}

		createdAssets[0].Meta.Name = "10MB"
		createdAssets[0].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
		createdAssets[0].Spec.GpuMemoryRequest = convutils.Ref("10MB")
		createdAssets[2].Meta.Name = "2GB"
		createdAssets[2].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
		createdAssets[2].Spec.GpuMemoryRequest = convutils.Ref("2GB")
		createdAssets[5].Meta.Name = "500MB"
		createdAssets[5].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
		createdAssets[5].Spec.GpuMemoryRequest = convutils.Ref("500MB")

		createdAssets[1].Meta.Name = "0.5"
		createdAssets[1].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
		createdAssets[1].Spec.GpuPortionRequest = convutils.Ref(0.5)
		createdAssets[3].Meta.Name = "0"
		createdAssets[3].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
		createdAssets[3].Spec.GpuPortionRequest = convutils.Ref(0.0)
		createdAssets[6].Meta.Name = "4"
		createdAssets[6].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
		createdAssets[6].Spec.GpuPortionRequest = convutils.Ref(4.0)
		createdAssets[8].Meta.Name = "0.8"
		createdAssets[8].Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
		createdAssets[8].Spec.GpuPortionRequest = convutils.Ref(0.8)

		createdAssets[4].Meta.Name = "no-gpu-b"
		createdAssets[4].Spec.GpuRequestType = nil
		createdAssets[7].Meta.Name = "no-gpu-a"
		createdAssets[7].Spec.GpuRequestType = nil
	})

	Context("sort the created assets", func() {
		It("should sort them by gpu request", func() {
			assets := make([]api.ComputeAsset, 0)
			for _, asset := range createdAssets {
				assets = append(assets, api.ComputeAsset{
					Meta: api.AssetMeta{
						Name: asset.Meta.Name,
					},
					Spec: convutils.Deref(asset.Spec, api.ComputeAssetSpec{}),
				})
			}

			expectedNames := []string{
				"no-gpu-a",
				"no-gpu-b",
				"10MB",
				"500MB",
				"2GB",
				"0",
				"0.5",
				"0.8",
				"4",
			}

			SortComputeAssets(assets)

			actualNames := make([]string, 0)
			for _, asset := range assets {
				actualNames = append(actualNames, asset.Meta.Name)
			}

			Expect(actualNames).To(Equal(expectedNames))
		})
	})
})
