/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"
	"time"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/settings"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Settings tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)
	otherTenantId := "2"
	ctx2 := SetContext(otherTenantId, TestUser)

	var projectIds []string
	var otherTenantProjectId string
	var otherTenantDepartmentId string
	var otherTenantDepartmentResources []api.ResourcesNullable
	var otherTenantProjectResources []api.ResourcesNullable
	var departmentId string
	var departmentId2 string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var orgUnitService api.OrgUnitAPIServicer

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.21")),
	}, nil)
	clusterServiceClient.On("GetClustersMinimalData", mock.Anything).Return(&clustersOpenapi.GetClustersMinimalData200Response{
		Clusters: []clustersOpenapi.ClustersMinimalData{
			{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.21")),
			},
		},
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	BeforeAll(func() {
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId = response.Body.(api.Department).Id
		orgUnitService = services.NewOrgUnitApiService(authorizationclient.NewAuthorizationClient(), &clusterServiceClient, mockTenantsClient, &clusterServiceClient)

		createDepartmentRequest = getFullDepartmentCreateRequest("parent", clusterId)
		otherTenantDepartmentResources = createDepartmentRequest.Resources
		response, err = departmentApiService.CreateDepartment(ctx2, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		otherTenantDepartmentId = response.Body.(api.Department).Id
		name := createProjectName()
		request := getFullProjectCreateRequest(name, name, &otherTenantDepartmentId, clusterId)
		otherTenantProjectResources = request.Resources
		createResponse, err := projectApiService.CreateProject(ctx2, request)
		if err != nil {
			Fail("Failed creating project")
		}
		otherTenantProjectId = createResponse.Body.(api.Project).Id
	})

	AfterAll(func() {
		for _, projectId := range projectIds {
			deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		projectIds = []string{}
		if departmentId2 != "" {
			_, err := departmentApiService.DeleteDepartment(ctx, departmentId2)
			if err != nil {
				Fail("Failed deleting department")
			}
			departmentId2 = ""
		}
		_, err := departmentApiService.DeleteDepartment(ctx, departmentId)
		if err != nil {
			Fail("Failed deleting department")
		}
		if otherTenantProjectId != "" {
			deleteResponse, err := projectApiService.DeleteProject(ctx2, otherTenantProjectId, "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
			otherTenantProjectId = ""
		}
		if otherTenantDepartmentId != "" {
			deleteResponse, err := departmentApiService.DeleteDepartment(ctx2, otherTenantDepartmentId)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
			otherTenantDepartmentId = ""
		}
	})

	Context("test cpu quota flag off", func() {
		BeforeEach(func() {
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now()))
		})

		AfterEach(func() {
			for _, projectId := range projectIds {
				updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			}
			if departmentId2 != "" {
				updateDepStatus(ctx, departmentApiService, departmentId2, clusterId, ptr.To(time.Now()))
			}
		})

		It("test change cpu quota to null - project", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			request.Resources[0].Cpu.Limit = nil
			request.Resources[0].Memory.Limit = nil
			timeBeforeReq := time.Now()
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			requestedResources := request.Resources[0]
			timeAfterReq := time.Now()

			Expect(created.Name).To(Equal(request.Name))
			Expect(len(created.Resources)).To(Equal(1))
			validateResourcesCpu(created.Resources[0], requestedResources, false)

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			validateResourcesCpu(getProject.Resources[0], requestedResources, false)

			getSyncResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			Expect(syncProject.Id).To(Equal(created.Id))
			Expect(syncProject.Name).To(Equal(created.Name))

			updateProjStatus(ctx, projectApiService, created.Id, clusterId, ptr.To(timeAfterReq))

			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  "lala",
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getProject.Resources[0], requestedResources, false)
			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "true",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getProject.Resources[0], requestedResources, false)
			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at changed
			validateResourcesCpu(getProject.Resources[0], requestedResources, true)

			getResponse, err = projectApiService.GetProject(ctx2, otherTenantProjectId)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(otherTenantProjectId))
			validateResourcesCpu(getProject.Resources[0], otherTenantProjectResources[0], false)

			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			Expect(syncProject.Id).To(Equal(created.Id))
			Expect(syncProject.Name).To(Equal(created.Name))
		})

		It("test change cpu quota to null - project unchanged", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			timeBeforeReq := time.Now()
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			requestedResources := request.Resources[0]
			timeAfterReq := time.Now()

			Expect(created.Name).To(Equal(request.Name))
			Expect(len(created.Resources)).To(Equal(1))
			validateResourcesCpu(created.Resources[0], requestedResources, true)

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			validateResourcesCpu(getProject.Resources[0], requestedResources, true)

			getSyncResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			Expect(syncProject.Id).To(Equal(created.Id))
			Expect(syncProject.Name).To(Equal(created.Name))

			updateProjStatus(ctx, projectApiService, created.Id, clusterId, ptr.To(timeAfterReq))

			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  "lala",
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getProject.Resources[0], requestedResources, true)
			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "true",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getProject.Resources[0], requestedResources, true)
			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getProject.Resources[0], requestedResources, true)

			getResponse, err = projectApiService.GetProject(ctx2, otherTenantProjectId)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(otherTenantProjectId))
			validateResourcesCpu(getProject.Resources[0], otherTenantProjectResources[0], false)

			getSyncResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeAfterReq, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getSyncResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))
		})

		It("test change cpu quota to null - department", func() {
			createDepartmentRequest := getFullDepartmentCreateRequest(createDepartmentName(), clusterId)
			timeBeforeReq := time.Now()
			createDepartmentRequest.Resources[0].Cpu.Limit = nil
			createDepartmentRequest.Resources[0].Memory.Limit = nil
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			created := response.Body.(api.Department)
			departmentId2 = created.Id
			requestedResources := createDepartmentRequest.Resources[0]
			timeAfterReq := time.Now()
			Expect(len(created.Resources)).To(Equal(1))
			validateResourcesCpu(created.Resources[0], requestedResources, false)

			getResponse, err := departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDep := getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			validateResourcesCpu(getDep.Resources[0], requestedResources, false)

			getSyncResponse, err := departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments := getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDep := syncDepartments[0]
			Expect(syncDep.Id).To(Equal(created.Id))
			Expect(syncDep.Name).To(Equal(created.Name))

			updateDepStatus(ctx, departmentApiService, created.Id, clusterId, ptr.To(timeAfterReq))

			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  "lala",
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getDep.Resources[0], requestedResources, false)
			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "true",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getDep.Resources[0], requestedResources, false)
			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at changed
			validateResourcesCpu(getDep.Resources[0], requestedResources, true)

			getResponse, err = departmentApiService.GetDepartment(ctx2, otherTenantDepartmentId)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(otherTenantDepartmentId))
			// validate resources and updated_at changed
			validateResourcesCpu(getDep.Resources[0], otherTenantDepartmentResources[0], false)

			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDep = syncDepartments[0]
			Expect(syncDep.Id).To(Equal(created.Id))
			Expect(syncDep.Name).To(Equal(created.Name))
		})

		It("test change cpu quota to null - department unchanged", func() {
			createDepartmentRequest := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			timeBeforeReq := time.Now()
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			created := response.Body.(api.Department)
			departmentId2 = created.Id
			requestedResources := createDepartmentRequest.Resources[0]
			timeAfterReq := time.Now()
			Expect(len(created.Resources)).To(Equal(1))
			validateResourcesCpu(created.Resources[0], requestedResources, true)

			getResponse, err := departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDep := getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			validateResourcesCpu(getDep.Resources[0], requestedResources, true)

			getSyncResponse, err := departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments := getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDep := syncDepartments[0]
			Expect(syncDep.Id).To(Equal(created.Id))
			Expect(syncDep.Name).To(Equal(created.Name))

			updateDepStatus(ctx, departmentApiService, created.Id, clusterId, ptr.To(timeAfterReq))

			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  "lala",
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getDep.Resources[0], requestedResources, true)
			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "true",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getDep.Resources[0], requestedResources, true)
			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			_, err = orgUnitService.UpdateOrgUnitSettings(ctx, api.OrgUnitSingleSetting{
				Name:  services.CpuResourcesFlag,
				Value: "false",
			})
			Expect(err).To(BeNil())
			getResponse, err = departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(created.Id))
			Expect(getDep.Name).To(Equal(created.Name))
			// validate resources and updated_at unchanged
			validateResourcesCpu(getDep.Resources[0], requestedResources, true)

			getResponse, err = departmentApiService.GetDepartment(ctx2, otherTenantDepartmentId)
			if err != nil {
				Fail("Failed getting project")
			}
			getDep = getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(otherTenantDepartmentId))
			// validate resources and updated_at changed
			validateResourcesCpu(getDep.Resources[0], otherTenantDepartmentResources[0], false)

			getSyncResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeAfterReq, 0, 0)
			Expect(err).To(BeNil())
			Expect(getSyncResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getSyncResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))
		})
	})
})

