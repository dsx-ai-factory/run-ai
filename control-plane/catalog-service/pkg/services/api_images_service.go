/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/constants"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/dao"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// imagesAPIService is a service that implements the logic for the ImagesAPIServicer
// This service should implement the business logic for every endpoint for the ImagesAPI API.
// Include any external packages or services that will be required by this service.

type NGCClient interface {
	GetImageTags(ctx context.Context, registryUrl string, imageName string, apiKey string) (api.ImageTags, error)
	GetImages(ctx context.Context, registryUrl string, apiKey string, search string, isNgcCatalog bool) ([]string, error)
}

type AssetsClient interface {
	GetNgcApiKeyByIdInternal(ctx context.Context, assetId string) (*assetsApi.NgcApiKeyInternal, *assetsApi.Error)
}

type imagesAPIService struct {
	ngcClient    NGCClient
	assetsClient AssetsClient
}

// NewImagesAPIService creates a default api service
func NewImagesAPIService(ngcClient NGCClient, assetsClient AssetsClient) *imagesAPIService {
	return &imagesAPIService{
		ngcClient:    ngcClient,
		assetsClient: assetsClient,
	}
}

// ListImages - List images by registry-integration id
func (s *imagesAPIService) ListImages(ctx context.Context, registryIntegrationId string, search string, assetId string) (api.ImplResponse, error) {
	registry, apiErr := dao.GetRegistryById(ctx, registryIntegrationId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	var ngcApiKey string
	if assetId != "" {
		ngcApiKeyInternal, err := s.assetsClient.GetNgcApiKeyByIdInternal(ctx, assetId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get NGC API key from assets service: %v", err)
			return api.Response(int(err.Code), err), nil
		}
		ngcApiKey = ngcApiKeyInternal.Key
	}

	isNgcCatalog := registry.Name == constants.InternalNgcIntegrationName

	images, err := s.ngcClient.GetImages(ctx, registry.Url, ngcApiKey, search, isNgcCatalog)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get NGC images: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "Failed to get images from NGC", err)), nil
	}

	resp := api.ImagesListResponse{
		Images: images,
		Prefix: convutils.Ref(constants.NgcImagePrefix),
	}

	return api.Response(http.StatusOK, resp), nil
}

// GetImageTags - Get the image's tags
func (s *imagesAPIService) GetImageTags(ctx context.Context, registryIntegrationId string, imageName string, assetId string) (api.ImplResponse, error) {
	registry, apiErr := dao.GetRegistryById(ctx, registryIntegrationId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	var ngcApiKey string
	if assetId != "" {
		ngcApiKeyInternal, err := s.assetsClient.GetNgcApiKeyByIdInternal(ctx, assetId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get NGC API key from assets service: %v", err)
			return api.Response(int(err.Code), err), nil
		}
		ngcApiKey = ngcApiKeyInternal.Key
	}

	imageTags, err := s.ngcClient.GetImageTags(ctx, registry.Url, imageName, ngcApiKey)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get image tags from NGC: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "Failed to get image tags from NGC", err)), nil
	}

	return api.Response(http.StatusOK, imageTags), nil
}
