/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	newNodepoolName  = "new-pool"
	newNodepoolId    = "11"
	newNodepool1Name = "new-pool1"
	newNodepool1Id   = "111"
)

var _ = Describe("Node pool resources API tests", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	nodePool2Name := "pool2"
	nodePool2Id := float32(2)
	mockNodepools := []openapi.NodepoolMinimal{
		{
			Name:                   defaultNodePoolName,
			Id:                     fmt.Sprint(defaultNodePoolId),
			SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
		},
		{
			Name:                   nodePool2Name,
			Id:                     fmt.Sprint(nodePool2Id),
			SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
		},
		{
			Name:                   newNodepoolName,
			Id:                     newNodepoolId,
			SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
		},
		{
			Name:                   newNodepool1Name,
			Id:                     newNodepool1Id,
			SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
		},
	}

	var notificationsService api.NotificationsAPIServicer
	var departmentService api.DepartmentsAPIServicer
	var projectService api.ProjectsAPIServicer
	clusterServiceClient := mocks.IClusterClient{}

	clusterId := uuid.New().String()
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *openapi.NewNullableString(ptr.To("2.17")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
		Nodepools: mockNodepools,
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	var departmentIds []string
	var projectIds []string

	BeforeEach(func() {
		notificationsService = services.NewNotificationsApiService(&clusterServiceClient)
		departmentService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
	})

	AfterEach(func() {
		for _, projectId := range projectIds {
			deleteResponse, err := projectService.DeleteProject(ctx, projectId, "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		projectIds = []string{}
		for _, departmentId := range departmentIds {
			deleteResponse, err := departmentService.DeleteDepartment(ctx, departmentId)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		departmentIds = []string{}
	})

	Context("test non default node pool", func() {
		BeforeEach(func() {
			createDepartmentRequest := getFullDepartmentCreateRequest("dep1", clusterId)
			createDepartmentRequest.DefaultNodePools = ptr.To([]string{nodePool2Name})
			createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, CreateNodePoolApiResources(false, "2", nodePool2Name))
			response, err := departmentService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			department1Id := department.Id
			departmentIds = append(departmentIds, department1Id)

			request := getProjectCreateRequest(createProjectName(), createProjectName(), department1Id, clusterId)
			request.Resources = append(request.Resources, CreateNodePoolApiResources(false, "2", nodePool2Name))
			request.DefaultNodePools = ptr.To([]string{"some-other-np-1", nodePool2Name, "some-other-np-2"})
			createResponse, err := projectService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1 := createResponse.Body.(api.Project)
			projectIds = append(projectIds, project1.Id)

			request = getProjectCreateRequest(createProjectName(), createProjectName(), department1Id, clusterId)
			request.Resources = append(request.Resources, CreateNodePoolApiResources(false, "2", nodePool2Name))
			request.DefaultNodePools = ptr.To([]string{"some-other-np-1", "some-other-np-2"})
			createResponse, err = projectService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project2 := createResponse.Body.(api.Project)
			projectIds = append(projectIds, project2.Id)

			time.Sleep(time.Duration(2) * time.Millisecond)
		})

		It("test create node pool", func() {
			updated, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        newNodepoolId,
				Name:      newNodepoolName,
				ClusterId: clusterId,
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))
			projectService, departmentService = setMockGetNodepools(mockNodepools[:3])

			getDepartmentResponse, err := departmentService.GetDepartment(ctx, departmentIds[0])
			if err != nil {
				Fail("Failed getting department")
			}
			department := getDepartmentResponse.Body.(api.Department)
			validateCreatedNodePoolResources(department.Resources, 0)

			getProjectResponse, err := projectService.GetProject(ctx, projectIds[0])
			if err != nil {
				Fail("Failed getting projects")
			}
			project := getProjectResponse.Body.(api.Project)
			validateCreatedNodePoolResources(project.Resources, 2)
			Expect(*project.TotalResources.GpuQuota).To(Equal(float64(4)))

			getProjectResponse, err = projectService.GetProject(ctx, projectIds[1])
			if err != nil {
				Fail("Failed getting projects")
			}
			project = getProjectResponse.Body.(api.Project)
			validateCreatedNodePoolResources(project.Resources, 2)
			Expect(*project.TotalResources.GpuQuota).To(Equal(float64(4)))
		})

		It("test delete node pool", func() {
			updated, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        "2",
				Name:      nodePool2Name,
				ClusterId: clusterId,
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))

			getDepartmentResponse, err := departmentService.GetDepartment(ctx, departmentIds[0])
			if err != nil {
				Fail("Failed getting department")
			}
			department := getDepartmentResponse.Body.(api.Department)
			validateDeleteNodePoolResources(department.Resources)
			Expect(department.DefaultNodePools).To(BeNil())

			getProjectResponse, err := projectService.GetProject(ctx, projectIds[0])
			if err != nil {
				Fail("Failed getting projects")
			}
			project := getProjectResponse.Body.(api.Project)
			validateDeleteNodePoolResources(project.Resources)
			Expect(*project.TotalResources.GpuQuota).To(Equal(float64(2)))
			Expect(project.DefaultNodePools).To(Not(BeNil()))
			Expect(len(*project.DefaultNodePools)).To(Equal(2))
			Expect(*project.DefaultNodePools).To(Equal([]string{"some-other-np-1", "some-other-np-2"}))

			getProjectResponse, err = projectService.GetProject(ctx, projectIds[1])
			if err != nil {
				Fail("Failed getting projects")
			}
			project = getProjectResponse.Body.(api.Project)
			validateDeleteNodePoolResources(project.Resources)
			Expect(*project.TotalResources.GpuQuota).To(Equal(float64(2)))
			Expect(project.DefaultNodePools).To(Not(BeNil()))
			Expect(len(*project.DefaultNodePools)).To(Equal(2))
			Expect(*project.DefaultNodePools).To(Equal([]string{"some-other-np-1", "some-other-np-2"}))
		})
	})

	Context("test default node pool", func() {
		It("test default node pool with unlimited", func() {
			newClusterId := uuid.New()
			updated, err := notificationsService.NodePoolUpdated(ctx, true, api.NodePool{
				Id:        "1",
				Name:      "default",
				ClusterId: newClusterId.String(),
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))

			getDepartmentResponse, err := departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", newClusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments := getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal("default"))
			Expect(departments[0].ClusterId).To(Equal(newClusterId.String()))

			updated, err = notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        newNodepool1Id,
				Name:      newNodepool1Name,
				ClusterId: newClusterId.String(),
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))
			projectService, departmentService = setMockGetNodepools(append(mockNodepools[:2], mockNodepools[3]))

			getDepartmentResponse, err = departmentService.GetDepartment(ctx, departments[0].Id)
			if err != nil {
				Fail("Failed getting department")
			}
			department := getDepartmentResponse.Body.(api.Department)
			validateDefaultDepartmentResources(department.Resources, true)

			updated, err = notificationsService.NodePoolUpdated(ctx, true, api.NodePool{
				Id:        "1",
				Name:      "default",
				ClusterId: newClusterId.String(),
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusBadRequest))
		})
		It("test default node pool without unlimited", func() {
			newClusterId := uuid.New()
			updated, err := notificationsService.NodePoolUpdated(ctx, true, api.NodePool{
				Id:        "1",
				Name:      "default",
				ClusterId: newClusterId.String(),
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))

			getDepartmentResponse, err := departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", newClusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments := getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal("default"))
			Expect(departments[0].ClusterId).To(Equal(newClusterId.String()))

			//getDepartmentResponse, err = departmentService.GetDepartment(ctx, departments[0].Id)
			//if err != nil {
			//	Fail("Failed getting department")
			//}
			//department := getDepartmentResponse.Body.(api.Department)
			updateRes, err := departmentService.UpdateDepartment(ctx, departments[0].Id, api.DepartmentUpdateRequest{
				Resources: []api.ResourcesNullable{{
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(10)),
					},
				}},
			})
			Expect(err).To(BeNil())
			Expect(updateRes.Code).To(Equal(http.StatusOK))

			updated, err = notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        newNodepool1Id,
				Name:      newNodepool1Name,
				ClusterId: newClusterId.String(),
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))
			projectService, departmentService = setMockGetNodepools(append(mockNodepools[:2], mockNodepools[3]))

			getDepartmentResponse, err = departmentService.GetDepartment(ctx, departments[0].Id)
			if err != nil {
				Fail("Failed getting department")
			}
			department := getDepartmentResponse.Body.(api.Department)
			validateDefaultDepartmentResources(department.Resources, false)

			updated, err = notificationsService.NodePoolUpdated(ctx, true, api.NodePool{
				Id:        "1",
				Name:      "default",
				ClusterId: newClusterId.String(),
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("test cluster", func() {
		deletedCusterId := uuid.New().String()
		savedCusterId := uuid.New().String()
		var departmentDeletedId string
		var departmentSavedId string
		var projectSavedIds []string

		BeforeEach(func() {
			notificationsService = services.NewNotificationsApiService(&clusterServiceClient)
			departmentService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			departmentDeletedId, departmentSavedId = addDepartments(ctx, deletedCusterId, savedCusterId, departmentService)
			departmentIds = append(departmentIds, departmentSavedId)

			_, projectSavedIds = createProjectsForDelete(ctx, departmentDeletedId, departmentSavedId, deletedCusterId, savedCusterId, projectService)
			projectIds = append(projectIds, projectSavedIds...)
		})

		It("test cluster deleted", func() {
			getDepartmentResponse, err := departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", deletedCusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments := getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal("dep1"))
			Expect(departments[0].ClusterId).To(Equal(deletedCusterId))
			Expect(len(departments[0].Resources)).To(Equal(1))
			Expect(departments[0].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*departments[0].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*departments[0].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*departments[0].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))

			getDepartmentResponse, err = departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", savedCusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments = getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal("dep2"))
			Expect(departments[0].ClusterId).To(Equal(savedCusterId))
			Expect(len(departments[0].Resources)).To(Equal(1))

			getProjResponse, err := projectService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", deletedCusterId)}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting department")
			}
			projects := getProjResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].ClusterId).To(Equal(deletedCusterId))
			Expect(projects[1].ClusterId).To(Equal(deletedCusterId))
			Expect(len(projects[0].Resources)).To(Equal(1))
			Expect(len(projects[1].Resources)).To(Equal(1))
			Expect(projects[0].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[0].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))
			Expect(projects[1].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[1].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))

			getProjResponse, err = projectService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", savedCusterId)}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting department")
			}
			projects = getProjResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].ClusterId).To(Equal(savedCusterId))
			Expect(projects[1].ClusterId).To(Equal(savedCusterId))
			Expect(len(projects[0].Resources)).To(Equal(1))
			Expect(len(projects[1].Resources)).To(Equal(1))
			Expect(projects[0].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[0].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))
			Expect(projects[1].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[1].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))

			updated, err := notificationsService.ClusterDeleted(ctx, deletedCusterId)
			Expect(err).To(BeNil())
			Expect(updated.Code).To(Equal(http.StatusOK))

			getDepartmentResponse, err = departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", deletedCusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments = getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(0))
			getProjResponse, err = projectService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", deletedCusterId)}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting department")
			}
			projects = getProjResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(0))

			getDepartmentResponse, err = departmentService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", savedCusterId)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			departments = getDepartmentResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal("dep2"))
			Expect(departments[0].ClusterId).To(Equal(savedCusterId))
			Expect(len(departments[0].Resources)).To(Equal(1))
			Expect(departments[0].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*departments[0].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*departments[0].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*departments[0].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))

			getProjResponse, err = projectService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", savedCusterId)}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting department")
			}
			projects = getProjResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].ClusterId).To(Equal(savedCusterId))
			Expect(projects[1].ClusterId).To(Equal(savedCusterId))
			Expect(len(projects[0].Resources)).To(Equal(1))
			Expect(len(projects[1].Resources)).To(Equal(1))
			Expect(projects[0].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[0].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[0].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))
			Expect(projects[1].OvertimeData.Range24hData).NotTo(BeNil())
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuAllocation).To(Equal(float64(50)))
			Expect((*projects[1].OvertimeData.Range24hData).AverageGpuUtilization).To(Equal(float64(60)))
			Expect(*(*projects[1].OvertimeData.Range24hData).AverageGpuMemoryUtilization).To(Equal(float64(70)))
		})
	})
})

