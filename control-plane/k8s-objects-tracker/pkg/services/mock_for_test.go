/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// RandomsBasedBackendMockClient is a mock client that works based on the hierarchy of RandomScopes
type RandomsBasedBackendMockClient struct {
	TenantId int32

	AllProjects     map[string]orgUnitApi.ProjectMetadata
	AllowedProjects map[string]orgUnitApi.ProjectMetadata

	TranslationTable map[string]*RandomScopes
}

func (rbc *RandomsBasedBackendMockClient) Client() orgUnitClient.ClientInterface {
	return rbc
}

func NewRandomsBasedBackendMockClient(randoms ...*RandomScopes) *RandomsBasedBackendMockClient {
	result := RandomsBasedBackendMockClient{
		TranslationTable: map[string]*RandomScopes{},

		AllProjects:     map[string]orgUnitApi.ProjectMetadata{},
		AllowedProjects: map[string]orgUnitApi.ProjectMetadata{},
	}

	for idx, randomsSet := range randoms {
		if idx == 0 {
			result.TenantId = randomsSet.TenantId
		}

		result.TranslationTable[randomsSet.ClusterId] = randomsSet
		result.TranslationTable[randomsSet.DepartmentId] = randomsSet
		result.TranslationTable[randomsSet.ProjectIdStr] = randomsSet

		ns := fmt.Sprintf("runai-%d", randomsSet.ProjectId)

		projectInfo := orgUnitApi.ProjectMetadata{
			Id:        randomsSet.ProjectIdStr,
			Name:      randomsSet.ProjectIdStr,
			Namespace: *orgUnitApi.NewNullableString(&ns),
		}

		result.AllProjects[randomsSet.ProjectIdStr] = projectInfo

		if !randomsSet.IsForbidden {
			result.AllowedProjects[randomsSet.ProjectIdStr] = projectInfo
		}
	}

	return &result
}

func (rbm *RandomsBasedBackendMockClient) GetAncestorsOf(_ context.Context, scopeType string, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	randoms, ok := rbm.TranslationTable[resourceId]
	if !ok {
		return nil, orgUnitApi.NewError(int32(http.StatusNotFound), fmt.Sprintf("%s %s does not exist", scopeType, resourceId))
	}

	response := orgUnitApi.Ancestors{
		TenantId: randoms.TenantIdStr,
	}

	if scopeType == string(orgUnitApi.SCOPETYPE_CLUSTER) || scopeType == string(orgUnitApi.SCOPETYPE_DEPARTMENT) || scopeType == string(orgUnitApi.SCOPETYPE_PROJECT) {
		response.SetClusterUuid(randoms.ClusterId)
	}
	if scopeType == string(orgUnitApi.SCOPETYPE_DEPARTMENT) || scopeType == string(orgUnitApi.SCOPETYPE_PROJECT) {
		response.SetDepartmentId(randoms.DepartmentId)
	}
	if scopeType == string(orgUnitApi.SCOPETYPE_PROJECT) {
		response.SetProjectId(randoms.ProjectIdStr)
	}

	return &response, nil
}

func (rbm *RandomsBasedBackendMockClient) GetAncestorsOfScope(ctx context.Context, scopeType orgUnitApi.ScopeType, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	return rbm.GetAncestorsOf(ctx, string(scopeType), resourceId)
}

func (rbm *RandomsBasedBackendMockClient) GetProjectsMetadataByCluster(ctx context.Context, _ string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	result := make([]orgUnitApi.ProjectMetadata, 0, len(rbm.AllProjects))

	if isAdmin(ctx) {
		for _, projectInfo := range rbm.AllProjects {
			result = append(result, projectInfo)
		}
	} else {
		for _, projectInfo := range rbm.AllowedProjects {
			result = append(result, projectInfo)
		}
	}

	return result, nil
}

func isAdmin(ctx context.Context) bool {
	requester := ctx.Value(runaictx.CfSubjectID).(string)
	return requester == "admin"
}

func (rbm *RandomsBasedBackendMockClient) CheckPermittedScopes(ctx context.Context, _ authorizationClient.ResourceType, action *authorizationClient.Action) (*authorizationClient.PermittedScopesActions, *api.Error) {
	scopes := authorizationClient.PermittedScopes{
		Clusters: []string{},
	}

	if isAdmin(ctx) {
		scopes.Tenants = append(scopes.Tenants, fmt.Sprintf("%d", rbm.TenantId))
		for projectId := range rbm.AllProjects {
			scopes.Projects = append(scopes.Projects, projectId)
		}
	} else {
		for projectId := range rbm.AllowedProjects {
			scopes.Projects = append(scopes.Projects, projectId)
		}
	}

	result := authorizationClient.PermittedScopesActions{}

	if action == nil || *action == authorizationClient.ACTION_READ {
		result.Read = scopes
	}
	if action == nil || *action == authorizationClient.ACTION_CREATE {
		result.Create = scopes
	}
	if action == nil || *action == authorizationClient.ACTION_DELETE {
		result.Delete = scopes
	}
	if action == nil || *action == authorizationClient.ACTION_UPDATE {
		result.Update = scopes
	}

	return &result, nil
}
