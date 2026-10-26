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

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
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

var _ = Describe("Org unit sync tests", Ordered, func() {
	var clusterId string
	ctx := SetContext(TestTenant, TestUser)

	var nodeType1Id string
	var nodeType1Name string
	var nodeType2Id string
	var nodeType2Name string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var nodeTypeService api.NodeTypesAPIServicer
	var notificationService api.NotificationsAPIServicer
	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.21")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
			},
		},
	}, nil)

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting org unit sync tests")

		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)
	})

	AfterAll(func() {
		_, err := nodeTypeService.DeleteNodeType(ctx, nodeType1Id)
		if err != nil {
			Fail("Failed deleting node type")
		}
		_, err = nodeTypeService.DeleteNodeType(ctx, nodeType2Id)
		if err != nil {
			Fail("Failed deleting node type")
		}
	})

	Context("test projects", func() {
		var department1Id string
		var department2Id string
		var projectIds []string

		BeforeEach(func() {
			clusterId = uuid.New().String()
			nodeTypeService = services.NewNodeTypeApiService()
			nodeType1Name = "nodeType1" + uuid.New().String()
			nodeTypeRequest := api.NodeTypeCreateRequest{Name: nodeType1Name, ClusterId: clusterId}
			nodeTypeResponse, err := nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
			nodeType1Id = nodeTypeResponse.Body.(api.NodeType).Id
			nodeType2Name = "nodeType2" + uuid.New().String()
			nodeTypeRequest = api.NodeTypeCreateRequest{Name: nodeType2Name, ClusterId: clusterId}
			nodeTypeResponse, err = nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
			nodeType2Id = nodeTypeResponse.Body.(api.NodeType).Id

			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			departmentName := "dep" + uuid.New().String()
			createDepartmentRequest := getDepartmentCreateRequest(departmentName, clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department1Id = response.Body.(api.Department).Id

			departmentName = "dep" + uuid.New().String()
			createDepartmentRequest = getDepartmentCreateRequest(departmentName, clusterId)
			createDepartmentRequest.DefaultNodePools = ptr.To([]string{"pool1"})
			createDepartmentRequest.SchedulingRules = ptr.To(api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(int32(100)),
			})
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training: []string{nodeType1Id},
			}
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department2Id = response.Body.(api.Department).Id
		})

		AfterEach(func() {
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, department1Id)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
			deleteResponse, err = departmentApiService.DeleteDepartment(ctx, department2Id)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		})

		It("basic project", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, department1Id, clusterId)
			request.SchedulingRules = ptr.To(api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
			})

			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)
			updatedAt := project.UpdatedAt

			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(updatedAt))

			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			validateSyncProject(syncProject, department1Id, false, false, nodeType1Id)

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Second), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department1Id, false, false, nodeType1Id)
			resources := syncProject.Resources[0]
			rankValue := services.MapRankToNumber(ptr.To(services.DefaultResourceRank))
			Expect(resources.Rank).To(Equal(rankValue))
			Expect(resources.Priority).To(Equal(rankValue))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, time.Now().Add(time.Duration(1)*time.Second), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			// update project
			time.Sleep(time.Duration(2) * time.Millisecond)
			updateRequest := getProjectUpdateRequest(project)
			updateRequest.Resources[0].Gpu.OverQuotaWeight = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Limit = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Deserved = ptr.To(float64(3))
			updateRequest.Resources[0].Rank = ptr.To(services.DefaultResourceRank)
			updateResponse, err := projectApiService.UpdateProject(ctx, project.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updatedAt = updateResponse.Body.(api.Project).UpdatedAt
			// update status updatedTime
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(updatedAt))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			resources = syncProject.Resources[0]
			Expect(resources.Rank).To(Equal(int32(100)))
			Expect(resources.Priority).To(Equal(int32(100)))
			validateSyncProject(syncProject, department1Id, true, false, nodeType1Id)

			newTime := time.Now().Add(time.Duration(1) * time.Second)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			// update project update time in cluster
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now().Add(time.Duration(-2)*time.Second)))

			// GET will return the project since update time is less than updatedAt
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department1Id, true, false, nodeType1Id)

			// update project update time in cluster
			updateProjStatus(ctx, projectApiService, projectId, clusterId, nil)

			// GET will return the project since update time is null
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department1Id, true, false, nodeType1Id)
		})

		It("permissions", func() {
			toMockAccessrulesDepId = ptr.To(department1Id)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, department1Id, clusterId)
			request.SchedulingRules = ptr.To(api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
			})

			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			Expect(syncProject.CreatedAt).ToNot(BeNil())
			Expect(syncProject.Parent.Id).To(Equal(department1Id))
			Expect(syncProject.Permissions.Users).ToNot(BeEmpty())

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			Expect(syncProject.CreatedAt).ToNot(BeNil())
			Expect(syncProject.Parent.Id).To(Equal(department1Id))
			Expect(syncProject.Permissions.Users).To(BeEmpty())

			toMockAccessrulesDepId = nil
		})

		It("inherited project", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, department2Id, clusterId)
			request.SchedulingRules = ptr.To(api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
			})

			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)
			updatedAt := project.UpdatedAt

			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(updatedAt))

			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			validateSyncProject(syncProject, department2Id, false, true, nodeType1Id)

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Second), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department2Id, false, true, nodeType1Id)

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, time.Now().Add(time.Duration(1)*time.Second), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			// update project
			time.Sleep(time.Duration(2) * time.Millisecond)
			updateRequest := getProjectUpdateRequest(project)
			updateRequest.Resources[0].Gpu.OverQuotaWeight = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Limit = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Deserved = ptr.To(float64(3))
			updateRequest.Resources[0].Rank = ptr.To("VeryLow")
			updateResponse, err := projectApiService.UpdateProject(ctx, project.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updatedAt = updateResponse.Body.(api.Project).UpdatedAt
			// update status updatedTime
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(updatedAt))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department2Id, true, true, nodeType1Id)

			resources := syncProject.Resources[0]
			Expect(resources.Rank).To(Equal(int32(20)))
			Expect(resources.Priority).To(Equal(int32(20)))

			newTime := time.Now().Add(time.Duration(1) * time.Second)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			// update project update in cluster
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now().Add(time.Duration(-2)*time.Second)))

			// GET will return the project since update time is less than updatedAt
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department2Id, true, true, nodeType1Id)

			// update project update time in cluster
			updateProjStatus(ctx, projectApiService, projectId, clusterId, nil)

			// GET will return the project since update time is null
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, newTime, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject = syncProjects[0]
			validateSyncProject(syncProject, department2Id, true, true, nodeType1Id)
		})

		It("multi projects with different update time", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, department1Id, clusterId)
			request.SchedulingRules = ptr.To(api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
			})

			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			project1Id := project.Id

			name2 := createProjectName()
			request = getProjectCreateRequest(name2, name2, department1Id, clusterId)

			time.Sleep(time.Duration(2) * time.Millisecond)

			createResponse, err = projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project2 := createResponse.Body.(api.Project)
			project2Id := project2.Id
			projectIds = append(projectIds, project2Id)
			updatedAt := project2.UpdatedAt

			updateProjStatus(ctx, projectApiService, project1Id, clusterId, ptr.To(updatedAt))
			updateProjStatus(ctx, projectApiService, project2Id, clusterId, ptr.To(updatedAt))

			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(2))
			Expect(syncProjects[0].Id).To(Equal(project1Id))
			Expect(syncProjects[1].Id).To(Equal(project2Id))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(project2Id))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			time.Sleep(time.Duration(2) * time.Millisecond)

			deleteProject, err := projectApiService.DeleteProject(ctx, project1Id, api.DELETIONTYPES_IF_EMPTY, true)
			Expect(err).To(BeNil())
			Expect(deleteProject.Code).To(Equal(http.StatusNoContent))

			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, updatedAt, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(project1Id))
		})

		It("GetSyncProjects reflects batch-patched resources", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, department1Id, clusterId)

			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))

			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			// Seed status so sync API uses calculated_updated_at correctly
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(project.UpdatedAt))
			time.Sleep(2 * time.Millisecond)

			// Batch patch: change GPU deserved/limit on default nodepool
			newGpu := 7.0
			patchResources := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			batchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, patchResources),
				},
			}
			batchReq.PatchRequests[0].Resources.Gpu.Deserved = &newGpu
			batchReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(newGpu)
			batchReq.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(newGpu)

			beforeUpdatedAt := time.Now()
			time.Sleep(2 * time.Millisecond) // ensure UpdatedAt will be later than beforeUpdatedAt
			syncResp, err := projectApiService.GetSyncProjects(ctx, clusterId, beforeUpdatedAt, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncProjects := syncResp.Body.(api.GetSyncProjects200Response).Projects
			// make sure that before the patch, no projects are returned
			Expect(len(syncProjects)).To(Equal(0))

			batchResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(batchResp.Code).To(Equal(http.StatusOK), fmt.Sprintf("Batch patch response body: %+v", batchResp.Body))
			batchBody := batchResp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(batchBody.TotalRequested).To(Equal(int32(1)))
			Expect(batchBody.TotalSuccessful).To(Equal(int32(1)))

			// GetSyncProjects with updatedAfter just before the patch should return the project
			syncResp, err = projectApiService.GetSyncProjects(ctx, clusterId, beforeUpdatedAt, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncProjects = syncResp.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			syncProject := syncProjects[0]
			Expect(syncProject.Id).To(Equal(projectId))

			// Validate patched GPU in sync payload
			var syncRes api.ResourcesSync
			for _, r := range syncProject.Resources {
				if r.NodePool != nil && r.NodePool.Id == fmt.Sprint(defaultNodePoolId) {
					syncRes = r
					break
				}
			}
			Expect(syncRes.NodePool).ToNot(BeNil())
			Expect(syncRes.Gpu.Deserved).To(Equal(newGpu))
			Expect(syncRes.Gpu.Limit).ToNot(BeNil())
			Expect(*syncRes.Gpu.Limit).To(Equal(newGpu))
		})
	})

	Context("test departments", func() {
		var departmentIds []string

		BeforeEach(func() {
			clusterId = uuid.New().String()
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		})

		AfterEach(func() {
			for _, departmentId := range departmentIds {
				deleteResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			departmentIds = []string{}
		})

		It("basic department", func() {
			name := createDepName()
			request := getDepartmentCreateRequest(name, clusterId)

			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			department := createResponse.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			updatedAt := department.UpdatedAt

			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(updatedAt))

			getResponse, err := departmentApiService.GetSyncDepartments(ctx, clusterId, time.Time{}, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments := getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment := syncDepartments[0]
			validateSyncDepartment(syncDepartment, false)

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Second), 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment = syncDepartments[0]
			validateSyncDepartment(syncDepartment, false)

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, time.Now().Add(time.Duration(1)*time.Second), 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			// update department
			time.Sleep(time.Duration(2) * time.Millisecond)
			updateRequest := getDepartmentUpdateRequest(department)
			updateRequest.Resources[0].Gpu.OverQuotaWeight = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Limit = ptr.To(float64(3))
			updateRequest.Resources[0].Gpu.Deserved = ptr.To(float64(3))
			updateResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, updateRequest)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updatedAt = updateResponse.Body.(api.Department).UpdatedAt
			// update status updatedTime
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(updatedAt))

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment = syncDepartments[0]
			validateSyncDepartment(syncDepartment, true)

			newTime := time.Now().Add(time.Duration(1) * time.Second)
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, newTime, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			// update department update time in cluster
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now().Add(time.Duration(-2)*time.Second)))

			// GET will return the department since update time is less than updatedAt
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, newTime, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment = syncDepartments[0]
			validateSyncDepartment(syncDepartment, true)

			// update department update time in cluster
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, nil)

			// GET will return the department since update time is null
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, newTime, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment = syncDepartments[0]
			validateSyncDepartment(syncDepartment, true)
		})

		It("multi departments with different update time", func() {
			name := createDepName()
			request := getDepartmentCreateRequest(name, clusterId)

			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			department := createResponse.Body.(api.Department)
			department1Id := department.Id

			name2 := createDepName()
			request = getDepartmentCreateRequest(name2, clusterId)

			time.Sleep(time.Duration(2) * time.Millisecond)

			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			department = createResponse.Body.(api.Department)
			department2Id := department.Id
			departmentIds = append(departmentIds, department2Id)
			updatedAt := department.UpdatedAt

			updateDepStatus(ctx, departmentApiService, department1Id, clusterId, ptr.To(updatedAt))
			updateDepStatus(ctx, departmentApiService, department2Id, clusterId, ptr.To(updatedAt))

			getResponse, err := departmentApiService.GetSyncDepartments(ctx, clusterId, time.Time{}, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments := getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(2))
			Expect(syncDepartments[0].Id).To(Equal(department1Id))
			Expect(syncDepartments[1].Id).To(Equal(department2Id))

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, updatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(department2Id))

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, updatedAt, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))

			time.Sleep(time.Duration(2) * time.Millisecond)

			deleteRes, err := departmentApiService.DeleteDepartment(ctx, department1Id)
			Expect(err).To(BeNil())
			Expect(deleteRes.Code).To(Equal(http.StatusNoContent))

			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, updatedAt, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(department1Id))
		})

		It("GetSyncDepartments reflects batch-patched resources", func() {
			name := createDepName()
			request := getDepartmentCreateRequest(name, clusterId)

			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))

			department := createResponse.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)

			// Seed status so sync API uses calculated_updated_at correctly
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(department.UpdatedAt))
			time.Sleep(2 * time.Millisecond)

			// Batch patch: change GPU deserved/limit on default nodepool
			newGpu := 5.0
			patchResources := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			batchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, patchResources),
				},
			}
			batchReq.PatchRequests[0].Resources.Gpu.Deserved = &newGpu
			batchReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(newGpu)
			batchReq.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(newGpu)

			beforeUpdatedAt := time.Now()
			time.Sleep(2 * time.Millisecond) // ensure UpdatedAt will be later than beforeUpdatedAt
			// GetSyncDepartments with updatedAfter just before the patch should return the department
			syncResp, err := departmentApiService.GetSyncDepartments(ctx, clusterId, beforeUpdatedAt.Add(-1*time.Millisecond), 0, 0)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncDepartments := syncResp.Body.(api.GetSyncDepartments200Response).Departments
			// make sure that before the patch, no departments are returned
			Expect(len(syncDepartments)).To(Equal(0))

			batchResp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(batchResp.Code).To(Equal(http.StatusOK), fmt.Sprintf("Batch patch response body: %+v", batchResp.Body))
			batchBody := batchResp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(batchBody.TotalRequested).To(Equal(int32(1)))
			Expect(batchBody.TotalSuccessful).To(Equal(int32(1)))

			// GetSyncDepartments with updatedAfter just before the patch should return the department
			syncResp, err = departmentApiService.GetSyncDepartments(ctx, clusterId, beforeUpdatedAt.Add(-1*time.Millisecond), 0, 0)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncDepartments = syncResp.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			syncDepartment := syncDepartments[0]
			Expect(syncDepartment.Id).To(Equal(departmentId))

			// Validate patched GPU in sync payload
			var syncRes api.ResourcesSync
			for _, r := range syncDepartment.Resources {
				if r.NodePool != nil && r.NodePool.Id == fmt.Sprint(defaultNodePoolId) {
					syncRes = r
					break
				}
			}
			Expect(syncRes.NodePool).ToNot(BeNil())
			Expect(syncRes.Gpu.Deserved).To(Equal(newGpu))
			Expect(syncRes.Gpu.Limit).ToNot(BeNil())
			Expect(*syncRes.Gpu.Limit).To(Equal(newGpu))
		})
	})

	Context("test calculated_updated_at", func() {
		var departmentIds []string
		var projectIds []string

		BeforeEach(func() {
			clusterId = uuid.New().String()
			notificationService = services.NewNotificationsApiService(&clusterServiceClient)
			nodeTypeService = services.NewNodeTypeApiService()
			nodeType1Name = "nodeType1" + uuid.New().String()
			nodeTypeRequest := api.NodeTypeCreateRequest{Name: nodeType1Name, ClusterId: clusterId}
			nodeTypeResponse, err := nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
			nodeType1Id = nodeTypeResponse.Body.(api.NodeType).Id
			nodeType2Name = "nodeType2" + uuid.New().String()
			nodeTypeRequest = api.NodeTypeCreateRequest{Name: nodeType2Name, ClusterId: clusterId}
			nodeTypeResponse, err = nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
			nodeType2Id = nodeTypeResponse.Body.(api.NodeType).Id
		})

		AfterEach(func() {
			for _, projectId := range projectIds {
				deleteProjResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteProjResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
			for _, departmentId := range departmentIds {
				deleteDepResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
				if err != nil || deleteDepResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			departmentIds = []string{}
		})

		It("test departments", func() {
			departmentName := "dep" + uuid.New().String()
			createDepartmentRequest := getDepartmentCreateRequest(departmentName, clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			departmentUpdatedAt := department.UpdatedAt

			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now()))

			// test updatedAt from department:
			getResponse, err := departmentApiService.GetSyncDepartments(ctx, clusterId, departmentUpdatedAt.Add(time.Duration(-1)*time.Millisecond), 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments := getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(departmentId))
			Expect(syncDepartments[0].UpdatedAt.UnixMilli()).To(Equal(departmentUpdatedAt.UnixMilli()))

			// test updatedAt from resources:
			timeBeforeExpectedUpdate := time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			nodePoolUpdated, err := notificationService.NodePoolUpdated(ctx, false, api.NodePool{
				Uid:       111,
				Id:        "111",
				Name:      "new_pool",
				ClusterId: clusterId,
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(nodePoolUpdated.Code).To(Equal(http.StatusOK))
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now()))
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(departmentId))
			Expect(syncDepartments[0].UpdatedAt.UnixMilli()).To(BeNumerically(">", departmentUpdatedAt.UnixMilli()))

			// test updatedAt with nodepool deleting from resources:
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			nodePoolUpdated, err = notificationService.NodePoolUpdated(ctx, false, api.NodePool{
				Uid:       111,
				Id:        "111",
				Name:      "new_pool",
				ClusterId: clusterId,
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(nodePoolUpdated.Code).To(Equal(http.StatusOK))
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now()))
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(departmentId))
			Expect(syncDepartments[0].UpdatedAt.UnixMilli()).To(BeNumerically(">", departmentUpdatedAt.UnixMilli()))

			// test updatedAt after node type change:
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			department.NodeTypes.Training = []string{nodeType1Id}
			updateResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, getUpdateRequestFromDepartment(department))
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updateDepStatus(ctx, departmentApiService, departmentId, clusterId, ptr.To(time.Now()))
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(1))
			Expect(syncDepartments[0].Id).To(Equal(departmentId))

			// test no change - should return 0 departments
			time.Sleep(time.Duration(2) * time.Millisecond)
			timeBeforeNoExpectedUpdate := time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			getResponse, err = departmentApiService.GetSyncDepartments(ctx, clusterId, timeBeforeNoExpectedUpdate, 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncDepartments = getResponse.Body.(api.GetSyncDepartments200Response).Departments
			Expect(len(syncDepartments)).To(Equal(0))
		})

		It("test projects", func() {
			departmentName := "dep" + uuid.New().String()
			createDepartmentRequest := getDepartmentCreateRequest(departmentName, clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)
			projectUpdatedAt := project.UpdatedAt

			// update status - else all the project will always return... (WHERE ... status_last_updated_time is null OR ....)
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))

			// test updatedAt from project:
			testUpdatedAfter := projectUpdatedAt.Add(time.Duration(-1) * time.Millisecond)
			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, testUpdatedAfter, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))
			Expect(syncProjects[0].UpdatedAt.UnixMilli()).To(Equal(projectUpdatedAt.UnixMilli()))

			// test no change in updated at - should return 0 projects
			time.Sleep(time.Duration(2) * time.Millisecond)
			timeBeforeNoExpectedUpdate := time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeNoExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))

			// test updatedAt from resources:
			timeBeforeExpectedUpdate := time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			nodePoolUpdated, err := notificationService.NodePoolUpdated(ctx, false, api.NodePool{
				Uid:       111,
				Id:        "111",
				Name:      "new_pool",
				ClusterId: clusterId,
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(nodePoolUpdated.Code).To(Equal(http.StatusOK))
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))
			Expect(syncProjects[0].UpdatedAt.UnixMilli()).To(BeNumerically(">", projectUpdatedAt.UnixMilli()))

			// test updatedAt from resources:
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			nodePoolUpdated, err = notificationService.NodePoolUpdated(ctx, false, api.NodePool{
				Uid:       111,
				Id:        "111",
				Name:      "new_pool",
				ClusterId: clusterId,
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(nodePoolUpdated.Code).To(Equal(http.StatusOK))
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))
			Expect(syncProjects[0].UpdatedAt.UnixMilli()).To(BeNumerically(">", projectUpdatedAt.UnixMilli()))

			// test updatedAt after department effective change:
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			department.SchedulingRules = &api.SchedulingRules{
				InteractiveJobTimeLimitSeconds:           nil,
				InteractiveJobMaxIdleDurationSeconds:     ptr.To(int32(20)),
				InteractiveJobPreemptIdleDurationSeconds: nil,
				TrainingJobMaxIdleDurationSeconds:        nil,
				TrainingJobTimeLimitSeconds:              nil,
			}
			updateDepResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, getUpdateRequestFromDepartment(department))
			Expect(err).To(BeNil())
			Expect(updateDepResponse.Code).To(Equal(http.StatusOK))
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))
			Expect(syncProjects[0].UpdatedAt.UnixMilli()).To(BeNumerically(">", projectUpdatedAt.UnixMilli()))

			// test updatedAt after node type change:
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			project.NodeTypes.Workspace = []string{nodeType1Id}
			updateResponse, err := projectApiService.UpdateProject(ctx, projectId, getUpdateRequestFromProject(project))
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))

			// test updatedAt after node type removal:
			// there is a bug in node type deletion - no "deleted_at" in org_unit_train_node_type tables...
			timeBeforeExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			_, err = nodeTypeService.DeleteNodeType(ctx, nodeType1Id)
			if err != nil {
				Fail("Failed deleting node type")
			}
			updateProjStatus(ctx, projectApiService, projectId, clusterId, ptr.To(time.Now()))
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1))
			Expect(syncProjects[0].Id).To(Equal(projectId))

			// test not change in updated at
			timeBeforeNoExpectedUpdate = time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBeforeNoExpectedUpdate, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0))
		})
	})
})

