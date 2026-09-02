/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"k8s.io/utils/ptr"
)

var _ = Describe("Revisions API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	revisionsService := NewRevisionsService()
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(mockTenantSettingsClient)

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid3, testing.BuildDisplayedCluster(testing.ClusterUuid3, testing.TestDefaultClusterVersion, convutils.Ref(true))))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	AfterEach(func() {
		// we use CleanTables and not individual workload and revision deletions,
		// as there's no api for deletion of revision children
		testing.CleanTables()
	})

	Context("Sync revisions", func() {
		It("Sanity", func() {
			// two workloads, one with one revision, one with two revisions
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			workloadId2 := uuid.New().String()
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadId2, SetupBaseWorkload(workloadId2, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid6, workloadId, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid5, workloadId2, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid7, workloadId2, DefaultProjectId, clusterId),
				},
			})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid6, workloadId, DefaultProjectId, clusterId, false)
			ValidateBaseRevisionChildren(ctx, testing.TestUuid6, DefaultChildrenIds)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid5, workloadId2, DefaultProjectId, clusterId, false)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid7, workloadId2, DefaultProjectId, clusterId, false)

			// get a non-existing revision
			apiResponse, _ = revisionsService.GetRevision(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Deleted and ignored", func() {
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid5, workloadId, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid6, workloadId, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid7, workloadId, DefaultProjectId, clusterId),
				},
			})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// we should have 3 revisions, 0 deleted
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid5, workloadId, DefaultProjectId, clusterId, false)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid6, workloadId, DefaultProjectId, clusterId, false)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid7, workloadId, DefaultProjectId, clusterId, false)

			// now sync again, with one revision deleted (TestUuid5), one left as (TestUuid6) is and one ignored (TestUuid7)
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid6, workloadId, DefaultProjectId, clusterId),
				},
				IdsToIgnore: []string{testing.TestUuid7},
			})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// now we should have 2 revisions, 1 deleted
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid5, workloadId, DefaultProjectId, clusterId, true)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid6, workloadId, DefaultProjectId, clusterId, false)
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid7, workloadId, DefaultProjectId, clusterId, false)
		})

		It("ClusterId - filters deletion by cluster", func() {
			// Create workloads in same project but different clusters
			workloadCluster3 := uuid.New().String()
			workloadCluster4 := uuid.New().String()

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadCluster3, SetupBaseWorkload(workloadCluster3, DefaultProjectId, testing.ClusterUuid3))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadCluster4, SetupBaseWorkload(workloadCluster4, DefaultProjectId, testing.ClusterUuid4))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Create revisions for both workloads in different clusters
			rev1Cluster3 := uuid.New().String()
			rev2Cluster3 := uuid.New().String()
			rev1Cluster4 := uuid.New().String()
			rev2Cluster4 := uuid.New().String()

			// Sync revisions for cluster3 with clusterId
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid3,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(rev1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
					SetupBaseRevision(rev2Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Sync revisions for cluster4 with clusterId
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid4,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(rev1Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
					SetupBaseRevision(rev2Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Validate all 4 revisions exist (not deleted)
			ValidateBaseRevision(ctx, revisionsService, rev1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3, false)
			ValidateBaseRevision(ctx, revisionsService, rev2Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3, false)
			ValidateBaseRevision(ctx, revisionsService, rev1Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4, false)
			ValidateBaseRevision(ctx, revisionsService, rev2Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4, false)

			// Sync cluster3 with only 1 revision - should soft-delete the other revision in cluster3 only
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid3,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(rev1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Cluster3 rev1 should still exist (not deleted)
			ValidateBaseRevision(ctx, revisionsService, rev1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3, false)
			// Cluster3 rev2 should be deleted
			ValidateBaseRevision(ctx, revisionsService, rev2Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3, true)

			// Cluster4 revisions should be unaffected - both should still exist (not deleted)
			ValidateBaseRevision(ctx, revisionsService, rev1Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4, false)
			ValidateBaseRevision(ctx, revisionsService, rev2Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4, false)
		})

		It("Pods as revision children", func() {
			// a workload with a revision that has pods as children, validate that the pods are returned as part of
			// the children list, even that they are synced separately (not part of the revision sync)
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			revision := SetupBaseRevision(testing.TestUuid6, workloadId, DefaultProjectId, clusterId)
			revision.ChildrenIds = nil
			revision.RequestedPods = api.RequestedPods{
				Min: ptr.To(int32(1)),
				Max: ptr.To(int32(2)),
			}
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{revision},
			})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// the revision should have a single child, the revision itself
			ValidateBaseRevisionChildren(ctx, testing.TestUuid6, []api.WorkloadChildrenIdsInner{})

			// sync a pod with the revision as parent
			pod := SetupBasePod(testing.TestUuid7, workloadId, DefaultProjectId, clusterId)
			pod.RevisionId = ptr.To(testing.TestUuid6)
			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:   DefaultProjectId,
				WorkloadId:  workloadId,
				Pods:        []api.PodSyncFields{pod},
				IdsToIgnore: nil,
			},
			})

			// the revision should now have two children, the revision itself and the pod
			ValidateBaseRevisionChildren(ctx, testing.TestUuid6, []api.WorkloadChildrenIdsInner{
				{
					Id:   testing.TestUuid7,
					Type: "Pod",
				},
			})

			// validate the RequestedPods entry in GetRevision response.
			// RequestedPods.Number is coming from the sum of pods, not from the revision itself
			apiResponse, _ = revisionsService.GetRevision(ctx, testing.TestUuid6)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.Revision)
			Expect(response.RequestedPods.Number).NotTo(BeNil())
			Expect(*response.RequestedPods.Number).To(Equal(int32(1)))
			Expect(response.RequestedPods.Min).NotTo(BeNil())
			Expect(*response.RequestedPods.Min).To(Equal(int32(1)))
			Expect(response.RequestedPods.Max).NotTo(BeNil())
			Expect(*response.RequestedPods.Max).To(Equal(int32(2)))
			Expect(response.RequestedPods.Parallelism).To(BeNil())
			Expect(response.RequestedPods.Completions).To(BeNil())
		})
	})

	Context("Get and delete revision by id", func() {
		It("Sanity", func() {
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid6, workloadId, DefaultProjectId, clusterId),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// get the revision and validate it
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid6, workloadId, DefaultProjectId, clusterId, false)

			// delete the revision
			apiResponse, _ = revisionsService.DeleteRevision(ctx, testing.TestUuid6)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// get the revision again, should be returned as deleted
			ValidateBaseRevision(ctx, revisionsService, testing.TestUuid6, workloadId, DefaultProjectId, clusterId, true)
		})
	})

	Context("Get and count inference workload revisions", func() {
		It("Get and count inference workload revisions", func() {
			// create two workloads
			clusterId := testing.ClusterUuid3
			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			workloadId2 := uuid.New().String()
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadId2, SetupBaseWorkload(workloadId2, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// sync 1 revision to the first workload and 2 revisions to the second one
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid5, workloadId, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid6, workloadId2, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid7, workloadId2, DefaultProjectId, clusterId),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// get the revisions and validate them
			ValidateGetRevisions(ctx, revisionsService, workloadId, false, []string{testing.TestUuid5})
			ValidateGetRevisions(ctx, revisionsService, workloadId, true, []string{})
			ValidateGetRevisions(ctx, revisionsService, workloadId2, false, []string{testing.TestUuid6, testing.TestUuid7})
			ValidateGetRevisions(ctx, revisionsService, workloadId2, true, []string{})

			// sync again with only 1 and 1, i.e. delete one revision from the second workload
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(testing.TestUuid5, workloadId, DefaultProjectId, clusterId),
					SetupBaseRevision(testing.TestUuid7, workloadId2, DefaultProjectId, clusterId),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			ValidateGetRevisions(ctx, revisionsService, workloadId, false, []string{testing.TestUuid5})
			ValidateGetRevisions(ctx, revisionsService, workloadId, true, []string{})
			ValidateGetRevisions(ctx, revisionsService, workloadId2, false, []string{testing.TestUuid7})
			ValidateGetRevisions(ctx, revisionsService, workloadId2, true, []string{testing.TestUuid6})
		})
	})
})
