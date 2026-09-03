/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

var _ = Describe("Workloads API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	podsService := NewPodsService(mockTenantSettingsClient)
	createdWorkloadsIds := []string{}
	createdPodsIds := []string{}

	BeforeAll(func() {
		id1 := uuid.NewString()
		id2 := uuid.NewString()
		id3 := uuid.NewString()
		id4 := uuid.NewString()
		id5 := uuid.NewString()
		id6 := uuid.NewString()
		id7 := uuid.NewString()
		id8 := uuid.NewString()
		// Project with deleted workloads
		workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: "10",
			Workloads: []api.WorkloadSyncFields{
				setupCustomWorkload(uuid.NewString(), "10", "DeletingProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
				setupCustomWorkload(uuid.NewString(), "10", "DeletingProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
				setupCustomWorkload(uuid.NewString(), "10", "DeletingProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
			},
		})
		workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: "10",
			Workloads: []api.WorkloadSyncFields{},
		})

		// Project 2 workloads
		apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: AnotherProjectId,
			Workloads: []api.WorkloadSyncFields{
				setupCustomWorkload(id1, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "1", "dep-1", DefaultType),
				setupCustomWorkload(id2, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "1", "dep-1", InferenceType),
				setupCustomWorkload(id3, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "1", "dep-1", WorkspaceType),
				setupCustomWorkload(id4, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "2", "dep-2", InferenceType),
				setupCustomWorkload(id5, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "1", "dep-1", DefaultType),
				setupCustomWorkload(id6, AnotherProjectId, "AnotherProjectId", testing.ClusterUuid2, "2", "dep-2", DefaultType),
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			Workloads: []api.WorkloadSyncFields{
				setupCustomWorkload(testing.TestUuid1, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
				setupCustomWorkload(testing.TestUuid2, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
				setupCustomWorkload(testing.TestUuid3, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
				setupCustomWorkload(id7, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "2", "dep-2", DefaultType),
				setupCustomWorkload(id8, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		createdWorkloadsIds = append(createdWorkloadsIds, id1, id2, id3, id4, id5, id6, id7, id8, testing.TestUuid1, testing.TestUuid2, testing.TestUuid3)
		podId1 := uuid.NewString()
		podId2 := uuid.NewString()
		podId3 := uuid.NewString()
		podId4 := uuid.NewString()
		podId5 := uuid.NewString()
		podId6 := uuid.NewString()
		podId7 := uuid.NewString()
		allocatedGpu1 := 2.0
		allocatedGpu2 := 1.0
		nodepool1 := DefaultNodepool
		nodepool2 := OtherNodepool
		apiPodResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
			ProjectId: AnotherProjectId,
			Pods: []api.PodSyncFields{
				setupCustomPod(podId1, id1, AnotherProjectId, testing.ClusterUuid2, &nodepool1, &allocatedGpu1, &allocatedGpu1),
				setupCustomPod(podId2, id1, AnotherProjectId, testing.ClusterUuid2, &nodepool2, &allocatedGpu2, &allocatedGpu1),
				setupCustomPod(podId3, id2, AnotherProjectId, testing.ClusterUuid2, &nodepool1, &allocatedGpu1, &allocatedGpu1),
				setupCustomPod(podId4, id3, AnotherProjectId, testing.ClusterUuid2, &nodepool1, &allocatedGpu1, &allocatedGpu1),
				setupCustomPod(podId5, id3, AnotherProjectId, testing.ClusterUuid2, &nodepool2, &allocatedGpu1, &allocatedGpu1),
				setupCustomPod(podId6, id4, AnotherProjectId, testing.ClusterUuid2, &nodepool2, &allocatedGpu1, &allocatedGpu1),
				setupCustomPod(podId7, id5, AnotherProjectId, testing.ClusterUuid2, &nodepool2, &allocatedGpu1, &allocatedGpu1),
			},
		})
		Expect(apiPodResponse.Code).To(Equal(http.StatusNoContent))
		createdPodsIds = append(createdPodsIds, podId1, podId2, podId3, podId4, podId5, podId6, podId7)
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("gpu allocation", func() {
		It("department filter with one group", func() {
			var groups []string
			groups = append(groups, CurrentNodepools)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, "", "", "1", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("nodepool filter", func() {
			var groups []string
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, testing.ClusterUuid2, "lala", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("0"))
		})
		It("cluster filter with 2 groups", func() {
			var groups []string
			groups = append(groups, DepartmentId, Type)
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				validateClusterFilterAllocation(g, telemetryValues)
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})
		It("group by category", func() {
			var groups []string
			groups = append(groups, Category)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(3))
			Expect(telemetryValues[0].Groups[0].Value).To(Equal(testing.Categories[WorkspaceType]))
			Expect(telemetryValues[1].Groups[0].Value).To(Equal(testing.Categories[InferenceType]))
			Expect(telemetryValues[2].Groups[0].Value).To(Equal(testing.Categories[DefaultType]))
		})
	})
	Context("workload count", func() {
		It("department filter", func() {
			var groups []string
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", "", "1", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("8"))
			Expect(telemetryValues[0].Groups).To(BeNil())
		})
		It("project filter", func() {
			var groups []string
			// DefaultProjectId has 5 workloads on cluster1
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", "", "", DefaultProjectId, groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("5"))
			Expect(telemetryValues[0].Groups).To(BeNil())

			// AnotherProjectId has 6 workloads on cluster2
			apiResponse, _ = workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", "", "", AnotherProjectId, groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry = httpResponseBodyToTelemetryResponse(apiResponse)
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("6"))
		})
		It("nodepool filter with 2 groups", func() {
			var groups []string
			groups = append(groups, Type, DepartmentId)
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(5))
				Expect(telemetryValues[0].Value).To(Equal("1"))
				Expect(telemetryValues[0].Groups).NotTo(BeNil())
				Expect(len(telemetryValues[0].Groups)).To(Equal(2))
				Expect(telemetryValues[0].Groups[0].Key).To(Equal(Type))
				Expect(telemetryValues[0].Groups[0].Value).To(Equal(InferenceType))
				Expect(telemetryValues[0].Groups[0].Name).To(Equal(""))
				Expect(telemetryValues[0].Groups[1].Key).To(Equal(DepartmentId))
				Expect(telemetryValues[0].Groups[1].Value).To(Equal("1"))
				Expect(telemetryValues[0].Groups[1].Name).To(Equal("dep-1"))
				Expect(len(telemetryValues[1].Groups)).To(Equal(2))
				Expect(telemetryValues[1].Groups[0].Key).To(Equal(Type))
				Expect(telemetryValues[1].Groups[0].Value).To(Equal(InferenceType))
				Expect(telemetryValues[1].Groups[0].Name).To(Equal(""))
				Expect(telemetryValues[1].Groups[1].Key).To(Equal(DepartmentId))
				Expect(telemetryValues[1].Groups[1].Value).To(Equal("2"))
				Expect(telemetryValues[1].Groups[1].Name).To(Equal("dep-2"))
				Expect(telemetryValues[1].Value).To(Equal("1"))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})
		It("cluster filter with one group", func() {
			var groups []string
			groups = append(groups, DepartmentId)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(2))
			Expect(telemetryValues[0].Value).To(Equal("4"))
			Expect(telemetryValues[0].Groups).NotTo(BeNil())
			Expect(len(telemetryValues[0].Groups)).To(Equal(1))
			Expect(telemetryValues[0].Groups[0].Key).To(Equal(DepartmentId))
			Expect(telemetryValues[0].Groups[0].Name).To(Equal("dep-1"))
			Expect(telemetryValues[0].Groups[0].Value).To(Equal("1"))
			Expect(telemetryValues[1].Value).To(Equal("2"))
			Expect(len(telemetryValues[1].Groups)).To(Equal(1))
			Expect(telemetryValues[1].Groups[0].Key).To(Equal(DepartmentId))
			Expect(telemetryValues[1].Groups[0].Value).To(Equal("2"))
			Expect(telemetryValues[1].Groups[0].Name).To(Equal("dep-2"))
		})
		It("cluster filter with bad group", func() {
			var groups []string
			groups = append(groups, "Name")
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("cluster filter with projectId group", func() {
			var groups []string
			groups = append(groups, ProjectId)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("6"))
			Expect(telemetryValues[0].Groups).NotTo(BeNil())
			Expect(len(telemetryValues[0].Groups)).To(Equal(1))
			Expect(telemetryValues[0].Groups[0].Key).To(Equal(ProjectId))
			Expect(telemetryValues[0].Groups[0].Value).To(Equal(AnotherProjectId))
			Expect(telemetryValues[0].Groups[0].Name).To(Equal("AnotherProjectId"))
		})
		It("cluster filter with phase group", func() {
			var groups []string
			groups = append(groups, Phase)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("6"))
			Expect(telemetryValues[0].Groups).NotTo(BeNil())
			Expect(len(telemetryValues[0].Groups)).To(Equal(1))
			Expect(telemetryValues[0].Groups[0].Key).To(Equal(Phase))
			Expect(telemetryValues[0].Groups[0].Name).To(Equal(""))
			Expect(telemetryValues[0].Groups[0].Value).To(Equal("Running"))
		})
		It("nodepool filter with clusterId group", func() {
			var groups []string
			groups = append(groups, ClusterId)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", DefaultNodepool, "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("department bad filter", func() {
			var groups []string
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", "", "21", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("0"))
		})
	})
	Context("idle allocated gpus", func() {
		idleGpuId1 := ""
		idleGpuId2 := ""
		idleGpuId3 := ""

		BeforeAll(func() {
			idleGpuId1 = uuid.NewString()
			idleGpuId2 = uuid.NewString()
			idleGpuId3 = uuid.NewString()

			// Create workloads for idle allocated gpus telemetry tests
			// Include original DefaultProjectId workloads to avoid replacing them (SyncWorkloads replaces all workloads for a project)
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					// Original workloads from top-level BeforeAll
					setupCustomWorkload(testing.TestUuid1, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
					setupCustomWorkload(testing.TestUuid2, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
					setupCustomWorkload(testing.TestUuid3, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
					setupCustomWorkload(createdWorkloadsIds[6], DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "2", "dep-2", DefaultType), // id7
					setupCustomWorkload(createdWorkloadsIds[7], DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType), // id8
					// New workloads for idle allocated gpus tests
					setupCustomWorkload(idleGpuId1, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", DefaultType),
					setupCustomWorkload(idleGpuId2, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "1", "dep-1", InferenceType),
					setupCustomWorkload(idleGpuId3, DefaultProjectId, "DefaultProjectId", testing.ClusterUuid1, "2", "dep-2", DefaultType),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Set idle_allocated_gpus values for the workloads
			tx := dao_utils.DB().Table(dao.WorkloadsTableName).Where("id = ?", idleGpuId1).
				Updates(map[string]interface{}{
					"idle_allocated_gpus": 2.5,
				})
			Expect(tx.Error).To(BeNil())
			tx = dao_utils.DB().Table(dao.WorkloadsTableName).Where("id = ?", idleGpuId2).
				Updates(map[string]interface{}{
					"idle_allocated_gpus": 1.5,
				})
			Expect(tx.Error).To(BeNil())
			tx = dao_utils.DB().Table(dao.WorkloadsTableName).Where("id = ?", idleGpuId3).
				Updates(map[string]interface{}{
					"idle_allocated_gpus": 3.0,
				})
			Expect(tx.Error).To(BeNil())
			createdWorkloadsIds = append(createdWorkloadsIds, idleGpuId1, idleGpuId2, idleGpuId3)
		})

		It("cluster filter without groups", func() {
			var groups []string
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, []api.Phase{})
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				g.Expect(len(telemetryValues)).To(Equal(1))
				g.Expect(telemetryValues[0].Groups).To(BeNil())
				g.Expect(telemetryValues[0].Value).To(Equal("7.00")) // 2.5 + 1.5 + 3.0 = 7.0
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("cluster filter with department group", func() {
			var groups []string
			groups = append(groups, DepartmentId)
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, []api.Phase{})
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				g.Expect(len(telemetryValues)).To(Equal(2))
				// Department 1: 2.5 + 1.5 = 4.0
				g.Expect(telemetryValues[0].Groups[0].Key).To(Equal(DepartmentId))
				g.Expect(telemetryValues[0].Groups[0].Value).To(Equal("1"))
				g.Expect(telemetryValues[0].Groups[0].Name).To(Equal("dep-1"))
				g.Expect(telemetryValues[0].Value).To(Equal("4.00"))
				// Department 2: 3.0
				g.Expect(telemetryValues[1].Groups[0].Key).To(Equal(DepartmentId))
				g.Expect(telemetryValues[1].Groups[0].Value).To(Equal("2"))
				g.Expect(telemetryValues[1].Groups[0].Name).To(Equal("dep-2"))
				g.Expect(telemetryValues[1].Value).To(Equal("3.00"))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("cluster filter with type group", func() {
			var groups []string
			groups = append(groups, Type)
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, []api.Phase{})
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				g.Expect(len(telemetryValues)).To(Equal(2))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("department filter", func() {
			var groups []string
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, "", "", "1", "", groups, []api.Phase{})
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				g.Expect(len(telemetryValues)).To(Equal(1))
				g.Expect(telemetryValues[0].Groups).To(BeNil())
				g.Expect(telemetryValues[0].Value).To(Equal("4.00")) // 2.5 + 1.5 = 4.0
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("nodepool filter with zero result", func() {
			var groups []string
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "non-existent-nodepool", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("0"))
		})

		It("phase filter - Running", func() {
			var groups []string
			phases := []api.Phase{api.PHASE_RUNNING}
			Eventually(func(g Gomega) {
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, phases)
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
				g.Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				g.Expect(len(telemetryValues)).To(Equal(1))
				g.Expect(telemetryValues[0].Value).To(Equal("7.00"))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("phase filter - Failed returns zero", func() {
			var groups []string
			phases := []api.Phase{api.PHASE_FAILED}
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, phases)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
		})

		It("invalid phase returns error", func() {
			var groups []string
			phases := []api.Phase{api.Phase("InvalidPhase")}
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, phases)
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("bad group returns error", func() {
			var groups []string
			groups = append(groups, "InvalidGroup")
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_IDLE_ALLOCATED_GPUS, testing.ClusterUuid1, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("workload count csv", func() {
		BeforeAll(func() {
			ctx = testing.SetContextWithAccept(testing.TestTenant, testing.TestUser, []string{}, CsvAccept)
		})
		AfterAll(func() {
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		})
		It("nodepool filter with 2 groups", func() {
			var groups []string
			groups = append(groups, Type, DepartmentId)
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			csvFile := apiResponse.Body.(runaihttp.CsvFile)
			telemetry := csvFile.Data
			Expect(len(telemetry)).To(Equal(6))
			Expect(len(telemetry[0])).To(Equal(5))
			Expect(telemetry[0][0]).To(Equal("value"))
			Expect(telemetry[0][1]).To(Equal(fmt.Sprintf("%s.value", Type)))
			Expect(telemetry[0][2]).To(Equal(fmt.Sprintf("%s.name", Type)))
			Expect(telemetry[0][3]).To(Equal(fmt.Sprintf("%s.value", DepartmentId)))
			Expect(telemetry[0][4]).To(Equal(fmt.Sprintf("%s.name", DepartmentId)))
			Expect(telemetry[1][0]).To(Equal("1"))
			Expect(telemetry[1][1]).To(Equal(InferenceType))
			Expect(telemetry[1][2]).To(Equal(""))
			Expect(telemetry[1][3]).To(Equal("1"))
			Expect(telemetry[1][4]).To(Equal("dep-1"))
			Expect(telemetry[2][0]).To(Equal("1"))
			Expect(telemetry[2][1]).To(Equal(InferenceType))
			Expect(telemetry[2][2]).To(Equal(""))
			Expect(telemetry[2][3]).To(Equal("2"))
			Expect(telemetry[2][4]).To(Equal("dep-2"))
		})
		It("department filter", func() {
			var groups []string
			apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, "", "", "1", "", groups, []api.Phase{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			csvFile := apiResponse.Body.(runaihttp.CsvFile)
			telemetry := csvFile.Data
			Expect(len(telemetry)).To(Equal(2))
			Expect(len(telemetry[0])).To(Equal(1))
			Expect(telemetry[0][0]).To(Equal("value"))
			Expect(telemetry[1][0]).To(Equal("10")) // 8 original + 2 from idle allocated gpus tests
		})
	})

	Context("phase filtering", func() {
		Context("workload count with phase filtering", func() {
			It("should filter by single phase - Running", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_RUNNING}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("6")) // All workloads in cluster2 are Running
			})

			It("should filter by multiple phases - Running and Failed", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_RUNNING, api.PHASE_FAILED}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("6")) // All workloads in cluster2 are Running
			})

			It("should return empty result for non-existent phase", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_FAILED}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("0")) // No Failed workloads in cluster2
			})

			It("should work with groupBy and phase filtering", func() {
				var groups []string
				groups = append(groups, Phase)
				phases := []api.Phase{api.PHASE_RUNNING}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("6"))
				Expect(telemetryValues[0].Groups).NotTo(BeNil())
				Expect(len(telemetryValues[0].Groups)).To(Equal(1))
				Expect(telemetryValues[0].Groups[0].Key).To(Equal(Phase))
				Expect(telemetryValues[0].Groups[0].Value).To(Equal("Running"))
			})

			It("should return all phases when no phase filter is provided", func() {
				var groups []string
				phases := []api.Phase{} // Empty phases array
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("6")) // All workloads in cluster2
			})
		})

		Context("gpu allocation with phase filtering", func() {
			It("should filter by single phase - Running", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_RUNNING}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				// Should have some GPU allocation value
				Expect(telemetryValues[0].Value).NotTo(Equal("0"))
			})

			It("should return zero allocation for non-existent phase", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_FAILED}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_GPU_ALLOCATION)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("0")) // No GPU allocation for Failed workloads
			})
		})

		Context("phase validation", func() {
			It("should return error for invalid phase", func() {
				var groups []string
				phases := []api.Phase{api.Phase("InvalidPhase")}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
			})

			It("should accept valid phases", func() {
				var groups []string
				phases := []api.Phase{api.PHASE_RUNNING, api.PHASE_PENDING, api.PHASE_FAILED}
				apiResponse, _ := workloadsService.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_WORKLOADS_COUNT, testing.ClusterUuid2, "", "", "", groups, phases)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
			})
		})

		Context("pending time distribution", func() {

			It("should group by cluster and bucket", func() {
				workloadIds := createPendingWorkloads(ctx, workloadsService, 3, func(i int) (clusterId, deptId, deptName string) {
					if i < 2 {
						return testing.ClusterUuid1, DefaultDepartmentId, DefaultDepartmentName
					}
					return testing.ClusterUuid2, DefaultDepartmentId, DefaultDepartmentName
				})
				createdWorkloadsIds = append(createdWorkloadsIds, workloadIds...)

				Eventually(func(g Gomega) {
					telemetryValues := getPendingTimeDistributionWithFilters(ctx, workloadsService, g, "", "", "", []string{ClusterId})
					g.Expect(len(telemetryValues)).To(Equal(8)) // 4 buckets * 2 clusters

					validatePendingTimeGroups(g, telemetryValues, ClusterId, []string{testing.ClusterUuid1, testing.ClusterUuid2})
					validatePendingDistributionGroupCounts(g, telemetryValues, map[string]int{
						testing.ClusterUuid1: 2,
						testing.ClusterUuid2: 1,
					}, 3)
				}, 2*time.Second, 100*time.Millisecond).Should(Succeed())
			})

			It("should group by department and bucket", func() {
				workloadIds := createPendingWorkloads(ctx, workloadsService, 4, func(i int) (clusterId, deptId, deptName string) {
					if i < 3 {
						return testing.ClusterUuid1, DefaultDepartmentId, DefaultDepartmentName
					}
					return testing.ClusterUuid1, "2", "dep-2"
				})
				createdWorkloadsIds = append(createdWorkloadsIds, workloadIds...)

				Eventually(func(g Gomega) {
					telemetryValues := getPendingTimeDistributionWithFilters(ctx, workloadsService, g, testing.ClusterUuid1, "", "", []string{DepartmentId})
					g.Expect(len(telemetryValues)).To(Equal(8)) // 4 buckets * 2 departments

					validatePendingTimeGroupsWithNames(g, telemetryValues, DepartmentId,
						map[string]string{DefaultDepartmentId: DefaultDepartmentName, "2": "dep-2"})
					validatePendingDistributionGroupCounts(g, telemetryValues, map[string]int{
						DefaultDepartmentId: 3,
						"2":                 1,
					}, 4)
				}, 2*time.Second, 100*time.Millisecond).Should(Succeed())
			})

			It("should filter by nodepool and group by project", func() {
				workloadIds := []string{}
				workloads1 := []api.WorkloadSyncFields{}
				workloads2 := []api.WorkloadSyncFields{}

				// Project 1 - 2 workloads requesting only "default" nodepool
				for i := 0; i < 2; i++ {
					id := uuid.NewString()
					workloadIds = append(workloadIds, id)
					workload := setupCustomWorkload(id, DefaultProjectId, DefaultProjectName, testing.ClusterUuid1, DefaultDepartmentId, DefaultDepartmentName, DefaultType)
					workload.Phase = "Pending"
					workload.RequestedNodePools = []string{DefaultNodepool}
					workloads1 = append(workloads1, workload)
				}

				// Project 1 - 1 workload requesting multiple nodepools (should be excluded by filter)
				id := uuid.NewString()
				workloadIds = append(workloadIds, id)
				workload := setupCustomWorkload(id, DefaultProjectId, DefaultProjectName, testing.ClusterUuid1, DefaultDepartmentId, DefaultDepartmentName, DefaultType)
				workload.Phase = "Pending"
				workload.RequestedNodePools = []string{DefaultNodepool, OtherNodepool}
				workloads1 = append(workloads1, workload)

				// Sync project 1 workloads
				workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
					ProjectId: DefaultProjectId,
					Workloads: workloads1,
				})

				// Project 2 - 1 workload requesting only "default" nodepool
				id = uuid.NewString()
				workloadIds = append(workloadIds, id)
				workload = setupCustomWorkload(id, AnotherProjectId, "test-2", testing.ClusterUuid1, DefaultDepartmentId, DefaultDepartmentName, DefaultType)
				workload.Phase = "Pending"
				workload.RequestedNodePools = []string{DefaultNodepool} // Only "default"
				workloads2 = append(workloads2, workload)

				// Sync project 2 workloads
				workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
					ProjectId: AnotherProjectId,
					Workloads: workloads2,
				})

				createdWorkloadsIds = append(createdWorkloadsIds, workloadIds...)

				Eventually(func(g Gomega) {
					telemetryValues := getPendingTimeDistributionWithFilters(ctx, workloadsService, g, testing.ClusterUuid1, DefaultNodepool, "", []string{ProjectId})
					g.Expect(len(telemetryValues)).To(Equal(8)) // 4 buckets * 2 projects

					validatePendingTimeGroupsWithNames(g, telemetryValues, ProjectId,
						map[string]string{DefaultProjectId: DefaultProjectName, AnotherProjectId: "test-2"})

					// Should only count the 3 workloads that request ONLY "default" nodepool
					validatePendingDistributionGroupCounts(g, telemetryValues, map[string]int{
						DefaultProjectId: 2, // 2 workloads with only "default"
						AnotherProjectId: 1, // 1 workload with only "default"
					}, 3)
				}, 2*time.Second, 100*time.Millisecond).Should(Succeed())
			})

		})
	})
})

