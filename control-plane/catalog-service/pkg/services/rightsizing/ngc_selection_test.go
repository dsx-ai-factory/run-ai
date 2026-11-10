// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"encoding/json"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU-agnostic exact-tag sizing", func() {
	for _, gpuType := range []string{"NVIDIA-A100-SXM4-40GB", "NVIDIA-A100-SXM4-80GB", "NVIDIA-H100-80GB-HBM3", "NVIDIA-B200", "unmapped-product"} {
		It("uses the generic profile for "+gpuType, func() {
			data, err := os.ReadFile("testdata/llama-3.1-8b-instruct-2.0.13.json")
			Expect(err).NotTo(HaveOccurred())
			var metadata ngcmetadata.Metadata
			Expect(json.Unmarshal(data, &metadata)).To(Succeed())
			capacity := int64(80_000_000_000)
			result := recommendForHardware(gpuinventory.Hardware{Type: gpuType, MemoryBytes: &capacity, MaxDevicesPerNode: 8}, metadata.Profiles, 90)
			Expect(string(result.Status)).To(Equal("recommended"))
			Expect(result.ProfileId).To(Equal("20739fe3e660205a0a64339c06bf2aa5c90f4935dd5e547b2b22d96f9c91fbc4"))
			Expect(result.Precision).To(Equal("nvfp4"))
			Expect(result.InferenceBackend).To(Equal("vllm"))
			Expect(result.MinimumGpuMemoryPerDeviceGb).To(Equal(24.0))
		})
	}
})

func sizingProfile(id string, minimum float64) ngcmetadata.Profile {
	one := 1
	return ngcmetadata.Profile{ID: id, MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}
}

var _ = Describe("GPU-agnostic profile applicability", func() {
	DescribeTable("does not infer hardware identity from product names or tested devices",
		func(targetGPU, targetDevice string, testedDevices []string, eligible bool) {
			profile := sizingProfile("published", 33)
			profile.GPU, profile.GPUDevice, profile.TestedGPUDevices = targetGPU, targetDevice, testedDevices
			capacity := int64(80_000_000_000)
			result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-A100-SXM4-80GB", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{profile}, 90)
			if eligible {
				Expect(string(result.Status)).To(Equal("recommended"))
				Expect(result.ProfileId).To(Equal("published"))
				Expect(result.Recommendation).NotTo(BeNil())
			} else {
				Expect(string(result.Status)).To(Equal("no_recommendation"))
				Expect(result.Reason).To(ContainSubstring("GPU-agnostic"))
				Expect(result.Recommendation).To(BeNil())
			}
		},
		Entry("no hardware fields", "", "", nil, true),
		Entry("empty tested-device list", "", "", []string{}, true),
		Entry("matching tested device does not change eligibility", "", "", []string{"20b2:10de"}, true),
		Entry("unrelated tested device does not restrict a generic profile", "", "", []string{"2901:10de"}, true),
		Entry("matching product name remains explicitly targeted", "NVIDIA-A100-SXM4-80GB", "", nil, false),
		Entry("matching device ID remains explicitly targeted", "", "20b2:10de", nil, false),
		Entry("matching tested device does not enable a targeted profile", "another-product", "", []string{"20b2:10de"}, false),
	)
})

var _ = It("keeps inventory GPU type filtering separate from profile selection", func() {
	small, large := int64(40_000_000_000), int64(80_000_000_000)
	generic := sizingProfile("generic", 60)
	targeted := sizingProfile("targeted", 10)
	targeted.GPUDevice = "20b2:10de"
	service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		return []gpuinventory.Hardware{
			{Type: "NVIDIA-A100-SXM4-80GB", MemoryBytes: &small, MaxDevicesPerNode: 8},
			{Type: "NVIDIA-A100-SXM4-80GB", MemoryBytes: &large, MaxDevicesPerNode: 8},
			{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &large, MaxDevicesPerNode: 8},
			{Type: "unmapped-product", MemoryBytes: &large, MaxDevicesPerNode: 8},
		}, nil
	}), metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, *api.Error) {
		return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{targeted, generic}}, nil
	}), 90)
	request := validRequest()
	response, err := service.Recommend(context.Background(), request)
	Expect(err).To(BeNil())
	Expect(response.Results).To(HaveLen(4))
	Expect(string(response.Results[0].Status)).To(Equal("no_recommendation"))
	for _, result := range response.Results[1:] {
		Expect(result.ProfileId).To(Equal("generic"))
	}
	gpuType := "NVIDIA-A100-SXM4-80GB"
	request.GpuType = &gpuType
	response, err = service.Recommend(context.Background(), request)
	Expect(err).To(BeNil())
	Expect(response.Results).To(HaveLen(2))
	Expect(string(response.Results[0].Status)).To(Equal("no_recommendation"))
	Expect(response.Results[1].ProfileId).To(Equal("generic"))
	for _, result := range response.Results {
		Expect(result.GpuType).To(Equal(gpuType))
	}
})

var _ = Describe("recommendation allocation boundaries", func() {
	capacity := int64(100_000_000_000)
	for _, tc := range []struct {
		name          string
		minimum       float64
		cutoff        int64
		allocation    string
		count, memory int64
	}{
		{"below 90 percent", 89.999999999, 90, "gpu_memory", 1, 89_999_999_999},
		{"exactly 90 percent", 90, 90, "gpu_count", 1, 100_000_000_000},
		{"above 90 percent", 91, 90, "gpu_count", 1, 100_000_000_000},
		{"exactly capacity", 100, 90, "gpu_count", 1, 100_000_000_000},
		{"configured 80 percent", 85, 80, "gpu_count", 1, 100_000_000_000},
	} {
		It(tc.name, func() {
			r := recommendForHardware(gpuinventory.Hardware{Type: "unknown", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{sizingProfile("generic", tc.minimum)}, tc.cutoff)
			a := r.Recommendation
			Expect(string(r.Status)).To(Equal("recommended"))
			Expect(r.MinimumGpuMemoryPerDeviceGb).To(Equal(tc.minimum))
			Expect(a).NotTo(BeNil())
			Expect(string(a.AllocationType)).To(Equal(tc.allocation))
			Expect(a.GpuCount).To(Equal(tc.count))
			Expect(a.GpuMemoryPerDeviceBytes).To(Equal(tc.memory))
		})
	}
})