func createProjectsForDelete(ctx context.Context, deletedDepartment string, savedDepartment string, deletedClusterId string, savedClusterId string, projectService api.ProjectsAPIServicer) ([]string, []string) {
	projectDeletedIds := []string{}
	projectSavedIds := []string{}
	projectName := createProjectName()
	request := getProjectCreateRequest(projectName, projectName, deletedDepartment, deletedClusterId)
	createResponse, err := projectService.CreateProject(ctx, request)
	if err != nil {
		Fail("Failed creating project")
	}
	Expect(createResponse.Code).To(Equal(http.StatusCreated))
	result := map[string]map[string]*models.OvertimeData{
		deletedClusterId: {
			projectName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	projectDeletedIds = append(projectDeletedIds, createResponse.Body.(api.Project).Id)
	projectName = createProjectName()
	request = getProjectCreateRequest(projectName, projectName, deletedDepartment, deletedClusterId)
	createResponse, err = projectService.CreateProject(ctx, request)
	if err != nil {
		Fail("Failed creating project")
	}
	Expect(createResponse.Code).To(Equal(http.StatusCreated))
	result = map[string]map[string]*models.OvertimeData{
		deletedClusterId: {
			projectName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	projectDeletedIds = append(projectDeletedIds, createResponse.Body.(api.Project).Id)

	projectName = createProjectName()
	request = getProjectCreateRequest(projectName, projectName, savedDepartment, savedClusterId)
	createResponse, err = projectService.CreateProject(ctx, request)
	if err != nil {
		Fail("Failed creating project")
	}
	Expect(createResponse.Code).To(Equal(http.StatusCreated))
	result = map[string]map[string]*models.OvertimeData{
		savedClusterId: {
			projectName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	projectSavedIds = append(projectSavedIds, createResponse.Body.(api.Project).Id)

	projectName = createProjectName()
	request = getProjectCreateRequest(projectName, projectName, savedDepartment, savedClusterId)
	createResponse, err = projectService.CreateProject(ctx, request)
	if err != nil {
		Fail("Failed creating project")
	}
	Expect(createResponse.Code).To(Equal(http.StatusCreated))
	result = map[string]map[string]*models.OvertimeData{
		savedClusterId: {
			projectName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	projectSavedIds = append(projectSavedIds, createResponse.Body.(api.Project).Id)

	return projectDeletedIds, projectSavedIds
}

func addDepartments(ctx context.Context, deletedCusterId string, savedCusterId string, departmentService api.DepartmentsAPIServicer) (string, string) {
	departmentName := "dep1"
	createDepartmentRequest := getFullDepartmentCreateRequest(departmentName, deletedCusterId)
	response, err := departmentService.CreateDepartment(ctx, createDepartmentRequest)
	if err != nil {
		Fail("Failed creating parent department")
	}
	Expect(response.Code).To(Equal(http.StatusCreated))

	result := map[string]map[string]*models.OvertimeData{
		deletedCusterId: {
			departmentName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	deletedId := response.Body.(api.Department).Id

	departmentName = "dep2"
	createDepartmentRequest = getFullDepartmentCreateRequest(departmentName, savedCusterId)
	response, err = departmentService.CreateDepartment(ctx, createDepartmentRequest)
	if err != nil {
		Fail("Failed creating parent department")
	}
	Expect(response.Code).To(Equal(http.StatusCreated))
	result = map[string]map[string]*models.OvertimeData{
		savedCusterId: {
			departmentName: &models.OvertimeData{
				AvgGpuAllocation24h:        ptr.To(float64(50)),
				AvgGpuUtilization24h:       ptr.To(float64(60)),
				AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
			},
		},
	}
	err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
	Expect(err).To(BeNil())
	savedId := response.Body.(api.Department).Id

	return deletedId, savedId
}

func validateCreatedNodePoolResources(resources []api.Resources, oqWeight float64) {
	Expect(len(resources)).To(Equal(3))

	foundNodePoolResources := map[string]bool{}
	for _, resource := range resources {
		if _, ok := foundNodePoolResources[resource.NodePool.Name]; ok {
			Fail(fmt.Sprintf("Duplicate node pool - %s", resource.NodePool.Name))
		}
		foundNodePoolResources[resource.NodePool.Name] = true

		if resource.NodePool.Name == "default" {
			Expect(resource.NodePool.Id).To(Equal("1"))
		} else if resource.NodePool.Name == "pool2" {
			Expect(resource.NodePool.Id).To(Equal("2"))
		} else if resource.NodePool.Name == newNodepoolName {
			Expect(resource.NodePool.Id).To(Equal(newNodepoolId))
			Expect(resource.Gpu.Deserved).To(Equal(float64(0)))
			Expect(*resource.Gpu.Limit).To(Equal(float64(-1)))
			Expect(*resource.Gpu.OverQuotaWeight).To(Equal(oqWeight))
			Expect(resource.Cpu).To(BeNil())
			Expect(resource.Memory).To(BeNil())
		} else {
			Fail(fmt.Sprintf("Unexpected node pool - %s", resource.NodePool.Name))
		}
	}

	Expect(foundNodePoolResources).To(HaveLen(3))
}

func validateDefaultDepartmentResources(resources []api.Resources, unlimitedGpu bool) {
	Expect(len(resources)).To(Equal(2))
	Expect(resources[0].NodePool.Name).ToNot(Equal(resources[1].NodePool.Name))

	for _, resource := range resources {
		if resource.NodePool.Name == "default" {
			Expect(resource.NodePool.Id).To(Equal("1"))
		} else if resource.NodePool.Name == newNodepool1Name {
			Expect(resource.NodePool.Id).To(Equal(newNodepool1Id))
			if unlimitedGpu {
				Expect(resource.Gpu.Deserved).To(Equal(float64(-1)))
				Expect(*resource.Gpu.Limit).To(Equal(float64(-1)))
				Expect(*resource.Gpu.OverQuotaWeight).To(Equal(float64(-1)))
			} else {
				Expect(resource.Gpu.Deserved).To(Equal(float64(0)))
				Expect(*resource.Gpu.Limit).To(Equal(float64(-1)))
				Expect(*resource.Gpu.OverQuotaWeight).To(Equal(float64(0)))
			}
			Expect(resource.Cpu).To(BeNil())
			Expect(resource.Memory).To(BeNil())
		} else {
			Fail(fmt.Sprintf("Unexpected node pool - %s", resource.NodePool.Name))
		}
	}
}

func validateDeleteNodePoolResources(resources []api.Resources) {
	Expect(len(resources)).To(Equal(1))
	Expect(resources[0].NodePool.Name).To(Equal("default"))
	Expect(resources[0].NodePool.Id).To(Equal("1"))
}

func setMockGetNodepools(mockNodepools []openapi.NodepoolMinimal) (api.ProjectsAPIServicer, api.DepartmentsAPIServicer) {
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
		Uuid:    uuid.New().String(),
		Version: *openapi.NewNullableString(ptr.To("2.17")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
		Nodepools: mockNodepools,
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	projectService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
	departmentService := departments.NewDepartmentApiService(&clusterServiceClient, nil)

	return projectService, departmentService
}
