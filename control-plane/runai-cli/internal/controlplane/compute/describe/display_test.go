/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/describe"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("ComputeFormatter", func() {
	Context("NewComputeFormatter", func() {
		It("should create formatter successfully with valid compute", func() {
			compute := &openapi.ComputeAsset{
				Meta: openapi.AssetMeta{
					Name: "test-compute",
				},
			}

			formatter, err := describe.NewComputeFormatter(compute)

			Expect(err).ToNot(HaveOccurred())
			Expect(formatter).ToNot(BeNil())
		})

		It("should return error for nil compute", func() {
			formatter, err := describe.NewComputeFormatter(nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("compute is nil"))
			Expect(formatter).To(BeNil())
		})
	})

	Context("Displayer", func() {
		It("should return valid displayer", func() {
			compute := &openapi.ComputeAsset{
				Meta: openapi.AssetMeta{
					Name: "test-compute",
				},
			}
			formatter, _ := describe.NewComputeFormatter(compute)

			displayer := formatter.Displayer()

			Expect(displayer).ToNot(BeNil())
		})
	})
})

var _ = Describe("ComputeDisplayer", func() {
	var (
		compute   *openapi.ComputeAsset
		displayer *describe.ComputeDisplayer
	)

	BeforeEach(func() {
		compute = newTestComputeAsset()
		displayer = describe.NewComputeDisplayer(compute)
	})

	Context("NewComputeDisplayer", func() {
		It("should create displayer successfully", func() {
			Expect(displayer).ToNot(BeNil())
		})
	})

	Context("AsTable", func() {
		It("should not return error", func() {
			err := displayer.AsTable()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("AsJSON", func() {
		It("should generate valid JSON", func() {
			err := displayer.AsJSON()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("AsYAML", func() {
		It("should generate valid YAML", func() {
			err := displayer.AsYAML()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("ToComputeAsset", func() {
	Context("Converting OpenAPI compute to output model", func() {
		It("should convert compute with all fields", func() {
			compute := newTestComputeAsset()

			result := describe.ToComputeAsset(compute)

			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("test-compute"))
			Expect(result.Meta.Scope).To(Equal("Project"))
			Expect(result.Spec.GpuDevicesRequest).ToNot(BeNil())
			Expect(*result.Spec.GpuDevicesRequest).To(Equal(int32(8)))
			Expect(result.Spec.GpuMemoryRequest).To(Equal("80GB"))
			Expect(result.Spec.CpuMemoryRequest).To(Equal("128GB"))
		})

		It("should handle nil compute", func() {
			result := describe.ToComputeAsset(nil)
			Expect(result).To(BeNil())
		})

		It("should omit empty fields in output", func() {
			compute := &openapi.ComputeAsset{
				Meta: openapi.AssetMeta{
					Name:      "minimal-compute",
					Scope:     openapi.Scope("Project"),
					CreatedBy: "test-user",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Spec: openapi.ComputeAssetSpec{},
			}

			result := describe.ToComputeAsset(compute)

			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Name).To(Equal("minimal-compute"))
			Expect(result.Spec.GpuDevicesRequest).To(BeNil())
			Expect(result.Spec.GpuMemoryRequest).To(BeEmpty())
			Expect(result.Spec.CpuMemoryRequest).To(BeEmpty())
		})

		It("should handle optional meta fields", func() {
			compute := &openapi.ComputeAsset{
				Meta: openapi.AssetMeta{
					Name:        "test-compute",
					Scope:       openapi.Scope("Project"),
					Description: *openapi.NewNullableString(openapi.PtrString("Test description")),
					ProjectName: *openapi.NewNullableString(openapi.PtrString("test-project")),
					CreatedBy:   "test-user",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Spec: openapi.ComputeAssetSpec{},
			}

			result := describe.ToComputeAsset(compute)

			Expect(result).ToNot(BeNil())
			Expect(result.Meta.Description).To(Equal("Test description"))
			Expect(result.Meta.ProjectName).To(Equal("test-project"))
		})

		It("should handle extended resources", func() {
			compute := &openapi.ComputeAsset{
				Meta: openapi.AssetMeta{
					Name:      "test-compute",
					Scope:     openapi.Scope("Project"),
					CreatedBy: "test-user",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Spec: openapi.ComputeAssetSpec{},
			}
			resource1 := openapi.ExtendedResource{}
			resource1.SetResource("rdma/rdma_shared_device")
			resource1.SetQuantity("1")
			resource2 := openapi.ExtendedResource{}
			resource2.SetResource("nvidia.com/gpu")
			resource2.SetQuantity("2")
			compute.Spec.ExtendedResources = []openapi.ExtendedResource{resource1, resource2}

			result := describe.ToComputeAsset(compute)

			Expect(result).ToNot(BeNil())
			Expect(result.Spec.ExtendedResources).To(HaveLen(2))
			Expect(result.Spec.ExtendedResources[0].Name).To(Equal("rdma/rdma_shared_device"))
			Expect(result.Spec.ExtendedResources[0].Quantity).To(Equal("1"))
		})
	})
})

// Helper function to create test compute asset for describe tests
func newTestComputeAsset() *openapi.ComputeAsset {
	// Create spec with all fields
	spec := openapi.ComputeAssetSpec{}
	spec.SetGpuDevicesRequest(8)
	gpuRequestType := openapi.GPUREQUESTTYPE_MEMORY
	spec.SetGpuRequestType(gpuRequestType)
	spec.SetGpuMemoryRequest("80GB")
	spec.SetGpuMemoryLimit("80GB")
	spec.SetCpuCoreRequest(32.0)
	spec.SetCpuCoreLimit(64.0)
	spec.SetCpuMemoryRequest("128GB")
	spec.SetCpuMemoryLimit("256GB")
	spec.SetLargeShmRequest(true)

	resource1 := openapi.ExtendedResource{}
	resource1.SetResource("rdma/rdma_shared_device")
	resource1.SetQuantity("1")
	spec.ExtendedResources = []openapi.ExtendedResource{resource1}

	projectId := int32(123)
	projectName := "test-project"

	return &openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Name:        "test-compute",
			Scope:       openapi.Scope("Project"),
			CreatedBy:   "test-user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ProjectId:   *openapi.NewNullableInt32(&projectId),
			ProjectName: *openapi.NewNullableString(&projectName),
		},
		Spec: spec,
	}
}
