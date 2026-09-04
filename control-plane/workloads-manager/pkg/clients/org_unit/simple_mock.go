package org_unit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

type SimpleMock struct {
	ProjectInfoById map[string]orgUnitApi.Project
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		ProjectInfoById: make(map[string]orgUnitApi.Project),
	}
}

func (smc *SimpleMock) WithProject(id string, project orgUnitApi.Project) *SimpleMock {
	smc.ProjectInfoById[id] = project
	return smc
}

func (smc *SimpleMock) GetProjectById(_ context.Context, projectId string) (*orgUnitApi.Project, *orgUnitApi.Error) {
	if result, ok := smc.ProjectInfoById[projectId]; ok {
		return &result, nil
	}

	return nil, orgUnitApi.NewError(http.StatusNotFound, "project does not exist")
}

func (smc *SimpleMock) WithRandoms(randoms *testutils.RandomValues) *SimpleMock {
	project := orgUnitApi.Project{
		Id:               randoms.ProjectIdStr,
		Name:             randoms.ProjectIdStr,
		ClusterId:        randoms.ClusterId,
		ParentId:         randoms.DepartmentId,
		DefaultNodePools: []string{"default"},
		Effective: orgUnitApi.ProjectEffective{
			DefaultNodePools: []string{"default"},
		},
		Parent: *orgUnitApi.NewNullableDepartmentProjectParent(&orgUnitApi.DepartmentProjectParent{
			Id:   randoms.DepartmentId,
			Name: randoms.DepartmentId,
		}),
	}

	// Set default namespace for testing
	project.SetRequestedNamespace(fmt.Sprintf("runai-%s", randoms.ProjectIdStr))

	smc.ProjectInfoById[randoms.ProjectIdStr] = project

	return smc
}

// WithNamespaceForProject sets a custom namespace for a specific project
func (smc *SimpleMock) WithNamespaceForProject(projectId string, namespace string) *SimpleMock {
	if project, ok := smc.ProjectInfoById[projectId]; ok {
		project.SetRequestedNamespace(namespace)
		smc.ProjectInfoById[projectId] = project
	}
	return smc
}

func (smc *SimpleMock) GetProjectsMetadataByCluster(_ context.Context, clusterId string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	projects := make([]orgUnitApi.ProjectMetadata, 0)
	for _, project := range smc.ProjectInfoById {
		if project.ClusterId == clusterId {
			namespace := ""
			if project.Status.Namespace.IsSet() && project.Status.Namespace.Get() != nil {
				namespace = *project.Status.Namespace.Get()
			} else if project.HasRequestedNamespace() {
				namespace = project.GetRequestedNamespace()
			}

			metadata := orgUnitApi.ProjectMetadata{
				Id:        project.Id,
				Name:      project.Name,
				ClusterId: project.ClusterId,
				Namespace: *orgUnitApi.NewNullableString(&namespace),
			}
			projects = append(projects, metadata)
		}
	}
	return projects, nil
}
