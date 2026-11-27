/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"context"
	"math"
	"net/http"
	"testing"

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

var _ = Describe("recommendation decision boundaries", func() {
	It("retains selected evidence when single-GPU allocation fails without selecting a parallel fallback", func() {
		capacity := int64(80_000_000_000)
		result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{publishedParallelProfile("parallel", 10, 2, 1), publishedParallelProfile("single", 20, 1, 1)}, 0)
		Expect(result).To(Equal(api.RecommendationResult{
			GpuType: "A100", GpuMemoryCapacityBytes: capacity,
			Status:    api.RECOMMENDATIONSTATUS_NO_RECOMMENDATION,
			ProfileId: "single", MinimumGpuMemoryPerDeviceGb: 20,
			Precision: "bf16", InferenceBackend: "vllm", TensorParallelSize: 1, PipelineParallelSize: 1,
			Reason: "GPU allocation sharing policy is invalid",
		}))
	})

	It("does not apply the single-GPU sharing cutoff to parallel allocation", func() {
		capacity := int64(80_000_000_000)
		result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{publishedParallelProfile("parallel", 20, 3, 2)}, 0)
		Expect(result).To(Equal(api.RecommendationResult{
			GpuType: "A100", GpuMemoryCapacityBytes: capacity,
			Status:    api.RECOMMENDATIONSTATUS_RECOMMENDED,
			ProfileId: "parallel", MinimumGpuMemoryPerDeviceGb: 20,
			Precision: "bf16", InferenceBackend: "vllm", TensorParallelSize: 3, PipelineParallelSize: 2,
			Recommendation: &api.Recommendation{AllocationType: api.RECOMMENDATIONALLOCATIONTYPE_GPU_COUNT,
				GpuCount: 6, GpuMemoryPerDeviceBytes: capacity},
		}))
	})

	It("ranks original GB minima even when they round to the same byte count", func() {
		capacity := int64(80_000_000_000)
		lower := publishedParallelProfile("z-lower", 20.0000000001, 1, 1)
		higher := publishedParallelProfile("a-higher", 20.0000000002, 1, 1)
		for _, profiles := range [][]ngcmetadata.Profile{{higher, lower}, {lower, higher}} {
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, profiles, 90)
			Expect(result.ProfileId).To(Equal(lower.ID))
			Expect(result.Recommendation.GpuMemoryPerDeviceBytes).To(Equal(int64(20_000_000_001)))
		}
	})

	It("retains the first profile when all ranking keys are identical", func() {
		capacity := int64(80_000_000_000)
		first := publishedParallelProfile("same", 20, 1, 1)
		second := first
		second.InferenceBackend = "other"
		result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{first, second}, 90)
		Expect(result.InferenceBackend).To(Equal(first.InferenceBackend))
	})

	for _, tc := range []struct {
		name    string
		minimum float64
	}{
		{"NaN", math.NaN()}, {"positive infinity", math.Inf(1)}, {"negative infinity", math.Inf(-1)},
		{"zero", 0}, {"negative", -1}, {"unrepresentable bytes", float64(math.MaxInt64) / 1e9},
	} {
		It("rejects a profile minimum of "+tc.name, func() {
			capacity := int64(80_000_000_000)
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
				[]ngcmetadata.Profile{publishedParallelProfile("invalid", tc.minimum, 1, 1)}, 90)
			Expect(result.Status).To(Equal(api.RECOMMENDATIONSTATUS_NO_RECOMMENDATION))
			Expect(result.ProfileId).To(BeEmpty())
			Expect(result.Recommendation).To(BeNil())
		})
	}
})

var _ = Describe("request validation precedes dependency calls", func() {
	for _, tc := range []struct {
		name   string
		change func(*api.RecommendationRequest)
	}{
		{"invalid image", func(r *api.RecommendationRequest) { r.Image = "invalid" }},
		{"empty tag", func(r *api.RecommendationRequest) { r.Tag = "" }},
		{"empty GPU type", func(r *api.RecommendationRequest) { value := ""; r.GpuType = &value }},
		{"whitespace GPU type", func(r *api.RecommendationRequest) { value := "  "; r.GpuType = &value }},
	} {
		It(tc.name, func() {
			request := validRequest()
			tc.change(&request)
			// Nil dependencies make an unexpected lookup fail the spec.
			response, err := New(nil, nil, 90).Recommend(context.Background(), request)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeNumerically("==", 400))
			Expect(response).To(Equal(api.RecommendationResponse{Image: request.Image, Tag: request.Tag, Results: []api.RecommendationResult{}}))
		})
	}

	It("matches the requested GPU type exactly without trimming nonempty names", func() {
		request := validRequest()
		gpuType := " A100 "
		request.GpuType = &gpuType
		service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
			return []gpuinventory.Hardware{{Type: "A100"}}, nil
		}), nil, 90)
		_, err := service.Recommend(context.Background(), request)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(Equal("Requested GPU type is not reported by the cluster"))
	})
})

func TestRightSizing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recommendation suite")
}

var _ = Describe("recommendation inventory errors", func() {
	DescribeTable("preserves the inventory response status",
		func(status int) {
			upstream := api.NewError(status, "Cluster inventory lookup failed", nil)
			inventory := inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
				return nil, upstream
			})
			response, err := New(inventory, nil, 90).Recommend(context.Background(), validRequest())
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(int32(status)))
			Expect(response.Results).To(BeEmpty())
		},
		Entry("missing cluster", http.StatusNotFound),
		Entry("authentication failure", http.StatusUnauthorized),
		Entry("authorization failure", http.StatusForbidden),
		Entry("dependency failure", http.StatusServiceUnavailable),
	)
})

var _ = It("invalid sharing policy does not become estimated GPU count", func() {
	capacity := int64(100_000_000_000)
	result := recommendForHardware(gpuinventory.Hardware{Type: "unknown", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{sizingProfile("generic", 20)}, 0)
	Expect(string(result.Status)).To(Equal("no_recommendation"))
	Expect(result.Recommendation).To(BeNil())
	Expect(result.Reason).NotTo(BeEmpty())
})
