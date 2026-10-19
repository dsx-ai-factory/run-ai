/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"math"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/runaimetrics"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

var _ = Describe("Workloads API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	internalWorkloadsService := workloadsService.(*WorkloadsService)
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	podsService := NewPodsService(mockTenantSettingsClient)

	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
		WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.22.0", nil)).
		WithCluster(testing.ClusterUuid3, testing.BuildDisplayedCluster(testing.ClusterUuid3, testing.TestDefaultClusterVersion, nil)).
		WithCluster(testing.ClusterUuid4, testing.BuildDisplayedCluster(testing.ClusterUuid4, testing.TestDefaultClusterVersion, convutils.Ref(true))))
	checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

	createdWorkloadsIds := []string{}
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()

	appId := uuid.NewString()
	appName := "application-1"

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting workload tests")

		// Project 2 workloads
		workload1 := SetupBaseWorkload(id1, AnotherProjectId, testing.ClusterUuid2)
		workload1.Priority = nil
		workload1.ApplicationId = &appId
		workload1.ApplicationName = &appName
		workload2 := SetupBaseWorkload(id2, AnotherProjectId, testing.ClusterUuid2)
		workload2.Priority = nil
		workload2.AdditionalFields = "{\"priority\":50}"
		workload2.DistributedFramework = PytorchFramework
		workload3 := SetupBaseWorkload(id3, AnotherProjectId, testing.ClusterUuid2)
		priority := int32(200)
		workload3.Priority = &priority
		apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: AnotherProjectId,
			Workloads: []api.WorkloadSyncFields{
				workload1,
				workload2,
				workload3,
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		createdWorkloadsIds = append(createdWorkloadsIds, id1, id2, id3)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
			WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.22.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	BeforeEach(func() {
		// Project 1 workloads
		apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			Workloads: []api.WorkloadSyncFields{
				SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
				SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1),
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

		waitForGetWorkloadsLength(ctx, workloadsService, 6, []string{})
		createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid1, testing.TestUuid2, testing.TestUuid3)

		apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
			ProjectId: DefaultProjectId,
			Pods: []api.PodSyncFields{
				SetupBasePod(testing.TestUuid4, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
			},
		})
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("Sync", func() {
		It("returns topology in GET workload and list", func() {
			// arrange: create one workload with topology and one without
			withTopoId := uuid.NewString()
			noTopoId := uuid.NewString()
			withTopo := SetupBaseWorkload(withTopoId, DefaultProjectId, testing.ClusterUuid1)
			withTopo.Topology = &api.WorkloadSyncFieldsTopology{
				Name:               "nvl-topology",
				PreferredPlacement: "rack-a",
				RequiredPlacement:  "zone-a",
			}
			noTopo := SetupBaseWorkload(noTopoId, DefaultProjectId, testing.ClusterUuid1)
			noTopo.Topology = nil

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{withTopo, noTopo},
			})
			Expect(apiResponse.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

			// act + assert: GET by id for workload with topology
			apiResponse, _ = workloadsService.GetWorkload(ctx, withTopoId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.Topology).ToNot(BeNil())
			Expect(wd.Topology.Name).To(Equal("nvl-topology"))
			Expect(wd.Topology.PreferredPlacement).To(Equal("rack-a"))
			Expect(wd.Topology.RequiredPlacement).To(Equal("zone-a"))
			Expect(wd.Topology.ActualPlacement).To(BeNil())

			// act + assert: GET by id for workload without topology
			apiResponse, _ = workloadsService.GetWorkload(ctx, noTopoId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.Topology).To(BeNil())

			// list endpoint should include topology for the first workload (filter by id is not supported)
			waitForGetWorkloadsLength(ctx, workloadsService, 5, []string{})
			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			list := httpResponseBodyToWorkloadsResponse(apiResponse)
			var found *api.Workload
			for i := range list.Workloads {
				if list.Workloads[i].Id == withTopoId {
					found = &list.Workloads[i]
					break
				}
			}
			Expect(found).ToNot(BeNil())
			Expect(found.Topology).ToNot(BeNil())
			Expect(found.Topology.Name).To(Equal("nvl-topology"))
			Expect(found.Topology.ActualPlacement).To(BeNil())
		})

		It("returns kubernetesRef only when both group and version are set", func() {
			withRefId := uuid.NewString()
			groupOnlyId := uuid.NewString()
			withoutRefId := uuid.NewString()

			withRef := SetupBaseWorkload(withRefId, DefaultProjectId, testing.ClusterUuid1)
			withRef.Group = "ray.io"
			withRef.Version = "v1"

			// Mirrors a native workload: runai_workload_parser sets Group but not Version.
			groupOnly := SetupBaseWorkload(groupOnlyId, DefaultProjectId, testing.ClusterUuid1)
			groupOnly.Group = "run.ai"

			withoutRef := SetupBaseWorkload(withoutRefId, DefaultProjectId, testing.ClusterUuid1)

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{withRef, groupOnly, withoutRef},
			})
			Expect(apiResponse.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

			apiResponse, _ = workloadsService.GetWorkload(ctx, withRefId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.KubernetesRef).ToNot(BeNil())
			Expect(wd.KubernetesRef.Group).To(Equal("ray.io"))
			Expect(wd.KubernetesRef.Version).To(Equal("v1"))

			apiResponse, _ = workloadsService.GetWorkload(ctx, groupOnlyId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.KubernetesRef).To(BeNil())

			apiResponse, _ = workloadsService.GetWorkload(ctx, withoutRefId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.KubernetesRef).To(BeNil())

			// list endpoint should apply the same rule. Eventually() because
			// the materialized view refreshes async (20ms cadence), and a
			// length check alone can race with the next snapshot swap.
			EventuallyWithOffset(1, func(g Gomega) {
				apiResp, _ := workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResp.Code).To(Equal(http.StatusOK))
				ls := httpResponseBodyToWorkloadsResponse(apiResp)
				var lWithRef, lGroupOnly, lWithoutRef *api.Workload
				for i := range ls.Workloads {
					switch ls.Workloads[i].Id {
					case withRefId:
						lWithRef = &ls.Workloads[i]
					case groupOnlyId:
						lGroupOnly = &ls.Workloads[i]
					case withoutRefId:
						lWithoutRef = &ls.Workloads[i]
					}
				}
				g.Expect(lWithRef).ToNot(BeNil())
				g.Expect(lWithRef.KubernetesRef).ToNot(BeNil())
				g.Expect(lWithRef.KubernetesRef.Group).To(Equal("ray.io"))
				g.Expect(lWithRef.KubernetesRef.Version).To(Equal("v1"))
				g.Expect(lGroupOnly).ToNot(BeNil())
				g.Expect(lGroupOnly.KubernetesRef).To(BeNil())
				g.Expect(lWithoutRef).ToNot(BeNil())
				g.Expect(lWithoutRef.KubernetesRef).To(BeNil())
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("returns nodeAffinities in GET workload and list", func() {
			withAffinityId := uuid.NewString()
			noAffinityId := uuid.NewString()

			withAffinity := SetupBaseWorkload(withAffinityId, DefaultProjectId, testing.ClusterUuid1)
			withAffinity.NodeAffinities = []api.WorkloadNodeAffinity{
				{
					SpecName: "worker",
					NodeAffinity: api.PodNodeAffinity{
						Required: &api.PodNodeAffinityRequired{
							NodeSelectorTerms: []api.NodeSelectorTerm{{
								MatchExpressions: []api.MatchExpression{
									{Key: "gpu-type", Operator: api.MATCHEXPRESSIONOPERATOR_IN, Values: []string{"a100", "h100"}},
								},
							}},
						},
					},
				},
				{
					SpecName: "leader",
					NodeAffinity: api.PodNodeAffinity{
						Preferred: &api.PodNodeAffinityPreferred{
							NodeSelectorTerms: []api.PodPreferredNodeSelectorTerm{{
								Weight:     10,
								Preference: &api.NodeSelectorTerm{MatchExpressions: []api.MatchExpression{{Key: "zone", Operator: api.MATCHEXPRESSIONOPERATOR_IN, Values: []string{"us-east-1"}}}},
							}},
						},
					},
				},
			}
			noAffinity := SetupBaseWorkload(noAffinityId, DefaultProjectId, testing.ClusterUuid1)

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{withAffinity, noAffinity},
			})
			Expect(apiResponse.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

			// GET detailed workload with affinities
			apiResponse, _ = workloadsService.GetWorkload(ctx, withAffinityId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.NodeAffinities).To(HaveLen(2))
			Expect(wd.NodeAffinities[0].SpecName).To(Equal("worker"))
			Expect(wd.NodeAffinities[0].NodeAffinity.Required).ToNot(BeNil())
			Expect(wd.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms).To(HaveLen(1))
			Expect(wd.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
			Expect(wd.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values).To(Equal([]string{"a100", "h100"}))
			Expect(wd.NodeAffinities[1].SpecName).To(Equal("leader"))
			Expect(wd.NodeAffinities[1].NodeAffinity.Preferred).ToNot(BeNil())
			Expect(wd.NodeAffinities[1].NodeAffinity.Preferred.NodeSelectorTerms[0].Weight).To(Equal(float32(10)))

			// GET detailed workload without affinities
			apiResponse, _ = workloadsService.GetWorkload(ctx, noAffinityId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			wd = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(wd.NodeAffinities).To(BeEmpty())

			// list endpoint should include nodeAffinities
			waitForGetWorkloadsLength(ctx, workloadsService, 5, []string{})
			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			list := httpResponseBodyToWorkloadsResponse(apiResponse)
			var foundWithAffinity *api.Workload
			var foundNoAffinity *api.Workload
			for i := range list.Workloads {
				if list.Workloads[i].Id == withAffinityId {
					foundWithAffinity = &list.Workloads[i]
				}
				if list.Workloads[i].Id == noAffinityId {
					foundNoAffinity = &list.Workloads[i]
				}
			}
			Expect(foundWithAffinity).ToNot(BeNil())
			Expect(foundWithAffinity.NodeAffinities).To(HaveLen(2))
			Expect(foundWithAffinity.NodeAffinities[0].SpecName).To(Equal("worker"))
			Expect(foundWithAffinity.NodeAffinities[0].NodeAffinity.Required).ToNot(BeNil())
			Expect(foundWithAffinity.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("gpu-type"))
			Expect(foundNoAffinity).ToNot(BeNil())
			Expect(foundNoAffinity.NodeAffinities).To(BeEmpty())
		})

		It("Sanity", func() {
			apiResponse, _ := workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadsResponse, testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
		})

		It("priority", func() {
			apiResponse, _ := workloadsService.GetWorkload(ctx, id1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.Preemptible).To(BeNil())
			apiResponse, _ = workloadsService.GetWorkload(ctx, id2)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(*workloadsResponse.Preemptible).To(BeTrue())
			apiResponse, _ = workloadsService.GetWorkload(ctx, id3)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(*workloadsResponse.Preemptible).To(BeFalse())
		})

		It("SubmittedBy tests", func() {
			newWorkload := SetupBaseWorkload(uuid.NewString(), DefaultProjectId, testing.ClusterUuid1)
			existingWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			existingWorkload.SubmittedBy = ""

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					existingWorkload,
					newWorkload,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, newWorkload.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal(DefaultSubmittedBy))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal(DefaultSubmittedBy))

			existingWorkload.SubmittedBy = "some-other-user"

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					existingWorkload,
					newWorkload,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, newWorkload.Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal(DefaultSubmittedBy))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal("some-other-user"))
		})

		It("Deletion - no To and From", func() {
			// Project 1 workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("projectId==%s", DefaultProjectId)})

			// Validate that deletion in one project does not affect another project
			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", AnotherProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(3))
			Expect(workloadsResponse.Next).To(Equal(int32(0)))
		})

		It("Deletion - filtered of other type", func() {
			// Project 1 workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{"type==someOtherType"}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(3)) // No deletions
			Expect(workloadsResponse.Next).To(Equal(int32(0)))
		})

		It("Deletion - filtered of real type", func() {
			// Project 1 workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{fmt.Sprintf("type==%s", DefaultType)}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// No deletions
			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("projectId==%s", DefaultProjectId)})
		})

		It("Deletion - with To and From", func() {
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
				To:   testing.TestUuid1,
				From: testing.TestUuid1,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
			Expect(len(workloadsResponse.Workloads)).To(Equal(3)) // No deletions
			Expect(workloadsResponse.Next).To(Equal(int32(0)))

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
				},
				From:        testing.TestUuid1,
				To:          testing.TestUuid3,
				IdsToIgnore: []string{testing.TestUuid3},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Should delete only workload with TestUuid2
			waitForGetWorkloadsLength(ctx, workloadsService, 2, []string{fmt.Sprintf("projectId==%s", DefaultProjectId)})
		})

		It("Deletion - with ClusterId filters by cluster", func() {
			// Create workloads in same project but different clusters
			workloadCluster1Id := uuid.NewString()
			workloadCluster2Id := uuid.NewString()
			workloadCluster2Id2 := uuid.NewString()

			// Sync workloads to cluster1
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(workloadCluster1Id, DefaultProjectId, testing.ClusterUuid3),
				},
				ClusterId: testing.ClusterUuid3,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Sync workloads to cluster2
			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(workloadCluster2Id, DefaultProjectId, testing.ClusterUuid4),
					SetupBaseWorkload(workloadCluster2Id2, DefaultProjectId, testing.ClusterUuid4),
				},
				ClusterId: testing.ClusterUuid4,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Validate all 3 workloads exist
			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)})
			waitForGetWorkloadsLength(ctx, workloadsService, 2, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid4)})

			// Sync with cluster2 but only include one workload - should delete only the other cluster2 workload
			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(workloadCluster2Id, DefaultProjectId, testing.ClusterUuid4),
				},
				ClusterId: testing.ClusterUuid4,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// Cluster1 workload should still exist (unaffected)
			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)})
			apiResponse, _ = workloadsService.GetWorkload(ctx, workloadCluster1Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body.(api.WorkloadDetailed).DeletedAt).To(BeNil())

			// Cluster2 should have 1 workload remaining (undeleted)
			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid4)})
			apiResponse, _ = workloadsService.GetWorkload(ctx, workloadCluster2Id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body.(api.WorkloadDetailed).DeletedAt).To(BeNil())

			// The other cluster2 workload should be soft-deleted
			apiResponse, _ = workloadsService.GetWorkload(ctx, workloadCluster2Id2)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(apiResponse.Body.(api.WorkloadDetailed).DeletedAt).ToNot(BeNil())

			// Cleanup
			workloadsService.DeleteWorkload(ctx, workloadCluster1Id)
			workloadsService.DeleteWorkload(ctx, workloadCluster2Id)
			workloadsService.DeleteWorkload(ctx, workloadCluster2Id2)
		})

		It("Wrong Project", func() {
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid1, AnotherProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("Internal Sync Sanity", func() {
			apiResponse, _ := workloadsService.InternalSyncWorkloads(ctx, []string{}, api.InternalWorkloadsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Workloads: []api.WorkloadSyncFields{SetupBaseWorkload(testing.TestUuid4, DefaultProjectId, testing.ClusterUuid3)},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid4)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadsResponse, testing.TestUuid4, DefaultProjectId, testing.ClusterUuid3)
		})

		It("Deletion", func() {
			apiResponse, _ := workloadsService.InternalSyncWorkloads(ctx, []string{}, api.InternalWorkloadsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Workloads: []api.WorkloadSyncFields{SetupBaseWorkload(testing.TestUuid4, DefaultProjectId, testing.ClusterUuid3)},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			waitForGetWorkloadsLength(ctx, workloadsService, 1, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)})

			apiResponse, _ = workloadsService.InternalSyncWorkloads(ctx, []string{}, api.InternalWorkloadsSyncRequest{
				ClusterId: testing.ClusterUuid3,
				Workloads: []api.WorkloadSyncFields{},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			waitForGetWorkloadsLength(ctx, workloadsService, 0, []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid3)})
		})
	})

	Context("Upsert", func() {
		It("Insert", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid4, DefaultProjectId, testing.ClusterUuid1)

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid4, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid4)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid4)
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadsResponse, testing.TestUuid4, DefaultProjectId, testing.ClusterUuid1)

			Expect(workloadsResponse.ActionsSupport.Delete).To(Equal(true))
			Expect(workloadsResponse.ActionsSupport.Suspend).To(Equal(true))
			Expect(workloadsResponse.PhaseReason).To(BeNil())
		})

		It("Update workload phase", func() {
			phaseMsg := "Failed to submit workload"
			apiResponse, _ := workloadsService.PatchWorkloadPhase(ctx, testing.TestUuid1, api.WorkloadPatchFields{Phase: api.PHASE_FAILED, PhaseMessage: phaseMsg})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.Id).To(Equal(testing.TestUuid1))
			Expect(workloadsResponse.Name).To(Equal(DefaultWorkloadName))
			Expect(workloadsResponse.Phase).To(Equal(api.PHASE_FAILED))
			Expect(workloadsResponse.PhaseMessage).To(Equal(phaseMsg))

			// Sleep, so an internal call to "PatchWorkloadPhaseHelper" with "updateOnlyIfLastUpdateSmallerThan" will have a different time.
			time.Sleep(100 * time.Millisecond)

			apiResponse, _ = workloadsService.PatchWorkloadPhase(ctx, testing.TestUuid1, api.WorkloadPatchFields{Phase: api.PHASE_RUNNING})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)

			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			Expect(workloadsResponse.Phase).To(Equal(api.PHASE_RUNNING))
			Expect(workloadsResponse.PhaseMessage).To(Equal(""))
		})

		It("Update workload phase - soft delete", func() {
			phaseMsg := "Workload with no project"

			apiResponse, _ := workloadsService.PatchWorkloadPhase(ctx, testing.TestUuid1, api.WorkloadPatchFields{Phase: api.PHASE_FAILED, PhaseMessage: phaseMsg, SoftDelete: true})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.DeletedAt).ToNot(BeNil())
		})

		It("SubmittedBy tests", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.SubmittedBy = ""

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal(DefaultSubmittedBy))

			baseWorkload.SubmittedBy = "some-other-user"

			apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.SubmittedBy).To(Equal("some-other-user"))
		})

		It("NodeAffinities via UpsertWorkload", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.NodeAffinities = []api.WorkloadNodeAffinity{
				{
					SpecName: "worker",
					NodeAffinity: api.PodNodeAffinity{
						Required: &api.PodNodeAffinityRequired{
							NodeSelectorTerms: []api.NodeSelectorTerm{{
								MatchExpressions: []api.MatchExpression{
									{Key: "nvidia.com/gpu", Operator: api.MATCHEXPRESSIONOPERATOR_EXISTS},
								},
							}},
						},
					},
				},
			}

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.NodeAffinities).To(HaveLen(1))
			Expect(workloadsResponse.NodeAffinities[0].SpecName).To(Equal("worker"))
			Expect(workloadsResponse.NodeAffinities[0].NodeAffinity.Required).ToNot(BeNil())
			Expect(workloadsResponse.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("nvidia.com/gpu"))
			Expect(workloadsResponse.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator).To(Equal(api.MATCHEXPRESSIONOPERATOR_EXISTS))

			// validate list endpoint - use Eventually because the materialized view needs to refresh
			// after the upsert (count doesn't change, so waitForGetWorkloadsLength won't help)
			Eventually(func(g Gomega) {
				apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				list := httpResponseBodyToWorkloadsResponse(apiResponse)
				var found *api.Workload
				for i := range list.Workloads {
					if list.Workloads[i].Id == testing.TestUuid1 {
						found = &list.Workloads[i]
						break
					}
				}
				g.Expect(found).ToNot(BeNil())
				g.Expect(found.NodeAffinities).To(HaveLen(1))
				g.Expect(found.NodeAffinities[0].SpecName).To(Equal("worker"))
				g.Expect(found.NodeAffinities[0].NodeAffinity.Required).ToNot(BeNil())
				g.Expect(found.NodeAffinities[0].NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("nvidia.com/gpu"))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())

			// Update to empty - should clear
			baseWorkload.NodeAffinities = []api.WorkloadNodeAffinity{}
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadsResponse.NodeAffinities).To(BeEmpty())
		})

		It("Conditions", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Conditions = []api.Condition{}

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(len(workloadsResponse.Conditions)).To(Equal(0))

			time := time.Now()
			baseWorkload.Conditions = []api.Condition{
				{
					Type:               "Ready",
					Status:             "False",
					Message:            "Resource validation failed: ...",
					Reason:             "ErrorConfig",
					LastTransitionTime: &time,
				},
				{
					Type:               "NotReady",
					Status:             "True",
					LastTransitionTime: nil,
				},
			}

			err := api.AssertWorkloadSyncFieldsConstraints(baseWorkload)
			Expect(err).To(BeNil())

			apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			workloadsResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(len(workloadsResponse.Conditions)).To(Equal(2))
			Expect(workloadsResponse.Conditions[0].Type).To(Equal(baseWorkload.Conditions[0].Type))
			Expect(workloadsResponse.Conditions[0].Status).To(Equal(baseWorkload.Conditions[0].Status))
			Expect(workloadsResponse.Conditions[0].Message).To(Equal(baseWorkload.Conditions[0].Message))
			Expect(workloadsResponse.Conditions[0].Reason).To(Equal(baseWorkload.Conditions[0].Reason))
			Expect(workloadsResponse.Conditions[0].LastTransitionTime.Unix()).ToNot(BeNil())
			Expect((*workloadsResponse.Conditions[0].LastTransitionTime).Unix()).To(Equal(time.Unix()))
			Expect(workloadsResponse.Conditions[1].Type).To(Equal(baseWorkload.Conditions[1].Type))
			Expect(workloadsResponse.Conditions[1].Status).To(Equal(baseWorkload.Conditions[1].Status))
			Expect(workloadsResponse.Conditions[1].Message).To(Equal(baseWorkload.Conditions[1].Message))
			Expect(workloadsResponse.Conditions[1].Reason).To(Equal(baseWorkload.Conditions[1].Reason))
			Expect(workloadsResponse.Conditions[1].LastTransitionTime).To(BeNil())
		})

		It("Network Status", func() {
			uuid := uuid.NewString()

			baseWorkload := SetupBaseWorkload(uuid, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.NetworkStatus = api.WorkloadNetworkStatus{
				Services: []api.WorkloadNetworkStatusServicesInner{
					{
						Name: "service1",
						Type: "ClusterIP",
						Ports: []api.WorkloadNetworkStatusServicesInnerPortsInner{
							{
								Name:       "port1",
								TargetPort: "80",
								Port:       80,
								Protocol:   "TCP",
							},
							{
								Name:       "port2",
								TargetPort: "90",
								Port:       90,
								Protocol:   "UDP",
							},
						},
						LoadBalancer: api.WorkloadNetworkStatusServicesInnerLoadBalancer{
							Ingress: []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
								{
									Ip:       "192.168.2.1",
									Hostname: "example.com",
								},
								{
									Ip:       "192.168.2.2",
									Hostname: "example2.com",
								},
							},
						},
					},
				},
			}

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, uuid, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			var workload2 *api.Workload
			Eventually(func(g Gomega) {
				apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
				for _, w := range workloadsResponse.Workloads {
					if w.Id == uuid {
						workload2 = &w
						break
					}
				}
				g.Expect(workload2).ToNot(BeNil())
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())

			//TODO: uncomment when network status is supported in the view
			/*
				apiResponse, err := workloadsService.GetWorkload(ctx, uuid)
				Expect(err).To(BeNil())
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workload := httpResponseBodyToWorkloadResponse(apiResponse)

				Expect(workload.NetworkStatus).ToNot(BeNil())
				Expect(workload.NetworkStatus.Services).ToNot(BeNil())
				Expect(len(workload.NetworkStatus.Services)).To(Equal(1))
				Expect(workload.NetworkStatus.Services[0].Name).To(Equal("service1"))
				Expect(workload.NetworkStatus.Services[0].Type).To(Equal("ClusterIP"))
				Expect(workload.NetworkStatus.Services[0].Ports).ToNot(BeNil())
				Expect(len(workload.NetworkStatus.Services[0].Ports)).To(Equal(2))
				Expect(workload.NetworkStatus.Services[0].Ports[0].Name).To(Equal("port1"))
				Expect(workload.NetworkStatus.Services[0].Ports[0].TargetPort).To(Equal("80"))
				Expect(workload.NetworkStatus.Services[0].Ports[0].Port).To(BeEquivalentTo(80))
				Expect(workload.NetworkStatus.Services[0].Ports[0].Protocol).To(Equal("TCP"))
				Expect(workload.NetworkStatus.Services[0].Ports[1].Name).To(Equal("port2"))
				Expect(workload.NetworkStatus.Services[0].Ports[1].TargetPort).To(Equal("90"))
				Expect(workload.NetworkStatus.Services[0].Ports[1].Port).To(BeEquivalentTo(90))
				Expect(workload.NetworkStatus.Services[0].Ports[1].Protocol).To(Equal("UDP"))
				Expect(workload.NetworkStatus.Services[0].LoadBalancer.Ingress).ToNot(BeNil())
				Expect(len(workload.NetworkStatus.Services[0].LoadBalancer.Ingress)).To(Equal(2))
				Expect(workload.NetworkStatus.Services[0].LoadBalancer.Ingress[0].Ip).To(Equal("192.168.2.1"))
				Expect(workload.NetworkStatus.Services[0].LoadBalancer.Ingress[0].Hostname).To(Equal("example.com"))
				Expect(workload.NetworkStatus.Services[0].LoadBalancer.Ingress[1].Ip).To(Equal("192.168.2.2"))
				Expect(workload.NetworkStatus.Services[0].LoadBalancer.Ingress[1].Hostname).To(Equal("example2.com"))

				Expect(workload2.NetworkStatus).ToNot(BeNil())
				Expect(workload2.NetworkStatus.Services).ToNot(BeNil())
				Expect(len(workload2.NetworkStatus.Services)).To(Equal(1))
				Expect(workload2.NetworkStatus.Services[0].Name).To(Equal("service1"))
				Expect(workload2.NetworkStatus.Services[0].Type).To(Equal("ClusterIP"))
				Expect(workload2.NetworkStatus.Services[0].Ports).ToNot(BeNil())
				Expect(len(workload2.NetworkStatus.Services[0].Ports)).To(Equal(2))
				Expect(workload2.NetworkStatus.Services[0].Ports[0].Name).To(Equal("port1"))
				Expect(workload2.NetworkStatus.Services[0].Ports[0].TargetPort).To(Equal("80"))
				Expect(workload2.NetworkStatus.Services[0].Ports[0].Port).To(BeEquivalentTo(80))
				Expect(workload2.NetworkStatus.Services[0].Ports[0].Protocol).To(Equal("TCP"))
				Expect(workload2.NetworkStatus.Services[0].Ports[1].Name).To(Equal("port2"))
				Expect(workload2.NetworkStatus.Services[0].Ports[1].TargetPort).To(Equal("90"))
				Expect(workload2.NetworkStatus.Services[0].Ports[1].Port).To(BeEquivalentTo(90))
				Expect(workload2.NetworkStatus.Services[0].Ports[1].Protocol).To(Equal("UDP"))
				Expect(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress).ToNot(BeNil())
				Expect(len(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress)).To(Equal(2))
				Expect(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress[0].Ip).To(Equal("192.168.2.1"))
				Expect(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress[0].Hostname).To(Equal("example.com"))
				Expect(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress[1].Ip).To(Equal("192.168.2.2"))
				Expect(workload2.NetworkStatus.Services[0].LoadBalancer.Ingress[1].Hostname).To(Equal("example2.com"))
			*/
		})

		It("Pending Scheduling Reason Message - GPU", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Phase = api.PHASE_PENDING
			baseWorkload.SchedulingConditions = ptr.To([]api.SchedulingCondition{
				{
					NodePool: "default",
					Reason:   api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_PROJECT,
						ExceedingResources: api.QueueResources{
							Gpu:    2,
							Cpu:    1,
							Memory: 1,
						},
						PodGroupRequestedNonPreemptibleResources: api.QueueResources{
							Gpu:    2,
							Cpu:    1,
							Memory: 1000,
						},
						QueueDeservedResources: api.QueueResources{
							Gpu:    4,
							Cpu:    1,
							Memory: 1000,
						},
						QueueAllocatedNonPreemptibleResources: api.QueueResources{
							Gpu:    3,
							Cpu:    1,
							Memory: 1000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    1,
							Cpu:    1,
							Memory: 1000,
						},
					},
				},
				{
					NodePool: "nodepool-a",
					Reason:   api.PHASEREASON_OVER_LIMIT,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_DEPARTMENT,
						ExceedingResources: api.QueueResources{
							Gpu:    2,
							Cpu:    0,
							Memory: 0,
						},
						PodGroupRequestedNonPreemptibleResources: api.QueueResources{
							Gpu:    3,
							Cpu:    1,
							Memory: 1000,
						},
						PodGroupRequestedResources: api.QueueResources{
							Gpu:    3,
							Cpu:    1,
							Memory: 1000,
						},
						QueueResourceLimits: api.QueueResources{
							Gpu:    3,
							Cpu:    1,
							Memory: 1000,
						},
						QueueAllocatedResources: api.QueueResources{
							Gpu:    2,
							Cpu:    1,
							Memory: 1000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    1,
							Cpu:    1,
							Memory: 1000,
						},
					},
				},
			})

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.PhaseReason).ToNot(BeNil())
			Expect(*workloadResponse.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(workloadResponse.PendingSchedulingMessages).ToNot(BeNil())
			Expect(len(*workloadResponse.PendingSchedulingMessages)).To(Equal(2))
			pm := *workloadResponse.PendingSchedulingMessages

			Expect(pm[0].PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(pm[0].NodePool).To(Equal("default"))
			Expect(pm[0].Reason).To(Equal(NonPreemtibleOverQuotaUserReason))
			Expect(pm[0].OrgType).ToNot(BeNil())
			Expect(*pm[0].OrgType).To(Equal(api.ORGTYPE_PROJECT))
			Expect(pm[0].UserMessage).ToNot(BeNil())
			Expect(*pm[0].UserMessage).To(Equal("GPU quota exceeded for non-preemptible workloads.\nRequested: 2 devices\nProject quota: 4 devices, with 3 devices already allocated for non-preemptible"))

			Expect(pm[1].PhaseReason).To(Equal(api.PHASEREASON_OVER_LIMIT))
			Expect(pm[1].NodePool).To(Equal("nodepool-a"))
			Expect(pm[1].Reason).To(Equal(fmt.Sprintf(OverLimitUserReason, "department")))
			Expect(pm[1].OrgType).ToNot(BeNil())
			Expect(*pm[1].OrgType).To(Equal(api.ORGTYPE_DEPARTMENT))
			Expect(pm[1].UserMessage).ToNot(BeNil())
			Expect(*pm[1].UserMessage).To(Equal("GPU quota exceeded.\nRequested: 3 devices\nDepartment quota: 3 devices, with 2 devices already allocated"))

			Eventually(func(g Gomega) {
				apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
				var testWorkload *api.Workload = nil
				for _, w := range workloadsResponse.Workloads {
					if w.Id == testing.TestUuid1 {
						testWorkload = &w
						break
					}
				}
				g.Expect(testWorkload).ToNot(BeNil())
				g.Expect(testWorkload.PhaseReason).ToNot(BeNil())
				g.Expect(*testWorkload.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("Pending Scheduling Reason Message - CPU", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Phase = api.PHASE_PENDING
			baseWorkload.SchedulingConditions = ptr.To([]api.SchedulingCondition{
				{
					NodePool: "default",
					Reason:   api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_PROJECT,
						ExceedingResources: api.QueueResources{
							Gpu:    0,
							Cpu:    1,
							Memory: 1,
						},
						PodGroupRequestedNonPreemptibleResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000,
						},
						QueueDeservedResources: api.QueueResources{
							Gpu:    4,
							Cpu:    4,
							Memory: 4000,
						},
						QueueAllocatedNonPreemptibleResources: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    1,
							Cpu:    1,
							Memory: 1000,
						},
					},
				},
				{
					NodePool: "nodepool-a",
					Reason:   api.PHASEREASON_OVER_LIMIT,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_DEPARTMENT,
						ExceedingResources: api.QueueResources{
							Gpu:    0,
							Cpu:    1,
							Memory: 1,
						},
						PodGroupRequestedNonPreemptibleResources: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000,
						},
						PodGroupRequestedResources: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000,
						},
						QueueResourceLimits: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000,
						},
						QueueAllocatedResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000,
						},
					},
				},
			})

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.PhaseReason).ToNot(BeNil())
			Expect(*workloadResponse.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(workloadResponse.PendingSchedulingMessages).ToNot(BeNil())
			Expect(len(*workloadResponse.PendingSchedulingMessages)).To(Equal(2))
			pm := *workloadResponse.PendingSchedulingMessages

			Expect(pm[0].PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(pm[0].NodePool).To(Equal("default"))
			Expect(pm[0].Reason).To(Equal(NonPreemtibleOverQuotaUserReason))
			Expect(pm[0].OrgType).ToNot(BeNil())
			Expect(*pm[0].OrgType).To(Equal(api.ORGTYPE_PROJECT))
			Expect(pm[0].UserMessage).ToNot(BeNil())
			Expect(*pm[0].UserMessage).To(Equal("CPU quota exceeded for non-preemptible workloads.\nRequested: 2 cores\nProject quota: 4 cores, with 3 cores already allocated for non-preemptible"))

			Expect(pm[1].PhaseReason).To(Equal(api.PHASEREASON_OVER_LIMIT))
			Expect(pm[1].NodePool).To(Equal("nodepool-a"))
			Expect(pm[1].Reason).To(Equal(fmt.Sprintf(OverLimitUserReason, "department")))
			Expect(pm[1].OrgType).ToNot(BeNil())
			Expect(*pm[1].OrgType).To(Equal(api.ORGTYPE_DEPARTMENT))
			Expect(pm[1].UserMessage).ToNot(BeNil())
			Expect(*pm[1].UserMessage).To(Equal("CPU quota exceeded.\nRequested: 3 cores\nDepartment quota: 3 cores, with 2 cores already allocated"))

			Eventually(func(g Gomega) {
				apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
				var testWorkload *api.Workload = nil
				for _, w := range workloadsResponse.Workloads {
					if w.Id == testing.TestUuid1 {
						testWorkload = &w
						break
					}
				}
				g.Expect(testWorkload).ToNot(BeNil())
				g.Expect(testWorkload.PhaseReason).ToNot(BeNil())
				g.Expect(*testWorkload.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})

		It("Pending Scheduling Reason Message - CPU Memory", func() {
			baseWorkload := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Phase = api.PHASE_PENDING
			baseWorkload.SchedulingConditions = ptr.To([]api.SchedulingCondition{
				{
					NodePool: "default",
					Reason:   api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_PROJECT,
						ExceedingResources: api.QueueResources{
							Gpu:    0,
							Cpu:    0,
							Memory: 1,
						},
						PodGroupRequestedNonPreemptibleResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000000,
						},
						QueueDeservedResources: api.QueueResources{
							Gpu:    4,
							Cpu:    4,
							Memory: 4000000,
						},
						QueueAllocatedNonPreemptibleResources: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    1,
							Cpu:    1,
							Memory: 1000000,
						},
					},
				},
				{
					NodePool: "nodepool-a",
					Reason:   api.PHASEREASON_OVER_LIMIT,
					QueueDetails: &api.SchedulingConditionQueueDetails{
						OrgType: api.ORGTYPE_DEPARTMENT,
						ExceedingResources: api.QueueResources{
							Gpu:    0,
							Cpu:    0,
							Memory: 1,
						},
						PodGroupRequestedResources: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000000,
						},
						QueueResourceLimits: api.QueueResources{
							Gpu:    3,
							Cpu:    3,
							Memory: 3000000,
						},
						QueueAllocatedResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000000,
						},
						QueueRequestedResources: api.QueueResources{
							Gpu:    2,
							Cpu:    2,
							Memory: 2000000,
						},
					},
				},
			})

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.PhaseReason).ToNot(BeNil())
			Expect(*workloadResponse.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(workloadResponse.PendingSchedulingMessages).ToNot(BeNil())
			Expect(len(*workloadResponse.PendingSchedulingMessages)).To(Equal(2))
			pm := *workloadResponse.PendingSchedulingMessages

			Expect(pm[0].PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			Expect(pm[0].NodePool).To(Equal("default"))
			Expect(pm[0].Reason).To(Equal(NonPreemtibleOverQuotaUserReason))
			Expect(pm[0].OrgType).ToNot(BeNil())
			Expect(*pm[0].OrgType).To(Equal(api.ORGTYPE_PROJECT))
			Expect(pm[0].UserMessage).ToNot(BeNil())
			Expect(*pm[0].UserMessage).To(Equal("CPU memory quota exceeded for non-preemptible workloads.\nRequested: 2 MB\nProject quota: 4 MB, with 3 MB already allocated for non-preemptible"))

			Expect(pm[1].PhaseReason).To(Equal(api.PHASEREASON_OVER_LIMIT))
			Expect(pm[1].NodePool).To(Equal("nodepool-a"))
			Expect(pm[1].Reason).To(Equal(fmt.Sprintf(OverLimitUserReason, "department")))
			Expect(pm[1].OrgType).ToNot(BeNil())
			Expect(*pm[1].OrgType).To(Equal(api.ORGTYPE_DEPARTMENT))
			Expect(pm[1].UserMessage).ToNot(BeNil())
			Expect(*pm[1].UserMessage).To(Equal("CPU memory quota exceeded.\nRequested: 3 MB\nDepartment quota: 3 MB, with 2 MB already allocated"))

			Eventually(func(g Gomega) {
				apiResponse, _ = workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", []string{}, "")
				g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
				var testWorkload *api.Workload = nil
				for _, w := range workloadsResponse.Workloads {
					if w.Id == testing.TestUuid1 {
						testWorkload = &w
						break
					}
				}
				g.Expect(testWorkload).ToNot(BeNil())
				g.Expect(testWorkload.PhaseReason).ToNot(BeNil())
				g.Expect(*testWorkload.PhaseReason).To(Equal(api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA))
			}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
		})
	})

	Context("Gets", func() {
		It("Filter - Category", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("category==%s", testing.TrainCategory)}, 6)
		})
		It("Filter - Cluster id", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid1)}, 3)
		})
		It("Filter - Project id", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, 3)
		})
		It("Filter - Nodepool", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("nodepool==%s", DefaultNodepool)}, 1)
		})
		It("Filter - Nodepool no result", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("nodepool==%s", "some-other-not-existing-np")}, 0)
		})
		It("Filter - Workload type", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("type==%s", DefaultType)}, 6)
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{"type==some-other-type"}, 0)
		})
		It("Filter - distributed framework", func() {
			testWorkloadsFilter(ctx, workloadsService, "distributedFramework", []string{fmt.Sprintf("distributedFramework==%s", PytorchFramework)}, 1)
		})
		It("Filter - Multi filters", func() {
			testWorkloadsFilter(ctx, workloadsService, "completedAt", []string{fmt.Sprintf("type==%s", DefaultType), "projectId!=452", "projectName=@tes"}, 6)
		})
		It("Filter - priority", func() {
			testWorkloadsFilter(ctx, workloadsService, "priority", []string{"priority==50"}, 4)
		})
		It("Filter - Phase sub sort", func() {
			workloads := testWorkloadsFilter(ctx, workloadsService, "phase", []string{"phase!=failed"}, 6)
			lastTime := time.Time{}
			var lastPhase api.Phase
			for _, workload := range workloads.Workloads {
				if lastPhase != workload.Phase {
					lastTime = workload.PhaseUpdatedAt
					lastPhase = workload.Phase
				} else {
					Expect(workload.PhaseUpdatedAt.Unix()).To(BeNumerically(">=", lastTime.Unix()))
				}
			}
		})
		It("Pagination", func() {
			offsets := []int32{0, 1, 4, 6}
			limits := []int32{3, 2, 3, 999}

			for i := 0; i < 4; i++ {
				apiResponse, _ := workloadsService.GetWorkloads(ctx, false, offsets[i], limits[i], "", "", []string{}, "")
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
				Expect(len(workloadsResponse.Workloads)).To(Equal(int(math.Min(float64(limits[i]), float64(6-offsets[i])))))
				nextRecord := offsets[i] + int32(len(workloadsResponse.Workloads))
				if nextRecord >= 6 {
					nextRecord = 0
				}

				Expect(workloadsResponse.Next).To(Equal(nextRecord))
			}
		})
		Context("Slim", func() {
			slimFilter := []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid1), fmt.Sprintf("name==%s", DefaultWorkloadName)}

			getSlimWorkloads := func(filterBy []string) []api.WorkloadSlim {
				apiResponse, _ := workloadsService.GetWorkloadsSlim(ctx, 0, 0, filterBy)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				bytes, err := json.Marshal(apiResponse.Body)
				Expect(err).To(BeNil())
				response := api.GetWorkloadsSlimResponse{}
				Expect(json.Unmarshal(bytes, &response)).To(Succeed())
				return response.Workloads
			}

			It("fills the identity fields", func() {
				workloads := getSlimWorkloads(slimFilter)
				Expect(workloads).To(HaveLen(3))
				for _, workload := range workloads {
					Expect(workload.Name).To(Equal(DefaultWorkloadName))
					Expect(workload.Type).To(Equal(DefaultType))
					Expect(workload.ProjectName).To(Equal(DefaultProjectName))
					Expect(workload.ClusterId).To(Equal(testing.ClusterUuid1))
					Expect(workload.DistributedFramework).To(BeEmpty())
				}
			})
			It("sorts by id ascending", func() {
				ids := []string{}
				for _, workload := range getSlimWorkloads(slimFilter) {
					ids = append(ids, workload.Id)
				}
				Expect(ids).To(Equal([]string{testing.TestUuid1, testing.TestUuid2, testing.TestUuid3}))
			})
			It("filters by distributed framework", func() {
				Expect(getSlimWorkloads(append(slimFilter, "distributedFramework!=PyTorchJob"))).To(HaveLen(3))
				Expect(getSlimWorkloads(append(slimFilter, "distributedFramework==PyTorchJob"))).To(BeEmpty())
			})
			It("returns a workload as soon as it's created", func() {
				newWorkload := SetupBaseWorkload(testing.TestUuid7, DefaultProjectId, testing.ClusterUuid1)
				newWorkload.Name = "slim-new-workload"
				apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
					ProjectId: DefaultProjectId,
					Workloads: []api.WorkloadSyncFields{
						newWorkload,
						SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
						SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
						SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1),
					},
				})
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
				createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid7)

				workloads := getSlimWorkloads([]string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid1), "name==slim-new-workload"})
				Expect(workloads).To(HaveLen(1))
				Expect(workloads[0].Id).To(Equal(testing.TestUuid7))
			})
			It("doesn't return deleted workloads", func() {
				_, err := runaidb.DB().Table(dao.WorkloadsTableName).Where("id = ?", testing.TestUuid3).Update("deleted_at", time.Now())
				Expect(err).To(BeNil())
				DeferCleanup(func() {
					_, err := runaidb.DB().Table(dao.WorkloadsTableName).Where("id = ?", testing.TestUuid3).Update("deleted_at", nil)
					Expect(err).To(BeNil())
				})

				workloads := getSlimWorkloads(slimFilter)
				Expect(workloads).To(HaveLen(2))
				for _, workload := range workloads {
					Expect(workload.Id).ToNot(Equal(testing.TestUuid3))
				}
			})
			It("rejects filters on fields that aren't slim fields", func() {
				apiResponse, _ := workloadsService.GetWorkloadsSlim(ctx, 0, 0, []string{"phase==Running"})
				Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Describe("Idle GPUs", func() {
			BeforeEach(func() {
				_, err := runaidb.DB().Table(dao.WorkloadsTableName).Where("id = ?", id3).
					Updates(map[string]interface{}{
						"idle_allocated_gpus": 11,
					})
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Table(dao.WorkloadsTableName).Where("id = ?", id1).
					Updates(map[string]interface{}{
						"idle_allocated_gpus": 3.8,
					})
				Expect(err).To(BeNil())
			})
			It("Filter - idleAllocatedGpus", func() {
				workloads := testWorkloadsFilter(ctx, workloadsService, "idleAllocatedGpus", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid2), "idleAllocatedGpus<=11"}, 2).Workloads
				Expect(workloads[0].Id).To(Equal(id1))
				Expect(*workloads[0].IdleAllocatedGpus).To(Equal(float32(3.8)))
				Expect(workloads[1].Id).To(Equal(id3))
				Expect(*workloads[1].IdleAllocatedGpus).To(Equal(float32(11)))

				workloads = testWorkloadsFilter(ctx, workloadsService, "idleAllocatedGpus", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid2), "idleAllocatedGpus>=4"}, 1).Workloads
				Expect(workloads[0].Id).To(Equal(id3))
				Expect(*workloads[0].IdleAllocatedGpus).To(Equal(float32(11)))

				workloads = testWorkloadsFilter(ctx, workloadsService, "idleAllocatedGpus", []string{fmt.Sprintf("clusterId==%s", testing.ClusterUuid2)}, 3).Workloads
				Expect(workloads[0].Id).To(Equal(id2))
				Expect(workloads[0].IdleAllocatedGpus).To(BeNil())
				Expect(workloads[1].Id).To(Equal(id1))
				Expect(*workloads[1].IdleAllocatedGpus).To(Equal(float32(3.8)))
				Expect(workloads[2].Id).To(Equal(id3))
				Expect(*workloads[2].IdleAllocatedGpus).To(Equal(float32(11)))
			})
		})
		Context("test search GetWorkloads", func() {
			BeforeEach(func() {
				w0 := SetupBaseWorkload(testing.TestUuid7, DefaultProjectId, testing.ClusterUuid1)
				w0.Name = "workload-f56bac34590dfe12"
				apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
					ProjectId: DefaultProjectId,
					Workloads: []api.WorkloadSyncFields{
						w0,
						SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1),
						SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
						SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1),
					},
				})
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
				waitForGetWorkloadsLength(ctx, workloadsService, 7, []string{})
				createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid7)
			})
			It("Search - by name", func() {
				testWorkloadsSearch(ctx, workloadsService, DefaultWorkloadName, 6)
			})
			It("Search - by partial name", func() {
				testWorkloadsSearch(ctx, workloadsService, DefaultWorkloadName[:3], 6)
			})
			It("Search - by partial name 2", func() {
				testWorkloadsSearch(ctx, workloadsService, "work", 1)
				testWorkloadsSearch(ctx, workloadsService, "workload-", 1)
				testWorkloadsSearch(ctx, workloadsService, "f56", 1)

				// searching for "dfe" should return 0 results -
				// because it is a string from the middle of the workload name,
				// and our search uses PostgreSQL full-text search,
				// which matches full words or the beginning of words only.
				// searching for "f56" works because it appears right after an "-"
				// and the search knows to match it as well.
				testWorkloadsSearch(ctx, workloadsService, "dfe", 0)
			})
			It("Search - by project name", func() {
				testWorkloadsSearch(ctx, workloadsService, DefaultProjectName[:2], 7)
			})
			It("Search - by department name", func() {
				testWorkloadsSearch(ctx, workloadsService, DefaultDepartmentName[:3], 7)
			})
			It("Search - by priority", func() {
				testWorkloadsSearch(ctx, workloadsService, "trai", 7)
			})
		})
	})

	Context("Delete", func() {
		It("Sanity", func() {
			apiResponse, _ := workloadsService.DeleteWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, testing.TestUuid1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			waitForGetWorkloadsLength(ctx, workloadsService, 2, []string{fmt.Sprintf("projectId==%s", DefaultProjectId)})

			// it exists in history
			apiResponse, _ = workloadsService.GetWorkloads(ctx, true, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
			found := false
			for _, workload := range workloadsResponse.Workloads {
				found = found || workload.Id == testing.TestUuid1
			}
			Expect(found).To(BeTrue())
		})

		It("deletes the V2 workload snapshot when cluster-sync deletes the workload", func() {
			workloadId := uuid.New()
			workloadIdStr := workloadId.String()

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(workloadIdStr, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			snapshot := &models_crud.WorkloadV2Snapshot{
				ID:            workloadId,
				CreatedAt:     time.Now(),
				CreatedBy:     testing.TestUser,
				UpdatedAt:     time.Now(),
				UpdatedBy:     testing.TestUser,
				TenantId:      testing.TestTenantId,
				ClusterId:     testing.ClusterUuid1,
				DepartmentId:  DefaultDepartmentId,
				ProjectId:     DefaultProjectId,
				WorkloadType:  "Training",
				RequestedName: "v2-workload-snapshot",
				Name:          "v2-workload-snapshot",
				Spec: &workloadsApi.SupersetSpec{
					Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				},
			}
			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			apiResponse, _ = workloadsService.DeleteWorkload(ctx, workloadIdStr)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			_, snapshotErr := dao_crud.DbGetWorkloadV2Snapshot(ctx, testing.TestTenantId, "Training", workloadIdStr, nil)
			Expect(snapshotErr).NotTo(BeNil())
			Expect(snapshotErr.Code).To(Equal(int32(http.StatusNotFound)))
		})

		It("deletes V2 workload snapshots for workloads removed by cluster-sync full sync", func() {
			workloadId, parseErr := uuid.Parse(testing.TestUuid1)
			Expect(parseErr).To(BeNil())

			snapshot := &models_crud.WorkloadV2Snapshot{
				ID:            workloadId,
				CreatedAt:     time.Now(),
				CreatedBy:     testing.TestUser,
				UpdatedAt:     time.Now(),
				UpdatedBy:     testing.TestUser,
				TenantId:      testing.TestTenantId,
				ClusterId:     testing.ClusterUuid1,
				DepartmentId:  DefaultDepartmentId,
				ProjectId:     DefaultProjectId,
				WorkloadType:  DefaultType,
				RequestedName: "v2-workload-snapshot",
				Name:          "v2-workload-snapshot",
				Spec: &workloadsApi.SupersetSpec{
					Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
				},
			}
			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				ClusterId: testing.ClusterUuid1,
				Workloads: []api.WorkloadSyncFields{
					SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, testing.ClusterUuid1),
					SetupBaseWorkload(testing.TestUuid3, DefaultProjectId, testing.ClusterUuid1),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			_, snapshotErr := dao_crud.DbGetWorkloadV2Snapshot(ctx, testing.TestTenantId, DefaultType, testing.TestUuid1, nil)
			Expect(snapshotErr).NotTo(BeNil())
			Expect(snapshotErr.Code).To(Equal(int32(http.StatusNotFound)))
		})

	})

	Context("Internal API workload creation", func() {
		It("Create and sync", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			createdWorkloadsIds = append(createdWorkloadsIds, id)
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.ProjectId).To(Equal(DefaultProjectId))
			Expect(workloadResponse.Datasources).To(Equal(workload.Datasources))
			Expect(workloadResponse.Environments).To(Equal(workload.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(workload.ExternalConnections))
			Expect(workloadResponse.DistributedFramework).To(Equal(workload.DistributedFramework))
			Expect(workloadResponse.Source).To(Equal(api.SOURCE_CONTROL_PLANE))
			Expect(workloadResponse.Phase).To(Equal(api.PHASE_CREATING))
			Expect(workloadResponse.PhaseMessage).To(Equal(ControlPlaneCreatingMessage))
			Expect(workloadResponse.ActionsSupport.Delete).To(Equal(true))
			// This is a WorkloadsV2 internal-create record using the legacy "Training" placeholder type.
			// No matching Karta is inserted for that type in this test, so V2 suspend is not supported.
			Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(false))
			Expect(workloadResponse.CreatedAt.IsZero()).To(BeFalse())
			Expect(workloadResponse.SourceApi).To(Not(BeNil()))
			Expect(*workloadResponse.SourceApi).To(Equal(api.SOURCEAPI_WORKLOADS_V2))

			// keep CreatedAt creation time
			createdAtTs := workloadResponse.CreatedAt

			// Validate that after cluster syncing the assets data is remains
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.DistributedFramework = "pytorch"

			// Modify CreatedAt timestamp
			baseWorkload.CreatedAt = createdAtTs.Add(time.Minute)
			apiResponse, _ = workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadResponse, id, DefaultProjectId, testing.ClusterUuid1)
			Expect(workloadResponse.ProjectName).To(Equal(DefaultProjectName))
			Expect(workloadResponse.Datasources).To(Equal(workload.Datasources))
			Expect(workloadResponse.Environments).To(Equal(workload.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(workload.ExternalConnections))
			Expect(workloadResponse.DistributedFramework).To(Equal(workload.DistributedFramework))

			// created at is keeping the creation timestamp
			Expect(workloadResponse.CreatedAt).To(Equal(createdAtTs))
		})

		It("sets ActionsSupport.Suspend=true when workload type has a Karta with suspendDefinition and cluster version >= 2.26", func() {
			typeName := "SuspendableJob-" + rand.String(5)

			// Create the workload type in the DB
			workloadType, apiErr := dao.DbUpsertWorkloadType(ctx, runaidb.DB(), dao.WorkloadTypeCreateFieldsParsed{
				WorkloadTypeCreateUpdateFieldsParsed: dao.WorkloadTypeCreateUpdateFieldsParsed{
					WorkloadTypeName: typeName,
				},
				Group: "batch",
			})
			Expect(apiErr).To(BeNil())
			createWorkloadTypeConfigForKartaLookup(workloadType.ID, string(api.SCOPETYPE_SYSTEM), nil)
			// Directly insert the Karta row (DbUpsertWorkloadType only creates the workload_types row)
			kartaRecord := models.WorkloadTypeKarta{
				ScopeType:      string(api.SCOPETYPE_SYSTEM),
				WorkloadTypeId: workloadType.ID,
				Version:        "v1",
				Karta:          buildTestKartaWithSuspend(typeName),
			}
			_, err := runaidb.DB().Create(&kartaRecord)
			Expect(err).To(BeNil())

			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id) // uses ClusterUuid1 (version 9.9.9 >= 2.26)
			workload.Type = typeName
			workload.Group = "batch"
			workload.Version = "v1"

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.ActionsSupport.Delete).To(Equal(true))
			Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(true))
		})

		It("returns the tenant-scoped Karta when the same GVK exists in tenant and system scopes", func() {
			typeName := "ScopedSuspendableJob-" + rand.String(5)

			workloadType, apiErr := dao.DbUpsertWorkloadType(ctx, runaidb.DB(), dao.WorkloadTypeCreateFieldsParsed{
				WorkloadTypeCreateUpdateFieldsParsed: dao.WorkloadTypeCreateUpdateFieldsParsed{
					WorkloadTypeName: typeName,
				},
				Group: "batch",
			})
			Expect(apiErr).To(BeNil())
			createWorkloadTypeConfigForKartaLookup(workloadType.ID, string(api.SCOPETYPE_SYSTEM), nil)
			createWorkloadTypeConfigForKartaLookup(workloadType.ID, string(api.SCOPETYPE_TENANT), ptr.To(testing.TestTenant))

			systemKarta := models.WorkloadTypeKarta{
				ScopeType:      string(api.SCOPETYPE_SYSTEM),
				WorkloadTypeId: workloadType.ID,
				Version:        "v1",
				Karta:          buildTestKartaWithSuspend(typeName),
			}
			_, err := runaidb.DB().Create(&systemKarta)
			Expect(err).To(BeNil())

			tenantKarta := models.WorkloadTypeKarta{
				TenantId:       ptr.To(testing.TestTenant),
				ScopeType:      string(api.SCOPETYPE_TENANT),
				WorkloadTypeId: workloadType.ID,
				Version:        "v1",
				Karta:          buildTestKartaWithSuspend(typeName),
			}
			_, err = runaidb.DB().Create(&tenantKarta)
			Expect(err).To(BeNil())

			karta, apiErr := dao.DbGetKartaByGVK(ctx, testing.TestTenantId, typeName, "batch", "v1")
			Expect(apiErr).To(BeNil())
			Expect(karta).ToNot(BeNil())
			Expect(karta.ID).To(Equal(tenantKarta.ID))
			Expect(karta.ScopeType).To(Equal(string(api.SCOPETYPE_TENANT)))
			Expect(karta.TenantId).ToNot(BeNil())
			Expect(*karta.TenantId).To(Equal(testing.TestTenant))

		})

		It("sets ActionsSupport.Suspend=false when cluster version < 2.26 even if Karta has suspendDefinition", func() {
			typeName := "SuspendableJob-" + rand.String(5)

			workloadType, apiErr := dao.DbUpsertWorkloadType(ctx, runaidb.DB(), dao.WorkloadTypeCreateFieldsParsed{
				WorkloadTypeCreateUpdateFieldsParsed: dao.WorkloadTypeCreateUpdateFieldsParsed{
					WorkloadTypeName: typeName,
				},
				Group: "batch",
			})
			Expect(apiErr).To(BeNil())
			createWorkloadTypeConfigForKartaLookup(workloadType.ID, string(api.SCOPETYPE_SYSTEM), nil)
			kartaRecord := models.WorkloadTypeKarta{
				ScopeType:      string(api.SCOPETYPE_SYSTEM),
				WorkloadTypeId: workloadType.ID,
				Version:        "v1",
				Karta:          buildTestKartaWithSuspend(typeName),
			}
			_, err := runaidb.DB().Create(&kartaRecord)
			Expect(err).To(BeNil())

			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = typeName
			workload.Group = "batch"
			workload.Version = "v1"
			workload.ClusterId = testing.ClusterUuid2 // version "2.22.0" < 2.26

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(false))
		})

		It("Test race condition", func() {
			id := uuid.NewString()

			// Validate that if cluster is synced before, we updates only the CP fields
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.DistributedFramework = "pytorch"
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			workload := setupInternalWorkloadCreationFields(id)

			apiResponse, _ = workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadResponse, id, DefaultProjectId, testing.ClusterUuid1)
			Expect(workloadResponse.Phase).ToNot(Equal(api.PHASE_CREATING))
			Expect(workloadResponse.PhaseMessage).ToNot(Equal(ControlPlaneCreatingMessage))
			Expect(workloadResponse.Datasources).To(Equal(workload.Datasources))
			Expect(workloadResponse.Environments).To(Equal(workload.Environments))
			Expect(workloadResponse.DistributedFramework).To(Equal(workload.DistributedFramework))
		})

		It("preserves creation fields when connection metadata is updated after internal create", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			connectionMetadata := setupWorkloadConnectionMetadata("after-create")

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiErr := internalWorkloadsService.UpdateWorkloadConnectionMetadata(ctx, id, connectionMetadata)
			Expect(apiErr).To(BeNil())

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.ProjectId).To(Equal(workload.ProjectId))
			Expect(workloadResponse.Datasources).To(Equal(workload.Datasources))
			Expect(workloadResponse.Environments).To(Equal(connectionMetadata.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(connectionMetadata.ExternalConnections))
			Expect(workloadResponse.SourceApi).To(Not(BeNil()))
			Expect(*workloadResponse.SourceApi).To(Equal(api.SOURCEAPI_WORKLOADS_V2))
			Expect(workloadResponse.Phase).To(Equal(api.PHASE_CREATING))
			Expect(workloadResponse.PhaseMessage).To(Equal(ControlPlaneCreatingMessage))
		})

		It("preserves connection metadata when internal create notification arrives after metadata update", func() {
			id := uuid.NewString()
			connectionMetadata := setupWorkloadConnectionMetadata("before-create")

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiErr := internalWorkloadsService.UpdateWorkloadConnectionMetadata(ctx, id, connectionMetadata)
			Expect(apiErr).To(BeNil())

			workload := setupInternalWorkloadCreationFields(id)
			workload.Datasources = nil
			workload.Environments = nil
			workload.ExternalConnections = nil

			apiResponse, _ = workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			ValidateBaseWorkload(workloadResponse, id, DefaultProjectId, testing.ClusterUuid1)
			Expect(workloadResponse.Environments).To(Equal(connectionMetadata.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(connectionMetadata.ExternalConnections))
			Expect(workloadResponse.SourceApi).To(Not(BeNil()))
			Expect(*workloadResponse.SourceApi).To(Equal(api.SOURCEAPI_WORKLOADS_V2))
			Expect(workloadResponse.Phase).ToNot(Equal(api.PHASE_CREATING))
			Expect(workloadResponse.PhaseMessage).ToNot(Equal(ControlPlaneCreatingMessage))
		})

		It("creates a fallback row when connection metadata arrives before the internal create notification", func() {
			id := uuid.NewString()
			connectionMetadata := setupWorkloadConnectionMetadata("fallback")
			workload := setupInternalWorkloadCreationFields(id)
			workload.Environments = nil
			workload.ExternalConnections = nil

			apiErr := internalWorkloadsService.EnsureWorkloadConnectionMetadata(ctx, workload, connectionMetadata)
			Expect(apiErr).To(BeNil())
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiResponse, _ := workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.Environments).To(Equal(connectionMetadata.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(connectionMetadata.ExternalConnections))
			Expect(workloadResponse.SourceApi).To(Not(BeNil()))
			Expect(*workloadResponse.SourceApi).To(Equal(api.SOURCEAPI_WORKLOADS_V2))

			// Simulate workloads-manager's delayed creation notification, which
			// does not include workloads-service-specific connection metadata.
			apiResponse, _ = workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse = httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.Environments).To(Equal(connectionMetadata.Environments))
			Expect(workloadResponse.ExternalConnections).To(Equal(connectionMetadata.ExternalConnections))
			Expect(workloadResponse.ProjectId).To(Equal(workload.ProjectId))
		})

		It("sets ActionsSupport.Suspend=true for V2 Job even when cluster sync created the row first", func() {
			id := uuid.NewString()

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Type = "Job"
			baseWorkload.ActionsSupport.Suspend = false
			apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = "Job"
			workload.Group = "batch"
			workload.Version = "v1"

			apiResponse, _ = workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.Type).To(Equal("Job"))
			Expect(workloadResponse.ActionsSupport.Delete).To(Equal(true))
			Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(true))
			Expect(workloadResponse.Phase).ToNot(Equal(api.PHASE_CREATING))
			Expect(workloadResponse.PhaseMessage).ToNot(Equal(ControlPlaneCreatingMessage))
		})

		It("sets ActionsSupport.Suspend=true for v1-converted Training backed by a V2 Job", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = "Training"
			workload.Group = "batch"
			workload.Version = "v1"

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.Type).To(Equal("Training"))
			Expect(workloadResponse.ActionsSupport.Delete).To(Equal(true))
			Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(true))
		})

		DescribeTable("sets ActionsSupport.Suspend=true for converted Distributed workloads backed by a suspendable Kubeflow Karta",
			func(framework api.DistributedFramework, version string) {
				id := uuid.NewString()
				workload := setupInternalWorkloadCreationFields(id)
				workload.Type = "Distributed"
				workload.Group = "kubeflow.org"
				workload.Version = version
				workload.DistributedFramework = string(framework)

				apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
				Expect(apiResponse.Code).To(Equal(http.StatusCreated))
				createdWorkloadsIds = append(createdWorkloadsIds, id)

				apiResponse, _ = workloadsService.GetWorkload(ctx, id)
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
				Expect(workloadResponse.Type).To(Equal("Distributed"))
				Expect(workloadResponse.DistributedFramework).To(Equal(string(framework)))
				Expect(workloadResponse.ActionsSupport.Delete).To(Equal(true))
				Expect(workloadResponse.ActionsSupport.Suspend).To(Equal(true))
			},
			Entry("PyTorch v1", api.DISTRIBUTEDFRAMEWORK_PY_TORCH, "v1"),
			Entry("TF v1", api.DISTRIBUTEDFRAMEWORK_TF, "v1"),
			Entry("XGBoost v1", api.DISTRIBUTEDFRAMEWORK_XG_BOOST, "v1"),
			Entry("JAX v1", api.DISTRIBUTEDFRAMEWORK_JAX, "v1"),
			Entry("MPI v1", api.DISTRIBUTEDFRAMEWORK_MPI, "v1"),
			Entry("MPI v2beta1", api.DISTRIBUTEDFRAMEWORK_MPI, "v2beta1"),
		)

		It("sets ActionsSupport.Suspend=false when the concrete Kubeflow version has no matching Karta", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = "Distributed"
			workload.Group = "kubeflow.org"
			workload.Version = "v2beta1"
			workload.DistributedFramework = string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH)

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.ActionsSupport.Suspend).To(BeFalse())
		})

		It("Test connections", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Urls = []string{"http://a.com", "http://b.com"}
			baseWorkload.AdditionalFields = "{}"

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(len(workloadResponse.Environments)).To(Equal(2))
			Expect(len(workloadResponse.Environments[0].Connections)).To(Equal(1))
			Expect(workloadResponse.Environments[0].Connections[0].Url).To(Equal("http://a.com"))
			Expect(len(workloadResponse.Environments[1].Connections)).To(Equal(1))
			Expect(workloadResponse.Environments[1].Connections[0].Url).To(Equal("http://b.com"))

			var expectedAdditionalFields map[string]interface{} = nil
			Expect(workloadResponse.AdditionalFields).To(Equal(expectedAdditionalFields))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test connections - inference serving port", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = InferenceType
			authorizedUsers := []string{"user1@run.ai", "user2@run.ai"}
			workload.Environments = []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:     "my-chatbot",
							ToolType: "Chatbot UI",
						},
						{
							Name:              "ServingPort",
							ConnectionType:    "ServingPort",
							ToolType:          "custom",
							ContainerPort:     8080,
							AuthorizationType: composer.ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS,
							AuthorizedUsers:   authorizedUsers,
						},
					},
					Name:        "inference env",
					Id:          testing.TestUuid3,
					ReplicaType: nil,
				},
			}

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Type = InferenceType
			baseWorkload.Urls = []string{"tool-url", "serving-port-url"}
			baseWorkload.AdditionalFields = "{}"

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(len(workloadResponse.Environments)).To(Equal(1))
			Expect(len(workloadResponse.Environments[0].Connections)).To(Equal(2))
			Expect(workloadResponse.Environments[0].Connections[0].Url).To(Equal("tool-url"))
			Expect(workloadResponse.Environments[0].Connections[1].Url).To(Equal("serving-port-url"))
			Expect(workloadResponse.Environments[0].Connections[1].AuthorizationType).To(Equal(composer.ConnectionAuthorizationType_AUTHORIZED_USERS_OR_GROUPS))
			Expect(workloadResponse.Environments[0].Connections[1].AuthorizedUsers).To(Equal(authorizedUsers))
			Expect(workloadResponse.Environments[0].Connections[1].AuthorizedGroups).To(BeNil())

			var expectedAdditionalFields map[string]interface{} = nil
			Expect(workloadResponse.AdditionalFields).To(Equal(expectedAdditionalFields))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test external connections - SyncWorkloads for external workloads", func() {
			id := uuid.NewString()
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			// Set ExternalWorkloadName to make it an external workload
			baseWorkload.Type = "DynamoGraphDeployment"
			baseWorkload.ExternalWorkloadName = "external-deployment-" + id
			baseWorkload.Urls = []string{
				"https://subdomain1.example.com/path1",
				"https://subdomain2.example.com:8080/path2",
			}

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{baseWorkload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Validate ExternalConnections are populated
			Expect(workloadResponse.ExternalConnections).ToNot(BeNil())
			Expect(len(workloadResponse.ExternalConnections)).To(Equal(2))

			// Validate first connection
			conn1 := workloadResponse.ExternalConnections[0]
			Expect(conn1.Url).To(Equal("https://subdomain1.example.com/path1"))
			Expect(conn1.ConnectionType).To(Equal("ExternalUrl"))
			Expect(conn1.CorrelationName).To(Equal("url"))
			Expect(conn1.ExposedUrlSubDomain).ToNot(BeNil())
			Expect(*conn1.ExposedUrlSubDomain).To(Equal("subdomain1"))
			Expect(conn1.ExposedUrlPath).ToNot(BeNil())
			Expect(*conn1.ExposedUrlPath).To(Equal("/path1"))
			Expect(conn1.ContainerPort).To(Equal(int32(8080))) // Default port

			// Validate second connection
			conn2 := workloadResponse.ExternalConnections[1]
			Expect(conn2.Url).To(Equal("https://subdomain2.example.com:8080/path2"))
			Expect(conn2.Name).To(Equal(baseWorkload.Type + "-2")) // Second URL: name with suffix
			Expect(conn2.ConnectionType).To(Equal("ExternalUrl"))
			Expect(conn2.CorrelationName).To(Equal("url"))
			Expect(conn2.ExposedUrlSubDomain).ToNot(BeNil())
			Expect(*conn2.ExposedUrlSubDomain).To(Equal("subdomain2"))
			Expect(conn2.ExposedUrlPath).ToNot(BeNil())
			Expect(*conn2.ExposedUrlPath).To(Equal("/path2"))
			Expect(conn2.ContainerPort).To(Equal(int32(8080)))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test external connections - UpsertWorkload for external workloads", func() {
			id := uuid.NewString()
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			// Set ExternalWorkloadName to make it an external workload
			baseWorkload.Type = "DynamoGraphDeployment"
			baseWorkload.ExternalWorkloadName = "external-statefulset-" + id
			baseWorkload.Urls = []string{"https://test.example.com/api"}

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Validate ExternalConnections are populated
			Expect(workloadResponse.ExternalConnections).ToNot(BeNil())
			Expect(len(workloadResponse.ExternalConnections)).To(Equal(1))

			conn := workloadResponse.ExternalConnections[0]
			Expect(conn.Url).To(Equal("https://test.example.com/api"))
			Expect(conn.Name).To(Equal(baseWorkload.Type))
			Expect(conn.ConnectionType).To(Equal("ExternalUrl"))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test external connections NOT populated - SyncWorkloads for non-external workloads with URLs", func() {
			id := uuid.NewString()
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			// Do NOT set ExternalWorkloadName - this is a native workload
			baseWorkload.ExternalWorkloadName = ""
			baseWorkload.Urls = []string{
				"https://native1.example.com/path1",
				"https://native2.example.com/path2",
			}

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{baseWorkload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Validate ExternalConnections are NOT populated for non-external workloads
			Expect(len(workloadResponse.ExternalConnections)).To(Equal(0))
			// URLs should still be present
			Expect(workloadResponse.Urls).To(Equal(baseWorkload.Urls))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test external connections NOT populated - UpsertWorkload for non-external workloads with URLs", func() {
			id := uuid.NewString()
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			// Do NOT set ExternalWorkloadName - this is a native workload
			baseWorkload.ExternalWorkloadName = ""
			baseWorkload.Urls = []string{"https://native-workload.example.com"}

			apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, baseWorkload)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Validate ExternalConnections are NOT populated for non-external workloads
			Expect(len(workloadResponse.ExternalConnections)).To(Equal(0))
			// URLs should still be present
			Expect(workloadResponse.Urls).To(Equal(baseWorkload.Urls))

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test endpoints - SyncWorkloads persists and returns endpoints", func() {
			id := uuid.NewString()
			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Type = "ExternalWorkload" // non-native type — DB-persisted endpoints only apply to non-native workloads
			baseWorkload.Endpoints = []api.EndpointSync{
				{
					Host:       "chat.example.com",
					Port:       8080,
					Path:       "/v1/completions",
					TlsEnabled: ptr.To(true),
					Protocol:   "HTTPS",
					Url:        "https://chat.example.com:8080/v1/completions",
					SourceRef: api.EndpointSyncAllOfSourceRef{
						Kind: "Ingress",
						Name: "my-ingress",
					},
				},
				{
					Host:     "api.example.com",
					Port:     443,
					Path:     "/health",
					Protocol: "HTTPS",
					Url:      "https://api.example.com/health",
				},
				{
					Host:       "api.example2.com",
					Port:       80,
					Path:       "/health",
					Protocol:   "HTTP",
					Url:        "http://api.example2.com/health",
					TlsEnabled: ptr.To(false),
				},
			}

			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{baseWorkload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			apiWorkloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			validateEndpoints(apiWorkloadResponse.Endpoints)

			apiResponse, _ = workloadsService.GetWorkloadEndpoints(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			bytes, err := json.Marshal(apiResponse.Body)
			Expect(err).To(BeNil())
			response := &api.GetWorkloadEndpoints200Response{}
			err = json.Unmarshal(bytes, response)
			Expect(err).To(BeNil())

			validateEndpoints(response.Endpoints)

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test endpoints - GetWorkloadEndpoints derives endpoints from connections for native workloads", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)
			workload.Type = InferenceType
			workload.Environments = []api.Environment{
				{
					Connections: []api.Connection{
						{
							Name:          "serving-port",
							ToolType:      "custom",
							ContainerPort: 8080,
						},
					},
					Name: "inference env",
					Id:   testing.TestUuid3,
				},
			}

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.Type = InferenceType
			baseWorkload.Urls = []string{"https://model.example.com/v1/completions"}

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{baseWorkload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			// GetWorkloadEndpoints should derive endpoints from connections for native workloads
			apiResponse, _ = workloadsService.GetWorkloadEndpoints(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			bytes, err := json.Marshal(apiResponse.Body)
			Expect(err).To(BeNil())
			endpointsResponse := &api.GetWorkloadEndpoints200Response{}
			err = json.Unmarshal(bytes, endpointsResponse)
			Expect(err).To(BeNil())

			Expect(len(endpointsResponse.Endpoints)).To(Equal(1))
			ep := endpointsResponse.Endpoints[0]
			Expect(ep.Host).To(Equal("model.example.com"))
			Expect(ep.Path).To(Equal("/v1/completions"))
			Expect(ep.Protocol).To(Equal("HTTPS"))
			Expect(ep.Url).To(Equal("https://model.example.com/v1/completions"))
			Expect(*ep.TlsEnabled).To(BeTrue())

			workloadsService.DeleteWorkload(ctx, id)
		})

		It("Test additional fields", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)

			apiResponse, _ := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))

			baseWorkload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
			baseWorkload.AdditionalFields = "{\"a\": \"b\", \"cliCommand\": \"runai submit --name test --image ubuntu\"}"

			apiResponse, _ = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{
					baseWorkload,
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkload(ctx, id)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)
			Expect(workloadResponse.AdditionalFields).ToNot(BeNil())
			Expect(len(workloadResponse.AdditionalFields)).To(Equal(2))
			Expect(workloadResponse.AdditionalFields["a"]).To(Equal("b"))
			Expect(workloadResponse.AdditionalFields["cliCommand"]).To(Equal("runai submit --name test --image ubuntu"))

			workloadsService.DeleteWorkload(ctx, id)
		})
		It("Test create and immediately count to see if name is taken", func() {
			id := uuid.NewString()
			workload := setupInternalWorkloadCreationFields(id)

			apiResponse, err := workloadsService.InternalCreateWorkload(ctx, workload)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusCreated))
			createdWorkloadsIds = append(createdWorkloadsIds, id)

			filterBy := []string{
				fmt.Sprintf("name==%s", workload.Name),
				fmt.Sprintf("clusterId==%s", workload.ClusterId),
				fmt.Sprintf("projectId==%s", workload.ProjectId),
			}
			apiResponse, err = workloadsService.CountWorkloads(ctx, false, filterBy, "", true)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsCountResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
			Expect(workloadsCountResponse.Count).To(BeEquivalentTo(1))
		})
	})

	Context("Sync Workloads", func() {
		It("Should Have Guaranteed time field with correct value", func() {
			// Create workloads with sensitive environment variables
			workloads := []api.WorkloadSyncFields{}
			workloadIds := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
			for _, id := range workloadIds {
				workloads = append(workloads, SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1))
			}

			// Sync the workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: workloads,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			waitForGetWorkloadsLength(ctx, workloadsService, 3+len(workloads), []string{})

			// Test with user1 (should redact user2's workload but not user1's)
			user1Ctx := testing.SetContext(testing.TestTenant, "user1", []string{})
			apiResponse, _ = workloadsService.GetWorkloads(user1Ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)

			for i := range workloadsResponse.Workloads {
				workload := &workloadsResponse.Workloads[i]
				Expect(workload.GuaranteedRuntimeEndsAt).To(Not(BeNil()))
				Expect(*workload.GuaranteedRuntimeEndsAt).To(BeTemporally("~", DefaultGuaranteedEndsAtTime, time.Second))
			}
			apiResponse, _ = workloadsService.GetWorkload(user1Ctx, workloads[0].Id)
			fetchedWorkload, ok := apiResponse.Body.(api.WorkloadDetailed)
			Expect(ok).To(BeTrue())
			Expect(fetchedWorkload.GuaranteedRuntimeEndsAt).To(Not(BeNil()))
			Expect(*fetchedWorkload.GuaranteedRuntimeEndsAt).To(BeTemporally("~", DefaultGuaranteedEndsAtTime, time.Second))
		})

		It("Should preserve ExternalWorkloadName", func() {
			workloads := []api.WorkloadSyncFields{}
			workloadIds := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
			for _, id := range workloadIds {
				workload := SetupBaseWorkload(id, DefaultProjectId, testing.ClusterUuid1)
				workloads = append(workloads, workload)
			}

			// Sync the workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: workloads,
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			waitForGetWorkloadsLength(ctx, workloadsService, 3+len(workloads), []string{})

			// Verify the value is preserved by both the list and details APIs.
			user1Ctx := testing.SetContext(testing.TestTenant, "user1", []string{})
			apiResponse, _ = workloadsService.GetWorkloads(user1Ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)

			for i := range workloadsResponse.Workloads {
				workload := &workloadsResponse.Workloads[i]
				Expect(workload.ExternalWorkloadName).To(Equal(fmt.Sprintf("ew-training-%s", workload.Id)))
			}
			apiResponse, _ = workloadsService.GetWorkload(user1Ctx, workloads[0].Id)
			fetchedWorkload, ok := apiResponse.Body.(api.WorkloadDetailed)
			Expect(ok).To(BeTrue())
			Expect(fetchedWorkload.ExternalWorkloadName).To(Equal(workloads[0].ExternalWorkloadName))
		})
	})

	Context("Get Workload and Pod Metrics", func() {
		// TODO: once we have a way to re-mock the metrics client, move pods tests
		const baseTimestampUnix = 1708958503653
		const numOfSamples = 10
		const numOfMultipleResult = 8
		const numOfNodePools = 3
		const numOfResultsPerNodePool = 3

		BeforeAll(func() {
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiAllocatedGpuCountPerWorkloadMetric{}.GetName():       getDefaultMetricMockMultipleNodePools(10, baseTimestampUnix, numOfSamples, numOfNodePools*numOfResultsPerNodePool, numOfNodePools),
				runaimetrics.RunaiGpuUtilizationPerWorkloadMetric{}.GetName():          getDefaultMetricMockResult(80, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryUsedBytesPerWorkloadMetric{}.GetName():      getDefaultMetricMockResult(900, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryRequestedBytesPerWorkloadMetric{}.GetName(): getConstMetricMockResult(1000, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryLimitBytesPerWorkloadMetric{}.GetName():     getConstMetricMockResult(1100, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuUsedCoresPerWorkloadMetric{}.GetName():            getDefaultMetricMockResult(50, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuRequestedCoresPerWorkloadMetric{}.GetName():       getConstMetricMockResult(1, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuLimitCoresPerWorkloadMetric{}.GetName():           getConstMetricMockResult(2, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiMemoryUsedBytesPerWorkloadMetric{}.GetName():         getDefaultMetricMockResult(150, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiMemoryRequestedBytesPerWorkloadMetric{}.GetName():    getConstMetricMockResult(100, baseTimestampUnix, numOfSamples),
				// metrics.RunaiMemoryLimitPerWorkloadMetric{}.GetName():      getConstMetricMockResult(500, baseTimestampUnix, numOfSamples),
				// Dont have a mock for metrics.RunaiMemoryLimitPerWorkloadMetric{}.GetName() to test a case where no result returned

				runaimetrics.RunaiGpuUtilizationPerPodPerGpuMetric{}.GetName():     getDefaultMetricMockMultipleResults(60, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuMemoryUsedBytesPerPodPerGpuMetric{}.GetName(): getDefaultMetricMockMultipleResults(700, baseTimestampUnix, numOfSamples, 2),
				runaimetrics.RunaiMemoryUsedBytesPerPodMetric{}.GetName():          getDefaultMetricMockResult(100, baseTimestampUnix, numOfSamples),
				// metrics.RunaiCpuUsagePerPodMetric{}.GetName():             getDefaultMetricMockResult(50, baseTimestampUnix, numOfSamples),
				// Dont have a mock for metrics.RunaiCpuUsagePerPodMetric{}.GetName() to test a case where no result returned

				runaimetrics.RunaiGpuGREngineActivePerPodPerGpuMetric{}.GetName():   getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuSmActivePerPodPerGpuMetric{}.GetName():         getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuSmOccupancyPerPodPerGpuMetric{}.GetName():      getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPipeTensorActivePerPodPerGpuMetric{}.GetName(): getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPipeFp64ActivePerPodPerGpuMetric{}.GetName():   getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPipeFp32ActivePerPodPerGpuMetric{}.GetName():   getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPipeFp16ActivePerPodPerGpuMetric{}.GetName():   getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuDramActivePerPodPerGpuMetric{}.GetName():       getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuNvlinkTxBytesPerPodPerGpuMetric{}.GetName():    getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuNvlinkRxBytesPerPodPerGpuMetric{}.GetName():    getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPcieTxBytesPerPodPerGpuMetric{}.GetName():      getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuPcieRxBytesPerPodPerGpuMetric{}.GetName():      getDefaultMetricMockMultipleResults(0, baseTimestampUnix, numOfSamples, numOfMultipleResult),

				// Add NIM pod metrics (time-series)
				runaimetrics.RunaiNimNumRequestsRunningPerPodMetric{}.GetName():  getDefaultMetricMockResult(2, 1710958503653, 4),
				runaimetrics.RunaiNimNumRequestsWaitingPerPodMetric{}.GetName():  getDefaultMetricMockResult(1, 1710958503653, 4),
				runaimetrics.RunaiNimNumRequestMaxPerPodMetric{}.GetName():       getConstMetricMockResult(8, 1710958503653, 4),
				runaimetrics.RunaiNimRequestSuccessTotalPerPodMetric{}.GetName(): getDefaultMetricMockResult(20, 1710958503653, 4),
				runaimetrics.RunaiNimRequestFailureTotalPerPodMetric{}.GetName(): getDefaultMetricMockResult(3, 1710958503653, 4),
				runaimetrics.RunaiNimGpuCacheUsagePercPerPodMetric{}.GetName():   getDefaultMetricMockResult(40, 1710958503653, 4),

				// Add NIM pod histogram bucket metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric{}.GetName():  getBucketMetricMock(1711058503000, [][2]string{{"0.1", "1"}, {"+Inf", "5"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerPodMetric{}.GetName(): getBucketMetricMock(1711058503000, [][2]string{{"0.5", "3"}, {"+Inf", "7"}}),

				// Add NIM pod histogram percentile metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric{}.GetName():  getPercentileMetricMock(1711158503000, [][2]string{{"50", "0.08"}, {"90", "0.20"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric{}.GetName(): getPercentileMetricMock(1711158503000, [][2]string{{"50", "1.2"}, {"99", "3.4"}}),

				// vLLM pod metrics use the same composite response paths as NIM.
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMNumRequestsRunning}).GetName():                getDefaultMetricMockResult(4, 1711958503653, 4),
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMNumRequestsWaiting}).GetName():                getDefaultMetricMockResult(2, 1711958503653, 4),
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMRequestSuccessRate}).GetName():                getDefaultMetricMockResult(95, 1711958503653, 4),
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMKVCacheUsagePerc}).GetName():                  getDefaultMetricMockResult(60, 1711958503653, 4),
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMTimeToFirstTokenSecondsBucket}).GetName():     getBucketMetricMock(1712058503000, [][2]string{{"0.1", "2"}, {"+Inf", "8"}}),
				(runaimetrics.RunaiVLLMPerPodMetric{Name: runaimetrics.VLLMRequestQueueTimeSecondsPercentile}).GetName(): getPercentileMetricMock(1712158503000, [][2]string{{"50", "0.03"}, {"95", "0.10"}, {"99", "0.15"}}),
			})
		})

		// Ensure the workload under test is marked as NIM-capable only for tests in this context
		BeforeEach(func() {
			w1 := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			w1.AdditionalFields = fmt.Sprintf(
				"{\"%s\":\"true\",\"%s\":\"true\"}",
				clusterCommonConstants.AdditionalNIMWorkload,
				clusterCommonConstants.AdditionalVLLMWorkload,
			)
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{w1},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: AnotherProjectId,
				Pods: []api.PodSyncFields{
					SetupBasePod(testing.TestUuid6, id1, AnotherProjectId, testing.ClusterUuid2),
				},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		})

		It("Get multiple metrics", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_GPU_UTILIZATION, api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES, api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES, api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES}, time.Now(), time.Now().Add(-5*time.Minute), 20)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(4))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_GPU_UTILIZATION))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES))
			Expect(metrics.Measurements[3].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
			Expect((*metrics.Measurements[0].Values)[9].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 9000)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("80"))
			Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("81"))
			Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("82"))
			Expect((*metrics.Measurements[0].Values)[8].Value).To(Equal("88"))
			Expect((*metrics.Measurements[0].Values)[9].Value).To(Equal("89"))

			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[1].Values)[5].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix) + 5000))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("1000"))
			Expect((*metrics.Measurements[1].Values)[5].Value).To(Equal("1000"))
		})

		It("Get multiple metrics with single failure metric", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid2, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_CPU_MEMORY_REQUEST_BYTES, api.WORKLOADMETRICTYPE_CPU_MEMORY_LIMIT_BYTES, api.WORKLOADMETRICTYPE_CPU_USAGE_CORES}, time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))

			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(3))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_CPU_MEMORY_REQUEST_BYTES))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_CPU_MEMORY_LIMIT_BYTES))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_CPU_USAGE_CORES))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect(metrics.Measurements[1].Values).To(BeNil())
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
		})

		It("Get metrics with unexisting workloadId", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, uuid.NewString(), []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_CPU_MEMORY_REQUEST_BYTES}, time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(404))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Failed to get clusterId of workload with id"))
		})

		It("Get allocation per nodepool", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid2, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_GPU_ALLOCATION}, time.Now(), time.Now().Add(-4*time.Minute), 10)
			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(err).To(BeNil())
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(numOfNodePools))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples * numOfResultsPerNodePool))
			Expect(resp.Code).To(Equal(200))
		})

		It("Get multiple pod metrics", func() {
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES, api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU, api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU},
				time.Now(), time.Now().Add(-5*time.Minute), 15)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(2 + numOfMultipleResult + 1))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU))
			Expect(metrics.Measurements[2+numOfMultipleResult-2].Type).To(BeEquivalentTo(api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU))
			Expect(metrics.Measurements[2+numOfMultipleResult-1].Type).To(BeEquivalentTo(api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU))
			Expect(metrics.Measurements[2+numOfMultipleResult].Type).To(BeEquivalentTo(api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect(len(*metrics.Measurements[5].Values)).To(Equal(numOfSamples))

			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[4].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[4].Values)[3].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 3000)))

			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("100"))
			Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("101"))
			Expect((*metrics.Measurements[0].Values)[8].Value).To(Equal("108"))
			Expect(metrics.Measurements[0].Labels).To(BeNil())
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("60"))
			Expect((*metrics.Measurements[1].Values)[3].Value).To(Equal("63"))
			Expect(metrics.Measurements[1].Labels).To(Not(BeNil()))
			Expect(*metrics.Measurements[1].Labels).To(BeEquivalentTo(map[string]string{"gpu": "0"}))
			Expect((*metrics.Measurements[1].Values)[7].Value).To(Equal("67"))
			Expect((*metrics.Measurements[6].Values)[0].Value).To(Equal("60"))
			Expect((*metrics.Measurements[6].Values)[6].Value).To(Equal("66"))
			Expect((*metrics.Measurements[6].Values)[9].Value).To(Equal("69"))
			Expect(metrics.Measurements[6].Labels).To(Not(BeNil()))
			Expect(*metrics.Measurements[6].Labels).To(BeEquivalentTo(map[string]string{"gpu": "5"}))
			Expect((*metrics.Measurements[2+numOfMultipleResult-1].Values)[0].Value).To(Equal("700"))
			Expect((*metrics.Measurements[2+numOfMultipleResult-1].Values)[2].Value).To(Equal("702"))
			Expect((*metrics.Measurements[2+numOfMultipleResult-1].Values)[5].Value).To(Equal("705"))
			Expect(metrics.Measurements[2+numOfMultipleResult-1].Labels).To(Not(BeNil()))
			Expect(metrics.Measurements[2+numOfMultipleResult].Labels).To(Not(BeNil()))
			Expect(*metrics.Measurements[2+numOfMultipleResult-1].Labels).To(BeEquivalentTo(map[string]string{"gpu": "0"}))
			Expect(*metrics.Measurements[2+numOfMultipleResult].Labels).To(BeEquivalentTo(map[string]string{"gpu": "1"}))
		})

		It("Get NIM pod basic metrics (time-series)", func() {
			const baseTimestampUnix = 1710958503653
			const numOfSamples = 4

			requested := []api.PodMetricType{
				api.PODMETRICTYPE_NIM_NUM_REQUESTS_RUNNING,
				api.PODMETRICTYPE_NIM_NUM_REQUESTS_WAITING,
				api.PODMETRICTYPE_NIM_NUM_REQUEST_MAX,
				api.PODMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL,
				api.PODMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL,
				api.PODMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC,
			}

			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requested, time.Now(), time.Now().Add(-5*time.Minute), 20)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metricsResp, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metricsResp.Measurements)).To(Equal(len(requested)))
			Expect(metricsResp.Measurements[0].Type).To(BeEquivalentTo(api.PODMETRICTYPE_NIM_NUM_REQUESTS_RUNNING))
			Expect(len(*metricsResp.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metricsResp.Measurements[0].Values)[0].Value).To(Equal("2"))
			Expect((*metricsResp.Measurements[2].Values)[0].Value).To(Equal("8"))
		})

		It("Get NIM pod histogram bucket metrics", func() {
			requested := []api.PodMetricType{
				api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS,
			}
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			ordered, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(ordered.Measurements)).To(Equal(0))
			Expect(ordered.Histogram).ToNot(BeNil())
			Expect(len(*ordered.Histogram)).To(Equal(2))
			Expect((*ordered.Histogram)[0].Type).To(Equal(string(api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)))
			Expect((*ordered.Histogram)[0].Values[0].Data["0.1"]).To(Equal("1"))
			Expect((*ordered.Histogram)[0].Values[0].Data["+Inf"]).To(Equal("5"))
		})

		It("Get NIM pod histogram percentile metrics", func() {
			requested := []api.PodMetricType{
				api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES,
				api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
			}
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			ordered, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(ordered.Measurements)).To(Equal(0))
			Expect(ordered.Histogram).ToNot(BeNil())
			Expect(len(*ordered.Histogram)).To(Equal(2))
			Expect((*ordered.Histogram)[0].Type).To(Equal(string(api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES)))
			Expect((*ordered.Histogram)[0].Values[0].Data["p50"]).To(Equal("0.08"))
			Expect((*ordered.Histogram)[0].Values[0].Data["p90"]).To(Equal("0.20"))
		})

		It("Get vLLM pod time-series, histogram, and percentile metrics", func() {
			requested := []api.PodMetricType{
				api.PODMETRICTYPE_VLLM_NUM_REQUESTS_RUNNING,
				api.PODMETRICTYPE_VLLM_NUM_REQUESTS_WAITING,
				api.PODMETRICTYPE_VLLM_REQUEST_SUCCESS_RATE,
				api.PODMETRICTYPE_VLLM_KV_CACHE_USAGE_PERC,
				api.PODMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.PODMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES,
			}

			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requested, time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			metricsResp, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(metricsResp.Measurements).To(HaveLen(4))
			Expect((*metricsResp.Measurements[0].Values)[0].Value).To(Equal("4"))
			Expect((*metricsResp.Measurements[2].Values)[0].Value).To(Equal("95"))
			Expect(metricsResp.Histogram).NotTo(BeNil())
			Expect(*metricsResp.Histogram).To(HaveLen(2))
			Expect((*metricsResp.Histogram)[0].Values[0].Data["+Inf"]).To(Equal("8"))
			Expect((*metricsResp.Histogram)[1].Values[0].Data["p95"]).To(Equal("0.10"))
			Expect((*metricsResp.Histogram)[1].Values[0].Data["p99"]).To(Equal("0.15"))
		})

		It("Get multiple pod metrics with single failure metric", func() {
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_CPU_USAGE_CORES, api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES, api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU},
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))

			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(4))
			Expect(metrics.Measurements[0].Labels).To(BeNil())
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.PODMETRICTYPE_CPU_USAGE_CORES))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU))
			Expect(metrics.Measurements[0].Values).To(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
		})

		It("Get pod metrics with unexisting podId/workloadId", func() {
			resp, err := podsService.GetWorkloadPodMetrics(ctx, uuid.NewString(), testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_CPU_USAGE_CORES}, time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(404))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Failed to get clusterId of pod with id"))

			resp, err = podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, uuid.NewString(),
				[]api.PodMetricType{api.PODMETRICTYPE_CPU_USAGE_CORES}, time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(404))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Failed to get clusterId of pod with id"))
		})

		It("Request too many workload metrics in 1 request", func() {
			requestedMetrics := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_GPU_UTILIZATION,
				api.WORKLOADMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
				api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES,
				api.WORKLOADMETRICTYPE_CPU_USAGE_CORES,
				api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES,
				api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES,
				api.WORKLOADMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
				api.WORKLOADMETRICTYPE_CPU_MEMORY_REQUEST_BYTES,
				api.WORKLOADMETRICTYPE_CPU_MEMORY_LIMIT_BYTES,
				api.WORKLOADMETRICTYPE_POD_COUNT,
				api.WORKLOADMETRICTYPE_RUNNING_POD_COUNT,
			}

			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requestedMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(400))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Too many metrics requested"))
		})

		It("Request too many pod metrics in 1 request + validate advanced gpu metrics", func() {
			mockTenantSettingsClient.On("GetAdvancedGpuMetricsSetting", mock.Anything).Return(true, nil).Once()
			requestedMetrics := []api.PodMetricType{
				api.PODMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU,
				api.PODMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU,
				api.PODMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU,
				api.PODMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU,
				api.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU,
				api.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU,
				api.PODMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU,
				api.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU,
				api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU,
				api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
				api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU,
			}

			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requestedMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(400))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Too many metrics requested"))

			requestedMetrics = requestedMetrics[:10]
			resp, err = podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requestedMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))
			metrics, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(1 + (len(requestedMetrics)-1)*numOfMultipleResult))
			Expect(metrics.Measurements[0].Labels).ToNot(BeNil())
			Expect(*metrics.Measurements[0].Labels).To(BeEquivalentTo(map[string]string{"gpu": "0"}))

			for i := 0; i < len(requestedMetrics)-1; i++ {
				for j := 0; j < numOfMultipleResult; j++ {
					Expect(metrics.Measurements[i*numOfMultipleResult+j].Labels).ToNot(BeNil(), "i=%d, j=%d", i, j)
					Expect(*metrics.Measurements[i*numOfMultipleResult+j].Labels).To(BeEquivalentTo(map[string]string{"gpu": fmt.Sprintf("%d", j)}), "i=%d, j=%d", i, j)
				}
			}

			mockTenantSettingsClient.On("GetAdvancedGpuMetricsSetting", mock.Anything).Return(false, nil).Once()

			resp, err = podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requestedMetrics,
				time.Now(), time.Now().Add(-4*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(400))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Advanced GPU metrics are not enabled"))
		})

		It("Get NIM workload basic metrics (time-series)", func() {
			const baseTimestampUnix = 1709958503653
			const numOfSamples = 5
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiNimNumRequestsRunningPerWorkloadMetric{}.GetName():  getDefaultMetricMockResult(3, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiNimNumRequestsWaitingPerWorkloadMetric{}.GetName():  getDefaultMetricMockResult(1, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiNimNumRequestMaxPerWorkloadMetric{}.GetName():       getConstMetricMockResult(10, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiNimRequestSuccessTotalPerWorkloadMetric{}.GetName(): getDefaultMetricMockResult(100, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiNimRequestFailureTotalPerWorkloadMetric{}.GetName(): getDefaultMetricMockResult(2, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiNimGpuCacheUsagePercPerWorkloadMetric{}.GetName():   getDefaultMetricMockResult(55, baseTimestampUnix, numOfSamples),
			})

			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING,
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING,
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUEST_MAX,
				api.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL,
				api.WORKLOADMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL,
				api.WORKLOADMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC,
			}

			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-5*time.Minute), 20)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metricsResp, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metricsResp.Measurements)).To(Equal(len(requested)))
			Expect(metricsResp.Measurements[0].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING))
			Expect(metricsResp.Measurements[1].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING))
			Expect(metricsResp.Measurements[2].Type).To(BeEquivalentTo(api.WORKLOADMETRICTYPE_NIM_NUM_REQUEST_MAX))
			Expect(len(*metricsResp.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metricsResp.Measurements[0].Values)[0].Value).To(Equal("3"))
			Expect((*metricsResp.Measurements[1].Values)[1].Value).To(Equal("2"))
			Expect((*metricsResp.Measurements[2].Values)[0].Value).To(Equal("10"))
			Expect((*metricsResp.Measurements[3].Values)[0].Value).To(Equal("100"))
			Expect((*metricsResp.Measurements[4].Values)[0].Value).To(Equal("2"))
			Expect((*metricsResp.Measurements[5].Values)[0].Value).To(Equal("55"))
		})

		It("Get NIM workload histogram bucket metrics", func() {
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric{}.GetName(): {
					Status:  "success",
					IsRange: true,
					Data: []openapi.Result{
						{
							MetricLabels: map[string]string{"le": "0.1"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710058503000), Value: "1"}, {Timestamp: time.UnixMilli(1710058503000 + 1000), Value: "2"}},
						},
						{
							MetricLabels: map[string]string{"le": "+Inf"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710058503000), Value: "5"}, {Timestamp: time.UnixMilli(1710058503000 + 1000), Value: "6"}},
						},
					},
				},
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric{}.GetName(): {
					Status:  "success",
					IsRange: true,
					Data: []openapi.Result{
						{
							MetricLabels: map[string]string{"le": "0.5"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710058503000), Value: "3"}, {Timestamp: time.UnixMilli(1710058503000 + 1000), Value: "4"}},
						},
						{
							MetricLabels: map[string]string{"le": "+Inf"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710058503000), Value: "7"}, {Timestamp: time.UnixMilli(1710058503000 + 1000), Value: "8"}},
						},
					},
				},
			})

			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS,
			}
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			ordered, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(ordered.Measurements)).To(Equal(0))
			Expect(ordered.Histogram).ToNot(BeNil())
			Expect(len(*ordered.Histogram)).To(Equal(2))
			// Validate first histogram series (TTFT)
			Expect((*ordered.Histogram)[0].Type).To(Equal(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)))
			Expect(len((*ordered.Histogram)[0].Values)).To(Equal(2))
			Expect((*ordered.Histogram)[0].Values[0].Data["0.1"]).To(Equal("1"))
			Expect((*ordered.Histogram)[0].Values[0].Data["+Inf"]).To(Equal("5"))
		})

		It("Get NIM workload histogram percentile metrics", func() {
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric{}.GetName(): {
					Status:  "success",
					IsRange: true,
					Data: []openapi.Result{
						{
							MetricLabels: map[string]string{"quantile_label": "50"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710158503000), Value: "0.08"}, {Timestamp: time.UnixMilli(1710158503000 + 1000), Value: "0.09"}},
						},
						{
							MetricLabels: map[string]string{"quantile_label": "90"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710158503000), Value: "0.20"}, {Timestamp: time.UnixMilli(1710158503000 + 1000), Value: "0.22"}},
						},
					},
				},
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric{}.GetName(): {
					Status:  "success",
					IsRange: true,
					Data: []openapi.Result{
						{
							MetricLabels: map[string]string{"quantile_label": "50"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710158503000), Value: "1.2"}, {Timestamp: time.UnixMilli(1710158503000 + 1000), Value: "1.3"}},
						},
						{
							MetricLabels: map[string]string{"quantile_label": "99"},
							Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(1710158503000), Value: "3.4"}, {Timestamp: time.UnixMilli(1710158503000 + 1000), Value: "3.6"}},
						},
					},
				},
			})

			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES,
				api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
			}
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			ordered, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(len(ordered.Measurements)).To(Equal(0))
			Expect(ordered.Histogram).ToNot(BeNil())
			Expect(len(*ordered.Histogram)).To(Equal(2))
			// Validate first percentile series (TTFT percentiles)
			Expect((*ordered.Histogram)[0].Type).To(Equal(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES)))
			Expect((*ordered.Histogram)[0].Values[0].Data["p50"]).To(Equal("0.08"))
			Expect((*ordered.Histogram)[0].Values[0].Data["p90"]).To(Equal("0.20"))
		})

		It("Get vLLM workload time-series, histogram, and percentile metrics", func() {
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLMNumRequestsRunning}).GetName():                 getDefaultMetricMockResult(6, 1712958503653, 4),
				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLMPromptTokensPerSecond}).GetName():              getDefaultMetricMockResult(120, 1712958503653, 4),
				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLME2ERequestLatencySecondsBucket}).GetName():     getBucketMetricMock(1713058503000, [][2]string{{"0.5", "4"}, {"+Inf", "9"}}),
				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLMInterTokenLatencySecondsPercentile}).GetName(): getPercentileMetricMock(1713158503000, [][2]string{{"50", "0.02"}, {"95", "0.06"}, {"99", "0.09"}}),
			})
			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_VLLM_NUM_REQUESTS_RUNNING,
				api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND,
				api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS,
				api.WORKLOADMETRICTYPE_VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES,
			}

			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			metricsResp, ok := resp.Body.(api.MetricsCompositeResponse)
			Expect(ok).To(BeTrue())
			Expect(metricsResp.Measurements).To(HaveLen(2))
			Expect((*metricsResp.Measurements[0].Values)[0].Value).To(Equal("6"))
			Expect((*metricsResp.Measurements[1].Values)[0].Value).To(Equal("120"))
			Expect(metricsResp.Histogram).NotTo(BeNil())
			Expect(*metricsResp.Histogram).To(HaveLen(2))
			Expect((*metricsResp.Histogram)[0].Values[0].Data["+Inf"]).To(Equal("9"))
			Expect((*metricsResp.Histogram)[1].Values[0].Data["p95"]).To(Equal("0.06"))
			Expect((*metricsResp.Histogram)[1].Values[0].Data["p99"]).To(Equal("0.09"))
		})

		It("Get unsupported cluster metric", func() {
			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
			}
			resp1, err := workloadsService.GetWorkloadMetrics(ctx, id1, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp1.Code).To(Equal(http.StatusBadRequest))

			requested = []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
			}
			resp1, err = workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp1.Code).To(Equal(http.StatusMultiStatus))

			requestedPodTypes := []api.PodMetricType{
				api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
			}
			resp2, err := podsService.GetWorkloadPodMetrics(ctx, id1, testing.TestUuid6, requestedPodTypes, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp2.Code).To(Equal(http.StatusBadRequest))

			requestedPodTypes = []api.PodMetricType{
				api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU,
			}
			resp2, err = podsService.GetWorkloadPodMetrics(ctx, id1, testing.TestUuid6, requestedPodTypes, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp2.Code).To(Equal(http.StatusBadRequest))

			requestedPodTypes = []api.PodMetricType{
				api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
				api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU,
			}
			resp2, err = podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4, requestedPodTypes, time.Now(), time.Now().Add(-1*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp2.Code).To(Equal(http.StatusMultiStatus))
		})
	})

	Context("Get Workload and Pod Metrics csv", func() {
		// TODO: once we have a way to re-mock the metrics client, move pods tests
		const baseTimestampUnix = 1708958503653
		const numOfSamples = 10
		const numOfMultipleResult = 8

		BeforeAll(func() {
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiGpuUtilizationPerWorkloadMetric{}.GetName():          getDefaultMetricMockResult(80, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryUsedBytesPerWorkloadMetric{}.GetName():      getDefaultMetricMockResult(900, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryRequestedBytesPerWorkloadMetric{}.GetName(): getConstMetricMockResult(1000, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiGpuMemoryLimitBytesPerWorkloadMetric{}.GetName():     getConstMetricMockResult(1100, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuUsedCoresPerWorkloadMetric{}.GetName():            getDefaultMetricMockResult(50, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuRequestedCoresPerWorkloadMetric{}.GetName():       getConstMetricMockResult(1, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiCpuLimitCoresPerWorkloadMetric{}.GetName():           getConstMetricMockResult(2, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiMemoryUsedBytesPerWorkloadMetric{}.GetName():         getDefaultMetricMockResult(150, baseTimestampUnix, numOfSamples),
				runaimetrics.RunaiMemoryRequestedBytesPerWorkloadMetric{}.GetName():    getConstMetricMockResult(100, baseTimestampUnix, numOfSamples),
				// metrics.RunaiMemoryLimitPerWorkloadMetric{}.GetName():      getConstMetricMockResult(500, baseTimestampUnix, numOfSamples),
				// Dont have a mock for metrics.RunaiMemoryLimitPerWorkloadMetric{}.GetName() to test a case where no result returned

				runaimetrics.RunaiGpuUtilizationPerPodPerGpuMetric{}.GetName():     getDefaultMetricMockMultipleResults(60, baseTimestampUnix, numOfSamples, numOfMultipleResult),
				runaimetrics.RunaiGpuMemoryUsedBytesPerPodPerGpuMetric{}.GetName(): getDefaultMetricMockMultipleResults(700, baseTimestampUnix, numOfSamples, 2),
				runaimetrics.RunaiMemoryUsedBytesPerPodMetric{}.GetName():          getDefaultMetricMockResult(100, baseTimestampUnix, numOfSamples),
				// metrics.RunaiCpuUsagePerPodMetric{}.GetName():             getDefaultMetricMockResult(50, baseTimestampUnix, numOfSamples),
				// Dont have a mock for metrics.RunaiCpuUsagePerPodMetric{}.GetName() to test a case where no result returned

				// NIM workload histogram bucket metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric{}.GetName():  getBucketMetricMock(1711058503000, [][2]string{{"0.1", "1"}, {"+Inf", "5"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric{}.GetName(): getBucketMetricMock(1711058503000, [][2]string{{"0.5", "3"}, {"+Inf", "7"}}),
				// NIM workload histogram percentile metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric{}.GetName():  getPercentileMetricMock(1711158503000, [][2]string{{"50", "0.08"}, {"90", "0.20"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric{}.GetName(): getPercentileMetricMock(1711158503000, [][2]string{{"50", "1.2"}, {"99", "3.4"}}),

				// NIM pod histogram bucket metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric{}.GetName():  getBucketMetricMock(1711058503000, [][2]string{{"0.1", "1"}, {"+Inf", "5"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerPodMetric{}.GetName(): getBucketMetricMock(1711058503000, [][2]string{{"0.5", "3"}, {"+Inf", "7"}}),
				// NIM pod histogram percentile metrics
				runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric{}.GetName():  getPercentileMetricMock(1711158503000, [][2]string{{"50", "0.08"}, {"90", "0.20"}}),
				runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric{}.GetName(): getPercentileMetricMock(1711158503000, [][2]string{{"50", "1.2"}, {"99", "3.4"}}),

				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLME2ERequestLatencySecondsBucket}).GetName():    getBucketMetricMock(1712058503000, [][2]string{{"0.5", "4"}, {"+Inf", "9"}}),
				(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: runaimetrics.VLLMRequestQueueTimeSecondsPercentile}).GetName(): getPercentileMetricMock(1712158503000, [][2]string{{"50", "0.03"}, {"95", "0.10"}, {"99", "0.15"}}),
			})
			ctx = testing.SetContextWithAccept(testing.TestTenant, testing.TestUser, []string{}, CsvAccept)
		})

		BeforeEach(func() {
			// Mark workload as NIM-capable to allow NIM workload histogram CSV metrics
			w1 := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, testing.ClusterUuid1)
			w1.AdditionalFields = fmt.Sprintf(
				"{\"%s\":\"true\",\"%s\":\"true\"}",
				clusterCommonConstants.AdditionalNIMWorkload,
				clusterCommonConstants.AdditionalVLLMWorkload,
			)
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{w1},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		})

		AfterAll(func() {
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		})

		It("Get multiple metrics csv", func() {
			samples := 20
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1,
				[]api.WorkloadMetricType{api.WORKLOADMETRICTYPE_GPU_UTILIZATION, api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES, api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES, api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES},
				time.Now(), time.Now().Add(-5*time.Minute), int32(samples))
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data

			Expect(metrics[0][0]).To(Equal("Timestamp"))
			Expect(metrics[0][1]).To(Equal(string(api.WORKLOADMETRICTYPE_GPU_UTILIZATION)))
			Expect(metrics[0][2]).To(Equal(string(api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES)))
			Expect(metrics[0][3]).To(Equal(string(api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES)))
			Expect(metrics[0][4]).To(Equal(string(api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES)))

			Expect(metrics[1][0]).To(Equal(stringTimestamp(baseTimestampUnix)))
			Expect(metrics[2][0]).To(Equal(stringTimestamp(baseTimestampUnix + 1000)))
			Expect(metrics[3][0]).To(Equal(stringTimestamp(baseTimestampUnix + 2000)))
			Expect(metrics[9][0]).To(Equal(stringTimestamp(baseTimestampUnix + 8000)))
			Expect(metrics[10][0]).To(Equal(stringTimestamp(baseTimestampUnix + 9000)))

			Expect(metrics[1][1]).To(Equal("80"))
			Expect(metrics[2][1]).To(Equal("81"))
			Expect(metrics[3][1]).To(Equal("82"))
			Expect(metrics[9][1]).To(Equal("88"))
			Expect(metrics[10][1]).To(Equal("89"))

			Expect(metrics[1][2]).To(Equal("1000"))
			Expect(metrics[6][2]).To(Equal("1000"))
		})

		It("Get multiple pod metrics csv", func() {
			samples := 15
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES},
				time.Now(), time.Now().Add(-5*time.Minute), int32(samples))
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data

			Expect(metrics[0][0]).To(Equal("Timestamp"))
			Expect(metrics[0][1]).To(Equal(string(api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES)))

			Expect(metrics[1][0]).To(Equal(stringTimestamp(baseTimestampUnix)))
			Expect(metrics[2][0]).To(Equal(stringTimestamp(baseTimestampUnix + 1000)))

			Expect(metrics[1][1]).To(Equal("100"))
			Expect(metrics[2][1]).To(Equal("101"))
			Expect(metrics[9][1]).To(Equal("108"))
		})

		It("Get workload histogram metrics csv (NIM)", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1,
				[]api.WorkloadMetricType{api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS, api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS},
				time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data
			Expect(metrics[0][0]).To(Equal("Timestamp"))
			// Expect headers include expanded histogram columns
			Expect(metrics[0][1]).To(ContainSubstring(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)))
			Expect(metrics[0][2]).To(ContainSubstring(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)))
			Expect(metrics[0][len(metrics[0])-1]).To(ContainSubstring(string(api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS)))
		})

		It("Get workload histogram and percentile metrics csv (vLLM)", func() {
			requested := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS,
				api.WORKLOADMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES,
			}
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1, requested, time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			Expect(response.Data[0]).To(ContainElement("VLLM_E2E_REQUEST_LATENCY_SECONDS[+Inf]"))
			Expect(response.Data[0]).To(ContainElement("VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES[p50]"))
			Expect(response.Data[0]).To(ContainElement("VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES[p95]"))
			Expect(response.Data[0]).To(ContainElement("VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES[p99]"))
		})

		It("Get workload histogram bucket+percentile metrics csv (NIM) values", func() {
			resp, err := workloadsService.GetWorkloadMetrics(ctx, testing.TestUuid1,
				[]api.WorkloadMetricType{api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS, api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES},
				time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data
			// headers include bucket keys and percentile keys
			Expect(metrics[0]).To(ContainElement(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS) + "[0.1]"))
			Expect(metrics[0]).To(ContainElement(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS) + "[+Inf]"))
			Expect(metrics[0]).To(ContainElement(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES) + "[p50]"))
			Expect(metrics[0]).To(ContainElement(string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES) + "[p90]"))

			// row with bucket timestamp
			bucketTs := int64(1711058503000)
			rowIdx := -1
			for i := 1; i < len(metrics); i++ {
				if metrics[i][0] == stringTimestamp(bucketTs) {
					rowIdx = i
					break
				}
			}
			Expect(rowIdx).To(BeNumerically(">=", 1))
			// find column indices
			var idxBucket01, idxBucketInf int
			for i, h := range metrics[0] {
				if h == string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)+"[0.1]" {
					idxBucket01 = i
				}
				if h == string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)+"[+Inf]" {
					idxBucketInf = i
				}
			}
			Expect(metrics[rowIdx][idxBucket01]).To(Equal("1"))
			Expect(metrics[rowIdx][idxBucketInf]).To(Equal("5"))

			// row with percentile timestamp
			pctTs := int64(1711158503000)
			var pctRow int = -1
			for i := 1; i < len(metrics); i++ {
				if metrics[i][0] == stringTimestamp(pctTs) {
					pctRow = i
					break
				}
			}
			Expect(pctRow).To(BeNumerically(">=", 1))
			var idxP50, idxP90 int
			for i, h := range metrics[0] {
				if h == string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES)+"[p50]" {
					idxP50 = i
				}
				if h == string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES)+"[p90]" {
					idxP90 = i
				}
			}
			Expect(metrics[pctRow][idxP50]).To(Equal("0.08"))
			Expect(metrics[pctRow][idxP90]).To(Equal("0.20"))
		})

		It("Get pod histogram bucket+percentile metrics csv (NIM) values", func() {
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS, api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES},
				time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data
			// headers include bucket keys and percentile keys
			Expect(metrics[0]).To(ContainElement(string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS) + "[0.5]"))
			Expect(metrics[0]).To(ContainElement(string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS) + "[+Inf]"))
			Expect(metrics[0]).To(ContainElement(string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES) + "[p50]"))
			Expect(metrics[0]).To(ContainElement(string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES) + "[p99]"))

			// bucket row
			bucketTs := int64(1711058503000)
			rowIdx := -1
			for i := 1; i < len(metrics); i++ {
				if metrics[i][0] == stringTimestamp(bucketTs) {
					rowIdx = i
					break
				}
			}
			Expect(rowIdx).To(BeNumerically(">=", 1))
			var idxBucket05, idxBucketInf int
			for i, h := range metrics[0] {
				if h == string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS)+"[0.5]" {
					idxBucket05 = i
				}
				if h == string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS)+"[+Inf]" {
					idxBucketInf = i
				}
			}
			Expect(metrics[rowIdx][idxBucket05]).To(Equal("3"))
			Expect(metrics[rowIdx][idxBucketInf]).To(Equal("7"))

			// percentile row
			pctTs := int64(1711158503000)
			var pctRow int = -1
			for i := 1; i < len(metrics); i++ {
				if metrics[i][0] == stringTimestamp(pctTs) {
					pctRow = i
					break
				}
			}
			Expect(pctRow).To(BeNumerically(">=", 1))
			var idxP50, idxP99 int
			for i, h := range metrics[0] {
				if h == string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES)+"[p50]" {
					idxP50 = i
				}
				if h == string(api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES)+"[p99]" {
					idxP99 = i
				}
			}
			Expect(metrics[pctRow][idxP50]).To(Equal("1.2"))
			Expect(metrics[pctRow][idxP99]).To(Equal("3.4"))
		})

		It("Get pod histogram metrics csv (NIM)", func() {
			resp, err := podsService.GetWorkloadPodMetrics(ctx, testing.TestUuid1, testing.TestUuid4,
				[]api.PodMetricType{api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS, api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS},
				time.Now(), time.Now().Add(-5*time.Minute), 10)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data
			Expect(metrics[0][0]).To(Equal("Timestamp"))
			Expect(metrics[0][1]).To(ContainSubstring(string(api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS)))
			Expect(metrics[0][2]).To(ContainSubstring("+Inf"))
		})
	})

	Context("Environment Variable Redaction", func() {
		It("Should redact USERCREDENTIAL: environment variables for different users", func() {
			// Create a workload with sensitive environment variables
			workloadId := uuid.NewString()
			workload := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
			workload.SubmittedBy = "user123"
			workload.EnvironmentVariables = map[string]string{
				"NORMAL_VAR":           "normal_value",
				"USERCREDENTIAL:USER":  "USERCREDENTIAL:secret_password_123",
				"USERCREDENTIAL:TOKEN": "USERCREDENTIAL:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				"ANOTHER_VAR":          "another_value",
			}

			// Sync the workload
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			waitForGetWorkloadsLength(ctx, workloadsService, 4, []string{})

			// Test with different user (should redact)
			differentUserCtx := testing.SetContext(testing.TestTenant, "different_user", []string{})
			apiResponse, _ = workloadsService.GetWorkload(differentUserCtx, workloadId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Verify redaction - sensitive variables should be removed entirely
			Expect(workloadResponse.EnvironmentVariables["NORMAL_VAR"]).To(Equal("normal_value"))
			Expect(workloadResponse.EnvironmentVariables).NotTo(HaveKey("USERCREDENTIAL:USER"))
			Expect(workloadResponse.EnvironmentVariables).NotTo(HaveKey("USERCREDENTIAL:TOKEN"))
			Expect(workloadResponse.EnvironmentVariables["ANOTHER_VAR"]).To(Equal("another_value"))
		})

		It("Should NOT redact USERCREDENTIAL: environment variables for same user", func() {
			// Create a workload with sensitive environment variables
			workloadId := uuid.NewString()
			workload := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
			workload.SubmittedBy = "user123"
			workload.EnvironmentVariables = map[string]string{
				"NORMAL_VAR":           "normal_value",
				"USERCREDENTIAL:USER":  "USERCREDENTIAL:secret_password_123",
				"USERCREDENTIAL:TOKEN": "USERCREDENTIAL:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				"ANOTHER_VAR":          "another_value",
			}

			// Sync the workload
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			waitForGetWorkloadsLength(ctx, workloadsService, 4, []string{})

			// Test with same user (should NOT redact)
			sameUserCtx := testing.SetContext(testing.TestTenant, "user123", []string{})
			apiResponse, _ = workloadsService.GetWorkload(sameUserCtx, workloadId)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadResponse := httpResponseBodyToWorkloadResponse(apiResponse)

			// Verify no redaction - all variables should be present
			Expect(workloadResponse.EnvironmentVariables["NORMAL_VAR"]).To(Equal("normal_value"))
			Expect(workloadResponse.EnvironmentVariables["USERCREDENTIAL:USER"]).To(Equal("USERCREDENTIAL:secret_password_123"))
			Expect(workloadResponse.EnvironmentVariables["USERCREDENTIAL:TOKEN"]).To(Equal("USERCREDENTIAL:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."))
			Expect(workloadResponse.EnvironmentVariables["ANOTHER_VAR"]).To(Equal("another_value"))
		})

		It("Should redact USERCREDENTIAL: environment variables in GetWorkloads for different users", func() {
			// Create workloads with sensitive environment variables
			workloadId1 := uuid.NewString()
			workloadId2 := uuid.NewString()

			workload1 := SetupBaseWorkload(workloadId1, DefaultProjectId, testing.ClusterUuid1)
			workload1.SubmittedBy = "user1"
			workload1.EnvironmentVariables = map[string]string{
				"USERCREDENTIAL:USER": "USERCREDENTIAL:secret1",
				"NORMAL_VAR":          "normal1",
			}

			workload2 := SetupBaseWorkload(workloadId2, DefaultProjectId, testing.ClusterUuid1)
			workload2.SubmittedBy = "user2"
			workload2.EnvironmentVariables = map[string]string{
				"USERCREDENTIAL:TOKEN": "USERCREDENTIAL:secret2",
				"NORMAL_VAR":           "normal2",
			}

			// Sync the workloads
			apiResponse, _ := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
				ProjectId: DefaultProjectId,
				Workloads: []api.WorkloadSyncFields{workload1, workload2},
			})
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
			waitForGetWorkloadsLength(ctx, workloadsService, 5, []string{})

			// Test with user1 (should redact user2's workload but not user1's)
			user1Ctx := testing.SetContext(testing.TestTenant, "user1", []string{})
			apiResponse, _ = workloadsService.GetWorkloads(user1Ctx, false, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", DefaultProjectId)}, "")
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)

			// Find our test workloads
			var testWorkload1, testWorkload2 *api.Workload
			for i := range workloadsResponse.Workloads {
				workload := &workloadsResponse.Workloads[i]
				if workload.Id == workloadId1 {
					testWorkload1 = workload
				} else if workload.Id == workloadId2 {
					testWorkload2 = workload
				}
			}

			Expect(testWorkload1).NotTo(BeNil())
			Expect(testWorkload2).NotTo(BeNil())

			// Verify redaction for user1's workload (should NOT be redacted)
			Expect(testWorkload1.EnvironmentVariables["USERCREDENTIAL:USER"]).To(Equal("USERCREDENTIAL:secret1"))
			Expect(testWorkload1.EnvironmentVariables["NORMAL_VAR"]).To(Equal("normal1"))

			// Verify redaction for user2's workload (should be redacted - removed entirely)
			Expect(testWorkload2.EnvironmentVariables).NotTo(HaveKey("USERCREDENTIAL:TOKEN"))
			Expect(testWorkload2.EnvironmentVariables["NORMAL_VAR"]).To(Equal("normal2"))
		})
	})
})

