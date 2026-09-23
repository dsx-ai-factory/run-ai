/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing_crud

import (
	"context"
	"fmt"
	"net/http"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

type ScopeTypeAndId struct {
	ScopeType scopeinfo.Scope
	ScopeId   string
}

// RandomsBasedAuthMockClient is a mock client that is based on RandomValues
type RandomsBasedAuthMockClient struct {
	RandomsByKey map[ScopeTypeAndId]*testutils.RandomValues

	ParentByKey map[ScopeTypeAndId]ScopeTypeAndId

	TenantSettings map[uint]datastructs.Set[string]
}

func NewRandomsBasedAuthMockClient() *RandomsBasedAuthMockClient {
	return &RandomsBasedAuthMockClient{
		RandomsByKey: make(map[ScopeTypeAndId]*testutils.RandomValues),
		ParentByKey:  make(map[ScopeTypeAndId]ScopeTypeAndId),
	}
}

func (mock *RandomsBasedAuthMockClient) Client() authorization.ClientInterface {
	return mock
}

// With adds a set of RandomValues to the mock. it adds the set of values as a project, department and cluster,
// so we will be able to access them all. it is the responsibility of the caller to be consistent with the
// values, i.e. not to send two randomValues for the same project but with different value for department
func (mc *RandomsBasedAuthMockClient) With(randomValues *testutils.RandomValues) *RandomsBasedAuthMockClient {
	projKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Project,
		ScopeId:   randomValues.ProjectIdStr,
	}

	mc.RandomsByKey[projKey] = randomValues

	deptKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Department,
		ScopeId:   randomValues.DepartmentId,
	}

	mc.RandomsByKey[deptKey] = randomValues

	clusterKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Cluster,
		ScopeId:   randomValues.ClusterId,
	}

	mc.RandomsByKey[clusterKey] = randomValues

	tenantKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Tenant,
		ScopeId:   randomValues.TenantIdStr,
	}

	mc.RandomsByKey[tenantKey] = randomValues

	mc.ParentByKey[projKey] = deptKey
	mc.ParentByKey[deptKey] = clusterKey
	mc.ParentByKey[clusterKey] = tenantKey

	return mc
}

func (mc *RandomsBasedAuthMockClient) WithSpecificProject(randomValues *testutils.RandomValues) *RandomsBasedAuthMockClient {
	projKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Project,
		ScopeId:   randomValues.ProjectIdStr,
	}

	mc.RandomsByKey[projKey] = randomValues

	deptKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Department,
		ScopeId:   randomValues.DepartmentId,
	}

	mc.ParentByKey[projKey] = deptKey

	return mc
}

func (mc *RandomsBasedAuthMockClient) WithDepartment(randomValues *testutils.RandomValues) *RandomsBasedAuthMockClient {
	deptKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Department,
		ScopeId:   randomValues.DepartmentId,
	}

	mc.RandomsByKey[deptKey] = randomValues

	clusterKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Cluster,
		ScopeId:   randomValues.ClusterId,
	}

	mc.RandomsByKey[clusterKey] = randomValues

	tenantKey := ScopeTypeAndId{
		ScopeType: scopeinfo.Tenant,
		ScopeId:   randomValues.TenantIdStr,
	}

	mc.RandomsByKey[tenantKey] = randomValues

	mc.ParentByKey[deptKey] = clusterKey
	mc.ParentByKey[clusterKey] = tenantKey

	return mc
}

func (mc *RandomsBasedAuthMockClient) WithSettings(values *testutils.RandomValues, settingsKey string) *RandomsBasedAuthMockClient {
	if mc.TenantSettings == nil {
		mc.TenantSettings = make(map[uint]datastructs.Set[string])
	}

	if mc.TenantSettings[values.TenantId] == nil {
		mc.TenantSettings[values.TenantId] = datastructs.NewSet[string]()
	}

	mc.TenantSettings[values.TenantId].Add(settingsKey)

	return mc
}

