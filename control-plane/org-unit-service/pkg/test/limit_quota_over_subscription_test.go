/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Limit quota over subscriptions tests", Ordered, func() {
	clusterId := uuid.New().String()
	clusterId2 := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var departmentId string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)
	nodePool1Name := "nodepool-a"
	nodePool1Id := float32(2)
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
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
			{
				Name:                   nodePool1Name,
				Id:                     fmt.Sprint(nodePool1Id),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)

	var (
		testProjectIds             []string
		departmentId2              string
		unlimitedDepartmentId      string
		depId2Int                  int64
		testDepIds                 []string
		projectIdToUpdate          int64
		projectIdToUpdateUnlimited int64
		dbProjectToUpdate          models.Project
		dbProjectToUpdateUnlimited models.Project
		projUpdateRequest          api.ProjectUpdateRequest
		projUpdateRequestUnlimited api.ProjectUpdateRequest
		projCreateRequest          api.ProjectCreationRequest
		projCreateRequestNew       api.ProjectCreationRequest
	)

	BeforeAll(func() {
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)

		createDepartmentRequest := getFullDepartmentCreateRequest("default", clusterId)
		createDepartmentRequest.Resources = getDep2Resources()
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId = response.Body.(api.Department).Id

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId, buildDisplayedCluster(clusterId, "2.17.0")).
			WithCluster(clusterId2, buildDisplayedCluster(clusterId2, "2.14.0")))
		checker.InitClusterCapabilitiesChecker("", getDefaultClustersCacheConfig(), "")

		createDepartmentRequest = getFullDepartmentCreateRequest("dep2", clusterId)
		createDepartmentRequest.Resources = getDep2Resources()
		response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating department dep2")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId2 = response.Body.(api.Department).Id
		testDepIds = append(testDepIds, departmentId2)

		depId2Int, err = strconv.ParseInt(departmentId2, 10, 64)
		Expect(err).To(BeNil())

		createDepartmentRequest = getFullDepartmentCreateRequest("dep3", clusterId)
		createDepartmentRequest.Resources = getDepResourcesUnlimited()
		response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating department dep2")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		unlimitedDepartmentId = response.Body.(api.Department).Id
		testDepIds = append(testDepIds, unlimitedDepartmentId)

		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

		projectsCreationRequests := []api.ProjectCreationRequest{getProj1CreationRequest(clusterId, departmentId2), getProj2CreationRequest(clusterId, departmentId2), getProj3CreationRequest(clusterId, unlimitedDepartmentId)}
		for _, projReq := range projectsCreationRequests {
			response, err := projectApiService.CreateProject(ctx, projReq)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			testProjectIds = append(testProjectIds, response.Body.(api.Project).Id)
		}

		projectIdToUpdateInt, err := strconv.Atoi(testProjectIds[0])
		Expect(err).To(BeNil())
		projectIdToUpdate = int64(projectIdToUpdateInt)

		dbProject, err := dao.GetProjectById(ctx, projectIdToUpdate, TestTenant, models.SlimVerbosity)
		Expect(err).To(BeNil())
		dbProjectToUpdate = *dbProject

		projectIdToUpdateInt, err = strconv.Atoi(testProjectIds[2])
		Expect(err).To(BeNil())
		projectIdToUpdateUnlimited = int64(projectIdToUpdateInt)

		dbProject, err = dao.GetProjectById(ctx, projectIdToUpdateUnlimited, TestTenant, models.SlimVerbosity)
		Expect(err).To(BeNil())
		dbProjectToUpdateUnlimited = *dbProject
	})

	BeforeEach(func() {
		// The over-subscription enforcement is gated on this setting; default it on for the whole
		// Describe (mirrors the previous always-true mock). Specs that exercise the disabled path
		// toggle it off within their own body.
		setLimitQuotaOverSubscription(true)
		projUpdateRequest = getProjUpdateRequest(modelResourcesDeservedToApiDeserved(dbProjectToUpdate.OrgUnitData.Resources))
		projUpdateRequestUnlimited = getProjUpdateRequest(modelResourcesDeservedToApiDeserved(dbProjectToUpdateUnlimited.OrgUnitData.Resources))
		projCreateRequest = getProjToCreateCreationRequest(clusterId, departmentId2)
		projCreateRequestNew = getProjToCreateCreationRequest(clusterId, departmentId)
	})

	AfterAll(func() {
		// Restore the default so the enabled flag doesn't leak into other Describe blocks
		// (the settings package is global/DB-backed and shared across the whole test binary).
		setLimitQuotaOverSubscription(false)

		for _, projectId := range testProjectIds {
			deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
		}
		for _, depId := range testDepIds {
			_, err := departmentApiService.DeleteDepartment(ctx, depId)
			if err != nil {
				Fail("Failed deleting department")
			}
		}

		_, err := departmentApiService.DeleteDepartment(ctx, departmentId2)
		if err != nil {
			Fail("Failed deleting department")
		}
		_, err = departmentApiService.DeleteDepartment(ctx, departmentId)
		if err != nil {
			Fail("Failed deleting department")
		}
	})

	Context("test validateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible - unlimited department", func() {
		var (
			projQuotaStatuses = []api.QuotaStatusNodePool{
				{
					NodePoolName: "default",
					NodePoolId:   "1",
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    1000,
						Memory: 100,
					},
				},
				{
					NodePoolName: "nodepool-a",
					NodePoolId:   "2",
					AllocatedNonPreemptible: api.QuotaStatusResource{
						Gpu:    1,
						Cpu:    1000,
						Memory: 100,
					},
				},
			}

			departmentId2Cluster2 string
			projCluster2Id        string
		)

		BeforeAll(func() {
			updateProjectStatus(ctx, projectApiService, strconv.Itoa(int(dbProjectToUpdateUnlimited.ID)), dbProjectToUpdateUnlimited.ClusterId.String(), projQuotaStatuses)

			dbProject, err := dao.GetProjectById(ctx, projectIdToUpdateUnlimited, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			dbProjectToUpdateUnlimited = *dbProject

			// setup dep+proj in another cluster
			createDepartmentRequest := getFullDepartmentCreateRequest("dep2", clusterId2)
			createDepartmentRequest.Resources = getDep2Resources()
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating department dep2 in cluster2")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			departmentId2Cluster2 = response.Body.(api.Department).Id
			projCreationn := getProj1CreationRequest(clusterId2, departmentId2Cluster2)
			response, err = projectApiService.CreateProject(ctx, projCreationn)
			if err != nil {
				Fail("Failed creating project in cluster2")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			projCluster2Id = response.Body.(api.Project).Id
			updateProjectStatus(ctx, projectApiService, projCluster2Id, clusterId2, projQuotaStatuses)
		})

		AfterAll(func() {
			updateProjectStatus(ctx, projectApiService, strconv.Itoa(int(dbProjectToUpdateUnlimited.ID)), dbProjectToUpdateUnlimited.ClusterId.String(),
				[]api.QuotaStatusNodePool{})

			dbProject, err := dao.GetProjectById(ctx, projectIdToUpdateUnlimited, TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			dbProjectToUpdateUnlimited = *dbProject

			_, err = departmentApiService.DeleteDepartment(ctx, departmentId2Cluster2)
			if err != nil {
				Fail("Failed deleting department")
			}

			_, err = projectApiService.DeleteProject(ctx, projCluster2Id, "", false)
			if err != nil {
				Fail("Failed deleting project")
			}
		})

		It("sanity - happy flow", func() {
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, dbProjectToUpdate.OrgUnitData.Resources, dbProjectToUpdateUnlimited)
			Expect(apiErr).To(BeNil())
		})

		It("sanity of failure with API", func() {
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			strProjectIdToUpdate := strconv.FormatInt(projectIdToUpdateUnlimited, 10)

			// block because of bad deserved gpus
			res, err := projectApiService.UpdateProject(ctx, strProjectIdToUpdate, api.ProjectUpdateRequest{Resources: projUpdateReqCopy.Resources})
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateProject to fail")
			res, err = projectApiService.UpdateProjectResources(ctx, strProjectIdToUpdate, projUpdateReqCopy.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateProjectResources to fail")
			res, err = projectApiService.PatchProjectResources(ctx, strProjectIdToUpdate, projUpdateReqCopy.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected PatchProjectResources to fail")
		})

		It("test blocks", func() {
			// block because of bad deserved gpus
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved GPUs for node pool 'default' because the requested amount is below the usage of non-preemptive workloads. The minimum allowed value is '1.00'."))

			// block because of bad deserved cpus
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[1].Cpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved CPUs for node pool 'nodepool-a' because the requested amount is below the usage of non-preemptive workloads. The minimum allowed value is '1000.00 MilliCores'."))

			// block because of bad deserved memory
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[0].Memory.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'default' because the requested amount is below the usage of non-preemptive workloads. The minimum allowed value is '100.00 Mb'."))

			// feature disabled
			setLimitQuotaOverSubscription(false)
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).To(BeNil())

			setLimitQuotaOverSubscription(true)

			// unsupported cluster version - should pass
			projCluster2IdInt, err := strconv.Atoi(projCluster2Id)
			Expect(err).To(BeNil())
			dbProjectInCluster2ToUpdate, err := dao.GetProjectById(ctx, int64(projCluster2IdInt), TestTenant, models.SlimVerbosity)
			Expect(err).To(BeNil())
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), *dbProjectInCluster2ToUpdate)
			Expect(apiErr).To(BeNil())
		})

		It("test allow", func() {
			// allow for unlimited cpu
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[1].Cpu.Deserved = ptr.To(float64(-1))
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).To(BeNil())
			projUpdateReqCopy.Resources[1].Cpu.Deserved = nil
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).To(BeNil())

			// no quotaStatuses at all - allow
			projQuotaStatusesCopy := projQuotaStatuses[1:]
			updateProjectStatus(ctx, projectApiService, strconv.Itoa(int(dbProjectToUpdateUnlimited.ID)), dbProjectToUpdateUnlimited.ClusterId.String(), projQuotaStatusesCopy)
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequestUnlimited)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateRequestUnlimited.Resources), dbProjectToUpdateUnlimited)
			Expect(apiErr).To(BeNil())
		})
	})

	Context("test validateProjectUpdateNotCausingOverSubscriptionOfDepartment", func() {
		BeforeEach(func() {
			setLimitQuotaOverSubscription(true)
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, getDep2Resources())
		})

		It("sanity - happy flow", func() {
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateRequest.Resources), dbProjectToUpdate)
			Expect(apiErr).To(BeNil())
		})

		It("sanity of failure with API", func() {
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(10))
			strProjectIdToUpdate := strconv.FormatInt(projectIdToUpdate, 10)

			// block because of bad deserved gpus
			res, err := projectApiService.UpdateProject(ctx, strProjectIdToUpdate, api.ProjectUpdateRequest{Resources: projUpdateReqCopy.Resources})
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateProject to fail")
			res, err = projectApiService.UpdateProjectResources(ctx, strProjectIdToUpdate, projUpdateReqCopy.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateProjectResources to fail")
			res, err = projectApiService.PatchProjectResources(ctx, strProjectIdToUpdate, projUpdateReqCopy.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected PatchProjectResources to fail")
		})

		It("test blocks", func() {
			// block because of bad deserved gpus
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Gpu.Deserved = ptr.To(float64(10))
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved GPUs for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '4.00'."))

			// block because of bad deserved cpus
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Cpu.Deserved = ptr.To(float64(10000))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved CPUs for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '4000.00 MilliCores'."))

			// block because of bad deserved memory
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Memory.Deserved = ptr.To(float64(1000))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '400.00 Mb'."))

			// block because of bad deserved memory (unlimited -1)
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[1].Memory.Deserved = ptr.To(float64(-1))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '100.00 Mb'."))
		})

		It("test allow", func() {
			// allow if department unlimited deserved
			newResources := getDep2Resources()
			newResources[1].Memory.Deserved = ptr.To(float64(-1))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[1].Memory.Deserved = ptr.To(float64(10000))
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).To(BeNil())
			projUpdateReqCopy.Resources[1].Memory.Deserved = ptr.To(float64(-1))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).To(BeNil())

			// allow for unlimited cpu
			newResources = getDep2Resources()
			newResources[0].Cpu.Deserved = nil
			newResources[0].Cpu.Limit = nil
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[0].Cpu.Deserved = nil
			projUpdateReqCopy.Resources[0].Cpu.Limit = nil
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).To(BeNil())
		})

		It("test blocks 2", func() {
			// block because no node pool resources for cpu - means unlimited
			setLimitQuotaOverSubscription(false)
			newResources := getDep2Resources()
			newResources[1].Cpu.Deserved = ptr.To(float64(0))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			setLimitQuotaOverSubscription(true)

			projUpdateReqCopy := copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[1].Cpu.Deserved = nil
			projUpdateReqCopy.Resources[1].Cpu.Limit = nil
			apiErr := services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved CPUs for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00 MilliCores'."))

			// block because no node pool resources for memory - means unlimited
			setLimitQuotaOverSubscription(false)
			newResources = getDep2Resources()
			newResources[1].Memory.Deserved = ptr.To(float64(0))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			setLimitQuotaOverSubscription(true)

			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[1].Memory = nil
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00 Mb'."))

			// block because of bad deserved gpus - max allowed deserved is 0
			setLimitQuotaOverSubscription(false)
			newResources = getDep2Resources()
			newResources[1].Gpu.Deserved = ptr.To(float64(0))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			setLimitQuotaOverSubscription(true)

			projUpdateReqCopy = copyProjUpdateReqeust(&projUpdateRequest)
			projUpdateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(1))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved GPUs for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00'."))

			// block modifying project because of bad deserved gpus in original project
			projUpdateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify project until department's projects quotas of GPUs for node pool 'nodepool-a' don't exceed the department's available quota."))

			// feature disabled - same API call should pass
			setLimitQuotaOverSubscription(false)
			apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projUpdateReqCopy.Resources), dbProjectToUpdate)
			Expect(apiErr).To(BeNil())
		})
	})

	Context("test ValidateProjectCreateNotCausingOverSubscriptionOfDepartment", func() {
		BeforeEach(func() {
			setLimitQuotaOverSubscription(true)
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, getDep2Resources())
		})

		It("sanity - happy flow", func() {
			apiErr := services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateRequest.Resources), *projCreateRequest.ParentId)
			Expect(apiErr).To(BeNil())
		})

		It("sanity of failure with API", func() {
			projCreateReqCopy := copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(10))

			// block because of bad deserved gpus
			res, err := projectApiService.CreateProject(ctx, projCreateReqCopy)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected CreateProject to fail")
		})

		It("sanity of failure with new project", func() {
			projCreateReqCopy := copyProjCreateRequest(&projCreateRequestNew)
			projCreateReqCopy.Resources[1].Cpu.Deserved = ptr.To(float64(5000))
			projCreateReqCopy.Resources[1].Cpu.Limit = ptr.To(float64(5000))

			// block because of bad deserved gpus
			res, err := projectApiService.CreateProject(ctx, projCreateReqCopy)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected CreateProject to fail")
		})

		It("test blocks", func() {
			parentId := *projCreateRequest.ParentId

			// block because of bad deserved gpus
			projCreateReqCopy := copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(10))
			apiErr := services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved GPUs for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00'."))

			// block because of bad deserved cpus
			projCreateReqCopy = copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[1].Cpu.Deserved = ptr.To(float64(10000))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved CPUs for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00 MilliCores'."))

			// block because of bad deserved memory
			projCreateReqCopy = copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[0].Memory.Deserved = ptr.To(float64(1000))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved memory for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '300.00 Mb'."))
			projCreateReqCopy.Resources[0].Memory.Deserved = nil
			projCreateReqCopy.Resources[0].Memory.Limit = nil
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved memory for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '300.00 Mb'."))

			// block because of bad deserved cpu (unlimited -1)
			projCreateReqCopy = copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[0].Cpu.Deserved = ptr.To(float64(-1))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved CPUs for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '3000.00 MilliCores'."))
			projCreateReqCopy.Resources[0].Cpu = nil
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved CPUs for node pool 'default' because the requested amount exceeds the department's available quota. The maximum allowed value is '3000.00 MilliCores'."))
		})

		It("test allow", func() {
			parentId := *projCreateRequest.ParentId

			// allow if department unlimited deserved
			newResources := getDep2Resources()
			newResources[0].Cpu.Deserved = ptr.To(float64(-1))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			projCreateReqCopy := copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[0].Cpu.Deserved = ptr.To(float64(-1))
			apiErr := services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).To(BeNil())

			// no node pool resources for cpu/memory - allow
			newResources = getDep2Resources()
			newResources[0].Cpu.Deserved = ptr.To(float64(-1))
			newResources[0].Memory.Deserved = nil
			newResources[0].Memory.Limit = nil
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			projCreateReqCopy = copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[0].Cpu.Deserved = ptr.To(float64(-1))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).To(BeNil())
			projCreateReqCopy.Resources[0].Cpu = nil
			projCreateReqCopy.Resources[0].Memory = nil
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).To(BeNil())

			// no node pool resources for gpu - allow
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, getDep2Resources())
			projCreateReqCopy = copyProjCreateRequest(&projCreateRequest)
			projCreateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).To(BeNil())
		})

		It("test block2", func() {
			parentId := *projCreateRequest.ParentId

			// block creating project because of bad deserved gpus in original department
			setLimitQuotaOverSubscription(false)
			newResources := getDep2Resources()
			newResources[1].Gpu.Deserved = ptr.To(float64(0))
			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, newResources)
			setLimitQuotaOverSubscription(true)

			projCreateReqCopy := copyProjCreateRequest(&projCreateRequest)
			apiErr := services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot create project until department's projects quotas of GPUs for node pool 'nodepool-a' don't exceed the department's available quota."))

			// block because of bad deserved gpus - max allowed deserved is 0
			projCreateReqCopy.Resources[1].Gpu.Deserved = ptr.To(float64(1))
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot set the deserved GPUs for node pool 'nodepool-a' because the requested amount exceeds the department's available quota. The maximum allowed value is '0.00'."))

			// feature disabled - same API call should pass
			setLimitQuotaOverSubscription(false)
			apiErr = services.ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(projCreateReqCopy.Resources), parentId)
			Expect(apiErr).To(BeNil())
		})
	})

	Context("test ValidateDepartmentUpdateNotCausingOverSubscription", func() {
		var (
			dbDepartment     *models.Department
			depUpdateReq     api.DepartmentUpdateRequest
			departmentId2Int int
		)

		BeforeAll(func() {
			var err error
			departmentId2Int, err = strconv.Atoi(departmentId2)
			Expect(err).To(BeNil())

			dbDepartment, err = dao.GetDepartmentById(ctx, int64(departmentId2Int), TestTenant, models.VerboseVerbosity, nil)
			Expect(err).To(BeNil())

			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
		})

		BeforeEach(func() {
			setLimitQuotaOverSubscription(true)

			updateDepartmentWithResources(ctx, departmentApiService, departmentId2, getDep2Resources())

			updateProjectWithResources(ctx, projectApiService, testProjectIds[1], getProj2Resources())

			var err error
			dbDepartment, err = dao.GetDepartmentById(ctx, int64(departmentId2Int), TestTenant, models.VerboseVerbosity, nil)
			Expect(err).To(BeNil())
		})

		It("sanity - happy flow", func() {
			apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())
		})

		It("sanity of failure with API", func() {
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			strDepIdToUpdate := strconv.FormatInt(dbDepartment.ID, 10)

			// block because of bad deserved gpus
			res, err := departmentApiService.UpdateDepartment(ctx, strDepIdToUpdate, depUpdateReq)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateDepartment to fail")
			res, err = departmentApiService.UpdateDepartmentResources(ctx, strDepIdToUpdate, depUpdateReq.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected UpdateDepartmentResources to fail")
			res, err = departmentApiService.PatchDepartmentResources(ctx, strDepIdToUpdate, depUpdateReq.Resources)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest), "Expected PatchDepartmentResources to fail")
		})

		It("test allow", func() {
			// allow for unlimited deserved
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Memory.Deserved = ptr.To(float64(-1))
			apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())
			depUpdateReq.Resources[0].Memory.Deserved = nil
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())
			depUpdateReq.Resources[0].Memory = nil
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())

			// updating unrelevant value - allow
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Gpu.Limit = ptr.To(float64(0))
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())
		})

		It("test blocks", func() {
			// block because of bad deserved gpus
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved GPUs for node pool 'default' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is '1.00'."))

			// block because of bad deserved cpus
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Cpu.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved CPUs for node pool 'default' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is '1000.00 MilliCores'."))

			// block because of bad deserved memory
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[0].Memory.Deserved = ptr.To(float64(0))
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'default' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is '100.00 Mb'."))
		})

		It("test block then allow", func() {
			// block because of bad deserved gpus
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[1].Gpu.Deserved = ptr.To(float64(1))
			apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved GPUs for node pool 'nodepool-a' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is '2.00'."))

			// update proj2 so the aggregated for the whole department for nodepool-a is now "1" and not "2"
			proj2ResourcesToUpdate := getProj2Resources()
			proj2ResourcesToUpdate[1].Gpu.Deserved = ptr.To(float64(0))
			proj2ResourcesToUpdate[1].Cpu.Deserved = ptr.To(float64(0))
			proj2ResourcesToUpdate[1].Memory.Deserved = ptr.To(float64(0))
			updateProjectWithResources(ctx, projectApiService, testProjectIds[1], proj2ResourcesToUpdate)

			var err error
			dbDepartment, err = dao.GetDepartmentById(ctx, int64(departmentId2Int), TestTenant, models.VerboseVerbosity, nil)
			Expect(err).To(BeNil())

			// now should pass
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())

			// update proj2 memory to unlimited
			proj2ResourcesToUpdate[1].Memory.Deserved = ptr.To(float64(-1))
			proj2ResourcesToUpdate[1].Memory.Limit = ptr.To(float64(-1))
			setLimitQuotaOverSubscription(false)
			updateProjectWithResources(ctx, projectApiService, testProjectIds[1], proj2ResourcesToUpdate)
			setLimitQuotaOverSubscription(true)
			dbDepartment, err = dao.GetDepartmentById(ctx, int64(departmentId2Int), TestTenant, models.VerboseVerbosity, nil)
			Expect(err).To(BeNil())

			// block because of bad deserved memory - projects aggregated is -1
			depUpdateReq = api.DepartmentUpdateRequest{Resources: getDep2Resources()}
			depUpdateReq.Resources[1].Memory.Deserved = ptr.To(float64(1000))
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify the deserved memory for node pool 'nodepool-a' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is -1 (Unlimited)."))

			// allow if both are -1
			depUpdateReq.Resources[1].Memory.Deserved = ptr.To(float64(-1))
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())

			// block modifying department until deserved is fixed
			proj2ResourcesToUpdate[0].Gpu.Deserved = ptr.To(float64(4))
			proj2ResourcesToUpdate[0].Gpu.Limit = ptr.To(float64(4))
			setLimitQuotaOverSubscription(false)
			updateProjectWithResources(ctx, projectApiService, testProjectIds[1], proj2ResourcesToUpdate)
			setLimitQuotaOverSubscription(true)
			dbDepartment, err = dao.GetDepartmentById(ctx, int64(departmentId2Int), TestTenant, models.VerboseVerbosity, nil)
			Expect(err).To(BeNil())
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).ToNot(BeNil())
			Expect(apiErr.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal("Cannot modify department until department's deserved GPUs for node pool 'default' won't be below the sum of department's projects quotas, which is '5.00'."))

			// feature disabled
			setLimitQuotaOverSubscription(false)
			apiErr = services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depId2Int)
			Expect(apiErr).To(BeNil())
		})
	})

	Context("test large memory quota update causing int32 overflow", func() {
		var (
			testDepartmentId string
			testProjectId    string
			depIdInt         int64
		)

		BeforeAll(func() {
			setLimitQuotaOverSubscription(true)

			// Create department with memory quota 2000 GB in Mib units
			createDepartmentRequest := getFullDepartmentCreateRequest("dep-large-memory", clusterId)
			createDepartmentRequest.Resources = []api.ResourcesNullable{
				{
					NodePool: &api.ResourcesNodePool{
						Id:   "1",
						Name: "default",
					},
					Gpu: api.ResourcesNullableGpu{
						Limit:    ptr.To(float64(10)),
						Deserved: ptr.To(float64(10)),
					},
					Cpu: &api.ResourcesCpu{
						Limit:    ptr.To(float64(10000)),
						Deserved: ptr.To(float64(10000)),
					},
					Memory: &api.ResourcesMemory{
						Limit:    ptr.To(float64(1907348.633)), // 2000 GB in MIB
						Deserved: ptr.To(float64(1907348.633)),
						Units:    api.UNITS_MIB,
					},
				},
			}
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating department with large memory quota")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			testDepartmentId = response.Body.(api.Department).Id

			depIdIntParsed, err := strconv.ParseInt(testDepartmentId, 10, 64)
			Expect(err).To(BeNil())
			depIdInt = depIdIntParsed

			// Create project with memory quota 1900 GB in Mib units
			config := createDefaultApiJobConfig()
			projectCreationRequest := api.ProjectCreationRequest{
				Name:            "proj-large-memory",
				ClusterId:       clusterId,
				ParentId:        &testDepartmentId,
				SchedulingRules: &config,
				Resources: []api.ResourcesNullable{
					{
						NodePool: &api.ResourcesNodePool{
							Id:   "1",
							Name: "default",
						},
						Gpu: api.ResourcesNullableGpu{
							Limit:    ptr.To(float64(5)),
							Deserved: ptr.To(float64(5)),
						},
						Cpu: &api.ResourcesCpu{
							Limit:    ptr.To(float64(5000)),
							Deserved: ptr.To(float64(5000)),
						},
						Memory: &api.ResourcesMemory{
							Limit:    ptr.To(float64(1811981.201)), // 1900 GB in MIB
							Deserved: ptr.To(float64(1811981.201)),
							Units:    api.UNITS_MIB,
						},
					},
				},
			}
			response, err = projectApiService.CreateProject(ctx, projectCreationRequest)
			if err != nil {
				Fail("Failed creating project with large memory quota")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			testProjectId = response.Body.(api.Project).Id
		})

		AfterAll(func() {
			if testProjectId != "" {
				_, err := projectApiService.DeleteProject(ctx, testProjectId, "", false)
				if err != nil {
					Fail("Failed deleting test project")
				}
			}
			if testDepartmentId != "" {
				_, err := departmentApiService.DeleteDepartment(ctx, testDepartmentId)
				if err != nil {
					Fail("Failed deleting test department")
				}
			}
		})

		It("should allow increasing department memory quota from 2000 GB to 3000 GB when project uses 1900 GB", func() {
			// Attempt to update department memory quota from 2000 GB to 3000 GB, using Mib units
			depUpdateReq := api.DepartmentUpdateRequest{
				Resources: []api.ResourcesNullable{
					{
						NodePool: &api.ResourcesNodePool{
							Id:   "1",
							Name: "default",
						},
						Gpu: api.ResourcesNullableGpu{
							Limit:    ptr.To(float64(10)),
							Deserved: ptr.To(float64(10)),
						},
						Cpu: &api.ResourcesCpu{
							Limit:    ptr.To(float64(10000)),
							Deserved: ptr.To(float64(10000)),
						},
						Memory: &api.ResourcesMemory{
							Limit:    ptr.To(float64(2861022.94921875)), // 3000 GB in MIB
							Deserved: ptr.To(float64(2861022.94921875)),
							Units:    api.UNITS_MIB,
						},
					},
				},
			}

			// BUG: Due to (a previous!) int32 overflow in roundNumber(), this validation incorrectly failed.
			// The overflow caused 3072000 MB to appear smaller than 1945600 MB
			// This should succeed because 3000 GB > 1900 GB (we're increasing the quota)
			apiErr := services.ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx,
				TestTenant, apiResourcesDeservedToModelDeserved(depUpdateReq.Resources), depIdInt)
			Expect(apiErr).To(BeNil(), "Department update should succeed when increasing quota from 2000 GB to 3000 GB (project uses 1900 GB)")

			// Also test with the actual API
			res, err := departmentApiService.UpdateDepartment(ctx, testDepartmentId, depUpdateReq)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK), "UpdateDepartment should succeed when increasing quota")
		})
	})

	Context("test sanity with only 1 nodepool", func() {
		var testProjectIds2 []string

		BeforeAll(func() {
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)

			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterAll(func() {
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&openapi.GetNodepoolsMinimal200Response{
				Nodepools: []openapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
					{
						Name:                   nodePool1Name,
						Id:                     fmt.Sprint(nodePool1Id),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)

			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		AfterEach(func() {
			for _, projectId := range testProjectIds2 {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail(fmt.Sprintf("Failed deleting, code: %d, reason: %s", deleteResponse.Code, deleteResponse.Body.(api.Error).Message))
				}
			}
			testProjectIds2 = []string{}
		})

		It("sanity - create a new project", func() {
			projCreationRequest := api.ProjectCreationRequest{
				Name:      "proj111",
				ClusterId: clusterId,
				ParentId:  &departmentId,
				Resources: []api.ResourcesNullable{
					{
						NodePool: &api.ResourcesNodePool{
							Id:   "1",
							Name: "default",
						},
						Gpu: api.ResourcesNullableGpu{
							Limit:    ptr.To(float64(1)),
							Deserved: ptr.To(float64(1)),
						},
						Cpu: &api.ResourcesCpu{
							Limit:    ptr.To(float64(0)),
							Deserved: ptr.To(float64(0)),
						},
						Memory: &api.ResourcesMemory{
							Limit:    ptr.To(float64(0)),
							Deserved: ptr.To(float64(0)),
							Units:    api.UNITS_MB,
						},
					},
				},
			}
			response, err := projectApiService.CreateProject(ctx, projCreationRequest)
			if err != nil {
				Fail("Failed creating project in cluster2")
			}
			if response.Code != http.StatusCreated {
				Expect(response.Code).To(Equal(http.StatusCreated), fmt.Sprintf("response message: %s", response.Body.(api.Error).Message))
			}

			testProjectIds2 = append(testProjectIds2, response.Body.(api.Project).Id)
		})

		It("sanity - create a new project without nodepool id in resources", func() {
			projCreationRequest := api.ProjectCreationRequest{
				Name:      "proj111",
				ClusterId: clusterId,
				ParentId:  &departmentId,
				Resources: []api.ResourcesNullable{
					{
						Gpu: api.ResourcesNullableGpu{
							Limit:    ptr.To(float64(1)),
							Deserved: ptr.To(float64(1)),
						},
						Cpu: &api.ResourcesCpu{
							Limit:    ptr.To(float64(0)),
							Deserved: ptr.To(float64(0)),
						},
						Memory: &api.ResourcesMemory{
							Limit:    ptr.To(float64(0)),
							Deserved: ptr.To(float64(0)),
							Units:    api.UNITS_MB,
						},
					},
				},
			}
			response, err := projectApiService.CreateProject(ctx, projCreationRequest)
			if err != nil {
				Fail("Failed creating project in cluster2")
			}
			if response.Code != http.StatusCreated {
				Expect(response.Code).To(Equal(http.StatusCreated), fmt.Sprintf("response message: %s", response.Body.(api.Error).Message))
			}

			testProjectIds2 = append(testProjectIds2, response.Body.(api.Project).Id)
		})
	})
})

func modelResourcesDeservedToApiDeserved(resources []models.Resources) []api.ResourcesNullable {
	apiResources := make([]api.ResourcesNullable, len(resources))

	for i := range resources {
		apiResources[i] = modelResourceDeservedToApiDeserved(resources[i])
	}
	sort.Slice(apiResources, func(i, j int) bool {
		return apiResources[i].NodePool.Id < apiResources[j].NodePool.Id
	})

	return apiResources
}

func modelResourceDeservedToApiDeserved(resource models.Resources) api.ResourcesNullable {
	apiResources := api.ResourcesNullable{
		NodePool: &api.ResourcesNodePool{
			Id:   resource.NodePoolId,
			Name: resource.NodePoolName,
		},
		Gpu: api.ResourcesNullableGpu{
			Deserved: ptr.To(*resource.GpuDeserved),
		},
	}

	if resource.CpuDeserved != nil {
		apiResources.Cpu = &api.ResourcesCpu{
			Deserved: resource.CpuDeserved,
		}
	}
	if resource.MemoryDeserved != nil {
		apiResources.Memory = &api.ResourcesMemory{
			Deserved: resource.MemoryDeserved,
			Units:    services.ModelUnitsToApi(resource.MemoryUnits),
		}
	}

	return apiResources
}

func getDep2Resources() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "1",
				Name: "default",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(4)),
				Deserved: ptr.To(float64(4)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(4000)),
				Deserved: ptr.To(float64(4000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(400)),
				Deserved: ptr.To(float64(400)),
				Units:    api.UNITS_MB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "2",
				Name: "nodepool-a",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(2)),
				Deserved: ptr.To(float64(2)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(2000)),
				Deserved: ptr.To(float64(2000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(200)),
				Deserved: ptr.To(float64(200)),
				Units:    api.UNITS_MB,
			},
		},
	}
}