func validateClusterFilterAllocation(g Gomega, telemetryValues []api.TelemetryResponseValuesInner) {
	g.Expect(len(telemetryValues)).To(Equal(5))
	g.Expect(telemetryValues[0].Value).To(Equal("2.00")) // 0
	g.Expect(telemetryValues[0].Groups).NotTo(BeNil())
	g.Expect(len(telemetryValues[0].Groups)).To(Equal(2))
	g.Expect(telemetryValues[0].Groups[0].Key).To(Equal(DepartmentId))
	g.Expect(telemetryValues[0].Groups[0].Name).To(Equal("dep-1"))
	g.Expect(telemetryValues[0].Groups[0].Value).To(Equal("1"))
	g.Expect(telemetryValues[0].Groups[1].Key).To(Equal(Type))
	g.Expect(telemetryValues[0].Groups[1].Value).To(Equal(InferenceType))
	g.Expect(telemetryValues[1].Value).To(Equal("5.00")) // 1
	g.Expect(telemetryValues[1].Groups).NotTo(BeNil())
	g.Expect(len(telemetryValues[1].Groups)).To(Equal(2))
	g.Expect(telemetryValues[1].Groups[0].Key).To(Equal(DepartmentId))
	g.Expect(telemetryValues[1].Groups[0].Value).To(Equal("1"))
	g.Expect(telemetryValues[1].Groups[0].Name).To(Equal("dep-1"))
	g.Expect(telemetryValues[1].Groups[1].Key).To(Equal(Type))
	g.Expect(telemetryValues[1].Groups[1].Value).To(Equal(DefaultType))
	g.Expect(telemetryValues[2].Value).To(Equal("4.00")) // 2
	g.Expect(telemetryValues[2].Groups).NotTo(BeNil())
	g.Expect(len(telemetryValues[2].Groups)).To(Equal(2))
	g.Expect(telemetryValues[2].Groups[0].Key).To(Equal(DepartmentId))
	g.Expect(telemetryValues[2].Groups[0].Name).To(Equal("dep-1"))
	g.Expect(telemetryValues[2].Groups[0].Value).To(Equal("1"))
	g.Expect(telemetryValues[2].Groups[1].Key).To(Equal(Type))
	g.Expect(telemetryValues[2].Groups[1].Value).To(Equal(WorkspaceType))
	g.Expect(telemetryValues[3].Value).To(Equal("2.00")) // 3
	g.Expect(telemetryValues[3].Groups).NotTo(BeNil())
	g.Expect(len(telemetryValues[3].Groups)).To(Equal(2))
	g.Expect(telemetryValues[3].Groups[0].Key).To(Equal(DepartmentId))
	g.Expect(telemetryValues[3].Groups[0].Value).To(Equal("2"))
	g.Expect(telemetryValues[3].Groups[0].Name).To(Equal("dep-2"))
	g.Expect(telemetryValues[3].Groups[1].Key).To(Equal(Type))
	g.Expect(telemetryValues[3].Groups[1].Value).To(Equal(InferenceType))
	g.Expect(telemetryValues[4].Value).To(Equal("0")) // 4
	g.Expect(telemetryValues[4].Groups).NotTo(BeNil())
	g.Expect(len(telemetryValues[4].Groups)).To(Equal(2))
	g.Expect(telemetryValues[4].Groups[0].Key).To(Equal(DepartmentId))
	g.Expect(telemetryValues[4].Groups[0].Name).To(Equal("dep-2"))
	g.Expect(telemetryValues[4].Groups[0].Value).To(Equal("2"))
	g.Expect(telemetryValues[4].Groups[1].Key).To(Equal(Type))
	g.Expect(telemetryValues[4].Groups[1].Value).To(Equal(DefaultType))
}

