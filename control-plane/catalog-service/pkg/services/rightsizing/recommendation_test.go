// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"math"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
