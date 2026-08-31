/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	openapi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
)

type TenantsManagerClient struct {
	api *openapi.APIClient
}

func NewTenantsManagerClient(baseURL string) *TenantsManagerClient {
	config := openapi.NewConfiguration()
	config.Host = strings.Replace(baseURL, "http://", "", 1)
	config.Scheme = "http"
	config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	return &TenantsManagerClient{
		api: openapi.NewAPIClient(config),
	}
}

func (tc *TenantsManagerClient) GetTenantIdByNameOrAlias(ctx context.Context, tenantName string) (string, error) {
	if tenantName == "" {
		return "", fmt.Errorf("invalid tenant name: empty")
	}

	tenant, _, err := tc.api.InternalAPI.GetTenantByNameOrAlias(ctx).TenantName(tenantName).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to resolve tenant %q: %w", tenantName, err)
	}

	return fmt.Sprintf("%d", tenant.GetId()), nil
}