func validateEndpoints(endpoints []api.Endpoint) {
	Expect(endpoints).ToNot(BeNil())
	Expect(len(endpoints)).To(Equal(3))

	ep1 := endpoints[0]
	Expect(ep1.Host).To(Equal("chat.example.com"))
	Expect(ep1.Port).To(Equal(int32(8080)))
	Expect(ep1.Path).To(Equal("/v1/completions"))
	Expect(*ep1.TlsEnabled).To(BeTrue())
	Expect(ep1.Protocol).To(Equal("HTTPS"))
	Expect(ep1.Url).To(Equal("https://chat.example.com:8080/v1/completions"))

	ep2 := endpoints[1]
	Expect(ep2.Host).To(Equal("api.example.com"))
	Expect(ep2.Port).To(Equal(int32(443)))
	Expect(ep2.Path).To(Equal("/health"))
	Expect(ep2.Protocol).To(Equal("HTTPS"))
	Expect(ep2.TlsEnabled).To(BeNil())

	ep3 := endpoints[2]
	Expect(ep3.Host).To(Equal("api.example2.com"))
	Expect(ep3.Port).To(Equal(int32(80)))
	Expect(ep3.Path).To(Equal("/health"))
	Expect(ep3.Protocol).To(Equal("HTTP"))
	Expect(ep3.Url).To(Equal("http://api.example2.com/health"))
	Expect(*ep3.TlsEnabled).To(BeFalse())
}

