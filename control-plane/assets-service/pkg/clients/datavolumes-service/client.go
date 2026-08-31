/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolumes_service

import (
	"context"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"net/http"
	"sync"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	datavolumes_service_client "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

type Client interface {
	CreateInternalDatavolume(ctx context.Context, request openapi.InternalDatavolumeCreationFields) (*openapi.InternalDatavolume, *http.Response, error)
	DeleteDatavolumeById(ctx context.Context, id string) (*openapi.HttpResponse, *http.Response, error)
	GetDataVolumesNotPaginated(ctx context.Context, requestType openapi.DatavolumeRequestType, usableInProjectId string,
		filter []string, sortBy string, sortOrder string) ([]openapi.DatavolumeNoSharedScopes, error)
}

var onceClient sync.Once
var client Client

func Get() Client {
	onceClient.Do(initRealClient)
	return client
}

func SetDatavolumesMockClient(c Client) {
	client = c
}

func initRealClient() {
	if client == nil {
		client = datavolumes_service_client.NewDatavolumesClient(runai_config.Get[*config.ServicesConfig]().DataVolumesServiceUrl)
	}
}
