/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"k8s.io/utils/ptr"
)

var _ = Describe("Workloads Preemptibility tests", Ordered, func() {
	var (
		ctx              context.Context
		workloadsService api.WorkloadsAPIServicer
	)

	BeforeAll(func() {
		ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
		_ = NewPodsService(mockTenantSettingsClient) // Setup pods service if needed by WorkloadsService indirectly
		workloadsService = NewWorkloadsService()

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
		testing.CreateTables()
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("SyncWorkloads with Preemptibility", func() {
		It("should save preemptibility when provided", func() {
			workloadId := uuid.NewString()
			workload := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
			workload.Preemptibility = ptr.To(string(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE))

			apiResponse, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Verify in DB: preemptibility stored as provided
			dbWorkload, dbErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
			Expect(dbErr).To(BeNil())
			Expect(dbWorkload.Preemptibility).ToNot(BeNil())
			Expect(*dbWorkload.Preemptibility).To(Equal(string(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)))
		})

		It("should not overwrite preemptibility when nil", func() {
			workloadId := uuid.NewString()
			// 1. Create with NON_PREEMPTIBLE
			workload := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
			workload.Preemptibility = ptr.To(string(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))

			apiResponse, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// 2. Update with nil
			workload.Preemptibility = nil
			apiResponse, err = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Verify remains as initially set
			dbWorkload, dbErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
			Expect(dbErr).To(BeNil())
			Expect(dbWorkload.Preemptibility).ToNot(BeNil())
			Expect(*dbWorkload.Preemptibility).To(Equal(string(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)))
		})

		It("should update preemptibility when changed", func() {
			workloadId := uuid.NewString()
			// 1. Create with PREEMPTIBLE
			workload := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
			workload.Preemptibility = ptr.To(string(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE))

			apiResponse, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// 2. Update to non-preemptible
			workload.Preemptibility = ptr.To(string(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			apiResponse, err = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Verify updated value stored
			dbWorkload, dbErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
			Expect(dbErr).To(BeNil())
			Expect(dbWorkload.Preemptibility).ToNot(BeNil())
			Expect(*dbWorkload.Preemptibility).To(Equal(string(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)))
		})
	})
})
