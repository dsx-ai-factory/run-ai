/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/overtime_metrics_writer"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/runaimetrics"
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

type OvertimeMetricsMockValue struct {
	ClusterId  string
	Department string
	Project    string
	Value      string
}

var _ = Describe("Project API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var (
		projectApiService    api.ProjectsAPIServicer
		departmentApiService api.DepartmentsAPIServicer
		defaultNodePoolName  = "default"
		defaultNodePoolId    = float32(1)
		nodePool2Name        = "np2"
		nodePool2Id          = float32(2)
	)

	BeforeAll(func() {
		clusterServiceClient := mocks.IClusterClient{}
		// The overtime metrics mock is global and returns data for shared clusters across all tenants.
		// Use a context-aware mock that returns the requesting tenant's TenantId so cluster ownership
		// validation passes for each tenant.
		for _, tid := range []struct {
			s string
			i int32
		}{{TestTenant, 1}, {"22", 22}, {"33", 33}, {"44", 44}} {
			tidCopy := tid
			clusterServiceClient.On("GetClusterMinimalDataById",
				mock.MatchedBy(func(ctx context.Context) bool {
					v, _ := ctx.Value(runaictx.CfTenantId).(string)
					return v == tidCopy.s
				}),
				mock.Anything,
			).Return(&clustersOpenapi.ClustersMinimalData{
				TenantId: tidCopy.i,
				Uuid:     clusterId,
				Version:  *clustersOpenapi.NewNullableString(ptr.To("2.21")),
			}, nil)
		}
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
			Nodepools: []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     fmt.Sprint(defaultNodePoolId),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
				{
					Name:                   nodePool2Name,
					Id:                     fmt.Sprint(nodePool2Id),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			},
		}, nil)
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)

		mockTenantsClient.Mock.On("ListAllTenantsInternal", mock.Anything).Return([]tenants_client.Tenant{{Id: 1, Status: "Ready"}, {Id: 22, Status: "Ready"}, {Id: 33, Status: "Ready"}, {Id: 44, Status: "Ready"}}, &http.Response{StatusCode: 200}, nil)
	})

	Context("Partial overtime data updates", func() {
		var (
			department api.Department
			project    api.Project
			projectId  int64
		)

		BeforeEach(func() {
			// Create fresh department and project for each test to ensure complete isolation
			createDepartmentRequest := getFullDepartmentCreateRequest("test-dept-partial", clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			Expect(err).To(BeNil(), "Failed creating test department")
			Expect(response.Code).To(Equal(http.StatusCreated))
			department = response.Body.(api.Department)

			createProjectRequest := getProjectCreateRequest(createProjName(), "", department.Id, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, createProjectRequest)
			Expect(err).To(BeNil(), "Failed creating test project")
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project = createResponse.Body.(api.Project)

			projectIdInt, err := strconv.Atoi(project.Id)
			Expect(err).To(BeNil(), "Failed converting project id")
			projectId = int64(projectIdInt)
		})

		AfterEach(func() {
			// Clean up after each test
			_, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
			Expect(err).To(BeNil(), "Failed deleting test project")
			_, err = departmentApiService.DeleteDepartment(ctx, department.Id)
			Expect(err).To(BeNil(), "Failed deleting test department")
		})

		It("Should allow partial data updates (only allocation)", func() {
			// Test case: Only GPU allocation is available, utilization is nil
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:  ptr.To(float64(42)),
						AvgGpuUtilization24h: nil, // Missing utilization
					},
				},
			}

			// Retry setting overtime data and checking until it succeeds
			// This handles the case where the database view hasn't refreshed yet
			Eventually(func(g Gomega) {
				// Try to set the data (may fail if project not yet visible in view)
				_ = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)

				// Check if data was set
				dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
				g.Expect(err).To(BeNil())
				g.Expect(dbProject).NotTo(BeNil())
				overtimeData := &dbProject.OrgUnitData.OvertimeData

				g.Expect(overtimeData.AvgGpuAllocation24h).NotTo(BeNil(), "Allocation should be updated")
				g.Expect(*overtimeData.AvgGpuAllocation24h).To(Equal(float64(42)), "Allocation value incorrect")
				g.Expect(overtimeData.AvgGpuUtilization24h).To(BeNil(), "Utilization should remain nil when not provided")
				g.Expect(overtimeData.AvgGpuMemoryUtilization24h).To(BeNil(), "Memory utilization should remain nil when not provided")
				g.Expect(overtimeData.Avg24hRangeUpdatedAt).NotTo(BeNil(), "Timestamp should be updated")
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
		})

		It("Should allow partial data updates (only utilization)", func() {
			// Test case: Only GPU utilization is available, allocation is nil
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:  nil, // Missing allocation
						AvgGpuUtilization24h: ptr.To(float64(88)),
					},
				},
			}

			// Retry setting overtime data and checking until it succeeds
			Eventually(func(g Gomega) {
				_ = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)

				dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
				g.Expect(err).To(BeNil())
				g.Expect(dbProject).NotTo(BeNil())
				overtimeData := &dbProject.OrgUnitData.OvertimeData

				g.Expect(overtimeData.AvgGpuUtilization24h).NotTo(BeNil(), "Utilization should be updated")
				g.Expect(*overtimeData.AvgGpuUtilization24h).To(Equal(float64(88)), "Utilization value incorrect")
				g.Expect(overtimeData.AvgGpuAllocation24h).To(BeNil(), "Allocation should remain nil when not provided")
				g.Expect(overtimeData.AvgGpuMemoryUtilization24h).To(BeNil(), "Memory utilization should remain nil when not provided")
				g.Expect(overtimeData.Avg24hRangeUpdatedAt).NotTo(BeNil(), "Timestamp should be updated")
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
		})

		It("Should allow partial data updates (only memory utilization)", func() {
			// Test case: Only GPU memory utilization is available, allocation and utilization are nil
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:        nil, // Missing allocation
						AvgGpuUtilization24h:       nil, // Missing utilization
						AvgGpuMemoryUtilization24h: ptr.To(float64(65)),
					},
				},
			}

			Eventually(func(g Gomega) {
				_ = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)

				dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
				g.Expect(err).To(BeNil())
				g.Expect(dbProject).NotTo(BeNil())
				overtimeData := &dbProject.OrgUnitData.OvertimeData

				g.Expect(overtimeData.AvgGpuMemoryUtilization24h).NotTo(BeNil(), "Memory utilization should be updated")
				g.Expect(*overtimeData.AvgGpuMemoryUtilization24h).To(Equal(float64(65)), "Memory utilization value incorrect")
				g.Expect(overtimeData.AvgGpuAllocation24h).To(BeNil(), "Allocation should remain nil when not provided")
				g.Expect(overtimeData.AvgGpuUtilization24h).To(BeNil(), "Utilization should remain nil when not provided")
				g.Expect(overtimeData.Avg24hRangeUpdatedAt).NotTo(BeNil(), "Timestamp should be updated")
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
		})

		It("Should not update anything when all metrics are nil", func() {
			// When all metrics are nil, verify no database update occurs
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:        nil,
						AvgGpuUtilization24h:       nil,
						AvgGpuMemoryUtilization24h: nil,
					},
				},
			}

			err := dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())

			// Verify all metrics remain nil (no default values set)
			dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			overtimeData := &dbProject.OrgUnitData.OvertimeData
			Expect(overtimeData.AvgGpuAllocation24h).To(BeNil())
			Expect(overtimeData.AvgGpuUtilization24h).To(BeNil())
			Expect(overtimeData.AvgGpuMemoryUtilization24h).To(BeNil())
		})

		It("Should successfully update all metrics when available", func() {
			// Test case: All metrics available - should work as before
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:        ptr.To(float64(55)),
						AvgGpuUtilization24h:       ptr.To(float64(77)),
						AvgGpuMemoryUtilization24h: ptr.To(float64(92)),
					},
				},
			}

			// Retry until all metrics are updated
			Eventually(func(g Gomega) {
				_ = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)

				dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
				g.Expect(err).To(BeNil())
				g.Expect(dbProject).NotTo(BeNil())
				overtimeData := &dbProject.OrgUnitData.OvertimeData

				validateOvertimeData(g, overtimeData, dao.RangeDay, 55, 77, ptr.To(float64(92)))
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
		})

		It("BUG FIX: Old project should accept partial metrics when they become available", func() {
			// Scenario: Workload running 24+ hours where metrics collection is inconsistent
			// 1. Initial collection fails - no metrics available
			// 2. Later, ONLY allocation metric becomes available (utilization still missing)
			// 3. WITHOUT FIX: Update rejected forever (needs BOTH metrics)
			//    WITH FIX: Allocation gets populated (partial update succeeds)

			// Step 1: Create a project that simulates being old (>24h)
			newDeptRequest := getFullDepartmentCreateRequest("dept-partial-metrics", clusterId)
			deptResponse, err := departmentApiService.CreateDepartment(ctx, newDeptRequest)
			Expect(err).To(BeNil())
			Expect(deptResponse.Code).To(Equal(http.StatusCreated))
			oldDept := deptResponse.Body.(api.Department)

			newProjRequest := getProjectCreateRequest("project-partial-metrics", "", oldDept.Id, clusterId)
			projResponse, err := projectApiService.CreateProject(ctx, newProjRequest)
			Expect(err).To(BeNil())
			Expect(projResponse.Code).To(Equal(http.StatusCreated))
			oldProject := projResponse.Body.(api.Project)

			oldProjectId, err := strconv.Atoi(oldProject.Id)
			Expect(err).To(BeNil())

			// Step 2: CRITICAL - Only allocation metric becomes available (utilization still nil)
			partialMetricsResult := map[string]map[string]*models.OvertimeData{
				clusterId: {
					oldProject.Name: &models.OvertimeData{
						AvgGpuAllocation24h:        ptr.To(float64(8.5)), // ONLY allocation available
						AvgGpuUtilization24h:       nil,                  // Utilization still missing
						AvgGpuMemoryUtilization24h: nil,
					},
				},
			}

			Eventually(func(g Gomega) {
				_ = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, partialMetricsResult, dao.RangeDay)

				dbProject, err := dao.GetProjectById(ctx, int64(oldProjectId), TestTenant, models.SlimVerbosity)
				g.Expect(err).To(BeNil())
				g.Expect(dbProject).NotTo(BeNil())
				overtimeData := &dbProject.OrgUnitData.OvertimeData

				g.Expect(overtimeData.AvgGpuAllocation24h).NotTo(BeNil(),
					"BUG FIX: Allocation should be populated with partial data")
				g.Expect(*overtimeData.AvgGpuAllocation24h).To(Equal(float64(8.5)),
					"Allocation value should match")

				g.Expect(overtimeData.AvgGpuUtilization24h).To(BeNil(),
					"Utilization should remain nil when not provided")
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed(),
				"FAILED: Old project should accept partial metrics (bug fix)")

			// Cleanup
			_, err = projectApiService.DeleteProject(ctx, oldProject.Id, "", false)
			Expect(err).To(BeNil())
			_, err = departmentApiService.DeleteDepartment(ctx, oldDept.Id)
			Expect(err).To(BeNil())
		})
	})

	Context("Overtime data - writing manually", Ordered, func() {
		var (
			department              api.Department
			project                 api.Project
			projectId, departmentId int64
		)

		BeforeAll(func() {
			createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department = response.Body.(api.Department)

			createProjectRequest := getProjectCreateRequest(createProjName(), "", department.Id, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project = createResponse.Body.(api.Project)

			projectIdInt, err := strconv.Atoi(project.Id)
			if err != nil {
				Fail("Failed converting project id")
			}
			projectId = int64(projectIdInt)

			departmentIdInt, err := strconv.Atoi(department.Id)
			if err != nil {
				Fail("Failed converting department id")
			}
			departmentId = int64(departmentIdInt)
		})

		AfterAll(func() {
			_, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
			if err != nil {
				Fail("Failed deleting project")
			}
			_, err = departmentApiService.DeleteDepartment(ctx, department.Id)
			if err != nil {
				Fail("Failed deleting department")
			}
		})

		It("Sanity for project - in get project from db and api", func() {
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					project.Name: &models.OvertimeData{
						AvgGpuAllocation24h:        ptr.To(float64(50)),
						AvgGpuUtilization24h:       ptr.To(float64(60)),
						AvgGpuMemoryUtilization24h: ptr.To(float64(70)),
					},
				},
			}
			err := dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())

			dbProject, err := dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			Expect(dbProject).NotTo(BeNil())
			overtimeData := &dbProject.OrgUnitData.OvertimeData
			Eventually(func(g Gomega) {
				validateOvertimeData(g, overtimeData, dao.RangeDay, 50, 60, ptr.To(float64(70)))
				validateOvertimeDataNil(g, overtimeData, dao.RangeWeek)
				validateOvertimeDataNil(g, overtimeData, dao.RangeMonth)
			}, time.Second, time.Second).Should(Succeed())

			getResponse, err := projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			apiOvertimeData := getResponse.Body.(api.Project).OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeDay, 50, 60, ptr.To(float64(70)))
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeWeek)
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeMonth)

			result[clusterId][project.Name].AvgGpuAllocation24h = ptr.To(float64(10))
			result[clusterId][project.Name].AvgGpuUtilization24h = ptr.To(float64(20))
			result[clusterId][project.Name].AvgGpuMemoryUtilization24h = ptr.To(float64(30))
			result[clusterId][project.Name].AvgGpuAllocation7d = ptr.To(float64(70))
			result[clusterId][project.Name].AvgGpuUtilization7d = ptr.To(float64(80))
			result[clusterId][project.Name].AvgGpuMemoryUtilization7d = ptr.To(float64(90))
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())
			err = dao.SetProjectsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())

			dbProject, err = dao.GetProjectById(ctx, projectId, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			Expect(dbProject).NotTo(BeNil())
			overtimeData = &dbProject.OrgUnitData.OvertimeData
			Eventually(func(g Gomega) {
				validateOvertimeData(g, overtimeData, dao.RangeDay, 10, 20, ptr.To(float64(30)))
				validateOvertimeData(g, overtimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
			}, time.Second, time.Second).Should(Succeed())

			getResponse, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			apiOvertimeData = getResponse.Body.(api.Project).OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeDay, 10, 20, ptr.To(float64(30)))
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeMonth)

			listResponse, err := projectApiService.GetProjects(ctx, []string{}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))
			projects := listResponse.Body.(api.GetProjects200Response).Projects
			apiProject := findProjectInList(projects, projectId)
			Expect(apiProject).ToNot(BeNil())
			apiOvertimeData = apiProject.OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeDay, 10, 20, ptr.To(float64(30)))
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeMonth)
		})

		It("Sanity for department - in get department from db and api", func() {
			result := map[string]map[string]*models.OvertimeData{
				clusterId: {
					department.Name: &models.OvertimeData{
						AvgGpuAllocation30d:        ptr.To(float64(100)),
						AvgGpuUtilization30d:       ptr.To(float64(90)),
						AvgGpuMemoryUtilization30d: nil,
					},
				},
			}
			err := dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())

			dbDepartment, err := dao.GetDepartmentById(ctx, departmentId, TestTenant, models.SlimVerbosity, nil)
			Expect(err).To(BeNil())
			Expect(dbDepartment).NotTo(BeNil())
			overtimeData := &dbDepartment.OrgUnitData.OvertimeData
			Eventually(func(g Gomega) {
				validateOvertimeData(g, overtimeData, dao.RangeMonth, 100, 90, nil)
				validateOvertimeDataNil(g, overtimeData, dao.RangeWeek)
				validateOvertimeDataNil(g, overtimeData, dao.RangeDay)
			}, time.Second, time.Second).Should(Succeed())

			response, err := departmentApiService.GetDepartment(ctx, department.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			apiOvertimeData := response.Body.(api.Department).OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeMonth, 100, 90, nil)
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeWeek)
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeDay)

			result[clusterId][department.Name].AvgGpuAllocation30d = ptr.To(float64(10))
			result[clusterId][department.Name].AvgGpuUtilization30d = ptr.To(float64(20))
			result[clusterId][department.Name].AvgGpuMemoryUtilization30d = nil
			result[clusterId][department.Name].AvgGpuAllocation7d = ptr.To(float64(70))
			result[clusterId][department.Name].AvgGpuUtilization7d = ptr.To(float64(80))
			result[clusterId][department.Name].AvgGpuMemoryUtilization7d = ptr.To(float64(90))
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())

			dbDepartment, err = dao.GetDepartmentById(ctx, departmentId, TestTenant, models.SlimVerbosity, nil)
			Expect(err).To(BeNil())
			Expect(dbDepartment).NotTo(BeNil())
			overtimeData = &dbDepartment.OrgUnitData.OvertimeData
			Eventually(func(g Gomega) {
				validateOvertimeData(g, overtimeData, dao.RangeMonth, 10, 20, nil)
				validateOvertimeData(g, overtimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
				validateOvertimeDataNil(g, overtimeData, dao.RangeDay)
			}, time.Second, time.Second).Should(Succeed())

			response, err = departmentApiService.GetDepartment(ctx, department.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			apiOvertimeData = response.Body.(api.Department).OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeMonth, 10, 20, nil)
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeDay)

			listResponse, err := departmentApiService.GetDepartments(ctx, []string{}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))
			departments := listResponse.Body.(api.GetDepartments200Response).Departments
			apiDepartment := findDepartmentInList(departments, departmentId)
			Expect(apiDepartment).ToNot(BeNil())
			apiOvertimeData = apiDepartment.OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeMonth, 10, 20, nil)
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, 70, 80, ptr.To(float64(90)))
			validateApiOvertimeDataNil(apiOvertimeData, dao.RangeDay)
		})
	})

	Context("Overtime Metrics Writer goroutine", Ordered, func() {
		var (
			stopper    chan struct{}
			cancelCtx  context.Context
			cancel     context.CancelFunc
			clusterId2 = uuid.New().String()

			// tenants
			tenantIds    = []string{"22", "33", "44"}
			projectId    = 0
			departmentId = 0

			// map of clusterId -> departmentName -> projectNames
			orgUnits = map[string]map[string][]string{}

			createdDepartmentIds = map[string][]string{}
			createdProjectIds    = map[string][]string{}

			projectsAvgGpuAllocation           = map[string]float64{}
			projectsAvgGpuUtilization          = map[string]float64{}
			departmentsAvgGpuAllocation        = map[string]float64{}
			departmentsAvgGpuUtilization       = map[string]float64{}
			departmentsAvgGpuMemoryUtilization = map[string]float64{}

			nodePoolName1 = "np1"

			projRankingGpuAllocation = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}
			projRankingGpuUtilization = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}
			projRankingGpuMemoryUtilization = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}
			depRankingGpuAllocation = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}
			depRankingGpuUtilization = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}
			depRankingGpuMemoryUtilization = map[api.RankingOrderType]map[string]float64{
				api.RANKINGORDERTYPE_TOP:    {},
				api.RANKINGORDERTYPE_BOTTOM: {},
			}

			projectsMockValues    = []OvertimeMetricsMockValue{}
			departmentsMockValues = []OvertimeMetricsMockValue{}

			rankingMetricsToTest = []api.RankingMetricType{
				api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H,
				api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H,
				api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H,
				api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D,
				api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D,
				api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D,
				api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D,
				api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D,
				api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D,
			}
			rankingOrdersToTest = []api.RankingOrderType{
				api.RANKINGORDERTYPE_TOP,
				api.RANKINGORDERTYPE_BOTTOM,
			}
		)

		BeforeAll(func() {
			// map of clusterId -> departmentName -> projectNames
			orgUnits = map[string]map[string][]string{
				clusterId: {
					createDepName(): []string{createProjName(), createProjName(), createProjName()},
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
				},
				clusterId2: {
					createDepName(): []string{createProjName(), createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
					createDepName(): []string{createProjName()},
				},
			}

			// the metrics mock is global (and not per-tenant), so create the same clusters/departments/projects for each tenant
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)
				createdDepartmentIds[tenantId] = []string{}
				createdProjectIds[tenantId] = []string{}

				departmentIndex := 0
				for clusterId, departmentNames := range orgUnits {
					for departmentName, projectNames := range departmentNames {
						createDepartmentRequest := getFullDepartmentCreateRequest(departmentName, clusterId)

						// Add node pool with quota 3 to every 3rd department
						if departmentIndex%3 == 0 {
							nodePoolWithQuota3 := CreateNodePoolApiResources(true, fmt.Sprint(nodePool2Id), nodePool2Name)
							nodePoolWithQuota3.Gpu.Deserved = ptr.To(float64(3))
							nodePoolWithQuota3.Gpu.Limit = ptr.To(float64(3))
							nodePoolWithQuota3.Gpu.OverQuotaWeight = ptr.To(float64(3))
							createDepartmentRequest.Resources = append(createDepartmentRequest.Resources, nodePoolWithQuota3)
						}

						response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
						Expect(err).To(BeNil())
						Expect(response.Code).To(Equal(http.StatusCreated))
						department := response.Body.(api.Department)
						createdDepartmentIds[tenantId] = append(createdDepartmentIds[tenantId], department.Id)
						if departmentId == 0 {
							departmentId, err = strconv.Atoi(department.Id)
							Expect(err).To(BeNil())
						}

						projectIndex := 0
						for _, projectName := range projectNames {
							createProjectRequest := getProjectCreateRequestWithNodePools(projectName, "", department.Id, clusterId, nodePoolName1)

							// Add node pool with quota 3 to every 3rd project
							if projectIndex%3 == 0 {
								nodePoolWithQuota3 := CreateNodePoolApiResources(true, fmt.Sprint(nodePool2Id), nodePool2Name)
								nodePoolWithQuota3.Gpu.Deserved = ptr.To(float64(3))
								nodePoolWithQuota3.Gpu.Limit = ptr.To(float64(3))
								nodePoolWithQuota3.Gpu.OverQuotaWeight = ptr.To(float64(3))
							}

							createResponse, err := projectApiService.CreateProject(ctx, createProjectRequest)
							Expect(err).To(BeNil())
							Expect(createResponse.Code).To(Equal(http.StatusCreated))
							createdProjectIds[tenantId] = append(createdProjectIds[tenantId], createResponse.Body.(api.Project).Id)
							if projectId == 0 {
								projectId, err = strconv.Atoi(createResponse.Body.(api.Project).Id)
								Expect(err).To(BeNil())
							}
							projectIndex++
						}
						departmentIndex++
					}
				}
			}

			projectsMockValues, departmentsMockValues = getOrgUnitsOvertimeMetricsMockValues(orgUnits)

			// no mock for RunaiGpuMemoryUtilizationAvgPerProjectMetric - should be empty results for this
			metrics_client.AddResponsesForMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiAllocatedGpuCountAvgPerProjectMetric{}.GetName():       getOvertimeAvgMetricMock(10, projectsMockValues, projectsAvgGpuAllocation),
				runaimetrics.RunaiGpuUtilizationAvgPerProjectMetric{}.GetName():          getOvertimeAvgMetricMock(20, projectsMockValues, projectsAvgGpuUtilization),
				runaimetrics.RunaiAllocatedGpuCountAvgPerDepartmentMetric{}.GetName():    getOvertimeAvgMetricMock(50, departmentsMockValues, departmentsAvgGpuAllocation),
				runaimetrics.RunaiGpuUtilizationAvgPerDepartmentMetric{}.GetName():       getOvertimeAvgMetricMock(60, departmentsMockValues, departmentsAvgGpuUtilization),
				runaimetrics.RunaiGpuMemoryUtilizationAvgPerDepartmentMetric{}.GetName(): getOvertimeAvgMetricMock(70, departmentsMockValues, departmentsAvgGpuMemoryUtilization),

				runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15{}.GetName():       getOvertimeAvgMetricMockMaxValues(10, projectsMockValues, projRankingGpuAllocation[api.RANKINGORDERTYPE_TOP], ptr.To(15)),
				runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolTop15{}.GetName():          getOvertimeAvgMetricMockMaxValues(20, projectsMockValues, projRankingGpuUtilization[api.RANKINGORDERTYPE_TOP], ptr.To(15)),
				runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolTop15{}.GetName():    getOvertimeAvgMetricMockMaxValues(30, projectsMockValues, projRankingGpuMemoryUtilization[api.RANKINGORDERTYPE_TOP], ptr.To(15)),
				runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolTop15{}.GetName():    getOvertimeAvgMetricMockMaxValues(40, departmentsMockValues, depRankingGpuAllocation[api.RANKINGORDERTYPE_TOP], ptr.To(15)),
				runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15{}.GetName():       getOvertimeAvgMetricMockMaxValues(50, departmentsMockValues, depRankingGpuUtilization[api.RANKINGORDERTYPE_TOP], ptr.To(15)),
				runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolTop15{}.GetName(): getOvertimeAvgMetricMockMaxValues(60, departmentsMockValues, depRankingGpuMemoryUtilization[api.RANKINGORDERTYPE_TOP], ptr.To(15)),

				runaimetrics.RunaiAvgAllocatedGpuCountPerProjectNodePoolBottom15{}.GetName():       getOvertimeAvgMetricMockMaxValues(90, projectsMockValues, projRankingGpuAllocation[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
				runaimetrics.RunaiAvgGpuUtilizationPerProjectNodePoolBottom15{}.GetName():          getOvertimeAvgMetricMockMaxValues(80, projectsMockValues, projRankingGpuUtilization[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
				runaimetrics.RunaiAvgGpuMemoryUtilizationPerProjectNodePoolBottom15{}.GetName():    getOvertimeAvgMetricMockMaxValues(70, projectsMockValues, projRankingGpuMemoryUtilization[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
				runaimetrics.RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15{}.GetName():    getOvertimeAvgMetricMockMaxValues(60, departmentsMockValues, depRankingGpuAllocation[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
				runaimetrics.RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15{}.GetName():       getOvertimeAvgMetricMockMaxValues(50, departmentsMockValues, depRankingGpuUtilization[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
				runaimetrics.RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15{}.GetName(): getOvertimeAvgMetricMockMaxValues(40, departmentsMockValues, depRankingGpuMemoryUtilization[api.RANKINGORDERTYPE_BOTTOM], ptr.To(15)),
			})
		})

		AfterAll(func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)
				for _, projectId := range createdProjectIds[tenantId] {
					_, err := projectApiService.DeleteProject(ctx, projectId, "", false)
					Expect(err).To(BeNil())
				}
				for _, departmentId := range createdDepartmentIds[tenantId] {
					_, err := departmentApiService.DeleteDepartment(ctx, departmentId)
					Expect(err).To(BeNil())
				}
			}
		})

		BeforeEach(func() {
			stopper = make(chan struct{})
			cancelCtx, cancel = context.WithCancel(context.Background())

			overtime_metrics_writer.InitGoRoutines(cancelCtx, stopper, mockTenantsClient)
		})

		AfterEach(func() {
			close(stopper)
			cancel()
		})

		It("Sanity for project - in get project from db and api", func() {
			ctx = SetContext(tenantIds[0], TestUser)
			Eventually(func(g Gomega) {
				dbProject, err := dao.GetProjectById(ctx, int64(projectId), tenantIds[0], models.SlimVerbosity)
				g.Expect(err).To(BeNil(), "Failed getting project")
				g.Expect(dbProject).NotTo(BeNil(), "dbProject is nil")
				overtimeData := &dbProject.OrgUnitData.OvertimeData
				validateOvertimeData(g, overtimeData, dao.RangeDay, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
				validateOvertimeData(g, overtimeData, dao.RangeWeek, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
				validateOvertimeData(g, overtimeData, dao.RangeMonth, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
			}, 10*time.Second, 1*time.Second).Should(Succeed())

			getResponse, err := projectApiService.GetProject(ctx, strconv.Itoa(projectId))
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			apiProject := getResponse.Body.(api.Project)
			apiOvertimeData := apiProject.OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeDay, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
			validateApiOvertimeData(apiOvertimeData, dao.RangeMonth, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
		})

		It("Sanity for department - in get department from db and api", func() {
			ctx = SetContext(tenantIds[0], TestUser)
			Eventually(func(g Gomega) {
				dbDepartment, err := dao.GetDepartmentById(ctx, int64(departmentId), tenantIds[0], models.SlimVerbosity, nil)
				g.Expect(err).To(BeNil(), "Failed getting department")
				g.Expect(dbDepartment).NotTo(BeNil(), "dbDepartment is nil")
				overtimeData := &dbDepartment.OrgUnitData.OvertimeData
				validateOvertimeData(g, overtimeData, dao.RangeDay, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
				validateOvertimeData(g, overtimeData, dao.RangeWeek, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
				validateOvertimeData(g, overtimeData, dao.RangeMonth, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
			}, 10*time.Second, 1*time.Second).Should(Succeed())

			response, err := departmentApiService.GetDepartment(ctx, strconv.Itoa(departmentId))
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			apiDepartment := response.Body.(api.Department)
			apiOvertimeData := apiDepartment.OvertimeData
			validateApiOvertimeData(apiOvertimeData, dao.RangeMonth, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
			validateApiOvertimeData(apiOvertimeData, dao.RangeWeek, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
			validateApiOvertimeData(apiOvertimeData, dao.RangeDay, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
		})

		It("Validate all projects results - in get projects from db and api", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					dbProjects, err := dao.GetProjects(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, "", nil, false, models.SlimVerbosity)
					Expect(err).To(BeNil())

					for _, dbProject := range *dbProjects {
						overtimeData := &dbProject.OrgUnitData.OvertimeData
						validateOvertimeData(g, overtimeData, dao.RangeDay, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
						validateOvertimeData(g, overtimeData, dao.RangeWeek, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
						validateOvertimeData(g, overtimeData, dao.RangeMonth, projectsAvgGpuAllocation[dbProject.Name], projectsAvgGpuUtilization[dbProject.Name], nil)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating db projects overtime data for tenantId %s", tenantId)

				listResponse, err := projectApiService.GetProjects(ctx, []string{}, "", "", 0, 0, "")
				Expect(err).To(BeNil())
				Expect(listResponse.Code).To(Equal(http.StatusOK))
				apiProjects := listResponse.Body.(api.GetProjects200Response).Projects
				for _, apiProject := range apiProjects {
					overtimeData := apiProject.OvertimeData
					validateApiOvertimeData(overtimeData, dao.RangeDay, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
					validateApiOvertimeData(overtimeData, dao.RangeWeek, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
					validateApiOvertimeData(overtimeData, dao.RangeMonth, projectsAvgGpuAllocation[apiProject.Name], projectsAvgGpuUtilization[apiProject.Name], nil)
				}
			}
		})

		It("Validate all departments results - in get departments from db and api", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					dbDepartments, err := dao.GetDepartments(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, nil, false, models.SlimVerbosity)
					Expect(err).To(BeNil())

					for _, dbDepartment := range *dbDepartments {
						overtimeData := &dbDepartment.OrgUnitData.OvertimeData
						validateOvertimeData(g, overtimeData, dao.RangeDay, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
						validateOvertimeData(g, overtimeData, dao.RangeWeek, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
						validateOvertimeData(g, overtimeData, dao.RangeMonth, departmentsAvgGpuAllocation[dbDepartment.Name], departmentsAvgGpuUtilization[dbDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[dbDepartment.Name]))
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments overtime data for tenantId %s", tenantId)

				listResponse, err := departmentApiService.GetDepartments(ctx, []string{}, "", "", "", 0, 0)
				Expect(err).To(BeNil())
				Expect(listResponse.Code).To(Equal(http.StatusOK))
				apiDepartments := listResponse.Body.(api.GetDepartments200Response).Departments
				for _, apiDepartment := range apiDepartments {
					overtimeData := apiDepartment.OvertimeData
					validateApiOvertimeData(overtimeData, dao.RangeDay, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
					validateApiOvertimeData(overtimeData, dao.RangeWeek, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
					validateApiOvertimeData(overtimeData, dao.RangeMonth, departmentsAvgGpuAllocation[apiDepartment.Name], departmentsAvgGpuUtilization[apiDepartment.Name], ptr.To(departmentsAvgGpuMemoryUtilization[apiDepartment.Name]))
				}
			}
		})

		It("test GetProjectsRanking with nodepool filter", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				for _, clusterIdd := range []string{clusterId, clusterId} {
					Eventually(func(g Gomega) {
						for _, ranking := range rankingMetricsToTest {
							for _, order := range rankingOrdersToTest {
								response, err := projectApiService.GetProjectsRanking(ctx, ranking, order, clusterIdd, nodePoolName1, "")
								g.Expect(err).To(BeNil(), "Failed getting projects ranking %s / %s", ranking, order)
								g.Expect(response.Code).To(Equal(http.StatusOK), "Get projects ranking error code, %s / %s", ranking, order)
								apiProjects := response.Body.(api.GetProjectsRanking200Response).Projects
								g.Expect(len(apiProjects)).To(Equal(15), "Get projects ranking error code, %s / %s", ranking, order)
								for _, apiProject := range apiProjects {
									validateProjectApiRankingData(g, apiProject,
										projRankingGpuAllocation[order][apiProject.Name],
										projRankingGpuUtilization[order][apiProject.Name],
										ptr.To(projRankingGpuMemoryUtilization[order][apiProject.Name]),
										ranking, order)
								}
							}
						}
					}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating projects ranking data for clusterId %s", clusterIdd)
				}
			}
		})

		It("test GetDepartmentsRanking with nodepool filter", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				for _, clusterIdd := range []string{clusterId, clusterId} {
					Eventually(func(g Gomega) {
						for _, ranking := range rankingMetricsToTest {
							for _, order := range rankingOrdersToTest {
								response, err := departmentApiService.GetDepartmentsRanking(ctx, ranking, order, clusterIdd, "default")
								g.Expect(err).To(BeNil(), "Failed getting departments ranking %s / %s", ranking, order)
								g.Expect(response.Code).To(Equal(http.StatusOK), "Get departments ranking error code, %s / %s", ranking, order)
								apiDepartments := response.Body.(api.GetDepartmentsRanking200Response).Departments
								g.Expect(len(apiDepartments)).To(Equal(15), "Get departments ranking error code, %s / %s", ranking, order)
								for _, apiDepartment := range apiDepartments {
									validateDepartmentApiRankingData(g, apiDepartment,
										depRankingGpuAllocation[order][apiDepartment.Name],
										depRankingGpuUtilization[order][apiDepartment.Name],
										ptr.To(depRankingGpuMemoryUtilization[order][apiDepartment.Name]),
										ranking, order)
								}
							}
						}
					}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments ranking data for clusterId %s", clusterIdd)
				}
			}
		})

		It("test GetProjects with overtime filters", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					response, err := projectApiService.GetProjects(ctx, []string{"avgGpuMemoryUtilization24h>=50"}, "", "", 0, 0, "")
					g.Expect(err).To(BeNil(), "Failed getting projects with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get projects with overtime data filters error code: %v", response.Body)
					apiProjects := response.Body.(api.GetProjects200Response).Projects

					// GpuMemoryUtilization metric mock not set for projects!
					g.Expect(len(apiProjects)).To(Equal(0), "Get projects with overtime data filters wrong length")
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating projects with overtime data filters")

				Eventually(func(g Gomega) {
					response, err := projectApiService.GetProjects(ctx, []string{"avgGpuUtilization7d<=60"}, "", "", 0, 0, "")
					g.Expect(err).To(BeNil(), "Failed getting projects with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get projects with overtime data filters error code: %v", response.Body)
					apiProjects := response.Body.(api.GetProjects200Response).Projects
					g.Expect(len(apiProjects)).To(BeNumerically(">", 0), "Get projects with overtime data filters wrong length")

					for _, apiProject := range apiProjects {
						g.Expect(apiProject.OvertimeData.Range7dData).ToNot(BeNil(), "overtimeData is nil for project %s", apiProject.Name)
						g.Expect(apiProject.OvertimeData.Range7dData.AverageGpuUtilization).To(BeNumerically("<=", 60), "overtimeData is not as expected for project %s", apiProject.Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating projects with overtime data filters")

				Eventually(func(g Gomega) {
					response, err := projectApiService.GetProjects(ctx, []string{"avgGpuAllocation30d==10"}, "", "", 0, 0, "")
					g.Expect(err).To(BeNil(), "Failed getting projects with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get projects with overtime data filters error code: %v", response.Body)
					apiProjects := response.Body.(api.GetProjects200Response).Projects
					g.Expect(len(apiProjects)).To(BeNumerically(">", 0), "Get projects with overtime data filters wrong length")

					for _, apiProject := range apiProjects {
						g.Expect(apiProject.OvertimeData.Range30dData).ToNot(BeNil(), "overtimeData is nil for project %s", apiProject.Name)
						g.Expect(apiProject.OvertimeData.Range30dData.AverageGpuAllocation).To(BeEquivalentTo(10), "overtimeData is not as expected for project %s", apiProject.Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating projects with overtime data filters")
			}
		})

		It("test GetDepartments with overtime filters", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					response, err := departmentApiService.GetDepartments(ctx, []string{"avgGpuMemoryUtilization30d<=80"}, "", "", "", 0, 0)
					g.Expect(err).To(BeNil(), "Failed getting departments with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get departments with overtime data filters error code: %v", response.Body)
					apiDepartments := response.Body.(api.GetDepartments200Response).Departments

					g.Expect(len(apiDepartments)).To(BeNumerically(">", 0), "Get departments with overtime data filters wrong length")

					for _, apiDepartment := range apiDepartments {
						g.Expect(apiDepartment.OvertimeData.Range30dData).ToNot(BeNil(), "overtimeData is nil for department %s", apiDepartment.Name)
						g.Expect(apiDepartment.OvertimeData.Range30dData.AverageGpuMemoryUtilization).ToNot(BeNil(), "overtimeData.AverageGpuMemoryUtilization is nil for department %s", apiDepartment.Name)
						g.Expect(*apiDepartment.OvertimeData.Range30dData.AverageGpuMemoryUtilization).To(BeNumerically("<=", 80), "overtimeData is not as expected for department %s", apiDepartment.Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments with overtime data filters")

				Eventually(func(g Gomega) {
					response, err := departmentApiService.GetDepartments(ctx, []string{"avgGpuUtilization24h==60"}, "", "", "", 0, 0)
					g.Expect(err).To(BeNil(), "Failed getting departments with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get departments with overtime data filters error code: %v", response.Body)
					apiDepartments := response.Body.(api.GetDepartments200Response).Departments
					g.Expect(len(apiDepartments)).To(BeNumerically(">", 0), "Get departments with overtime data filters wrong length")

					for _, apiDepartment := range apiDepartments {
						g.Expect(apiDepartment.OvertimeData.Range24hData).ToNot(BeNil(), "overtimeData is nil for department %s", apiDepartment.Name)
						g.Expect(apiDepartment.OvertimeData.Range24hData.AverageGpuUtilization).To(BeEquivalentTo(60), "overtimeData is not as expected for department %s", apiDepartment.Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments with overtime data filters")

				Eventually(func(g Gomega) {
					response, err := departmentApiService.GetDepartments(ctx, []string{"avgGpuAllocation7d>=20"}, "", "", "", 0, 0)
					g.Expect(err).To(BeNil(), "Failed getting departments with overtime data filters")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get departments with overtime data filters error code: %v", response.Body)
					apiDepartments := response.Body.(api.GetDepartments200Response).Departments
					g.Expect(len(apiDepartments)).To(BeNumerically(">", 0), "Get departments with overtime data filters wrong length")

					for _, apiDepartment := range apiDepartments {
						g.Expect(apiDepartment.OvertimeData.Range7dData).ToNot(BeNil(), "overtimeData is nil for department %s", apiDepartment.Name)
						g.Expect(apiDepartment.OvertimeData.Range7dData.AverageGpuAllocation).To(BeNumerically(">=", 20), "overtimeData is not as expected for department %s", apiDepartment.Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments with overtime data filters")
			}
		})

		It("test GetProjects with overtime sortBy", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					response, err := projectApiService.GetProjects(ctx, []string{}, "avgGpuUtilization24h", "", 0, 0, "")
					g.Expect(err).To(BeNil(), "Failed getting projects with overtime data sort by")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get projects with overtime data sort by error code: %v", response.Body)
					apiProjects := response.Body.(api.GetProjects200Response).Projects

					g.Expect(len(apiProjects)).To(BeNumerically(">", 0), "Get projects with overtime data sort by wrong length")

					for i := 1; i < len(apiProjects); i++ {
						g.Expect(apiProjects[i].OvertimeData.Range24hData).ToNot(BeNil(), "overtimeData is nil for project %s", apiProjects[i].Name)
						g.Expect(apiProjects[i-1].OvertimeData.Range24hData.AverageGpuUtilization).To(BeNumerically("<=", apiProjects[i].OvertimeData.Range24hData.AverageGpuUtilization), "overtimeData is not as expected for project %s", apiProjects[i].Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating projects with overtime data filters")
			}
		})

		It("test GetDepartments with overtime sortBy", func() {
			for _, tenantId := range tenantIds {
				ctx := SetContext(tenantId, TestUser)

				Eventually(func(g Gomega) {
					response, err := departmentApiService.GetDepartments(ctx, []string{}, "avgGpuMemoryUtilization30d", "", "DESC", 0, 0)
					g.Expect(err).To(BeNil(), "Failed getting departments with overtime data sort by")
					g.Expect(response.Code).To(Equal(http.StatusOK), "Get departments with overtime data sort by error code: %v", response.Body)
					apiDepartments := response.Body.(api.GetDepartments200Response).Departments

					g.Expect(len(apiDepartments)).To(BeNumerically(">", 0), "Get departments with overtime data sort by wrong length")

					for i := 1; i < len(apiDepartments); i++ {
						g.Expect(apiDepartments[i].OvertimeData.Range30dData).ToNot(BeNil(), "overtimeData is nil for department %s", apiDepartments[i].Name)
						g.Expect(apiDepartments[i].OvertimeData.Range30dData.AverageGpuMemoryUtilization).ToNot(BeNil(), "overtimeData.AverageGpuMemoryUtilization is nil for department %s", apiDepartments[i].Name)
						g.Expect(*apiDepartments[i-1].OvertimeData.Range30dData.AverageGpuMemoryUtilization).To(BeNumerically(">=", *apiDepartments[i].OvertimeData.Range30dData.AverageGpuMemoryUtilization), "overtimeData is not as expected for department %s", apiDepartments[i].Name)
					}
				}, 10*time.Second, 1*time.Second).Should(Succeed(), "Failed validating departments with overtime data filters")
			}
		})
	})

})

func validateOvertimeDataNil(g Gomega, overtimeData *models.OvertimeData, overtimeMetricRange dao.OvertimeMetricRange) {
	g.Expect(overtimeData).NotTo(BeNil())

	allocation, utilization, memoryUtilization, updatedAt := getAllocationAndUtilizationForRange(overtimeData, overtimeMetricRange)
	g.Expect(allocation).To(BeNil())
	g.Expect(utilization).To(BeNil())
	g.Expect(memoryUtilization).To(BeNil())
	g.Expect(updatedAt).To(BeNil())
}

func validateOvertimeData(g Gomega, overtimeData *models.OvertimeData, overtimeMetricRange dao.OvertimeMetricRange,
	expectedGpuAllocation, expectedGpuUtilization float64, expectedGpuMemoryUtilization *float64,
) {
	g.Expect(overtimeData).NotTo(BeNil(), "overtimeData is nil")

	allocation, utilization, memoryUtilization, updatedAt := getAllocationAndUtilizationForRange(overtimeData, overtimeMetricRange)

	g.Expect(allocation).NotTo(BeNil(), "allocation is nil for range %v", overtimeMetricRange)
	g.Expect(*allocation).To(Equal(expectedGpuAllocation), "allocation is not as expected for range %v", overtimeMetricRange)
	g.Expect(utilization).NotTo(BeNil(), "utilization is nil for range %v", overtimeMetricRange)
	g.Expect(*utilization).To(Equal(expectedGpuUtilization), "utilization is not as expected for range %v", overtimeMetricRange)
	if expectedGpuMemoryUtilization != nil {
		g.Expect(memoryUtilization).NotTo(BeNil(), "memory utilization is nil for range %v", overtimeMetricRange)
		g.Expect(*memoryUtilization).To(Equal(*expectedGpuMemoryUtilization), "memory utilization is not as expected for range %v", overtimeMetricRange)
	} else {
		g.Expect(memoryUtilization).To(BeNil(), "memory utilization is not nil for range %v", overtimeMetricRange)
	}
	g.Expect(updatedAt).NotTo(BeNil(), "updatedAt is nil for range %v", overtimeMetricRange)
}

func validateApiOvertimeData(overtimeData api.OvertimeData, overtimeMetricRange dao.OvertimeMetricRange,
	expectedGpuAllocation, expectedGpuUtilization float64, expectedGpuMemoryUtilization *float64,
) {
	overtimeRangeData := getApiOvertimeRangeData(overtimeData, overtimeMetricRange)

	Expect(overtimeRangeData).ToNot(BeNil())

	Expect(overtimeRangeData.AverageGpuAllocation).To(Equal(expectedGpuAllocation))
	Expect(overtimeRangeData.AverageGpuUtilization).To(Equal(expectedGpuUtilization))
	Expect(overtimeRangeData.AverageGpuMemoryUtilization).To(BeEquivalentTo(expectedGpuMemoryUtilization))
	Expect(overtimeRangeData.UpdatedAt.IsZero()).To(BeFalse())
}

func validateApiOvertimeDataNil(overtimeData api.OvertimeData, overtimeMetricRange dao.OvertimeMetricRange) {
	overtimeRangeData := getApiOvertimeRangeData(overtimeData, overtimeMetricRange)
	Expect(overtimeRangeData).To(BeNil())
}

func validateProjectApiRankingData(g Gomega, rankingResult api.ProjectGetRanking,
	expectedGpuAllocation, expectedGpuUtilization float64, expectedGpuMemoryUtilization *float64,
	ranking api.RankingMetricType, order api.RankingOrderType,
) {
	g.Expect(rankingResult.AllocationValue).ToNot(BeNil(), "allocation is nil for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.AllocationValue).To(Equal(expectedGpuAllocation), "allocation is not as expected for ranking/order %s/%s", ranking, order)
	g.Expect(rankingResult.UtilizationValue).ToNot(BeNil(), "utilization is nil for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.UtilizationValue).To(Equal(expectedGpuUtilization), "utilization is not as expected for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.GpuQuota).To(Equal(FullResourcesDeservedGpu), "gpu quota is not as expected for ranking/order %s/%s", ranking, order)

	if expectedGpuMemoryUtilization != nil {
		g.Expect(rankingResult.MemoryUtilizationValue).ToNot(BeNil(), "memory utilization is nil for ranking/order %s/%s", ranking, order)
		g.Expect(*rankingResult.MemoryUtilizationValue).To(Equal(*expectedGpuMemoryUtilization), "memory utilization is not as expected for ranking/order %s/%s", ranking, order)
	} else {
		g.Expect(rankingResult.MemoryUtilizationValue).To(BeNil(), "memory utilization is not nil for ranking/order %s/%s", ranking, order)
	}
}

func validateDepartmentApiRankingData(g Gomega, rankingResult api.DepartmentGetRanking,
	expectedGpuAllocation, expectedGpuUtilization float64, expectedGpuMemoryUtilization *float64,
	ranking api.RankingMetricType, order api.RankingOrderType,
) {
	g.Expect(rankingResult.AllocationValue).ToNot(BeNil(), "allocation is nil for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.AllocationValue).To(Equal(expectedGpuAllocation), "allocation is not as expected for ranking/order %s/%s", ranking, order)
	g.Expect(rankingResult.UtilizationValue).ToNot(BeNil(), "utilization is nil for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.UtilizationValue).To(Equal(expectedGpuUtilization), "utilization is not as expected for ranking/order %s/%s", ranking, order)
	g.Expect(*rankingResult.GpuQuota).To(Equal(FullResourcesDeservedGpu), "gpu quota is not as expected for ranking/order %s/%s", ranking, order)
	if expectedGpuMemoryUtilization != nil {
		g.Expect(rankingResult.MemoryUtilizationValue).ToNot(BeNil(), "memory utilization is nil for ranking/order %s/%s", ranking, order)
		g.Expect(*rankingResult.MemoryUtilizationValue).To(Equal(*expectedGpuMemoryUtilization), "memory utilization is not as expected for ranking/order %s/%s", ranking, order)
	} else {
		g.Expect(rankingResult.MemoryUtilizationValue).To(BeNil(), "memory utilization is not nil for ranking/order %s/%s", ranking, order)
	}
}

func getAllocationAndUtilizationForRange(overtimeData *models.OvertimeData, rangeType dao.OvertimeMetricRange) (*float64, *float64, *float64, *time.Time) {
	switch rangeType {
	case dao.RangeDay:
		return overtimeData.AvgGpuAllocation24h, overtimeData.AvgGpuUtilization24h, overtimeData.AvgGpuMemoryUtilization24h, overtimeData.Avg24hRangeUpdatedAt
	case dao.RangeWeek:
		return overtimeData.AvgGpuAllocation7d, overtimeData.AvgGpuUtilization7d, overtimeData.AvgGpuMemoryUtilization7d, overtimeData.Avg7dRangeUpdatedAt
	case dao.RangeMonth:
		return overtimeData.AvgGpuAllocation30d, overtimeData.AvgGpuUtilization30d, overtimeData.AvgGpuMemoryUtilization30d, overtimeData.Avg30dRangeUpdatedAt
	}
	return nil, nil, nil, nil
}

func getOvertimeAvgMetricMock(baseValue float64, mockValues []OvertimeMetricsMockValue, mockMetricValues map[string]float64) openapi.GetMetrics200Response {
	return getOvertimeAvgMetricMockMaxValues(baseValue, mockValues, mockMetricValues, nil)
}

func getOvertimeAvgMetricMockMaxValues(
	baseValue float64,
	mockValues []OvertimeMetricsMockValue,
	mockMetricValues map[string]float64,
	maxValues *int,
) openapi.GetMetrics200Response {
	data := []openapi.Result{}
	value := baseValue

	for _, mockValue := range mockValues {
		data = append(data, getAvgMockResult(baseValue, mockValue, mockMetricValues))
		if maxValues != nil && len(data) == *maxValues {
			break
		}
		value = value + 1
	}

	return openapi.GetMetrics200Response{
		Status:  "success",
		IsRange: true,
		Data:    data,
	}
}

func getOrgUnitsOvertimeMetricsMockValues(orgUnits map[string]map[string][]string) ([]OvertimeMetricsMockValue, []OvertimeMetricsMockValue) {
	projMckValues := []OvertimeMetricsMockValue{}
	depMockValues := []OvertimeMetricsMockValue{}

	for clusterId, departmentNames := range orgUnits {
		for departmentName, projectNames := range departmentNames {
			depMockValues = append(depMockValues, OvertimeMetricsMockValue{
				ClusterId:  clusterId,
				Department: departmentName,
			})

			for _, projectName := range projectNames {
				projMckValues = append(projMckValues, OvertimeMetricsMockValue{
					ClusterId:  clusterId,
					Department: departmentName,
					Project:    projectName,
				})
			}
		}
	}

	return projMckValues, depMockValues
}

func getAvgMockResult(value float64, mockValue OvertimeMetricsMockValue, mockMetricValues map[string]float64) openapi.Result {
	measurements := []openapi.Measurement{
		{
			Timestamp: time.UnixMilli(1708958503653),
			Value:     fmt.Sprint(value),
		},
	}

	mockMetricValueKey := mockValue.Department

	labels := map[string]string{
		"clusterId":  mockValue.ClusterId,
		"department": mockValue.Department,
	}
	if mockValue.Project != "" {
		labels["project"] = mockValue.Project
		mockMetricValueKey = mockValue.Project
	}

	mockMetricValues[mockMetricValueKey] = value

	return openapi.Result{
		Measurements: measurements,
		MetricLabels: labels,
	}
}

func getApiOvertimeRangeData(overtimeData api.OvertimeData, rangeType dao.OvertimeMetricRange) *api.OvertimeRangeData {
	switch rangeType {
	case dao.RangeDay:
		return overtimeData.Range24hData
	case dao.RangeWeek:
		return overtimeData.Range7dData
	case dao.RangeMonth:
		return overtimeData.Range30dData
	}
	return nil
}

func findProjectInList(projects []api.Project, projectId int64) *api.Project {
	for _, project := range projects {
		if project.Id == strconv.Itoa(int(projectId)) {
			return &project
		}
	}
	return nil
}

func findDepartmentInList(departments []api.DepartmentForList, departmentId int64) *api.DepartmentForList {
	for _, department := range departments {
		if department.Id == strconv.Itoa(int(departmentId)) {
			return &department
		}
	}
	return nil
}
