/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"context"
	"math"
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
	inventory Inventory
	metadata  Metadata
}

func New(inventory Inventory, metadata Metadata) *Service {
	return &Service{inventory: inventory, metadata: metadata}
}

func (s *Service) Recommend(ctx context.Context, request api.RightSizingRequest) (api.RightSizingResponse, *api.Error) {
	response := api.RightSizingResponse{Image: request.Image, Tag: request.Tag, Results: []api.RightSizingResponseResultsInner{}}
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
		response.Results = append(response.Results, recommendForHardware(gpu, metadata.Profiles))
	}
	return response, nil
}

func recommendForHardware(gpu gpuinventory.Hardware, profiles []ngcmetadata.Profile) api.RightSizingResponseResultsInner {
	result := api.RightSizingResponseResultsInner{GpuType: gpu.Type, Status: "no_recommendation"}
	if gpu.MemoryBytes == nil || *gpu.MemoryBytes <= 0 {
		result.Reason = "GPU memory capacity is not reported by the cluster"
		return result
	}
	result.GpuMemoryCapacityBytes = *gpu.MemoryBytes
	var chosen *ngcmetadata.Profile
	for i := range profiles {
		profile := &profiles[i]
		// Only explicit single-device profiles establish the baseline for this
		// API slice. Multi-device profile selection belongs to a separate effort.
		if profile.ID == "" || profile.TensorParallelSize == nil || *profile.TensorParallelSize != 1 ||
			profile.PipelineParallelSize == nil || *profile.PipelineParallelSize != 1 ||
			!matchesHardware(*profile) || profile.MinimumVRAMPerDeviceGB == nil {
			continue
		}
		minimum := *profile.MinimumVRAMPerDeviceGB
		if math.IsNaN(minimum) || math.IsInf(minimum, 0) || minimum <= 0 || minimum >= float64(math.MaxInt64)/1e9 {
			continue
		}
		if chosen == nil || minimum < *chosen.MinimumVRAMPerDeviceGB ||
			(minimum == *chosen.MinimumVRAMPerDeviceGB && profile.ID < chosen.ID) {
			chosen = profile
		}
	}
	if chosen == nil {
		result.Reason = "No compatible explicit single-GPU profile with a positive minimum memory value"
		return result
	}
	result.MinimumGpuMemoryPerDeviceGb = *chosen.MinimumVRAMPerDeviceGB
	result.ProfileId = chosen.ID
	result.Precision = chosen.Precision
	result.InferenceBackend = chosen.InferenceBackend
	// NGC reports GB; inventory reports bytes. Round requirements upward.
	minimumBytes := int64(math.Ceil(*chosen.MinimumVRAMPerDeviceGB * 1e9))
	allocation, err := SingleGPUAllocation(minimumBytes, *gpu.MemoryBytes)
	if err != nil {
		// The baseline here is explicitly TP=1, PP=1. This remains only a naive
		// estimate; it must not be mistaken for a supported parallel profile.
		count := minimumBytes / *gpu.MemoryBytes
		if minimumBytes%*gpu.MemoryBytes != 0 {
			count++
		}
		result.Status = "recommended"
		result.Recommendation = &api.RightSizingResponseResultsInnerRecommendation{
			AllocationType: "gpu_count", GpuCount: count, Estimated: true,
			Warning: profileSizingWarning(*chosen) + " Estimated from a single-device profile. This count is not a validated tensor-parallel configuration.",
		}
		return result
	}
	result.Status = "recommended"
	allocationType := "gpu_memory"
	if allocation.WholeGPU {
		allocationType = "gpu_count"
	}
	result.Recommendation = &api.RightSizingResponseResultsInnerRecommendation{
		AllocationType: allocationType, GpuCount: 1,
		GpuMemoryPerDeviceBytes: allocation.AllocatedMemoryBytes,
		Estimated:               false,
		Warning:                 profileSizingWarning(*chosen),
	}
	return result
}

func matchesHardware(profile ngcmetadata.Profile) bool {
	// NGC requires PCI identifiers to verify targeted profiles; inventory currently
	// exposes only display names. Tested devices are history, not restrictions.
	return profile.GPU == "" && profile.GPUDevice == ""
}

// A published memory floor does not establish that the profile can execute on
// every GPU. Keep this prerequisite visible even for exact allocation arithmetic.
func profileSizingWarning(profile ngcmetadata.Profile) string {
	warning := "Memory sizing only; compatibility of the selected profile, precision, backend and runtime configuration with the target GPU is unverified. Verify the profile can run on the target hardware before deploying."
	if strings.EqualFold(profile.Precision, "nvfp4") {
		warning += " NVFP4 requires Blackwell by default. Pre-Blackwell emulation requires a supported vLLM or SGLang image" +
			" and NIM_ALLOW_NVFP4_EMULATION=1; TensorRT-LLM does not support emulation. Verify the exact image tag and host configuration."
	}
	return warning
}
