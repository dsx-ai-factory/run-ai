/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"net/http"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clusterMocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	tenants_client "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Internal Org Tree API", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)

	var (
		orgUnitService       api.OrgUnitAPIServicer
		departmentApiService api.DepartmentsAPIServicer
		projectApiService    api.ProjectsAPIServicer
		clusterServiceClient clusterMocks.IClusterClient
		clusterId            string
		createdDepartmentId  string
		createdProjectId     string
	)

	BeforeAll(func() {
		clusterId = uuid.New().String()
		clusterServiceClient = clusterMocks.IClusterClient{}

		// Clusters used by org tree (GetClusters)
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
			Name:     "cluster-1",
			Uuid:     clusterId,
			TenantId: 1,
		}, nil)

		// Clusters used by org tree (GetClusters)
		clusterServiceClient.On("GetClustersMinimalData", mock.Anything).Return(&clustersOpenapi.GetClustersMinimalData200Response{
			Clusters: []clustersOpenapi.ClustersMinimalData{
				{
					Name:     "cluster-1",
					Uuid:     clusterId,
					TenantId: 1,
				},
			},
		}, nil)

		// Clusters used by org tree (GetClusters)
		clusterServiceClient.On("GetInternalClusters", mock.Anything, mock.Anything).Return([]clustersOpenapi.DisplayedCluster{
			{
				Uuid:     clusterId,
				Name:     "cluster-1",
				TenantId: 1,
			},
			{
				Uuid:     "cluster-2",
				Name:     "cluster-2",
				TenantId: 1,
			},
		}, nil)

		// Needed by departments/projects flows
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
			Uuid:    clusterId,
			Version: *clustersOpenapi.NewNullableString(ptr.To("2.21")),
		}, nil)
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
			Nodepools: []clustersOpenapi.NodepoolMinimal{
				{Name: "default", Id: "1", SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false)},
			},
		}, nil)

		// Tenants manager used by GetTenantOrgUnit
		mockTenantsClient.Mock.On("GetTenantByIdInternal", mock.Anything, int32(1)).Return(&tenants_client.Tenant{Id: 1, Name: "Tenant-1"}, &http.Response{StatusCode: 200}, nil)

		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)

		// Build service under test with mocks
		orgUnitService = services.NewOrgUnitApiService(
			authorizationclient.NewAuthorizationClient(),
			&clusterServiceClient,
			mockTenantsClient,
			&clusterServiceClient,
		)
	})

	AfterAll(func() {
		if createdProjectId != "" {
			deleteResponse, err := projectApiService.DeleteProject(ctx, createdProjectId, "", false)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			createdProjectId = ""
		}
		if createdDepartmentId != "" {
			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, createdDepartmentId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			createdDepartmentId = ""
		}
	})

	It("returns full tree without filtering by permitted scopes", func() {
		// Seed a department and a project
		depReq := getFullDepartmentCreateRequest("dep-internal", clusterId)
		depResp, err := departmentApiService.CreateDepartment(ctx, depReq)
		Expect(err).To(BeNil())
		Expect(depResp.Code).To(Equal(http.StatusCreated))
		dep := depResp.Body.(api.Department)
		createdDepartmentId = dep.Id

		projName := createProjectName()
		projReq := getFullProjectCreateRequest(projName, "", &dep.Id, clusterId)
		projResp, err := projectApiService.CreateProject(ctx, projReq)
		Expect(err).To(BeNil())
		Expect(projResp.Code).To(Equal(http.StatusCreated))
		createdProjectId = projResp.Body.(api.Project).Id

		// Call internal org tree
		resp, err := orgUnitService.GetInternalOrgTree(ctx)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))
		body := resp.Body.(api.GetInternalOrgTree200Response)

		// Verify tenant, cluster, department, project exist
		types := map[api.ScopeType]bool{}
		ids := map[string]bool{}
		for _, u := range body.OrgUnits {
			types[u.Type] = true
			ids[u.Id] = true
		}
		Expect(types[api.SCOPETYPE_TENANT]).To(BeTrue())
		Expect(types[api.SCOPETYPE_CLUSTER]).To(BeTrue())
		Expect(types[api.SCOPETYPE_DEPARTMENT]).To(BeTrue())
		Expect(types[api.SCOPETYPE_PROJECT]).To(BeTrue())
		Expect(ids[dep.Id]).To(BeTrue())

		resp2, err := orgUnitService.GetOrgUnitsNames(ctx, "project", createdProjectId)
		Expect(err).To(BeNil())
		Expect(resp2.Code).To(Equal(http.StatusOK))
		names := resp2.Body.([]api.OrgUnit)
		Expect(len(names)).To(Equal(1))
		Expect(names[0].Name).To(Equal(projName))
		Expect(names[0].Id).To(Equal(createdProjectId))
		Expect(names[0].Type).To(Equal(api.SCOPETYPE_PROJECT))
	})

	It("GetOrgTree is filtered by permitted scopes while GetInternalOrgTree is full", func() {
		// Create two departments and projects
		depReq1 := getFullDepartmentCreateRequest("dep-filt-1", clusterId)
		depResp1, err := departmentApiService.CreateDepartment(ctx, depReq1)
		Expect(err).To(BeNil())
		Expect(depResp1.Code).To(Equal(http.StatusCreated))
		dep1 := depResp1.Body.(api.Department)

		depReq2 := getFullDepartmentCreateRequest("dep-filt-2", clusterId)
		depResp2, err := departmentApiService.CreateDepartment(ctx, depReq2)
		Expect(err).To(BeNil())
		Expect(depResp2.Code).To(Equal(http.StatusCreated))
		dep2 := depResp2.Body.(api.Department)

		projReq1 := getFullProjectCreateRequest(createProjectName(), "", &dep1.Id, clusterId)
		projResp1, err := projectApiService.CreateProject(ctx, projReq1)
		Expect(err).To(BeNil())
		Expect(projResp1.Code).To(Equal(http.StatusCreated))
		proj1 := projResp1.Body.(api.Project)

		projReq2 := getFullProjectCreateRequest(createProjectName(), "", &dep2.Id, clusterId)
		projResp2, err := projectApiService.CreateProject(ctx, projReq2)
		Expect(err).To(BeNil())
		Expect(projResp2.Code).To(Equal(http.StatusCreated))
		proj2 := projResp2.Body.(api.Project)

		// Build permitted scopes only for dep1/proj1
		permitted := &authorization_api.PermittedScopesActions{
			Read: authorization_api.PermittedScopes{
				Departments: []string{dep1.Id},
				Projects:    []string{proj1.Id},
			},
		}
		authMock := auth_mocks.IAuthorizationClient{}
		authMock.On("GetPermittedScopes", mock.Anything, authorization_api.RESOURCETYPE_DEPARTMENT, mock.Anything).Return(permitted, 0, nil)
		authMock.On("GetPermittedScopes", mock.Anything, authorization_api.RESOURCETYPE_PROJECT, mock.Anything).Return(permitted, 0, nil)

		orgUnitServiceFiltered := services.NewOrgUnitApiService(&authMock, &clusterServiceClient, mockTenantsClient, &clusterServiceClient)

		// Get filtered org tree
		respFiltered, err := orgUnitServiceFiltered.GetOrgTree(ctx)
		Expect(err).To(BeNil())
		Expect(respFiltered.Code).To(Equal(http.StatusOK))
		filtered := respFiltered.Body.(api.GetInternalOrgTree200Response)

		// Get full internal org tree
		respFull, err := orgUnitServiceFiltered.GetInternalOrgTree(ctx)
		Expect(err).To(BeNil())
		Expect(respFull.Code).To(Equal(http.StatusOK))
		full := respFull.Body.(api.GetInternalOrgTree200Response)

		// Helpers
		contains := func(units []api.OrgUnit, id string, t api.ScopeType) bool {
			for _, u := range units {
				if u.Id == id && u.Type == t {
					return true
				}
			}
			return false
		}

		// Filtered tree should include dep1/proj1 but not dep2/proj2
		Expect(contains(filtered.OrgUnits, dep1.Id, api.SCOPETYPE_DEPARTMENT)).To(BeTrue())
		Expect(contains(filtered.OrgUnits, proj1.Id, api.SCOPETYPE_PROJECT)).To(BeTrue())
		Expect(contains(filtered.OrgUnits, dep2.Id, api.SCOPETYPE_DEPARTMENT)).To(BeFalse())
		Expect(contains(filtered.OrgUnits, proj2.Id, api.SCOPETYPE_PROJECT)).To(BeFalse())

		// Full internal tree should include both
		Expect(contains(full.OrgUnits, dep1.Id, api.SCOPETYPE_DEPARTMENT)).To(BeTrue())
		Expect(contains(full.OrgUnits, dep2.Id, api.SCOPETYPE_DEPARTMENT)).To(BeTrue())
		Expect(contains(full.OrgUnits, proj1.Id, api.SCOPETYPE_PROJECT)).To(BeTrue())
		Expect(contains(full.OrgUnits, proj2.Id, api.SCOPETYPE_PROJECT)).To(BeTrue())

		// Cleanup entities from this test
		_, _ = projectApiService.DeleteProject(ctx, proj1.Id, "", false)
		_, _ = projectApiService.DeleteProject(ctx, proj2.Id, "", false)
		_, _ = departmentApiService.DeleteDepartment(ctx, dep1.Id)
		_, _ = departmentApiService.DeleteDepartment(ctx, dep2.Id)
	})
})
