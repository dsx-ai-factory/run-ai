/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type inventoryFunc func(context.Context, string) ([]gpuinventory.Hardware, *api.Error)

func (f inventoryFunc) Get(ctx context.Context, id string) ([]gpuinventory.Hardware, *api.Error) {
	return f(ctx, id)
}

type metadataFunc func(context.Context, string, string) (ngcmetadata.Metadata, *api.Error)

func (f metadataFunc) Get(ctx context.Context, image, tag string) (ngcmetadata.Metadata, *api.Error) {
	return f(ctx, image, tag)
}

var _ = It("recommend keeps capacities and applies cutoff", func() {
	small, large, minimum, one := int64(40_000_000_000), int64(80_000_000_000), 36.0, 1
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		return []gpuinventory.Hardware{{Type: "NVIDIA-A100", MemoryBytes: &small, MaxDevicesPerNode: 8}, {Type: "NVIDIA-A100", MemoryBytes: &large, MaxDevicesPerNode: 8}, {Type: "T4"}}, nil
	}), metadataFunc(func(_ context.Context, image, tag string) (ngcmetadata.Metadata, *api.Error) {
		Expect(image).To(Equal("nvcr.io/nim/nvidia/example"))
		Expect(tag).To(Equal("2.1.4"))
		return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{{ID: "profile", Precision: "bf16", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}}, nil
	}), 90)
	result, err := s.Recommend(context.Background(), validRequest())
	Expect(err).To(BeNil())
	Expect(result.Results).To(HaveLen(3))
	Expect(string(result.Results[0].Recommendation.AllocationType)).To(Equal("gpu_count"))
	Expect(string(result.Results[1].Recommendation.AllocationType)).To(Equal("gpu_memory"))
	Expect(result.Results[1].Recommendation.GpuMemoryPerDeviceBytes).To(BeNumerically("==", 36_000_000_000))
	Expect(string(result.Results[2].Status)).To(Equal("no_recommendation"))
	gpu := "NVIDIA-A100"
	request := validRequest()
	request.GpuType = &gpu
	filtered, err := s.Recommend(context.Background(), request)
	Expect(err).To(BeNil())
	Expect(filtered.Results).To(HaveLen(2))
})

var _ = It("recommend uses multi device minimum per device", func() {
	capacity, minimum, tp, pp := int64(80_000_000_000), 35.0, 4, 1
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}}, nil
	}),
		metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, *api.Error) {
			return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{{ID: "tp4", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &tp, PipelineParallelSize: &pp}}}, nil
		}), 90)
	result, err := s.Recommend(context.Background(), validRequest())
	Expect(err).To(BeNil())
	allocation := result.Results[0].Recommendation
	Expect(allocation).NotTo(BeNil())
	Expect(allocation.GpuCount).To(BeNumerically("==", 4))
	Expect(allocation.GpuMemoryPerDeviceBytes).To(Equal(capacity))
	Expect(result.Results[0].MinimumGpuMemoryPerDeviceGb).To(Equal(minimum))
})

var _ = It("recommend missing metadata and upstream errors", func() {
	capacity := int64(80_000_000_000)
	inventory := inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}}, nil
	})
	s := New(inventory, metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, *api.Error) {
		return ngcmetadata.Metadata{}, nil
	}), 90)
	result, err := s.Recommend(context.Background(), validRequest())
	Expect(err).To(BeNil())
	Expect(string(result.Results[0].Status)).To(Equal("no_recommendation"))
	failure := api.NewError(http.StatusServiceUnavailable, "upstream unavailable", nil)
	s = New(inventory, metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, *api.Error) {
		return ngcmetadata.Metadata{}, failure
	}), 90)
	{
		_, err = s.Recommend(context.Background(), validRequest())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(BeNumerically("==", 503))
	}
	s = New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) { return nil, failure }), nil, 90)
	{
		_, err = s.Recommend(context.Background(), validRequest())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(BeNumerically("==", 503))
	}
})

var _ = It("recommend rejects invalid context before lookup", func() {
	s := New(nil, nil, 90)
	request := validRequest()
	request.ClusterId = "invalid"
	{
		_, err := s.Recommend(context.Background(), request)
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(BeNumerically("==", 400))
	}
})

func validRequest() api.RecommendationRequest {
	return api.RecommendationRequest{ClusterId: "4b94efde-b05a-4118-a03a-1dedff81fa8e", Image: "nvcr.io/nim/nvidia/example", Tag: "2.1.4"}
}

var _ = Describe("rejects unverified targeted profiles", func() {
	capacity, minimum, one := int64(80_000_000_000), 36.0, 1
	for _, tc := range []struct{ name, gpu, device string }{
		{"display name only", "A100", ""}, {"device target only", "", "20b2:10de"}, {"GPU and device target", "A100", "ffff:10de"},
	} {
		It(tc.name, func() {
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{{ID: "targeted", GPU: tc.gpu, GPUDevice: tc.device, MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}, 90)
			Expect(string(result.Status)).To(Equal("no_recommendation"))
			Expect(result.Recommendation).To(BeNil())
		})
	}
})

var _ = It("oversized single device baseline does not produce estimate", func() {
	capacity, minimum, one := int64(40_000_000_000), 81.0, 1
	result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{{ID: "generic", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}, 90)
	Expect(string(result.Status)).To(Equal("no_recommendation"))
	Expect(result.Recommendation).To(BeNil())
	Expect(result.Reason).NotTo(BeEmpty())
})

var _ = Describe("profile selection requires evidence and is deterministic", func() {
	capacity, one, minimum, larger := int64(80_000_000_000), 1, 16.0, 32.0
	baseline := ngcmetadata.Profile{ID: "b", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}
	for _, tc := range []struct {
		name     string
		profiles []ngcmetadata.Profile
		want     string
	}{
		{"missing memory", []ngcmetadata.Profile{{ID: "unknown", TensorParallelSize: &one, PipelineParallelSize: &one}}, ""},
		{"missing parallelism", []ngcmetadata.Profile{{ID: "unknown", MinimumVRAMPerDeviceGB: &minimum}}, ""},
		{"smallest memory then stable profile ID", []ngcmetadata.Profile{baseline, {ID: "larger", MinimumVRAMPerDeviceGB: &larger, TensorParallelSize: &one, PipelineParallelSize: &one}, {ID: "a", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}, "a"},
		{"tested devices are history", []ngcmetadata.Profile{{ID: "generic", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one, TestedGPUDevices: []string{"different-pci"}}}, "generic"},
	} {
		It(tc.name, func() {
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, tc.profiles, 90)
			Expect(result.ProfileId).To(Equal(tc.want))
			if tc.want == "" {
				Expect(result.Recommendation).To(BeNil())
			}
		})
	}
})

var _ = It("unavailable requested GPU fails before metadata", func() {
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) { return nil, nil }), nil, 90)
	request := validRequest()
	missing := "missing"
	request.GpuType = &missing
	{
		_, err := s.Recommend(context.Background(), request)
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(BeNumerically("==", 400))
	}
	request.GpuType = nil
	response, err := s.Recommend(context.Background(), request)
	Expect(err).To(BeNil())
	Expect(response.Results).NotTo(BeNil())
	Expect(response.Results).To(HaveLen(0))
})
