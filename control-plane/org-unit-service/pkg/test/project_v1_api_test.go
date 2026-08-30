package testing

import (
	"fmt"
	"net/http"
	"time"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"

	clustersOpenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	v1 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/v1"
)

var _ = Describe("Project v1 API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var projectIds []string
	var nodeTypeIds []string
	var departmentApiService api.DepartmentsAPIServicer
	var projectV1ApiService api.ProjectsV1APIServicer
	var projectApiService api.ProjectsAPIServicer
	var nodeTypesApiService api.NodeTypesAPIServicer

	var defaultDepartmentId string

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.17")),
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

	BeforeAll(func() {
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		nodeTypesApiService = services.NewNodeTypeApiService()
		defaultDepartment, err := departmentApiService.CreateDepartment(ctx, getDefaultDepartmentV1Request(clusterId))
		if err != nil || defaultDepartment.Code != http.StatusCreated {
			Fail("Failed creating default department")
		}
		defaultDepartmentId = defaultDepartment.Body.(api.Department).Id
	})

	Context("cpu flag on, oq flag off", func() {

		BeforeEach(func() {
			clusterServiceClient := mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.17")),
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
			projectDeletion := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

			departmentApiService = api.NewDepartmentsAPIService()
			projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, projectDeletion)
			projectApiService = projects.NewProjectApiService(projectDeletion, &clusterServiceClient, nil)
			for i := 0; i < 2; i++ {
				nodeType, err := nodeTypesApiService.CreateNodeType(ctx, api.NodeTypeCreateRequest{
					ClusterId: clusterId,
					Name:      fmt.Sprintf("type%d", i),
				})
				Expect(err).To(BeNil())
				Expect(nodeType.Code).To(Equal(http.StatusCreated))
				nodeTypeIds = append(nodeTypeIds, nodeType.Body.(api.NodeType).Id)
			}
		})
		AfterEach(func() {
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, fmt.Sprint(projectId), "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
			for _, nodeTypeId := range nodeTypeIds {
				response, err := nodeTypesApiService.DeleteNodeType(ctx, fmt.Sprint(nodeTypeId))
				if err != nil || response.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			nodeTypeIds = []string{}
		})

		It("test update status", func() {
			clusterServiceClient := mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.17")),
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
						Name:                   "pool1",
						Id:                     "2",
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectDeletion := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
			projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, projectDeletion)
			projectApiService = projects.NewProjectApiService(projectDeletion, &clusterServiceClient, nil)

			projectName := createProjectName()
			request := getProjectV1RequestWithNp(projectName, defaultDepartmentId, api.ResourcesNodePool{
				Id:   "2",
				Name: "pool1",
			}, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, (request))
			if err != nil {
				Fail("Failed creating project")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Project)
			projectId := created.Id
			projectIds = append(projectIds, projectId)
			Expect(created.ClusterId).To(Equal(clusterId))

			quotaStatus := []api.QuotaStatusV1Inner{{
				NodePoolName: "default",
				Allocated: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				Requested: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    1,
					Memory: 1000,
				},
			}, {
				NodePoolName: "pool1",
				Allocated: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				AllocatedNonPreemptible: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
				Requested: api.QuotaStatusResourceListV1{
					Gpu:    1,
					Cpu:    2,
					Memory: 3000,
				},
			}}
			Expect(err).To(BeNil())
			newStatus := "{\"namespace\":\"runai-proj1\",\"phase\":\"Ready\",\"message\":\"some message 2\",\"conditions\":[{\"type\":\"NamespaceReady\",\"lastProbeTime\":null,\"lastTransitionTime\":\"2024-02-15T14:58:19Z\"}],\"quotaStatus\":{}}"

			updateResponse, err := projectV1ApiService.UpdateProjectStatusV1(ctx, clusterId, api.ProjectV1StatusRequest{
				Status:         newStatus,
				Name:           projectName,
				DepartmentName: "default",
				QuotaStatuses:  quotaStatus,
			})
			if err != nil {
				Fail("Failed updating project status")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			var nilTime *time.Time = nil

			projIdStr := projectId

			getResponseV2, err := projectApiService.GetProject(ctx, projIdStr)
			if err != nil {
				Fail("Failed getting project v2")
			}
			getV2 := getResponseV2.Body.(api.Project)
			Expect(string(getV2.Status.Phase)).To(Equal("Ready"))
			Expect(getV2.Status.PhaseMessage).To(Equal(ptr.To("some message 2")))
			Expect(getV2.Status.Namespace).To(Equal(ptr.To("runai-proj1")))
			Expect(getV2.Status.QuotaStatus).To(Equal(api.QuotaStatus{
				Allocated: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4194.304,
				},
				AllocatedNonPreemptible: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    4,
					Memory: 6291.456,
				},
				Requested: api.QuotaStatusResource{
					Gpu:    2,
					Cpu:    3,
					Memory: 4194.304,
				},
			}))
			Expect(getV2.Status.NodePoolQuotaStatuses).To(ConsistOf([]api.QuotaStatusNodePool{
				{
					NodePoolName: "default",
					NodePoolId:   "1",
					Allocated: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    1,
						Memory: 1048.576,
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    2,
						Memory: 3145.728,
					},
					Requested: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    1,
						Memory: 1048.576,
					},
				}, {
					NodePoolName: "pool1",
					NodePoolId:   "2",
					Allocated: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    2,
						Memory: 3145.728,
					},
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    2,
						Memory: 3145.728,
					},
					Requested: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    2,
						Memory: 3145.728,
					},
				},
			}))
			Expect(getV2.Status.AdditionalStatusData).ToNot(BeNil())
			Expect(*getV2.Status.AdditionalStatusData).To(Equal(map[string]interface{}{
				"conditions": models.ConditionArray{
					{
						Type:               "NamespaceReady",
						LastTransitionTime: ptr.To(time.Date(2024, 2, 15, 14, 58, 19, 0, time.UTC)),
						LastProbeTime:      nilTime,
					},
				},
			}))
		})
	})

	Context("test get projects", func() {

		BeforeAll(func() {
			clusterServiceClient := mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.17")),
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
			projectDeletion := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

			departmentApiService = api.NewDepartmentsAPIService()
			projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, projectDeletion)
			projectApiService = projects.NewProjectApiService(projectDeletion, &clusterServiceClient, nil)

			projectIds = []string{}

			projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, projectDeletion)

			for i := 0; i < 3; i++ {
				// create a project
				request := getProjectV1Request(createProjectName(), defaultDepartmentId, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				if err != nil {
					Fail("Failed creating project")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				projectIds = append(projectIds, createResponse.Body.(api.Project).Id)
			}
		})
		AfterAll(func() {
			toMockAccessrulesDepId = nil

			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, fmt.Sprint(projectId), "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting project")
				}
			}
			projectIds = []string{}
		})

		It("test get projects with permissions", func() {
			toMockAccessrulesDepId = ptr.To(defaultDepartmentId)
			res, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, false, true)
			if err != nil {
				Fail("Failed getting projects")
			}
			Expect(res.Code).To(Equal(http.StatusOK), "failed getting projects: %v", res.Body)

			projects := res.Body.([]api.ProjectV1)
			Expect(len(projects)).To(Equal(3))

			for _, p := range projects {
				Expect(p.Permissions.Users).ToNot(BeNil())
				Expect(len(*p.Permissions.Users)).To(Equal(2))
				Expect(p.Permissions.Groups).ToNot(BeNil())
				Expect(len(*p.Permissions.Groups)).To(Equal(1))
			}
		})
	})
})

func getProjectV1Request(name string, departmentId string, clusterId string) api.ProjectCreationRequest {
	resources := createDefaultApiV1NodePoolResources(false, []api.ResourcesNodePool{{Id: "1", Name: "default"}})
	return api.ProjectCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		ParentId:  ptr.To(departmentId),
		Resources: resources,
	}
}

func getProjectV1RequestWithNp(name string, departmentId string, pool api.ResourcesNodePool, clusterId string) api.ProjectCreationRequest {
	resources := createDefaultApiV1NodePoolResources(false, []api.ResourcesNodePool{{Id: "1", Name: "default"}, pool})
	return api.ProjectCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		ParentId:  ptr.To(departmentId),
		Resources: resources,
	}
}

func getDefaultDepartmentV1Request(clusterId string) api.DepartmentCreationRequest {
	resources := createDefaultApiV1NodePoolResources(false, []api.ResourcesNodePool{{Id: "1", Name: "default"}})
	return api.DepartmentCreationRequest{
		Resources: resources,
		Name:      "default",
		ClusterId: clusterId,
	}
}
