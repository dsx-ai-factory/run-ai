/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

type TenantsClienter interface {
	GetTenantIds(ctx context.Context) ([]int32, error)
}

type tenantsClient struct {
	client *api.APIClient
}

func newTenantsClient(tenantsManagerUrl string) *tenantsClient {
	return &tenantsClient{
		client: api.NewAPIClient(&api.Configuration{
			Servers: api.ServerConfigurations{{URL: tenantsManagerUrl}},
		}),
	}
}

func (c *tenantsClient) GetTenantIds(ctx context.Context) ([]int32, error) {
	tenants, _, err := c.client.InternalAPI.ManagementGetTenantsInternal(ctx).Execute()
	if err != nil {
		return nil, err
	}

	tenantIds := make([]int32, 0, len(tenants))
	for _, tenant := range tenants {
		tenantIds = append(tenantIds, tenant.Id)
	}

	return tenantIds, nil
}
