/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	workloadsOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	workloadsMocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("Department API tests", Ordered, func() {
	clusterId := uuid.New().String()

	ctx := SetContext(TestTenant, TestUser)

	var departmentIds []string
	var departmentApiService api.DepartmentsAPIServicer
	var departmentV1ApiService api.DepartmentsV1APIServicer

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *openapi.NewNullableString(ptr.To("2.21")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
		Nodepools: []openapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)

	BeforeEach(func() {
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		departmentV1ApiService = v1.NewDepartmentV1ApiService(&clusterServiceClient)
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

	Context("test create/get departments", func() {
		var projectApiService api.ProjectsAPIServicer

		BeforeAll(func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test create/get one department without setting rank", func() {
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			request.Resources[0].Rank = nil
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			createdDepartment := createResponse.Body.(api.Department)
			Expect(len(createdDepartment.Resources)).To(Equal(1))
			resources := createdDepartment.Resources[0]
			Expect(resources.Rank).To(Equal(ptr.To(services.DefaultResourceRank)))

			getResponse, err := departmentApiService.GetDepartment(ctx, createdDepartment.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getDepartment := getResponse.Body.(api.Department)
			resources = getDepartment.Resources[0]
			Expect(resources.Rank).To(Equal(ptr.To(services.DefaultResourceRank)))

			_, deleteErr := departmentApiService.DeleteDepartment(ctx, getDepartment.Id)
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("with adjustable weights off and deserved=0, every department route returns MinOverQuotaWeight for GPU, CPU, and Memory", func() {
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

			// POST /departments
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			request.Resources = zeroResources
			createResp, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))
			created := createResp.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)
			assertAllResourcesFloored(created.Resources[0])

			// GET /departments/{id}
			getResp, err := departmentApiService.GetDepartment(ctx, created.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(getResp.Body.(api.Department).Resources[0])

			// GET /departments (list)
			listResp, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			listed := listResp.Body.(api.GetDepartments200Response).Departments
			var listedDept *api.DepartmentForList
			for i := range listed {
				if listed[i].Id == created.Id {
					listedDept = &listed[i]
				}
			}
			Expect(listedDept).ToNot(BeNil())
			assertAllResourcesFloored(listedDept.Resources[0])

			// PUT /departments/{id} — round-trip the GET response; validation must accept MinOverQuotaWeight when deserved=0
			updateReq := getDepartmentUpdateRequest(getResp.Body.(api.Department))
			updateResp, err := departmentApiService.UpdateDepartment(ctx, created.Id, updateReq)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(updateResp.Body.(api.Department).Resources[0])

			// PUT /departments/{id}/resources
			putResourcesResp, err := departmentApiService.UpdateDepartmentResources(ctx, created.Id, zeroResources)
			Expect(err).To(BeNil())
			Expect(putResourcesResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(putResourcesResp.Body.([]api.Resources)[0])

			// PATCH /departments/{id}/resources — round-trip the previous response
			patchResp, err := departmentApiService.PatchDepartmentResources(ctx, created.Id, getDepartmentUpdateRequest(updateResp.Body.(api.Department)).Resources)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			assertAllResourcesFloored(patchResp.Body.([]api.Resources)[0])

			// GET /sync/departments — cluster-scheduler feed
			syncResp, err := departmentApiService.GetSyncDepartments(ctx, clusterId, time.Time{}, 0, 0)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusOK))
			syncDepartments := syncResp.Body.(api.GetSyncDepartments200Response).Departments
			var syncedDept *api.DepartmentSync
			for i := range syncDepartments {
				if syncDepartments[i].Id == created.Id {
					syncedDept = &syncDepartments[i]
				}
			}
			Expect(syncedDept).ToNot(BeNil())
			Expect(*syncedDept.Resources[0].Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(syncedDept.Resources[0].Cpu).ToNot(BeNil())
			Expect(*syncedDept.Resources[0].Cpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(syncedDept.Resources[0].Memory).ToNot(BeNil())
			Expect(*syncedDept.Resources[0].Memory.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
		})

		It("test create department with unlimited gpu", func() {
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)

			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			request.Resources[0].Gpu = api.ResourcesNullableGpu{
				Deserved:        ptr.To(float64(-1)),
				Limit:           nil,
				OverQuotaWeight: nil,
			}
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
		})

		It("test create/get one department no override", func() {
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)
			Expect(created.CreatedAt).ToNot(BeNil())
			Expect(created.UpdatedAt).To(Equal(created.CreatedAt))
			Expect(created.Name).To(Equal(request.Name))
			Expect(created.Description).To(Equal(fmt.Sprintf("%s department description", request.Name)))
			Expect(created.ClusterId).To(Equal(request.ClusterId))
			// Expect(created.ParentId).To(BeNil())
			// Expect(created.SchedulingRules).To(BeNil())
			Expect(len(created.Resources)).To(Equal(1))
			resources := created.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(ptr.Deref(request.Resources[0].Gpu.Deserved, 0)))
			Expect(resources.Gpu.Limit).To(Equal(request.Resources[0].Gpu.Limit))
			Expect(resources.Gpu.OverQuotaWeight).To(Equal(request.Resources[0].Gpu.OverQuotaWeight))
			Expect(resources.Cpu).To(BeNil())
			Expect(resources.Memory).To(BeNil())
			Expect(resources.Rank).To(Equal(ptr.To(services.DefaultResourceRank)))
			Expect(created.CreatedBy).To(Equal(TestUser))

			getResponse, err := departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getDepartment := getResponse.Body.(api.Department)
			Expect(getDepartment.Id).To(Equal(created.Id))
			Expect(getDepartment.ClusterId).To(Equal(created.ClusterId))
			Expect(getDepartment.Description).To(Equal(fmt.Sprintf("%s department description", request.Name)))
			Expect(getDepartment.Resources).To(Equal(created.Resources))
			Expect(getDepartment.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(getDepartment.UpdatedAt.UnixMilli()).To(Equal(created.UpdatedAt.UnixMilli()))

			ancestorsResponse, err := departmentApiService.GetDepartmentAncestors(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department ancestors")
			}
			ancestors := ancestorsResponse.Body.(authorization_utils.Ancestors)
			Expect(ancestors).To(Not(BeNil()))
			Expect(ancestors.TenantId).To(Equal(TestTenant))
			Expect(ancestors.ClusterUuid).To(Not(BeNil()))
			Expect(*ancestors.ClusterUuid).To(Equal(created.ClusterId))
			Expect(ancestors.DepartmentId).To(BeNil())
			Expect(ancestors.ProjectId).To(BeNil())
		})

		It("test create/get departments with override", func() {
			request := getFullDepartmentCreateRequest(createDepartmentName(), clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			createdParent := response.Body.(api.Department)
			Expect(createdParent.CreatedAt).ToNot(BeNil())
			Expect(createdParent.UpdatedAt).To(Equal(createdParent.CreatedAt))
			Expect(createdParent.Name).To(Equal(request.Name))
			Expect(createdParent.ClusterId).To(Equal(request.ClusterId))

			Expect(len(createdParent.Resources)).To(Equal(1))
			resources := createdParent.Resources[0]
			Expect(resources.NodePool.Id).To(Equal(request.Resources[0].NodePool.Id))
			Expect(resources.NodePool.Name).To(Equal(request.Resources[0].NodePool.Name))
			Expect(resources.Gpu.Deserved).To(Equal(ptr.Deref(request.Resources[0].Gpu.Deserved, 0)))
			Expect(resources.Gpu.Limit).To(Equal(request.Resources[0].Gpu.Limit))
			Expect(resources.Gpu.OverQuotaWeight).To(Equal(request.Resources[0].Gpu.OverQuotaWeight))
			Expect(resources.Cpu).To(Equal(request.Resources[0].Cpu))
			Expect(resources.Memory).To(Equal(request.Resources[0].Memory))

			request = getFullDepartmentCreateRequest(createDepartmentName(), clusterId)
			response, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			createdChild := response.Body.(api.Department)

			getResponse, err := departmentApiService.GetDepartment(ctx, createdChild.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDepartment1 := getResponse.Body.(api.Department)
			Expect(getDepartment1.Id).To(Equal(createdChild.Id))

			Expect(getDepartment1.ClusterId).To(Equal(createdChild.ClusterId))

			Expect(getDepartment1.Resources).To(Equal(createdChild.Resources))
			Expect(getDepartment1.CreatedAt.UnixMilli()).To(Equal(createdChild.CreatedAt.UnixMilli()))
			Expect(getDepartment1.UpdatedAt.UnixMilli()).To(Equal(createdChild.UpdatedAt.UnixMilli()))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 10)
			if err != nil {
				Fail("Failed getting departments")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDepartments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(getDepartments)).To(Equal(3))

			Expect(getDepartments).To(ContainElement(WithTransform(func(dep api.DepartmentForList) string { return dep.Id }, Equal(createdParent.Id))))
			Expect(getDepartments).To(ContainElement(WithTransform(func(dep api.DepartmentForList) string { return dep.Id }, Equal(createdChild.Id))))

			for _, dep := range getDepartments {
				departmentIds = append(departmentIds, dep.Id)
			}

			getResponseV1, err := departmentV1ApiService.GetDepartmentsV1(ctx, clusterId, true, false)
			if err != nil {
				Fail("Failed getting departments")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDepartmentsV1 := getResponseV1.Body.([]api.DepartmentForGetV1)
			Expect(len(getDepartmentsV1)).To(Equal(3))

			Expect(getDepartments).To(ContainElement(WithTransform(func(dep api.DepartmentForList) string { return dep.Id }, Equal(createdParent.Id))))
			Expect(getDepartments).To(ContainElement(WithTransform(func(dep api.DepartmentForList) string { return dep.Id }, Equal(createdChild.Id))))
		})

		It("test create department with existing name", func() {
			name := "a-" + createDepartmentName()
			request := getFullDepartmentCreateRequest(name, clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			firstDep := response.Body.(api.Department)
			departmentIds = append(departmentIds, firstDep.Id)

			response, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusConflict))
		})

		It("test pagination", func() {
			name1 := "a-" + createDepartmentName()
			request := getFullDepartmentCreateRequest(name1, clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			firstDep := response.Body.(api.Department)

			name2 := "b-" + createDepartmentName()
			request = getFullDepartmentCreateRequest(name2, clusterId)
			response, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			secondDep := response.Body.(api.Department)

			departmentIds = append(departmentIds, firstDep.Id)
			departmentIds = append(departmentIds, secondDep.Id)

			// test pagination
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{}, "name", "", "", 0, 1)
			if err != nil {
				Fail("Failed getting departments")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDepartments := getResponse.Body.(api.GetDepartments200Response).Departments
			next := getResponse.Body.(api.GetDepartments200Response).Next
			Expect(len(getDepartments)).To(Equal(1))
			Expect(firstDep.Id).To(Equal(firstDep.Id))
			Expect(next).To(Equal(int32(1)))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{}, "name", "", "", 1, 1)
			if err != nil {
				Fail("Failed getting departments")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDepartments = getResponse.Body.(api.GetDepartments200Response).Departments
			next = getResponse.Body.(api.GetDepartments200Response).Next
			Expect(len(getDepartments)).To(Equal(1))
			Expect(secondDep.Id).To(Equal(secondDep.Id))
			Expect(next).To(Equal(int32(2)))
		})

		It("test get non existing department", func() {
			getResponse, err := departmentApiService.GetDepartment(ctx, "777")
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test create projects in department and validate children in get department", func() {
			var projectIds []string
			request := getFullDepartmentCreateRequest(createDepartmentName(), clusterId)
			response, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			dep := response.Body.(api.Department)
			departmentIds = append(departmentIds, dep.Id)

			projects := map[string]api.Project{}
			for i := 0; i < 3; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, dep.Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				if err != nil {
					Fail("Failed creating project")
				}

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				projects[created.Id] = created
				projectIds = append(projectIds, created.Id)
			}

			getResponse, err := departmentApiService.GetDepartments(ctx, []string{}, "", api.VERBOSITY_VERBOSE, "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			var ourDep *api.DepartmentForList = nil
			for _, department := range departments {
				if department.Id == dep.Id {
					ourDep = &department
					break
				}
			}
			Expect(ourDep).ToNot(BeNil())
			Expect(ourDep.Id).To(Equal(dep.Id))
			Expect(ourDep.Children).ToNot(BeNil())
			Expect(*ourDep.Children).To(HaveLen(3))
			for _, projChild := range *ourDep.Children {
				proj, ok := projects[projChild.Id]
				Expect(ok).To(BeTrue())
				Expect(proj.Name).To(Equal(projChild.Name))
			}

			getResponse, err = departmentApiService.GetDepartment(ctx, dep.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			gottenDep := getResponse.Body.(api.Department)
			Expect(gottenDep.Id).To(Equal(dep.Id))
			Expect(gottenDep.Children).ToNot(BeNil())
			Expect(*gottenDep.Children).To(HaveLen(3))
			for _, projChild := range *gottenDep.Children {
				proj, ok := projects[projChild.Id]
				Expect(ok).To(BeTrue())
				Expect(proj.Name).To(Equal(projChild.Name))
			}

			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{}, "", api.VERBOSITY_VERBOSE, "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			for _, department := range departments {
				if department.Id == dep.Id {
					ourDep = &department
					break
				}
			}
			Expect(ourDep).ToNot(BeNil())
			Expect(ourDep.Id).To(Equal(dep.Id))
			Expect(ourDep.Children).ToNot(BeNil())
			Expect(*ourDep.Children).To(HaveLen(0))

			getResponse, err = departmentApiService.GetDepartment(ctx, dep.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			gottenDep = getResponse.Body.(api.Department)
			Expect(gottenDep.Id).To(Equal(dep.Id))
			Expect(gottenDep.Children).ToNot(BeNil())
			Expect(*gottenDep.Children).To(HaveLen(0))
		})

		It("test create project with version below 2.20 with not default rank", func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *openapi.NewNullableString(ptr.To("2.19")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
					},
				},
			}, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			request.Resources = []api.ResourcesNullable{{
				Rank: ptr.To("VeryHigh"),
			}}
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("test with another nodepool", func() {
		var (
			nodePoolName1         = "nodepool-a"
			nodePoolId1           = float32(2)
			departmentApiService2 api.DepartmentsAPIServicer
		)

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *openapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					}, {
						Name:                   nodePoolName1,
						Id:                     fmt.Sprint(nodePoolId1),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
		})
		departmentApiService2 = departments.NewDepartmentApiService(&clusterServiceClient, nil)

		AfterAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *openapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			departmentApiService2 = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		})

		It("test create department without nodepoolname - returned with name in GET", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			request.Resources = append(request.Resources, CreateNodePoolApiResources(false, "2", nodePoolName1))
			request.Resources[0].NodePool.Name = ""
			request.Resources[1].NodePool.Name = ""
			createResponse, err := departmentApiService2.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)

			getResponse, err := departmentApiService2.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDepartment := getResponse.Body.(api.Department)
			for _, resources := range getDepartment.Resources {
				Expect(resources.NodePool).To(Not(BeNil()))
				if resources.NodePool.Id == "2" {
					Expect(resources.NodePool.Name).To(Equal(nodePoolName1))
				} else {
					Expect(resources.NodePool.Name).To(Equal(defaultNodePoolName))
				}
			}
		})
	})

	Context("test update departments", func() {
		var projectApiService api.ProjectsAPIServicer

		BeforeEach(func() {
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test update department ok status", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)
			// Expect(created.SchedulingRules).To(BeNil())

			updateRequest := getDepartmentUpdateRequest(created)
			updateRequest.Description = "new description"
			//timeLimits := int32(1000)
			//updateRequest.SchedulingRules = &api.DataUpdatableFieldsSchedulingRules{
			//	InteractiveJobTimeLimitSeconds: &timeLimits,
			//}

			updateResponse, err := departmentApiService.UpdateDepartment(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating department")
			}
			updated := updateResponse.Body.(api.Department)
			Expect(updated.Id).To(Equal(created.Id))
			// Expect(updated.ParentId).To(Equal(created.ParentId))
			Expect(updated.ClusterId).To(Equal(created.ClusterId))
			Expect(updated.Description).To(Equal("new description"))
			//Expect(updated.SchedulingRules).To(Equal(&api.DataUpdatableFieldsSchedulingRules{
			//	InteractiveJobTimeLimitSeconds: &timeLimits,
			//}))
			Expect(updated.Resources).To(Equal(created.Resources))
			Expect(updated.CreatedAt.UnixMilli()).To(Equal(created.CreatedAt.UnixMilli()))
			Expect(updated.UpdatedAt.UnixMilli()).NotTo(Equal(created.UpdatedAt.UnixMilli()))

			getResponse, err := departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			getDepartment := getResponse.Body.(api.Department)
			Expect(getDepartment.Id).To(Equal(updated.Id))
			// Expect(getDepartment.ParentId).To(Equal(updated.ParentId))
			Expect(getDepartment.ClusterId).To(Equal(updated.ClusterId))
			// Expect(getDepartment.SchedulingRules).To(Equal(updated.SchedulingRules))
			Expect(getDepartment.Description).To(Equal("new description"))
			Expect(getDepartment.Resources).To(Equal(updated.Resources))
			Expect(getDepartment.CreatedAt.UnixMilli()).To(Equal(updated.CreatedAt.UnixMilli()))
			Expect(getDepartment.UpdatedAt.UnixMilli()).To(Equal(updated.UpdatedAt.UnixMilli()))
			Expect(getDepartment.UpdatedBy).To(Equal(TestUser))
		})

		It("test update department returns ProjectsAggregatedResources", func() {
			// Create department
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			dep := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, dep.Id)

			// Create projects under the department with specific GPU resources
			var projectIds []string
			proj1GpuDeserved := float64(5)
			proj2GpuDeserved := float64(10)

			proj1Request := api.ProjectCreationRequest{
				Name:      createProjectName(),
				ClusterId: clusterId,
				ParentId:  &dep.Id,
				Resources: []api.ResourcesNullable{{
					NodePool: &api.ResourcesNodePool{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName},
					Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(proj1GpuDeserved)},
				}},
			}
			proj1Response, err := projectApiService.CreateProject(ctx, proj1Request)
			Expect(err).To(BeNil())
			Expect(proj1Response.Code).To(Equal(http.StatusCreated))
			proj1 := proj1Response.Body.(api.Project)
			projectIds = append(projectIds, proj1.Id)

			proj2Request := api.ProjectCreationRequest{
				Name:      createProjectName(),
				ClusterId: clusterId,
				ParentId:  &dep.Id,
				Resources: []api.ResourcesNullable{{
					NodePool: &api.ResourcesNodePool{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName},
					Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(proj2GpuDeserved)},
				}},
			}
			proj2Response, err := projectApiService.CreateProject(ctx, proj2Request)
			Expect(err).To(BeNil())
			Expect(proj2Response.Code).To(Equal(http.StatusCreated))
			proj2 := proj2Response.Body.(api.Project)
			projectIds = append(projectIds, proj2.Id)

			// Update department
			updateRequest := getDepartmentUpdateRequest(dep)
			updateRequest.Description = "updated description"
			updateResponse, err := departmentApiService.UpdateDepartment(ctx, dep.Id, updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updated := updateResponse.Body.(api.Department)

			// Validate ProjectsAggregatedResources is populated
			Expect(updated.ProjectsAggregatedResources).ToNot(BeNil())
			Expect(*updated.ProjectsAggregatedResources).To(HaveLen(1))
			aggregated := (*updated.ProjectsAggregatedResources)[0]
			Expect(aggregated.NodePool.Id).To(Equal(fmt.Sprint(defaultNodePoolId)))
			Expect(aggregated.NodePool.Name).To(Equal(defaultNodePoolName))
			expectedTotalGpu := proj1GpuDeserved + proj2GpuDeserved
			Expect(*aggregated.Gpu.Deserved).To(Equal(expectedTotalGpu))

			// Validate Children is populated with the two projects
			Expect(updated.Children).ToNot(BeNil())
			Expect(*updated.Children).To(HaveLen(2))
			childIds := []string{}
			childNames := []string{}
			for _, child := range *updated.Children {
				childIds = append(childIds, child.Id)
				childNames = append(childNames, child.Name)
			}
			Expect(childIds).To(ContainElement(proj1.Id))
			Expect(childIds).To(ContainElement(proj2.Id))
			Expect(childNames).To(ContainElement(proj1.Name))
			Expect(childNames).To(ContainElement(proj2.Name))

			// Validate Status is populated (QuotaStatus should have zero values since no status updates were made)
			Expect(updated.Status.QuotaStatus.Allocated.Gpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.Allocated.Cpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.Allocated.Memory).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.AllocatedNonPreemptible.Gpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.AllocatedNonPreemptible.Cpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.AllocatedNonPreemptible.Memory).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.Requested.Gpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.Requested.Cpu).To(Equal(float64(0)))
			Expect(updated.Status.QuotaStatus.Requested.Memory).To(Equal(float64(0)))

			// Verify UpdateDepartment returns the same verbose fields as GetDepartment
			getResponse, err := departmentApiService.GetDepartment(ctx, dep.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDep := getResponse.Body.(api.Department)
			Expect(updated.ProjectsAggregatedResources).To(Equal(getDep.ProjectsAggregatedResources))
			Expect(updated.Children).To(Equal(getDep.Children))
			Expect(updated.Status).To(Equal(getDep.Status))

			// Cleanup projects
			for _, projectId := range projectIds {
				deleteResponse, delErr := projectApiService.DeleteProject(ctx, projectId, "", false)
				Expect(delErr).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			}
		})

		It("test GetDepartment ProjectsAggregatedResources returns -1 when any project has unlimited", func() {
			// Create department
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			dep := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, dep.Id)

			// Create two projects under the department:
			// Project 1: unlimited values (gpuDeserved=-1, cpuDeserved=-1, memoryDeserved=-1)
			// Project 2: specific values (gpuDeserved=5, cpuDeserved=100, memoryDeserved=1000)
			var projectIds []string

			// Project 1 with unlimited resources (-1 means unlimited)
			proj1Request := api.ProjectCreationRequest{
				Name:      createProjectName(),
				ClusterId: clusterId,
				ParentId:  &dep.Id,
				Resources: []api.ResourcesNullable{{
					NodePool: &api.ResourcesNodePool{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName},
					Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(float64(-1))},
					Cpu: &api.ResourcesCpu{
						Deserved: ptr.To(float64(-1)),
					},
					Memory: &api.ResourcesMemory{
						Deserved: ptr.To(float64(-1)),
						Units:    api.UNITS_MB,
					},
				}},
			}
			proj1Response, err := projectApiService.CreateProject(ctx, proj1Request)
			Expect(err).To(BeNil())
			Expect(proj1Response.Code).To(Equal(http.StatusCreated))
			proj1 := proj1Response.Body.(api.Project)
			projectIds = append(projectIds, proj1.Id)

			// Project 2 with specific values
			proj2GpuDeserved := float64(5)
			proj2CpuDeserved := float64(100)
			proj2MemoryDeserved := float64(1000)
			proj2Request := api.ProjectCreationRequest{
				Name:      createProjectName(),
				ClusterId: clusterId,
				ParentId:  &dep.Id,
				Resources: []api.ResourcesNullable{{
					NodePool: &api.ResourcesNodePool{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName},
					Gpu:      api.ResourcesNullableGpu{Deserved: ptr.To(proj2GpuDeserved)},
					Cpu: &api.ResourcesCpu{
						Deserved: ptr.To(proj2CpuDeserved),
					},
					Memory: &api.ResourcesMemory{
						Deserved: ptr.To(proj2MemoryDeserved),
						Units:    api.UNITS_MB,
					},
				}},
			}
			proj2Response, err := projectApiService.CreateProject(ctx, proj2Request)
			Expect(err).To(BeNil())
			Expect(proj2Response.Code).To(Equal(http.StatusCreated))
			proj2 := proj2Response.Body.(api.Project)
			projectIds = append(projectIds, proj2.Id)

			// Get department and verify ProjectsAggregatedResources
			getResponse, err := departmentApiService.GetDepartment(ctx, dep.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDep := getResponse.Body.(api.Department)

			// Validate ProjectsAggregatedResources returns -1 for all resources since proj1 has unlimited
			// Expected: all aggregated values should be -1 because at least one project has unlimited
			Expect(getDep.ProjectsAggregatedResources).ToNot(BeNil())
			Expect(*getDep.ProjectsAggregatedResources).To(HaveLen(1))
			aggregated := (*getDep.ProjectsAggregatedResources)[0]
			Expect(aggregated.NodePool.Id).To(Equal(fmt.Sprint(defaultNodePoolId)))
			Expect(aggregated.NodePool.Name).To(Equal(defaultNodePoolName))

			// GPU: should be -1 (unlimited) because proj1 has unlimited
			Expect(aggregated.Gpu.Deserved).ToNot(BeNil())
			Expect(*aggregated.Gpu.Deserved).To(Equal(float64(-1)))

			// CPU: should be -1 (unlimited) because proj1 has unlimited
			Expect(aggregated.Cpu.Deserved).ToNot(BeNil())
			Expect(*aggregated.Cpu.Deserved).To(Equal(float64(-1)))

			// Memory: should be -1 (unlimited) because proj1 has unlimited
			Expect(aggregated.Memory.Deserved).ToNot(BeNil())
			Expect(*aggregated.Memory.Deserved).To(Equal(float64(-1)))
			Expect(aggregated.Memory.Units).To(Equal(api.UNITS_MB))

			// Cleanup projects
			for _, projectId := range projectIds {
				deleteResponse, delErr := projectApiService.DeleteProject(ctx, projectId, "", false)
				Expect(delErr).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			}
		})

		It("test update department with unlimited gpu", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)

			updateRequest := getDepartmentUpdateRequest(created)
			updateRequest.Description = "new description"
			updateRequest.Resources[0].Gpu = api.ResourcesNullableGpu{
				Deserved:        ptr.To(float64(-1)),
				Limit:           nil,
				OverQuotaWeight: nil,
			}

			updateResponse, err := departmentApiService.UpdateDepartment(ctx, created.Id, updateRequest)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
		})

		It("test update department not found", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)

			updateRequest := getDepartmentUpdateRequest(created)

			updateResponse, err := departmentApiService.UpdateDepartment(ctx, "444", updateRequest)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("test update department resources", func() {
		BeforeAll(func() {
			nodePoolName := "pool2"
			nodePoolId := float32(2)
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&openapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *openapi.NewNullableString(ptr.To("2.21")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
					{
						Name:                   nodePoolName,
						Id:                     fmt.Sprint(nodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
		})

		BeforeEach(func() {
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		})

		It("test update department resources ok status", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
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
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.Department)
			departmentIds = append(departmentIds, created.Id)
			validateResourcesArrayToNullable(request.Resources, created.Resources)

			updateResources := getUpdateResources()
			updateResources[0].Rank = ptr.To("Low")
			updateResponse, err := departmentApiService.UpdateDepartmentResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusOK))
			updatedResources := updateResponse.Body.([]api.Resources)
			validateResourcesArrayToNullable(updateResources, updatedResources)

			getResponse, err := departmentApiService.GetDepartment(ctx, created.Id)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getDepartment := getResponse.Body.(api.Department)
			validateResourcesArrayToNullable(updateResources, getDepartment.Resources)

			updateResources = getUpdateResources()
			updateResources[0].Rank = ptr.To("LowUnknown")
			updateResponse, err = departmentApiService.UpdateDepartmentResources(ctx, created.Id, updateResources)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test update department not found", func() {
			updateResources := getUpdateResources()
			updateResponse, err := departmentApiService.UpdateDepartmentResources(ctx, "444", updateResources)
			if err != nil {
				Fail("Failed updating department")
			}
			Expect(updateResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("test delete departments", func() {
		var (
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			projectIds        = []string{}
		)

		AfterEach(func() {
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
		})

		It("test delete department success", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentId := createResponse.Body.(api.Department).Id
			departmentIds = append(departmentIds, departmentId)

			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
			if err != nil {
				Fail("Failed deleting first department")
			}
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			departmentIds = []string{}

			getResponse, err := departmentApiService.GetDepartment(ctx, departmentId)
			if err != nil {
				Fail("Failed getting department")
			}
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("test delete department with projects in it - should fail", func() {
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentId := createResponse.Body.(api.Department).Id
			departmentIds = append(departmentIds, departmentId)

			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			projRequest := getProjectCreateRequest(createProjectName(), "", departmentId, clusterId)
			createResponse, err = projectApiService.CreateProject(ctx, projRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projectIds = append(projectIds, createResponse.Body.(api.Project).Id)

			deleteResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
			if err != nil {
				Fail("Failed deleting department")
			}
			Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("test sort/filter", func() {
		var department1Id string
		var department2Id string
		var department2Name string

		BeforeEach(func() {
			// department 1: total gpu quota: 2, total cpu quota: 2, total memory quota: 2
			// departtment 2: total gpu quota: 0, total cpu quota: 10, total memory quota: 1
			nodePoolName1 := "nodepool-a"
			nodePoolId1 := float32(2)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			department1Name := createDepartmentName()
			request := getDepartmentCreateRequest(department1Name, clusterId)
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
				},
			})
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			department1Id = createResponse.Body.(api.Department).Id

			department2Name = createDepartmentName()
			request = getDepartmentCreateRequest(department2Name, clusterId)
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
				},
			})
			createResponse, err = departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			department2Id = createResponse.Body.(api.Department).Id

			departmentIds = append(departmentIds, department1Id, department2Id)
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

		It("test sort/filter", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "totalGpuQuota", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(2))
			Expect(departments[0].Id).To(Equal(department2Id))
			Expect(departments[1].Id).To(Equal(department1Id))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "createdAt", "", "DESC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(2))
			Expect(departments[0].Id).To(Equal(department2Id))
			Expect(departments[1].Id).To(Equal(department1Id))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("name==%s", department2Name)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Id).To(Equal(department2Id))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("name!=%s", department2Name)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Id).To(Equal(department1Id))
		})
	})

	Context("test sort/filter by avg overtime data", func() {
		var (
			testDepartments    []api.Department
			d0, d1, d2, d3, d4 string
		)

		BeforeAll(func() {
			for i := 0; i < 5; i++ {
				request := getDepartmentCreateRequest(fmt.Sprintf("d%d-%s", i, rand.String(3)), clusterId)
				createResponse, err := departmentApiService.CreateDepartment(ctx, request)
				if err != nil {
					Fail("Failed creating department")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testDepartments = append(testDepartments, createResponse.Body.(api.Department))
			}

			result := map[string]map[string]*models.OvertimeData{clusterId: {}}

			var i float32 = 0
			for i = 0; i < 5; i++ {
				department := testDepartments[int(i)]

				result[clusterId][department.Name] = &models.OvertimeData{
					AvgGpuAllocation24h:  ptr.To(float64(i * 10)),            // 0, 10, 20, 30, 40
					AvgGpuUtilization24h: ptr.To(float64(i * 20)),            // 0, 20, 40, 60, 80
					AvgGpuAllocation7d:   ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160
					AvgGpuUtilization7d:  ptr.To(float64(mod(2-i, 10) * 20)), // 20, 10, 0, 180, 160
					AvgGpuAllocation30d:  ptr.To(float64((5 - i) * 20)),      // 100, 80, 60, 40, 20
					AvgGpuUtilization30d: ptr.To(float64((5 - i) * 10)),      // 50, 40, 30, 20, 10
				}
			}

			err := dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())

			d0 = testDepartments[0].Name
			d1 = testDepartments[1].Name
			d2 = testDepartments[2].Name
			d3 = testDepartments[3].Name
			d4 = testDepartments[4].Name
		})

		AfterAll(func() {
			for _, department := range testDepartments {
				deleteResponse, err := departmentApiService.DeleteDepartment(ctx, department.Id)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			testDepartments = []api.Department{}
		})

		It("test sort by avg gpu 24h", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation24h", "", "ASC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d0, d1, d2, d3, d4))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization24h", "", "ASC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames = getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d0, d1, d2, d3, d4))
		})

		It("test sort by avg gpu 7d", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation7d", "", "DESC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d3, d4, d0, d1, d2))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization7d", "", "ASC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames = getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d2, d1, d0, d4, d3))
		})

		It("test sort by avg gpu 30d", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuAllocation30d", "", "ASC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d4, d3, d2, d1, d0))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "avgGpuUtilization30d", "", "ASC", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(5))
			apiDepNames = getDepartmentNames(departments)
			Expect(apiDepNames).To(HaveExactElements(d4, d3, d2, d1, d0))
		})

		It("test filter by invalid cluster id", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", "invalid-uuid"), "avgGpuAllocation24h>=10"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test filter by unknown field", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("something==%s", "invalid-uuid"), "avgGpuAllocation24h>=10"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("test filter by avg gpu 24h", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation24h>=10"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(4))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(ConsistOf(d1, d2, d3, d4))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization24h==60"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal(d3))
		})

		It("test filter by avg gpu 7d", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation7d<=40"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(3))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(ConsistOf(d0, d1, d2))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization7d!=180"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(4))
			apiDepNames = getDepartmentNames(departments)
			Expect(apiDepNames).To(ConsistOf(d0, d1, d2, d4))
		})

		It("test filter by avg gpu 30d", func() {
			getResponse, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuAllocation30d<=60"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments := getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(3))
			apiDepNames := getDepartmentNames(departments)
			Expect(apiDepNames).To(ConsistOf(d2, d3, d4))

			getResponse, err = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), "avgGpuUtilization30d==20"}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			departments = getResponse.Body.(api.GetDepartments200Response).Departments
			Expect(len(departments)).To(Equal(1))
			Expect(departments[0].Name).To(Equal(d3))
		})
	})

	Context("count departments", func() {
		AfterEach(func() {
			for _, departmentId := range departmentIds {
				deleteResponse, err := departmentApiService.DeleteDepartment(ctx, departmentId)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			departmentIds = []string{}
		})
		It("count departments filter by department name", func() {
			departmentName := createDepartmentName()
			request := getDepartmentCreateRequest(departmentName, clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			if err != nil {
				Fail("Failed creating department")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)

			getCountResponse, err := departmentApiService.CountDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId), fmt.Sprintf("name==%s", departmentName)})
			Expect(err).To(BeNil())
			Expect(getCountResponse.Code).To(Equal(http.StatusOK))
			count := getCountResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(1)))
		})

		It("create and delete departments then count and validate", func() {
			getCountResponse, err := departmentApiService.CountDepartments(ctx, []string{})
			Expect(err).To(BeNil())
			Expect(getCountResponse.Code).To(Equal(http.StatusOK))
			countBefore := getCountResponse.Body.(api.CountResponse).Count

			testDepartmentIds := []string{}
			for i := 0; i < 10; i++ {
				request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
				createResponse, err := departmentApiService.CreateDepartment(ctx, request)
				if err != nil {
					Fail("Failed creating department")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testDepartmentIds = append(testDepartmentIds, createResponse.Body.(api.Department).Id)
				if i >= 5 {
					departmentIds = append(departmentIds, createResponse.Body.(api.Department).Id)
				}
			}
			for i := 0; i < 5; i++ {
				deleteResponse, err := departmentApiService.DeleteDepartment(ctx, testDepartmentIds[i])
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}

			getCountResponse, err = departmentApiService.CountDepartments(ctx, []string{})
			Expect(err).To(BeNil())
			Expect(getCountResponse.Code).To(Equal(http.StatusOK))
			count := getCountResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(countBefore + 5))
		})
	})

	Context("test get department descendants", func() {
		var (
			projectApiService   api.ProjectsAPIServicer
			department1Id       string
			department2Id       string
			department1Projects []string
			department2Projects []string
		)

		BeforeAll(func() {
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		It("test get department descendants with projects", func() {
			// Create first department
			request1 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse1, err := departmentApiService.CreateDepartment(ctx, request1)
			if err != nil {
				Fail("Failed creating department 1")
			}
			Expect(createResponse1.Code).To(Equal(http.StatusCreated))
			department1Id = createResponse1.Body.(api.Department).Id
			departmentIds = append(departmentIds, department1Id)

			// Create second department
			request2 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse2, err := departmentApiService.CreateDepartment(ctx, request2)
			if err != nil {
				Fail("Failed creating department 2")
			}
			Expect(createResponse2.Code).To(Equal(http.StatusCreated))
			department2Id = createResponse2.Body.(api.Department).Id
			departmentIds = append(departmentIds, department2Id)

			// Create projects in department 1
			for i := 0; i < 3; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department1Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				if err != nil {
					Fail("Failed creating project in department 1")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department1Projects = append(department1Projects, created.Id)
			}

			// Create projects in department 2
			for i := 0; i < 2; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department2Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				if err != nil {
					Fail("Failed creating project in department 2")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department2Projects = append(department2Projects, created.Id)
			}

			// Cast to concrete type to access GetDepartmentsDescendants method
			concreteService := departmentApiService.(*departments.DepartmentApiService)

			// Test GetDepartmentsDescendants for department 1
			descendantsResponse, err := concreteService.GetDepartmentsDescendants(ctx, department1Id)
			if err != nil {
				Fail("Failed getting department descendants")
			}
			Expect(descendantsResponse.Code).To(Equal(http.StatusOK))

			// Validate the response structure
			descendants := descendantsResponse.Body.(authorization_utils.Descendants)
			Expect(descendants).ToNot(BeNil())
			Expect(descendants.ProjectIds).ToNot(BeNil())

			// Convert returned project IDs to slice for easier comparison
			returnedProjectIds := descendants.ProjectIds
			Expect(len(returnedProjectIds)).To(Equal(len(department1Projects)))

			// Validate that all expected project IDs are present
			for _, expectedProjectId := range department1Projects {
				Expect(returnedProjectIds).To(ContainElement(expectedProjectId))
			}

			// Validate that no projects from department 2 are present
			for _, department2ProjectId := range department2Projects {
				Expect(returnedProjectIds).ToNot(ContainElement(department2ProjectId))
			}

			// Test GetDepartmentsDescendants for department 2
			descendantsResponse2, err := concreteService.GetDepartmentsDescendants(ctx, department2Id)
			if err != nil {
				Fail("Failed getting department 2 descendants")
			}
			Expect(descendantsResponse2.Code).To(Equal(http.StatusOK))

			descendants2 := descendantsResponse2.Body.(authorization_utils.Descendants)
			Expect(descendants2).ToNot(BeNil())
			Expect(descendants2.ProjectIds).ToNot(BeNil())

			returnedProjectIds2 := descendants2.ProjectIds
			Expect(len(returnedProjectIds2)).To(Equal(len(department2Projects)))

			// Validate that all expected project IDs from department 2 are present
			for _, expectedProjectId := range department2Projects {
				Expect(returnedProjectIds2).To(ContainElement(expectedProjectId))
			}

			// Validate that no projects from department 1 are present
			for _, department1ProjectId := range department1Projects {
				Expect(returnedProjectIds2).ToNot(ContainElement(department1ProjectId))
			}

			// Clean up projects
			allProjects := append(department1Projects, department2Projects...)
			for _, projectId := range allProjects {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting project")
				}
			}
		})

		It("test get departments descendants with multiple departments", func() {
			// Create two departments with projects
			var department1Id, department2Id string
			var department1Projects, department2Projects []string

			// Create department 1
			request1 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse1, err := departmentApiService.CreateDepartment(ctx, request1)
			Expect(err).To(BeNil())
			Expect(createResponse1.Code).To(Equal(http.StatusCreated))
			department1Id = createResponse1.Body.(api.Department).Id
			departmentIds = append(departmentIds, department1Id)

			// Create department 2
			request2 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse2, err := departmentApiService.CreateDepartment(ctx, request2)
			Expect(err).To(BeNil())
			Expect(createResponse2.Code).To(Equal(http.StatusCreated))
			department2Id = createResponse2.Body.(api.Department).Id
			departmentIds = append(departmentIds, department2Id)

			// Create projects for department 1
			for i := 0; i < 2; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department1Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department1Projects = append(department1Projects, created.Id)
			}

			// Create projects for department 2
			for i := 0; i < 3; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department2Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department2Projects = append(department2Projects, created.Id)
			}

			// Cast to concrete type to access GetDepartmentsDescendants method
			concreteService := departmentApiService.(*departments.DepartmentApiService)

			// Test GetDepartmentsDescendants with both departments
			combinedDepartmentIds := department1Id + "," + department2Id
			descendantsResponse, err := concreteService.GetDepartmentsDescendants(ctx, combinedDepartmentIds)
			Expect(err).To(BeNil())
			Expect(descendantsResponse.Code).To(Equal(http.StatusOK))

			// Validate the response structure
			descendants := descendantsResponse.Body.(authorization_utils.Descendants)
			Expect(descendants).ToNot(BeNil())
			Expect(descendants.ProjectIds).ToNot(BeNil())

			// Convert returned project IDs to slice for easier comparison
			returnedProjectIds := descendants.ProjectIds
			expectedTotalProjects := len(department1Projects) + len(department2Projects)
			Expect(len(returnedProjectIds)).To(Equal(expectedTotalProjects))

			// Validate that all projects from department 1 are present
			for _, expectedProjectId := range department1Projects {
				Expect(returnedProjectIds).To(ContainElement(expectedProjectId))
			}

			// Validate that all projects from department 2 are present
			for _, expectedProjectId := range department2Projects {
				Expect(returnedProjectIds).To(ContainElement(expectedProjectId))
			}

			// Clean up projects
			allProjects := append(department1Projects, department2Projects...)
			for _, projectId := range allProjects {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting project")
				}
			}
		})

		It("test get department descendants with department having no projects", func() {
			// Create a department without any projects
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentId := createResponse.Body.(api.Department).Id
			departmentIds = append(departmentIds, departmentId)

			// Cast to concrete type to access GetDepartmentsDescendants method
			concreteService := departmentApiService.(*departments.DepartmentApiService)

			// Test GetDepartmentsDescendants for department with no projects
			descendantsResponse, err := concreteService.GetDepartmentsDescendants(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(descendantsResponse.Code).To(Equal(http.StatusOK))

			// Validate the response structure
			descendants := descendantsResponse.Body.(authorization_utils.Descendants)
			Expect(descendants).ToNot(BeNil())
			Expect(descendants.ProjectIds).ToNot(BeNil())
			Expect(len(descendants.ProjectIds)).To(Equal(0))
		})

		It("test get department descendants with large number of projects", func() {
			// Create a department
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentId := createResponse.Body.(api.Department).Id
			departmentIds = append(departmentIds, departmentId)

			// Create many projects in the department
			var projectIds []string
			for i := 0; i < 10; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, departmentId, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				projectIds = append(projectIds, created.Id)
			}

			concreteService := departmentApiService.(*departments.DepartmentApiService)

			descendantsResponse, err := concreteService.GetDepartmentsDescendants(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(descendantsResponse.Code).To(Equal(http.StatusOK))

			descendants := descendantsResponse.Body.(authorization_utils.Descendants)
			Expect(descendants).ToNot(BeNil())
			Expect(descendants.ProjectIds).ToNot(BeNil())

			returnedProjectIds := descendants.ProjectIds
			Expect(len(returnedProjectIds)).To(Equal(len(projectIds)))

			for _, expectedProjectId := range projectIds {
				Expect(returnedProjectIds).To(ContainElement(expectedProjectId))
			}

			// Clean up projects
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			}
		})

		It("test get department descendants cache behavior", func() {
			// Create a department
			request := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse, err := departmentApiService.CreateDepartment(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			departmentId := createResponse.Body.(api.Department).Id
			departmentIds = append(departmentIds, departmentId)

			// Create a project in the department
			projectName := createProjectName()
			projRequest := getProjectCreateRequest(projectName, projectName, departmentId, clusterId)
			createResponse, err = projectApiService.CreateProject(ctx, projRequest)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			projectId := createResponse.Body.(api.Project).Id

			// Cast to concrete type to access GetDepartmentsDescendants method
			concreteService := departmentApiService.(*departments.DepartmentApiService)

			// First call - should be a cache miss and populate cache
			descendantsResponse1, err := concreteService.GetDepartmentsDescendants(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(descendantsResponse1.Code).To(Equal(http.StatusOK))

			descendants1 := descendantsResponse1.Body.(authorization_utils.Descendants)
			Expect(descendants1).ToNot(BeNil())
			Expect(descendants1.ProjectIds).ToNot(BeNil())
			Expect(len(descendants1.ProjectIds)).To(Equal(1))
			Expect((descendants1.ProjectIds)[0]).To(Equal(projectId))

			// Second call - should be a cache hit
			descendantsResponse2, err := concreteService.GetDepartmentsDescendants(ctx, departmentId)
			Expect(err).To(BeNil())
			Expect(descendantsResponse2.Code).To(Equal(http.StatusOK))

			descendants2 := descendantsResponse2.Body.(authorization_utils.Descendants)
			Expect(descendants2).ToNot(BeNil())
			Expect(descendants2.ProjectIds).ToNot(BeNil())
			Expect(len(descendants2.ProjectIds)).To(Equal(1))
			Expect((descendants2.ProjectIds)[0]).To(Equal(projectId))

			// Clean up project
			deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("test get department descendants with different departments isolation", func() {
			// Create first department
			request1 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse1, err := departmentApiService.CreateDepartment(ctx, request1)
			Expect(err).To(BeNil())
			Expect(createResponse1.Code).To(Equal(http.StatusCreated))
			department1Id := createResponse1.Body.(api.Department).Id
			departmentIds = append(departmentIds, department1Id)

			// Create second department
			request2 := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			createResponse2, err := departmentApiService.CreateDepartment(ctx, request2)
			Expect(err).To(BeNil())
			Expect(createResponse2.Code).To(Equal(http.StatusCreated))
			department2Id := createResponse2.Body.(api.Department).Id
			departmentIds = append(departmentIds, department2Id)

			// Create projects in both departments
			var department1Projects []string
			var department2Projects []string

			// Create 2 projects in department 1
			for i := 0; i < 2; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department1Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department1Projects = append(department1Projects, created.Id)
			}

			// Create 3 projects in department 2
			for i := 0; i < 3; i++ {
				projectName := createProjectName()
				projRequest := getProjectCreateRequest(projectName, projectName, department2Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, projRequest)
				Expect(err).To(BeNil())
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				created := createResponse.Body.(api.Project)
				department2Projects = append(department2Projects, created.Id)
			}

			// Cast to concrete type to access GetDepartmentsDescendants method
			concreteService := departmentApiService.(*departments.DepartmentApiService)

			// Test GetDepartmentsDescendants for department 1
			descendantsResponse1, err := concreteService.GetDepartmentsDescendants(ctx, department1Id)
			Expect(err).To(BeNil())
			Expect(descendantsResponse1.Code).To(Equal(http.StatusOK))

			descendants1 := descendantsResponse1.Body.(authorization_utils.Descendants)
			Expect(descendants1).ToNot(BeNil())
			Expect(descendants1.ProjectIds).ToNot(BeNil())
			Expect(len(descendants1.ProjectIds)).To(Equal(2))

			// Test GetDepartmentsDescendants for department 2
			descendantsResponse2, err := concreteService.GetDepartmentsDescendants(ctx, department2Id)
			Expect(err).To(BeNil())
			Expect(descendantsResponse2.Code).To(Equal(http.StatusOK))

			descendants2 := descendantsResponse2.Body.(authorization_utils.Descendants)
			Expect(descendants2).ToNot(BeNil())
			Expect(descendants2.ProjectIds).ToNot(BeNil())
			Expect(len(descendants2.ProjectIds)).To(Equal(3))

			// Verify isolation - department 1 should not have department 2's projects
			returnedProjectIds1 := descendants1.ProjectIds
			for _, department2ProjectId := range department2Projects {
				Expect(returnedProjectIds1).ToNot(ContainElement(department2ProjectId))
			}

			// Verify isolation - department 2 should not have department 1's projects
			returnedProjectIds2 := descendants2.ProjectIds
			for _, department1ProjectId := range department1Projects {
				Expect(returnedProjectIds2).ToNot(ContainElement(department1ProjectId))
			}

			// Clean up projects
			allProjects := append(department1Projects, department2Projects...)
			for _, projectId := range allProjects {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
			}
		})
	})

	Context("test get departments ranking", func() {
		var (
			testDepartments    []api.Department
			d0, d1, d2, d3, d4 string
			rankingProjectIds  []string
		)

		BeforeAll(func() {
			rankingProjectApiService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			for i := 0; i < 5; i++ {
				request := getDepartmentCreateRequest(fmt.Sprintf("d%d-%s", i, rand.String(3)), clusterId)
				createResponse, err := departmentApiService.CreateDepartment(ctx, request)
				if err != nil {
					Fail("Failed creating department")
				}
				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				testDepartments = append(testDepartments, createResponse.Body.(api.Department))
			}

			result := map[string]map[string]*models.OvertimeData{clusterId: {}}

			var i float32 = 0
			for i = 0; i < 5; i++ {
				department := testDepartments[int(i)]

				result[clusterId][department.Name] = &models.OvertimeData{
					AvgGpuAllocation24h:        ptr.To(float64(i * 10)),            // 0, 10, 20, 30, 40
					AvgGpuUtilization24h:       ptr.To(float64(i * 20)),            // 0, 20, 40, 60, 80
					AvgGpuMemoryUtilization24h: ptr.To(float64(i * 15)),            // 0, 15, 30, 45, 60
					AvgGpuAllocation7d:         ptr.To(float64(mod(2-i, 10) * 20)), // 40, 20, 0, 180, 160
					AvgGpuUtilization7d:        ptr.To(float64(mod(2-i, 10) * 20)), // 20, 10, 0, 180, 160
					AvgGpuMemoryUtilization7d:  ptr.To(float64(mod(2-i, 10) * 15)), // 30, 15, 0, 135, 120
					AvgGpuAllocation30d:        ptr.To(float64((5 - i) * 20)),      // 100, 80, 60, 40, 20
					AvgGpuUtilization30d:       ptr.To(float64((5 - i) * 10)),      // 50, 40, 30, 20, 10
					AvgGpuMemoryUtilization30d: ptr.To(float64((5 - i) * 15)),      // 75, 60, 45, 30, 15
				}
			}

			err := dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeDay)
			Expect(err).To(BeNil())
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeWeek)
			Expect(err).To(BeNil())
			err = dao.SetDepartmentsOvertimeDataForTenant(ctx, TestTenant, result, dao.RangeMonth)
			Expect(err).To(BeNil())

			// Create one child project per department and seed distinct gpu_allocated so ratio
			// ordering can be validated: dept[i] gets i*0.4 allocated, quota=2 → ratio 0, 0.2, 0.4, 0.6, 0.8
			for idx, dept := range testDepartments {
				projReq := getProjectCreateRequest(fmt.Sprintf("ratio-child-%d-%s", idx, rand.String(3)), "", dept.Id, clusterId)
				projResp, err := rankingProjectApiService.CreateProject(ctx, projReq)
				if err != nil {
					Fail("Failed creating child project for dept ranking")
				}
				Expect(projResp.Code).To(Equal(http.StatusCreated))
				childProject := projResp.Body.(api.Project)
				rankingProjectIds = append(rankingProjectIds, childProject.Id)

				allocValue := float64(idx) * 0.4 // 0.0, 0.4, 0.8, 1.2, 1.6
				Expect(runaidb.DB().Actual().Exec(
					"UPDATE quota_status SET gpu_allocated = ? FROM project WHERE project.quota_status_id = quota_status.id AND project.id::text = ?",
					allocValue, childProject.Id,
				).Error).To(BeNil())
			}

			d0 = testDepartments[0].Name
			d1 = testDepartments[1].Name
			d2 = testDepartments[2].Name
			d3 = testDepartments[3].Name
			d4 = testDepartments[4].Name
		})

		AfterAll(func() {
			rankingProjectApiService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
			for _, projectId := range rankingProjectIds {
				_, _ = rankingProjectApiService.DeleteProject(ctx, projectId, "", false)
			}
			rankingProjectIds = nil
			for _, department := range testDepartments {
				deleteResponse, err := departmentApiService.DeleteDepartment(ctx, department.Id)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			testDepartments = []api.Department{}
		})

		It("should get top 5 departments by GPU allocation 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d4))
			Expect(ranking[1].Name).To(Equal(d3))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d0))
		})

		It("should get bottom 5 departments by GPU allocation 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d0))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d3))
			Expect(ranking[4].Name).To(Equal(d4))
		})

		It("should get top 5 departments by GPU utilization 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d4))
			Expect(ranking[1].Name).To(Equal(d3))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d0))
			Expect(ranking[4].Name).To(Equal(d0))
		})

		It("should get bottom 5 departments by GPU utilization 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d0))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d3))
			Expect(ranking[4].Name).To(Equal(d4))
		})

		It("should get top 5 departments by GPU memory utilization 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d4))
			Expect(ranking[1].Name).To(Equal(d3))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d0))
			Expect(ranking[4].Name).To(Equal(d0))
		})

		It("should get bottom 5 departments by GPU memory utilization 24h", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d0))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d3))
			Expect(ranking[4].Name).To(Equal(d4))
		})

		It("should get top 5 departments by GPU allocation 7d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d3))
			Expect(ranking[1].Name).To(Equal(d4))
			Expect(ranking[2].Name).To(Equal(d0))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d2))
		})

		It("should get bottom 5 departments by GPU allocation 7d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d2))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d0))
			Expect(ranking[3].Name).To(Equal(d4))
			Expect(ranking[4].Name).To(Equal(d3))
		})

		It("should get top 5 departments by GPU memory utilization 7d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d3))
			Expect(ranking[1].Name).To(Equal(d4))
			Expect(ranking[2].Name).To(Equal(d0))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d2))
		})

		It("should get bottom 5 departments by GPU memory utilization 7d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d2))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d0))
			Expect(ranking[3].Name).To(Equal(d4))
			Expect(ranking[4].Name).To(Equal(d3))
		})

		It("should get top 5 departments by GPU memory utilization 30d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d0))
			Expect(ranking[1].Name).To(Equal(d1))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d3))
			Expect(ranking[4].Name).To(Equal(d4))
		})

		It("should get bottom 5 departments by GPU memory utilization 30d", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			Expect(len(ranking)).To(Equal(5))
			Expect(ranking[0].Name).To(Equal(d4))
			Expect(ranking[1].Name).To(Equal(d3))
			Expect(ranking[2].Name).To(Equal(d2))
			Expect(ranking[3].Name).To(Equal(d1))
			Expect(ranking[4].Name).To(Equal(d0))
		})

		It("should return GPU quota for each department in ranking", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			for _, department := range ranking {
				Expect(department.GpuQuota).NotTo(BeNil(), "GPU quota should not be nil for department %s", department.Name)
				Expect(*department.GpuQuota).To(Equal(float64(2)), "GPU quota should be 2 for department %s", department.Name)
			}
		})

		It("should get departments ranked by GPU allocation ratio (bottom = lowest first)", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments
			Expect(len(ranking)).To(Equal(5))
			// dept[0] has ratio 0.0, dept[4] has ratio 0.8 — BOTTOM returns lowest first
			for _, dept := range ranking {
				Expect(dept.AllocationValue).NotTo(BeNil())
			}
			Expect(ranking[0].Name).To(Equal(d0), "BOTTOM should start with lowest ratio (d0, ratio=0)")
			Expect(ranking[4].Name).To(Equal(d4), "BOTTOM should end with highest ratio (d4, ratio=0.8)")
			for i := 1; i < len(ranking); i++ {
				Expect(*ranking[i-1].AllocationValue).To(BeNumerically("<=", *ranking[i].AllocationValue))
			}
		})

		It("should get departments ranked by GPU allocation ratio (top = highest first)", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments
			Expect(len(ranking)).To(Equal(5))
			// dept[4] has ratio 0.8, dept[0] has ratio 0.0 — TOP returns highest first
			for _, dept := range ranking {
				Expect(dept.AllocationValue).NotTo(BeNil())
			}
			Expect(ranking[0].Name).To(Equal(d4), "TOP should start with highest ratio (d4, ratio=0.8)")
			Expect(ranking[4].Name).To(Equal(d0), "TOP should end with lowest ratio (d0, ratio=0)")
			for i := 1; i < len(ranking); i++ {
				Expect(*ranking[i-1].AllocationValue).To(BeNumerically(">=", *ranking[i].AllocationValue))
			}
		})

		It("should handle validation error gracefully", func() {
			response, err := departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, api.RANKINGORDERTYPE_BOTTOM, "invalid-cluster-id", "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = departmentApiService.GetDepartmentsRanking(ctx, api.RankingMetricType("blabla"), api.RANKINGORDERTYPE_BOTTOM, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = departmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H, "wrongggg", clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return department ranking with workload counts enrichment", func() {
			workloadsClient := &workloadsMocks.IWorkloadsClient{}

			department1Id := testDepartments[0].Id
			department2Id := testDepartments[1].Id

			workloadsClient.On("InternalEnrichOrgUnitRanking", mock.Anything, mock.MatchedBy(func(req workloadsOpenapi.InternalEnrichOrgUnitRankingRequest) bool {
				return req.ClusterId == clusterId &&
					req.OrgUnitType == "department" &&
					len(req.Ids) > 0
			})).Return(&workloadsOpenapi.InternalEnrichOrgUnitRankingResponse{
				Items: []workloadsOpenapi.OrgUnitRankingItem{
					{
						Id:                    department1Id,
						RunningWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(10))),
						PendingWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(4))),
					},
					{
						Id:                    department2Id,
						RunningWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(7))),
						PendingWorkloadsCount: *workloadsOpenapi.NewNullableInt32(ptr.To(int32(3))),
					},
				},
			}, nil)

			enrichedDepartmentApiService := departments.NewDepartmentApiService(&clusterServiceClient, workloadsClient)

			response, err := enrichedDepartmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments

			var department1Ranking, department2Ranking *api.DepartmentGetRanking
			for i := range ranking {
				if ranking[i].Id == department1Id {
					department1Ranking = &ranking[i]
				}
				if ranking[i].Id == department2Id {
					department2Ranking = &ranking[i]
				}
			}

			Expect(department1Ranking).NotTo(BeNil())
			Expect(department1Ranking.RunningWorkloadsCount).NotTo(BeNil())
			Expect(*department1Ranking.RunningWorkloadsCount).To(Equal(int32(10)))
			Expect(department1Ranking.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*department1Ranking.PendingWorkloadsCount).To(Equal(int32(4)))

			Expect(department2Ranking).NotTo(BeNil())
			Expect(department2Ranking.RunningWorkloadsCount).NotTo(BeNil())
			Expect(*department2Ranking.RunningWorkloadsCount).To(Equal(int32(7)))
			Expect(department2Ranking.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*department2Ranking.PendingWorkloadsCount).To(Equal(int32(3)))

			workloadsClient.AssertExpectations(GinkgoT())
		})

		It("should handle workload enrichment failure gracefully for departments", func() {
			// Mock workloads client that returns an error
			workloadsClient := &workloadsMocks.IWorkloadsClient{}
			workloadsClient.On("InternalEnrichOrgUnitRanking", mock.Anything, mock.Anything).Return(
				(*workloadsOpenapi.InternalEnrichOrgUnitRankingResponse)(nil),
				fmt.Errorf("workloads service unavailable"),
			)

			// Create service with failing workloads client
			enrichedDepartmentApiService := departments.NewDepartmentApiService(&clusterServiceClient, workloadsClient)

			// Call GetDepartmentsRanking - should succeed despite enrichment failure
			response, err := enrichedDepartmentApiService.GetDepartmentsRanking(ctx, api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H, api.RANKINGORDERTYPE_TOP, clusterId, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			ranking := response.Body.(api.GetDepartmentsRanking200Response).Departments
			Expect(len(ranking) > 0).To(BeTrue())

			// Verify workload counts are nil (enrichment failed)
			for _, department := range ranking {
				Expect(department.RunningWorkloadsCount).To(BeNil())
				Expect(department.PendingWorkloadsCount).To(BeNil())
			}
		})
	})
})

func getDepartmentCreateRequest(name string, clusterId string) api.DepartmentCreationRequest {
	return api.DepartmentCreationRequest{
		Name:        name,
		Description: fmt.Sprintf("%s department description", name),
		ClusterId:   clusterId,
		Resources:   createDefaultApiResources(false),
	}
}

func getFullDepartmentCreateRequest(name string, clusterId string) api.DepartmentCreationRequest {
	return api.DepartmentCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		Resources: createDefaultApiResources(true),
	}
}

func createDepartmentName() string {
	return fmt.Sprint("dep-", uuid.New())
}

func getDepartmentUpdateRequest(department api.Department) api.DepartmentUpdateRequest {
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
		// SchedulingRules: department.SchedulingRules,
		Resources: resources,
	}
}

func getDepartmentUpdateRequestFromListResponse(department api.DepartmentForList) api.DepartmentUpdateRequest {
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
		// SchedulingRules: department.SchedulingRules,
		Resources: resources,
	}
}

func getDepartmentNames(departments []api.DepartmentForList) []string {
	names := []string{}
	for _, department := range departments {
		names = append(names, department.Name)
	}
	return names
}
