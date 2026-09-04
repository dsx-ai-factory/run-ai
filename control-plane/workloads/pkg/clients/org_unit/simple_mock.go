package org_unit

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/structs"
)

type SimpleMock struct {
	ProjectInfoById     map[string]orgUnitApi.Project
	DepartmentNamesById map[string]string
	Randoms             *testutils.RandomValues
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		ProjectInfoById:     make(map[string]orgUnitApi.Project),
		DepartmentNamesById: make(map[string]string),
		Randoms:             nil,
	}
}

func (smc *SimpleMock) WithProject(id string, project orgUnitApi.Project) *SimpleMock {
	smc.ProjectInfoById[id] = project
	return smc
}

func (smc *SimpleMock) WithDepartment(id string, name string) *SimpleMock {
	smc.DepartmentNamesById[id] = name
	return smc
}

func (smc *SimpleMock) WithRandoms(randoms *testutils.RandomValues) *SimpleMock {
	smc.Randoms = randoms
	return smc
}

func (smc *SimpleMock) GetProjectById(_ context.Context, projectId string) (*orgUnitApi.Project, *orgUnitApi.Error) {
	if result, ok := smc.ProjectInfoById[projectId]; ok {
		return &result, nil
	}
	return nil, orgUnitApi.NewError(http.StatusNotFound, "project does not exist")
}

func (smc *SimpleMock) GetAncestorsOfScope(_ context.Context, scopeType orgUnitApi.ScopeType, scopeId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	if smc.Randoms == nil {
		return nil, orgUnitApi.NewError(http.StatusNotFound, "randoms not set")
	}

	clusterUuid := orgUnitApi.NewNullableString(&smc.Randoms.ClusterId)
	departmentId := orgUnitApi.NewNullableString(&smc.Randoms.DepartmentId)
	projectId := orgUnitApi.NewNullableString(&smc.Randoms.ProjectIdStr)
	response := orgUnitApi.Ancestors{
		TenantId:     smc.Randoms.TenantIdStr,
		ClusterUuid:  *clusterUuid,
		DepartmentId: *departmentId,
		ProjectId:    *projectId,
	}

	return &response, nil
}

func (smc *SimpleMock) GetProjectsMetadataByDepartment(_ context.Context, departmentId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	return nil, orgUnitApi.NewError(http.StatusNotFound, "projects metadata by department not found")
}

func (smc *SimpleMock) GetOrgTreeMap(_ context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error) {
	nameMap := map[structs.OrgUnitKey]string{}
	for id, name := range smc.DepartmentNamesById {
		nameMap[structs.OrgUnitKey{Scope: string(orgUnitApi.SCOPETYPE_DEPARTMENT), Id: id}] = name
	}
	return structs.OrgUnitMap{}, nameMap, nil
}

func (smc *SimpleMock) GetDepartmentName(_ context.Context, departmentId string) (string, *orgUnitApi.Error) {
	if name, ok := smc.DepartmentNamesById[departmentId]; ok {
		return name, nil
	}
	return "", nil
}
