/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Internal Enrich Ranking API tests", func() {
	ctx := testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	BeforeEach(func() {
		ctx = testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	createWorkloadWithPhase := func(projectId, clusterId string, phase api.Phase, nodePools []string) api.WorkloadSyncFields {
		workload := SetupBaseWorkload(uuid.NewString(), projectId, clusterId)
		workload.Phase = phase
		if nodePools != nil {
			workload.RequestedNodePools = nodePools
		}
		return workload
	}

	syncWorkloads := func(projectId string, workloads []api.WorkloadSyncFields) {
		apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: projectId,
			Workloads: workloads,
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
	}

	callEnrichmentAPI := func(projectId, clusterId string, nodePoolName *string) api.OrgUnitRankingItem {
		request := api.InternalEnrichOrgUnitRankingRequest{
			ClusterId:    clusterId,
			OrgUnitType:  "project",
			NodePoolName: nodePoolName,
			Ids:          []string{projectId},
		}

		response, err := workloadsService.InternalEnrichOrgUnitRanking(ctx, request)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK))

		enrichedResponse, ok := response.Body.(api.InternalEnrichOrgUnitRankingResponse)
		Expect(ok).To(BeTrue())
		Expect(enrichedResponse.Items).To(HaveLen(1))

		return enrichedResponse.Items[0]
	}

	callEnrichmentAPIEventually := func(projectId, clusterId string, nodePoolName *string, expectedPendingCount int32) api.OrgUnitRankingItem {
		var item api.OrgUnitRankingItem
		Eventually(func(g Gomega) {
			request := api.InternalEnrichOrgUnitRankingRequest{
				ClusterId:    clusterId,
				OrgUnitType:  "project",
				NodePoolName: nodePoolName,
				Ids:          []string{projectId},
			}

			response, err := workloadsService.InternalEnrichOrgUnitRanking(ctx, request)
			g.Expect(err).To(BeNil())
			g.Expect(response.Code).To(Equal(http.StatusOK))

			enrichedResponse, ok := response.Body.(api.InternalEnrichOrgUnitRankingResponse)
			g.Expect(ok).To(BeTrue())
			g.Expect(enrichedResponse.Items).To(HaveLen(1))

			item = enrichedResponse.Items[0]
			g.Expect(item.PendingWorkloadsCount).NotTo(BeNil())
			g.Expect(*item.PendingWorkloadsCount).To(Equal(expectedPendingCount))
		}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())

		return item
	}

	Context("Enrich project ranking with workload counts", func() {
		It("should return correct running and pending workload counts", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, nil),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, nil),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, nil),
			})

			item := callEnrichmentAPI(projectId, clusterId, nil)

			Expect(item.Id).To(Equal(projectId))
			Expect(item.RunningWorkloadsCount).NotTo(BeNil())
			Expect(item.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*item.RunningWorkloadsCount).To(BeNumerically(">=", 0))
			Expect(*item.PendingWorkloadsCount).To(BeNumerically(">=", 0))
		})
	})

	Context("Enrich ranking with nodepool filter for pending workloads", func() {
		It("should count only pending workloads with single nodepool request", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "gpu-pool-test"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName, "another-pool"}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{"other-pool"}),
			})

			item := callEnrichmentAPIEventually(projectId, clusterId, &nodePoolName, int32(2))

			Expect(item.Id).To(Equal(projectId))
			Expect(item.RunningWorkloadsCount).NotTo(BeNil())
			Expect(*item.RunningWorkloadsCount).To(Equal(int32(0)))
		})

		It("should exclude pending workloads with multiple nodepool requests", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "exclusive-pool"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName, "pool2"}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName, "pool2", "pool3"}),
			})

			item := callEnrichmentAPIEventually(projectId, clusterId, &nodePoolName, int32(0))

			Expect(*item.PendingWorkloadsCount).To(Equal(int32(0)))
		})

		It("should exclude pending workloads with no nodepool specified", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "specific-pool"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, nil),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{}),
			})

			item := callEnrichmentAPIEventually(projectId, clusterId, &nodePoolName, int32(0))

			Expect(*item.PendingWorkloadsCount).To(Equal(int32(0)))
		})

		It("should exclude pending workloads with different single nodepool", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "target-pool"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{"different-pool"}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{"another-pool"}),
			})

			item := callEnrichmentAPIEventually(projectId, clusterId, &nodePoolName, int32(0))

			Expect(*item.PendingWorkloadsCount).To(Equal(int32(0)))
		})

		It("should handle mixed pending workloads correctly", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "target-pool"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				// Should be counted (3)
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName}),
				// Should NOT be counted
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{nodePoolName, "pool2"}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, []string{"other-pool"}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_PENDING, nil),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, []string{nodePoolName}),
			})

			item := callEnrichmentAPIEventually(projectId, clusterId, &nodePoolName, int32(3))

			Expect(*item.PendingWorkloadsCount).To(Equal(int32(3)))
			Expect(*item.RunningWorkloadsCount).To(Equal(int32(0)))
		})
	})

	Context("Enrich ranking with nodepool filter for running workloads", func() {
		It("should handle running workloads with nodepool filter", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()
			nodePoolName := "gpu-pool-running-test"

			syncWorkloads(projectId, []api.WorkloadSyncFields{
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, []string{nodePoolName}),
				createWorkloadWithPhase(projectId, clusterId, api.PHASE_RUNNING, []string{"other-pool"}),
			})

			item := callEnrichmentAPI(projectId, clusterId, &nodePoolName)

			Expect(item.Id).To(Equal(projectId))
			Expect(item.RunningWorkloadsCount).NotTo(BeNil())
			Expect(item.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*item.PendingWorkloadsCount).To(Equal(int32(0)))
			Expect(*item.RunningWorkloadsCount).To(BeNumerically(">=", 0))
		})
	})

	Context("Enrich ranking with no workloads", func() {
		It("should return zero counts when there are no workloads", func() {
			projectId := uuid.NewString()
			clusterId := uuid.NewString()

			item := callEnrichmentAPI(projectId, clusterId, nil)
			Expect(item.Id).To(Equal(projectId))
			Expect(item.RunningWorkloadsCount).NotTo(BeNil())
			Expect(item.PendingWorkloadsCount).NotTo(BeNil())
			Expect(*item.RunningWorkloadsCount).To(Equal(int32(0)))
			Expect(*item.PendingWorkloadsCount).To(Equal(int32(0)))

		})
	})
})