func getDepResourcesUnlimited() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "1",
				Name: "default",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(1000)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(-1)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(-1)),
				Units:    api.UNITS_MB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "2",
				Name: "nodepool-a",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(1000)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(-1)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(-1)),
				Deserved: ptr.To(float64(-1)),
				Units:    api.UNITS_MB,
			},
		},
	}
}

func getProjUpdateRequest(resources []api.ResourcesNullable) api.ProjectUpdateRequest {
	return api.ProjectUpdateRequest{
		Resources: resources,
	}
}

func getProj1CreationRequest(clusterId string, depId string) api.ProjectCreationRequest {
	config := createDefaultApiJobConfig()
	return api.ProjectCreationRequest{
		Name:            "proj1",
		ClusterId:       clusterId,
		ParentId:        &depId,
		SchedulingRules: &config,
		Resources:       getProj1Resources(),
	}
}

func getProj3CreationRequest(clusterId string, depId string) api.ProjectCreationRequest {
	config := createDefaultApiJobConfig()
	return api.ProjectCreationRequest{
		Name:            "proj3",
		ClusterId:       clusterId,
		ParentId:        &depId,
		SchedulingRules: &config,
		Resources:       getProj1Resources(),
	}
}

func getProj2CreationRequest(clusterId string, depId string) api.ProjectCreationRequest {
	config := createDefaultApiJobConfig()
	return api.ProjectCreationRequest{
		Name:            "proj2",
		ClusterId:       clusterId,
		ParentId:        &depId,
		SchedulingRules: &config,
		Resources:       getProj2Resources(),
	}
}