func setupCustomPod(id string, workloadId string, projectId string, clusterId string, nodepool *string, allocatedGpu *float64, allocatedCpu *float64) api.PodSyncFields {
	return api.PodSyncFields{
		Name:               "",
		PriorityClassName:  "",
		Id:                 id,
		WorkloadId:         workloadId,
		ClusterId:          clusterId,
		ProjectId:          projectId,
		NodeName:           nil,
		CreatedAt:          time.Time{},
		CompletedAt:        nil,
		Containers:         nil,
		CurrentNodePool:    nodepool,
		RequestedNodePools: nil,
		RequestedResources: nil,
		AllocatedResources: &api.AllocatedResources{
			Gpu:               allocatedGpu,
			GpuMemory:         nil,
			Cpu:               allocatedCpu,
			CpuMemory:         nil,
			ExtendedResources: nil,
		},
		Tolerations: nil,
		K8sPhase:    "",
	}
}

func setupCustomWorkload(id string, projectId string, projectName string, clusterId string, departmentId string, departmentName string, workloadType string) api.WorkloadSyncFields {
	return api.WorkloadSyncFields{
		Type:           workloadType,
		Name:           DefaultWorkloadName,
		Source:         DefaultSource,
		Id:             id,
		SubmittedBy:    DefaultSubmittedBy,
		ClusterId:      clusterId,
		ProjectName:    projectName,
		ProjectId:      projectId,
		DepartmentId:   departmentId,
		DepartmentName: departmentName,
		Namespace:      DefaultNamespace,
		CreatedAt:      DefaultCreationTime,
		RequestedResources: &api.WorkloadRequestResources{
			GpuRequestType: (*api.GpuRequestType)(pointer.String(string(DefaultGPURequestType))),
			Gpu: &api.RequestResourceCores{
				Limit:   pointer.Float64(DefaultGPULimit),
				Request: pointer.Float64(DefaultGPURequest),
			},
		},
		Phase:              DefaultPhase,
		PhaseMessage:       "",
		K8sPhase:           DefaultK8sPhase,
		RequestedPods:      api.RequestedPods{Number: pointer.Int32(DefaultNumberOfRequestedPods)},
		RequestedNodePools: DefaultRequestedNodePools,
		Images:             DefaultImages,
		ChildrenIds:        DefaultChildrenIds,
		Urls:               DefaultUrls,
		AdditionalFields:   "",
		Category:           testing.Categories[workloadType],
	}
}

