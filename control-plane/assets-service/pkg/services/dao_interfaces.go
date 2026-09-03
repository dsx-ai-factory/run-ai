/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
)

// ConfigMapsDBAccessor is an interface to interact with the db ConfigMap assets table
type ConfigMapsDBAccessor interface {
	// CreateConfigMapAsset insets a new ConfigMap assets into the ConfigMaps and assets metadata tables
	CreateConfigMapAsset(ctx context.Context, request api.ConfigMapCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ConfigMapAsset, *api.Error)
	// DeleteConfigMapAsset deletes a ConfigMap asset
	DeleteConfigMapAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType dao.DeletionType) *api.Error
	// GetConfigMapAssetById returns a ConfigMap asset by it asset id
	GetConfigMapAssetById(ctx context.Context, assetId string) (*api.ConfigMapAsset, *api.Error)
	// ListConfigMapAssets returns a list of  ConfigMap assets filter by query params
	ListConfigMapAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.ConfigMapListResponse, *api.Error)
	// UpdateConfigMapAsset updates a ConfigMap assets
	UpdateConfigMapAsset(ctx context.Context, assetId string, request api.ConfigMapUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.ConfigMapAsset, *api.Error)
}

// SecretVolumeDBAccessor is an interface to interact with the DB's Secret data storage assets table
type SecretVolumeDBAccessor interface {
	// CreateSecretVolumeAsset insets a new secret asset into the Secret and assets metadata tables
	CreateSecretVolumeAsset(ctx context.Context, request api.SecretAssetCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.SecretAsset, *api.Error)
	// DeleteSecretVolumeAsset deletes a Secret asset
	DeleteSecretVolumeAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType dao.DeletionType) *api.Error
	// GetSecretVolumeAssetById returns a Secret asset by it asset id
	GetSecretVolumeAssetById(ctx context.Context, assetId string) (*api.SecretAsset, *api.Error)
	// ListSecretVolumeAssets returns a list of  Secret assets filter by query params
	ListSecretVolumeAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.SecretAssetListResponse, *api.Error)
	// UpdateSecretVolumeAsset updates a Secret asset
	UpdateSecretVolumeAsset(ctx context.Context, assetId string, request api.SecretAssetUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.SecretAsset, *api.Error)
}