func getProjToCreateCreationRequest(clusterId string, depId string) api.ProjectCreationRequest {
	config := createDefaultApiJobConfig()
	return api.ProjectCreationRequest{
		Name:            "proj-to-create",
		ClusterId:       clusterId,
		ParentId:        &depId,
		SchedulingRules: &config,
		Resources:       getProjToCreateResources(),
	}
}

func getProj1Resources() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "1",
				Name: "default",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(1)),
				Deserved: ptr.To(float64(1)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(1000)),
				Deserved: ptr.To(float64(1000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(100)),
				Deserved: ptr.To(float64(100)),
				Units:    api.UNITS_MB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "2",
				Name: "nodepool-a",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(1)),
				Deserved: ptr.To(float64(1)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(1000)),
				Deserved: ptr.To(float64(1000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(100)),
				Deserved: ptr.To(float64(100)),
				Units:    api.UNITS_MB,
			},
		},
	}
}

func getProj2Resources() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "1",
				Name: "default",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
				Units:    api.UNITS_MB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "2",
				Name: "nodepool-a",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(1)),
				Deserved: ptr.To(float64(1)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(1000)),
				Deserved: ptr.To(float64(1000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(100)),
				Deserved: ptr.To(float64(100)),
				Units:    api.UNITS_MB,
			},
		},
	}
}

