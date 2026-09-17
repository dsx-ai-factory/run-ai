/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/google/uuid"
)

// backendMockClient implements the following simple hierarchy
//
//	tenant-1 -> dept-11 -> proj-111, 112
//	         -> dept-12 -> proj-121, 122
//	tenant-2 -> dept-21 -> proj-211, 212
//	         -> dept-22 -> proj-221, 222
//
// So if we need ancestors of project 212 we return department 1 tenant 2
// The only parameter we need is the maximum org unit entities for each level (2 in this case)
// for any org unit with number higher than this maximum we return 404 not found
// cluster id we are returning "cluster-t" where t is the tenant number
type BackendMockClient struct {
	MaxOrgUnitsPerLevel int

	// we allow multiple clusters for the same tenant, but all the projects and departments are
	// considered to be related to the first cluster
	ClusterIdPerTenant map[string][]string

	// for authentication purposes, we have this action to permitted scopes mapping, that test can set using SetPermittedScope
	// function below. anything not in the map will return permitted scopes that allows access for the entire tenant.
	PermittedScopes map[authClient.Action]*authClient.PermittedScopes

	// DeletedScopeEntities tracks scope entities that have been "deleted" from the backend.
	// Key format: "scopeType:resourceId" (e.g. "department:111", "project:1111").
	// When GetAncestors is called for one of these, it returns 404.
	DeletedScopeEntities map[string]bool
}

func NewBackendMockClient(maxOrgUnitsPerLevel int, numClusters int) *BackendMockClient {
	if numClusters <= 0 {
		numClusters = 1 // minimum 1 cluster per tenant
	}

	mock := BackendMockClient{
		MaxOrgUnitsPerLevel: maxOrgUnitsPerLevel,
		ClusterIdPerTenant:  make(map[string][]string),
	}
	for tenant := 1; tenant <= maxOrgUnitsPerLevel; tenant += 1 {
		clustersOfThisTenant := make([]string, 0, numClusters)

		for clusterIdx := 0; clusterIdx < numClusters; clusterIdx++ {
			clusterId := uuid.NewString()
			clustersOfThisTenant = append(clustersOfThisTenant, clusterId)
		}

		mock.ClusterIdPerTenant[strconv.Itoa(tenant)] = clustersOfThisTenant
	}

	mock.PermittedScopes = make(map[authClient.Action]*authClient.PermittedScopes)
	mock.DeletedScopeEntities = make(map[string]bool)

	return &mock
}

type PoliciesAuthClient struct {
	backendMockClient *BackendMockClient
}

func NewPoliciesAuthClient(backendMockClient *BackendMockClient) *PoliciesAuthClient {
	return &PoliciesAuthClient{backendMockClient: backendMockClient}
}

func (mc *PoliciesAuthClient) Client() authorization.ClientInterface {
	return mc
}

func (mc *PoliciesAuthClient) CheckPermittedScopes(_ context.Context, _ /*resourceType*/ authClient.ResourceType,
	action authClient.Action,
) (*authClient.PermittedScopes, *authClient.Error) {
	if mc.backendMockClient.PermittedScopes[action] != nil {
		return mc.backendMockClient.PermittedScopes[action], nil
	}
	return &authClient.PermittedScopes{
		Tenants: []string{"1", "2"},
	}, nil
}

func (mc *PoliciesAuthClient) SetPermittedScopes(action authClient.Action, permittedScopes *authClient.PermittedScopes) {
	mc.backendMockClient.PermittedScopes[action] = permittedScopes
}

type PoliciesOrgUnitClient struct {
	backendMockClient *BackendMockClient
}

func NewPoliciesOrgUnitClient(backendMockClient *BackendMockClient) *PoliciesOrgUnitClient {
	return &PoliciesOrgUnitClient{backendMockClient: backendMockClient}
}

func (mc *PoliciesOrgUnitClient) Client() orgunit.ClientInterface {
	return mc
}