func stringTimestamp(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).String()
}

func testWorkloadsFilter(ctx context.Context, workloadsService api.WorkloadsAPIServicer, sortBy string, filters []string, expectedCount int) api.GetWorkloads200Response {
	apiResponse, _ := workloadsService.CountWorkloads(ctx, false, filters, "", true)
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
	Expect(countResponse.Count).To(Equal(int64(expectedCount)))

	var workloadsResponse api.GetWorkloads200Response
	Eventually(func(g Gomega) {
		apiResponse, _ := workloadsService.GetWorkloads(ctx, false, 0, 0, "ASC", sortBy, filters, "")
		g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
		workloadsResponse = httpResponseBodyToWorkloadsResponse(apiResponse)
		g.Expect(len(workloadsResponse.Workloads)).To(Equal(expectedCount))
	}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())

	apiResponse, _ = workloadsService.CountWorkloads(ctx, false, filters, "", false)
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	countResponse = httpResponseBodyToWorkloadsCountResponse(apiResponse)
	Expect(countResponse.Count).To(Equal(int64(expectedCount)))

	return workloadsResponse
}

func testWorkloadsSearch(ctx context.Context, workloadsService api.WorkloadsAPIServicer, search string, expectedCount int) api.GetWorkloads200Response {
	apiResponse, _ := workloadsService.CountWorkloads(ctx, false, []string{}, search, true)
	ExpectWithOffset(1, apiResponse.Code).To(Equal(http.StatusOK))
	countResponse := httpResponseBodyToWorkloadsCountResponse(apiResponse)
	ExpectWithOffset(1, countResponse.Count).To(Equal(int64(expectedCount)))

	var workloadsResponse api.GetWorkloads200Response
	Eventually(func(g Gomega) {
		apiResponse, _ := workloadsService.GetWorkloads(ctx, false, 0, 0, "ASC", "", []string{}, search)
		g.ExpectWithOffset(1, apiResponse.Code).To(Equal(http.StatusOK))
		workloadsResponse = httpResponseBodyToWorkloadsResponse(apiResponse)
		g.ExpectWithOffset(1, len(workloadsResponse.Workloads)).To(Equal(expectedCount))
	}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())

	apiResponse, _ = workloadsService.CountWorkloads(ctx, false, []string{}, search, false)
	ExpectWithOffset(1, apiResponse.Code).To(Equal(http.StatusOK))
	countResponse = httpResponseBodyToWorkloadsCountResponse(apiResponse)
	ExpectWithOffset(1, countResponse.Count).To(Equal(int64(expectedCount)))

	return workloadsResponse
}

