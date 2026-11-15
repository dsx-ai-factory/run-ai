/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"
	"strconv"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Parity API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var projectIds []string
	var departmentIds []string
	var nodeType1Id string
	var nodeType1Name string
	var nodeType2Id string
	var nodeType2Name string
	var projectV1ApiService api.ProjectsV1APIServicer
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var nodeTypeService api.NodeTypesAPIServicer
	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.19")),
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
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	BeforeAll(func() {
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

	BeforeEach(func() {
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, projectDeletionService)
	})

	AfterEach(func() {
		for _, projectId := range projectIds {
			deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		projectIds = []string{}
		for _, departmentId := range departmentIds {
			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		departmentIds = []string{}
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

	Context("test new department fields", func() {
		It("test POST/PUT", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			updateDepartmentRequest := getEffectiveDepartmentUpdateRequest(department)
			updateDepartmentRequest.DefaultNodePools = &[]string{"default"}
			updateDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Workspace: []string{nodeType1Id},
				Training:  nil,
			}
			timeout2 := int32(100)
			updateDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: ptr.To(timeout2),
			}

			updateResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, updateDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			department = updateResponse.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).ToNot(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(1))
			Expect(*department.DefaultNodePools).To(Equal([]string{"default"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Workspace)).To(Equal(1))
			Expect(department.NodeTypes.Workspace[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getDep, err = departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).NotTo(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(1))
			Expect(*department.DefaultNodePools).To(Equal([]string{"default"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Workspace)).To(Equal(1))
			Expect(department.NodeTypes.Workspace[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
		})
	})

	Context("test override fields", func() {
		It("test no override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).To(BeNil())
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).To(BeNil())
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).NotTo(BeNil())
			Expect(*project.Effective.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).To(BeNil())
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).To(BeNil())
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).NotTo(BeNil())
			Expect(*project.Effective.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getProjectsResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("name==%s", name)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getProjectsResponse.Code).To(Equal(http.StatusOK))
			project2 := getProjectsResponse.Body.(api.GetProjects200Response).Projects[0]
			Expect(project2.CreatedAt).ToNot(BeNil())
			Expect(project2.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project2.ParentId).To(Equal(departmentId))
			Expect(project2.DefaultNodePools).To(BeNil())
			Expect(project2.NodeTypes.Training).To(BeNil())
			Expect(project2.NodeTypes.Workspace).To(BeNil())
			Expect(project2.SchedulingRules).To(BeNil())
			Expect(len(project2.Resources)).To(Equal(1))
			Expect(project2.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project2.Effective.DefaultNodePools).NotTo(BeNil())
			Expect(*project2.Effective.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project2.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == projectId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(projectV1.CreatedAt).ToNot(BeNil())
			Expect(strconv.FormatInt(int64(projectV1.DepartmentId), 10)).To(Equal(departmentId))
			Expect(len(projectV1.DefaultNodePools)).To(Equal(1))
			Expect(projectV1.DefaultNodePools[0]).To(Equal("default"))
			Expect(projectV1.NodeAffinity.Train.AffinityType).To(Equal("no_limit"))
			Expect(projectV1.NodeAffinity.Interactive.AffinityType).To(Equal("no_limit"))
			Expect(projectV1.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(len(projectV1.NodePoolsResources)).To(Equal(1))
		})

		It("test project override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).To(BeNil())
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			request.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			request.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout),
			}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).NotTo(BeNil())
			Expect(len(*project.DefaultNodePools)).To(Equal(3))
			Expect(*project.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(project.NodeTypes).NotTo(BeNil())
			Expect(len(project.NodeTypes.Training)).To(Equal(1))
			Expect(project.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).NotTo(BeNil())
			Expect(*project.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(project.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).To(BeNil())
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).NotTo(BeNil())
			Expect(len(*project.DefaultNodePools)).To(Equal(3))
			Expect(*project.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(project.NodeTypes).NotTo(BeNil())
			Expect(len(project.NodeTypes.Training)).To(Equal(1))
			Expect(project.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).NotTo(BeNil())
			Expect(*project.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(project.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getProjectsResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("name==%s", name)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getProjectsResponse.Code).To(Equal(http.StatusOK))
			project2 := getProjectsResponse.Body.(api.GetProjects200Response).Projects[0]
			Expect(project2.CreatedAt).ToNot(BeNil())
			Expect(project2.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project2.ParentId).To(Equal(departmentId))
			Expect(project2.DefaultNodePools).NotTo(BeNil())
			Expect(len(*project2.DefaultNodePools)).To(Equal(3))
			Expect(*project2.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(project2.NodeTypes).NotTo(BeNil())
			Expect(len(project2.NodeTypes.Training)).To(Equal(1))
			Expect(project2.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(project2.NodeTypes.Workspace).To(BeNil())
			Expect(project2.SchedulingRules).NotTo(BeNil())
			Expect(*project2.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(project2.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project2.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project2.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project2.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(len(project2.Resources)).To(Equal(1))
			Expect(project2.Effective.SchedulingRules).To(Equal(project.SchedulingRules))
			Expect(project2.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project2.Effective.NodeTypes).To(Equal(project.NodeTypes))

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == projectId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(projectV1.CreatedAt).ToNot(BeNil())
			Expect(strconv.FormatInt(int64(projectV1.DepartmentId), 10)).To(Equal(departmentId))
			Expect(projectV1.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(projectV1.NodeAffinity.Train.AffinityType).To(Equal("only_selected"))
			Expect(len(projectV1.NodeAffinity.Train.SelectedTypes)).To(Equal(1))
			Expect(projectV1.NodeAffinity.Train.SelectedTypes[0].Name).To(Equal(nodeType1Name))
			Expect(projectV1.NodeAffinity.Interactive.AffinityType).To(Equal("no_limit"))
			Expect(projectV1.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobTimeLimitSecs).To(Equal(float32(10)))
			Expect(len(projectV1.NodePoolsResources)).To(Equal(1))
		})

		It("test department override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).To(BeNil())
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(department.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(department.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project.Effective.NodeTypes.Workspace).To(Equal(department.NodeTypes.Workspace))

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(len(department.Resources)).To(Equal(1))

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(project.DefaultNodePools).To(BeNil())
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(project.SchedulingRules).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(department.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(department.DefaultNodePools))
			Expect(project.Effective.NodeTypes).To(Equal(department.NodeTypes))

			getProjectsResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("name==%s", name)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getProjectsResponse.Code).To(Equal(http.StatusOK))
			project2 := getProjectsResponse.Body.(api.GetProjects200Response).Projects[0]
			Expect(project2.CreatedAt).ToNot(BeNil())
			Expect(project2.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project2.ParentId).To(Equal(departmentId))
			Expect(project2.DefaultNodePools).To(BeNil())
			Expect(project2.NodeTypes.Training).To(BeNil())
			Expect(project2.NodeTypes.Workspace).To(BeNil())
			Expect(project2.SchedulingRules).To(BeNil())
			Expect(len(project2.Resources)).To(Equal(1))
			Expect(project2.Effective.SchedulingRules).To(Equal(department.SchedulingRules))
			Expect(project2.Effective.DefaultNodePools).To(Equal(department.DefaultNodePools))
			Expect(project2.Effective.NodeTypes).To(Equal(department.NodeTypes))

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == projectId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(projectV1.CreatedAt).ToNot(BeNil())
			Expect(strconv.FormatInt(int64(projectV1.DepartmentId), 10)).To(Equal(departmentId))
			Expect(projectV1.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(projectV1.NodeAffinity.Train.AffinityType).To(Equal("only_selected"))
			Expect(len(projectV1.NodeAffinity.Train.SelectedTypes)).To(Equal(1))
			Expect(projectV1.NodeAffinity.Train.SelectedTypes[0].Name).To(Equal(nodeType1Name))
			Expect(projectV1.NodeAffinity.Interactive.AffinityType).To(Equal("no_limit"))
			Expect(projectV1.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobTimeLimitSecs).To(Equal(float32(10)))
			Expect(len(projectV1.NodePoolsResources)).To(Equal(1))
		})

		It("test department override fields with update", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.SchedulingRules).To(BeNil())
			Expect(project.Effective.SchedulingRules).To(Equal(department.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(department.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project.Effective.NodeTypes.Workspace).To(Equal(department.NodeTypes.Workspace))

			timeout2 := int32(100)
			createDepartmentRequest.SchedulingRules.TrainingJobTimeLimitSeconds = ptr.To(timeout2)
			updateResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, api.DepartmentUpdateRequest{
				Description:      createDepartmentRequest.Description,
				Resources:        createDepartmentRequest.Resources,
				SchedulingRules:  createDepartmentRequest.SchedulingRules,
				DefaultNodePools: createDepartmentRequest.DefaultNodePools,
				NodeTypes:        createDepartmentRequest.NodeTypes,
			})
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			department = response.Body.(api.Department)
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.SchedulingRules).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.SchedulingRules).To(Equal(department.SchedulingRules))
			Expect(project.Effective.DefaultNodePools).To(Equal(department.DefaultNodePools))
			Expect(project.Effective.NodeTypes).To(Equal(department.NodeTypes))

			getProjectsResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("name==%s", name)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getProjectsResponse.Code).To(Equal(http.StatusOK))
			project2 := getProjectsResponse.Body.(api.GetProjects200Response).Projects[0]
			Expect(project2.Effective.SchedulingRules).To(Equal(department.SchedulingRules))

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == projectId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(projectV1.CreatedAt).ToNot(BeNil())
			Expect(strconv.FormatInt(int64(projectV1.DepartmentId), 10)).To(Equal(departmentId))
			Expect(projectV1.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobTimeLimitSecs).To(Equal(float32(100)))
		})

		It("test department and projects override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			timeout1 := int32(100)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds:       ptr.To(timeout),
				TrainingJobMaxIdleDurationSeconds: ptr.To(timeout1),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(*department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			timeout2 := int32(20)
			timeout3 := int32(30)
			request.DefaultNodePools = &[]string{defaultNodePoolName}
			request.SchedulingRules = &api.SchedulingRules{
				InteractiveJobPreemptIdleDurationSeconds: ptr.To(timeout2),
				TrainingJobMaxIdleDurationSeconds:        ptr.To(timeout3),
			}
			request.NodeTypes = api.NodeTypesPerWorkload{
				Workspace: []string{nodeType2Id},
			}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(*project.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).NotTo(BeNil())
			Expect(project.NodeTypes.Workspace).To(Equal([]string{nodeType2Id}))
			Expect(project.SchedulingRules).NotTo(BeNil())
			Expect(*project.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project.Effective.NodeTypes.Workspace).To(Equal(project.NodeTypes.Workspace))
			Expect(*project.Effective.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(*project.Effective.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project.Effective.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project.Effective.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())

			getDep, err := departmentApiService.GetDepartment(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(getDep.Code).To(Equal(http.StatusOK))
			department = getDep.Body.(api.Department)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.DefaultNodePools).NotTo(BeNil())
			Expect(len(*department.DefaultNodePools)).To(Equal(3))
			Expect(*department.DefaultNodePools).To(Equal([]string{"pool2", "pool1", "pool3"}))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(*department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(*project.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(Equal([]string{nodeType2Id}))
			Expect(project.SchedulingRules).NotTo(BeNil())
			Expect(*project.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project.Effective.NodeTypes.Workspace).To(Equal(project.NodeTypes.Workspace))
			Expect(len(*project.Effective.NodeTypes.Names)).To(Equal(2))
			Expect(*project.Effective.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(*project.Effective.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project.Effective.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project.Effective.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())

			getProjectsResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("name==%s", name)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getProjectsResponse.Code).To(Equal(http.StatusOK))
			project2 := getProjectsResponse.Body.(api.GetProjects200Response).Projects[0]
			Expect(project2.CreatedAt).ToNot(BeNil())
			Expect(project2.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project2.ParentId).To(Equal(departmentId))
			Expect(*project2.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project2.NodeTypes.Training).To(BeNil())
			Expect(project2.NodeTypes.Workspace).To(Equal([]string{nodeType2Id}))
			Expect(project2.SchedulingRules).NotTo(BeNil())
			Expect(*project2.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project2.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project2.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project2.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project2.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(len(project2.Resources)).To(Equal(1))
			Expect(project2.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project2.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project2.Effective.NodeTypes.Workspace).To(Equal(project.NodeTypes.Workspace))
			Expect(len(*project2.Effective.NodeTypes.Names)).To(Equal(2))
			Expect(*project2.Effective.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(*project2.Effective.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(20)))
			Expect(*project2.Effective.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(30)))
			Expect(project2.Effective.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(project2.Effective.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == projectId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(projectV1.CreatedAt).ToNot(BeNil())
			Expect(strconv.FormatInt(int64(projectV1.DepartmentId), 10)).To(Equal(departmentId))
			Expect(projectV1.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(projectV1.NodeAffinity.Train.AffinityType).To(Equal("only_selected"))
			Expect(len(projectV1.NodeAffinity.Train.SelectedTypes)).To(Equal(1))
			Expect(projectV1.NodeAffinity.Train.SelectedTypes[0].Name).To(Equal(nodeType1Name))
			Expect(projectV1.NodeAffinity.Interactive.AffinityType).To(Equal("only_selected"))
			Expect(len(projectV1.NodeAffinity.Interactive.SelectedTypes)).To(Equal(1))
			Expect(projectV1.NodeAffinity.Interactive.SelectedTypes[0].Name).To(Equal(nodeType2Name))
			Expect(projectV1.InteractiveJobMaxIdleDurationSecs).To(Equal(float32(0)))
			Expect(projectV1.InteractivePreemptibleJobMaxIdleDurationSecs).To(Equal(float32(20)))
			Expect(projectV1.InteractiveJobTimeLimitSecs).To(Equal(float32(0)))
			Expect(projectV1.TrainingJobMaxIdleDurationSecs).To(Equal(float32(30)))
			Expect(projectV1.TrainingJobTimeLimitSecs).To(Equal(float32(10)))
			Expect(len(projectV1.NodePoolsResources)).To(Equal(1))
		})

		It("test department and projects override fields error", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			timeout1 := int32(100)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds:       ptr.To(timeout),
				TrainingJobMaxIdleDurationSeconds: ptr.To(timeout1),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			timeout2 := int32(200)
			request.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout2),
			}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test department and projects override fields after department update", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			timeout := int32(10)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds: ptr.To(timeout),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			department := response.Body.(api.Department)
			departmentId := department.Id
			departmentIds = append(departmentIds, departmentId)
			Expect(department.CreatedAt).ToNot(BeNil())
			Expect(department.UpdatedAt).To(Equal(department.CreatedAt))
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(len(department.NodeTypes.Training)).To(Equal(1))
			Expect(department.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(department.NodeTypes.Workspace).To(BeNil())
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(*department.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.DefaultNodePools = &[]string{defaultNodePoolName}
			timeout2 := int32(20)
			request.SchedulingRules = &api.SchedulingRules{
				InteractiveJobMaxIdleDurationSeconds: &timeout2,
			}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			projectId := project.Id
			projectIds = append(projectIds, projectId)

			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(*project.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(*project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(Equal(int32(20)))
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(Equal(department.NodeTypes.Training))
			Expect(project.Effective.NodeTypes.Workspace).To(BeNil())
			Expect(*project.Effective.SchedulingRules.TrainingJobTimeLimitSeconds).To(Equal(int32(10)))
			Expect(project.Effective.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.InteractiveJobTimeLimitSeconds).To(BeNil())
			Expect(*project.Effective.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(Equal(int32(20)))

			timeout3 := int32(100)
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
			depUpdateReq := api.DepartmentUpdateRequest{
				Description: department.Description,
				Resources:   resources,
				SchedulingRules: &api.SchedulingRules{
					InteractiveJobTimeLimitSeconds: ptr.To(timeout3),
				},
				DefaultNodePools: department.DefaultNodePools,
				NodeTypes: api.NodeTypesPerWorkload{
					Training:  nil,
					Workspace: []string{nodeType2Id},
				},
			}
			updateDep, err := departmentApiService.UpdateDepartment(ctx, department.Id, depUpdateReq)
			Expect(err).To(BeNil())
			department = updateDep.Body.(api.Department)
			Expect(len(department.Resources)).To(Equal(1))
			Expect(department.NodeTypes).NotTo(BeNil())
			Expect(department.NodeTypes.Training).To(BeNil())
			Expect(len(department.NodeTypes.Workspace)).To(Equal(1))
			Expect(department.NodeTypes.Workspace[0]).To(Equal(nodeType2Id))
			Expect(department.SchedulingRules).NotTo(BeNil())
			Expect(department.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(department.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(department.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(BeNil())
			Expect(*department.SchedulingRules.InteractiveJobTimeLimitSeconds).To(Equal(int32(100)))
			Expect(department.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			Expect(err).To(BeNil())
			project = getResponse.Body.(api.Project)
			Expect(project.CreatedAt).ToNot(BeNil())
			Expect(project.UpdatedAt).To(Equal(project.CreatedAt))
			Expect(project.ParentId).To(Equal(departmentId))
			Expect(*project.DefaultNodePools).To(Equal([]string{defaultNodePoolName}))
			Expect(project.NodeTypes.Training).To(BeNil())
			Expect(project.NodeTypes.Workspace).To(BeNil())
			Expect(*project.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(Equal(int32(20)))
			Expect(len(project.Resources)).To(Equal(1))
			Expect(project.Effective.DefaultNodePools).To(Equal(project.DefaultNodePools))
			Expect(project.Effective.NodeTypes.Training).To(BeNil())
			Expect(project.Effective.NodeTypes.Workspace).To(Equal(department.NodeTypes.Workspace))
			Expect(len(*project.Effective.NodeTypes.Names)).To(Equal(1))
			Expect(project.Effective.SchedulingRules.TrainingJobTimeLimitSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
			Expect(project.Effective.SchedulingRules.TrainingJobMaxIdleDurationSeconds).To(BeNil())
			Expect(*project.Effective.SchedulingRules.InteractiveJobTimeLimitSeconds).To(Equal(int32(100)))
			Expect(*project.Effective.SchedulingRules.InteractiveJobMaxIdleDurationSeconds).To(Equal(int32(20)))
		})
	})

	Context("test v1 PUT", func() {
		It("test PUT without override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
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
		})

		It("test PUT with override fields", func() {
			createDepartmentRequest := getEffectiveDepartmentCreateRequest("parent", clusterId)
			createDepartmentRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: []string{nodeType1Id},
			}
			timeout := int32(10)
			timeout1 := int32(100)
			createDepartmentRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobTimeLimitSeconds:       ptr.To(timeout),
				TrainingJobMaxIdleDurationSeconds: ptr.To(timeout1),
			}

			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil())
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
		})
	})
})

func getEffectiveDepartmentCreateRequest(name string, clusterId string) api.DepartmentCreationRequest {
	departmentRequest := api.DepartmentCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		Resources: createDefaultApiResources(false),
	}
	return departmentRequest
}

func getEffectiveDepartmentUpdateRequest(department api.Department) api.DepartmentUpdateRequest {
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
	departmentRequest := api.DepartmentUpdateRequest{
		Description:      department.Description,
		Resources:        resources,
		SchedulingRules:  department.SchedulingRules,
		DefaultNodePools: department.DefaultNodePools,
		NodeTypes:        department.NodeTypes,
	}
	return departmentRequest
}
