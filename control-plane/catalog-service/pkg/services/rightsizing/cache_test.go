// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/require"
)

type metadataTransport func(*http.Request) (*http.Response, error)

func (f metadataTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCachedMetadataKeepsInventoryAndAllocationFresh(t *testing.T) {
	cachetesting.EnableCache()
	t.Cleanup(cachetesting.DisableCache)
	metadataCalls, inventoryCalls := 0, 0
	metadata := ngcmetadata.New(&http.Client{Transport: metadataTransport(func(r *http.Request) (*http.Response, error) {
		metadataCalls++
		body := `{"profiles":[{"profile_id":"baseline","precision":"bf16","tensor_parallel_size":1,"pipeline_parallel_size":1,"min_vram_per_device_gb":36}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}, 15*time.Minute)
	service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		inventoryCalls++
		capacity := int64(40_000_000_000)
		if inventoryCalls > 1 {
			capacity = 80_000_000_000
		}
		return []gpuinventory.Hardware{{Type: "NVIDIA-A100", MemoryBytes: &capacity, MaxDevicesPerNode: 8}}, nil
	}), metadata, 90)
	first, apiErr := service.Recommend(context.Background(), validRequest())
	require.Nil(t, apiErr)
	require.Len(t, first.Results, 1)
	require.NotNil(t, first.Results[0].Recommendation)
	require.Equal(t, "gpu_count", string(first.Results[0].Recommendation.AllocationType))
	second, apiErr := service.Recommend(context.Background(), validRequest())
	require.Nil(t, apiErr)
	require.Len(t, second.Results, 1)
	require.NotNil(t, second.Results[0].Recommendation)
	require.Equal(t, "gpu_memory", string(second.Results[0].Recommendation.AllocationType))
	require.Equal(t, int64(36_000_000_000), second.Results[0].Recommendation.GpuMemoryPerDeviceBytes)
	require.Equal(t, 2, inventoryCalls)
	require.Equal(t, 1, metadataCalls)
}

func TestCachedParallelProfilesStillRequireFreshSingleNodeCounts(t *testing.T) {
	cachetesting.EnableCache()
	t.Cleanup(cachetesting.DisableCache)
	metadataCalls, inventoryCalls := 0, 0
	metadata := ngcmetadata.New(&http.Client{Transport: metadataTransport(func(*http.Request) (*http.Response, error) {
		metadataCalls++
		body := `{"profiles":[{"profile_id":"tp2","tensor_parallel_size":2,"pipeline_parallel_size":1,"max_gpu_count":2,"min_vram_per_device_gb":57}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}, 15*time.Minute)
	service := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		inventoryCalls++
		capacity := int64(80_000_000_000)
		count := int64(2)
		if inventoryCalls > 1 {
			count = 1
		}
		return []gpuinventory.Hardware{{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: count}}, nil
	}), metadata, 90)
	first, apiErr := service.Recommend(context.Background(), validRequest())
	require.Nil(t, apiErr)
	require.Equal(t, int64(2), first.Results[0].Recommendation.GpuCount)
	second, apiErr := service.Recommend(context.Background(), validRequest())
	require.Nil(t, apiErr)
	require.Equal(t, "no_recommendation", string(second.Results[0].Status))
	require.Nil(t, second.Results[0].Recommendation)
	require.Equal(t, 2, inventoryCalls)
	require.Equal(t, 1, metadataCalls)
}
