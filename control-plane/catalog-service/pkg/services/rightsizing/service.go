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

// Recommend evaluates installed GPU capacity, not currently free resources.
// Read fresh requested hardware, load the exact image/tag profiles (which may be
// cached), then construct one result per GPU type/capacity variant. Missing sizing
// evidence yields no_recommendation; retrieval/access failures are API errors.
func (s *Service) Recommend(ctx context.Context, request api.RecommendationRequest) (api.RecommendationResponse, *api.Error) {
	response := api.RecommendationResponse{
		Image:   request.Image,
		Tag:     request.Tag,
		Results: []api.RecommendationResult{},
	}
	if err := validateRequest(request); err != nil {
		return response, err
	}
	hardware, err := s.loadRequestedHardware(ctx, request.ClusterId, request.GpuType)
	if err != nil {
		return response, err
	}
	// An existing cluster without GPUs has no variants to evaluate, so no
	// metadata lookup is needed. An explicitly requested absent type fails above.
	if len(hardware) == 0 {
		return response, nil
	}
	metadata, err := s.metadata.Get(ctx, request.Image, request.Tag)
	if err != nil {
		return response, err
	}
	for _, gpu := range hardware {
		response.Results = append(response.Results, recommendForHardware(gpu, metadata.Profiles, s.cutoffPercent))
	}
	return response, nil
}

func validateRequest(request api.RecommendationRequest) *api.Error {
	if err := uuid.Validate(request.ClusterId); err != nil {
		return api.NewError(http.StatusBadRequest, "Invalid cluster ID", err)
	}
	if err := ngcmetadata.Validate(request.Image, request.Tag); err != nil {
		return err
	}
	if request.GpuType != nil && strings.TrimSpace(*request.GpuType) == "" {
		return api.NewError(http.StatusBadRequest, "GPU type must be nonempty or null", nil)
	}
	return nil
}

// loadRequestedHardware retains each capacity variant of the requested type.
// Inventory lookup establishes cluster access/existence before type selection.
func (s *Service) loadRequestedHardware(ctx context.Context, clusterID string, gpuType *string) ([]gpuinventory.Hardware, *api.Error) {
	hardware, err := s.inventory.Get(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	selected := make([]gpuinventory.Hardware, 0, len(hardware))
	for _, gpu := range hardware {
		if gpuType == nil || gpu.Type == *gpuType {
			selected = append(selected, gpu)
		}
	}
	if gpuType != nil && len(selected) == 0 {
		return nil, api.NewError(http.StatusBadRequest, "Requested GPU type is not reported by the cluster", nil)
	}
	return selected, nil
}