func getUpdateRequestFromDepartment(department api.Department) api.DepartmentUpdateRequest {
	resources := []api.ResourcesNullable{}
	for i := range department.Resources {
		resources = append(resources, api.ResourcesNullable{
			NodePool: department.Resources[i].NodePool,
			Gpu: api.ResourcesNullableGpu{
				Deserved:        ptr.To(department.Resources[i].Gpu.Deserved),
				Limit:           department.Resources[i].Gpu.Limit,
				OverQuotaWeight: department.Resources[i].Gpu.OverQuotaWeight,
			},
			Cpu:    department.Resources[i].Cpu,
			Memory: department.Resources[i].Memory,
			Rank:   department.Resources[i].Rank,
		})
	}
	return api.DepartmentUpdateRequest{
		Description:      department.Description,
		Resources:        resources,
		SchedulingRules:  department.SchedulingRules,
		DefaultNodePools: department.DefaultNodePools,
		NodeTypes:        department.NodeTypes,
	}
}

func getUpdateRequestFromProject(project api.Project) api.ProjectUpdateRequest {
	resources := []api.ResourcesNullable{}
	for i := range project.Resources {
		resources = append(resources, api.ResourcesNullable{
			NodePool: project.Resources[i].NodePool,
			Gpu: api.ResourcesNullableGpu{
				Deserved:        ptr.To(project.Resources[i].Gpu.Deserved),
				Limit:           project.Resources[i].Gpu.Limit,
				OverQuotaWeight: project.Resources[i].Gpu.OverQuotaWeight,
			},
			Cpu:    project.Resources[i].Cpu,
			Memory: project.Resources[i].Memory,
			Rank:   project.Resources[i].Rank,
		})
	}
	return api.ProjectUpdateRequest{
		Description:      project.Description,
		Resources:        resources,
		SchedulingRules:  project.SchedulingRules,
		DefaultNodePools: project.DefaultNodePools,
		NodeTypes:        project.NodeTypes,
	}
}

