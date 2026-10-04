/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunit

import (
	"context"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

type ClientInterface interface {
	GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error)
	ListProjects(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error)
	GetAncestors(ctx context.Context, scopeType, scopeId string) (*openapi.Ancestors, *openapi.Error)
}
