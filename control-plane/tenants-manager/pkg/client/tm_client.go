/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenants_manager_client

import (
	"context"
	"net/http"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

type ITenantsManager interface {
	ListAllTenantsInternal(ctx context.Context) ([]openapi.Tenant, *http.Response, error)
	GetTenantByIdInternal(ctx context.Context, id int32) (*openapi.Tenant, *http.Response, error)
	GetAllLogos(ctx context.Context, offset int32, limit int32) (*openapi.TenantLogosResponse, *http.Response, error)
}

type TenantsManagerClient struct {
	apiClient *openapi.APIClient
}

// NewTenantsManagerClient creates a new API client.  Requires a base URL.
func NewTenantsManagerClient(baseURL string) ITenantsManager {
	config := openapi.NewConfiguration()
	config.Host = strings.Replace(baseURL, "http://", "", 1)
	config.Scheme = "http"
	return &TenantsManagerClient{
		apiClient: openapi.NewAPIClient(config),
	}
}

func (t TenantsManagerClient) ListAllTenantsInternal(ctx context.Context) ([]openapi.Tenant, *http.Response, error) {
	tenants, httpResponse, err := t.apiClient.InternalAPI.ManagementGetTenantsInternal(ctx).Execute()
	if err != nil {
		return nil, httpResponse, err
	}
	return tenants, httpResponse, nil
}

func (t TenantsManagerClient) GetTenantByIdInternal(ctx context.Context, id int32) (*openapi.Tenant, *http.Response, error) {
	tenant, httpResponse, err := t.apiClient.TenantsAPI.GetTenantByIdInternal(ctx, id).Execute()
	if err != nil {
		return nil, httpResponse, err
	}
	return tenant, httpResponse, nil
}

func (t TenantsManagerClient) GetAllLogos(ctx context.Context, offset int32, limit int32) (*openapi.TenantLogosResponse, *http.Response, error) {
	logos, httpResponse, err := t.apiClient.InternalAPI.GetAllLogos(ctx).Offset(offset).Limit(limit).Execute()
	if err != nil {
		return nil, httpResponse, err
	}
	return logos, httpResponse, nil
}
