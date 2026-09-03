/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunit

import (
	"context"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

type FailingMockClient struct {
	FailGetOrgTree bool
}

func (fmc *FailingMockClient) Client() ClientInterface {
	return fmc
}

func (fmc *FailingMockClient) GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error) {
	if fmc.FailGetOrgTree {
		return nil, openapi.NewError(http.StatusServiceUnavailable, "mock: failed to get org tree")
	}

	return []openapi.OrgUnit{}, nil
}

func (fmc *FailingMockClient) ListProjects(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error) {
	return []openapi.ProjectMetadata{}, nil
}

func (fmc *FailingMockClient) GetAncestors(ctx context.Context, scopeType, scopeId string) (*openapi.Ancestors, *openapi.Error) {
	return &openapi.Ancestors{}, nil
}
