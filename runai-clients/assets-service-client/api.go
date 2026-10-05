/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets_service_client

import (
	"context"
	"net/http"

	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/internal"
)

type IAssetsExternalClient interface {
	ListComputeAssets(ctx context.Context, name string) (*openapi.ComputeListResponse, *http.Response, error)
	CreateComputeAsset(ctx context.Context, request *openapi.ComputeCreationRequest) (*openapi.ComputeAsset, *http.Response, error)
	ListEnvironmentAssets(ctx context.Context, name string) (*openapi.EnvironmentListResponse, *http.Response, error)
	CreateEnvironmentAsset(ctx context.Context, request *openapi.EnvironmentCreationRequest) (*openapi.EnvironmentAsset, *http.Response, error)
	ListModelAssets(ctx context.Context, name string) (*openapi.ModelListResponse, *http.Response, error)
	CreateModelAsset(ctx context.Context, request *openapi.ModelCreationRequest) (*openapi.ModelAsset, *http.Response, error)
	ListDatasourceAssets(ctx context.Context, ids []string) (*openapi.DatasourceListResponse, *http.Response, error)

	RequestAssetFlatFormat(ctx context.Context, request openapi.AssetFlatRequest) (*openapi.AssetFlatResponse, *openapi.Error)

	CredentialsSyncUpdates(ctx context.Context, clusterId string, fromUpdate int64, request openapi.ClusterSyncReport) (*openapi.CredentialsUpdatesResp, *openapi.Error)
	DatasourceSyncUpdates(ctx context.Context, clusterId string, fromUpdate int64, request openapi.ClusterSyncReport) (*openapi.DatasourceUpdatesResp, *openapi.Error)
	ListTemplatesOfTenant(ctx context.Context) (*openapi.WorkloadTemplateListResponse, *openapi.Error)
	SyncEvents(ctx context.Context, request openapi.EventsSyncRequest) (*http.Response, error)
	SyncAssetsResourcesIds(ctx context.Context, request openapi.AssetResourcesIdsSyncRequest) (*http.Response, error)
	GetNgcApiKeyByIdInternal(ctx context.Context, assetId string) (*openapi.NgcApiKeyInternal, *openapi.Error)
}

type IAssetsClient interface {
	IAssetsExternalClient
	ListCredentialsAssets(ctx context.Context, privateOnly bool) (*openapi.CredentialsListResponse, *openapi.Error)
	SyncEvents(ctx context.Context, request openapi.EventsSyncRequest) (*http.Response, error)
}

func NewAssetsClient(baseURL string) IAssetsClient {
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return &internal.AssetsClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewAssetsExternalClient(baseURL string) IAssetsExternalClient {
	return NewAssetsClient(client_utils.GetExternalUrl(baseURL))
}

func NewExternalClientWithConfig(config *openapi.Configuration) IAssetsExternalClient {
	return &internal.AssetsClient{ApiClient: openapi.NewAPIClient(config)}
}
