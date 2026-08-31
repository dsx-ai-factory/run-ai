/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Org-Unit status API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var projectIds []string
	var departmentId string
	var nodeType1Id string
	var nodeType2Id string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var nodeTypeService api.NodeTypesAPIServicer
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *openapi.NewNullableString(ptr.To("2.17")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
		Nodepools: []openapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				Phase:                  ptr.To(openapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   "pool1",
				Id:                     "2",
				Phase:                  ptr.To(openapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)

	BeforeAll(func() {
		nodeTypeService = services.NewNodeTypeApiService()
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId = response.Body.(api.Department).Id
		nodeTypeRequest := api.NodeTypeCreateRequest{Name: "nodeType1" + uuid.New().String(), ClusterId: clusterId}
		nodeTypeResponse, err := nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
		if err != nil {
			Fail("Failed creating node type")
		}
		Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
		nodeType1Id = nodeTypeResponse.Body.(api.NodeType).Id
		nodeTypeRequest = api.NodeTypeCreateRequest{Name: "nodeType2" + uuid.New().String(), ClusterId: clusterId}
		nodeTypeResponse, err = nodeTypeService.CreateNodeType(ctx, nodeTypeRequest)
		if err != nil {
			Fail("Failed creating node type")
		}
		Expect(nodeTypeResponse.Code).To(Equal(http.StatusCreated))
		nodeType2Id = nodeTypeResponse.Body.(api.NodeType).Id
	})

	BeforeEach(func() {
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
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

	Context("test update project status", func() {
		It("update status successful", func() {
			name := createProjectName()
			request := getProjectCreateRequest(name, name, departmentId, clusterId)
			request.Resources = append(request.Resources, CreateNodePoolApiResources(false, "2", "pool1"))
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated), "Create project response: %+v", createResponse)
			project := createResponse.Body.(api.Project)

			projectId := project.Id
			projectIds = append(projectIds, projectId)
			getResponse, err := projectApiService.GetProject(ctx, projectId)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			project = getResponse.Body.(api.Project)
			Expect(string(project.Status.Phase)).To(Equal("Creating"))
			Expect(project.Status.Namespace).To(BeNil())

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
			}, {
				NodePoolName: "pool1",
				NodePoolId:   "2",
				Allocated: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
			}}
			quotaStatus := api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6000,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4000,
				},
			}
			updateStatusRequest := api.ProjectStatusUpdateRequest{
				Id:        project.Id,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase:                 "Ready",
					PhaseMessage:          ptr.To("some messagee"),
					Namespace:             ptr.To("runai-proj1"),
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

			updateResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
			if err != nil {
				Fail("Failed update project status")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			var nilTime *time.Time = nil

			getResponse, err = projectApiService.GetProject(ctx, projectId)
			if err != nil {
				Fail("Failed getting project")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			project = getResponse.Body.(api.Project)
			Expect(string(project.Status.Phase)).To(Equal("Ready"))
			Expect(*project.Status.Namespace).To(Equal("runai-proj1"))
			Expect(*project.Status.PhaseMessage).To(Equal("some messagee"))
			Expect(project.Status.NodePoolQuotaStatuses).To(ConsistOf(quotaStatuses))
			Expect(project.Status.QuotaStatus).To(Equal(quotaStatus))
			Expect(project.Status.AdditionalStatusData).ToNot(BeNil())
			apiConditions, ok := (*project.Status.AdditionalStatusData)["conditions"]
			Expect(ok).To(BeTrue())

			marshaled, err := json.Marshal(apiConditions)
			if err != nil {
				Fail("Failed marshaling conditions")
			}
			conditions := []models.Condition{}
			err = json.Unmarshal(marshaled, &conditions)
			if err != nil {
				Fail("Failed unmarshaling conditions")
			}

			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Type).To(Equal("NamespaceReady"))
			Expect(conditions[0].Status).To(Equal("True"))
			Expect(conditions[0].Message).To(Equal("some message"))
			Expect(conditions[0].Reason).To(Equal("ProjectNamespace"))
			Expect(conditions[0].LastProbeTime).To(BeEquivalentTo(nilTime))
			Expect(*conditions[0].LastTransitionTime).To(Equal(time.Date(2024, 2, 15, 14, 58, 19, 0, time.UTC)))
		})
	})

	Context("projects in department", func() {
		var (
			nodePoolName1 = "pool1"
			nodePoolId1   = float32(2)
			dep2Id        string
			dep2Name      = "dep-1"
			testProjects  = []api.Project{}

			quotaStatuses = []api.QuotaStatusNodePool{{
				NodePoolName: "default",
				NodePoolId:   "1",
				Allocated: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    1,
					Memory: 1,
				},
			}, {
				NodePoolName: "pool1",
				NodePoolId:   "2",
				Allocated: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    1,
					Cpu:    2,
					Memory: 3,
				},
			}}
			quotaStatus = api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4,
				},
			}
			updateStatusRequest = api.ProjectStatusUpdateRequest{
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase:                 "Ready",
					PhaseMessage:          ptr.To("some messagee"),
					Namespace:             ptr.To("runai-proj1"),
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
		)

		BeforeAll(func() {
			// department+projects service with 2 nodepools
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			createDepartmentRequest := getFullDepartmentCreateRequest(dep2Name, clusterId)
			createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, CreateNodePoolApiResources(true, "2", nodePoolName1))
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating dep-1 department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			dep2Id = response.Body.(api.Department).Id
		})

		AfterAll(func() {
			// delete created projects and department
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			for _, project := range testProjects {
				_, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
				if err != nil {
					Fail("Failed deleting project")
				}
			}
			_, err := departmentApiService.DeleteDepartment(ctx, dep2Id)
			if err != nil {
				Fail("Failed deleting department")
			}

			// department+projects service with 1 nodepool
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test department ProjectAggregatedNodePoolsResources + QuotaStatus + QuotaStatuses", func() {
			// create 2 projects in the department
			for i := 0; i < 2; i++ {
				name := createProjectName()
				request := getProjectCreateRequestWithNodePools(name, name, dep2Id, clusterId, nodePoolName1)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testProjects = append(testProjects, createResponse.Body.(api.Project))
			}

			// update status of projects
			for _, project := range testProjects {
				updateStatusRequest.Id = project.Id
				updateResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
				if err != nil {
					Fail("Failed update project status")
				}
				Expect(updateResponse.Code).To(Equal(http.StatusOK))
			}

			// get department and validate status + projectAggregated response
			getResponse, err := departmentApiService.GetDepartment(ctx, dep2Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDep := getResponse.Body.(api.Department)
			Expect(getDep.Id).To(Equal(dep2Id))

			// sum for both projects
			expectedQuotaStatuses := []api.QuotaStatusNodePool{{
				NodePoolName: "default",
				NodePoolId:   "1",
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    2,
					Memory: 2,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    2,
					Memory: 2,
				},
			}, {
				NodePoolName: "pool1",
				NodePoolId:   "2",
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6,
				},
			}}
			expectedQuotaStatus := api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    4,
					Cpu:    6,
					Memory: 8,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    4,
					Cpu:    8,
					Memory: 12,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    4,
					Cpu:    6,
					Memory: 8,
				},
			}
			expectedProjectAggregatedResourcesMap := map[string]api.AggregatedResources{}
			for _, project := range testProjects {
				for _, npResources := range project.Resources {
					eNpR, ok := expectedProjectAggregatedResourcesMap[npResources.NodePool.Name]
					if !ok {
						r := api.AggregatedResources{
							NodePool: api.AggregatedResourcesNodePool{
								Id:   npResources.NodePool.Id,
								Name: npResources.NodePool.Name,
							},
							Gpu: api.DeservedResource{
								Deserved: ptr.To(npResources.Gpu.Deserved),
							},
							Cpu: api.DeservedResource{
								Deserved: ptr.To(*npResources.Cpu.Deserved),
							},
							Memory: api.AggregatedResourcesMemory{
								Deserved: ptr.To(*npResources.Memory.Deserved),
								Units:    npResources.Memory.Units,
							},
						}
						expectedProjectAggregatedResourcesMap[npResources.NodePool.Name] = r
						continue
					}

					eNpR.Gpu.Deserved = ptr.To(*eNpR.Gpu.Deserved + npResources.Gpu.Deserved)
					*eNpR.Cpu.Deserved += *npResources.Cpu.Deserved
					*eNpR.Memory.Deserved += *npResources.Memory.Deserved
					expectedProjectAggregatedResourcesMap[npResources.NodePool.Name] = eNpR
				}
			}
			expectedProjectAggregatedResources := []api.AggregatedResources{}
			for _, r := range expectedProjectAggregatedResourcesMap {
				expectedProjectAggregatedResources = append(expectedProjectAggregatedResources, r)
			}

			Expect(getDep.Status.NodePoolQuotaStatuses).To(ConsistOf(expectedQuotaStatuses))
			Expect(getDep.Status.QuotaStatus).To(Equal(expectedQuotaStatus))
			Expect(getDep.ProjectsAggregatedResources).ToNot(BeNil())
			Expect(*getDep.ProjectsAggregatedResources).To(ConsistOf(expectedProjectAggregatedResources))

			getDepartmentsResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("name==%s", dep2Name)}, "", "verbose", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getDepartmentsResponse.Code).To(Equal(http.StatusOK))
			getDeps := getDepartmentsResponse.Body.(api.GetDepartments200Response)
			Expect(len(getDeps.Departments)).To(Equal(1))
			Expect(getDeps.Departments[0].Status.QuotaStatus).To(Equal(expectedQuotaStatus))

			getDepartmentsResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("name==%s", dep2Name)}, "", "", "", 0, 0)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getDepartmentsResponse.Code).To(Equal(http.StatusOK))
			getDeps = getDepartmentsResponse.Body.(api.GetDepartments200Response)
			Expect(len(getDeps.Departments)).To(Equal(1))
			Expect(getDeps.Departments[0].Status).To(BeNil())
		})
	})
})
