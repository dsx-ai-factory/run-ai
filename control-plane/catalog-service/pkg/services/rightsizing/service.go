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
	Get(context.Context, string) ([]gpuinventory.Hardware, error)
}

type Metadata interface {
	Get(context.Context, string, string) (ngcmetadata.Metadata, error)
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

func (s *Service) Recommend(ctx context.Context, request api.RightSizingRequest) (api.RightSizingResponse, *api.Error) {
	response := api.RightSizingResponse{
		Image:   request.Image,
		Tag:     request.Tag,
		Results: []api.RightSizingResult{},
	}
	if err := uuid.Validate(request.ClusterId); err != nil {
		return response, api.NewError(http.StatusBadRequest, "Invalid cluster ID", err)
	}
	if err := ngcmetadata.Validate(request.Image, request.Tag); err != nil {
		return response, api.NewError(http.StatusBadRequest, err.Error(), nil)
	}
	if request.GpuType != nil && strings.TrimSpace(*request.GpuType) == "" {
		return response, api.NewError(http.StatusBadRequest, "GPU type must be nonempty or null", nil)
	}
	hardware, err := s.inventory.Get(ctx, request.ClusterId)
	if err != nil {
		return response, api.NewError(http.StatusServiceUnavailable, "Cluster GPU inventory unavailable", err)
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
		return response, api.NewError(http.StatusServiceUnavailable, "NGC sizing metadata unavailable", err)
	}
	for _, gpu := range selected {
		response.Results = append(response.Results, recommendForHardware(gpu, metadata.Profiles, s.cutoffPercent))
	}
	return response, nil
}

func recommendForHardware(gpu gpuinventory.Hardware, profiles []ngcmetadata.Profile, cutoffPercent int64) api.RightSizingResult {
	result := api.RightSizingResult{
		GpuType: gpu.Type,
		Status:  api.RIGHTSIZINGSTATUS_NO_RECOMMENDATION,
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
	pciDevice, pciError := gpuinventory.ResolvePCIDevice(gpu.Type)
	chosen := selectProfile(profiles, pciDevice, *gpu.MemoryBytes, gpu.MaxDevicesPerNode)
	if chosen == nil {
		result.Reason = "No applicable published profile with explicit TP/PP and a positive minimum fits per-device memory and the matching GPU count on one node"
		if pciError != nil {
			result.Reason += "; " + pciError.Error()
		}
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
		result.Status = api.RIGHTSIZINGSTATUS_RECOMMENDED
		result.Recommendation = &api.RightSizingRecommendation{
			AllocationType:          api.RIGHTSIZINGALLOCATIONTYPE_GPU_COUNT,
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
	result.Status = api.RIGHTSIZINGSTATUS_RECOMMENDED
	allocationType := api.RIGHTSIZINGALLOCATIONTYPE_GPU_MEMORY
	if allocation.WholeGPU {
		allocationType = api.RIGHTSIZINGALLOCATIONTYPE_GPU_COUNT
	}
	result.Recommendation = &api.RightSizingRecommendation{
		AllocationType:          allocationType,
		GpuCount:                1,
		GpuMemoryPerDeviceBytes: allocation.AllocatedMemoryBytes,
	}
	return result
}

type hardwareMatch int

const (
	directHardwareMatch hardwareMatch = iota
	testedHardwareMatch
	genericHardwareMatch
	noHardwareMatch
)

func matchHardware(profile ngcmetadata.Profile, pciDevice string) hardwareMatch {
	if pciDevice != "" {
		if strings.EqualFold(profile.GPUDevice, pciDevice) {
			return directHardwareMatch
		}
		for _, tested := range profile.TestedGPUDevices {
			if strings.EqualFold(tested, pciDevice) {
				return testedHardwareMatch
			}
		}
	}
	// NGC treats an untargeted profile as GPU-agnostic. A tested-device list
	// records validation history; it does not turn this into a SKU allowlist.
	if profile.GPU == "" && profile.GPUDevice == "" {
		return genericHardwareMatch
	}
	return noHardwareMatch
}
