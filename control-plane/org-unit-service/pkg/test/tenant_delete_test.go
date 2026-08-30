package testing

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersOpenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	openapi2 "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	dvMocks "github.com/run-ai/runai/runai-clients/datavolumes-service-client/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project delete API tests", Ordered, func() {
	tenantID := 77
	TenantID := fmt.Sprintf("%d", tenantID)
	clusterId := uuid.New().String()
	ctx := SetContext(TenantID, TestUser)

	anotherTenant := 88
	anotherTenantId := fmt.Sprintf("%d", anotherTenant)
	anotherCtx := SetContext(anotherTenantId, TestUser)

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	dvsClient := dvMocks.IDatavolumesClient{}
	var nodeTypeApiService api.NodeTypesAPIServicer
	var projectApiService api.ProjectsAPIServicer
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To(services.ProjectDeletionMinVersion)),
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
	departmentApiService := departments.NewDepartmentApiService(&clusterServiceClient, nil)
	hardDeleteService := services.NewHardTenantDeletionService()

	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.20")),
	}, nil)
	dvsClient.On("GetDatavolumes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&openapi2.GetDatavolumes200Response{
		Datavolumes: []openapi2.DatavolumeNoSharedScopes{},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, &dvsClient)
	projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
	nodeTypeApiService = services.NewNodeTypeApiService()

	Context("test hard deletion service", func() {
		numOfTestDepartments := 5
		numOfTestProjects := 5

		anotherNumOfTestDepartments := 5
		anotherNumOfTestProjects := anotherNumOfTestDepartments
		var nodeType api.NodeType
		var anotherNodeType api.NodeType

		var departmentWithOvertimeData api.Department
		var projectWithOvertimeData api.Project
		var anotherDepartmentWithOvertimeData api.Department
		var anotherProjectWithOvertimeData api.Project

		AfterAll(func() {
			// Delete all departments and projects of Another Tenant
			err := hardDeleteService.HardDelete(context.Background(), anotherTenantId)
			Expect(err).To(BeNil())

			// Validate all deleted
			allDepartments, err := dao.GetAllTenantDepartments(anotherCtx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(len(allDepartments)).Should(BeZero())

			allProjects, err := dao.GetAllTenantProjects(anotherCtx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(len(allProjects)).Should(BeZero())

			allNodeTypes, err := dao.GetAllTenantNodeTypes(anotherCtx, TenantID)
			Expect(err).To(BeNil())
			Expect(len(allNodeTypes)).Should(BeZero())
		})
		BeforeAll(func() {
			for i := 0; i < numOfTestDepartments; i++ {
				name := fmt.Sprintf("department-%d", i)
				createDepartmentRequest := getFullDepartmentCreateRequest(name, clusterId)
				response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))
				if i == 0 {
					department := response.Body.(api.Department)
					departmentWithOvertimeData = department
					result := map[string]map[string]*models.OvertimeData{
						clusterId: {
							department.Name: &models.OvertimeData{
								AvgGpuAllocation24h:        ptr.To(float64(50)),
								AvgGpuUtilization24h:       ptr.To(float64(60)),
								AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
							},
						},
					}
					err := dao.SetDepartmentsOvertimeDataForTenant(ctx, TenantID, result, dao.RangeDay)
					Expect(err).To(BeNil())
				}
			}

			// Create Another Departments - those departments belong to AnotherTenant and should bot be deleted
			for i := 0; i < anotherNumOfTestDepartments; i++ {
				name := fmt.Sprintf("another-department-%d", i)
				createDepartmentRequest := getFullDepartmentCreateRequest(name, clusterId)
				response, err := departmentApiService.CreateDepartment(anotherCtx, createDepartmentRequest)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))
				if i == 0 {
					department := response.Body.(api.Department)
					anotherDepartmentWithOvertimeData = department
					result := map[string]map[string]*models.OvertimeData{
						clusterId: {
							department.Name: &models.OvertimeData{
								AvgGpuAllocation24h:        ptr.To(float64(50)),
								AvgGpuUtilization24h:       ptr.To(float64(60)),
								AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
							},
						},
					}
					err := dao.SetDepartmentsOvertimeDataForTenant(ctx, anotherTenantId, result, dao.RangeDay)
					Expect(err).To(BeNil())
				}
			}

			anotherNodeTypeResponse, anotherNodeTypeErr := nodeTypeApiService.CreateNodeType(anotherCtx, api.NodeTypeCreateRequest{
				Name:      "type-X",
				ClusterId: clusterId,
			})
			Expect(anotherNodeTypeErr).To(BeNil())
			anotherNodeType = anotherNodeTypeResponse.Body.(api.NodeType)

			nodeTypeResponse, nodeTypeErr := nodeTypeApiService.CreateNodeType(ctx, api.NodeTypeCreateRequest{
				Name:      "type-X",
				ClusterId: clusterId,
			})
			Expect(nodeTypeErr).To(BeNil())
			Expect(nodeType).To(Not(BeNil()))
			nodeType = nodeTypeResponse.Body.(api.NodeType)

			departments, err := dao.GetAllTenantDepartments(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(departments).To(Not(BeNil()))
			Expect(len(departments)).Should(Equal(numOfTestDepartments))

			anotherDepartments, err := dao.GetAllTenantDepartments(ctx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(anotherDepartments).To(Not(BeNil()))
			Expect(len(anotherDepartments)).Should(Equal(anotherNumOfTestDepartments))

			// Create Project for each Another Department - those projects belong to AnotherTenant and should bot be deleted
			for i := 0; i < anotherNumOfTestDepartments; i++ {
				anotherDepartment := anotherDepartments[i]
				name := createProjectName()
				request := getProjectCreateRequest(name, name, strconv.FormatInt(anotherDepartment.ID, 10), clusterId)
				request.NodeTypes = api.NodeTypesPerWorkload{
					Training:  nil,
					Workspace: nil,
				}
				if i%2 == 0 {
					request.NodeTypes.Workspace = []string{anotherNodeType.Id}
				} else {
					request.NodeTypes.Training = []string{anotherNodeType.Id}
				}
				createProjectResponse, err := projectApiService.CreateProject(anotherCtx, request)
				Expect(err).To(BeNil())
				Expect(createProjectResponse.Code).To(Equal(http.StatusCreated))
				if i == 0 {
					project := createProjectResponse.Body.(api.Project)
					anotherProjectWithOvertimeData = project
					result := map[string]map[string]*models.OvertimeData{
						clusterId: {
							project.Name: &models.OvertimeData{
								AvgGpuAllocation24h:        ptr.To(float64(50)),
								AvgGpuUtilization24h:       ptr.To(float64(60)),
								AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
							},
						},
					}
					err := dao.SetProjectsOvertimeDataForTenant(ctx, anotherTenantId, result, dao.RangeDay)
					Expect(err).To(BeNil())
				}
			}

			for _, department := range departments {
				for i := 0; i < numOfTestProjects; i++ {
					name := createProjectName()
					request := getProjectCreateRequest(name, name, strconv.FormatInt(department.ID, 10), clusterId)
					request.NodeTypes = api.NodeTypesPerWorkload{
						Training:  nil,
						Workspace: nil,
					}
					if i%2 == 0 {
						request.NodeTypes.Workspace = []string{nodeType.Id}
					} else {
						request.NodeTypes.Training = []string{nodeType.Id}
					}
					createProjectResponse, err := projectApiService.CreateProject(ctx, request)
					Expect(err).To(BeNil())
					Expect(createProjectResponse.Code).To(Equal(http.StatusCreated))
					if i == 0 {
						project := createProjectResponse.Body.(api.Project)
						projectWithOvertimeData = project
						result := map[string]map[string]*models.OvertimeData{
							clusterId: {
								project.Name: &models.OvertimeData{
									AvgGpuAllocation24h:        ptr.To(float64(50)),
									AvgGpuUtilization24h:       ptr.To(float64(60)),
									AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
								},
							},
						}
						err := dao.SetProjectsOvertimeDataForTenant(ctx, TenantID, result, dao.RangeDay)
						Expect(err).To(BeNil())
					}
				}
			}
		})

		It("Delete Tenant", func() {
			// Validate the Projects, Departments and nodeTypes exists before the deletion
			allProjects, err := dao.GetAllTenantProjects(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allProjects).To(Not(BeNil()))
			Expect(len(allProjects)).Should(Equal(numOfTestDepartments * numOfTestProjects))
			projectId, err := strconv.ParseInt(projectWithOvertimeData.Id, 10, 64)
			Expect(err).To(BeNil())
			project, err := dao.GetProjectById(ctx, projectId, TenantID, models.VerboseVerbosity)
			validateOrgUnitOvertimeData(project.OrgUnitData)

			allDepartments, err := dao.GetAllTenantDepartments(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allDepartments).To(Not(BeNil()))
			Expect(len(allDepartments)).Should(Not(BeZero()))
			departmentId, err := strconv.ParseInt(departmentWithOvertimeData.Id, 10, 64)
			Expect(err).To(BeNil())
			department, err := dao.GetDepartmentById(ctx, departmentId, TenantID, models.VerboseVerbosity, nil)
			validateOrgUnitOvertimeData(department.OrgUnitData)

			allNodeTypes, err := dao.GetAllTenantNodeTypes(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allNodeTypes).To(Not(BeNil()))
			Expect(len(allNodeTypes)).Should(Not(BeZero()))

			// Delete all departments and projects
			err = hardDeleteService.HardDelete(context.Background(), TenantID)
			Expect(err).To(BeNil())

			// Validate the Projects, Departments and nodeTypes DOES NOT exist after the deletion
			allDepartments, err = dao.GetAllTenantDepartments(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allDepartments).To(BeNil())
			Expect(len(allDepartments)).Should(BeZero())

			allProjects, err = dao.GetAllTenantProjects(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allProjects).To(BeNil())
			Expect(len(allProjects)).Should(BeZero())

			allNodeTypes, err = dao.GetAllTenantNodeTypes(ctx, TenantID)
			Expect(err).To(BeNil())
			Expect(allNodeTypes).To(BeNil())
			Expect(len(allNodeTypes)).Should(BeZero())

			// Validate the Projects, Departments and nodeTypes DOES NOT exist after the deletion
			allDepartments, err = dao.GetAllTenantDepartments(anotherCtx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(allDepartments).To(Not(BeNil()))
			Expect(len(allDepartments)).Should(Equal(anotherNumOfTestDepartments))

			allProjects, err = dao.GetAllTenantProjects(anotherCtx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(allProjects).To(Not(BeNil()))
			Expect(len(allProjects)).Should(Equal(anotherNumOfTestProjects))

			allNodeTypes, err = dao.GetAllTenantNodeTypes(anotherCtx, anotherTenantId)
			Expect(err).To(BeNil())
			Expect(len(allNodeTypes)).Should(Equal(1))

			departmentId, err = strconv.ParseInt(anotherDepartmentWithOvertimeData.Id, 10, 64)
			Expect(err).To(BeNil())
			department, err = dao.GetDepartmentById(ctx, departmentId, anotherTenantId, models.VerboseVerbosity, nil)
			validateOrgUnitOvertimeData(department.OrgUnitData)

			projectId, err = strconv.ParseInt(anotherProjectWithOvertimeData.Id, 10, 64)
			Expect(err).To(BeNil())
			project, err = dao.GetProjectById(ctx, projectId, anotherTenantId, models.VerboseVerbosity)
			validateOrgUnitOvertimeData(project.OrgUnitData)
		})
	})
})

func validateOrgUnitOvertimeData(orgUnitData models.OrgUnitData) {
	Expect(orgUnitData.OvertimeData.AvgGpuAllocation24h).ToNot(BeNil())
	Expect(*orgUnitData.OvertimeData.AvgGpuAllocation24h).To(Equal(float64(50)))
	Expect(orgUnitData.OvertimeData.AvgGpuUtilization24h).ToNot(BeNil())
	Expect(*orgUnitData.OvertimeData.AvgGpuUtilization24h).To(Equal(float64(60)))
	Expect(orgUnitData.OvertimeData.AvgGpuMemoryUtilization24h).ToNot(BeNil())
	Expect(*orgUnitData.OvertimeData.AvgGpuMemoryUtilization24h).To(Equal(float64(70)))
}