func setupInternalWorkloadCreationFields(id string) api.WorkloadCreationFields {
	return api.WorkloadCreationFields{
		Type:               DefaultType,
		Name:               fmt.Sprintf("%s-%s", DefaultWorkloadName, rand.String(4)),
		Id:                 id,
		SubmittedBy:        DefaultSubmittedBy,
		ClusterId:          testing.ClusterUuid1,
		ProjectId:          DefaultProjectId,
		CreatedAt:          DefaultCreationTime,
		RequestedResources: nil,
		RequestedPods:      api.RequestedPods{Number: ptr.To(int32(DefaultNumberOfRequestedPods))},
		RequestedNodePools: DefaultRequestedNodePools,
		Datasources: []api.Datasource{{
			Type: "Storage",
			Name: "S3",
			Id:   testing.TestUuid2,
		}},
		Environments: []api.Environment{
			{
				Connections: []api.Connection{{Name: "my-pytorch-env-1", ToolType: "pytorch", AuthorizedUsers: []string{"adam@company.ai", "eve@company.ai"}}},
				Name:        "pytorch env",
				Id:          testing.TestUuid3,
				ReplicaType: (*api.ReplicaType)(ptr.To(string(api.REPLICATYPE_MASTER))),
			},
			{
				Connections: []api.Connection{{Name: "my-pytorch-env-2", ToolType: "pytorch"}},
				Name:        "pytorch env",
				Id:          testing.TestUuid3,
				ReplicaType: nil,
			},
		},
		ExternalConnections:  []api.Connection{{Name: "weights and biases", ToolType: "wandb", Url: "www.wandb.com"}},
		DistributedFramework: "pytorch",
		Images:               DefaultImages,
		SourceApi:            ptr.To(api.SOURCEAPI_WORKLOADS_V2),
	}
}

