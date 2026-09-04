/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"net/http"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/services"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("event test", Ordered, func() {
	projectId1 := "1"
	projectId2 := "2"
	projectId3 := "3"
	clusterId1 := uuid.New().String()
	clusterId2 := uuid.New().String()
	clusterId3 := uuid.New().String()
	departmentId1 := uuid.New().String()
	departmentId2 := uuid.New().String()
	departmentId3 := uuid.New().String()

	projectToAncestors := map[string][]string{
		projectId1: {departmentId1, clusterId1, TestTenant},
		projectId2: {departmentId2, clusterId2, TestTenant2},
		projectId3: {departmentId3, clusterId3, TestTenant},
	}

	var datavolumesApiService api.DatavolumesAPIServicer
	ctx := SetContext(TestTenant, TestUser)

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId1, buildDisplayedCluster(clusterId1, "2.18.0")).
			WithCluster(clusterId2, buildDisplayedCluster(clusterId2, "2.18.0")).
			WithCluster(clusterId3, buildDisplayedCluster(clusterId3, "2.18.0")))
		checker.InitClusterCapabilitiesChecker("", getDefaultClustersCacheConfig(), "")

		for projectId, ancestors := range projectToAncestors {
			SetProjectAncestorsCache(ctx, projectId, &authorization_utils.Ancestors{
				TenantId:     ancestors[2],
				ClusterUuid:  &ancestors[1],
				DepartmentId: &ancestors[0],
			})
			SetAncestorsCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, &authorization_utils.Ancestors{
				TenantId:    ancestors[2],
				ClusterUuid: &ancestors[1],
			})
			SetAncestorsCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, &authorization_utils.Ancestors{
				TenantId: ancestors[2],
			})
			SetScopeNameCache(ctx, projectId, api.SCOPETYPE_PROJECT, ancestors[1], getTestScopeName(projectId, api.SCOPETYPE_PROJECT))
			SetScopeNameCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, ancestors[1], getTestScopeName(ancestors[0], api.SCOPETYPE_DEPARTMENT))
			SetScopeNameCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, ancestors[1], getTestScopeName(ancestors[1], api.SCOPETYPE_CLUSTER))
		}
	})
	AfterAll(func() {
		for projectId, ancestors := range projectToAncestors {
			InvalidateProjectAncestorsCache(ctx, projectId)
			InvalidateAncestorsCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT)
			InvalidateAncestorsCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER)
			InvalidateScopeNameCache(ctx, projectId, api.SCOPETYPE_PROJECT, ancestors[1])
			InvalidateScopeNameCache(ctx, ancestors[0], api.SCOPETYPE_DEPARTMENT, ancestors[1])
			InvalidateScopeNameCache(ctx, ancestors[1], api.SCOPETYPE_CLUSTER, ancestors[1])
		}
	})

	BeforeEach(func() {
		datavolumesApiService = services.NewDatavolumesApiService()
	})

	Context("test project deletion", func() {
		It("basic flow", func() {
			By("creating project scoped data-volumes (project 1")
			projectDv1Response := createDatavolume(ctx, datavolumesApiService, projectId1, "project-dv-1").Body.(api.Datavolume)

			By("verifying data-volume from project 1 exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv1Response.Id)

			By("creating project scoped data-volumes (project 2")
			projectDv2Response := createDatavolume(ctx, datavolumesApiService, projectId3, "project-dv-2").Body.(api.Datavolume)

			By("verifying data-volume from project 2 exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv1Response.Id)

			By("triggering cluster deletion event for project 1")
			delErr := services.HandleProjectDeleteEvent(ctx, &events.ProjectDeleted{
				ProjectId: projectId1,
			})
			Expect(delErr).To(BeNil())

			By("verifying data-volume from project 1 was deleted")
			assertDatavolumeDeleted(ctx, datavolumesApiService, projectDv1Response.Id)

			By("verifying data-volume from project 2 still exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv2Response.Id)

			By("cleanup")
			datavolumesApiService.DeleteDatavolume(ctx, projectDv2Response.Id)
		})
	})
	Context("test cluster deletion", func() {
		It("basic flow", func() {
			By("creating project scoped data-volumes (project 1, cluster 1)")
			projectDv1Response := createDatavolume(ctx, datavolumesApiService, projectId1, "project-dv-1").Body.(api.Datavolume)

			By("verifying data-volume from project 1 exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv1Response.Id)

			By("creating project scoped data-volumes (project 2, cluster 2)")
			projectDv2Response := createDatavolume(ctx, datavolumesApiService, projectId3, "project-dv-2").Body.(api.Datavolume)

			By("verifying data-volume from project 2 exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv1Response.Id)

			By("triggering cluster deletion event for cluster 1")
			delErr := services.HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: clusterId1,
			})
			Expect(delErr).To(BeNil())

			By("verifying data-volume from project 1 that's in cluster 1 was deleted")
			assertDatavolumeDeleted(ctx, datavolumesApiService, projectDv1Response.Id)

			By("verifying data-volume from project 2 that's in cluster 2 still exist")
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv2Response.Id)

			By("cleanup")
			datavolumesApiService.DeleteDatavolume(ctx, projectDv2Response.Id)
		})
	})

	Context("test tenant deletion", func() {
		It("basic flow", func() {
			By("creating project scoped data-volumes for tenant 1")
			projectDv1Response := createDatavolume(ctx, datavolumesApiService, "1", "project-dv-1").Body.(api.Datavolume)

			By("creating project scoped data-volumes for tenant 2")
			ctx := SetContext(TestTenant2, TestUser)
			projectDv2Response := createDatavolume(ctx, datavolumesApiService, projectId2, "project-dv-2").Body.(api.Datavolume)
			ctx = SetContext(TestTenant, TestUser)

			By("triggering tenant deletion event for tenant 1")
			delErr := services.HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: TestTenant,
			})
			Expect(delErr).To(BeNil())

			By("verifying data-volume from project 1 that's in tenant 1 was deleted")
			assertDatavolumeDeleted(ctx, datavolumesApiService, projectDv1Response.Id)

			By("verifying data-volume from project 2 that's in tenant 2 still exist")
			ctx = SetContext(TestTenant2, TestUser)
			assertDatavolumeExists(ctx, datavolumesApiService, projectDv2Response.Id)

			By("cleanup")
			datavolumesApiService.DeleteDatavolume(ctx, projectDv2Response.Id)
		})
	})
})

func createDatavolume(ctx context.Context, apiService api.DatavolumesAPIServicer, projectId string, name string) api.ImplResponse {
	request := api.DatavolumeCreationFields{
		Name:          name,
		Description:   "test datavolume",
		ProjectId:     projectId,
		OriginPvcName: name,
	}

	res, err := apiService.CreateDatavolume(ctx, request)
	Expect(err).To(BeNil(), fmt.Sprintf("Create datavolume %s should succeed", name))
	Expect(res.Code).To(Equal(http.StatusCreated), fmt.Sprintf("Create datavolume %s should return 201", name))
	return res
}

func assertDatavolumeDeleted(ctx context.Context, apiService api.DatavolumesAPIServicer, id string) {
	res, _ := apiService.GetDatavolume(ctx, id)
	Expect(res.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("Datavolume %s should be deleted", id))
}

func assertDatavolumeExists(ctx context.Context, apiService api.DatavolumesAPIServicer, id string) {
	res, _ := apiService.GetDatavolume(ctx, id)
	Expect(res.Code).To(Equal(http.StatusOK), fmt.Sprintf("Datavolume %s should exist", id))
}