func getProjToCreateResources() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "1",
				Name: "default",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(2)),
				Deserved: ptr.To(float64(2)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(2000)),
				Deserved: ptr.To(float64(2000)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(200)),
				Deserved: ptr.To(float64(200)),
				Units:    api.UNITS_MB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id:   "2",
				Name: "nodepool-a",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
			},
			Memory: &api.ResourcesMemory{
				Limit:    ptr.To(float64(0)),
				Deserved: ptr.To(float64(0)),
				Units:    api.UNITS_MB,
			},
		},
	}
}

func updateDepartment(ctx context.Context, departmentApiService api.DepartmentsAPIServicer, departmentId string, updateRequest api.DepartmentUpdateRequest) {
	updateResponse, err := departmentApiService.UpdateDepartment(ctx, departmentId, updateRequest)
	if err != nil {
		Fail(fmt.Sprintf("Failed updating department %s: %+v", departmentId, updateResponse.Body))
	}
	Expect(updateResponse.Code).To(Equal(http.StatusOK), "Failed updating department %s: %+v", departmentId, updateResponse.Body)
}

func updateDepartmentWithResources(ctx context.Context, departmentApiService api.DepartmentsAPIServicer, departmentId string, resources []api.ResourcesNullable) {
	updateRequest := api.DepartmentUpdateRequest{
		Resources: resources,
	}
	updateDepartment(ctx, departmentApiService, departmentId, updateRequest)
}