func validateSyncProject(syncProject api.ProjectSync, departmentId string, afterUpdate bool, withEffectiveData bool, nodeType1Id string) {
	Expect(syncProject.CreatedAt).ToNot(BeNil())
	Expect(syncProject.Parent.Id).To(Equal(departmentId))
	Expect(len(syncProject.Resources)).To(Equal(1))
	if afterUpdate {
		ExpectWithOffset(1, syncProject.UpdatedAt).NotTo(Equal(syncProject.CreatedAt))
	} else {
		ExpectWithOffset(1, syncProject.UpdatedAt).To(Equal(syncProject.CreatedAt))
	}
	if withEffectiveData {
		Expect(*syncProject.DefaultNodePools).To(Equal([]string{"pool1"}))
		Expect(syncProject.NodeTypes.Training).To(Equal([]string{nodeType1Id}))
		Expect(syncProject.NodeTypes.Workspace).To(BeNil())
		Expect(*syncProject.SchedulingRules).To(Equal(api.SchedulingRules{
			InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
			TrainingJobTimeLimitSeconds:    ptr.To(int32(100)),
		}))
	} else {
		Expect(*syncProject.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
		Expect(syncProject.NodeTypes.Training).To(BeNil())
		Expect(syncProject.NodeTypes.Workspace).To(BeNil())
		Expect(*syncProject.SchedulingRules).To(Equal(api.SchedulingRules{
			InteractiveJobTimeLimitSeconds: ptr.To(int32(20)),
		}))
	}
}

func validateSyncDepartment(syncDepartment api.DepartmentSync, afterUpdate bool) {
	Expect(syncDepartment.CreatedAt).ToNot(BeNil())
	Expect(len(syncDepartment.Resources)).To(Equal(1))
	if afterUpdate {
		Expect(syncDepartment.UpdatedAt).NotTo(Equal(syncDepartment.CreatedAt))
	} else {
		Expect(syncDepartment.UpdatedAt).To(Equal(syncDepartment.CreatedAt))
	}
}

func updateProjStatus(ctx context.Context, projectApiService api.ProjectsAPIServicer, projectId, clusterId string, lastUpdatedTime *time.Time) {
	updateStatusRequest := api.ProjectStatusUpdateRequest{
		Id:        projectId,
		ClusterId: clusterId,
		Status: api.ProjectStatus{
			LastUpdatedTime: lastUpdatedTime,
			Phase:           api.PROJECTPHASE_READY,
		},
	}
	updateStatusRes, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
	Expect(err).To(BeNil())
	Expect(updateStatusRes.Code).To(Equal(http.StatusOK))
}

func updateDepStatus(ctx context.Context, departmentApiService api.DepartmentsAPIServicer, departmentId, clusterId string, lastUpdatedTime *time.Time) {
	updateStatusRequest := api.DepartmentStatusUpdateRequest{
		ClusterId: clusterId,
		Status: api.DepartmentStatus{
			LastUpdatedTime: lastUpdatedTime,
		},
	}
	updateStatusRes, err := departmentApiService.UpdateDepartmentStatus(ctx, departmentId, updateStatusRequest)
	Expect(err).To(BeNil())
	Expect(updateStatusRes.Code).To(Equal(http.StatusNoContent))
}
