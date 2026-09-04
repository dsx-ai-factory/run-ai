/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunit

import (
	"context"

	org_unit_service_client "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

type Client struct {
	apiClient org_unit_service_client.IOrgUnitClient
}

func (c *Client) GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error) {
	return c.apiClient.GetOrgTree(ctx)
}

func (c *Client) ListProjects(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error) {
	return c.apiClient.GetProjectsMetadataByCluster(ctx, clusterId)
}

func (c *Client) GetAncestors(ctx context.Context, scopeType, scopeId string) (*openapi.Ancestors, *openapi.Error) {
	return c.apiClient.GetAncestorsOfScope(ctx, openapi.ScopeType(scopeType), scopeId)
}
