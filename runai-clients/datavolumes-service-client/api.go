/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolumes_service_client

import (
	"context"
	"net/http"

	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/internal"
)

type IDatavolumesExternalClient interface {
}

type IDatavolumesClient interface {
	IDatavolumesExternalClient
	GetDatavolumeAncestorsSync(ctx context.Context, resourceId string) (*openapi.Ancestors, error)
	UpdateDatavolumeStatus(ctx context.Context, id string, datavolumeStatusInternal openapi.DatavolumeStatusInternal) (*http.Response, error)
	SyncDatavolumes(ctx context.Context, datavolumes []openapi.DatavolumeStatusForSync, clusterId string, from *string, to *string) (*http.Response, error)
	GetDatavolumes(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
		limit, offset int32, filter []string, sortBy string, sortOrder string) (*openapi.GetDatavolumes200Response, error)
	GetDataVolumesNotPaginated(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
		filter []string, sortBy string, sortOrder string) ([]openapi.DatavolumeNoSharedScopes, error)
	CreateInternalDatavolume(ctx context.Context, request openapi.InternalDatavolumeCreationFields) (*openapi.InternalDatavolume, *http.Response, error)
	DeleteDatavolumeById(ctx context.Context, id string) (*openapi.HttpResponse, *http.Response, error)
}

func NewDatavolumesExternalClient(baseURL string) IDatavolumesExternalClient {
	return NewDatavolumesClient(client_utils.GetExternalUrl(baseURL))
}

func NewDatavolumesClient(baseURL string) IDatavolumesClient {
	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return NewDatavolumesClientWithConfig(clientConfig)
}

func NewDatavolumesClientWithConfig(config *openapi.Configuration) IDatavolumesClient {
	return &internal.DatavolumesClient{ApiClient: openapi.NewAPIClient(config)}
}
