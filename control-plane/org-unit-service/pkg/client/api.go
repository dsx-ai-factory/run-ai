/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit_service_client

import (
	"context"
	"net/http"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/internal"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
)

type IOrgUnitExternalClient interface {
	GetProjectsSync(ctx context.Context, clusterId string, limit int32, offset int32, UpdatedAfter time.Time, excludePermissions bool) (*openapi.GetSyncProjects200Response, error)
	UpdateProjectsStatus(ctx context.Context, projectStatusUpdateRequest openapi.ProjectStatusUpdateRequest) (*http.Response, error)
	SyncClusterProjects(ctx context.Context, clusterId string, syncProjects []openapi.ProjectInSyncRequest) (*http.Response, error)
	GetDepartmentsSync(ctx context.Context, clusterId string, limit int32, offset int32, UpdatedAfter time.Time) (*openapi.GetSyncDepartments200Response, error)
	UpdateDepartmentsStatus(ctx context.Context, departmentId string, departmentStatusUpdateRequest openapi.DepartmentStatusUpdateRequest) (*http.Response, error)
}

type IOrgUnitClient interface {
	IOrgUnitExternalClient
	NodePoolCreated(ctx context.Context, clusterId string, nodePoolId string, nodePoolName string) error
	NodePoolDeleted(ctx context.Context, clusterId string, nodePoolId string, nodePoolName string) error
	DefaultNodePoolCreated(ctx context.Context, clusterId string, nodePoolId string) error
	ClusterDeleted(ctx context.Context, clusterId string) error
	GetProjectById(ctx context.Context, projectId string) (*openapi.Project, *openapi.Error)
	GetOrgUnits(ctx context.Context, scopeType *openapi.ScopeType, scopeId *string) ([]openapi.OrgUnit, *openapi.Error)
	// GetOrgTreeMap Returns a map from org unit to its parent (if exists)
	GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *openapi.Error)
	GetProjectsMetadataByCluster(ctx context.Context, clusterId string) ([]openapi.ProjectMetadata, *openapi.Error)
	GetProjectsMetadataByDepartment(ctx context.Context, departmentId string) ([]openapi.ProjectMetadata, *openapi.Error)
	GetProjectAncestors(ctx context.Context, projectId string) (*openapi.Ancestors, *openapi.Error)
	GetDepartmentAncestors(ctx context.Context, departmentId string) (*openapi.Ancestors, *openapi.Error)
	GetDepartmentsDescendants(ctx context.Context, departmentIds string) (*openapi.Descendants, *openapi.Error)
	GetAncestorsOfScope(ctx context.Context, scopeType openapi.ScopeType, scopeId string) (*openapi.Ancestors, *openapi.Error)
	CountClusterDepartments(ctx context.Context, clusterId string) (int64, *openapi.Error)
	GetInternalDeletedOrgUnits(ctx context.Context, deletedAfter time.Time) ([]openapi.DeletedOrgUnit, *openapi.Error)
	GetOrgTree(ctx context.Context) ([]openapi.OrgUnit, *openapi.Error)
	UpdateOrgUnitSettings(ctx context.Context, name string, value string) error
	NodePoolSummary(ctx context.Context, nodePoolId string, summaryType openapi.OrgUnitsNodepoolSummaryType) ([]openapi.OrgUnitsNodepoolSummaryInner, error)
	NodePoolResetResources(ctx context.Context, clusterId string, nodePoolId string, action openapi.NodepoolResourcesResetAction) error
}

func NewOrgUnitClient(baseURL string) IOrgUnitClient {
	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	client := openapi.NewAPIClient(clientConfig)
	return &internal.OrgUnitClient{ApiClient: client}
}

func NewOrgUnitExternalClient(baseURL string) IOrgUnitExternalClient {
	return NewOrgUnitClient(client_utils.GetExternalUrl(baseURL))
}

func NewOrgUnitClientWithConfig(config *openapi.Configuration) IOrgUnitClient {
	return &internal.OrgUnitClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewOrgUnitExternalClientWithConfig(config *openapi.Configuration) IOrgUnitExternalClient {
	return NewOrgUnitClientWithConfig(config)
}
