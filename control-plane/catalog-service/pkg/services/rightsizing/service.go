/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"github.com/google/uuid"
)

type Inventory interface {
	Get(context.Context, string) ([]gpuinventory.Hardware, *api.Error)
}

type Metadata interface {
	Get(context.Context, string, string) (ngcmetadata.Metadata, *api.Error)
}

type Service struct {
	inventory     Inventory
	metadata      Metadata
	cutoffPercent int64
}

func New(inventory Inventory, metadata Metadata, cutoffPercent int64) *Service {
	return &Service{
		inventory:     inventory,
		metadata:      metadata,
		cutoffPercent: cutoffPercent,
	}
}

// Recommend validates the request, reads fresh cluster inventory, and loads
// published profiles for the exact image/tag (metadata may be cached). Each
// type/capacity is evaluated independently: filter GPU-agnostic profiles by memory
// and single-node fit, rank them, then apply the allocation policy.
// Missing sizing evidence is a result; retrieval/access failures are API errors.
func (s *Service) Recommend(ctx context.Context, request api.RecommendationRequest) (api.RecommendationResponse, *api.Error) {
	response := api.RecommendationResponse{
		Image:   request.Image,
		Tag:     request.Tag,
		Results: []api.RecommendationResult{},
	}
	if err := uuid.Validate(request.ClusterId); err != nil {
		return response, api.NewError(http.StatusBadRequest, "Invalid cluster ID", err)
	}
	if err := ngcmetadata.Validate(request.Image, request.Tag); err != nil {
		return response, err
	}
	if request.GpuType != nil && strings.TrimSpace(*request.GpuType) == "" {
		return response, api.NewError(http.StatusBadRequest, "GPU type must be nonempty or null", nil)
	}
	hardware, err := s.inventory.Get(ctx, request.ClusterId)
	if err != nil {
		return response, err
	}
	selected := make([]gpuinventory.Hardware, 0, len(hardware))
	for _, gpu := range hardware {
		if request.GpuType == nil || gpu.Type == *request.GpuType {
			selected = append(selected, gpu)
		}
	}
	if request.GpuType != nil && len(selected) == 0 {
		return response, api.NewError(http.StatusBadRequest, "Requested GPU type is not reported by the cluster", nil)
	}
	if len(selected) == 0 {
		return response, nil
	}
	metadata, err := s.metadata.Get(ctx, request.Image, request.Tag)
	if err != nil {
		return response, err
	}
	for _, gpu := range selected {
		response.Results = append(response.Results, recommendForHardware(gpu, metadata.Profiles, s.cutoffPercent))
	}
	return response, nil
}

// recommendForHardware keeps the published profile evidence in the result.
// Parallel profiles allocate exactly TP × PP whole GPUs. Single-GPU profiles
// use the configured sharing cutoff to choose memory or a whole GPU.
func recommendForHardware(gpu gpuinventory.Hardware, profiles []ngcmetadata.Profile, cutoffPercent int64) api.RecommendationResult {
	result := api.RecommendationResult{
		GpuType: gpu.Type,
		Status:  api.RECOMMENDATIONSTATUS_NO_RECOMMENDATION,
	}
	if gpu.MemoryBytes == nil || *gpu.MemoryBytes <= 0 {
		result.Reason = "GPU memory capacity is not reported by the cluster"
		return result
	}
	result.GpuMemoryCapacityBytes = *gpu.MemoryBytes
	if gpu.MaxDevicesPerNode <= 0 {
		result.Reason = "Matching single-node GPU count is not reported by the cluster"
		return result
	}
	chosen := selectProfile(profiles, *gpu.MemoryBytes, gpu.MaxDevicesPerNode)
	if chosen == nil {
		result.Reason = "No GPU-agnostic published profile with explicit TP/PP and a positive minimum fits per-device memory and the matching GPU count on one node"
		return result
	}
	profile := chosen.profile
	result.MinimumGpuMemoryPerDeviceGb = *profile.MinimumVRAMPerDeviceGB
	result.ProfileId = profile.ID
	result.Precision = profile.Precision
	result.InferenceBackend = profile.InferenceBackend
	result.TensorParallelSize = int64(*profile.TensorParallelSize)
	result.PipelineParallelSize = int64(*profile.PipelineParallelSize)
	if chosen.count > 1 {
		result.Status = api.RECOMMENDATIONSTATUS_RECOMMENDED
		result.Recommendation = &api.Recommendation{
			AllocationType:          api.RECOMMENDATIONALLOCATIONTYPE_GPU_COUNT,
			GpuCount:                chosen.count,
			GpuMemoryPerDeviceBytes: *gpu.MemoryBytes,
		}
		return result
	}
	allocation, err := SingleGPUAllocation(chosen.minimumBytes, *gpu.MemoryBytes, cutoffPercent)
	if err != nil {
		result.Reason = "GPU allocation sharing policy is invalid"
		return result
	}
	result.Status = api.RECOMMENDATIONSTATUS_RECOMMENDED
	allocationType := api.RECOMMENDATIONALLOCATIONTYPE_GPU_MEMORY
	if allocation.WholeGPU {
		allocationType = api.RECOMMENDATIONALLOCATIONTYPE_GPU_COUNT
	}
	result.Recommendation = &api.Recommendation{
		AllocationType:          allocationType,
		GpuCount:                1,
		GpuMemoryPerDeviceBytes: allocation.AllocatedMemoryBytes,
	}
	return result
}
