// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type metadataTransport func(*http.Request) (*http.Response, error)

func (f metadataTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

var _ = It("cached metadata keeps inventory and allocation fresh", func() {
	cachetesting.EnableCache()
	DeferCleanup(cachetesting.DisableCache)
	metadataCalls, inventoryCalls := 0, 0
	metadata := ngcmetadata.New(&http.Client{Transport: metadataTransport(func(r *http.Request) (*http.Response, error) {
		metadataCalls++
		body := `{"profiles":[{"profile_id":"baseline","precision":"bf16","tensor_parallel_size":1,"pipeline_parallel_size":1,"min_vram_per_device_gb":36}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}, 15*time.Minute)
	service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		inventoryCalls++
		capacity := int64(40_000_000_000)
		if inventoryCalls > 1 {
			capacity = 80_000_000_000
		}
		return []gpuinventory.Hardware{{Type: "NVIDIA-A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}}, nil
	}), metadata, 90)
	first, apiErr := service.Recommend(context.Background(), validRequest())
	Expect(apiErr).To(BeNil())
	Expect(first.Results).To(HaveLen(1))
	Expect(first.Results[0].Recommendation).To(Not(BeNil()))
	Expect(string(first.Results[0].Recommendation.AllocationType)).To(Equal("gpu_count"))
	second, apiErr := service.Recommend(context.Background(), validRequest())
	Expect(apiErr).To(BeNil())
	Expect(second.Results).To(HaveLen(1))
	Expect(second.Results[0].Recommendation).To(Not(BeNil()))
	Expect(string(second.Results[0].Recommendation.AllocationType)).To(Equal("gpu_memory"))
	Expect(second.Results[0].Recommendation.GpuMemoryPerDeviceBytes).To(Equal(int64(36_000_000_000)))
	Expect(inventoryCalls).To(BeNumerically("==", 2))
	Expect(metadataCalls).To(BeNumerically("==", 1))
})

var _ = It("cached parallel profiles still require fresh single node counts", func() {
	cachetesting.EnableCache()
	DeferCleanup(cachetesting.DisableCache)
	metadataCalls, inventoryCalls := 0, 0
	metadata := ngcmetadata.New(&http.Client{Transport: metadataTransport(func(*http.Request) (*http.Response, error) {
		metadataCalls++
		body := `{"profiles":[{"profile_id":"tp2","tensor_parallel_size":2,"pipeline_parallel_size":1,"max_gpu_count":2,"min_vram_per_device_gb":57}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}, 15*time.Minute)
	service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, *api.Error) {
		inventoryCalls++
		capacity := int64(80_000_000_000)
		count := int64(2)
		if inventoryCalls > 1 {
			count = 1
		}
		return []gpuinventory.Hardware{{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: count}}, nil
	}), metadata, 90)
	first, apiErr := service.Recommend(context.Background(), validRequest())
	Expect(apiErr).To(BeNil())
	Expect(first.Results[0].Recommendation.GpuCount).To(Equal(int64(2)))
	second, apiErr := service.Recommend(context.Background(), validRequest())
	Expect(apiErr).To(BeNil())
	Expect(string(second.Results[0].Status)).To(Equal("no_recommendation"))
	Expect(second.Results[0].Recommendation).To(BeNil())
	Expect(inventoryCalls).To(BeNumerically("==", 2))
	Expect(metadataCalls).To(BeNumerically("==", 1))
})
