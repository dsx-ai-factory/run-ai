/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/common"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("workloads - event tests", func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
		WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, testing.TestDefaultClusterVersion, nil)))

	Context("basic flow", func() {
		It("cluster deletion event", func() {
			By("Create workload1 in one cluster")
			workload1 := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId, Workloads: []api.WorkloadSyncFields{workload1}})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			By("Validate that workload1 created")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload1.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			By("Create workload2 in another cluster")
			workload2 := SetupBaseWorkload(testing.TestUuid2, AnotherProjectId, testing.ClusterUuid2)
			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: AnotherProjectId, Workloads: []api.WorkloadSyncFields{workload2}})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			By("Validate that workload2 created")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload2.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			By("triggering cluster deletion event for cluster1")
			delErr := common.HandleClusterDeleteRequest(ctx, &events.ClusterDelete{
				ClusterId: testing.ClusterUuid1,
			})
			Expect(delErr).To(BeNil())

			By("verifying workload1 from project 1 that's in cluster 1 was deleted")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload1.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))

			By("verifying workload2 from project 2 that's in cluster 2 still exist")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload2.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body.(api.WorkloadDetailed).DeletedAt).To(BeNil())

			By("cleanup")
			workloadsService.DeleteWorkload(ctx, workload2.Id)
		})

		It("tenant deletion event", func() {
			By("Create workload1 in cluster1 & tenant1")
			workload1 := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId, Workloads: []api.WorkloadSyncFields{workload1}})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			By("Validate that workload1 created")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload1.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			By("Create workload2 in another tenant")
			ctx = testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
			workload2 := SetupBaseWorkload(testing.TestUuid2, AnotherProjectId, testing.ClusterUuid2)
			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: AnotherProjectId, Workloads: []api.WorkloadSyncFields{workload2}})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			By("Validate that workload2 created")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload2.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			By("triggering tenant deletion event for tenant1")
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
			delErr := common.HandleTenantDeleteRequest(ctx, &events.TenantDelete{
				TenantId: testing.TestTenant,
			})
			Expect(delErr).To(BeNil())

			By("verifying workload1 from project 1 that's in tenant 1 was deleted")
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload1.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))

			By("verifying workload2 from project 2 that's in tenant 2 still exist")
			ctx = testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
			apiResponse, _ = workloadsService.GetWorkload(ctx, workload2.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body.(api.WorkloadDetailed).DeletedAt).To(BeNil())

			By("cleanup")
			workloadsService.DeleteWorkload(ctx, workload2.Id)
		})
	})
})