func httpResponseBodyToTelemetryResponse(httpResponse api.ImplResponse) api.TelemetryResponse {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.TelemetryResponse{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func createPendingWorkloads(ctx context.Context, service api.WorkloadsAPIServicer, count int, factory func(i int) (clusterId, deptId, deptName string)) []string {
	workloadIds := make([]string, count)
	workloads := make([]api.WorkloadSyncFields, count)

	for i := 0; i < count; i++ {
		workloadIds[i] = uuid.NewString()
		clusterId, deptId, deptName := factory(i)
		workload := setupCustomWorkload(workloadIds[i], DefaultProjectId, DefaultProjectName, clusterId, deptId, deptName, DefaultType)
		workload.Phase = "Pending"
		workloads[i] = workload
	}

	service.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
		ProjectId: DefaultProjectId,
		Workloads: workloads,
	})

	return workloadIds
}

func getPendingTimeDistributionWithFilters(ctx context.Context, service api.WorkloadsAPIServicer, g Gomega, clusterId, nodepoolName, departmentId string, groups []string) []api.TelemetryResponseValuesInner {
	apiResponse, _ := service.GetWorkloadsTelemetry(ctx, api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION, clusterId, nodepoolName, departmentId, "", groups, []api.Phase{})
	g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
	telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
	g.Expect(telemetry.Type).To(Equal(string(api.WORKLOADTELEMETRYTYPE_PENDING_TIME_DISTRIBUTION)))
	g.Expect(telemetry.Timestamp).NotTo(BeNil())
	return telemetry.Values
}

