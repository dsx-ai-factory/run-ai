/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	catalogconfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services/rightsizing"
	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type RightSizingAPIService struct {
	service *rightsizing.Service
}

func NewRightSizingAPIService() *RightSizingAPIService {
	servicesConfig := config.Get[*catalogconfig.Services]()
	inventory := gpuinventory.New(clusterclient.NewClusterClient(servicesConfig.ClusterServiceUrl), time.Duration(servicesConfig.InventoryRequestTimeoutSec)*time.Second)
	metadata := ngcmetadata.New(&http.Client{
		Timeout: time.Duration(config.Get[*catalogconfig.NGC]().RequestTimeoutSec) * time.Second,
	})
	return &RightSizingAPIService{service: rightsizing.New(inventory, metadata)}
}

func (s *RightSizingAPIService) RecommendInferenceResources(ctx context.Context, request api.RightSizingRequest) (api.ImplResponse, error) {
	response, apiErr := s.service.Recommend(ctx, request)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr), nil
	}
	return api.Response(http.StatusOK, response), nil
}