func updateProject(ctx context.Context, projectApiService api.ProjectsAPIServicer, projectId string, updateRequest api.ProjectUpdateRequest) {
	updateResponse, err := projectApiService.UpdateProject(ctx, projectId, updateRequest)
	if err != nil {
		Fail(fmt.Sprintf("Failed updating project %s", projectId))
	}
	Expect(updateResponse.Code).To(Equal(http.StatusOK), "Failed updating project %s: %+v", projectId, updateResponse.Body)
}

func updateProjectStatus(ctx context.Context, projectApiService api.ProjectsAPIServicer, projId, clusterId string,
	quotaStatuses []api.QuotaStatusNodePool,
) {
	updateStatusRequest := api.ProjectStatusUpdateRequest{
		Id:        projId,
		ClusterId: clusterId,
		Status: api.ProjectStatus{
			Phase:                 "Ready",
			NodePoolQuotaStatuses: quotaStatuses,
		},
	}
	updateResponse, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
	if err != nil {
		Fail(fmt.Sprintf("Failed updating project %s status: %+v", projId, updateResponse.Body))
	}
	Expect(updateResponse.Code).To(Equal(http.StatusOK), "message body: %+v", updateResponse.Body)
}

func updateProjectWithResources(ctx context.Context, projectApiService api.ProjectsAPIServicer, projectId string, resources []api.ResourcesNullable) {
	updateRequest := api.ProjectUpdateRequest{
		Resources: resources,
	}
	updateProject(ctx, projectApiService, projectId, updateRequest)
}