func validatePendingTimeGroups(g Gomega, values []api.TelemetryResponseValuesInner, groupKey string, expectedValues []string) {
	for _, value := range values {
		g.Expect(len(value.Groups)).To(Equal(2))
		// First group is always bucket
		g.Expect(value.Groups[0].Key).To(Equal(Bucket))
		g.Expect(value.Groups[0].Value).NotTo(BeEmpty())
		g.Expect(value.Groups[0].Name).NotTo(BeEmpty())
		// Second group
		g.Expect(value.Groups[1].Key).To(Equal(groupKey))
		g.Expect(value.Groups[1].Value).To(BeElementOf(expectedValues))
		g.Expect(value.Groups[1].Name).To(Equal(value.Groups[1].Value)) // For ClusterId, value=name
	}
}

func validatePendingTimeGroupsWithNames(g Gomega, values []api.TelemetryResponseValuesInner, groupKey string, valueToName map[string]string) {
	for _, value := range values {
		g.Expect(len(value.Groups)).To(Equal(2))
		g.Expect(value.Groups[0].Key).To(Equal(Bucket))
		g.Expect(value.Groups[1].Key).To(Equal(groupKey))

		groupValue := value.Groups[1].Value
		expectedName, exists := valueToName[groupValue]
		g.Expect(exists).To(BeTrue(), "Unexpected group value: %s", groupValue)
		g.Expect(value.Groups[1].Name).To(Equal(expectedName))
	}
}

func validatePendingDistributionGroupCounts(g Gomega, values []api.TelemetryResponseValuesInner, expectedCounts map[string]int, expectedTotal int) {
	actualCounts := make(map[string]int)
	totalCount := 0

	for _, value := range values {
		var count int
		fmt.Sscanf(value.Value, "%d", &count)
		totalCount += count
		groupValue := value.Groups[1].Value
		actualCounts[groupValue] += count
	}

	g.Expect(totalCount).To(Equal(expectedTotal), "Total workload count mismatch")
	for groupValue, expectedCount := range expectedCounts {
		g.Expect(actualCounts[groupValue]).To(Equal(expectedCount), "Count mismatch for group %s", groupValue)
	}
}