func setupWorkloadConnectionMetadata(suffix string) WorkloadConnectionMetadata {
	return WorkloadConnectionMetadata{
		Environments: []api.Environment{
			{
				Connections: []api.Connection{
					{
						Name:                "Jupyter",
						ToolType:            "jupyter-notebook",
						ConnectionType:      "NodePort",
						ContainerPort:       8888,
						PortResourceName:    ptr.To("port-8888"),
						ServiceResourceName: ptr.To(fmt.Sprintf("job-test-%s-np", suffix)),
					},
				},
				Name: "env-" + suffix,
				Id:   "env-id-" + suffix,
			},
		},
		ExternalConnections: []api.Connection{
			{
				Name:                "TensorBoard",
				ToolType:            "tensorboard",
				ConnectionType:      "ExternalUrl",
				ExposedUrlPath:      ptr.To(fmt.Sprintf("/p1/test-%s", suffix)),
				ServiceResourceName: ptr.To(fmt.Sprintf("job-test-%s-url", suffix)),
			},
		},
	}
}

func httpResponseBodyToWorkloadsResponse(httpResponse api.ImplResponse) api.GetWorkloads200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.GetWorkloads200Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func httpResponseBodyToWorkloadsCountResponse(httpResponse api.ImplResponse) api.CountWorkloads200Response {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.CountWorkloads200Response{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func httpResponseBodyToWorkloadResponse(httpResponse api.ImplResponse) api.WorkloadDetailed {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.WorkloadDetailed{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func getConstMetricMockResult(baseValue int, baseTimestampUnix float64, numOfSamples int) openapi.GetMetrics200Response {
	return getMetricMockResultInner(baseValue, baseTimestampUnix, true, numOfSamples, 1, 1)
}

func getDefaultMetricMockResult(baseValue int, baseTimestampUnix float64, numOfSamples int) openapi.GetMetrics200Response {
	return getMetricMockResultInner(baseValue, baseTimestampUnix, false, numOfSamples, 1, 1)
}

// Helpers for histogram-style mocks used in NIM tests
func getBucketMetricMock(baseTimestampUnix int64, buckets [][2]string) openapi.GetMetrics200Response {
	data := make([]openapi.Result, 0, len(buckets))
	for _, b := range buckets {
		data = append(data, openapi.Result{
			MetricLabels: map[string]string{"le": b[0]},
			Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(baseTimestampUnix), Value: b[1]}},
		})
	}
	return openapi.GetMetrics200Response{Status: "success", IsRange: true, Data: data}
}

func getPercentileMetricMock(baseTimestampUnix int64, pairs [][2]string) openapi.GetMetrics200Response {
	data := make([]openapi.Result, 0, len(pairs))
	for _, p := range pairs {
		data = append(data, openapi.Result{
			MetricLabels: map[string]string{"quantile_label": p[0]},
			Measurements: []openapi.Measurement{{Timestamp: time.UnixMilli(baseTimestampUnix), Value: p[1]}},
		})
	}
	return openapi.GetMetrics200Response{Status: "success", IsRange: true, Data: data}
}

func getDefaultMetricMockMultipleResults(baseValue int, baseTimestampUnix float64, numOfSamples, numberOfResults int) openapi.GetMetrics200Response {
	return getMetricMockResultInner(baseValue, baseTimestampUnix, false, numOfSamples, numberOfResults, 1)
}

func getDefaultMetricMockMultipleNodePools(baseValue int, baseTimestampUnix float64, numOfSamples, numberOfResults int, numberOfNodePools int) openapi.GetMetrics200Response {
	return getMetricMockResultInner(baseValue, baseTimestampUnix, false, numOfSamples, numberOfResults, numberOfNodePools)
}

func getMetricMockResultInner(baseValue int, baseTimestampUnix float64, constValues bool, numOfSamples int, numberOfResults int, numberOfNodePools int) openapi.GetMetrics200Response {
	baseValueStr := fmt.Sprintf("%d", baseValue)
	measurements := []openapi.Measurement{}

	for i := 0; i < numOfSamples; i++ {
		value := baseValueStr
		if !constValues {
			value = fmt.Sprintf("%d", baseValue+i)
		}
		measurements = append(measurements, openapi.Measurement{
			Timestamp: time.UnixMilli(int64(baseTimestampUnix) + int64(i*1000)),
			Value:     value,
		})
	}

	data := []openapi.Result{}
	for i := 0; i < numberOfResults; i++ {
		data = append(data, openapi.Result{
			Measurements: measurements,
			MetricLabels: map[string]string{"gpu": fmt.Sprintf("%d", i), nodepoolLabel: fmt.Sprintf("nodepool%d", i%numberOfNodePools)},
		})
	}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data:    data,
	}
}

func getMetricResponse(workloadIds []string, metricValues []string) openapi.GetMetrics200Response {
	results := []openapi.Result{}
	for i := range workloadIds {
		measurements := []openapi.Measurement{
			{
				Timestamp: time.UnixMilli(1708958503653),
				Value:     metricValues[i],
			},
		}
		labels := map[string]string{
			"workload_id": workloadIds[i],
		}
		results = append(results, openapi.Result{
			Measurements: measurements,
			MetricLabels: labels,
		})

	}
	return openapi.GetMetrics200Response{
		Status: "success",
		Data:   results,
	}
}

func waitForGetWorkloadsLength(ctx context.Context, workloadsService api.WorkloadsAPIServicer, expectedLength int,
	filterBy []string,
) {
	EventuallyWithOffset(1, func(g Gomega) {
		apiResponse, err := workloadsService.GetWorkloads(ctx, false, 0, 0, "", "", filterBy, "")
		g.Expect(err).To(BeNil())
		g.Expect(apiResponse.Code).To(Equal(http.StatusOK))
		workloadsResponse := httpResponseBodyToWorkloadsResponse(apiResponse)
		g.Expect(len(workloadsResponse.Workloads)).To(Equal(expectedLength))
		g.Expect(workloadsResponse.Next).To(Equal(int32(0)))
	}, TimeoutWhenWaitingForView, IntervalWhenWaitingForView).Should(Succeed())
}

func validateWorkloadTypeConfig(actual, expected api.WorkloadTypeConfig) {
	ExpectWithOffset(1, actual.CategoryId).To(Equal(expected.CategoryId), "CategoryId")
	ExpectWithOffset(1, actual.PriorityId).To(Equal(expected.PriorityId), "PriorityId")
	ExpectWithOffset(1, actual.Name).To(Equal(expected.Name), "Name")
	ExpectWithOffset(1, actual.CategoryName).To(Equal(expected.CategoryName), "CategoryName")
	ExpectWithOffset(1, actual.PriorityName).To(Equal(expected.PriorityName), "PriorityName")
	ExpectWithOffset(1, actual.CreatedBy).To(Equal(expected.CreatedBy), "CreatedBy")
	ExpectWithOffset(1, actual.Group).To(Equal(expected.Group), "Group")
	ExpectWithOffset(1, actual.Source).To(Equal(expected.Source), "Source")
	ExpectWithOffset(1, actual.Preemptibility).To(Equal(expected.Preemptibility), "Preemptibility")

	ExpectWithOffset(1, len(actual.ResourceInterfaces)).To(Equal(len(expected.ResourceInterfaces)), "ResourceInterfaces length")
	for i := range actual.ResourceInterfaces {
		actualResourceInterfaceWithVersion := actual.ResourceInterfaces[i]
		expectedResourceInterfaceWithVersion := expected.ResourceInterfaces[i]

		ExpectWithOffset(1, actualResourceInterfaceWithVersion.Version).To(Equal(expectedResourceInterfaceWithVersion.Version), "ResourceInterfaces[%d] Version", i)

		actualResourceInterface := actualResourceInterfaceWithVersion.ResourceInterface
		expectedResourceInterface := expectedResourceInterfaceWithVersion.ResourceInterface

		ExpectWithOffset(1, len(actualResourceInterface)).To(Equal(len(expectedResourceInterface)), "ResourceInterfaces[%d] length", i)
		for key, actualValue := range actualResourceInterface {
			expectedValue, ok := expectedResourceInterface[key]
			ExpectWithOffset(1, ok).To(BeTrue(), "ResourceInterfaces[%d] missing key %s", i, key)
			ExpectWithOffset(1, actualValue).To(Equal(expectedValue), "ResourceInterfaces[%d][%s]", i, key)
		}
	}
}

func findCategoryByName(categories []api.WorkloadCategory, name string) *api.WorkloadCategory {
	for _, category := range categories {
		if category.Name == name {
			return &category
		}
	}
	Fail("Category not found: " + name)
	return nil
}

func findPriorityByName(priorities []api.WorkloadPriority, name string) *api.WorkloadPriority {
	for _, priority := range priorities {
		if priority.Name == name {
			return &priority
		}
	}
	Fail("Priority not found: " + name)
	return nil
}

func findRunaiWorkloadTypes(workloadTypes []api.WorkloadTypeConfig) (*api.WorkloadTypeConfig, *api.WorkloadTypeConfig, *api.WorkloadTypeConfig, *api.WorkloadTypeConfig) {
	var training *api.WorkloadTypeConfig = nil
	var dist *api.WorkloadTypeConfig = nil
	var interactive *api.WorkloadTypeConfig = nil
	var inference *api.WorkloadTypeConfig = nil
	for i := range workloadTypes {
		workloadType := &workloadTypes[i]
		if workloadType.Name == "TrainingWorkload" {
			training = workloadType
		} else if workloadType.Name == "DistributedWorkload" {
			dist = workloadType
		} else if workloadType.Name == "InteractiveWorkload" {
			interactive = workloadType
		} else if workloadType.Name == "InferenceWorkload" {
			inference = workloadType
		}
	}

	return training, dist, interactive, inference
}

// createWorkloadTypeConfigForKartaLookup creates the config row needed by external_workloads_type_config_view.
func createWorkloadTypeConfigForKartaLookup(workloadTypeId uuid.UUID, scopeType string, tenantId *string) {
	var category models.WorkloadsCategory
	ExpectWithOffset(1, runaidb.DB().
		Where("name = ?", "Train").
		Where("scope_type = ?", string(api.SCOPETYPE_SYSTEM)).
		First(&category)).To(BeNil())

	var priority models.Priority
	ExpectWithOffset(1, runaidb.DB().
		Where("name = ?", "low").
		Where("scope_type = ?", string(api.SCOPETYPE_SYSTEM)).
		First(&priority)).To(BeNil())

	now := time.Now()
	config := models.ExternalWorkloadTypeConfig{
		ID:             uuid.New(),
		TenantId:       tenantId,
		ScopeType:      scopeType,
		WorkloadTypeId: workloadTypeId,
		CategoryID:     ptr.To(category.ID),
		PriorityID:     ptr.To(priority.ID),
		CreatedAt:      now,
		CreatedBy:      ptr.To(testing.TestUser),
		UpdatedAt:      now,
		UpdatedBy:      ptr.To(testing.TestUser),
	}
	_, err := runaidb.DB().Create(&config)
	ExpectWithOffset(1, err).To(BeNil())
}

// buildTestKartaWithSuspend returns a minimal models.Karta with a non-nil SuspendDefinition
// so that isSuspendSupported returns true for that workload type.
func buildTestKartaWithSuspend(kind string) models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    kind,
					},
					SuspendDefinition: &kartav1alpha1.SuspendDefinition{
						SuspendActions: []kartav1alpha1.SuspendAction{
							{Path: ".spec.suspend", Value: "true"},
						},
						ResumeActions: []kartav1alpha1.SuspendAction{
							{Path: ".spec.suspend", Value: "false"},
						},
					},
				},
			},
		},
	}
}
