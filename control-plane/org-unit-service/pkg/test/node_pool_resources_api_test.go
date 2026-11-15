/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project and Department nodepool resources API tests", Ordered, func() {
	cluster1Id := uuid.New().String()
	cluster2Id := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)
	defaultNodePoolName1 := "default"
	defaultNodePoolId1 := float32(1)
	nodePoolName1 := "nodepool-a"
	nodePoolId1 := float32(555)
	nodePoolName2 := "nodepool-b"
	nodePoolId2 := float32(22)
	defaultNodePoolNameC2 := "default"
	defaultNodePoolIdC2 := float32(10)
	nodePoolName1c2 := "nodepool-a"
	nodePoolId1c2 := float32(20)
	nodePoolName2c2 := "nodepool-b"
	nodePoolId2c2 := float32(220)

	var defaultDepartment1 api.Department
	var defaultDepartment2 api.Department
	var projectIds []string
	var departmentIds []string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
	}, nil)
	clusterServiceClient.On("GetClusterByUuid", mock.Anything, mock.Anything).Return(&clustersOpenapi.DisplayedCluster{
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, cluster1Id).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName1,
				Id:                     fmt.Sprint(defaultNodePoolId1),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   nodePoolName1,
				Id:                     fmt.Sprint(nodePoolId1),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   nodePoolName2,
				Id:                     fmt.Sprint(nodePoolId2),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
			},
		},
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, cluster2Id).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolNameC2,
				Id:                     fmt.Sprint(defaultNodePoolIdC2),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   nodePoolName1c2,
				Id:                     fmt.Sprint(nodePoolId1c2),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   nodePoolName2c2,
				Id:                     fmt.Sprint(nodePoolId2c2),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	clusterServiceClient.On("GetNodepoolMinimal", mock.Anything, fmt.Sprint(nodePoolId2)).Return(&clustersOpenapi.NodepoolMinimal{
		Name:                   nodePoolName2,
		Id:                     fmt.Sprint(nodePoolId2),
		Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
		SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
	}, nil)
	clusterServiceClient.On("GetNodepoolMinimal", mock.Anything, fmt.Sprint(nodePoolId1)).Return(&clustersOpenapi.NodepoolMinimal{
		Name:                   nodePoolName1,
		Id:                     fmt.Sprint(nodePoolId1),
		Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
		SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting nodepool resources tests")

		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
	})

	BeforeAll(func() {
		createDepartmentRequest := getParentDepartmentCreateRequest(cluster1Id)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating default department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		defaultDepartment1 = response.Body.(api.Department)
		departmentIds = append(departmentIds, defaultDepartment1.Id)

		createDepartmentRequest = getParentDepartmentCreateRequest(cluster2Id)
		response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating default department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		defaultDepartment2 = response.Body.(api.Department)
		departmentIds = append(departmentIds, defaultDepartment2.Id)
	})

	AfterAll(func() {
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

	Context("test projects", func() {

		var projects = []api.Project{}

		BeforeAll(func() {
			name1 := fmt.Sprintf("1-%s", createProjectName())
			resources := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolId1),
						Name: defaultNodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(1)),
						Limit:           ptr.To(float64(1)),
						OverQuotaWeight: ptr.To(float64(1)),
					},
					Rank: ptr.To("Low"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(2)),
						Limit:           ptr.To(float64(2)),
						OverQuotaWeight: ptr.To(float64(2)),
					},
					Rank: ptr.To("High"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(3)),
						Limit:           ptr.To(float64(3)),
						OverQuotaWeight: ptr.To(float64(3)),
					},
					Rank: ptr.To("VeryHigh"),
				},
			}
			request := getProjectResourcesCreateRequest(name1, defaultDepartment1.Id, cluster1Id, resources)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projects = append(projects, createResponse.Body.(api.Project))
			projectIds = append(projectIds, createResponse.Body.(api.Project).Id)

			name2 := fmt.Sprintf("2-%s", createProjectName())
			resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolId1),
						Name: defaultNodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(10)),
						Limit:           ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(10)),
					},
					Rank: ptr.To("VeryHigh"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(20)),
						Limit:           ptr.To(float64(20)),
						OverQuotaWeight: ptr.To(float64(20)),
					},
					Rank: ptr.To("Low"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(30)),
						Limit:           ptr.To(float64(30)),
						OverQuotaWeight: ptr.To(float64(30)),
					},
					Rank: ptr.To("High"),
				},
			}
			request = getProjectResourcesCreateRequest(name2, defaultDepartment1.Id, cluster1Id, resources)
			createResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projects = append(projects, createResponse.Body.(api.Project))
			projectIds = append(projectIds, createResponse.Body.(api.Project).Id)

			name3 := fmt.Sprintf("3-%s", createProjectName())
			resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolIdC2),
						Name: defaultNodePoolNameC2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(11)),
						Limit:           ptr.To(float64(11)),
						OverQuotaWeight: ptr.To(float64(11)),
					},
					Rank: ptr.To("High"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1c2),
						Name: nodePoolName1c2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(21)),
						Limit:           ptr.To(float64(21)),
						OverQuotaWeight: ptr.To(float64(21)),
					},
					Rank: ptr.To("VeryHigh"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2c2),
						Name: nodePoolName2c2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(31)),
						Limit:           ptr.To(float64(31)),
						OverQuotaWeight: ptr.To(float64(31)),
					},
					Rank: ptr.To("Low"),
				},
			}
			request = getProjectResourcesCreateRequest(name3, defaultDepartment2.Id, cluster2Id, resources)
			createResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projects = append(projects, createResponse.Body.(api.Project))
			projectIds = append(projectIds, createResponse.Body.(api.Project).Id)

			name4 := fmt.Sprintf("4-%s", createProjectName())
			resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolIdC2),
						Name: defaultNodePoolNameC2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(110)),
						Limit:           ptr.To(float64(110)),
						OverQuotaWeight: ptr.To(float64(110)),
					},
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1c2),
						Name: nodePoolName1c2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(210)),
						Limit:           ptr.To(float64(210)),
						OverQuotaWeight: ptr.To(float64(210)),
					},
					Rank: ptr.To("High"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2c2),
						Name: nodePoolName2c2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(310)),
						Limit:           ptr.To(float64(310)),
						OverQuotaWeight: ptr.To(float64(310)),
					},
					Rank: ptr.To("VeryHigh"),
				},
			}
			request = getProjectResourcesCreateRequest(name4, defaultDepartment2.Id, cluster2Id, resources)
			createResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projects = append(projects, createResponse.Body.(api.Project))
			projectIds = append(projectIds, createResponse.Body.(api.Project).Id)

		})

		It("test projects count", func() {
			response, err := projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(6)))

			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id), fmt.Sprintf("nodePoolName==%s", nodePoolName1c2)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(2)))

			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId2c2)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(2)))

			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitName==%s", projects[0].Name)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))

			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster2Id), fmt.Sprintf("rank==%s", models.DefaultResourceRank)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(1)))

			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id), "rank==High"}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(2)))

			// test search exact
			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{}, projects[0].Name)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))

			// test search partial
			project0PartialName := projects[0].Name[0 : len(projects[0].Name)-2]
			response, err = projectApiService.CountProjectNodePoolResources(ctx, []string{}, project0PartialName)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))
		})

		It("test projects get", func() {
			response, err := projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(len(response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources)).To(Equal(6))

			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId1c2)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_ORG_UNIT_NAME, "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(2))
			Expect(nodePoolResources[0].Id).ToNot(BeEmpty())
			Expect(nodePoolResources[0].OrgUnitId).To(Equal(projects[3].Id))
			Expect(nodePoolResources[0].OrgUnitName).To(Equal(projects[3].Name))
			Expect(nodePoolResources[0].ParentId).To(Equal(projects[3].ParentId))
			Expect(nodePoolResources[0].ClusterId).To(Equal(projects[3].ClusterId))
			Expect(nodePoolResources[0].NodePoolId).To(Equal(fmt.Sprint(nodePoolId1c2)))
			Expect(nodePoolResources[0].NodePoolName).To(Equal(fmt.Sprint(nodePoolName1c2)))
			Expect(nodePoolResources[0].Rank).To(Equal("High"))
			Expect(nodePoolResources[0].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(210))))
			Expect(nodePoolResources[1].NodePoolId).To(Equal(fmt.Sprint(nodePoolId1c2)))
			Expect(nodePoolResources[1].OrgUnitName).To(Equal(projects[2].Name))
			Expect(nodePoolResources[1].OrgUnitId).To(Equal(projects[2].Id))
			Expect(nodePoolResources[1].ParentId).To(Equal(projects[2].ParentId))
			Expect(nodePoolResources[1].Rank).To(Equal("VeryHigh"))
			Expect(nodePoolResources[1].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(21))))
			// test the DESC order by orgUnitName
			Expect(strings.Compare(nodePoolResources[0].OrgUnitName, nodePoolResources[1].OrgUnitName)).To(BeNumerically(">", 0))

			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{"rank==High", fmt.Sprintf("nodePoolId==%v", defaultNodePoolIdC2)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_ORG_UNIT_NAME, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources = response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(1))
			Expect(nodePoolResources[0].NodePoolId).To(Equal(fmt.Sprint(defaultNodePoolIdC2)))
			Expect(nodePoolResources[0].NodePoolName).To(Equal(defaultNodePoolNameC2))
			Expect(nodePoolResources[0].OrgUnitName).To(Equal(projects[2].Name))
			Expect(nodePoolResources[0].OrgUnitId).To(Equal(projects[2].Id))
			Expect(nodePoolResources[0].ParentId).To(Equal(projects[2].ParentId))
			Expect(nodePoolResources[0].Rank).To(Equal("High"))
			Expect(nodePoolResources[0].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(11))))

			// test search exact
			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "", "", 0, 0, projects[1].Name)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources = response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(3))
			for _, res := range nodePoolResources {
				Expect(res.OrgUnitId).To(Equal(projects[1].Id))
				Expect(res.OrgUnitName).To(Equal(projects[1].Name))
			}

			// test search partial
			proj2NamePartial := projects[1].Name[0 : len(projects[1].Name)-2]
			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "", "", 0, 0, proj2NamePartial)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources = response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(3))
			for _, res := range nodePoolResources {
				Expect(res.OrgUnitId).To(Equal(projects[1].Id))
				Expect(res.OrgUnitName).To(Equal(projects[1].Name))
			}
		})

		It("test projects get with gpuOverQuotaWeight", func() {
			// filterBy gpuOverQuotaWeight not allowed
			response, err := projectApiService.GetProjectNodePoolResources(ctx, []string{"gpuOverQuotaWeight==3"}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_ORG_UNIT_NAME, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			// sortBy gpuOverQuotaWeight not allowed without node pool filterBy
			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitName==balbla")}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			// sortBy gpuOverQuotaWeight not allowed when filtered nodepool doesn't have adjustable weights
			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId1c2)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId2)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK), fmt.Sprintf("err: %v, response body: %+v", err, response.Body))
			// validate each next item has equal or higher weight than the previous one
			nodePoolResources := response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			prevWeight := float64(-1)
			for _, res := range nodePoolResources {
				currentWeight := ptr.Deref(res.Resources.Gpu.OverQuotaWeight, 0)
				Expect(currentWeight).To(BeNumerically(">=", prevWeight))
				prevWeight = currentWeight
			}
		})

		It("test projects get with gpuDeserved sort", func() {
			// sortBy gpuDeserved ASC
			response, err := projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_DESERVED, "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			// validate each next item has equal or higher deserved than the previous one
			prevDeserved := float64(-1)
			for _, res := range nodePoolResources {
				currentDeserved := ptr.Deref(res.Resources.Gpu.Deserved, 0)
				Expect(currentDeserved).To(BeNumerically(">=", prevDeserved))
				prevDeserved = currentDeserved
			}

			// sortBy gpuDeserved DESC
			response, err = projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_DESERVED, "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources = response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			// validate each next item has equal or lower deserved than the previous one
			prevDeserved = ptr.Deref(nodePoolResources[0].Resources.Gpu.Deserved, 0)
			for _, res := range nodePoolResources[1:] {
				currentDeserved := ptr.Deref(res.Resources.Gpu.Deserved, 0)
				Expect(currentDeserved).To(BeNumerically("<=", prevDeserved))
				prevDeserved = currentDeserved
			}
		})

		It("test sum weights", func() {
			response, err := projectApiService.SumProjectNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id), "rank==High"})
			Expect(err).To(BeNil())
			weightsSum := response.Body.(api.SumResponse).WeightsSum
			Expect(weightsSum).To(Equal(float64(30)))

			response, err = projectApiService.SumProjectNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			weightsSum = response.Body.(api.SumResponse).WeightsSum
			Expect(weightsSum).To(Equal(float64(33)))
		})

		It("test sum with permitted scopes", func() {
			// Baseline: sum for cluster1 + nodePoolId2 without any permitted-scopes restriction
			baseResp, err := projectApiService.SumProjectNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(baseResp.Code).To(Equal(http.StatusOK))
			baseSum := baseResp.Body.(api.SumResponse).WeightsSum
			// From the previous test we know this should be 33 (3 from project[0], 30 from project[1])
			Expect(baseSum).To(Equal(float64(33)))

			// When only the second project is permitted, sum should include only its 30 weight on nodePoolId2
			limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projects[1].Id},
			})
			resp, err := projectApiService.SumProjectNodePoolResources(limitedCtx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			sum := resp.Body.(api.SumResponse).WeightsSum
			Expect(sum).To(Equal(float64(30)))

			// When only the first project is permitted, sum should include only its 3 weight on nodePoolId2
			limitedCtx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projects[0].Id},
			})
			resp, err = projectApiService.SumProjectNodePoolResources(limitedCtx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			sum = resp.Body.(api.SumResponse).WeightsSum
			Expect(sum).To(Equal(float64(3)))
		})

		It("test sum weights uses deserved when adjustable weights is disabled", func() {
			// nodePoolId1 has AdjustableWeights=false, so WeightsSum should equal sum of Deserved values
			// projects[0] on nodePoolId1: Deserved=2, OverQuotaWeight=2
			// projects[1] on nodePoolId1: Deserved=20, OverQuotaWeight=20
			// Total Deserved = 2 + 20 = 22
			response, err := projectApiService.SumProjectNodePoolResources(ctx, fmt.Sprint(nodePoolId1), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			weightsSum := response.Body.(api.SumResponse).WeightsSum
			// When AdjustableWeights is false, WeightsSum should be calculated from Deserved, not OverQuotaWeight
			Expect(weightsSum).To(Equal(float64(22)))
		})

		It("test projects get returns QuotaStatus after status update", func() {
			// Update status for projects[0] with specific quota values for each node pool
			nodePoolQuotaStatuses := []api.QuotaStatusNodePool{
				{
					NodePoolName: defaultNodePoolName1,
					NodePoolId:   fmt.Sprint(defaultNodePoolId1),
					Allocated: api.QuotaStatusResource{
						Gpu:    1.5,
						Cpu:    100,
						Memory: 1000,
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    0.5,
						Cpu:    50,
						Memory: 500,
					},
					Requested: api.QuotaStatusResource{
						Gpu:    2,
						Cpu:    150,
						Memory: 1500,
					},
				},
				{
					NodePoolName: nodePoolName1,
					NodePoolId:   fmt.Sprint(nodePoolId1),
					Allocated: api.QuotaStatusResource{
						Gpu:    3,
						Cpu:    200,
						Memory: 2000,
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    75,
						Memory: 750,
					},
					Requested: api.QuotaStatusResource{
						Gpu:    4,
						Cpu:    250,
						Memory: 2500,
					},
				},
				{
					NodePoolName: nodePoolName2,
					NodePoolId:   fmt.Sprint(nodePoolId2),
					Allocated: api.QuotaStatusResource{
						Gpu:    5.5,
						Cpu:    300,
						Memory: 3000,
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    2.5,
						Cpu:    125,
						Memory: 1250,
					},
					Requested: api.QuotaStatusResource{
						Gpu:    6,
						Cpu:    350,
						Memory: 3500,
					},
				},
			}

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        projects[0].Id,
				ClusterId: cluster1Id,
				Status: api.ProjectStatus{
					Phase:                 "Ready",
					NodePoolQuotaStatuses: nodePoolQuotaStatuses,
					QuotaStatus: api.QuotaStatus{
						Allocated:               api.QuotaStatusResource{Gpu: 10, Cpu: 600, Memory: 6000},
						AllocatedNonPreemptible: api.QuotaStatusResource{Gpu: 4, Cpu: 250, Memory: 2500},
						Requested:               api.QuotaStatusResource{Gpu: 12, Cpu: 750, Memory: 7500},
					},
				},
			}

			updateResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			// Retrieve node pool resources and verify QuotaStatus is populated
			response, err := projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", projects[0].Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(3))

			// Find and validate each node pool's QuotaStatus
			for _, npr := range nodePoolResources {
				switch npr.NodePoolId {
				case fmt.Sprint(defaultNodePoolId1):
					Expect(npr.QuotaStatus.Allocated.Gpu).To(Equal(float64(1.5)))
					Expect(npr.QuotaStatus.Allocated.Cpu).To(Equal(float64(100)))
					Expect(npr.QuotaStatus.Allocated.Memory).To(Equal(float64(1000)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(0.5)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Cpu).To(Equal(float64(50)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Memory).To(Equal(float64(500)))
					Expect(npr.QuotaStatus.Requested.Gpu).To(Equal(float64(2)))
					Expect(npr.QuotaStatus.Requested.Cpu).To(Equal(float64(150)))
					Expect(npr.QuotaStatus.Requested.Memory).To(Equal(float64(1500)))
				case fmt.Sprint(nodePoolId1):
					Expect(npr.QuotaStatus.Allocated.Gpu).To(Equal(float64(3)))
					Expect(npr.QuotaStatus.Allocated.Cpu).To(Equal(float64(200)))
					Expect(npr.QuotaStatus.Allocated.Memory).To(Equal(float64(2000)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(1)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Cpu).To(Equal(float64(75)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Memory).To(Equal(float64(750)))
					Expect(npr.QuotaStatus.Requested.Gpu).To(Equal(float64(4)))
					Expect(npr.QuotaStatus.Requested.Cpu).To(Equal(float64(250)))
					Expect(npr.QuotaStatus.Requested.Memory).To(Equal(float64(2500)))
				case fmt.Sprint(nodePoolId2):
					Expect(npr.QuotaStatus.Allocated.Gpu).To(Equal(float64(5.5)))
					Expect(npr.QuotaStatus.Allocated.Cpu).To(Equal(float64(300)))
					Expect(npr.QuotaStatus.Allocated.Memory).To(Equal(float64(3000)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(2.5)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Cpu).To(Equal(float64(125)))
					Expect(npr.QuotaStatus.AllocatedNonPreemptible.Memory).To(Equal(float64(1250)))
					Expect(npr.QuotaStatus.Requested.Gpu).To(Equal(float64(6)))
					Expect(npr.QuotaStatus.Requested.Cpu).To(Equal(float64(350)))
					Expect(npr.QuotaStatus.Requested.Memory).To(Equal(float64(3500)))
				default:
					Fail(fmt.Sprintf("Unexpected nodePoolId: %s", npr.NodePoolId))
				}
			}
		})

		It("test projects get returns zero QuotaStatus when no status update was made", func() {
			// projects[1] has no status update, so QuotaStatus should be zero values
			response, err := projectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", projects[1].Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(3))

			// All QuotaStatus fields should be zero since no status update was made
			for _, npr := range nodePoolResources {
				Expect(npr.QuotaStatus.Allocated.Gpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.Allocated.Cpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.Allocated.Memory).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.AllocatedNonPreemptible.Cpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.AllocatedNonPreemptible.Memory).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.Requested.Gpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.Requested.Cpu).To(Equal(float64(0)))
				Expect(npr.QuotaStatus.Requested.Memory).To(Equal(float64(0)))
			}
		})
	})

	Context("test departments", func() {

		var departments = []api.Department{}

		BeforeAll(func() {
			name1 := fmt.Sprintf("1-%s", createDepartmentName())
			resources := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolId1),
						Name: defaultNodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(1)),
						Limit:           ptr.To(float64(1)),
						OverQuotaWeight: ptr.To(float64(1)),
					},
					Rank: ptr.To("Low"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(2)),
						Limit:           ptr.To(float64(2)),
						OverQuotaWeight: ptr.To(float64(2)),
					},
					Rank: ptr.To("High"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(3)),
						Limit:           ptr.To(float64(3)),
						OverQuotaWeight: ptr.To(float64(3)),
					},
					Rank: ptr.To("VeryHigh"),
				},
			}
			request := getDepartmentResourcesCreateRequest(name1, cluster1Id, resources)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departments = append(departments, createResponse.Body.(api.Department))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			name2 := fmt.Sprintf("2-%s", createDepartmentName())
			resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolId1),
						Name: defaultNodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(10)),
						Limit:           ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(10)),
					},
					Rank: ptr.To("VeryHigh"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(20)),
						Limit:           ptr.To(float64(20)),
						OverQuotaWeight: ptr.To(float64(20)),
					},
					Rank: ptr.To("Low"),
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(30)),
						Limit:           ptr.To(float64(30)),
						OverQuotaWeight: ptr.To(float64(30)),
					},
					Rank: ptr.To("High"),
				},
			}
			request = getDepartmentResourcesCreateRequest(name2, cluster1Id, resources)
			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departments = append(departments, createResponse.Body.(api.Department))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			name3 := fmt.Sprintf("3-%s", createDepartmentName())
			resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(defaultNodePoolId1),
						Name: defaultNodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(10)),
						Limit:           ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(10)),
					},
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(20)),
						Limit:           ptr.To(float64(20)),
						OverQuotaWeight: ptr.To(float64(20)),
					},
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(30)),
						Limit:           ptr.To(float64(30)),
						OverQuotaWeight: ptr.To(float64(30)),
					},
				},
			}
			request = getDepartmentResourcesCreateRequest(name3, cluster1Id, resources)
			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departments = append(departments, createResponse.Body.(api.Department))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)
		})

		It("test departments count", func() {
			response, err := departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster2Id)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(12)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id), fmt.Sprintf("nodePoolName==%s", nodePoolName1)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(4)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId2c2)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(1)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitName==%s", departments[0].Name)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id), fmt.Sprintf("rank==%s", models.DefaultResourceRank)}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(6)))

			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id), "rank==High"}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(2)))

			// test search
			response, err = departmentApiService.CountDepartmentNodePoolResources(ctx, []string{}, departments[0].Name)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.CountResponse).Count).To(Equal(int64(3)))
		})

		It("test departments get", func() {
			response, err := departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(len(response.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources)).To(Equal(12))

			response, err = departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId1), "orgUnitName!@default"}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_ORG_UNIT_NAME, "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(Equal(3))
			Expect(nodePoolResources[0].Id).ToNot(BeEmpty())
			Expect(nodePoolResources[0].OrgUnitId).To(Equal(departments[0].Id))
			Expect(nodePoolResources[0].OrgUnitName).To(Equal(departments[0].Name))
			Expect(nodePoolResources[0].ClusterId).To(Equal(departments[0].ClusterId))
			Expect(nodePoolResources[0].NodePoolId).To(Equal(fmt.Sprint(nodePoolId1)))
			Expect(nodePoolResources[0].NodePoolName).To(Equal(fmt.Sprint(nodePoolName1)))
			Expect(nodePoolResources[0].Rank).To(Equal("High"))
			Expect(nodePoolResources[0].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(2))))
			Expect(nodePoolResources[1].NodePoolId).To(Equal(fmt.Sprint(nodePoolId1)))
			Expect(nodePoolResources[1].OrgUnitName).To(Equal(departments[1].Name))
			Expect(nodePoolResources[1].OrgUnitId).To(Equal(departments[1].Id))
			Expect(nodePoolResources[1].Rank).To(Equal("Low"))
			Expect(nodePoolResources[1].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(20))))
			Expect(nodePoolResources[2].NodePoolId).To(Equal(fmt.Sprint(nodePoolId1)))
			Expect(nodePoolResources[2].OrgUnitName).To(Equal(departments[2].Name))
			Expect(nodePoolResources[2].OrgUnitId).To(Equal(departments[2].Id))
			Expect(nodePoolResources[2].Rank).To(Equal("MediumLow"))
			Expect(nodePoolResources[2].Resources.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(20))))
			// test the ASC order by orgUnitName
			Expect(strings.Compare(nodePoolResources[0].OrgUnitName, nodePoolResources[1].OrgUnitName)).To(BeNumerically("<", 0))
			Expect(strings.Compare(nodePoolResources[1].OrgUnitName, nodePoolResources[2].OrgUnitName)).To(BeNumerically("<", 0))
		})

		It("test departments get with gpuOverQuotaWeight", func() {
			// filterBy gpuOverQuotaWeight not allowed
			response, err := departmentApiService.GetDepartmentNodePoolResources(ctx, []string{"gpuOverQuotaWeight==3"}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_NODE_POOL_NAME, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			// sortBy gpuOverQuotaWeight not allowed without node pool filterBy
			response, err = departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitName==balbla")}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			// sortBy gpuOverQuotaWeight not allowed when filtered nodepool doesn't have adjustable weights
			response, err = departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId1c2)}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("nodePoolId==%v", nodePoolId2)}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_OVER_QUOTA_WEIGHT, "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			// validate each next item has equal or higher weight than the previous one
			nodePoolResources := response.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			prevWeight := float64(-1)
			for _, res := range nodePoolResources {
				currentWeight := ptr.Deref(res.Resources.Gpu.OverQuotaWeight, 0)
				Expect(currentWeight).To(BeNumerically(">=", prevWeight))
				prevWeight = currentWeight
			}
		})

		It("test departments get with gpuDeserved sort", func() {
			// sortBy gpuDeserved ASC
			response, err := departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_DESERVED, "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources := response.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			// validate each next item has equal or higher deserved than the previous one
			prevDeserved := float64(-1)
			for _, res := range nodePoolResources {
				currentDeserved := ptr.Deref(res.Resources.Gpu.Deserved, 0)
				Expect(currentDeserved).To(BeNumerically(">=", prevDeserved))
				prevDeserved = currentDeserved
			}

			// sortBy gpuDeserved DESC
			response, err = departmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", cluster1Id)}, api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_GPU_DESERVED, "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodePoolResources = response.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			Expect(len(nodePoolResources)).To(BeNumerically(">", 0))
			// validate each next item has equal or lower deserved than the previous one
			prevDeserved = ptr.Deref(nodePoolResources[0].Resources.Gpu.Deserved, 0)
			for _, res := range nodePoolResources[1:] {
				currentDeserved := ptr.Deref(res.Resources.Gpu.Deserved, 0)
				Expect(currentDeserved).To(BeNumerically("<=", prevDeserved))
				prevDeserved = currentDeserved
			}
		})

		It("test sum deserved and weights", func() {
			response, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id), "rank==High"})
			Expect(err).To(BeNil())
			sumResponse := response.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(30)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(30)))

			response, err = departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			sumResponse = response.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(63)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(63)))
		})

		It("test sum with different deserved and weights values", func() {
			// Create departments where OverQuotaWeight differs from Deserved
			deptDiffWeights1 := fmt.Sprintf("dept-diff-weights-1-%s", createDepartmentName())
			resources1 := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(10)),
						Limit:           ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(5)), // Weight is half of Deserved
					},
				},
			}
			request := getDepartmentResourcesCreateRequest(deptDiffWeights1, cluster1Id, resources1)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			deptDiffWeights2 := fmt.Sprintf("dept-diff-weights-2-%s", createDepartmentName())
			resources2 := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(20)),
						Limit:           ptr.To(float64(20)),
						OverQuotaWeight: ptr.To(float64(8)), // Weight is less than half of Deserved
					},
				},
			}
			request = getDepartmentResourcesCreateRequest(deptDiffWeights2, cluster1Id, resources2)
			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			// Query sum for nodePoolId2 in cluster1
			// Existing departments on nodePoolId2: departments[0]=3/3, departments[1]=30/30, departments[2]=30/30
			// New departments: deptDiffWeights1=10/5, deptDiffWeights2=20/8
			// DeservedSum = 3 + 30 + 30 + 10 + 20 = 93
			// WeightsSum = 3 + 30 + 30 + 5 + 8 = 76
			response, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			sumResponse := response.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(93)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(76)))
		})

		It("test sum weights with zero weight department", func() {
			// Create a department with Deserved > 0 but OverQuotaWeight = 0
			deptZeroWeight := fmt.Sprintf("dept-zero-weight-%s", createDepartmentName())
			resources := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(15)),
						Limit:           ptr.To(float64(15)),
						OverQuotaWeight: ptr.To(float64(0)), // Zero weight - should be included in sum as 0
					},
				},
			}
			request := getDepartmentResourcesCreateRequest(deptZeroWeight, cluster1Id, resources)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			// Query sum for nodePoolId2 in cluster1
			// After previous test: DeservedSum was 93, WeightsSum was 76
			// Adding deptZeroWeight: Deserved=15, Weight=0
			// DeservedSum = 93 + 15 = 108
			// WeightsSum = 76 + 0 = 76 (unchanged)
			response, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			sumResponse := response.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(108)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(76)))
		})

		It("test sum deserved returns -1 when any department has unlimited", func() {
			// Create a department with gpu_deserved = -1 (meaning unlimited/not set)
			deptWithNegative := fmt.Sprintf("dept-negative-%s", createDepartmentName())
			resourcesNegative := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved: ptr.To(float64(-1)), // -1 means unlimited/not set - should make entire sum -1
					},
				},
			}
			request := getDepartmentResourcesCreateRequest(deptWithNegative, cluster1Id, resourcesNegative)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			createdDeptNegative := createResponse.Body.(api.Department)

			// Sum should be -1 for cluster1 + nodePoolId2 since one department has unlimited (-1)
			response, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			sumResponse := response.Body.(api.DepartmentSumResponse)
			// departments[0]: 3, departments[1]: 30, departments[2]: 30, deptWithNegative: -1 (unlimited)
			// DeservedSum should be -1 (unlimited) because at least one department is unlimited
			// WeightsSum should still be calculated normally: 76
			Expect(sumResponse.DeservedSum).To(Equal(float64(-1)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(76)))

			// Create a department with gpu_deserved = 0 - should be included in sum normally
			deptWithZero := fmt.Sprintf("dept-zero-%s", createDepartmentName())
			resourcesZero := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId2),
						Name: nodePoolName2,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved: ptr.To(float64(0)), // 0 should be included in sum
					},
				},
			}
			request = getDepartmentResourcesCreateRequest(deptWithZero, cluster1Id, resourcesZero)
			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			createdDeptZero := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, createdDeptZero.Id)

			// Adding deptWithZero: Deserved=0 (included), Weight=0 (included)
			// But since deptWithNegative still exists with -1, DeservedSum should still be -1
			// WeightsSum = 76 + 0 = 76
			response, err = departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			sumResponse = response.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(-1)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(76)))

			// delete createdDeptNegative.Id
			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, createdDeptNegative.Id)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("test sum with permitted scopes", func() {
			// Baseline: sum for cluster1 + nodePoolId2 without any permitted-scopes restriction
			// After previous tests: DeservedSum=108, WeightsSum=76
			baseResp, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(baseResp.Code).To(Equal(http.StatusOK))
			baseSumResponse := baseResp.Body.(api.DepartmentSumResponse)
			Expect(baseSumResponse.DeservedSum).To(Equal(float64(108)))
			Expect(baseSumResponse.WeightsSum).To(Equal(float64(76)))

			// When only departments[1] is permitted, sum should include only its 30 deserved / 30 weight on nodePoolId2
			limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{departments[1].Id},
			})
			resp, err := departmentApiService.SumDepartmentNodePoolResources(limitedCtx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			sumResponse := resp.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(30)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(30)))

			// When only departments[0] is permitted, sum should include only its 3 deserved / 3 weight on nodePoolId2
			limitedCtx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{departments[0].Id},
			})
			resp, err = departmentApiService.SumDepartmentNodePoolResources(limitedCtx, fmt.Sprint(nodePoolId2), []string{
				fmt.Sprintf("clusterId==%v", cluster1Id),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			sumResponse = resp.Body.(api.DepartmentSumResponse)
			Expect(sumResponse.DeservedSum).To(Equal(float64(3)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(3)))
		})

		It("test sum uses deserved when adjustable weights is disabled", func() {
			// nodePoolId1 has AdjustableWeights=false, so WeightsSum should equal DeservedSum
			// departments[0] on nodePoolId1: Deserved=2, OverQuotaWeight=2
			// departments[1] on nodePoolId1: Deserved=20, OverQuotaWeight=20
			// departments[2] on nodePoolId1: Deserved=20, OverQuotaWeight=20
			// defaultDepartment1 on nodePoolId1: Deserved=0
			// Total Deserved = 2 + 20 + 20 + 0 = 42
			response, err := departmentApiService.SumDepartmentNodePoolResources(ctx, fmt.Sprint(nodePoolId1), []string{fmt.Sprintf("clusterId==%v", cluster1Id)})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			sumResponse := response.Body.(api.DepartmentSumResponse)
			// When AdjustableWeights is false, both DeservedSum and WeightsSum should be calculated from Deserved
			Expect(sumResponse.DeservedSum).To(Equal(float64(42)))
			Expect(sumResponse.WeightsSum).To(Equal(float64(42)))
		})
	})

	Context("test nodepool delete and recreate with same name", func() {
		// This test validates that when a nodepool is deleted and recreated with the same name
		// but a different ID, the node-pool-resources APIs return correct results without duplications

		testClusterId := uuid.New().String()
		originalNodePoolId := "1000"
		recreatedNodePoolId := "2000"
		nodePoolName := "test-pool"
		defaultNodePoolId := "1"
		defaultNodePoolName := "default"

		var testDepartment api.Department
		var testProject api.Project
		var notificationsService api.NotificationsAPIServicer
		var testProjectApiService api.ProjectsAPIServicer
		var testDepartmentApiService api.DepartmentsAPIServicer
		var testProjectIds []string
		var testDepartmentIds []string

		// Helper to create mock cluster client with specified nodepools
		createMockClientWithNodepools := func(nodepools []clustersOpenapi.NodepoolMinimal) *mocks.IClusterClient {
			mockClient := &mocks.IClusterClient{}
			mockClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
			}, nil)
			mockClient.On("GetClusterByUuid", mock.Anything, mock.Anything).Return(&clustersOpenapi.DisplayedCluster{
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.24")),
			}, nil)
			mockClient.On("GetNodePoolsMinimalByCluster", mock.Anything, testClusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: nodepools,
			}, nil)
			for _, np := range nodepools {
				mockClient.On("GetNodepoolMinimal", mock.Anything, np.Id).Return(&clustersOpenapi.NodepoolMinimal{
					Name:                   np.Name,
					Id:                     np.Id,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				}, nil)
			}
			return mockClient
		}

		BeforeAll(func() {
			// Setup initial nodepools: default + test-pool
			initialNodepools := []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     defaultNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
				{
					Name:                   nodePoolName,
					Id:                     originalNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			}
			mockClient := createMockClientWithNodepools(initialNodepools)
			projectDeletionSvc := services.NewProjectDeletionService(mockClient, nil, nil)
			testDepartmentApiService = departments.NewDepartmentApiService(mockClient, nil)
			testProjectApiService = projects.NewProjectApiService(projectDeletionSvc, mockClient, nil)
			notificationsService = services.NewNotificationsApiService(mockClient)

			// Create department with resources on both default and test-pool
			deptResources := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   defaultNodePoolId,
						Name: defaultNodePoolName,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(5)),
						Limit:           ptr.To(float64(10)),
						OverQuotaWeight: ptr.To(float64(5)),
					},
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   originalNodePoolId,
						Name: nodePoolName,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(15)),
						Limit:           ptr.To(float64(20)),
						OverQuotaWeight: ptr.To(float64(15)),
					},
				},
			}
			deptRequest := api.DepartmentCreationRequest{
				Name:      fmt.Sprintf("dept-np-recreate-%s", uuid.New().String()[:8]),
				ClusterId: testClusterId,
				Resources: deptResources,
			}
			deptResponse, err := testDepartmentApiService.CreateDepartment(ctx, deptRequest)
			Expect(err).To(BeNil())
			Expect(deptResponse.Code).To(Equal(http.StatusCreated))
			testDepartment = deptResponse.Body.(api.Department)
			testDepartmentIds = append(testDepartmentIds, testDepartment.Id)

			// Create project with resources on both default and test-pool
			projResources := []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   defaultNodePoolId,
						Name: defaultNodePoolName,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(2)),
						Limit:           ptr.To(float64(5)),
						OverQuotaWeight: ptr.To(float64(2)),
					},
				},
				{
					NodePool: &api.ResourcesNodePool{
						Id:   originalNodePoolId,
						Name: nodePoolName,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(8)),
						Limit:           ptr.To(float64(15)),
						OverQuotaWeight: ptr.To(float64(8)),
					},
				},
			}
			projRequest := api.ProjectCreationRequest{
				Name:      fmt.Sprintf("proj-np-recreate-%s", uuid.New().String()[:8]),
				ClusterId: testClusterId,
				ParentId:  &testDepartment.Id,
				Resources: projResources,
			}
			projResponse, err := testProjectApiService.CreateProject(ctx, projRequest)
			Expect(err).To(BeNil())
			Expect(projResponse.Code).To(Equal(http.StatusCreated))
			testProject = projResponse.Body.(api.Project)
			testProjectIds = append(testProjectIds, testProject.Id)
		})

		AfterAll(func() {
			// Cleanup: recreate services with all nodepools to allow deletion
			allNodepools := []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     defaultNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
				{
					Name:                   nodePoolName,
					Id:                     recreatedNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			}
			mockClient := createMockClientWithNodepools(allNodepools)
			projectDeletionSvc := services.NewProjectDeletionService(mockClient, nil, nil)
			cleanupProjectService := projects.NewProjectApiService(projectDeletionSvc, mockClient, nil)
			cleanupDepartmentService := departments.NewDepartmentApiService(mockClient, nil)

			for _, projectId := range testProjectIds {
				deleteResponse, err := cleanupProjectService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail(fmt.Sprintf("Failed deleting project %s", projectId))
				}
			}
			testProjectIds = []string{}
			for _, departmentId := range testDepartmentIds {
				deleteResponse, err := cleanupDepartmentService.DeleteDepartment(ctx, departmentId)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail(fmt.Sprintf("Failed deleting department %s", departmentId))
				}
			}
			testDepartmentIds = []string{}
		})

		It("returns no duplications after nodepool delete and recreate with same name", func() {
			// Step 1: Verify initial state - both department and project have 2 nodepool resources
			deptNpResponse, err := testDepartmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", testDepartment.Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(deptNpResponse.Code).To(Equal(http.StatusOK))
			initialDeptResources := deptNpResponse.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			Expect(len(initialDeptResources)).To(Equal(2))

			projNpResponse, err := testProjectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", testProject.Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(projNpResponse.Code).To(Equal(http.StatusOK))
			initialProjResources := projNpResponse.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			Expect(len(initialProjResources)).To(Equal(2))

			// Step 2: Delete the nodepool via callback
			deleteResponse, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        originalNodePoolId,
				Name:      nodePoolName,
				ClusterId: testClusterId,
				Operation: services.DeletedOperation,
			})
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusOK))

			// Step 3: Recreate nodepool with same name but different ID
			// First, update the mock to reflect the new nodepool
			recreatedNodepools := []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     defaultNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
				{
					Name:                   nodePoolName,
					Id:                     recreatedNodePoolId,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			}
			mockClient := createMockClientWithNodepools(recreatedNodepools)
			projectDeletionSvc := services.NewProjectDeletionService(mockClient, nil, nil)
			testDepartmentApiService = departments.NewDepartmentApiService(mockClient, nil)
			testProjectApiService = projects.NewProjectApiService(projectDeletionSvc, mockClient, nil)
			notificationsService = services.NewNotificationsApiService(mockClient)

			createResponse, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
				Id:        recreatedNodePoolId,
				Name:      nodePoolName,
				ClusterId: testClusterId,
				Operation: services.CreatedOperation,
			})
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			// Step 4: Verify department node-pool-resources - should have exactly 2 resources, no duplications
			deptNpResponse, err = testDepartmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", testDepartment.Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(deptNpResponse.Code).To(Equal(http.StatusOK))
			finalDeptResources := deptNpResponse.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources

			// Validate no duplications and correct resources
			Expect(len(finalDeptResources)).To(Equal(2), "Department should have exactly 2 nodepool resources (no duplications)")

			deptNodePoolIds := make(map[string]bool)
			deptNodePoolNames := make(map[string]bool)
			for _, npr := range finalDeptResources {
				// Check for duplicate nodepool IDs
				Expect(deptNodePoolIds[npr.NodePoolId]).To(BeFalse(), fmt.Sprintf("Duplicate nodepool ID found in department resources: %s", npr.NodePoolId))
				deptNodePoolIds[npr.NodePoolId] = true

				// Verify nodepool names
				deptNodePoolNames[npr.NodePoolName] = true

				// Verify correct org unit association
				Expect(npr.OrgUnitId).To(Equal(testDepartment.Id))
				Expect(npr.OrgUnitName).To(Equal(testDepartment.Name))
				Expect(npr.ClusterId).To(Equal(testClusterId))
			}

			// Verify we have both expected nodepools (default and the recreated one)
			Expect(deptNodePoolIds[defaultNodePoolId]).To(BeTrue(), "Default nodepool should be present")
			Expect(deptNodePoolIds[recreatedNodePoolId]).To(BeTrue(), "Recreated nodepool should be present")
			Expect(deptNodePoolIds[originalNodePoolId]).To(BeFalse(), "Original (deleted) nodepool should NOT be present")

			// Step 5: Verify project node-pool-resources - should have exactly 2 resources, no duplications
			projNpResponse, err = testProjectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("orgUnitId==%s", testProject.Id)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(projNpResponse.Code).To(Equal(http.StatusOK))
			finalProjResources := projNpResponse.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources

			// Validate no duplications and correct resources
			Expect(len(finalProjResources)).To(Equal(2), "Project should have exactly 2 nodepool resources (no duplications)")

			projNodePoolIds := make(map[string]bool)
			projNodePoolNames := make(map[string]bool)
			for _, npr := range finalProjResources {
				// Check for duplicate nodepool IDs
				Expect(projNodePoolIds[npr.NodePoolId]).To(BeFalse(), fmt.Sprintf("Duplicate nodepool ID found in project resources: %s", npr.NodePoolId))
				projNodePoolIds[npr.NodePoolId] = true

				// Verify nodepool names
				projNodePoolNames[npr.NodePoolName] = true

				// Verify correct org unit association
				Expect(npr.OrgUnitId).To(Equal(testProject.Id))
				Expect(npr.OrgUnitName).To(Equal(testProject.Name))
				Expect(npr.ClusterId).To(Equal(testClusterId))
				Expect(npr.ParentId).To(Equal(testDepartment.Id))
			}

			// Verify we have both expected nodepools (default and the recreated one)
			Expect(projNodePoolIds[defaultNodePoolId]).To(BeTrue(), "Default nodepool should be present in project")
			Expect(projNodePoolIds[recreatedNodePoolId]).To(BeTrue(), "Recreated nodepool should be present in project")
			Expect(projNodePoolIds[originalNodePoolId]).To(BeFalse(), "Original (deleted) nodepool should NOT be present in project")

			// Step 6: Verify via cluster-wide query that there are no stale resources
			clusterDeptNpResponse, err := testDepartmentApiService.GetDepartmentNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", testClusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(clusterDeptNpResponse.Code).To(Equal(http.StatusOK))
			clusterDeptResources := clusterDeptNpResponse.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources

			// Count resources per department to ensure no duplications cluster-wide
			deptResourceCount := 0
			for _, npr := range clusterDeptResources {
				if npr.OrgUnitId == testDepartment.Id {
					deptResourceCount++
					Expect(npr.NodePoolId).ToNot(Equal(originalNodePoolId), "Original nodepool should not appear in cluster-wide department query")
				}
			}
			Expect(deptResourceCount).To(Equal(2), "Department should have exactly 2 resources in cluster-wide query")

			clusterProjNpResponse, err := testProjectApiService.GetProjectNodePoolResources(ctx, []string{fmt.Sprintf("clusterId==%s", testClusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(clusterProjNpResponse.Code).To(Equal(http.StatusOK))
			clusterProjResources := clusterProjNpResponse.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources

			// Count resources per project to ensure no duplications cluster-wide
			projResourceCount := 0
			for _, npr := range clusterProjResources {
				if npr.OrgUnitId == testProject.Id {
					projResourceCount++
					Expect(npr.NodePoolId).ToNot(Equal(originalNodePoolId), "Original nodepool should not appear in cluster-wide project query")
				}
			}
			Expect(projResourceCount).To(Equal(2), "Project should have exactly 2 resources in cluster-wide query")
		})
	})
})

func getProjectResourcesCreateRequest(name string, parentId string, clusterId string, resources []api.ResourcesNullable) api.ProjectCreationRequest {
	return api.ProjectCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		ParentId:  &parentId,
		Resources: resources,
	}
}

func getDepartmentResourcesCreateRequest(name string, clusterId string, resources []api.ResourcesNullable) api.DepartmentCreationRequest {
	return api.DepartmentCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		Resources: resources,
	}
}

func getParentDepartmentCreateRequest(clusterId string) api.DepartmentCreationRequest {
	return api.DepartmentCreationRequest{
		Name:      "default-p",
		ClusterId: clusterId,
		Resources: []api.ResourcesNullable{{
			Gpu: api.ResourcesNullableGpu{
				Deserved: ptr.To(float64(0)),
			},
		}},
	}
}
