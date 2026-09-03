/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
)

var _ = Describe("Workloads-Pods API advanced tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	podsService := NewPodsService(mockTenantSettingsClient)
	revisionsService := NewRevisionsService()
	createdWorkloadsIds := []string{}
	createdRevisionsIds := []string{}
	createdPods := []api.PodSyncFields{}

	workloadsDao := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]())
	var tenantUint uint = 1

	AfterEach(func() {
		for _, pod := range createdPods {
			apiResponse, _ := podsService.DeletePod(ctx, pod.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		}
		createdPods = []api.PodSyncFields{}

		for _, id := range createdRevisionsIds {
			apiResponse, _ := revisionsService.DeleteRevision(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		}
		createdRevisionsIds = []string{}

		for _, id := range createdWorkloadsIds {
			apiResponse, _ := workloadsService.DeleteWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		}
		createdWorkloadsIds = []string{}

		waitForGetWorkloadsLength(ctx, workloadsService, 0, []string{})
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("Workload view computed fields", func() {
		It("Workload request without pods", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.RequestedResources.Cpu = &api.RequestResourceCores{
				Limit:   pointer.Float64(16),
				Request: pointer.Float64(8),
			}
			baseWorkload.RequestedResources.CpuMemory = &api.RequestResourceQuantity{
				Limit:   pointer.String("200Mi"),
				Request: pointer.String("2G"),
			}
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid1)

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.WorkloadRequestedResources.Cpu.Request).ToNot(BeNil())
			Expect(*workloadResponse.WorkloadRequestedResources.Cpu.Request).To(Equal(float64(8)))
			Expect(workloadResponse.WorkloadRequestedResources.Cpu.Limit).ToNot(BeNil())
			Expect(*workloadResponse.WorkloadRequestedResources.Cpu.Limit).To(Equal(float64(16)))
			Expect(workloadResponse.WorkloadRequestedResources.CpuMemory.Request).ToNot(BeNil())
			Expect(*workloadResponse.WorkloadRequestedResources.CpuMemory.Request).To(Equal("2G"))
			Expect(workloadResponse.WorkloadRequestedResources.CpuMemory.Limit).ToNot(BeNil())
			Expect(*workloadResponse.WorkloadRequestedResources.CpuMemory.Limit).To(Equal("200Mi"))
			Expect(workloadResponse.CompletedAt).To(BeNil())
		})
		It("Workload multiple pods", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid1)

			id1 := uuid.NewString()
			basePod := SetupBasePod(id1, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			basePod.RequestedResources.Cpu = &api.RequestResourceCores{
				Limit:   pointer.Float64(16),
				Request: pointer.Float64(8),
			}
			basePod.RequestedResources.CpuMemory = &api.RequestResourceQuantity{
				Limit:   pointer.String("200Mi"),
				Request: pointer.String("2G"),
			}
			basePod.AllocatedResources = &api.AllocatedResources{
				Gpu:       basePod.RequestedResources.Gpu.Request,
				Cpu:       basePod.RequestedResources.Cpu.Request,
				CpuMemory: basePod.RequestedResources.CpuMemory.Request,
			}
			basePod.CompletedAt = convutils.Ref(time.Now())
			id2 := uuid.NewString()
			basePod2 := basePod
			basePod2.Id = id2
			basePod2.CurrentNodePool = pointer.String(OtherNodepool)
			basePod2.CompletedAt = convutils.Ref(basePod.CompletedAt.Add(1 * time.Hour))

			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods: []api.PodSyncFields{
					basePod,
					basePod2,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, basePod, basePod2)

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.PodsRequestedResources.Cpu.Request).ToNot(BeNil())
			Expect(*workloadResponse.PodsRequestedResources.Cpu.Request).To(Equal(float64(16)))
			Expect(workloadResponse.PodsRequestedResources.Cpu.Limit).ToNot(BeNil())
			Expect(*workloadResponse.PodsRequestedResources.Cpu.Limit).To(Equal(float64(32)))
			Expect(workloadResponse.PodsRequestedResources.CpuMemory.Request).ToNot(BeNil())
			Expect(*workloadResponse.PodsRequestedResources.CpuMemory.Request).To(Equal("4G"))
			Expect(workloadResponse.PodsRequestedResources.CpuMemory.Limit).ToNot(BeNil())
			Expect(*workloadResponse.PodsRequestedResources.CpuMemory.Limit).To(Equal("400Mi"))
			Expect(len(workloadResponse.CurrentNodePools)).To(Equal(2))
			Expect(workloadResponse.CurrentNodePools).To(ContainElements(DefaultNodepool, OtherNodepool))
			Expect(workloadResponse.CompletedAt).To(BeNil())

			baseWorkload.Phase = api.PHASE_COMPLETED
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse = httpResponseBodyToWorkloadResponse(apiResponse)

			Expect(workloadResponse.CompletedAt).ToNot(BeNil())
			Expect(workloadResponse.CompletedAt.Unix()).To(Equal(workloadResponse.PhaseUpdatedAt.Unix()))
		})
	})

	Context("Total pending time", func() {
		It("workload in pending phase - calculate total pending and running times", func() {
			uid := uuid.New().String()
			baseWorkload := SetupBaseWorkload(uid, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, uid)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, nil, ptr.To(int32(0)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, nil, nil)
			time.Sleep(1 * time.Second)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, nil, ptr.To(int32(1)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, nil, ptr.To(int32(1)))

			baseWorkload.Phase = api.PHASE_PENDING
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			time.Sleep(1 * time.Second)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(1)), ptr.To(int32(1)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(1)), ptr.To(int32(1)))
			time.Sleep(1 * time.Second)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(2)), ptr.To(int32(1)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(2)), ptr.To(int32(1)))

			baseWorkload.Phase = api.PHASE_RUNNING
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(2)), ptr.To(int32(1)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(2)), ptr.To(int32(1)))
			time.Sleep(1 * time.Second)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(2)), ptr.To(int32(2)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(2)), ptr.To(int32(2)))

			baseWorkload.Phase = api.PHASE_PENDING
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(2)), ptr.To(int32(2)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(2)), ptr.To(int32(2)))
			time.Sleep(2 * time.Second)
			validateWorkloadTotalPendingRunningTimeSecondsDb(ctx, workloadsDao, tenantUint, uid, ptr.To(int32(4)), ptr.To(int32(2)))
			validateWorkloadTotalPendingRunningTimeSeconds(ctx, workloadsService, uid, ptr.To(int32(4)), ptr.To(int32(2)))
		})

		It("Test sortBy/filterBy totalPendingTimeSeconds and totalRunningTimeSeconds", func() {
			var wg sync.WaitGroup
			for i := 0; i < 4; i += 1 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					uid := uuid.New().String()
					baseWorkload := SetupBaseWorkload(uid, DefaultProjectId, testing.ClusterUuid1)
					baseWorkload.Phase = api.PHASE_PENDING
					apiResponse, _ := workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
					Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
					createdWorkloadsIds = append(createdWorkloadsIds, uid)

					if i%2 == 0 {
						time.Sleep(1 * time.Second)
					} else {
						time.Sleep(2 * time.Second)
					}

					baseWorkload.Phase = api.PHASE_RUNNING
					apiResponse, _ = workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
					Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

					if i%2 == 0 {
						time.Sleep(1 * time.Second)
					} else {
						time.Sleep(2 * time.Second)
					}

					baseWorkload.Phase = api.PHASE_FAILED
					apiResponse, _ = workloadsService.UpsertWorkload(ctx, uid, baseWorkload)
					Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
				}()
			}

			// Wait for all goroutines to finish
			wg.Wait()

			apiResponse, _ := workloadsService.GetWorkloads(ctx, false, 0, 0, "", "totalPendingTimeSeconds", []string{}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(4))
			Expect(workloadsResponse.Workloads[0].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[0].TotalPendingTimeSeconds).To(Equal(int32(1)))
			Expect(workloadsResponse.Workloads[1].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[1].TotalPendingTimeSeconds).To(Equal(int32(1)))
			Expect(workloadsResponse.Workloads[2].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[2].TotalPendingTimeSeconds).To(Equal(int32(2)))
			Expect(workloadsResponse.Workloads[3].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[3].TotalPendingTimeSeconds).To(Equal(int32(2)))

			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{"totalPendingTimeSeconds>=2"}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(2))
			Expect(workloadsResponse.Workloads[0].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[0].TotalPendingTimeSeconds).To(Equal(int32(2)))
			Expect(workloadsResponse.Workloads[1].TotalPendingTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[1].TotalPendingTimeSeconds).To(Equal(int32(2)))

			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "totalRunningTimeSeconds", []string{}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(4))
			Expect(workloadsResponse.Workloads[0].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[0].TotalRunningTimeSeconds).To(Equal(int32(1)))
			Expect(workloadsResponse.Workloads[1].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[1].TotalRunningTimeSeconds).To(Equal(int32(1)))
			Expect(workloadsResponse.Workloads[2].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[2].TotalRunningTimeSeconds).To(Equal(int32(2)))
			Expect(workloadsResponse.Workloads[3].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[3].TotalRunningTimeSeconds).To(Equal(int32(2)))

			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{"totalRunningTimeSeconds<=1"}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(2))
			Expect(workloadsResponse.Workloads[0].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[0].TotalRunningTimeSeconds).To(Equal(int32(1)))
			Expect(workloadsResponse.Workloads[1].TotalRunningTimeSeconds).ToNot(BeNil())
			Expect(*workloadsResponse.Workloads[1].TotalRunningTimeSeconds).To(Equal(int32(1)))
		})
	})

	Context("Revision view computed fields", func() {
		It("Revision request without pods", func() {
			workloadId := testing.TestUuid1
			revisionId := testing.TestUuid6

			revision := SetupBaseRevision(revisionId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			revision.RequestedResources.Cpu = &api.RequestResourceCores{
				Limit:   pointer.Float64(16),
				Request: pointer.Float64(8),
			}
			revision.RequestedResources.CpuMemory = &api.RequestResourceQuantity{
				Limit:   pointer.String("200Mi"),
				Request: pointer.String("2G"),
			}

			apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{revision},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			createdRevisionsIds = append(createdRevisionsIds, revisionId)

			apiResponse, _ = revisionsService.GetRevision(ctx, revisionId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body).ToNot(BeNil())
			revisionResponse := apiResponse.Body.(api.Revision)

			Expect(revisionResponse.RevisionRequestedResources.Cpu.Request).ToNot(BeNil())
			Expect(*revisionResponse.RevisionRequestedResources.Cpu.Request).To(Equal(float64(8)))
			Expect(revisionResponse.RevisionRequestedResources.Cpu.Limit).ToNot(BeNil())
			Expect(*revisionResponse.RevisionRequestedResources.Cpu.Limit).To(Equal(float64(16)))
			Expect(revisionResponse.RevisionRequestedResources.CpuMemory.Request).ToNot(BeNil())
			Expect(*revisionResponse.RevisionRequestedResources.CpuMemory.Request).To(Equal("2G"))
			Expect(revisionResponse.RevisionRequestedResources.CpuMemory.Limit).ToNot(BeNil())
			Expect(*revisionResponse.RevisionRequestedResources.CpuMemory.Limit).To(Equal("200Mi"))
		})
		It("Revision multiple pods", func() {
			workloadId := testing.TestUuid1
			revisionId := testing.TestUuid6

			// setup a revision
			revision := SetupBaseRevision(revisionId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{revision},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			createdRevisionsIds = append(createdRevisionsIds, revisionId)

			// setup two pods, only one of them should be running
			id1 := uuid.NewString()
			basePod := SetupBasePod(id1, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			basePod.RevisionId = &revisionId
			basePod.K8sPhase = "Running"
			basePod.RequestedResources.Cpu = &api.RequestResourceCores{
				Limit:   pointer.Float64(16),
				Request: pointer.Float64(8),
			}
			basePod.RequestedResources.CpuMemory = &api.RequestResourceQuantity{
				Limit:   pointer.String("3G"),
				Request: pointer.String("2G"),
			}
			basePod.AllocatedResources = &api.AllocatedResources{
				Gpu:       basePod.RequestedResources.Gpu.Request,
				Cpu:       basePod.RequestedResources.Cpu.Request,
				CpuMemory: basePod.RequestedResources.CpuMemory.Request,
			}
			id2 := uuid.NewString()
			basePod2 := basePod
			basePod2.Id = id2
			basePod2.RevisionId = &revisionId
			basePod.K8sPhase = "Terminating"
			basePod2.CurrentNodePool = pointer.String(OtherNodepool)
			basePod2.AllocatedResources = &api.AllocatedResources{
				Gpu:       basePod.RequestedResources.Gpu.Limit,
				Cpu:       basePod.RequestedResources.Cpu.Limit,
				CpuMemory: basePod.RequestedResources.CpuMemory.Limit,
			}

			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods: []api.PodSyncFields{
					basePod,
					basePod2,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, basePod, basePod2)

			apiResponse, _ = revisionsService.GetRevision(ctx, revisionId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body).ToNot(BeNil())
			revisionResponse := apiResponse.Body.(api.Revision)

			Expect(revisionResponse.PodsRequestedResources.Cpu.Request).ToNot(BeNil())
			Expect(*revisionResponse.PodsRequestedResources.Cpu.Request).To(Equal(float64(16)))
			Expect(revisionResponse.PodsRequestedResources.Cpu.Limit).ToNot(BeNil())
			Expect(*revisionResponse.PodsRequestedResources.Cpu.Limit).To(Equal(float64(32)))
			Expect(revisionResponse.PodsRequestedResources.CpuMemory.Request).ToNot(BeNil())
			Expect(*revisionResponse.PodsRequestedResources.CpuMemory.Request).To(Equal("4G"))
			Expect(revisionResponse.PodsRequestedResources.CpuMemory.Limit).ToNot(BeNil())
			Expect(*revisionResponse.PodsRequestedResources.CpuMemory.Limit).To(Equal("6G"))
			Expect(*revisionResponse.AllocatedResources.Gpu).To(Equal(3.0))
			Expect(revisionResponse.AllocatedResources.GpuMemory).To(BeNil())
			Expect(*revisionResponse.AllocatedResources.Cpu).To(Equal(24.0))
			Expect(*revisionResponse.AllocatedResources.CpuMemory).To(Equal("5G"))
			Expect(len(revisionResponse.CurrentNodePools)).To(Equal(2))
			Expect(revisionResponse.CurrentNodePools).To(ContainElements(DefaultNodepool, OtherNodepool))
			Expect(revisionResponse.RunningPods).To(Equal(int32(1)))
		})
	})
})

func validateWorkloadTotalPendingRunningTimeSeconds(ctx context.Context, workloadsService api.WorkloadsAPIServicer,
	workloadId string, expectedTotalPendingTimeSeconds, expectedTotalRunningTimeSeconds *int32) {
	apiResponse, err := workloadsService.GetWorkload(ctx, workloadId)
	ExpectWithOffset(1, err).To(BeNil())
	ExpectWithOffset(1, apiResponse.Code).To(Equal(http.StatusOK))
	workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

	if expectedTotalPendingTimeSeconds == nil {
		ExpectWithOffset(1, workloadResponse.TotalPendingTimeSeconds).To(BeNil(), "Expected total pending time to be nil")
		return
	} else {
		ExpectWithOffset(1, workloadResponse.TotalPendingTimeSeconds).ToNot(BeNil(), "Expected total pending time to be not nil")
		ExpectWithOffset(1, *workloadResponse.TotalPendingTimeSeconds).To(Equal(*expectedTotalPendingTimeSeconds), "Expected total pending time to be %d", *expectedTotalPendingTimeSeconds)
	}

	if expectedTotalRunningTimeSeconds == nil {
		ExpectWithOffset(1, workloadResponse.TotalRunningTimeSeconds).To(BeNil(), "Expected total running time to be nil")
		return
	} else {
		ExpectWithOffset(1, workloadResponse.TotalRunningTimeSeconds).ToNot(BeNil(), "Expected total running time to be not nil")
		ExpectWithOffset(1, *workloadResponse.TotalRunningTimeSeconds).To(Equal(*expectedTotalRunningTimeSeconds), "Expected total running time to be %d", *expectedTotalRunningTimeSeconds)
	}
}

func validateWorkloadTotalPendingRunningTimeSecondsDb(ctx context.Context, workloadsDao *dao.WorkloadsDao,
	tenantId uint, workloadId string, expectedTotalPendingTimeSeconds, expectedTotalRunningTimeSeconds *int32) {
	wUid := uuid.MustParse(workloadId)

	workloadView, err := workloadsDao.DbGetWorkload(ctx, tenantId, wUid)
	ExpectWithOffset(1, err).To(BeNil())

	if expectedTotalPendingTimeSeconds == nil {
		ExpectWithOffset(1, workloadView.TotalPendingTimeSeconds).To(BeNil(), "Expected total pending time to be nil")
	} else {
		ExpectWithOffset(1, workloadView.TotalPendingTimeSeconds).ToNot(BeNil(), "Expected total pending time to be not nil")
		ExpectWithOffset(1, *workloadView.TotalPendingTimeSeconds).To(Equal(*expectedTotalPendingTimeSeconds), "Expected total pending time to be %d", *expectedTotalPendingTimeSeconds)
	}

	if expectedTotalRunningTimeSeconds == nil {
		ExpectWithOffset(1, workloadView.TotalRunningTimeSeconds).To(BeNil(), "Expected total running time to be nil")
	} else {
		ExpectWithOffset(1, workloadView.TotalRunningTimeSeconds).ToNot(BeNil(), "Expected total running time to be not nil")
		ExpectWithOffset(1, *workloadView.TotalRunningTimeSeconds).To(Equal(*expectedTotalRunningTimeSeconds), "Expected total running time to be %d", *expectedTotalRunningTimeSeconds)
	}

}