func copyProjUpdateReqeust(req *api.ProjectUpdateRequest) api.ProjectUpdateRequest {
	newReq := &api.ProjectUpdateRequest{}

	newReqBytes, _ := json.Marshal(req)
	json.Unmarshal(newReqBytes, newReq)
	return *newReq
}

func copyProjCreateRequest(req *api.ProjectCreationRequest) api.ProjectCreationRequest {
	newReq := &api.ProjectCreationRequest{}

	newReqBytes, _ := json.Marshal(req)
	json.Unmarshal(newReqBytes, newReq)
	return *newReq
}

func apiResourcesDeservedToModelDeserved(resources []api.ResourcesNullable) []models.Resources {
	apiResources := make([]models.Resources, len(resources))

	for i := range resources {
		apiResources[i] = apiResourceDeservedToModelDeserved(resources[i])
	}

	return apiResources
}

func apiResourceDeservedToModelDeserved(resource api.ResourcesNullable) models.Resources {
	result := models.Resources{
		NodePoolId:   resource.NodePool.Id,
		NodePoolName: resource.NodePool.Name,
		GpuDeserved:  resource.Gpu.Deserved,
	}

	if resource.Cpu != nil {
		result.CpuDeserved = resource.Cpu.Deserved
	}
	if resource.Memory != nil {
		result.MemoryDeserved = resource.Memory.Deserved
		result.MemoryUnits = apiUnitsToModel(resource.Memory.Units)
	}

	return result
}

func apiUnitsToModel(units api.Units) models.Units {
	switch units {
	case api.UNITS_MB:
		return models.MB
	case api.UNITS_GB:
		return models.GB
	case api.UNITS_MIB:
		return models.Mib
	}
	return models.Mib
}

// setLimitQuotaOverSubscription writes the migrated LimitQuotaOverSubscription setting for the test tenant
// via the settings package (the source of truth after migration to org-unit-service).
func setLimitQuotaOverSubscription(enabled bool) {
	err := setting_pkg.Update[*settings.LimitQuotaOverSubscription](TestTenant, settings.LimitQuotaOverSubscription{Enabled: enabled}).Execute()
	Expect(err).To(BeNil())
}
