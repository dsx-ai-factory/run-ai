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

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	api2 "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
)

// Test plan:
// In 2 projects, create a workload in each, sync 2 pods in each project
// - One pod under the existing workload - validate it accessible directly after it
// - One pod not under this workload, but create its workload afterward - validate workload and pods are not accessible, and after it synced its accessible with the pod
// Delete a workload and see that we can access its pods
var _ = Describe("Pods API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	podsService := NewPodsService(mockTenantSettingsClient)
	workloadsService := NewWorkloadsService()
	revisionsService := NewRevisionsService()
	createdWorkloadsIds := []string{}
	createdPods := []api.PodSyncFields{}

	BeforeEach(func() {
		// Set default mock to prevent HTTP calls during pod sync
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock())
		setupProjectWithWorkloadAndPods(ctx, workloadsService, podsService, &createdWorkloadsIds, &createdPods, DefaultProjectId, testing.TestUuid1, 1)
		setupProjectWithWorkloadAndPods(ctx, workloadsService, podsService, &createdWorkloadsIds, &createdPods, AnotherProjectId, testing.TestUuid2, 1)
	})

	AfterEach(func() {
		for _, pod := range createdPods {
			apiResponse, _ := podsService.DeletePod(ctx, pod.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		}
		createdPods = []api.PodSyncFields{}

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

	Context("Get workload pod by id", func() {
		It("Returns the pod when it exists in the workload", func() {
			// testing.TestUuid1 is the workload of a seeded pod (see Get workload pods context).
			workloadPods, _ := podsService.GetWorkloadPods(ctx, testing.TestUuid1, false, 0, 0, "", "", "", nil)
			pods := httpResponseBodyToPodsResponse(workloadPods)
			Expect(len(pods.Pods)).To(BeNumerically(">", 0))
			seedPodId := pods.Pods[0].Id

			apiResponse, _ := podsService.GetWorkloadPod(ctx, testing.TestUuid1, seedPodId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			pod := httpResponseBodyToPod(apiResponse)
			Expect(pod.Id).To(Equal(seedPodId))
			Expect(pod.WorkloadId).To(Equal(testing.TestUuid1))
			// SetupBasePod seeds containers, so the preload should populate them.
			Expect(len(pod.Containers)).To(BeNumerically(">", 0), "pod should include containers (DbGetPod must preload)")
		})

		It("Returns 404 when the pod exists but belongs to a different workload", func() {
			// Defense-in-depth: a caller authorized for one workload shouldn't be able
			// to fetch pods belonging to another workload by guessing podIds.
			workloadPods, _ := podsService.GetWorkloadPods(ctx, testing.TestUuid1, false, 0, 0, "", "", "", nil)
			pods := httpResponseBodyToPodsResponse(workloadPods)
			Expect(len(pods.Pods)).To(BeNumerically(">", 0))
			seedPodId := pods.Pods[0].Id

			apiResponse, _ := podsService.GetWorkloadPod(ctx, uuid.NewString(), seedPodId)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Returns 404 when the pod does not exist", func() {
			apiResponse, _ := podsService.GetWorkloadPod(ctx, testing.TestUuid1, uuid.NewString())
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("Returns 400 when podId is not a uuid", func() {
			apiResponse, _ := podsService.GetWorkloadPod(ctx, testing.TestUuid1, "not-a-uuid")
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Returns 400 when workloadId is not a uuid", func() {
			apiResponse, _ := podsService.GetWorkloadPod(ctx, "not-a-uuid", uuid.NewString())
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Returns 404 for soft-deleted pods", func() {
			workloadPods, _ := podsService.GetWorkloadPods(ctx, testing.TestUuid1, false, 0, 0, "", "", "", nil)
			pods := httpResponseBodyToPodsResponse(workloadPods)
			Expect(len(pods.Pods)).To(BeNumerically(">", 0))
			seedPodId := pods.Pods[0].Id

			delResp, _ := podsService.DeletePod(ctx, seedPodId)
			Expect(delResp.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ := podsService.GetWorkloadPod(ctx, testing.TestUuid1, seedPodId)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

	})

	Context("DRA resource claims", func() {
		upsertAndGetPod := func(pod api.PodSyncFields) api.Pod {
			apiResponse, _ := podsService.UpsertPod(ctx, pod.Id, pod)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPod(ctx, pod.WorkloadId, pod.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			return httpResponseBodyToPod(apiResponse)
		}

		It("Persists template-generated and standalone claims as reported", func() {
			claims := []api.PodResourceClaim{
				{Name: "gpu-unbound", ResourceClaimTemplateName: ptr.To("single-gpu")},
				{Name: "gpu-bound", ResourceClaimTemplateName: ptr.To("single-gpu"), AllocatedResourceClaimName: ptr.To("dra-pod-gpu-bound-mk4s8")},
				{Name: "ib", ResourceClaimName: ptr.To("shared-infiniband"), AllocatedResourceClaimName: ptr.To("shared-infiniband")},
			}
			pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod.ResourceClaims = &claims
			createdPods = append(createdPods, pod)

			result := upsertAndGetPod(pod)
			Expect(result.ResourceClaims).NotTo(BeNil())
			Expect(*result.ResourceClaims).To(Equal(claims))
		})

		It("Omits resource claims for a pod that reports none", func() {
			pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			createdPods = append(createdPods, pod)

			result := upsertAndGetPod(pod)
			Expect(result.ResourceClaims).To(BeNil())
		})

		It("Overwrites resource claims on re-sync and clears them when no longer reported", func() {
			pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			createdPods = append(createdPods, pod)

			pod.ResourceClaims = &[]api.PodResourceClaim{
				{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
			}
			result := upsertAndGetPod(pod)
			Expect(*result.ResourceClaims).To(HaveLen(1))
			Expect((*result.ResourceClaims)[0].AllocatedResourceClaimName).To(BeNil())

			pod.ResourceClaims = &[]api.PodResourceClaim{
				{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu"), AllocatedResourceClaimName: ptr.To("dra-pod-gpu-1-abcde")},
			}
			result = upsertAndGetPod(pod)
			Expect(*result.ResourceClaims).To(HaveLen(1))
			Expect((*result.ResourceClaims)[0].AllocatedResourceClaimName).To(Equal(ptr.To("dra-pod-gpu-1-abcde")))

			pod.ResourceClaims = nil
			result = upsertAndGetPod(pod)
			Expect(result.ResourceClaims).To(BeNil())
		})
	})

	Context("Get workload pods", func() {
		It("Workload with pod", func() {
			apiResponse, _ := podsService.GetWorkloadPods(ctx, testing.TestUuid1, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(podsResponse.Next).To(Equal(int32(0)))
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1, nil)
		})

		It("Deleted workload with pod", func() {
			apiResponse, _ := podsService.GetWorkloadPods(ctx, testing.TestUuid2, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			countOfPods := len(podsResponse.Pods)
			Expect(countOfPods).To(BeNumerically(">", 0))

			apiResponse, _ = workloadsService.DeleteWorkload(ctx, testing.TestUuid2)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			removeWorkloadFromCreated(&createdWorkloadsIds, testing.TestUuid2)

			apiResponse, _ = podsService.GetWorkloadPods(ctx, testing.TestUuid2, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(BeNumerically(">=", countOfPods))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, testing.TestUuid2, AnotherProjectId, testing.ClusterUuid1, nil)
		})
	})
	Context("Get revision pods", func() {
		It("Revision with pod", func() {
			// we set up 3 pods:
			// one of the revision
			// one of another revision
			// one of no revision
			revisionId := uuid.NewString()
			pod := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod.RevisionId = ptr.To(revisionId)
			pod2 := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod2.RevisionId = ptr.To(uuid.NewString())
			pod3 := SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod, pod2, pod3},
			})
			createdPods = append(createdPods, pod, pod2, pod3)

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{ // required in order to sync the pod with a revision
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(revisionId, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetRevisionPods(ctx, revisionId, false)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1, &revisionId)
		})

		It("Deleted revision with pod", func() {
			revisionId := uuid.NewString()

			pod := SetupBasePod(uuid.NewString(), testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1)
			pod.RevisionId = ptr.To(revisionId)

			pod2 := SetupBasePod(uuid.NewString(), testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1)
			pod2.RevisionId = ptr.To(revisionId)
			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod, pod2},
			})
			createdPods = append(createdPods, pod, pod2)

			// sync two pods and then sync again with only one of them,
			// so the second is marked as deleted
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod, pod2},
			})
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod},
			})

			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{ // required in order to sync the pod with a revision
				ProjectId: DefaultProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(revisionId, testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetRevisionPods(ctx, revisionId, false)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], pod.Id, testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1, &revisionId)

			apiResponse, _ = podsService.GetRevisionPods(ctx, revisionId, true)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], pod2.Id, testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1, &revisionId)
		})
	})
	Context("Gets", func() {
		var pod1, pod2, pod3, pod4 api.PodSyncFields
		var revId string

		BeforeEach(func() {
			pod1 = SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod1.NodeName = pointer.String("worker-1")
			pod1.Name = "pod-1"
			pod2 = SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod2.NodeName = pointer.String("worker-2")
			pod2.CurrentNodePool = pointer.String("np-a")
			pod2.Name = "pod-2"
			pod3 = SetupBasePod(uuid.NewString(), testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			pod3.NodeName = pointer.String("worker-3")
			pod3.Name = "pod-3"
			// completed pod
			pod3.CompletedAt = ptr.To(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))

			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods: []api.PodSyncFields{
					pod1, pod2, pod3,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			workloadId := testing.TestUuid2
			clusterId := uuid.NewString()
			pod4 = SetupBasePod(uuid.NewString(), workloadId, AnotherProjectId, clusterId)
			pod4.NodeName = pointer.String("worker-1")
			pod4.Name = "pod-4"
			revId = uuid.NewString()
			pod4.RevisionId = ptr.To(revId)
			apiResponse, _ = revisionsService.SyncRevisions(ctx, api.RevisionsSyncRequest{ // required in order to sync the pod with a revision
				ProjectId: AnotherProjectId,
				Revisions: []api.RevisionSyncFields{
					SetupBaseRevision(revId, workloadId, AnotherProjectId, clusterId),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: AnotherProjectId,
				Pods: []api.PodSyncFields{
					pod4,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("Filter - Cluster id", func() {
			testPodsFilter(ctx, podsService, []string{fmt.Sprintf("clusterId==%s", pod4.ClusterId)}, 1, "")
			testPodsFilter(ctx, podsService, []string{fmt.Sprintf("clusterId==%s", uuid.NewString())}, 0, "")
		})
		It("Filter - Workload id", func() {
			testPodsFilter(ctx, podsService, []string{fmt.Sprintf("workloadId==%s", testing.TestUuid1)}, 3, "")
			testPodsFilter(ctx, podsService, []string{fmt.Sprintf("workloadId==%s", testing.TestUuid2)}, 1, "")
		})
		It("Filter - Node name", func() {
			testPodsFilter(ctx, podsService, []string{"nodeName=@rker-"}, 4, "")
			testPodsFilter(ctx, podsService, []string{"nodeName==worker-12"}, 0, "")
		})
		It("Filter - Current nodepool", func() {
			testPodsFilter(ctx, podsService, []string{"currentNodepool==np-a"}, 1, "")
			testPodsFilter(ctx, podsService, []string{"currentNodepool==some-other-np"}, 0, "")
		})
		It("Filter - Multi filters", func() {
			testPodsFilter(ctx, podsService, []string{"nodeName=@rker-", "currentNodepool!=some-other-np", fmt.Sprintf("clusterId==%s", pod4.ClusterId)}, 1, "")
		})
		It("Filter - Revision id", func() {
			pods := testPodsFilter(ctx, podsService, []string{fmt.Sprintf("revisionId==%s", revId)}, 1, "")
			Expect(len(pods.Pods)).To(Equal(1))
			Expect(pods.Pods[0].RevisionId).To(Not(BeNil()))
			Expect(*pods.Pods[0].RevisionId).To(Equal(revId))
		})
		It("Filter - Task", func() {
			testPodsFilter(ctx, podsService, []string{"task==Leader"}, 4, "")
			testPodsFilter(ctx, podsService, []string{"task==Worker"}, 0, "")
			testPodsFilter(ctx, podsService, []string{"task==blabla"}, 0, "")
		})
		It("Filter - Element id", func() {
			// Re-sync pod1 with elements
			pod1.Elements = &[]api.PodElement{
				{ElementId: "elem-abc", HierarchyLevel: 1},
				{ElementId: "elem-def", HierarchyLevel: 2},
			}
			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod1, pod2, pod3},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// brief verbosity (+ count via testPodsFilter)
			testPodsFilter(ctx, podsService, []string{"elementId==elem-abc"}, 1, "")
			testPodsFilter(ctx, podsService, []string{"elementId==elem-def"}, 1, "")
			testPodsFilter(ctx, podsService, []string{"elementId==non-existent"}, 0, "")
		})
		It("Filter - Element id with full verbosity", func() {
			pod1.Elements = &[]api.PodElement{
				{ElementId: "elem-abc", HierarchyLevel: 1},
				{ElementId: "elem-def", HierarchyLevel: 2},
			}
			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods:      []api.PodSyncFields{pod1, pod2, pod3},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// full verbosity path
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{"elementId==elem-abc"}, "full", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Name).To(Equal(pod1.Name))
			Expect(podsResponse.Pods[0].WorkloadName).To(Equal(DefaultWorkloadName))

			// non-matching — full verbosity
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{"elementId==non-existent"}, "full", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))
		})

		It("Search - Node name", func() {
			testPodsFilter(ctx, podsService, []string{}, 2, "worker-1")
			testPodsFilter(ctx, podsService, []string{}, 4, "worker-")
			testPodsFilter(ctx, podsService, []string{}, 0, "worker-12")
		})

		It("Verbosity - Get list pods verbosity - full", func() {
			apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"clusterId==" + testing.ClusterUuid1}, "full", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(podsResponse.Pods[0].WorkloadName).To(Equal(DefaultWorkloadName))
			Expect(podsResponse.Pods[0].ProjectName).To(Equal(DefaultProjectName))
		})
		It("Verbosity - Return invalid verbosity", func() {
			apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"clusterId==" + pod1.ClusterId}, "invalid", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("Completed query param - true", func() {
			apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"clusterId==" + testing.ClusterUuid1}, "full", "true", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].CompletedAt).ToNot(BeNil())
			Expect(podsResponse.Pods[0].Id).To(Equal(pod3.Id))
		})
		It("Completed query param - false", func() {
			apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"clusterId==" + testing.ClusterUuid1}, "full", "false", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))
			Expect(podsResponse.Pods[0].CompletedAt).To(BeNil())
			Expect(podsResponse.Pods[1].CompletedAt).To(BeNil())
			Expect([]string{podsResponse.Pods[0].Id, podsResponse.Pods[1].Id}).To(ConsistOf(pod1.Id, pod2.Id))
		})
		It("Completed query param - count matches the listed pods", func() {
			filters := []string{"clusterId==" + testing.ClusterUuid1}

			apiResponse, _ := podsService.CountPods(ctx, false, filters, "", "false")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(httpResponseBodyToWorkloadsCountResponse(apiResponse).Count).To(Equal(int64(2)))

			apiResponse, _ = podsService.CountPods(ctx, false, filters, "", "true")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(httpResponseBodyToWorkloadsCountResponse(apiResponse).Count).To(Equal(int64(1)))

			apiResponse, _ = podsService.CountPods(ctx, false, filters, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(httpResponseBodyToWorkloadsCountResponse(apiResponse).Count).To(Equal(int64(3)))
		})
		It("Sync twice - no duplicated containers but updated", func() {
			Expect(len(pod1.Containers)).To(Equal(1))
			Expect(pod1.Containers[0].Image).To(Equal(DefaultImages[0]))

			pod1.Containers[0].Image = "test"
			pod1.Containers[0].StartedAt = nil
			apiResponse, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				Pods: []api.PodSyncFields{
					pod1,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{"clusterId==" + pod1.ClusterId}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))

			Expect(len(podsResponse.Pods[0].Containers)).To(Equal(1))
			Expect(podsResponse.Pods[0].Containers[0].Image).To(Equal(pod1.Containers[0].Image))
			Expect(podsResponse.Pods[0].Containers[0].StartedAt).To(Equal(pod1.Containers[0].StartedAt))
		})
	})

	Context("SyncWorkloadsPods with ClusterId", func() {
		It("ClusterId - filters deletion by cluster", func() {
			// Create workloads in same project but different clusters
			workloadCluster3 := uuid.New().String()
			workloadCluster4 := uuid.New().String()

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadCluster3, SetupBaseWorkload(workloadCluster3, DefaultProjectId, testing.ClusterUuid3))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadCluster4, SetupBaseWorkload(workloadCluster4, DefaultProjectId, testing.ClusterUuid4))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadCluster3, workloadCluster4)

			// Create pods for both workloads in different clusters
			pod1Cluster3 := uuid.New().String()
			pod2Cluster3 := uuid.New().String()
			pod1Cluster4 := uuid.New().String()
			pod2Cluster4 := uuid.New().String()

			// Sync all pods initially (with clusterId)
			syncRequest := api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid3,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
					SetupBasePod(pod2Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			}
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, syncRequest)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, syncRequest.Pods...)

			syncRequest = api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid4,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
					SetupBasePod(pod2Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
				},
			}
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, syncRequest)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, syncRequest.Pods...)

			// Validate all 4 pods exist (2 per cluster)
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid4)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			// Sync cluster3 with only 1 pod - should soft-delete the other pod in cluster3 only
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid3,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Cluster3 should have 1 non-deleted pod
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Id).To(Equal(pod1Cluster3))

			// Cluster3 should have 1 deleted pod
			apiResponse, _ = podsService.ListPods(ctx, true, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Id).To(Equal(pod2Cluster3))

			// Cluster4 pods should be unaffected - still have 2 non-deleted pods
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid4)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			// Cluster4 should have 0 deleted pods
			apiResponse, _ = podsService.ListPods(ctx, true, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid4)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(0))

			// Cleanup
			podsService.DeletePod(ctx, pod1Cluster3)
			podsService.DeletePod(ctx, pod2Cluster3)
			podsService.DeletePod(ctx, pod1Cluster4)
			podsService.DeletePod(ctx, pod2Cluster4)
		})
	})

	Context("Internal Sync", func() {
		It("Sanity", func() {
			workloadId := testing.TestUuid3
			clusterId := testing.ClusterUuid3
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			apiResponse, _ = podsService.InternalSyncPods(ctx, api.InternalPodsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Pods:      []api.PodSyncFields{SetupBasePod(testing.TestUuid4, workloadId, DefaultProjectId, clusterId)},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, workloadId, DefaultProjectId, clusterId, nil)
		})
		It("Sync pod with no workload", func() {
			workloadId := testing.TestUuid3
			clusterId := testing.ClusterUuid3
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			apiResponse, _ = podsService.InternalSyncPods(ctx, api.InternalPodsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Pods: []api.PodSyncFields{
					SetupBasePod(testing.TestUuid4, workloadId, DefaultProjectId, clusterId),        // existing workload
					SetupBasePod(testing.TestUuid1, testing.TestUuid4, DefaultProjectId, clusterId), // not existing workload
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusMultiStatus))

			podsResponse := httpResponseBodyToSync207Response(apiResponse)
			Expect(len(podsResponse.SuccessfullySyncedPods)).To(Equal(1))
			Expect(podsResponse.SuccessfullySyncedPods[0]).To(Equal(testing.TestUuid4))
			Expect(len(podsResponse.FailedPods)).To(Equal(1))
			Expect(podsResponse.FailedPods[0]).To(Equal(testing.TestUuid1))
		})
	})
	Context("Per workload Sync", func() {
		It("Sanity", func() {
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			workloadId2 := uuid.New().String()
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadId2, SetupBaseWorkload(workloadId2, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			createdWorkloadsIds = append(createdWorkloadsIds, workloadId, workloadId2)

			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:   DefaultProjectId,
				WorkloadId:  workloadId,
				Pods:        []api.PodSyncFields{SetupBasePod(testing.TestUuid6, workloadId, DefaultProjectId, clusterId)},
				IdsToIgnore: nil,
			}, {
				ProjectId:  DefaultProjectId,
				WorkloadId: workloadId2,
				Pods: []api.PodSyncFields{SetupBasePod(testing.TestUuid5, workloadId2, DefaultProjectId, clusterId),
					SetupBasePod(testing.TestUuid7, workloadId2, DefaultProjectId, clusterId)},
				IdsToIgnore: nil,
			}})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, workloadId, DefaultProjectId, clusterId, nil)

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId2, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, workloadId2, DefaultProjectId, clusterId, nil)
			ValidateBasePod(podsResponse.Pods[1], podsResponse.Pods[1].Id, workloadId2, DefaultProjectId, clusterId, nil)

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId2, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))
		})

		It("ignoredIds", func() {
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:  DefaultProjectId,
				WorkloadId: workloadId,
				Pods: []api.PodSyncFields{SetupBasePod(testing.TestUuid5, workloadId, DefaultProjectId, clusterId),
					SetupBasePod(testing.TestUuid6, workloadId, DefaultProjectId, clusterId),
					SetupBasePod(testing.TestUuid7, workloadId, DefaultProjectId, clusterId)},
				IdsToIgnore: nil,
			}})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(3))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))

			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:   DefaultProjectId,
				WorkloadId:  workloadId,
				Pods:        []api.PodSyncFields{SetupBasePod(testing.TestUuid6, workloadId, DefaultProjectId, clusterId)},
				IdsToIgnore: []string{testing.TestUuid7},
			}})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Id).To(Equal(testing.TestUuid5))
		})

		It("ClusterId - filters deletion by cluster", func() {
			// Create two workloads in same project but different clusters
			workloadCluster3 := uuid.New().String()
			workloadCluster4 := uuid.New().String()

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadCluster3, SetupBaseWorkload(workloadCluster3, DefaultProjectId, testing.ClusterUuid3))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadCluster4, SetupBaseWorkload(workloadCluster4, DefaultProjectId, testing.ClusterUuid4))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadCluster3, workloadCluster4)

			// Create pods for both workloads
			pod1Cluster3 := uuid.New().String()
			pod2Cluster3 := uuid.New().String()
			pod1Cluster4 := uuid.New().String()
			pod2Cluster4 := uuid.New().String()

			// Sync pods for cluster3 workload with clusterId
			syncRequest := api.WorkloadPodsSyncRequest{
				ProjectId:  DefaultProjectId,
				WorkloadId: workloadCluster3,
				ClusterId:  testing.ClusterUuid3,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
					SetupBasePod(pod2Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			}
			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{syncRequest})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, syncRequest.Pods...)

			// Sync pods for cluster4 workload with clusterId
			syncRequest = api.WorkloadPodsSyncRequest{
				ProjectId:  DefaultProjectId,
				WorkloadId: workloadCluster4,
				ClusterId:  testing.ClusterUuid4,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
					SetupBasePod(pod2Cluster4, workloadCluster4, DefaultProjectId, testing.ClusterUuid4),
				},
			}
			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{syncRequest})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, syncRequest.Pods...)

			// Validate all 4 pods exist
			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadCluster3, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadCluster4, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			// Sync cluster3 with only 1 pod - should soft-delete the other pod in cluster3 only
			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:  DefaultProjectId,
				WorkloadId: workloadCluster3,
				ClusterId:  testing.ClusterUuid3,
				Pods: []api.PodSyncFields{
					SetupBasePod(pod1Cluster3, workloadCluster3, DefaultProjectId, testing.ClusterUuid3),
				},
			}})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Cluster3 should have 1 non-deleted pod
			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadCluster3, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(1))
			Expect(httpResponseBodyToPodsResponse(apiResponse).Pods[0].Id).To(Equal(pod1Cluster3))

			// Cluster3 should have 1 deleted pod
			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadCluster3, true, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(1))
			Expect(httpResponseBodyToPodsResponse(apiResponse).Pods[0].Id).To(Equal(pod2Cluster3))

			// Cluster4 pods should be unaffected - still have 2 non-deleted pods
			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadCluster4, false, 0, 0, "", "", "", nil)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(len(httpResponseBodyToPodsResponse(apiResponse).Pods)).To(Equal(2))

			// Cleanup
			podsService.DeletePod(ctx, pod1Cluster3)
			podsService.DeletePod(ctx, pod2Cluster3)
			podsService.DeletePod(ctx, pod1Cluster4)
			podsService.DeletePod(ctx, pod2Cluster4)
		})
	})
	Context("Workload pods with extra filters", func() {
		It("Task filter", func() {
			clusterId := testing.ClusterUuid3

			workloadId := uuid.New().String()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			leaderPod := SetupBasePod(testing.TestUuid6, workloadId, DefaultProjectId, clusterId)
			workerPod := SetupBasePod(testing.TestUuid7, workloadId, DefaultProjectId, clusterId)
			workerPod.Task = ptr.To("Worker")
			apiResponse, _ = podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{{
				ProjectId:   DefaultProjectId,
				WorkloadId:  workloadId,
				Pods:        []api.PodSyncFields{leaderPod, workerPod},
				IdsToIgnore: nil,
			}})

			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", []string{"task==Leader"})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			ValidateBasePod(podsResponse.Pods[0], podsResponse.Pods[0].Id, workloadId, DefaultProjectId, clusterId, nil)

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", []string{"task==Leader"})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))

			apiResponse, _ = podsService.GetWorkloadPods(ctx, workloadId, false, 0, 0, "", "", "", []string{"task==blabla"})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse = httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))
		})
	})

	Context("Network Topology Labels", func() {
		It("Should populate topology labels from cluster-service", func() {
			clusterId := testing.ClusterUuid3
			workloadId := uuid.New().String()

			// Setup cluster-service mock with topology labels for specific nodes
			nodeName1 := "worker-node-1"
			nodeName2 := "worker-node-2"
			nodeName3 := "worker-node-3"

			networkTopologyLabelValues1 := []string{
				"us-west-2",
				"zone-west-2a",
				"rack-10",
				"hostname-10",
			}

			networkTopologyLabelValues2 := []string{
				"region=us-east-1",
				"us-east-1b",
			}

			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodeNetworkTopologyLabelValues(clusterId, nodeName1, networkTopologyLabelValues1).
				WithNodeNetworkTopologyLabelValues(clusterId, nodeName2, networkTopologyLabelValues2))

			// Create workload
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			// Create pods with different nodes
			pod1 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod1.NodeName = ptr.To(nodeName1)
			pod1.Name = "pod-with-topology-1"

			pod2 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod2.NodeName = ptr.To(nodeName2)
			pod2.Name = "pod-with-topology-2"

			pod3 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod3.NodeName = ptr.To(nodeName3)
			pod3.Name = "pod-without-topology"

			pod4 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod4.NodeName = nil
			pod4.Name = "pod-no-node"

			// Sync pods
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: clusterId,
				Pods:      []api.PodSyncFields{pod1, pod2, pod3, pod4},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, pod1, pod2, pod3, pod4)

			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", clusterId)}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(4))
		})

		It("Should use cache and not call API multiple times for same node during SyncWorkloadsPods", func() {
			clusterId := uuid.New().String()
			workloadId := uuid.New().String()
			clusterVersion := "2.25.0"
			nodeName := "worker-node-cached"

			// Setup cluster with version 2.25 (supports topology address)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(clusterId, testing.BuildDisplayedCluster(clusterId, clusterVersion, nil)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			networkTopologyLabelValues := []string{
				"us-west-2",
				"zone-west-2a",
				"rack-10",
				"hostname-10",
			}

			// Setup cluster-service mock with topology labels and call tracking
			mock := clientsclusterservice.NewSimpleMock().
				WithNodeNetworkTopologyLabelValues(clusterId, nodeName, networkTopologyLabelValues)
			clusterservice.SetMock(mock)

			// Create workload
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			// Create multiple pods on the SAME node
			pod1 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod1.NodeName = ptr.To(nodeName)
			pod1.Name = "pod-same-node-1"

			pod2 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod2.NodeName = ptr.To(nodeName)
			pod2.Name = "pod-same-node-2"

			pod3 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod3.NodeName = ptr.To(nodeName)
			pod3.Name = "pod-same-node-3"

			// Sync pods - all 3 pods are on the same node
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: clusterId,
				Pods:      []api.PodSyncFields{pod1, pod2, pod3},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, pod1, pod2, pod3)

			// Verify caching behavior:
			// - GetNodeNetworkTopologyLabelValues is called 3 times (once per pod)
			// - GetNodeInternal (actual API) is called only once (caching)
			Expect(mock.GetNodeNetworkTopologyLabelValuesCallCount()).To(Equal(3))
			Expect(mock.GetNodeInternalCallCount()).To(Equal(1))
			Expect(mock.GetNodeInternalCallParams()).To(ConsistOf(clusterId + ":" + nodeName))
		})

		It("Should NOT populate topology labels for cluster version 2.24", func() {
			clusterId := uuid.New().String()
			workloadId := uuid.New().String()
			clusterVersion := "2.24.0"

			// Setup cluster with version 2.24 (below MinVersionTopologyAddress of 2.25.0)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(clusterId, testing.BuildDisplayedCluster(clusterId, clusterVersion, nil)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			// Setup cluster-service mock with topology labels
			nodeName1 := "worker-node-old-cluster"
			networkTopologyLabelValues := []string{
				"us-west-2",
				"zone-west-2a",
				"rack-20",
				"hostname-20",
			}

			clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
				WithNodeNetworkTopologyLabelValues(clusterId, nodeName1, networkTopologyLabelValues))

			// Create workload
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId)

			pod1 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, clusterId)
			pod1.NodeName = ptr.To(nodeName1)
			pod1.Name = "pod-old-cluster-version"

			// Sync pods
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: clusterId,
				Pods:      []api.PodSyncFields{pod1},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, pod1)

			filterBy := []string{fmt.Sprintf("clusterId==%s", clusterId)}
			filtersParam, _ := queryapi.NewParser(models.Pod{}.SortAndFilterFieldsToColumn()).Filters(filterBy)
			podsModelsResponse, err := dao.DbListPods(ctx, testing.TestTenantId, false, 0, 0, "", "", filtersParam, api2.VerbosityBrief, "", "", nil)
			Expect(err).To(BeNil())
			podsModels := podsModelsResponse.(*[]models.Pod)
			Expect(podsModels).NotTo(BeNil())
			Expect((*podsModels)[0].NetworkTopologyLabelValues).To(BeNil())
		})

		It("Should update topology labels via internal endpoint for all pods on node", func() {
			clusterId := testing.ClusterUuid3
			workloadId1 := uuid.New().String()
			workloadId2 := uuid.New().String()
			nodeName := "worker-node-bulk-update"

			// Create two workloads
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId1, SetupBaseWorkload(workloadId1, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId1)

			apiResponse, _ = workloadsService.UpsertWorkload(ctx, workloadId2, SetupBaseWorkload(workloadId2, DefaultProjectId, clusterId))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, workloadId2)

			// Create 3 pods on the same node from different workloads
			pod1 := SetupBasePod(uuid.NewString(), workloadId1, DefaultProjectId, clusterId)
			pod1.NodeName = ptr.To(nodeName)
			pod1.Name = "pod-1-workload-1"

			pod2 := SetupBasePod(uuid.NewString(), workloadId1, DefaultProjectId, clusterId)
			pod2.NodeName = ptr.To(nodeName)
			pod2.Name = "pod-2-workload-1"

			pod3 := SetupBasePod(uuid.NewString(), workloadId2, DefaultProjectId, clusterId)
			pod3.NodeName = ptr.To(nodeName)
			pod3.Name = "pod-1-workload-2"

			// Sync pods initially (without topology labels from cluster-service)
			clusterservice.SetMock(clientsclusterservice.NewSimpleMock())

			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: clusterId,
				Pods:      []api.PodSyncFields{pod1, pod2, pod3},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, pod1, pod2, pod3)

			newTopologyLabels := []string{
				"us-central-1",
				"us-central-1a",
				"rack-100",
				"switch-100",
			}

			updateRequest := api.UpdateNodePodsRequest{
				NetworkTopologyLabelValues: &newTopologyLabels,
			}

			apiResponse, _ = podsService.InternalUpdateNodePods(ctx, clusterId, nodeName, updateRequest)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			filterBy := []string{fmt.Sprintf("clusterId==%s", clusterId), "nodeName==" + nodeName}
			apiResponse, _ = podsService.ListPods(ctx, false, 0, 0, "", "", filterBy, api.PODVERBOSITY_BRIEF, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(3))

			filtersParam, _ := queryapi.NewParser(models.Pod{}.SortAndFilterFieldsToColumn()).Filters(filterBy)
			podsModelsResponse, err := dao.DbListPods(ctx, testing.TestTenantId, false, 0, 0, "", "", filtersParam, api2.VerbosityBrief, "", "", nil)
			Expect(err).To(BeNil())
			podsModels := podsModelsResponse.(*[]models.Pod)
			Expect(podsModels).NotTo(BeNil())
			Expect(len(*podsModels)).To(Equal(3))
			Expect([]string(*(*podsModels)[0].NetworkTopologyLabelValues)).To(ConsistOf(newTopologyLabels))
		})
	})

	Context("ListPods and CountPods with permitted scopes", func() {
		It("Tenant-scoped permissions returns all pods", func() {
			apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))

			apiResponse, _ = podsService.CountPods(ctx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(2)))
		})

		It("Project-scoped permissions returns only pods in permitted project", func() {
			projectCtx := testing.SetProjectContext(testing.TestTenant, testing.TestUser, []string{}, DefaultProjectId)
			apiResponse, _ := podsService.ListPods(projectCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].ProjectId).To(Equal(DefaultProjectId))

			apiResponse, _ = podsService.CountPods(projectCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(1)))
		})

		It("Project-scoped permissions for other project returns only that project pods", func() {
			projectCtx := testing.SetProjectContext(testing.TestTenant, testing.TestUser, []string{}, AnotherProjectId)
			apiResponse, _ := podsService.ListPods(projectCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].ProjectId).To(Equal(AnotherProjectId))

			apiResponse, _ = podsService.CountPods(projectCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(1)))
		})

		It("Project-scoped permissions for non-matching project returns no pods", func() {
			projectCtx := testing.SetProjectContext(testing.TestTenant, testing.TestUser, []string{}, "888888")
			apiResponse, _ := podsService.ListPods(projectCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))

			apiResponse, _ = podsService.CountPods(projectCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(0)))
		})

		It("Cluster-scoped permissions returns only pods in permitted cluster", func() {
			// Create a workload + pod in a different cluster to verify it gets excluded
			otherClusterWorkloadId := uuid.NewString()
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, otherClusterWorkloadId, SetupBaseWorkload(otherClusterWorkloadId, DefaultProjectId, testing.ClusterUuid2))
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, otherClusterWorkloadId)

			otherClusterPod := SetupBasePod(uuid.NewString(), otherClusterWorkloadId, DefaultProjectId, testing.ClusterUuid2)
			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid2,
				Pods:      []api.PodSyncFields{otherClusterPod},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdPods = append(createdPods, otherClusterPod)

			// Verify the extra pod is visible with tenant-scoped ctx
			apiResponse, _ = podsService.CountPods(ctx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(httpResponseBodyToWorkloadsCountResponse(apiResponse).Count).To(Equal(int64(3)))

			// Now list with cluster-scoped permissions for ClusterUuid1
			clusterCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Clusters: []string{testing.ClusterUuid1}},
			)
			apiResponse, _ = podsService.ListPods(clusterCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))
			for _, pod := range podsResponse.Pods {
				Expect(pod.ClusterId).To(Equal(testing.ClusterUuid1))
			}

			apiResponse, _ = podsService.CountPods(clusterCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(2)))
		})

		It("Cluster-scoped permissions for non-matching cluster returns no pods", func() {
			clusterCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Clusters: []string{uuid.NewString()}},
			)
			apiResponse, _ := podsService.ListPods(clusterCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))

			apiResponse, _ = podsService.CountPods(clusterCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(0)))
		})

		It("Multiple project scopes returns pods from all permitted projects", func() {
			multiCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Projects: []string{DefaultProjectId, AnotherProjectId}},
			)
			apiResponse, _ := podsService.ListPods(multiCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))

			apiResponse, _ = podsService.CountPods(multiCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(2)))
		})

		It("Department-scoped permissions returns pods whose workload belongs to that department", func() {
			deptCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Departments: []string{DefaultDepartmentId}},
			)
			apiResponse, _ := podsService.ListPods(deptCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(2))

			apiResponse, _ = podsService.CountPods(deptCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(2)))
		})

		It("Department-scoped permissions for non-matching department returns no pods", func() {
			deptCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Departments: []string{"888888"}},
			)
			apiResponse, _ := podsService.ListPods(deptCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(0))

			apiResponse, _ = podsService.CountPods(deptCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(0)))
		})

		It("Department and project scoped permissions returns union of matching pods", func() {
			deptAndProjCtx := context.WithValue(
				testing.SetContext(testing.TestTenant, testing.TestUser, []string{}),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{
					Departments: []string{"88888"},
					Projects:    []string{DefaultProjectId},
				},
			)
			apiResponse, _ := podsService.ListPods(deptAndProjCtx, false, 0, 0, "", "", []string{}, "brief", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].ProjectId).To(Equal(DefaultProjectId))

			apiResponse, _ = podsService.CountPods(deptAndProjCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(1)))
		})

		It("Project-scoped permissions with full verbosity returns filtered pods", func() {
			projectCtx := testing.SetProjectContext(testing.TestTenant, testing.TestUser, []string{}, DefaultProjectId)
			apiResponse, _ := podsService.ListPods(projectCtx, false, 0, 0, "", "", []string{}, "full", "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(apiResponse)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].ProjectId).To(Equal(DefaultProjectId))
			Expect(podsResponse.Pods[0].WorkloadName).To(Equal(DefaultWorkloadName))

			apiResponse, _ = podsService.CountPods(projectCtx, false, []string{}, "", "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(countResponse.Count).To(Equal(int64(1)))
		})

		It("Nil permitted scopes returns all pods", func() {
			filtersParam, _ := queryapi.NewParser(models.Pod{}.SortAndFilterFieldsToColumn()).Filters([]string{})
			podsModelsResponse, err := dao.DbListPods(ctx, testing.TestTenantId, false, 0, 0, "", "", filtersParam, api2.VerbosityBrief, "", "", nil)
			Expect(err).To(BeNil())
			podsModels := podsModelsResponse.(*[]models.Pod)
			Expect(len(*podsModels)).To(Equal(2))

			count, countErr := dao.DbCountPods(ctx, testing.TestTenantId, false, filtersParam, "", "", nil)
			Expect(countErr).To(BeNil())
			Expect(count).To(Equal(int64(2)))
		})
	})
})