func (mc *RandomsBasedAuthMockClient) GetAncestors(_ context.Context, scope scopeinfo.Scope, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	randoms, ok := mc.RandomsByKey[ScopeTypeAndId{
		ScopeType: scope,
		ScopeId:   resourceId,
	}]
	if !ok {
		return nil, orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("%s %s does not exist", scope, resourceId))
	}

	response := orgUnitApi.Ancestors{
		TenantId: randoms.TenantIdStr,
	}
	response.SetClusterUuid(randoms.ClusterId)
	response.SetDepartmentId(randoms.DepartmentId)
	response.SetProjectId(randoms.ProjectIdStr)

	return &response, nil
}

// AllowedProjectsToContext adds to a permitted scopes element to the context, containing
// all the projects the user is allowed to see
func (mc *RandomsBasedAuthMockClient) AllowedProjectsToContext(ctx *context.Context) {
	permittedScopes := authClient.PermittedScopes{
		Clusters:    make([]string, 0),
		Departments: make([]string, 0),
		Projects:    make([]string, 0),
	}

	allProjects := datastructs.Set[string]{}

	for _, randomValues := range mc.RandomsByKey {
		if !allProjects.Has(randomValues.ProjectIdStr) {
			if !randomValues.IsForbidden {
				permittedScopes.Projects = append(permittedScopes.Projects, randomValues.ProjectIdStr)
				allProjects.Add(randomValues.ProjectIdStr)
			}
		}
	}

	*ctx = context.WithValue(*ctx, authorization_utils.PermittedScopesKey, &permittedScopes)
}

// AllowedDepartmentsToContext adds to a permitted scopes element to the context, containing
// all the departments the user is allowed to see
func (mc *RandomsBasedAuthMockClient) AllowedDepartmentsToContext(ctx *context.Context) {
	permittedScopes := authClient.PermittedScopes{
		Clusters:    make([]string, 0),
		Departments: make([]string, 0),
		Projects:    make([]string, 0),
	}

	allDepartments := datastructs.Set[string]{}

	for _, randomValues := range mc.RandomsByKey {
		if !allDepartments.Has(randomValues.DepartmentId) {
			if !randomValues.IsForbidden {
				permittedScopes.Departments = append(permittedScopes.Departments, randomValues.DepartmentId)
				allDepartments.Add(randomValues.DepartmentId)
			}
		}
	}

	*ctx = context.WithValue(*ctx, authorization_utils.PermittedScopesKey, &permittedScopes)
}

func (mc *RandomsBasedAuthMockClient) CheckPermittedScopes(ctx context.Context, _ authClient.ResourceType, _ authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	permittedScopes := authClient.PermittedScopes{
		Clusters:    make([]string, 0),
		Departments: make([]string, 0),
		Projects:    make([]string, 0),
	}

	if isAdmin(ctx) {
		tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
		if tenantId != nil {
			permittedScopes.Tenants = append(permittedScopes.Tenants, *tenantId)
		}
		return &permittedScopes, nil
	}

	allClusters := datastructs.Set[string]{}
	allDepartments := datastructs.Set[string]{}
	allProjects := datastructs.Set[string]{}

	for _, randomValues := range mc.RandomsByKey {
		if randomValues.IsForbidden {
			continue
		}

		if !allDepartments.Has(randomValues.ClusterId) {
			permittedScopes.Departments = append(permittedScopes.Departments, randomValues.ClusterId)
			allClusters.Add(randomValues.ClusterId)
		}
		if !allDepartments.Has(randomValues.DepartmentId) {
			permittedScopes.Departments = append(permittedScopes.Departments, randomValues.DepartmentId)
			allDepartments.Add(randomValues.DepartmentId)
		}
		if !allDepartments.Has(randomValues.ProjectIdStr) {
			permittedScopes.Projects = append(permittedScopes.Projects, randomValues.ProjectIdStr)
			allProjects.Add(randomValues.ProjectIdStr)
		}
	}

	return &permittedScopes, nil
}

func isAdmin(ctx context.Context) bool {
	return auth.SubjectIdOf(ctx) == "admin"
}
