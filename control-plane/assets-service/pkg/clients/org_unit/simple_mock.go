/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

type SimpleMock struct {
	authMock *testing.RandomsBasedAuthMockClient
}

func NewSimpleMock(randoms ...*testutils.RandomValues) ClientInterface {
	return newCacheMockClient(randoms...)
}

func newSimpleMock(randoms ...*testutils.RandomValues) ClientInterface {
	authMock := testing.NewRandomsBasedAuthMockClient(randoms...)
	orgunit.SetMock(newOrgUnitFactoryMock(authMock))
	return &SimpleMock{
		authMock: authMock,
	}
}

// orgUnitFactoryMock adapts RandomsBasedAuthMockClient to orgunit.FactoryInterface and orgunit.ClientInterface
type orgUnitFactoryMock struct {
	authMock *testing.RandomsBasedAuthMockClient
}

// newOrgUnitFactoryMock returns an orgunit.FactoryInterface mock backed by the provided auth mock data
func newOrgUnitFactoryMock(authMock *testing.RandomsBasedAuthMockClient) orgunit.FactoryInterface {
	return &orgUnitFactoryMock{authMock: authMock}
}

func (a *orgUnitFactoryMock) ListProjects(_ context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	result := make([]orgUnitApi.ProjectMetadata, 0)

	for _, project := range a.authMock.AllowedProjects {
		if project.ClusterId == clusterId {
			result = append(result, *project)
		}
	}

	return result, nil
}

func (a *orgUnitFactoryMock) GetAncestors(ctx context.Context, scopeType, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	return a.authMock.GetAncestors(ctx, scope.Scope(scopeType), resourceId)
}

func (a *orgUnitFactoryMock) Client() orgunit.ClientInterface {
	return a
}

func (a *orgUnitFactoryMock) GetOrgTree(ctx context.Context) ([]orgUnitApi.OrgUnit, *orgUnitApi.Error) {
	result := make([]orgUnitApi.OrgUnit, 0)

	ctxTenantId := auth.TenantIdOf(ctx)
	if ctxTenantId != a.authMock.TenantId {
		return result, nil
	}

	if strings.HasPrefix(auth.SubjectIdOf(ctx), "admin") {
		for clusterId := range a.authMock.AllClusters {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_CLUSTER, Id: clusterId})
		}
		for departmentId := range a.authMock.AllDepartments {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_DEPARTMENT, Id: departmentId})
		}
		for projectId := range a.authMock.AllProjects {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_PROJECT, Id: projectId})
		}
	} else {
		for clusterId := range a.authMock.AllClusters {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_CLUSTER, Id: clusterId})
		}
		for departmentId := range a.authMock.AllowedDepartments {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_DEPARTMENT, Id: departmentId})
		}
		for projectId := range a.authMock.AllowedProjects {
			result = append(result, orgUnitApi.OrgUnit{Type: orgUnitApi.SCOPETYPE_PROJECT, Id: projectId})
		}
	}

	return result, nil
}

func (smc *SimpleMock) GetProjectAncestors(ctx context.Context, projectId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	ancestors, err := smc.authMock.GetAncestors(ctx, scope.Project, projectId)
	if err != nil {
		return nil, orgUnitApi.NewError(err.Code, err.Message)
	}

	clusterUuid := orgUnitApi.NewNullableString(ancestors.ClusterUuid.Get())
	departmentId := orgUnitApi.NewNullableString(ancestors.DepartmentId.Get())
	orgProjectId := orgUnitApi.NewNullableString(ancestors.ProjectId.Get())
	orgAncestors := &orgUnitApi.Ancestors{
		TenantId:     ancestors.TenantId,
		ClusterUuid:  *clusterUuid,
		DepartmentId: *departmentId,
		ProjectId:    *orgProjectId,
	}
	return orgAncestors, nil
}

func (smc *SimpleMock) GetDepartmentAncestors(ctx context.Context, departmentId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	ancestors, err := smc.authMock.GetAncestors(ctx, scope.Department, departmentId)
	if err != nil {
		return nil, orgUnitApi.NewError(err.Code, err.Message)
	}

	clusterUuid := orgUnitApi.NewNullableString(ancestors.ClusterUuid.Get())
	orgAncestors := &orgUnitApi.Ancestors{
		TenantId:    ancestors.TenantId,
		ClusterUuid: *clusterUuid,
	}
	return orgAncestors, nil
}

func (smc SimpleMock) GetProjectsMetadataByDepartment(_ context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	var departmentProjects []orgUnitApi.ProjectMetadata

	for _, project := range smc.authMock.AllowedProjects {
		departmentOfProject := smc.authMock.DepartmentOfProject[project.Id]

		if departmentOfProject == departmentId {
			projectMetadata := orgUnitApi.ProjectMetadata{
				Id: project.Id,
			}
			departmentProjects = append(departmentProjects, projectMetadata)
		}
	}

	return departmentProjects, nil
}

func (smc SimpleMock) GetProjectsMetadataByCluster(_ context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	var clusterProjects []orgUnitApi.ProjectMetadata

	for _, project := range smc.authMock.AllowedProjects {
		if project.ClusterId == clusterId {
			projectMetadata := orgUnitApi.ProjectMetadata{
				Id: project.Id,
			}
			clusterProjects = append(clusterProjects, projectMetadata)
		}
	}

	return clusterProjects, nil
}

func (smc SimpleMock) GetAncestorsOfScope(_ context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	switch scopeType {
	case orgUnitApi.SCOPETYPE_PROJECT:
		return smc.GetProjectAncestors(context.Background(), scopeId)
	case orgUnitApi.SCOPETYPE_DEPARTMENT:
		return smc.GetDepartmentAncestors(context.Background(), scopeId)
	default:
		return nil, orgUnitApi.NewError(http.StatusInternalServerError, fmt.Sprintf("Invalid scope type: %v", scopeType))
	}
}

func newCacheMockClient(randoms ...*testutils.RandomValues) ClientInterface {
	client := newSimpleMock(randoms...)
	return newCacheClient(client)
}