func setupProjectWithWorkloadAndPods(ctx context.Context, workloadsService api.WorkloadsAPIServicer, podsService api.PodsAPIServicer, createdWorkloadsIds *[]string, createdPods *[]api.PodSyncFields, projectId string, workloadId string, podsUnderTheWorkloadCount int) (podsNotUnderTheWorkloadIds []string) {
	podsNotUnderTheWorkloadIds = []string{}

	apiResponse, _ := workloadsService.UpsertWorkload(ctx, workloadId, SetupBaseWorkload(workloadId, projectId, testing.ClusterUuid1))
	Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
	*createdWorkloadsIds = append(*createdWorkloadsIds, workloadId)

	podsToCreate := []api.PodSyncFields{}
	podsIdsToCreate := []string{}
	for i := 0; i < podsUnderTheWorkloadCount; i++ {
		id := uuid.NewString()
		podsToCreate = append(podsToCreate, SetupBasePod(id, workloadId, projectId, testing.ClusterUuid1))
		podsIdsToCreate = append(podsIdsToCreate, id)
	}

	apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
		ProjectId: projectId,
		Pods:      podsToCreate,
	})
	Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
	*createdPods = append(*createdPods, podsToCreate...)

	return podsNotUnderTheWorkloadIds
}

func httpResponseBodyToPodsResponse(httpResponse api.ImplResponse) api.ListPods200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.ListPods200Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func httpResponseBodyToPod(httpResponse api.ImplResponse) api.Pod {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	pod := &api.Pod{}
	err = json.Unmarshal(bytes, pod)
	Expect(err).To(BeNil())
	return *pod
}

func httpResponseBodyToSync207Response(httpResponse api.ImplResponse) api.InternalSyncPods207Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.InternalSyncPods207Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func testPodsFilter(ctx context.Context, podsService api.PodsAPIServicer, filters []string, expectedCount int, search string) api.ListPods200Response {
	apiResponse, _ := podsService.ListPods(ctx, false, 0, 0, "", "completedAt", filters, "brief", "", search)
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	podsResponse := httpResponseBodyToPodsResponse(apiResponse)
	Expect(len(podsResponse.Pods)).To(Equal(expectedCount))

	apiResponse, _ = podsService.CountPods(ctx, false, filters, search, "")
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)

	Expect(countResponse.Count).To(Equal(int64(expectedCount)))

	return podsResponse
}

func removeWorkloadFromCreated(createdWorkloadsIds *[]string, workloadId string) {
	for i, id := range *createdWorkloadsIds {
		if id == workloadId {
			*createdWorkloadsIds = append((*createdWorkloadsIds)[:i], (*createdWorkloadsIds)[i+1:]...)
			return
		}
	}
}