var _ = Describe("OrgUnit settings API", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)
	settingsService := services.NewSettingsService()

	AfterEach(func() {
		// reset to the default disabled state so other specs are unaffected
		err := setting_pkg.Update[*settings.LimitQuotaOverSubscription](TestTenant, settings.LimitQuotaOverSubscription{Enabled: false}).Execute()
		Expect(err).To(BeNil())
	})

	It("GetOrgUnitSettings returns limitQuotaOverSubscription with its display metadata", func() {
		err := setting_pkg.Update[*settings.LimitQuotaOverSubscription](TestTenant, settings.LimitQuotaOverSubscription{Enabled: false}).Execute()
		Expect(err).To(BeNil())

		resp, err := settingsService.GetOrgUnitSettings(ctx, "")
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body, ok := resp.Body.(*api.OrgUnitSettingsGetResponse)
		Expect(ok).To(BeTrue())
		Expect(body.LimitQuotaOverSubscription.Enabled).To(Equal(ptr.To(false)))
		Expect(body.LimitQuotaOverSubscription.DisplayName).ToNot(BeEmpty())
		Expect(body.LimitQuotaOverSubscription.Stage).ToNot(BeEmpty())
	})

	It("UpdateOrgUnitSettingByKey persists the value and GetOrgUnitSettingByKey reads it back", func() {
		updateResp, err := settingsService.UpdateOrgUnitSettingByKey(ctx,
			api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION,
			api.UpdateOrgUnitSettingByKeyRequest{Enabled: ptr.To(true)}, "")
		Expect(err).To(BeNil())
		Expect(updateResp.Code).To(Equal(http.StatusOK))
		updateBody, ok := updateResp.Body.(api.UpdateOrgUnitSettingByKeyRequest)
		Expect(ok).To(BeTrue())
		Expect(updateBody.Enabled).To(Equal(ptr.To(true)))

		getResp, err := settingsService.GetOrgUnitSettingByKey(ctx, api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION, "")
		Expect(err).To(BeNil())
		Expect(getResp.Code).To(Equal(http.StatusOK))
		setting, ok := getResp.Body.(*settings.LimitQuotaOverSubscription)
		Expect(ok).To(BeTrue())
		Expect(setting.Enabled).To(BeTrue())

		// flip it back off and confirm both the update response and the read reflect false
		updateResp, err = settingsService.UpdateOrgUnitSettingByKey(ctx,
			api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION,
			api.UpdateOrgUnitSettingByKeyRequest{Enabled: ptr.To(false)}, "")
		Expect(err).To(BeNil())
		Expect(updateResp.Code).To(Equal(http.StatusOK))
		updateBody, ok = updateResp.Body.(api.UpdateOrgUnitSettingByKeyRequest)
		Expect(ok).To(BeTrue())
		Expect(updateBody.Enabled).To(Equal(ptr.To(false)))

		getResp, err = settingsService.GetOrgUnitSettingByKey(ctx, api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION, "")
		Expect(err).To(BeNil())
		Expect(getResp.Code).To(Equal(http.StatusOK))
		setting, ok = getResp.Body.(*settings.LimitQuotaOverSubscription)
		Expect(ok).To(BeTrue())
		Expect(setting.Enabled).To(BeFalse())
	})

	It("GetOrgUnitSettingByKey returns 400 for an unknown key", func() {
		resp, err := settingsService.GetOrgUnitSettingByKey(ctx, api.OrgUnitSettingsKeyEnum("invalid-key"), "")
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
	})

	It("UpdateOrgUnitSettingByKey returns 400 when enabled is missing", func() {
		resp, err := settingsService.UpdateOrgUnitSettingByKey(ctx,
			api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION,
			api.UpdateOrgUnitSettingByKeyRequest{}, "")
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 401 when a non-management tenant targets another tenant", func() {
		resp, err := settingsService.GetOrgUnitSettings(ctx, "other-tenant")
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusUnauthorized))
	})
})

