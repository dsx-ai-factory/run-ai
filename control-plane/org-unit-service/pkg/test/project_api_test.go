/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	workloadsOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	workloadsMocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project API tests", Ordered, func() {
	clusterId := uuid.New().String()
	TestTenantId, _ := strconv.Atoi(TestTenant)
	ctx := SetContext(TestTenant, TestUser)

	var projectIds []string
	var departmentId string
	var nodeType1Id string
	var nodeType1Name string
	var nodeType2Id string
	var nodeType2Name string
	var projectApiService api.ProjectsAPIServicer
	var projectV1ApiService api.ProjectsV1APIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var nodeTypeService api.NodeTypesAPIServicer

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		TenantId: int32(TestTenantId),
		Uuid:     clusterId,
		Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting project tests")

		nodeTypeService = services.NewNodeTypeApiService()
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId = response.Body.(api.Department).Id
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
	})

	AfterAll(func() {
		_, err := departmentApiService.DeleteDepartment(ctx, departmentId)
		if err != nil {
			Fail("Failed deleting department")
		}
		_, err = nodeTypeService.DeleteNodeType(ctx, nodeType1Id)
		if err != nil {
			Fail("Failed deleting node type")
		}
		_, err = nodeTypeService.DeleteNodeType(ctx, nodeType2Id)
		if err != nil {
			Fail("Failed deleting node type")
		}
	})

	Context("test create/get projects", func() {
		It("test create/get one project no override", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			request.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			request.EnforceRunaiScheduler = ptr.To(true)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.CreatedAt).ToNot(BeNil())
			Expect(created.UpdatedAt).To(Equal(created.CreatedAt))
			Expect(created.Name).To(Equal(request.Name))
			Expect(created.Description).To(Equal(fmt.Sprintf("%s project description", name)))
			Expect(*created.RequestedNamespace).To(Equal(request.RequestedNamespace))
			Expect(created.EnforceRunaiScheduler).To(Equal(true))
			Expect(created.ClusterId).To(Equal(request.ClusterId))
			Expect(created.ParentId).To(Equal(departmentId))
			Expect(created.DefaultNodePools).To(Not(BeNil()))
			Expect(len(*created.DefaultNodePools)).To(Equal(3))
			Expect((*created.DefaultNodePools)[0]).To(Equal("pool2"))
			Expect((*created.DefaultNodePools)[1]).To(Equal("pool1"))
			Expect((*created.DefaultNodePools)[2]).To(Equal("pool3"))
			Expect(created.NodeTypes).To(Not(BeNil()))
			Expect(created.NodeTypes.Workspace).To(BeNil())
			Expect(created.NodeTypes.Training).To(Not(BeNil()))
			Expect(len(created.NodeTypes.Training)).To(Equal(1))
			Expect(created.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(created.NodeTypes.Names).NotTo(BeNil())
			Expect(len(*created.NodeTypes.Names)).To(Equal(1))
			createdNodeTypeName, createdFound := (*created.NodeTypes.Names)[nodeType1Id]
			Expect(createdFound).To(Equal(true))
			Expect(createdNodeTypeName).To(Equal(nodeType1Name))
			Expect(created.SchedulingRules).To(BeNil())
			Expect(len(created.Resources)).To(Equal(1))
			resources := created.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(ptr.Deref(request.Resources[0].Gpu.Deserved, 0)))
			Expect(resources.Gpu.Limit).To(Equal(request.Resources[0].Gpu.Limit))
			Expect(resources.Gpu.OverQuotaWeight).To(Equal(request.Resources[0].Gpu.OverQuotaWeight))
			Expect(resources.Cpu).To(BeNil())
			Expect(resources.Memory).To(BeNil())
			Expect(created.CreatedBy).To(Equal(TestUser))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			Expect(getProject.RequestedNamespace).To(Equal(created.RequestedNamespace))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(true))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.Parent.Id).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.Description).To(Equal(fmt.Sprintf("%s project description", name)))
			Expect(getProject.SchedulingRules).To(Equal(created.SchedulingRules))
			Expect(getProject.Resources).To(Equal(created.Resources))
			Expect(getProject.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(getProject.UpdatedAt.UnixMilli()).To(Equal(created.UpdatedAt.UnixMilli()))
			Expect(getProject.DefaultNodePools).To(Not(BeNil()))
			Expect(len(*getProject.DefaultNodePools)).To(Equal(3))
			Expect((*getProject.DefaultNodePools)[0]).To(Equal("pool2"))
			Expect((*getProject.DefaultNodePools)[1]).To(Equal("pool1"))
			Expect((*getProject.DefaultNodePools)[2]).To(Equal("pool3"))
			Expect(getProject.NodeTypes).To(Not(BeNil()))
			Expect(getProject.NodeTypes.Workspace).To(BeNil())
			Expect(getProject.NodeTypes.Training).To(Not(BeNil()))
			Expect(len(getProject.NodeTypes.Training)).To(Equal(1))
			Expect(getProject.NodeTypes.Training[0]).To(Equal(nodeType1Id))
			Expect(getProject.NodeTypes.Names).NotTo(BeNil())
			Expect(len(*getProject.NodeTypes.Names)).To(Equal(1))
			nodeTypeName, found := (*getProject.NodeTypes.Names)[nodeType1Id]
			Expect(found).To(Equal(true))
			Expect(nodeTypeName).To(Equal(nodeType1Name))

			ancestorsResponse, err := projectApiService.GetProjectAncestors(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project ancestors")
			}
			ancestors := ancestorsResponse.Body.(authorization_utils.Ancestors)
			Expect(ancestors).To(Not(BeNil()))
			Expect(ancestors.TenantId).To(Equal(TestTenant))
			Expect(ancestors.ClusterUuid).To(Not(BeNil()))
			Expect(*ancestors.ClusterUuid).To(Equal(created.ClusterId))
			Expect(ancestors.DepartmentId).To(Not(BeNil()))
			Expect(*ancestors.DepartmentId).To(Equal(created.ParentId))
			Expect(ancestors.ProjectId).To(BeNil())
		})

		It("test create/get one project 0 deserved", func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources = []api.ResourcesNullable{{
				Gpu: api.ResourcesNullableGpu{
					Deserved: ptr.To(float64(0)),
				},
				Rank: ptr.To("VeryHigh"),
			}}
			request.EnforceRunaiScheduler = ptr.To(false)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.CreatedAt).ToNot(BeNil())
			Expect(created.UpdatedAt).To(Equal(created.CreatedAt))
			Expect(created.Name).To(Equal(request.Name))
			Expect(*created.RequestedNamespace).To(Equal(request.RequestedNamespace))
			Expect(created.EnforceRunaiScheduler).To(Equal(false))
			Expect(created.ClusterId).To(Equal(request.ClusterId))
			Expect(created.ParentId).To(Equal(departmentId))
			Expect(created.DefaultNodePools).To(BeNil())
			Expect(len(created.Resources)).To(Equal(1))
			resources := created.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(float64(0)))
			Expect(resources.Gpu.Limit).To(BeNil())
			// adjustable weights off + deserved 0 → response floors over-quota-weight to MinOverQuotaWeight
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(resources.Cpu).To(BeNil())
			Expect(resources.Memory).To(BeNil())
			Expect(created.CreatedBy).To(Equal(TestUser))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			Expect(getProject.RequestedNamespace).To(Equal(created.RequestedNamespace))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(false))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.Parent.Id).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.SchedulingRules).To(Equal(created.SchedulingRules))
			Expect(len(getProject.Resources)).To(Equal(1))
			resources = getProject.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(float64(0)))
			Expect(resources.Gpu.Limit).To(BeNil())
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(resources.Cpu).To(BeNil())
			Expect(resources.Memory).To(BeNil())
			Expect(getProject.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(getProject.UpdatedAt.UnixMilli()).To(Equal(created.UpdatedAt.UnixMilli()))
			Expect(resources.Rank).To(Equal(ptr.To("VeryHigh")))
			// Expect(getProject.DefaultNodePools).To(Not(BeNil()))
		})

		It("with adjustable weights off and deserved=0, every project route returns MinOverQuotaWeight for GPU, CPU, and Memory", func() {
			// Reset cluster mock to cpuQuota=true, adjustable weights=false so CPU/Memory are validated.
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			assertAllResourcesFloored := func(r api.Resources) {
				Expect(r.Gpu.Deserved).To(Equal(float64(0)))
				Expect(r.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*r.Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
				Expect(r.Cpu).ToNot(BeNil())
				Expect(*r.Cpu.Deserved).To(Equal(float64(0)))
				Expect(*r.Cpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
				Expect(r.Memory).ToNot(BeNil())
				Expect(*r.Memory.Deserved).To(Equal(float64(0)))
				Expect(*r.Memory.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			}

			zeroResources := []api.ResourcesNullable{{
				NodePool: &api.ResourcesNodePool{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName},
				Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(float64(0)), Limit: ptr.To(float64(-1))},
				Cpu:      &api.ResourcesCpu{Deserved: ptr.To(float64(0)), Limit: ptr.To(float64(-1))},
				Memory:   &api.ResourcesMemory{Deserved: ptr.To(float64(0)), Limit: ptr.To(float64(-1)), Units: api.UNITS_MIB},
			}}

			// POST /projects
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources = zeroResources
			request.EnforceRunaiScheduler = ptr.To(false)
			createResp, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))
			created := createResp.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			assertAllResourcesFloored(created.Resources[0])

			// GET /projects/{id}
			getResp, err := projectApiService.GetProject(ctx, created.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(getResp.Body.(api.Project).Resources[0])

			// GET /projects (list)
			listResp, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			listed := listResp.Body.(api.GetProjects200Response).Projects
			var listedProject *api.Project
			for i := range listed {
				if listed[i].Id == created.Id {
					listedProject = &listed[i]
				}
			}
			Expect(listedProject).ToNot(BeNil())
			assertAllResourcesFloored(listedProject.Resources[0])

			// PUT /projects/{id} — round-trip the GET response; validation must accept MinOverQuotaWeight when deserved=0
			updateReq := getProjectUpdateRequest(*listedProject)
			updateResp, err := projectApiService.UpdateProject(ctx, created.Id, updateReq)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(updateResp.Body.(api.Project).Resources[0])

			// PUT /projects/{id}/resources
			putResourcesResp, err := projectApiService.UpdateProjectResources(ctx, created.Id, zeroResources)
			Expect(err).To(BeNil())
			Expect(putResourcesResp.Code).To(Equal(http.StatusOK))
			updatedResources := putResourcesResp.Body.([]api.Resources)
			assertAllResourcesFloored(updatedResources[0])

			// PATCH /projects/{id}/resources — round-trip the previous response
			patchResp, err := projectApiService.PatchProjectResources(ctx, created.Id, getProjectUpdateRequest(updateResp.Body.(api.Project)).Resources)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(patchResp.Body.([]api.Resources)[0])

			// GET /sync/projects — cluster-scheduler feed
			syncResp, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncProjects := syncResp.Body.(api.GetSyncProjects200Response).Projects
			var syncedProj *api.ProjectSync
			for i := range syncProjects {
				if syncProjects[i].Id == created.Id {
					syncedProj = &syncProjects[i]
				}
			}
			Expect(syncedProj).ToNot(BeNil())
			Expect(*syncedProj.Resources[0].Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(syncedProj.Resources[0].Cpu).ToNot(BeNil())
			Expect(*syncedProj.Resources[0].Cpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(syncedProj.Resources[0].Memory).ToNot(BeNil())
			Expect(*syncedProj.Resources[0].Memory.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
		})

		It("test create project with version below 2.20 with not Normal rank", func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.19")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources = []api.ResourcesNullable{{
				Rank: ptr.To("VeryHigh"),
			}}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		Describe("", Ordered, func() {
			nodePoolName1 := "nodepool-a"
			nodePoolId1 := float32(2)

			BeforeAll(func() {
				clusterServiceClient = mocks.IClusterClient{}
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TestTenantId),
					Uuid:     clusterId,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.19")),
				}, nil)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
						{
							Name:                   nodePoolName1,
							Id:                     fmt.Sprint(nodePoolId1),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
					},
				}, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			})

			AfterAll(func() {
				clusterServiceClient = mocks.IClusterClient{}
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TestTenantId),
					Uuid:     clusterId,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.19")),
				}, nil)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
					},
				}, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			})

			It("test create/get one project total quota values", func() {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, departmentId, clusterId)
				request.Resources[0].Gpu = api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(-1)),
					OverQuotaWeight: ptr.To(float64(1)),
				}
				request.Resources[0].Cpu = &api.ResourcesCpu{
					Deserved:        ptr.To(float64(-1)),
					Limit:           ptr.To(float64(-1)),
					OverQuotaWeight: ptr.To(float64(-1)),
				}
				request.Resources[0].Memory = &api.ResourcesMemory{
					Deserved:        nil,
					Limit:           nil,
					OverQuotaWeight: nil,
				}
				request.Resources = append(request.Resources, api.ResourcesNullable{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(2)),
						Limit:           ptr.To(float64(2)),
						OverQuotaWeight: ptr.To(float64(2)),
					},
					Cpu: &api.ResourcesCpu{
						Deserved:        ptr.To(float64(1)),
						Limit:           ptr.To(float64(1)),
						OverQuotaWeight: ptr.To(float64(1)),
					},
					Memory: &api.ResourcesMemory{
						Deserved:        nil,
						Limit:           nil,
						OverQuotaWeight: nil,
					},
				})
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				proj := createResponse.Body.(api.Project)

				projectIds = append(projectIds, proj.Id)

				getResponse, err := projectApiService.GetProject(ctx, proj.Id)
				if err != nil {
					Fail("Failed getting project")
				}
				proj = getResponse.Body.(api.Project)
				Expect(proj.Id).To(Equal(proj.Id))
				Expect(proj.Name).To(Equal(proj.Name))
				Expect(proj.RequestedNamespace).To(Equal(proj.RequestedNamespace))
				Expect(proj.ParentId).To(Equal(proj.ParentId))
				Expect(proj.Parent.Id).To(Equal(proj.ParentId))
				Expect(proj.ClusterId).To(Equal(proj.ClusterId))
				Expect(proj.Description).To(Equal(fmt.Sprintf("%s project description", name)))
				Expect(proj.SchedulingRules).To(Equal(proj.SchedulingRules))
				Expect(proj.Resources).To(Equal(proj.Resources))
				Expect(*proj.TotalResources.GpuQuota).To(Equal(float64(3)))
				Expect(*proj.TotalResources.CpuQuota).To(Equal(float64(-1)))
				Expect(proj.TotalResources.MemoryQuota).To(BeNil())
				Expect(proj.CreatedAt.UnixMilli()).To(Equal(proj.CreatedAt.UnixMilli()))
				Expect(proj.UpdatedAt.UnixMilli()).To(Equal(proj.UpdatedAt.UnixMilli()))
			})

			It("test create/get one project total quota values different units", func() {
				nodePoolName1 := "nodepool-a"
				nodePoolId1 := float32(2)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
				name := createProjectName()
				request := getProjectCreateRequest(name, name, departmentId, clusterId)
				request.Resources[0].Gpu = api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(-1)),
					OverQuotaWeight: ptr.To(float64(1)),
				}
				request.Resources[0].Cpu = &api.ResourcesCpu{
					Deserved:        ptr.To(float64(-1)),
					Limit:           ptr.To(float64(-1)),
					OverQuotaWeight: ptr.To(float64(-1)),
				}
				request.Resources[0].Memory = &api.ResourcesMemory{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(1)),
					OverQuotaWeight: ptr.To(float64(1)),
					Units:           "Mib",
				}
				request.Resources = append(request.Resources, api.ResourcesNullable{
					NodePool: &api.ResourcesNodePool{
						Id:   fmt.Sprint(nodePoolId1),
						Name: nodePoolName1,
					},
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(2)),
						Limit:           ptr.To(float64(2)),
						OverQuotaWeight: ptr.To(float64(2)),
					},
					Cpu: &api.ResourcesCpu{
						Deserved:        ptr.To(float64(1)),
						Limit:           ptr.To(float64(1)),
						OverQuotaWeight: ptr.To(float64(1)),
					},
					Memory: &api.ResourcesMemory{
						Deserved:        ptr.To(float64(1)),
						Limit:           ptr.To(float64(1)),
						OverQuotaWeight: ptr.To(float64(1)),
						Units:           "GB",
					},
				})
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				proj := createResponse.Body.(api.Project)

				projectIds = append(projectIds, proj.Id)

				getResponse, err := projectApiService.GetProject(ctx, proj.Id)
				if err != nil {
					Fail("Failed getting project")
				}
				proj = getResponse.Body.(api.Project)
				Expect(proj.Name).To(Equal(request.Name))
				Expect(proj.RequestedNamespace).ToNot(BeNil())
				Expect(*proj.RequestedNamespace).To(Equal(request.RequestedNamespace))
				Expect(proj.ParentId).To(Equal(*request.ParentId))
				Expect(proj.Parent.Id).To(Equal(*request.ParentId))
				Expect(proj.ClusterId).To(Equal(request.ClusterId))
				Expect(proj.Description).To(Equal(fmt.Sprintf("%s project description", name)))
				Expect(proj.SchedulingRules).To(Equal(request.SchedulingRules))

				defaultNpResources := api.Resources{}
				otherNpResources := api.Resources{}
				for _, r := range proj.Resources {
					if r.NodePool.Id == fmt.Sprint(defaultNodePoolId) {
						defaultNpResources = r
					} else {
						otherNpResources = r
					}
				}

				validateResourcesCpu(defaultNpResources, request.Resources[0], false)
				validateResourcesCpu(otherNpResources, request.Resources[1], false)
				Expect(*proj.TotalResources.GpuQuota).To(Equal(float64(3)))
				Expect(*proj.TotalResources.CpuQuota).To(Equal(float64(-1)))
				Expect(*proj.TotalResources.MemoryQuota).To(Equal(float64(1001.049)))
			})

			It("test create/get one project with only 1 nodepool resources even though we have 2 nodepools in the system", func() {
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
				name := createProjectName()
				request := getProjectCreateRequest(name, name, departmentId, clusterId)
				request.Resources[0].NodePool = nil

				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				proj := createResponse.Body.(api.Project)

				projectIds = append(projectIds, proj.Id)

				getResponse, err := projectApiService.GetProject(ctx, proj.Id)
				if err != nil {
					Fail("Failed getting project")
				}
				proj = getResponse.Body.(api.Project)
				Expect(proj.Id).To(Equal(proj.Id))
				Expect(proj.Name).To(Equal(request.Name))
				Expect(len(proj.Resources)).To(Equal(2))
			})
		})

		It("test create/get projects with override", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			desc := "first project"
			request.Description = desc
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			created1 := response.Body.(api.Project)
			Expect(created1.CreatedAt).ToNot(BeNil())
			Expect(created1.UpdatedAt).To(Equal(created1.CreatedAt))
			Expect(created1.Name).To(Equal(request.Name))
			Expect(created1.Description).To(Equal(desc))
			Expect(*created1.RequestedNamespace).To(Equal(request.RequestedNamespace))
			Expect(created1.ClusterId).To(Equal(request.ClusterId))
			Expect(created1.ParentId).To(Equal(departmentId))
			// Expect(created1.DefaultNodePools).ToNot(BeNil())
			// Expect(*created1.DefaultNodePools).To(Equal([]string{"default"}))
			Expect(created1.NodeTypes).To(Not(BeNil()))
			Expect(created1.NodeTypes.Workspace).To(BeNil())
			Expect(created1.NodeTypes.Training).To(BeNil())
			jobConf := created1.SchedulingRules
			Expect(*jobConf.TrainingJobMaxIdleDurationSeconds).To(Equal(int32(1000)))
			Expect(*jobConf.InteractiveJobPreemptIdleDurationSeconds).To(Equal(int32(1000)))
			Expect(*jobConf.InteractiveJobTimeLimitSeconds).To(Equal(int32(1000)))
			Expect(*jobConf.InteractiveJobMaxIdleDurationSeconds).To(Equal(int32(1000)))
			Expect(len(created1.Resources)).To(Equal(1))
			resources := created1.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(ptr.Deref(request.Resources[0].Gpu.Deserved, 0)))
			Expect(resources.Gpu.Limit).To(Equal(request.Resources[0].Gpu.Limit))
			Expect(resources.Gpu.OverQuotaWeight).To(Equal(request.Resources[0].Gpu.OverQuotaWeight))
			Expect(resources.Cpu).To(Equal(request.Resources[0].Cpu))
			Expect(resources.Memory).To(Equal(request.Resources[0].Memory))

			name2 := createProjectName()
			request = getFullProjectCreateRequest(name2, name2, &departmentId, clusterId)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			created2 := response.Body.(api.Project)

			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 10, "")
			if err != nil {
				Fail("Failed getting projects")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProjects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(getProjects)).To(Equal(2))
			sort.Slice(getProjects, func(i, j int) bool {
				return getProjects[i].CreatedAt.Before(getProjects[j].CreatedAt)
			})
			Expect(getProjects[0].Id).To(Equal(created1.Id))
			Expect(getProjects[0].Description).To(Equal(desc))
			Expect(getProjects[1].Id).To(Equal(created2.Id))

			projectIds = append(projectIds, created2.Id)
			projectIds = append(projectIds, created1.Id)
		})

		It("test pagination", func() {
			name1 := "a-" + createProjectName()
			request := getFullProjectCreateRequest(name1, name1, &departmentId, clusterId)
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			created1 := response.Body.(api.Project)

			name2 := "b-" + createProjectName()
			request = getFullProjectCreateRequest(name2, name2, &departmentId, clusterId)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			created2 := response.Body.(api.Project)
			projectIds = append(projectIds, created2.Id)
			projectIds = append(projectIds, created1.Id)

			// test pagination
			getResponse, err := projectApiService.GetProjects(ctx, []string{}, "name", "", 0, 1, "")
			if err != nil {
				Fail("Failed getting projects")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProjects := getResponse.Body.(api.GetProjects200Response).Projects
			next := getResponse.Body.(api.GetProjects200Response).Next
			Expect(len(getProjects)).To(Equal(1))
			Expect(getProjects[0].Id).To(Equal(created1.Id))
			Expect(next).To(Equal(int32(1)))

			getResponse, err = projectApiService.GetProjects(ctx, []string{}, "name", "", 1, 1, "")
			if err != nil {
				Fail("Failed getting projects")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProjects = getResponse.Body.(api.GetProjects200Response).Projects
			next = getResponse.Body.(api.GetProjects200Response).Next
			Expect(len(getProjects)).To(Equal(1))
			Expect(getProjects[0].Id).To(Equal(created2.Id))
			Expect(next).To(Equal(int32(2)))
		})

		It("test create/get projects with override job limit good request", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			timeLimit := int32(300)
			request.SchedulingRules.InteractiveJobTimeLimitSeconds = &timeLimit
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			projectIds = append(projectIds, response.Body.(api.Project).Id)
		})

		//It("test create/get projects with override job limit bad request", func() {
		//	name := createProjectName()
		//	request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
		//	timeLimit := int32(1300)
		//	request.SchedulingRules.InteractiveJobTimeLimitSeconds = &timeLimit
		//	response, err := projectApiService.CreateProject(ctx, request)
		//	if err != nil {
		//		Fail("Failed creating project")
		//	}
		//	Expect(err).To(BeNil())
		//	Expect(response.Code).To(Equal(http.StatusBadRequest))
		//})

		It("test get non existing project", func() {
			getResponse, err := projectApiService.GetProject(ctx, "444")
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test create project and delete node type", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			nodeTypeRequest := api.NodeTypeCreateRequest{Name: "nodeTypeTest-" + uuid.New().String(), ClusterId: clusterId}
			nodeTypeResponse, err := nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
			nodeTypeId := nodeTypeResponse.Body.(api.NodeType).Id

			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			request.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeTypeId},
				Workspace: nil,
			}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.CreatedAt).ToNot(BeNil())
			Expect(created.UpdatedAt).To(Equal(created.CreatedAt))
			Expect(created.Name).To(Equal(request.Name))
			Expect(created.NodeTypes).To(Not(BeNil()))
			Expect(created.NodeTypes.Workspace).To(BeNil())
			Expect(created.NodeTypes.Training).To(Not(BeNil()))
			Expect(len(created.NodeTypes.Training)).To(Equal(1))
			Expect(created.NodeTypes.Training[0]).To(Equal(nodeTypeId))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			Expect(getProject.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(getProject.UpdatedAt.UnixMilli()).To(Equal(created.UpdatedAt.UnixMilli()))
			Expect(getProject.NodeTypes).To(Not(BeNil()))
			Expect(getProject.NodeTypes.Workspace).To(BeNil())
			Expect(getProject.NodeTypes.Training).To(Not(BeNil()))
			Expect(len(created.NodeTypes.Training)).To(Equal(1))
			Expect(created.NodeTypes.Training[0]).To(Equal(nodeTypeId))

			_, err = nodeTypeService.DeleteNodeType(ctx, nodeTypeId)
			if err != nil {
				Fail("Failed deleting node type")
			}

			getResponse, err = projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject = getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.Name).To(Equal(created.Name))
			Expect(getProject.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(getProject.UpdatedAt.UnixMilli()).To(Equal(created.UpdatedAt.UnixMilli()))
			Expect(getProject.NodeTypes).To(Not(BeNil()))
			Expect(getProject.NodeTypes.Workspace).To(BeNil())
			Expect(getProject.NodeTypes.Training).To(BeNil())
		})

		It("test create project that starts with number", func() {
			// can start with number
			name := "13p"
			request := getFullProjectCreateRequest(name, "", &departmentId, clusterId)
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			// cannot end with hyphen
			name = "13p-"
			request = getFullProjectCreateRequest(name, "", &departmentId, clusterId)
			response, err = projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			// cannot be only numbers
			name = "1337"
			request = getFullProjectCreateRequest(name, "", &departmentId, clusterId)
			response, err = projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("test create/get one project without setting priority", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources[0].Rank = nil

			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			validateResourcesArrayTo(getProject.Resources, created.Resources)

			_, deleteError := projectApiService.DeleteProject(ctx, getProject.Id, "", false)
			Expect(deleteError).NotTo(HaveOccurred())
		})

		It("test create project with existing name", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			response, err = projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusConflict))
		})

		It("test create forbidden", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			ctx2 := SetContext("2", TestUser)
			response, err := projectApiService.CreateProject(ctx2, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("test project with another nodepool", func() {
		var (
			nodePoolName1         = "nodepool-a"
			nodePoolId1           = float32(2)
			departmentApiService2 api.DepartmentsAPIServicer
			projectApiService2    api.ProjectsAPIServicer
			dep0Id                string
		)

		BeforeAll(func() {
			// department+projects service with 2 nodepools
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.19")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			departmentApiService2 = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService2 = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			createDepartmentRequest := getFullDepartmentCreateRequest("dep-0", clusterId)
			createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, CreateNodePoolApiResources(true, "2", nodePoolName1))
			response, err := departmentApiService2.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			dep0Id = response.Body.(api.Department).Id
		})

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.19")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			// delete department dep0Id
			_, err := departmentApiService2.DeleteDepartment(ctx, dep0Id)
			if err != nil {
				Fail("Failed deleting parent department")
			}
		})

		It("test create without nodepool id - error", func() {
			name := createProjectName()
			request := getProjectCreateRequestWithNodePools(name, "", dep0Id, clusterId, nodePoolName1)
			id0 := request.Resources[0].NodePool.Id
			name0 := request.Resources[0].NodePool.Name
			request.Resources[0].NodePool.Id = ""
			request.Resources[0].NodePool.Name = ""
			createResponse, err := projectApiService2.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))

			request.Resources[0].NodePool.Id = id0
			request.Resources[0].NodePool.Name = name0
			request.Resources[1].NodePool.Id = ""
			request.Resources[1].NodePool.Name = ""
			createResponse, err = projectApiService2.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test create without nodepoolname - returned with name in GET", func() {
			name := createProjectName()
			request := getProjectCreateRequestWithNodePools(name, "", dep0Id, clusterId, nodePoolName1)
			request.Resources[0].NodePool.Name = ""
			request.Resources[1].NodePool.Name = ""
			createResponse, err := projectApiService2.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			createdProjectId := createResponse.Body.(api.Project).Id
			projectIds = append(projectIds, createdProjectId)

			getResponse, err := projectApiService2.GetProject(ctx, createdProjectId)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)

			for _, resources := range getProject.Resources {
				Expect(resources.NodePool).To(Not(BeNil()))
				if resources.NodePool.Id == "2" {
					Expect(resources.NodePool.Name).To(Equal(nodePoolName1))
				} else {
					Expect(resources.NodePool.Name).To(Equal(defaultNodePoolName))
				}
			}
		})
	})

	Context("test update projects", func() {
		It("test update project ok status", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Description = ""
			request.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			request.EnforceRunaiScheduler = ptr.To(true)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)
			timeLimits := int32(1000)
			updateRequest.SchedulingRules = &api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: &timeLimits,
			}
			updateRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  nil,
				Workspace: []string{nodeType2Id},
			}
			updateRequest.EnforceRunaiScheduler = ptr.To(false)

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.EnforceRunaiScheduler).To(Equal(false))
			Expect(updated.Description).To(Equal(""))
			Expect(updated.SchedulingRules).To(Equal(&api.SchedulingRules{
				InteractiveJobTimeLimitSeconds: &timeLimits,
			}))
			Expect(updated.NodeTypes).To(Not(BeNil()))
			Expect(updated.NodeTypes.Training).To(BeNil())
			Expect(updated.NodeTypes.Workspace).To(Not(BeNil()))
			Expect(len(updated.NodeTypes.Workspace)).To(Equal(1))
			Expect(updated.NodeTypes.Workspace[0]).To(Equal(nodeType2Id))
			Expect(updated.NodeTypes.Names).NotTo(BeNil())
			Expect(len(*updated.NodeTypes.Names)).To(Equal(1))
			updatedNodeTypeName, updatedFound := (*updated.NodeTypes.Names)[nodeType2Id]
			Expect(updatedFound).To(Equal(true))
			Expect(updatedNodeTypeName).To(Equal(nodeType2Name))
			Expect(updated.Resources).To(Equal(created.Resources))
			Expect(updated.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(updated.UpdatedAt.UnixMilli()).NotTo(Equal(created.UpdatedAt.UnixMilli()))
			Expect(string(updated.Status.Phase)).To(Equal("Updating"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(updated.Id))
			Expect(getProject.ParentId).To(Equal(updated.ParentId))
			Expect(getProject.ClusterId).To(Equal(updated.ClusterId))
			Expect(getProject.Description).To(Equal(""))
			Expect(getProject.SchedulingRules).To(Equal(updated.SchedulingRules))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(false))
			Expect(getProject.Resources).To(Equal(updated.Resources))
			Expect(getProject.CreatedAt.UnixMilli()).To(Equal(updated.CreatedAt.UnixMilli()))
			Expect(getProject.UpdatedAt.UnixMilli()).To(Equal(updated.UpdatedAt.UnixMilli()))
			Expect(getProject.NodeTypes).To(Not(BeNil()))
			Expect(getProject.NodeTypes.Training).To(BeNil())
			Expect(getProject.NodeTypes.Workspace).To(Not(BeNil()))
			Expect(len(getProject.NodeTypes.Workspace)).To(Equal(1))
			Expect(getProject.NodeTypes.Workspace[0]).To(Equal(nodeType2Id))
			Expect(getProject.NodeTypes.Names).NotTo(BeNil())
			Expect(len(*getProject.NodeTypes.Names)).To(Equal(1))
			nodeTypeName, found := (*getProject.NodeTypes.Names)[nodeType2Id]
			Expect(found).To(Equal(true))
			Expect(nodeTypeName).To(Equal(nodeType2Name))
			Expect(string(getProject.Status.Phase)).To(Equal("Updating"))
		})

		It("test update project no change", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest := getProjectUpdateRequest(created)

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(updated.Status.Phase)).To(Equal("Ready"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(getProject.Status.Phase)).To(Equal("Ready"))
		})

		It("test update project no change with cpu", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved: ptr.To(9536.7431640625),
				Limit:    nil,
				Units:    api.UNITS_MIB,
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated), fmt.Sprintf("Error creating project: %v", createResponse.Body))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved: ptr.To(9536.7431640625),
				Limit:    nil,
				Units:    api.UNITS_MIB,
			})

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK), fmt.Sprintf("Error updating project: %v", updateResponse.Body))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(updated.Status.Phase)).To(Equal("Ready"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(getProject.Status.Phase)).To(Equal("Ready"))
		})

		It("test update project no change different memory units", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved: ptr.To(float64(200)),
				Limit:    ptr.To(float64(200)),
				Units:    api.UNITS_GB,
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated), fmt.Sprintf("Error creating project: %v", createResponse.Body))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest := getProjectUpdateRequest(created)
			// same but in MB instead of GB
			updateRequest.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved: ptr.To(float64(200000)),
				Limit:    ptr.To(float64(200000)),
				Units:    api.UNITS_MB,
			})

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK), fmt.Sprintf("Error updating project: %v", updateResponse.Body))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(updated.Status.Phase)).To(Equal("Ready"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.EnforceRunaiScheduler).To(Equal(true))
			Expect(string(getProject.Status.Phase)).To(Equal("Ready"))
		})

		It("test update project with change different memory units", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved:        ptr.To(float64(2)),
				Limit:           ptr.To(float64(-1)),
				OverQuotaWeight: ptr.To(float64(2)),
				Units:           api.UNITS_GB,
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Resources[0].Memory = ptr.To(api.ResourcesMemory{
				Deserved:        ptr.To(float64(2000)),
				Limit:           ptr.To(float64(-1)),
				OverQuotaWeight: ptr.To(float64(2000)),
				Units:           api.UNITS_MB,
			})

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(string(updated.Status.Phase)).NotTo(Equal("Updating"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(string(getProject.Status.Phase)).NotTo(Equal("Updating"))
		})

		It("test update project with change", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training: []string{nodeType1Id},
			}

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(string(updated.Status.Phase)).To(Equal("Updating"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(string(getProject.Status.Phase)).To(Equal("Updating"))
		})

		It("test update project description", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Description = ""
			request.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			request.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Description = "new description"
			updateRequest.NodeTypes = api.NodeTypesPerWorkload{
				Training:  []string{nodeType1Id},
				Workspace: nil,
			}
			updateRequest.DefaultNodePools = &[]string{"pool2", "pool1", "pool3"}

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.Description).To(Equal("new description"))
			Expect(string(updated.Status.Phase)).NotTo(Equal("Updating"))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)
			Expect(getProject.Id).To(Equal(created.Id))
			Expect(getProject.ParentId).To(Equal(created.ParentId))
			Expect(getProject.ClusterId).To(Equal(created.ClusterId))
			Expect(getProject.Description).To(Equal("new description"))
			Expect(string(getProject.Status.Phase)).NotTo(Equal("Updating"))
		})

		It("test update project not found", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)

			updateResponse, err := projectApiService.UpdateProject(ctx, "444", updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test update project with change - more fields", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, "", departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.SchedulingRules = &api.SchedulingRules{
				TrainingJobMaxIdleDurationSeconds: ptr.To(int32(1000)),
			}

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(string(updated.Status.Phase)).To(Equal("Updating"))

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest = getProjectUpdateRequest(created)
			updateRequest.SchedulingRules = &api.SchedulingRules{
				InteractiveJobPreemptIdleDurationSeconds: ptr.To(int32(1000)),
			}

			updateResponse, err = projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated = updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(string(updated.Status.Phase)).To(Equal("Updating"))

			updateStatusRequest = api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err = projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest = getProjectUpdateRequest(created)
			updateRequest.SchedulingRules = &api.SchedulingRules{
				InteractiveJobMaxIdleDurationSeconds: ptr.To(int32(1000)),
			}

			updateResponse, err = projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated = updateResponse.Body.(api.Project)
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(string(updated.Status.Phase)).To(Equal("Updating"))
		})

		It("test update project response contains full status", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)

			// Set full status via UpdateProjectStatus
			quotaStatus := api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 8000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 4000,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    3,
					Cpu:    6,
					Memory: 12000,
				},
			}
			statusNamespace := fmt.Sprintf("runai-%s", name)
			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase:        "Ready",
					PhaseMessage: ptr.To("project is ready"),
					Namespace:    &statusNamespace,
					QuotaStatus:  quotaStatus,
					AdditionalStatusData: &map[string]interface{}{
						"conditions": []models.Condition{
							{
								Type:               "NamespaceReady",
								Status:             "True",
								Message:            "namespace is ready",
								Reason:             "ProjectNamespace",
								LastTransitionTime: ptr.To(time.Date(2024, 2, 15, 14, 58, 19, 0, time.UTC)),
							},
						},
					},
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			// Now perform an update on the project
			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Description = "updated description"
			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)

			// Verify the update response contains the full status - same as what GET returns
			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			getProject := getResponse.Body.(api.Project)

			// Status fields should match between update response and GET response
			Expect(updated.Status.Namespace).To(Equal(getProject.Status.Namespace))
			Expect(updated.Status.PhaseMessage).To(Equal(getProject.Status.PhaseMessage))
			Expect(updated.Status.QuotaStatus).To(Equal(getProject.Status.QuotaStatus))
			Expect(updated.Status.AdditionalStatusData).To(Equal(getProject.Status.AdditionalStatusData))
			Expect(updated.Status.NodePoolQuotaStatuses).To(Equal(getProject.Status.NodePoolQuotaStatuses))

			// Verify specific status values are present
			Expect(updated.Status.Namespace).NotTo(BeNil())
			Expect(*updated.Status.Namespace).To(Equal(statusNamespace))
			Expect(updated.Status.QuotaStatus.Allocated.Gpu).To(Equal(float64(2)))
			Expect(updated.Status.QuotaStatus.Allocated.Cpu).To(Equal(float64(4)))
			Expect(updated.Status.QuotaStatus.Allocated.Memory).To(Equal(float64(8000)))
			Expect(updated.Status.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(1)))
			Expect(updated.Status.AdditionalStatusData).NotTo(BeNil())
		})

		It("test update project keeps node pool quota statuses", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			createResponse, err := projectApiService.CreateProject(ctx, getProjectCreateRequest(name, name, departmentId, clusterId))
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)

			nodePool := created.Resources[0].NodePool
			nodePoolQuotaStatuses := []api.QuotaStatusNodePool{{
				NodePoolName:            nodePool.Name,
				NodePoolId:              nodePool.Id,
				Allocated:               api.QuotaStatusResource{Gpu: 2, Cpu: 4, Memory: 8000},
				AllocatedNonPreemptible: api.QuotaStatusResource{Gpu: 1, Cpu: 2, Memory: 4000},
				Requested:               api.QuotaStatusResource{Gpu: 3, Cpu: 6, Memory: 12000},
			}}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase:                 "Ready",
					Namespace:             ptr.To(fmt.Sprintf("runai-%s", name)),
					NodePoolQuotaStatuses: nodePoolQuotaStatuses,
				},
			})
			Expect(err).To(BeNil())
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Resources[0].Gpu.Limit = ptr.To(float64(4))
			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Project)
			Expect(*updated.Resources[0].Gpu.Limit).To(Equal(float64(4)))
			Expect(updated.Status.NodePoolQuotaStatuses).To(Equal(nodePoolQuotaStatuses))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Body.(api.Project).Status.NodePoolQuotaStatuses).To(Equal(nodePoolQuotaStatuses))
		})

		It("test update project does not overwrite a status synced during the update", func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			name := createProjectName()
			createResponse, err := projectApiService.CreateProject(ctx, getProjectCreateRequest(name, name, departmentId, clusterId))
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)

			nodePool := created.Resources[0].NodePool
			setNonPreemptibleGpu := func(gpu float64) {
				response, err := projectApiService.UpdateProjectStatus(ctx, api.ProjectStatusUpdateRequest{
					Id:        created.Id,
					ClusterId: clusterId,
					Status: api.ProjectStatus{
						Phase:     "Ready",
						Namespace: ptr.To(fmt.Sprintf("runai-%s", name)),
						NodePoolQuotaStatuses: []api.QuotaStatusNodePool{{
							NodePoolName:            nodePool.Name,
							NodePoolId:              nodePool.Id,
							AllocatedNonPreemptible: api.QuotaStatusResource{Gpu: gpu},
						}},
					},
				})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
			}

			setNonPreemptibleGpu(1)
			projectId, err := strconv.ParseInt(created.Id, 10, 64)
			Expect(err).To(BeNil())
			staleProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			setNonPreemptibleGpu(2)

			staleProject.OrgUnitData.Resources[0].GpuLimit = ptr.To(float64(4))
			_, err = dao.UpdateProject(ctx, staleProject, false, false)
			Expect(err).To(BeNil())

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			Expect(err).To(BeNil())
			got := getResponse.Body.(api.Project)
			Expect(*got.Resources[0].Gpu.Limit).To(Equal(float64(4)))
			Expect(got.Status.NodePoolQuotaStatuses[0].AllocatedNonPreemptible.Gpu).To(Equal(float64(2)))
		})
	})

	Context("test update project resources", func() {
		nodePoolName := "pool2"
		nodePoolId := float32(2)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
					{
						Name:                   nodePoolName,
						Id:                     fmt.Sprint(nodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		BeforeEach(func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test update project resources ok status", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
				Rank: ptr.To("Low"),
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateResources := getUpdateResources()
			updateResources[0].Rank = ptr.To("High")
			updateResponse, err := projectApiService.UpdateProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updatedResources := updateResponse.Body.([]api.Resources)
			Expect(len(updatedResources)).To(Equal(len(updateResources)))
			validateResourcesArrayToNullable(updateResources, updatedResources)

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProject := getResponse.Body.(api.Project)
			Expect(len(getProject.Resources)).To(Equal(len(updateResources)))
			validateResourcesArrayToNullable(updateResources, updatedResources)
			Expect(string(getProject.Status.Phase)).To(Equal("Updating"))
		})

		It("test update project not found", func() {
			updateResources := getUpdateResources()
			updateResponse, err := projectApiService.UpdateProjectResources(ctx, "777", updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test update project resources bad request", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateResources := getUpdateResources()
			updateResources[0].Cpu.Deserved = nil
			updateResponse, err := projectApiService.UpdateProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))

			updateResources = getUpdateResources()
			updateResources[0].Gpu.Deserved = ptr.To(float64(10))
			updateResponse, err = projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test update project resources no change", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			resources := []api.ResourcesNullable{}
			for i := range created.Resources {
				resources = append(resources, api.ResourcesNullable{
					NodePool: created.Resources[i].NodePool,
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(created.Resources[i].Gpu.Deserved),
						Limit:           created.Resources[i].Gpu.Limit,
						OverQuotaWeight: created.Resources[i].Gpu.OverQuotaWeight,
					},
					Cpu:    created.Resources[i].Cpu,
					Memory: created.Resources[i].Memory,
					Rank:   created.Resources[i].Rank,
				})
			}
			updateResponse, err := projectApiService.UpdateProjectResources(ctx, created.Id, resources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProject := getResponse.Body.(api.Project)
			Expect(string(getProject.Status.Phase)).To(Equal("Ready"))
		})
	})

	Context("test patch project resources", func() {
		nodePoolName := "pool2"
		nodePoolId := float32(2)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
					{
						Name:                   nodePoolName,
						Id:                     fmt.Sprint(nodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test patch project resources ok status", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
				Rank: ptr.To("High"),
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			updateResources := getUpdateResources()
			updateResources[0].Rank = ptr.To("Low")
			patchResponse, err := projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(patchResponse.Code).To(Equal(http.StatusOK))
			patchedResources := patchResponse.Body.([]api.Resources)
			Expect(len(patchedResources)).To(Equal(len(updateResources)))
			validateResourcesArrayToNullable(updateResources, patchedResources)

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProject := getResponse.Body.(api.Project)
			Expect(len(getProject.Resources)).To(Equal(len(updateResources)))
			validateResourcesArrayToNullable(updateResources, getProject.Resources)
			Expect(string(getProject.Status.Phase)).To(Equal("Updating"))

			updateResources = getUpdateResources()
			updateResources[0].Rank = ptr.To("LowUnknown")
			patchResponse, err = projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(patchResponse.Code).To(Equal(http.StatusBadRequest))

			// Success to Update Resources with empty (nil) priority
			updateResources[0].Rank = nil
			patchResponse, err = projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			Expect(patchResponse.Code).To(Equal(http.StatusOK))
		})

		It("test fail create project resources with unknown priority", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
				Rank: ptr.To("HighUnknown"),
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test patch project not found", func() {
			updateResources := getUpdateResources()
			patchResponse, err := projectApiService.PatchProjectResources(ctx, "777", updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(patchResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test patch project resources bad request", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateResources := getUpdateResources()
			updateResources[0].Memory.Deserved = nil
			updateResponse, err := projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))

			updateResources = getUpdateResources()
			updateResources[0].Gpu.Limit = ptr.To(float64(1))
			updateResponse, err = projectApiService.PatchProjectResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test patch project resources no change", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			limit := float64(1)
			deserved := float64(1)
			oqWeight := float64(1)
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   "2",
					Name: "pool2",
				},
				Gpu: api.ResourcesNullableGpu{
					Limit:           &limit,
					Deserved:        ptr.To(deserved),
					OverQuotaWeight: &oqWeight,
				},
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        created.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
				},
			}
			updateStatusResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateStatusResponse.Code).To(Equal(http.StatusOK))

			resources := []api.ResourcesNullable{}
			for i := range created.Resources {
				resources = append(resources, api.ResourcesNullable{
					NodePool: created.Resources[i].NodePool,
					Gpu: api.ResourcesNullableGpu{
						Deserved:        ptr.To(created.Resources[i].Gpu.Deserved),
						Limit:           created.Resources[i].Gpu.Limit,
						OverQuotaWeight: created.Resources[i].Gpu.OverQuotaWeight,
					},
					Cpu:    created.Resources[i].Cpu,
					Memory: created.Resources[i].Memory,
					Rank:   created.Resources[i].Rank,
				})
			}
			patchResponse, err := projectApiService.PatchProjectResources(ctx, created.Id, resources)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(patchResponse.Code).To(Equal(http.StatusOK))

			getResponse, err := projectApiService.GetProject(ctx, created.Id)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getProject := getResponse.Body.(api.Project)
			Expect(string(getProject.Status.Phase)).To(Equal("Ready"))
		})
	})

	Context("test delete projects", func() {
		It("test delete project success", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projectId := createResponse.Body.(api.Project).Id

			deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
			if err != nil {
				Fail("Failed deleting first project")
			}
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

			getResponse, err := projectApiService.GetProject(ctx, projectId)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test update project with unlimited resource succeeds", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectIds = append(projectIds, created.Id)

			updateRequest := getProjectUpdateRequest(created)
			updateRequest.Description = "new description"
			updateRequest.Resources[0].Gpu = api.ResourcesNullableGpu{
				Deserved:        ptr.To(float64(-1)),
				Limit:           nil,
				OverQuotaWeight: nil,
			}

			updateResponse, err := projectApiService.UpdateProject(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating project")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
		})
	})

	Context("test sort/filter", func() {
		var project1Id string
		var project2Id string
		nodePoolName1 := "nodepool-a"
		nodePoolId1 := float32(2)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		BeforeEach(func() {
			// project 1: total gpu quota: 2, total cpu quota: 2, total memory quota: 2
			// project 2: total gpu quota: 0, total cpu quota: 10, total memory quota: 1
			project1Name := createProjectName()
			request := getProjectCreateRequest(project1Name, project1Name, departmentId, clusterId)
			request.Resources[0].Gpu = api.ResourcesNullableGpu{
				Deserved:        ptr.To(float64(1)),
				Limit:           ptr.To(float64(1)),
				OverQuotaWeight: ptr.To(float64(1)),
			}
			request.Resources[0].Cpu = &api.ResourcesCpu{
				Deserved:        ptr.To(float64(1)),
				Limit:           ptr.To(float64(1)),
				OverQuotaWeight: ptr.To(float64(1)),
			}
			request.Resources[0].Memory = &api.ResourcesMemory{
				Deserved:        ptr.To(float64(1)),
				Limit:           ptr.To(float64(1)),
				OverQuotaWeight: ptr.To(float64(1)),
				Units:           api.UNITS_MB,
			}
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(nodePoolId1),
					Name: nodePoolName1,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(1)),
					OverQuotaWeight: ptr.To(float64(1)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(1)),
					OverQuotaWeight: ptr.To(float64(1)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(1)),
					Limit:           ptr.To(float64(1)),
					OverQuotaWeight: ptr.To(float64(1)),
					Units:           api.UNITS_MB,
				},
			})
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1Id = createResponse.Body.(api.Project).Id

			project2Name := createProjectName()
			request = getProjectCreateRequest(project2Name, project2Name, departmentId, clusterId)
			request.Resources[0].Gpu = api.ResourcesNullableGpu{
				Deserved:        ptr.To(float64(0)),
				Limit:           ptr.To(float64(0)),
				OverQuotaWeight: ptr.To(float64(0)),
			}
			request.Resources[0].Cpu = &api.ResourcesCpu{
				Deserved:        ptr.To(float64(10)),
				Limit:           ptr.To(float64(10)),
				OverQuotaWeight: ptr.To(float64(10)),
			}
			request.Resources[0].Memory = &api.ResourcesMemory{
				Deserved:        ptr.To(float64(1)),
				Limit:           ptr.To(float64(1)),
				OverQuotaWeight: ptr.To(float64(1)),
				Units:           api.UNITS_MB,
			}
			request.Resources = append(request.Resources, api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(nodePoolId1),
					Name: nodePoolName1,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
					Units:           api.UNITS_MB,
				},
			})

			createResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project2Id = createResponse.Body.(api.Project).Id
		})

		AfterEach(func() {
			if project1Id != "" {
				deleteResponse, err := projectApiService.DeleteProject(ctx, project1Id, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
				project1Id = ""
			}
			if project2Id != "" {
				deleteResponse, err := projectApiService.DeleteProject(ctx, project2Id, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
				project2Id = ""
			}
		})

		It("test project total quota values sort", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalGpuQuota", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalGpuQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project1Id))
			Expect(projects[1].Id).To(Equal(project2Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalCpuQuota", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project1Id))
			Expect(projects[1].Id).To(Equal(project2Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalCpuQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalMemoryQuota", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalMemoryQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project1Id))
			Expect(projects[1].Id).To(Equal(project2Id))
		})

		It("test project total quota values filter", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalGpuQuota>=2"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(1))
			Expect(projects[0].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalGpuQuota==0"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(1))
			Expect(projects[0].Id).To(Equal(project2Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalCpuQuota<=100"}, "totalCpuQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalCpuQuota>=100"}, "totalCpuQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(0))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalMemoryQuota!=0"}, "totalMemoryQuota", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "totalMemoryQuota!=1"}, "totalMemoryQuota", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(1))
			Expect(projects[0].Id).To(Equal(project1Id))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("parentId==%s", departmentId)}, "totalMemoryQuota", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(2))
			Expect(projects[0].Id).To(Equal(project2Id))
			Expect(projects[1].Id).To(Equal(project1Id))
		})
	})

	Context("test sort/filter by avg overtime data", func() {
		var (
			testProjects       []api.Project
			p0, p1, p2, p3, p4 string
		)

		BeforeAll(func() {
			for i := 0; i < 5; i++ {
				request := getProjectCreateRequest(fmt.Sprintf("p%d-%s", i, rand.String(3)), "", departmentId, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testProjects = append(testProjects, createResponse.Body.(api.Project))
			}

			result := map[string]map[string]*models.OvertimeData{clusterId: {}}

			var i float32 = 0
			for i = 0; i < 5; i++ {
				project := testProjects[int(i)]

				result[clusterId][project.Name] = &models.OvertimeData{
					AvgGpuAllocation24h:  ptr.To(float64(i * 10)),            // 0, 10, 20, 30, 40
					AvgGpuUtilization24h: ptr.To(float64(i * 20)),            // 0, 20, 40, 60, 80
					AvgGpuAllocation7d:   ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160
					AvgGpuUtilization7d:  ptr.To(float64(mod(2-i, 10) * 20)), // 20, 10, 0, 180, 160
					AvgGpuAllocation30d:  ptr.To(float64((5 - i) * 20)),      // 100, 80, 60, 40, 20
					AvgGpuUtilization30d: ptr.To(float64((5 - i) * 10)),      // 50, 40, 30, 20, 10
				}
			}

			err := dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())

			p0 = testProjects[0].Name
			p1 = testProjects[1].Name
			p2 = testProjects[2].Name
			p3 = testProjects[3].Name
			p4 = testProjects[4].Name
		})

		AfterAll(func() {
			for _, project := range testProjects {
				deleteResponse, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			testProjects = []api.Project{}
		})

		It("test sort by avg gpu 24h", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation24h", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p0, p1, p2, p3, p4))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization24h", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames = getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p0, p1, p2, p3, p4))
		})

		It("test sort by avg gpu 7d", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation7d", "DESC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p3, p4, p0, p1, p2))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization7d", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames = getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p2, p1, p0, p4, p3))
		})

		It("test sort by avg gpu 30d", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation30d", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p4, p3, p2, p1, p0))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization30d", "ASC", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(5))
			apiProjectNames = getProjectNames(projects)
			Expect(apiProjectNames).To(HaveExactElements(p4, p3, p2, p1, p0))
		})

		It("test filter by avg gpu 24h", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation24h>=10"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(4))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(ConsistOf(p1, p2, p3, p4))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization24h==60"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(1))
			Expect(projects[0].Name).To(Equal(p3))
		})

		It("test filter by avg gpu 7d", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation7d<=40"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(3))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(ConsistOf(p0, p1, p2))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization7d!=180"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(4))
			apiProjectNames = getProjectNames(projects)
			Expect(apiProjectNames).To(ConsistOf(p0, p1, p2, p4))
		})

		It("test filter by avg gpu 30d", func() {
			getResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation30d<=60"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects := getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(3))
			apiProjectNames := getProjectNames(projects)
			Expect(apiProjectNames).To(ConsistOf(p2, p3, p4))

			getResponse, err = projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization30d==20"}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			projects = getResponse.Body.(api.GetProjects200Response).Projects
			Expect(len(projects)).To(Equal(1))
			Expect(projects[0].Name).To(Equal(p3))
		})
	})

	Context("test count", func() {
		It("count projects filter by project name", func() {
			name1 := ""
			for i := 0; i < 4; i++ {
				name := createProjectName()
				if i == 0 {
					name1 = name
				}
				request := getProjectCreateRequest(name, name, departmentId, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				projectIds = append(projectIds, createResponse.Body.(api.Project).Id)
			}

			countResponse, err := projectApiService.CountProjects(ctx, []string{fmt.Sprintf("name==%s", name1)}, "")
			if err != nil {
				Fail("Failed counting projects")
			}
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			Expect(countResponse.Body.(api.CountResponse).Count).To(Equal(int64(1)))
		})

		It("create and delete projects then count and validate", func() {
			countResponse, err := projectApiService.CountProjects(ctx, []string{}, "")
			if err != nil {
				Fail("Failed counting projects")
			}
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			countBefore := countResponse.Body.(api.CountResponse).Count

			testProjectIds := []string{}
			for i := 0; i < 12; i++ {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, departmentId, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				testProjectIds = append(testProjectIds, created.Id)
				if i >= 6 {
					projectIds = append(projectIds, created.Id)
				}
			}

			for i := 0; i < 6; i++ {
				deleteResponse, err := projectApiService.DeleteProject(ctx, testProjectIds[i], "", false)
				if err != nil {
					Fail("Failed deleting project")
				}
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			}

			countResponse, err = projectApiService.CountProjects(ctx, []string{}, "")
			if err != nil {
				Fail("Failed counting projects")
			}
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(countBefore + 6))
		})

		It("internal count ignores the caller's permitted scopes", func() {
			createdIds := []string{}
			for i := 0; i < 2; i++ {
				name := createProjectName()
				createResponse, err := projectApiService.CreateProject(ctx, getProjectCreateRequest(name, name, departmentId, clusterId))
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				createdIds = append(createdIds, created.Id)
				projectIds = append(projectIds, created.Id)
			}

			unscoped, err := projectApiService.CountProjects(ctx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(unscoped.Code).To(Equal(http.StatusOK))
			total := unscoped.Body.(api.CountResponse).Count
			Expect(total).To(BeNumerically(">=", 2))

			limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{createdIds[0]},
			})

			scoped, err := projectApiService.CountProjects(limitedCtx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(scoped.Code).To(Equal(http.StatusOK))
			Expect(scoped.Body.(api.CountResponse).Count).To(Equal(int64(1)))

			internal, err := projectApiService.CountInternalProjects(limitedCtx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(internal.Code).To(Equal(http.StatusOK))
			Expect(internal.Body.(api.CountResponse).Count).To(Equal(total))
		})

	})

	Context("memory units", func() {
		var (
			nodePoolName1 = "nodepool-a"
			nodePoolId1   = float32(2)
		)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test units", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			request.Resources = []api.ResourcesNullable{{
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(defaultNodePoolId),
					Name: defaultNodePoolName,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(100)),
					Limit:           ptr.To(float64(100)),
					OverQuotaWeight: ptr.To(float64(100)),
					Units:           "MB",
				},
			}, {
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(nodePoolId1),
					Name: nodePoolName1,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(100)),
					Limit:           ptr.To(float64(100)),
					OverQuotaWeight: ptr.To(float64(100)),
					Units:           "Mib",
				},
			}}

			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			createdId := created.Id
			projectIds = append(projectIds, createdId)
			Expect(len(created.Resources)).To(Equal(2))
			Expect(created.Resources[0].Memory.Deserved).To(Equal(ptr.To(float64(100))))
			Expect(created.Resources[0].Memory.Units).To(Equal(api.UNITS_MB))
			Expect(created.Resources[1].Memory.Deserved).To(Equal(ptr.To(float64(100))))
			Expect(created.Resources[1].Memory.Units).To(Equal(api.UNITS_MIB))

			getResponse, err := projectApiService.GetProject(ctx, createdId)
			if err != nil {
				Fail("Failed getting project")
			}
			project := getResponse.Body.(api.Project)

			Expect(len(project.Resources)).To(Equal(2))
			Expect(project.Resources[0].Memory.Deserved).To(Equal(ptr.To(float64(100))))
			Expect(created.Resources[0].Memory.Units).To(Equal(api.UNITS_MB))
			Expect(created.Resources[1].Memory.Deserved).To(Equal(ptr.To(float64(100))))
			Expect(created.Resources[1].Memory.Units).To(Equal(api.UNITS_MIB))

			getV1Response, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, false)
			Expect(err).To(BeNil())
			projectsV1 := getV1Response.Body.([]api.ProjectV1)
			var projectV1 api.ProjectV1
			for i := range projectsV1 {
				if fmt.Sprint(projectsV1[i].Id) == createdId {
					projectV1 = projectsV1[i]
				}
			}
			Expect(len(projectV1.NodePoolsResources)).To(Equal(2))
			Expect(projectV1.NodePoolsResources[0].Memory.Deserved).To(Not(BeNil()))
			Expect(projectV1.NodePoolsResources[1].Memory.Deserved).To(Not(BeNil()))
			Expect([]float32{*projectV1.NodePoolsResources[0].Memory.Deserved, *projectV1.NodePoolsResources[1].Memory.Deserved}).To(
				ContainElements(float32(services.MbToMib(100)), float32(100)))
		})

		It("test units PUT", func() {
			name := createProjectName()
			request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
			request.Resources = []api.ResourcesNullable{{
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(defaultNodePoolId),
					Name: defaultNodePoolName,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(250)),
					Limit:           ptr.To(float64(250)),
					OverQuotaWeight: ptr.To(float64(250)),
					Units:           "GB",
				},
			}, {
				NodePool: &api.ResourcesNodePool{
					Id:   fmt.Sprint(nodePoolId1),
					Name: nodePoolName1,
				},
				Gpu: api.ResourcesNullableGpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Cpu: &api.ResourcesCpu{
					Deserved:        ptr.To(float64(0)),
					Limit:           ptr.To(float64(0)),
					OverQuotaWeight: ptr.To(float64(0)),
				},
				Memory: &api.ResourcesMemory{
					Deserved:        ptr.To(float64(250)),
					Limit:           ptr.To(float64(250)),
					OverQuotaWeight: ptr.To(float64(250)),
					Units:           "GB",
				},
			}}

			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			createdId := created.Id
			projectIds = append(projectIds, createdId)
			Expect(len(created.Resources)).To(Equal(2))
			Expect(created.Resources[0].Memory.Deserved).To(Equal(ptr.To(float64(250))))
			Expect(created.Resources[0].Memory.Units).To(Equal(api.UNITS_GB))
			Expect(created.Resources[1].Memory.Deserved).To(Equal(ptr.To(float64(250))))
			Expect(created.Resources[1].Memory.Units).To(Equal(api.UNITS_GB))

			getResponse, err := projectApiService.GetProject(ctx, createdId)
			if err != nil {
				Fail("Failed getting project")
			}
			project := getResponse.Body.(api.Project)

			Expect(len(project.Resources)).To(Equal(2))
			Expect(project.Resources[0].Memory.Deserved).To(Equal(ptr.To(float64(250))))
			Expect(created.Resources[0].Memory.Units).To(Equal(api.UNITS_GB))
			Expect(created.Resources[1].Memory.Deserved).To(Equal(ptr.To(float64(250))))
			Expect(created.Resources[1].Memory.Units).To(Equal(api.UNITS_GB))
		})
	})

	Context("projects in department", func() {
		var (
			nodePoolName1 = "nodepool-a"
			nodePoolId1   = float32(2)
			depId         string
			testProjects  = []api.Project{}
		)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			// department+projects service with 2 nodepools
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			createDepartmentRequest := getFullDepartmentCreateRequest("dep-1", clusterId)
			createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, CreateNodePoolApiResources(true, "2", nodePoolName1))
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating dep-1 department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			depId = response.Body.(api.Department).Id
		})

		AfterAll(func() {
			// delete created projects and department
			for _, project := range testProjects {
				_, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
				if err != nil {
					Fail("Failed deleting project")
				}
			}
			_, err := departmentApiService.DeleteDepartment(ctx, depId)
			if err != nil {
				Fail("Failed deleting department")
			}

			// department+projects service with 1 nodepool
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test delete department", func() {
			name := createProjectName()
			request := getProjectCreateRequestWithNodePools(name, name, depId, clusterId, nodePoolName1)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project := createResponse.Body.(api.Project)
			testProjects = append(testProjects, project)

			getResponse, err := projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			Expect(getResponse.Body.(api.Project).Id).To(Equal(project.Id))

			delResponse, err := departmentApiService.DeleteDepartment(ctx, depId)
			Expect(err).To(BeNil())
			Expect(delResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("test get projects metadata", func() {
		var (
			projectApi          api.ProjectsAPIServicer
			testCreatedProjects []api.Project
		)

		BeforeAll(func() {
			projectApi = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			quotaStatuses := []api.QuotaStatusNodePool{{
				NodePoolName: "default",
				NodePoolId:   "1",
				Allocated: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
			}}
			quotaStatus := api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
			}
			updateStatusRequest := api.ProjectStatusUpdateRequest{
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase:                 "Ready",
					PhaseMessage:          ptr.To("some messagee"),
					NodePoolQuotaStatuses: quotaStatuses,
					QuotaStatus:           quotaStatus,
					AdditionalStatusData: &map[string]interface{}{
						"conditions": []models.Condition{
							{
								Type:               "NamespaceReady",
								Status:             "True",
								Message:            "some message",
								Reason:             "ProjectNamespace",
								LastProbeTime:      nil,
								LastTransitionTime: ptr.To(time.Date(2024, 2, 15, 14, 58, 19, 0, time.UTC)),
							},
						},
					},
				},
			}

			for i := 0; i < 2; i++ {
				name := fmt.Sprintf("a%d-%s", i, createProjectName())
				request := getFullProjectCreateRequest(name, name, &departmentId, clusterId)
				response, err := projectApi.CreateProject(ctx, request)
				if err != nil {
					Fail(fmt.Sprintf("Failed creating project %s", name))
				}
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))
				proj := response.Body.(api.Project)
				testCreatedProjects = append(testCreatedProjects, proj)

				updateStatusRequest.Id = proj.Id
				updateStatusRequest.Status.Namespace = ptr.To(fmt.Sprintf("runai-%s", name))
				updateResponse, err := projectApi.UpdateProjectStatus(ctx, updateStatusRequest)
				if err != nil {
					Fail("Failed update project status")
				}
				Expect(updateResponse.Code).To(Equal(http.StatusOK))
			}
		})

		AfterAll(func() {
			for _, project := range testCreatedProjects {
				_, err := projectApi.DeleteProject(ctx, project.Id, "", false)
				if err != nil {
					Fail("Failed deleting project")
				}
			}
		})

		It("validate correct response", func() {
			metadataResponse, err := projectApi.GetProjectsMetadata(ctx, []string{}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata := metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(2))

			gottenProjects := map[string]api.ProjectMetadata{}
			for _, p := range getProjectsMetadata {
				gottenProjects[p.Id] = p
			}

			for _, proj := range testCreatedProjects {
				gottenProj, ok := gottenProjects[proj.Id]
				Expect(ok).To(BeTrue())

				Expect(gottenProj.Id).To(Equal(proj.Id))
				Expect(gottenProj.Name).To(Equal(proj.Name))
				Expect(*gottenProj.Namespace).To(Equal(fmt.Sprintf("runai-%s", proj.Name)))
				Expect(*gottenProj.Phase).To(Equal("Ready"))
				Expect(gottenProj.ClusterId).To(Equal(clusterId))
				Expect(gottenProj.Allocated).To(Equal(api.QuotaStatusResource{Gpu: 1, Cpu: 1, Memory: 1000}))
				Expect(gottenProj.Deserved).To(Equal(api.QuotaStatusResource{Gpu: proj.Resources[0].Gpu.Deserved, Cpu: *proj.Resources[0].Cpu.Deserved, Memory: *proj.Resources[0].Memory.Deserved * 1000}))
			}
		})

		It("test filter by", func() {
			metadataResponse, err := projectApi.GetProjectsMetadata(ctx, []string{fmt.Sprintf("name==%s", testCreatedProjects[0].Name)}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata := metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(1))
			Expect(getProjectsMetadata[0].Id).To(Equal(testCreatedProjects[0].Id))

			metadataResponse, err = projectApi.GetProjectsMetadata(ctx, []string{"phase==Ready"}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata = metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(2))

			metadataResponse, err = projectApi.GetProjectsMetadata(ctx, []string{"createdAt==2021-01-01T00:00:00Z"}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata = metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(0))

			metadataResponse, err = projectApi.GetProjectsMetadata(ctx, []string{"createdAt!=2021-01-01T00:00:00Z"}, "", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata = metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(2))
		})

		It("test search", func() {
			metadataResponse, err := projectApi.GetProjectsMetadata(ctx, []string{}, "", "", 0, 0, testCreatedProjects[0].Name)
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata := metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(1))
			Expect(getProjectsMetadata[0].Id).To(Equal(testCreatedProjects[0].Id))

			metadataResponse, err = projectApi.GetProjectsMetadata(ctx, []string{}, "", "", 0, 0, "Ready")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata = metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(2))
		})

		It("test sort by", func() {
			metadataResponse, err := projectApi.GetProjectsMetadata(ctx, []string{}, "name", "", 0, 0, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata := metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			Expect(len(getProjectsMetadata)).To(Equal(2))
			Expect(getProjectsMetadata[0].Id).To(Equal(testCreatedProjects[0].Id))
			Expect(getProjectsMetadata[1].Id).To(Equal(testCreatedProjects[1].Id))
		})

		It("test pagination", func() {
			metadataResponse, err := projectApi.GetProjectsMetadata(ctx, []string{}, "name", "", 0, 1, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata := metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			next := metadataResponse.Body.(api.GetProjectsMetadata200Response).Next
			Expect(len(getProjectsMetadata)).To(Equal(1))
			Expect(getProjectsMetadata[0].Id).To(Equal(testCreatedProjects[0].Id))
			Expect(next).To(Equal(int32(1)))

			metadataResponse, err = projectApi.GetProjectsMetadata(ctx, []string{}, "name", "", 1, 1, "")
			if err != nil {
				Fail("Failed getting projects metadata")
			}
			Expect(metadataResponse.Code).To(Equal(http.StatusOK))
			getProjectsMetadata = metadataResponse.Body.(api.GetProjectsMetadata200Response).Projects
			next = metadataResponse.Body.(api.GetProjectsMetadata200Response).Next
			Expect(len(getProjectsMetadata)).To(Equal(1))
			Expect(getProjectsMetadata[0].Id).To(Equal(testCreatedProjects[1].Id))
			Expect(next).To(Equal(int32(2)))
		})
	})

	Context("test get projects ranking", func() {
		cluster2Id := uuid.New().String()
		nodePoolName := "nodepool-a"
		var depId string
		var testProjects []api.Project
		var p0, p1, p2, p3, p4, p5, p6, p7, p8, p9 string

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: int32(TestTenantId),
				Uuid:     cluster2Id,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName,
						Id:                     "2",
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectDeletionService = services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			createDepartmentRequest := getFullDepartmentCreateRequest("parent", cluster2Id)
			createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, CreateNodePoolApiResources(true, "2", nodePoolName))
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating dep-1 department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated), "Failed creating department: %v", response.Body)
			depId = response.Body.(api.Department).Id
			for i := 0; i < 10; i++ {
				// Include nodepool-a resources so the nodepool filter test can return per-nodepool ratios
				request := getProjectCreateRequestWithNodePools(fmt.Sprintf("ranking-top-p-%d-%s", i, rand.String(3)), "", depId, cluster2Id, nodePoolName)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testProjects = append(testProjects, createResponse.Body.(api.Project))
			}

			result := map[string]map[string]*models.OvertimeData{cluster2Id: {}}

			var i float32 = 0
			for i = 0; i < 10; i++ {
				project := testProjects[int32(i)]

				result[cluster2Id][project.Name] = &models.OvertimeData{
					AvgGpuAllocation24h:        ptr.To(float64(i * 10)),            // 0, 10, 20, 30, 40, 50, 60, 70, 80, 90
					AvgGpuUtilization24h:       ptr.To(float64(i * 20)),            // 0, 20, 40, 60, 80, 100, 120, 140, 160, 180
					AvgGpuMemoryUtilization24h: ptr.To(float64(i * 15)),            // 0, 15, 30, 45, 60, 75, 90, 105, 120, 135
					AvgGpuAllocation7d:         ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160, 140, 120, 100, 80, 60
					AvgGpuUtilization7d:        ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160, 140, 120, 100, 80, 60
					AvgGpuMemoryUtilization7d:  ptr.To(float64(mod(2-i, 10) * 15)), // 30, 15, 0, 135, 120, 105, 90, 75, 60, 45
					AvgGpuAllocation30d:        ptr.To(float64((10 - i) * 10)),     // 100, 90, 80, 70, 60, 50, 40, 30, 20, 10
					AvgGpuUtilization30d:       ptr.To(float64((10 - i) * 5)),      // 50, 45, 40, 35, 30, 25, 20, 15, 10, 5
					AvgGpuMemoryUtilization30d: ptr.To(float64((10 - i) * 7)),      // 70, 63, 56, 49, 42, 35, 28, 21, 14, 7
				}
			}

			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())

			// Seed distinct allocated GPU so TOP/BOTTOM ratio ordering is meaningful.
			// project[i] gets i * 0.2 allocated (quota is 2), so ratio goes 0.0, 0.1, ..., 0.9.
			for idx, project := range testProjects {
				allocValue := float64(idx) * 0.2
				Expect(runaidb.DB().Actual().Exec(
					"UPDATE quota_status SET gpu_allocated = ? FROM project WHERE project.quota_status_id = quota_status.id AND project.id::text = ?",
					allocValue, project.Id,
				).Error).To(BeNil())
				// Also seed resources.allocated_gpu for the nodepool-a row used by the nodepool filter test
				Expect(runaidb.DB().Actual().Exec(
					"UPDATE resources SET allocated_gpu = ? FROM org_unit_data, project WHERE resources.org_unit_data_id = org_unit_data.id AND project.org_unit_data_id = org_unit_data.id AND project.id::text = ? AND resources.node_pool_name = ?",
					allocValue, project.Id, nodePoolName,
				).Error).To(BeNil())
			}

			p0 = testProjects[0].Name
			p1 = testProjects[1].Name
			p2 = testProjects[2].Name
			p3 = testProjects[3].Name
			p4 = testProjects[4].Name
			p5 = testProjects[5].Name
			p6 = testProjects[6].Name
			p7 = testProjects[7].Name
			p8 = testProjects[8].Name
			p9 = testProjects[9].Name
		})

		AfterAll(func() {
			for _, project := range testProjects {
				deleteResponse, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			testProjects = []api.Project{}
		})

		It("should get top 5 projects by GPU allocation 24h", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p9))
			Expect(ranking[1].Name).To(Equal(p8))
			Expect(ranking[2].Name).To(Equal(p7))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p5))
		})

		It("should get bottom 5 projects by GPU allocation 24h", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p0))
			Expect(ranking[1].Name).To(Equal(p1))
			Expect(ranking[2].Name).To(Equal(p2))
			Expect(ranking[3].Name).To(Equal(p3))
			Expect(ranking[4].Name).To(Equal(p4))
		})

		It("should get top 5 projects by GPU memory utilization 24h", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p9))
			Expect(ranking[1].Name).To(Equal(p8))
			Expect(ranking[2].Name).To(Equal(p7))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p5))
		})

		It("should get top 5 projects by GPU utilization 24h", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p9))
			Expect(ranking[1].Name).To(Equal(p8))
			Expect(ranking[2].Name).To(Equal(p7))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p5))
		})

		It("should get bottom 5 projects by GPU memory utilization 24h", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p0))
			Expect(ranking[1].Name).To(Equal(p1))
			Expect(ranking[2].Name).To(Equal(p2))
			Expect(ranking[3].Name).To(Equal(p3))
			Expect(ranking[4].Name).To(Equal(p4))
		})

		It("should get top 5 projects by GPU utilization 7d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p3))
			Expect(ranking[1].Name).To(Equal(p4))
			Expect(ranking[2].Name).To(Equal(p5))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p7))
		})

		It("should get top 5 projects by GPU memory utilization 7d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p3))
			Expect(ranking[1].Name).To(Equal(p4))
			Expect(ranking[2].Name).To(Equal(p5))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p7))
		})

		It("should get bottom 5 projects by GPU memory utilization 7d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p2))
			Expect(ranking[1].Name).To(Equal(p1))
			Expect(ranking[2].Name).To(Equal(p0))
			Expect(ranking[3].Name).To(Equal(p9))
			Expect(ranking[4].Name).To(Equal(p8))
		})

		It("should get top 5 projects by GPU utilization 30d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p0))
			Expect(ranking[1].Name).To(Equal(p1))
			Expect(ranking[2].Name).To(Equal(p2))
			Expect(ranking[3].Name).To(Equal(p3))
			Expect(ranking[4].Name).To(Equal(p4))
		})

		It("should get top 5 projects by GPU memory utilization 30d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p0))
			Expect(ranking[1].Name).To(Equal(p1))
			Expect(ranking[2].Name).To(Equal(p2))
			Expect(ranking[3].Name).To(Equal(p3))
			Expect(ranking[4].Name).To(Equal(p4))
		})

		It("should get bottom 5 projects by GPU memory utilization 30d", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			Expect(ranking[0].Name).To(Equal(p9))
			Expect(ranking[1].Name).To(Equal(p8))
			Expect(ranking[2].Name).To(Equal(p7))
			Expect(ranking[3].Name).To(Equal(p6))
			Expect(ranking[4].Name).To(Equal(p5))
		})

		It("should return GPU quota for each project in ranking", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects

			for _, project := range ranking {
				Expect(project.GpuQuota).NotTo(BeNil(), "GPU quota should not be nil for project %s", project.Name)
				// Projects now have 2 nodepools (default + nodepool-a), each with deserved=2, so total_gpu_quota=4
				Expect(*project.GpuQuota).To(Equal(float64(4)), "GPU quota should be 4 for project %s", project.Name)
				Expect(project.ParentName).NotTo(BeEmpty(), "Parent name should be populated for project %s", project.Name)
			}
		})

		It("should get projects ranked by GPU allocation ratio (bottom = lowest ratio)", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			// Verify AllocationValue is present and ascending (lowest ratio first = p0)
			for _, project := range ranking {
				Expect(project.AllocationValue).NotTo(BeNil())
				Expect(project.GpuQuota).NotTo(BeNil())
			}
			Expect(ranking[0].Name).To(Equal(p0), "BOTTOM should start with lowest ratio (p0, ratio=0)")
			Expect(ranking[9].Name).To(Equal(p9), "BOTTOM should end with highest ratio (p9, ratio=0.9)")
			for i := 1; i < len(ranking); i++ {
				Expect(*ranking[i-1].AllocationValue).To(BeNumerically("<=", *ranking[i].AllocationValue))
			}
		})

		It("should get projects ranked by GPU allocation ratio (top = highest ratio)", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
			// Verify AllocationValue is present and descending (highest ratio first = p9)
			for _, project := range ranking {
				Expect(project.AllocationValue).NotTo(BeNil())
			}
			Expect(ranking[0].Name).To(Equal(p9), "TOP should start with highest ratio (p9, ratio=0.9)")
			Expect(ranking[9].Name).To(Equal(p0), "TOP should end with lowest ratio (p0, ratio=0)")
			for i := 1; i < len(ranking); i++ {
				Expect(*ranking[i-1].AllocationValue).To(BeNumerically(">=", *ranking[i].AllocationValue))
			}
		})

		It("should filter projects ranking by departmentId", func() {
			// Create a second department with one project to prove the filter excludes it
			dep2Request := getFullDepartmentCreateRequest("dep2-filter-test", cluster2Id)
			dep2Response, err := departmentApiService.CreateDepartment(ctx, dep2Request)
			Expect(err).To(BeNil())
			Expect(dep2Response.Code).To(Equal(http.StatusCreated))
			dep2Id := dep2Response.Body.(api.Department).Id

			otherRequest := getProjectCreateRequest("other-dept-project", "", dep2Id, cluster2Id)
			otherResponse, err := projectApiService.CreateProject(ctx, otherRequest)
			Expect(err).To(BeNil())
			Expect(otherResponse.Code).To(Equal(http.StatusCreated))
			otherProjectId := otherResponse.Body.(api.Project).Id

			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", depId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			// Only the 10 projects in depId should be returned; the dep2 project should be excluded
			Expect(len(ranking)).To(Equal(10))
			for _, p := range ranking {
				Expect(p.Id).NotTo(Equal(otherProjectId), "project from dep2 must not appear in depId-filtered results")
			}

			// Cleanup
			_, _ = projectApiService.DeleteProject(ctx, otherProjectId, "", false)
			_, _ = departmentApiService.DeleteDepartment(ctx, dep2Id)
		})

		It("should get projects ranked by GPU allocation ratio with nodepool filter", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, nodePoolName, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			// All 10 projects have nodepool-a resources, so all 10 should be returned
			Expect(len(ranking)).To(Equal(10))
		})

		It("should return empty when filtering by non-existent departmentId", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "99999")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(0))
		})

		It("should handle validation error gracefully", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, "invalid-cluster-id", "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = projectApiService.GetProjectsRanking(ctx, api.RankingMetricType("blabla"), api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, "wrongggg", cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should handle workload count enrichment gracefully when workloads service is unavailable", func() {
			response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking)).To(Equal(10))
		})

		It("should return project ranking with workload counts enrichment", func() {
			workloadsClient := &workloadsMocks.IWorkloadsClient{}

			project1Id := testProjects[0].Id
			project2Id := testProjects[1].Id

			workloadsClient.On("InternalEnrichOrgUnitRanking", mock.Anything, mock.MatchedBy(func(req workloadsOpenapi.InternalEnrichOrgUnitRankingRequest) bool {
				return req.ClusterId == cluster2Id &&
					req.OrgUnitType == "project" &&
					len(req.Ids) > 0
			})).Return(&workloadsOpenapi.InternalEnrichOrgUnitRankingResponse{
				Items: []workloadsOpenapi.OrgUnitRankingItem{
					{
						Id:                    project1Id,
						RunningWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(5))),
						PendingWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(2))),
					},
					{
						Id:                    project2Id,
						RunningWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(3))),
						PendingWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(1))),
					},
				},
			}, nil)

			enrichedProjectApiService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, workloadsClient)

			response, err := enrichedProjectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects

			var project1Ranking, project2Ranking *api.ProjectGetRanking
			for i := range ranking {
				if ranking[i].Id == project1Id {
					project1Ranking = &ranking[i]
				}
				if ranking[i].Id == project2Id {
					project2Ranking = &ranking[i]
				}
			}

			Expect(project1Ranking).NotTo(BeNil())
			Expect(project1Ranking.RunningWorkloadsCount).NotTo(BeNil())
			Expect(*project1Ranking.RunningWorkloadsCount).To(Equal(int32(5)))
			Expect(project1Ranking.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*project1Ranking.PendingWorkloadsCount).To(Equal(int32(2)))

			Expect(project2Ranking).NotTo(BeNil())
			Expect(project2Ranking.RunningWorkloadsCount).NotTo(BeNil())
			Expect(*project2Ranking.RunningWorkloadsCount).To(Equal(int32(3)))
			Expect(project2Ranking.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*project2Ranking.PendingWorkloadsCount).To(Equal(int32(1)))

			workloadsClient.AssertExpectations(GinkgoT())
		})

		It("should handle workload enrichment failure gracefully", func() {
			workloadsClient := &workloadsMocks.IWorkloadsClient{}
			workloadsClient.On("InternalEnrichOrgUnitRanking", mock.Anything, mock.Anything).Return(
				(*workloadsOpenapi.InternalEnrichOrgUnitRankingResponse)(nil),
				fmt.Errorf("workloads service unavailable"),
			)

			enrichedProjectApiService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, workloadsClient)

			response, err := enrichedProjectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, cluster2Id, "", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetProjectsRanking200Response).Projects
			Expect(len(ranking) > 0).To(BeTrue())

			for _, project := range ranking {
				Expect(project.RunningWorkloadsCount).To(BeNil())
				Expect(project.PendingWorkloadsCount).To(BeNil())
			}
		})

		Context("older cluster", func() {
			var (
				depApiService  api.DepartmentsAPIServicer
				projApiService api.ProjectsAPIServicer
			)

			BeforeAll(func() {
				clusterServiceClient2 := mocks.IClusterClient{}
				clusterServiceClient2.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TestTenantId),
					Uuid:     cluster2Id,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.20")),
				}, nil)
				depApiService = departments.NewDepartmentApiService(&clusterServiceClient2, nil)
				projApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient2, nil)
			})

			It("should fail getting top/bottom 5 projects/departments by GPU memory utilization", func() {
				rankingToTest := []api.RankingMetricType{api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D}
				ordersToTest := []api.RankingOrderType{api.RANKINGORDERTYPE_TOP, api.RANKINGORDERTYPE_BOTTOM}

				for _, ranking := range rankingToTest {
					for _, order := range ordersToTest {
						response, err := projApiService.GetProjectsRanking(ctx, ranking, order, cluster2Id, "", "")
						Expect(err).To(BeNil())
						Expect(response.Code).To(Equal(http.StatusBadRequest))

						response, err = depApiService.GetDepartmentsRanking(ctx, ranking, order, cluster2Id, "")
						Expect(err).To(BeNil())
						Expect(response.Code).To(Equal(http.StatusBadRequest))
					}
				}

				// sanity for a different metric
				response, err := projApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, cluster2Id, "", "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				response, err = depApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D, api.RANKINGORDERTYPE_TOP, cluster2Id, "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
			})
		})

		Context("new cluster but gpu_memory_utilization is not set", func() {
			var testProjects2 []api.Project
			var clusterId2, depId2 string

			BeforeAll(func() {
				clusterId2 = uuid.New().String()

				clusterServiceClient := mocks.IClusterClient{}
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TestTenantId),
					Uuid:     clusterId2,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
				}, nil)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
					},
				}, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
				departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)

				createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId2)
				response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
				if err != nil {
					Fail("Failed creating parent department")
				}
				Expect(response.Code).To(Equal(http.StatusCreated))
				depId2 = response.Body.(api.Department).Id

				for i := 0; i < 3; i++ {
					request := getProjectCreateRequest(fmt.Sprintf("ranking-top-p-%d-%s", i, rand.String(3)), "", depId2, clusterId2)
					createResponse, err := projectApiService.CreateProject(ctx, request)
					if err != nil {
						Fail("Failed creating project")
					}
					Expect(createResponse.Code).To(Equal(http.StatusCreated))
					testProjects2 = append(testProjects2, createResponse.Body.(api.Project))
				}

				result := map[string]map[string]*models.OvertimeData{clusterId2: {}}

				var i float32 = 0
				for i = 0; i < 3; i++ {
					project := testProjects2[int32(i)]

					result[clusterId2][project.Name] = &models.OvertimeData{
						AvgGpuAllocation24h:        ptr.To(float64(i * 10)), // 0, 10, 20, 30, 40, 50, 60, 70, 80, 90
						AvgGpuUtilization24h:       ptr.To(float64(i * 20)), // 0, 20, 40, 60, 80, 100, 120, 140, 160, 180
						AvgGpuMemoryUtilization24h: nil,
						AvgGpuAllocation7d:         ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160, 140, 120, 100, 80, 60
						AvgGpuUtilization7d:        ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160, 140, 120, 100, 80, 60
						AvgGpuMemoryUtilization7d:  nil,
						AvgGpuAllocation30d:        ptr.To(float64((10 - i) * 10)), // 100, 90, 80, 70, 60, 50, 40, 30, 20, 10
						AvgGpuUtilization30d:       ptr.To(float64((10 - i) * 5)),  // 50, 45, 40, 35, 30, 25, 20, 15, 10, 5
						AvgGpuMemoryUtilization30d: nil,
					}
				}

				err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
				Expect(err).To(BeNil())
				err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
				Expect(err).To(BeNil())
				err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
				Expect(err).To(BeNil())
			})

			AfterAll(func() {
				for _, project := range testProjects2 {
					deleteResponse, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
					if err != nil || deleteResponse.Code != http.StatusNoContent {
						Fail("Failed deleting project")
					}
				}
				testProjects2 = []api.Project{}

				_, err := departmentApiService.DeleteDepartment(ctx, depId2)
				if err != nil {
					Fail("Failed deleting department")
				}
			})

			It("ranking for gpu memory utilization when it is not set should still return all projects/departments with null values", func() {
				response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGORDERTYPE_TOP, clusterId2, "", "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				ranking := response.Body.(api.GetProjectsRanking200Response).Projects
				// Should return all 3 projects even though gpu_memory_utilization is not set
				// Projects should be ordered by name since all have NULL values
				Expect(len(ranking)).To(Equal(3))
				for _, project := range ranking {
					Expect(project.MemoryUtilizationValue).To(BeNil())
				}

				response, err = departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D, api.RANKINGORDERTYPE_BOTTOM, clusterId2, "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				ranking2 := response.Body.(api.GetDepartmentsRanking200Response).Departments
				// Should return the department even though gpu_memory_utilization is not set
				Expect(len(ranking2)).To(Equal(1))
				Expect(ranking2[0].MemoryUtilizationValue).To(BeNil())
			})
		})

		Context("ranking with mixed data - some projects have data, some don't", func() {
			var testProjectsMixed []api.Project
			var clusterId3, depId3 string
			// Projects with data (will get overtime data set)
			var pWithData1, pWithData2 string
			// Projects without data (will NOT get overtime data set)
			var pNoData1, pNoData2 string

			BeforeAll(func() {
				clusterId3 = uuid.New().String()

				clusterServiceClient := mocks.IClusterClient{}
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TestTenantId),
					Uuid:     clusterId3,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
				}, nil)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
					},
				}, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
				departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)

				createDepartmentRequest := getFullDepartmentCreateRequest("parent-mixed", clusterId3)
				response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
				if err != nil {
					Fail("Failed creating parent department")
				}
				Expect(response.Code).To(Equal(http.StatusCreated))
				depId3 = response.Body.(api.Department).Id

				// Create 4 projects with predictable names for ordering verification
				// Names are chosen so alphabetical order is: aaa-no-data, bbb-with-data, ccc-no-data, ddd-with-data
				projectNames := []string{"aaa-no-data", "bbb-with-data", "ccc-no-data", "ddd-with-data"}
				for _, name := range projectNames {
					request := getProjectCreateRequest(name, "", depId3, clusterId3)
					createResponse, err := projectApiService.CreateProject(ctx, request)
					if err != nil {
						Fail("Failed creating project")
					}
					Expect(createResponse.Code).To(Equal(http.StatusCreated))
					testProjectsMixed = append(testProjectsMixed, createResponse.Body.(api.Project))
				}

				pNoData1 = testProjectsMixed[0].Name   // aaa-no-data
				pWithData1 = testProjectsMixed[1].Name // bbb-with-data
				pNoData2 = testProjectsMixed[2].Name   // ccc-no-data
				pWithData2 = testProjectsMixed[3].Name // ddd-with-data

				// Only set overtime data for projects with "with-data" in name
				result := map[string]map[string]*models.OvertimeData{clusterId3: {}}

				result[clusterId3][pWithData1] = &models.OvertimeData{
					AvgGpuAllocation24h:        ptr.To(float64(50)),
					AvgGpuUtilization24h:       ptr.To(float64(60)),
					AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
				}
				result[clusterId3][pWithData2] = &models.OvertimeData{
					AvgGpuAllocation24h:        ptr.To(float64(80)),
					AvgGpuUtilization24h:       ptr.To(float64(90)),
					AvgGpuMemoryUtilization24h: ptr.To(float64(100)),
				}

				err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
				Expect(err).To(BeNil())
			})

			AfterAll(func() {
				for _, project := range testProjectsMixed {
					deleteResponse, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
					if err != nil || deleteResponse.Code != http.StatusNoContent {
						Fail("Failed deleting project")
					}
				}
				testProjectsMixed = []api.Project{}

				_, err := departmentApiService.DeleteDepartment(ctx, depId3)
				if err != nil {
					Fail("Failed deleting department")
				}
			})

			It("should return all 4 projects with data-having projects first (ordered by value DESC), then no-data projects (ordered by name)", func() {
				response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, clusterId3, "", "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				ranking := response.Body.(api.GetProjectsRanking200Response).Projects
				Expect(len(ranking)).To(Equal(4))

				// TOP order (DESC): projects with data first (highest value first), then no-data projects by name
				// ddd-with-data (80) -> bbb-with-data (50) -> aaa-no-data (null) -> ccc-no-data (null)
				Expect(ranking[0].Name).To(Equal(pWithData2)) // ddd-with-data, value=80
				Expect(ranking[0].AllocationValue).NotTo(BeNil())
				Expect(*ranking[0].AllocationValue).To(Equal(float64(80)))

				Expect(ranking[1].Name).To(Equal(pWithData1)) // bbb-with-data, value=50
				Expect(ranking[1].AllocationValue).NotTo(BeNil())
				Expect(*ranking[1].AllocationValue).To(Equal(float64(50)))

				Expect(ranking[2].Name).To(Equal(pNoData1)) // aaa-no-data (alphabetically first among nulls)
				Expect(ranking[2].AllocationValue).To(BeNil())

				Expect(ranking[3].Name).To(Equal(pNoData2)) // ccc-no-data (alphabetically second among nulls)
				Expect(ranking[3].AllocationValue).To(BeNil())
			})

			It("should return all 4 projects with data-having projects first (ordered by value ASC), then no-data projects (ordered by name)", func() {
				response, err := projectApiService.GetProjectsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_BOTTOM, clusterId3, "", "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				ranking := response.Body.(api.GetProjectsRanking200Response).Projects
				Expect(len(ranking)).To(Equal(4))

				// BOTTOM order (ASC): projects with data first (lowest value first), then no-data projects by name
				// bbb-with-data (50) -> ddd-with-data (80) -> aaa-no-data (null) -> ccc-no-data (null)
				Expect(ranking[0].Name).To(Equal(pWithData1)) // bbb-with-data, value=50
				Expect(ranking[0].AllocationValue).NotTo(BeNil())
				Expect(*ranking[0].AllocationValue).To(Equal(float64(50)))

				Expect(ranking[1].Name).To(Equal(pWithData2)) // ddd-with-data, value=80
				Expect(ranking[1].AllocationValue).NotTo(BeNil())
				Expect(*ranking[1].AllocationValue).To(Equal(float64(80)))

				Expect(ranking[2].Name).To(Equal(pNoData1)) // aaa-no-data (alphabetically first among nulls)
				Expect(ranking[2].AllocationValue).To(BeNil())

				Expect(ranking[3].Name).To(Equal(pNoData2)) // ccc-no-data (alphabetically second among nulls)
				Expect(ranking[3].AllocationValue).To(BeNil())
			})
		})

		Context("cross-tenant isolation", func() {
			// Tenant A owns the cluster and has projects on it.
			// Tenant B is an attacker who knows Tenant A's cluster ID and tries
			// to call GetProjectsRanking with it. The request should be rejected
			// because the cluster does not belong to Tenant B.
			var (
				tenantAId              = TestTenant
				tenantBId              = "999"
				tenantACluster         string
				tenantADepId           string
				tenantAProjects        []api.Project
				crossTenantProjService api.ProjectsAPIServicer
				crossTenantDepService  api.DepartmentsAPIServicer
			)

			BeforeAll(func() {
				tenantACluster = uuid.New().String()
				TenantAId, _ := strconv.Atoi(tenantAId)
				crossTenantClusterClient := mocks.IClusterClient{}
				// Mock returns Tenant A's TenantId — Tenant B's requests will fail cluster ownership validation
				crossTenantClusterClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					TenantId: int32(TenantAId),
					Uuid:     tenantACluster,
					Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
				}, nil)
				crossTenantClusterClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
						},
					},
				}, nil)

				crossTenantDeletionService := services.NewProjectDeletionService(&crossTenantClusterClient, nil, nil)
				crossTenantDepService = departments.NewDepartmentApiService(&crossTenantClusterClient, nil)
				crossTenantProjService = projects.NewProjectApiService(crossTenantDeletionService, &crossTenantClusterClient, nil)

				// Create department + project under Tenant A (the cluster owner)
				ctxA := SetContext(tenantAId, TestUser)
				depReq := getFullDepartmentCreateRequest("cross-tenant-dep", tenantACluster)
				resp, err := crossTenantDepService.CreateDepartment(ctxA, depReq)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusCreated))
				tenantADepId = resp.Body.(api.Department).Id

				projReq := getProjectCreateRequest("cross-tenant-proj", "", tenantADepId, tenantACluster)
				createResp, err := crossTenantProjService.CreateProject(ctxA, projReq)
				Expect(err).To(BeNil())
				Expect(createResp.Code).To(Equal(http.StatusCreated))
				tenantAProjects = append(tenantAProjects, createResp.Body.(api.Project))

				// Set overtime data for Tenant A's project
				result := map[string]map[string]*models.OvertimeData{tenantACluster: {
					tenantAProjects[0].Name: {
						AvgGpuAllocation24h:  ptr.To(float64(50)),
						AvgGpuUtilization24h: ptr.To(float64(60)),
					},
				}}
				err = dao.SetProjectsOvertimeDataForTenant(ctxA, tenantAId, result, dao.RangeDay)
				Expect(err).To(BeNil())
			})

			AfterAll(func() {
				ctxA := SetContext(tenantAId, TestUser)
				for _, project := range tenantAProjects {
					crossTenantProjService.DeleteProject(ctxA, project.Id, "", false)
				}
				crossTenantDepService.DeleteDepartment(ctxA, tenantADepId)
			})

			It("GetProjectsRanking should reject a different tenant querying another tenant's cluster", func() {
				ctxB := SetContext(tenantBId, TestUser)
				response, err := crossTenantProjService.GetProjectsRanking(ctxB, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, tenantACluster, "", "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusUnauthorized),
					"Tenant B should be rejected when querying Tenant A's cluster")
			})
		})
	})

	//Context("test update project parent", func() {
	//	var department2Id string
	//	BeforeAll(func() {
	//		createDepartmentRequest := getFullDepartmentCreateRequest("parent2", clusterId)
	//		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
	//		if err != nil {
	//			Fail("Failed creating parent department")
	//		}
	//		Expect(response.Code).To(Equal(http.StatusCreated))
	//		department2Id = response.Body.(api.Department).Id
	//	})
	//
	//	AfterAll(func() {
	//		_, err := departmentApiService.DeleteDepartment(ctx, department2Id)
	//		if err != nil {
	//			Fail("Failed deleting department")
	//		}
	//	})
	//
	//	It("update parent successful", func() {
	//		name := createProjectName()
	//		request := getProjectCreateRequest(name, name, departmentId, clusterId)
	//		createResponse, err := projectApiService.CreateProject(ctx, request)
	//		if err != nil {
	//			Fail("Failed creating project")
	//		}
	//		Expect(createResponse.Code).To(Equal(http.StatusCreated))
	//		project := createResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(departmentId))
	//
	//		getResponse, err := projectApiService.GetProject(ctx, project.Id, clusterId)
	//		if err != nil {
	//			Fail("Failed getting project")
	//		}
	//		Expect(getResponse.Code).To(Equal(http.StatusOK))
	//		project = getResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(departmentId))
	//
	//		updateParentRequest := api.UpdateProjectParentRequest{
	//			ParentId: department2Id,
	//		}
	//		updateResponse, err := projectApiService.UpdateProjectParent(ctx, project.Id, updateParentRequest)
	//		if err != nil {
	//			Fail("Failed update project parent")
	//		}
	//		Expect(updateResponse.Code).To(Equal(http.StatusOK))
	//
	//		getResponse, err = projectApiService.GetProject(ctx, project.Id, clusterId)
	//		if err != nil {
	//			Fail("Failed getting project")
	//		}
	//		Expect(getResponse.Code).To(Equal(http.StatusOK))
	//		project = getResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(department2Id))
	//	})
	//
	//	It("update parent fail", func() {
	//		name := createProjectName()
	//		request := getProjectCreateRequest(name, name, departmentId, clusterId)
	//		createResponse, err := projectApiService.CreateProject(ctx, request)
	//		if err != nil {
	//			Fail("Failed creating project")
	//		}
	//		Expect(createResponse.Code).To(Equal(http.StatusCreated))
	//		project := createResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(departmentId))
	//
	//		updateParentRequest := api.UpdateProjectParentRequest{
	//			ParentId: uuid.New().String(),
	//		}
	//		updateResponse, err := projectApiService.UpdateProjectParent(ctx, project.Id, updateParentRequest)
	//		if err != nil {
	//			Fail("Failed update project parent")
	//		}
	//		Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
	//	})
	//})

	// add when there is multi hierarchy
	//Context("test verbosity", func() {
	//	var department2Id string
	//	var department3Id string
	//	BeforeEach(func() {
	//		createDepartmentRequest := getFullDepartmentCreateRequest("parent2", clusterId)
	//		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
	//		if err != nil {
	//			Fail("Failed creating parent department")
	//		}
	//		Expect(response.Code).To(Equal(http.StatusCreated))
	//		department2Id = response.Body.(api.Department).Id
	//		createDepartment3Request := getFullDepartmentCreateRequest("parent3", clusterId)
	//		response3, err := departmentApiService.CreateDepartment(ctx, createDepartment3Request)
	//		if err != nil {
	//			Fail("Failed creating parent department")
	//		}
	//		Expect(response3.Code).To(Equal(http.StatusCreated))
	//		department3Id = response3.Body.(api.Department).Id
	//	})
	//
	//	AfterEach(func() {
	//		_, err := departmentApiService.DeleteDepartment(ctx, department2Id)
	//		if err != nil {
	//			Fail("Failed deleting department")
	//		}
	//		_, err = departmentApiService.DeleteDepartment(ctx, department3Id)
	//		if err != nil {
	//			Fail("Failed deleting department")
	//		}
	//	})
	//
	//	It("test get verbosity", func() {
	//		request := getProjectCreateRequest("proj1", "proj1-ns", department2Id, clusterId)
	//		createResponse, err := projectApiService.CreateProject(ctx, request)
	//		if err != nil {
	//			Fail("Failed creating project")
	//		}
	//		Expect(createResponse.Code).To(Equal(http.StatusCreated))
	//		project := createResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(department2Id))
	//
	//		getResponse, err := projectApiService.GetProject(ctx, project.Id, clusterId, AncestorsVerbosity)
	//		if err != nil {
	//			Fail("Failed getting project")
	//		}
	//		Expect(getResponse.Code).To(Equal(http.StatusOK))
	//		project = getResponse.Body.(api.Project)
	//		Expect(*project.ParentId).To(Equal(department2Id))
	//		parent := project.Parent
	//		Expect(parent.Id).To(Equal(department2Id))
	//		Expect(parent.Name).To(Equal("parent2"))
	//		parent = parent.Parent
	//		Expect(parent.Id).To(Equal(departmentId))
	//		Expect(parent.Name).To(Equal("parent"))
	//		Expect(parent.Parent).To(BeNil())
	//
	//		getResponse, err = departmentApiService.GetDepartment(ctx, departmentId, clusterId, TreeVerbosity)
	//		if err != nil {
	//			Fail("Failed getting project")
	//		}
	//		Expect(getResponse.Code).To(Equal(http.StatusOK))
	//		department := getResponse.Body.(api.Department)
	//		children := *department.Children
	//		Expect(len(children)).To(Equal(1))
	//		child1 := children[0]
	//		Expect(child1.Id).To(Equal(department2Id))
	//		Expect(child1.Name).To(Equal("parent2"))
	//		children = child1.Children
	//		Expect(len(children)).To(Equal(2))
	//		child1 = children[0]
	//		child2 := children[1]
	//		Expect(child1.Id).To(Equal(project.Id))
	//		Expect(child1.Name).To(Equal("proj1"))
	//		Expect(child2.Id).To(Equal(department3Id))
	//		Expect(child2.Name).To(Equal("parent3"))
	//	})
	//})
})