func (mc *PoliciesOrgUnitClient) GetOrgTree(_ context.Context) ([]orgUnitApi.OrgUnit, *orgUnitApi.Error) {
	orgUnits := make([]orgUnitApi.OrgUnit, 0)
	for tenant := 1; tenant <= mc.backendMockClient.MaxOrgUnitsPerLevel; tenant += 1 {
		tenantId := strconv.Itoa(tenant)
		tenantType := orgUnitApi.SCOPETYPE_TENANT
		tenantOrgUnit := orgUnitApi.OrgUnit{
			Type: tenantType,
			Id:   tenantId,
			Name: fmt.Sprintf("tenant-%s", tenantId),
		}
		orgUnits = append(orgUnits, tenantOrgUnit)

		for clusterIdx := 1; clusterIdx <= len(mc.backendMockClient.ClusterIdPerTenant[tenantId]); clusterIdx += 1 {
			clusterId := mc.backendMockClient.ClusterIdPerTenant[tenantId][clusterIdx-1]
			clusterType := orgUnitApi.SCOPETYPE_CLUSTER
			clusterOrgUnit := orgUnitApi.OrgUnit{
				Type:       clusterType,
				Id:         clusterId,
				Name:       fmt.Sprintf("cluster-%s", clusterId),
				ParentId:   &tenantId,
				ParentType: &tenantType,
			}
			orgUnits = append(orgUnits, clusterOrgUnit)

			for deptIdx := 1; deptIdx <= mc.backendMockClient.MaxOrgUnitsPerLevel; deptIdx += 1 {
				deptId := fmt.Sprintf("%s%d%d", tenantId, clusterIdx, deptIdx)
				departmentType := orgUnitApi.SCOPETYPE_DEPARTMENT
				deptOrgUnit := orgUnitApi.OrgUnit{
					Type:       departmentType,
					Id:         deptId,
					Name:       fmt.Sprintf("dept-%s", deptId),
					ParentId:   &clusterId,
					ParentType: &clusterType,
				}
				orgUnits = append(orgUnits, deptOrgUnit)

				for projIdx := 1; projIdx <= mc.backendMockClient.MaxOrgUnitsPerLevel; projIdx += 1 {
					projId := fmt.Sprintf("%s%d", deptId, projIdx)
					projectType := orgUnitApi.SCOPETYPE_PROJECT
					projOrgUnit := orgUnitApi.OrgUnit{
						Type:       projectType,
						Id:         projId,
						Name:       fmt.Sprintf("proj-%s", projId),
						ParentId:   &deptId,
						ParentType: &departmentType,
					}
					orgUnits = append(orgUnits, projOrgUnit)
				}
			}
		}

	}

	return orgUnits, nil
}

func (mc *PoliciesOrgUnitClient) GetOrgTreeMap(ctx context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error) {
	asList, err := mc.GetOrgTree(ctx)
	if err != nil {
		return nil, nil, orgUnitApi.NewError(err.Code, err.Message)
	}

	orgUnitMap := structs.OrgUnitMap{}
	orgUnitToStr := map[structs.OrgUnitKey]string{} // we keep it empty in the mock, no need for it

	for _, entry := range asList {
		orgUnitKey := structs.OrgUnitKey{
			Scope: string(entry.Type),
			Id:    entry.Id,
		}

		if entry.ParentType != nil && entry.ParentId != nil {
			orgUnitMap[orgUnitKey] = structs.OrgUnitKey{
				Scope: string(*entry.ParentType),
				Id:    *entry.ParentId,
			}
		}
	}

	return orgUnitMap, orgUnitToStr, nil
}

func (mc *BackendMockClient) MarkScopeEntityDeleted(scopeType scope.Scope, resourceId string) {
	mc.DeletedScopeEntities[fmt.Sprintf("%s:%s", scopeType, resourceId)] = true
}

func (mc *PoliciesOrgUnitClient) GetAncestors(_ context.Context, scopeType string, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	if mc.backendMockClient.DeletedScopeEntities[fmt.Sprintf("%s:%s", scopeType, resourceId)] {
		return nil, orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("%s '%s' does not exist", scopeType, resourceId))
	}

	response := orgUnitApi.Ancestors{}

	// Handle cluster scope separately since cluster IDs are UUIDs
	if scopeType == string(scope.Cluster) {
		for tenantId, clusterIds := range mc.backendMockClient.ClusterIdPerTenant {
			for _, clusterId := range clusterIds {
				if clusterId == resourceId {
					response.TenantId = tenantId
					response.SetClusterUuid(clusterId)
					return &response, nil
				}
			}
		}
		return nil, orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("cluster '%s' not found", resourceId))
	}

	// For non-cluster scopes, validate digit pattern
	maxDigit := fmt.Sprintf("%d", mc.backendMockClient.MaxOrgUnitsPerLevel)
	for c := 0; c < len(resourceId); c++ {
		if resourceId[c] > maxDigit[0] {
			return nil, orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("%s '%s' not found", scopeType, resourceId))
		}
	}

	switch scope.Scope(scopeType) {
	case scope.Tenant:
		if len(resourceId) != 1 {
			return nil, orgUnitApi.NewError(http.StatusBadRequest, "expecting single digit resource id for tenant")
		}
		response.TenantId = resourceId[0:1]
	case scope.Department:
		if len(resourceId) != 3 {
			return nil, orgUnitApi.NewError(http.StatusBadRequest, "expecting three digits resource id for department")
		}
		response.TenantId = resourceId[0:1]
		clusterIdx := resourceId[1] - '1'
		response.SetClusterUuid(mc.backendMockClient.ClusterIdPerTenant[response.TenantId][clusterIdx])
		response.SetDepartmentId(resourceId[0:3])
	default:
		if len(resourceId) != 4 {
			return nil, orgUnitApi.NewError(http.StatusBadRequest, "expecting four digits resource id for project")
		}
		response.TenantId = resourceId[0:1]
		clusterIdx := resourceId[1] - '1'
		response.SetDepartmentId(resourceId[0:3])
		response.SetClusterUuid(mc.backendMockClient.ClusterIdPerTenant[response.TenantId][clusterIdx])
		response.SetProjectId(resourceId)
	}

	return &response, nil
}

func (mc *PoliciesOrgUnitClient) ListProjects(_ context.Context, _ string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	return nil, orgUnitApi.NewError(http.StatusNotImplemented, "PoliciesBackendMockClient does not implement list projects")
}