// validateResourcesCpu
// validates that the cpu and memory resources in the response match the expected resources,
// assuming adjustableWeights is turned off.
func validateResourcesCpu(resources api.Resources, expected api.ResourcesNullable, cpuResourcesTurnedToNull bool) {
	ExpectWithOffset(1, resources.NodePool.Id).To(Equal(expected.NodePool.Id), "node pool id mismatch")
	ExpectWithOffset(1, resources.NodePool.Name).To(Equal(expected.NodePool.Name), "node pool name mismatch")
	ExpectWithOffset(1, resources.Gpu.Deserved).To(Equal(ptr.Deref(expected.Gpu.Deserved, 0)), "gpu deserved mismatch")
	ExpectWithOffset(1, resources.Gpu.Limit).To(Equal(expected.Gpu.Limit), "gpu limit mismatch")
	ExpectWithOffset(1, resources.Gpu.OverQuotaWeight).To(Equal(expected.Gpu.OverQuotaWeight), "gpu over quota weight mismatch")
	if cpuResourcesTurnedToNull {
		ExpectWithOffset(1, resources.Cpu).To(BeNil(), "cpu should be nil")
		ExpectWithOffset(1, resources.Memory).To(BeNil(), "memory should be nil")
	} else {
		validateCpuResponseValue(resources.Cpu.Deserved, expected.Cpu.Deserved, "cpu deserved")
		validateCpuResponseValue(resources.Cpu.Limit, expected.Cpu.Limit, "cpu limit")
		ExpectWithOffset(1, resources.Cpu.OverQuotaWeight).To(Equal(expected.Cpu.OverQuotaWeight), "cpu over quota weight mismatch")

		validateCpuResponseValue(resources.Memory.Deserved, expected.Memory.Deserved, "memory deserved")
		validateCpuResponseValue(resources.Memory.Limit, expected.Memory.Limit, "memory limit")
		ExpectWithOffset(1, resources.Memory.OverQuotaWeight).To(Equal(expected.Memory.OverQuotaWeight), "memory over quota weight mismatch")
	}
}

func validateCpuResponseValue(actual, expected *float64, message string) {
	ExpectWithOffset(2, actual).ToNot(BeNil(), fmt.Sprintf("%s - actual is nil", message))
	if expected == nil {
		ExpectWithOffset(2, *actual).To(Equal(float64(-1)), fmt.Sprintf("%s - expected -1 for nil", message))
	} else {
		ExpectWithOffset(2, *actual).To(Equal(*expected), fmt.Sprintf("%s - value mismatch", message))
	}
}