func getProjectCreateRequest(name string, namespace string, parentId string, clusterId string) api.ProjectCreationRequest {
	return api.ProjectCreationRequest{
		Name:               name,
		Description:        fmt.Sprintf("%s project description", name),
		RequestedNamespace: namespace,
		ClusterId:          clusterId,
		ParentId:           &parentId,
		Resources:          createDefaultApiResources(false),
	}
}

func getProjectCreateRequestWithNodePools(name string, namespace string, parentId string, clusterId, nodePoolName1 string) api.ProjectCreationRequest {
	return api.ProjectCreationRequest{
		Name:               name,
		RequestedNamespace: namespace,
		ClusterId:          clusterId,
		ParentId:           &parentId,
		Resources:          []api.ResourcesNullable{createDefaultApiResources(true)[0], CreateNodePoolApiResources(true, "2", nodePoolName1)},
	}
}

func getFullProjectCreateRequest(name string, namespace string, parentId *string, clusterId string) api.ProjectCreationRequest {
	config := createDefaultApiJobConfig()
	return api.ProjectCreationRequest{
		Name:               name,
		RequestedNamespace: namespace,
		ClusterId:          clusterId,
		ParentId:           parentId,
		SchedulingRules:    &config,
		Resources:          createDefaultApiResources(true),
	}
}

func getProjectUpdateRequest(project api.Project) api.ProjectUpdateRequest {
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
		SchedulingRules: project.SchedulingRules,
		Resources:       resources,
	}
}

func getProjectNames(projects []api.Project) []string {
	names := []string{}
	for _, project := range projects {
		names = append(names, project.Name)
	}
	return names
}

func mod(a, b float32) float32 {
	result := int(a) % int(b)
	if result < 0 {
		result += int(b)
	}
	return float32(result)
}
