/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/nvidia"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/go_migrations/ri/source/kubernetes"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/policy"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

var _ = Describe("Workloads types API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadPropertiesService := NewWorkloadPropertiesService()

	BeforeAll(func() {
		policy.SetMock(policy.NewSimpleMock())
	})

	BeforeEach(func() {
		ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("Categories", func() {
		It("ListCategories - Get all categories", func() {
			ctx := testing.SetSystemContext(testing.TestUser)
			apiResponse, err := workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			response, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			Expect(len(response.Categories)).To(BeNumerically(">=", 3)) // At least the system categories
		})

		It("GetCategoryById - Get existing category", func() {
			ctx := testing.SetSystemContext(testing.TestUser)
			// First get all categories to find a valid ID
			apiResponse, err := workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			response, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			Expect(len(response.Categories)).To(BeNumerically(">=", 1))

			// Get the first category by ID
			categoryId := response.Categories[0].Id
			apiResponse, err = workloadPropertiesService.GetCategoryById(ctx, categoryId)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			category, ok := apiResponse.Body.(*api.WorkloadCategory)
			Expect(ok).To(BeTrue())
			Expect(category.Id).To(Equal(categoryId))
			Expect(category.Name).To(Equal(response.Categories[0].Name))
			Expect(category.Description).To(Equal(response.Categories[0].Description))
		})

		It("GetCategoryById - Non-existent category", func() {
			nonExistentId := uuid.NewString()
			apiResponse, err := workloadPropertiesService.GetCategoryById(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Priorities", Ordered, func() {
		predefinedPriorities := []api.WorkloadPriority{
			{
				Source: RunaiSource,
				Name:   "high",
				Value:  125,
			},
			{
				Source: RunaiSource,
				Name:   "low",
				Value:  40,
			},
			{
				Source: RunaiSource,
				Name:   "medium",
				Value:  80,
			},
			{
				Source: RunaiSource,
				Name:   "medium-high",
				Value:  90,
			},
			{
				Source: RunaiSource,
				Name:   "medium-low",
				Value:  65,
			},
			{
				Source: RunaiSource,
				Name:   "very-high",
				Value:  150,
			},
			{
				Source: RunaiSource,
				Name:   "very-low",
				Value:  25,
			},
		}
		predefinedPrioritiesMap := map[string]api.SyncPriority{
			"high": {
				Source: RunaiSource,
				Name:   "high",
				Value:  125,
			},
			"low": {
				Source: RunaiSource,
				Name:   "low",
				Value:  40,
			},
			"medium": {
				Source: RunaiSource,
				Name:   "medium",
				Value:  80,
			},
			"medium-high": {
				Source: RunaiSource,
				Name:   "medium-high",
				Value:  90,
			},
			"medium-low": {
				Source: RunaiSource,
				Name:   "medium-low",
				Value:  65,
			},
			"very-high": {
				Source: RunaiSource,
				Name:   "very-high",
				Value:  150,
			},
			"very-low": {
				Source: RunaiSource,
				Name:   "very-low",
				Value:  25,
			},
		}

		AfterAll(func() {
			testing.CleanTables()
		})

		Context("Get priorities", func() {
			It("get all priorities", func() {
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "name", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(7))
				for i := range priorities {
					Expect(priorities[i].Name).To(Equal(predefinedPriorities[i].Name))
					Expect(priorities[i].Value).To(Equal(predefinedPriorities[i].Value))
					Expect(priorities[i].Source).To(Equal(predefinedPriorities[i].Source))
					Expect(priorities[i].AggregatedPhase).To(Equal(StatusReady))
				}
			})

			It("get with filter", func() {
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{"name==medium-low"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Name).To(Equal("medium-low"))
				Expect(priorities[0].Value).To(Equal(int32(65)))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[0].AggregatedPhase).To(Equal(StatusReady))
			})
		})

		Context("Get priorities with status", func() {
			clusterId1, _ := uuid.NewUUID()
			clusterId2, _ := uuid.NewUUID()
			clusterId3, _ := uuid.NewUUID()
			var (
				priority1 api.WorkloadPriority
				priority2 api.WorkloadPriority
			)
			BeforeAll(func() {
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "name", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities

				priority1 = priorities[1]
				priority2 = priorities[2]

				query := "INSERT INTO priority_statuses (\"id\",\"priority_id\", \"cluster_id\", \"phase\") VALUES"
				valuePriority1Cluster1 := fmt.Sprintf("'%s', '%s', '%s', '%s'", uuid.New(), priority1.Id, clusterId1, StatusCreating)
				valuePriority1Cluster2 := fmt.Sprintf("'%s', '%s', '%s', '%s'", uuid.New(), priority1.Id, clusterId2, StatusFailed)
				valuePriority2Cluster1 := fmt.Sprintf("'%s', '%s', '%s', '%s'", uuid.New(), priority2.Id, clusterId1, StatusReady)
				rows, err := runaidb.DB().Raw(fmt.Sprintf("%s (%s), (%s), (%s);", query, valuePriority1Cluster1, valuePriority2Cluster1, valuePriority1Cluster2)).Rows()
				Expect(err).To(BeNil())
				Expect(rows.Err()).To(BeNil())
			})

			It("get with filter", func() {
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{"name==low"}) // priority1
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Name).To(Equal("low"))
				Expect(priorities[0].Value).To(Equal(int32(40)))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[0].AggregatedPhase).To(Equal(StatusFailed))
				Expect(len(priorities[0].Status)).To(Equal(2))
				clusterId := priorities[0].Status[0].ClusterId
				if clusterId == clusterId1.String() {
					Expect(priorities[0].Status[0].Phase).To(Equal(StatusCreating))
				} else if clusterId == clusterId2.String() {
					Expect(priorities[0].Status[0].Phase).To(Equal(StatusFailed))
				} else {
					Fail(fmt.Sprintf("clusterId %s invalid", clusterId))
				}
				clusterId = priorities[0].Status[1].ClusterId
				if clusterId == clusterId1.String() {
					Expect(priorities[0].Status[1].Phase).To(Equal(StatusCreating))
				} else if clusterId == clusterId2.String() {
					Expect(priorities[0].Status[1].Phase).To(Equal(StatusFailed))
				} else {
					Fail(fmt.Sprintf("clusterId %s invalid", clusterId))
				}

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{"name==medium"}) // priority2
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Name).To(Equal("medium"))
				Expect(priorities[0].Value).To(Equal(int32(80)))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[0].AggregatedPhase).To(Equal(StatusReady))
				Expect(len(priorities[0].Status)).To(Equal(1))
				Expect(priorities[0].Status[0].ClusterId).To(Equal(clusterId1.String()))
				Expect(priorities[0].Status[0].Phase).To(Equal(StatusReady))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "value", []string{"value==150"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Value).To(BeEquivalentTo(150))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "source", []string{"source==Run.ai"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(7))
				for _, p := range priorities {
					Expect(p.Source).To(Equal(RunaiSource))
				}
			})

			It("get cluster with 2 priorities", func() {
				response, err := workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId1.String(), "", "name", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(2))

				Expect(priorities[0].Name).To(Equal("low"))
				Expect(priorities[0].Value).To(Equal(int32(40)))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[0].Phase).To(Equal(StatusCreating))

				Expect(priorities[1].Name).To(Equal("medium"))
				Expect(priorities[1].Value).To(Equal(int32(80)))
				Expect(priorities[1].Source).To(Equal(RunaiSource))
				Expect(priorities[1].Phase).To(Equal(StatusReady))
			})

			It("get cluster with 1 priority", func() {
				response, err := workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId2.String(), "", "name", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))

				Expect(priorities[0].Name).To(Equal("low"))
				Expect(priorities[0].Value).To(Equal(int32(40)))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[0].Phase).To(Equal(StatusFailed))
			})

			It("get cluster with no priorities", func() {
				response, err := workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId3.String(), "", "name", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(0))

				// test db func as well
				dbPriorities, err := dao.GetClusterPriorities(ctx, testing.TestTenant, clusterId3.String(), "", "", []queryapi.Filter{})
				Expect(err).To(BeNil())
				Expect(len(*dbPriorities)).To(Equal(0))
			})

			It("get cluster priorities with filters", func() {
				response, err := workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId1.String(), "", "name", []string{"name==low"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities := response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Name).To(Equal("low"))

				response, err = workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId1.String(), "", "source", []string{"source==Run.ai"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(2))
				Expect(priorities[0].Source).To(Equal(RunaiSource))
				Expect(priorities[1].Source).To(Equal(RunaiSource))

				response, err = workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId1.String(), "", "value", []string{"value!=40"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(1))
				Expect(priorities[0].Value).ToNot(BeEquivalentTo(40))

				response, err = workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId1.String(), "", "phase", []string{"phase!=Updating"})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				priorities = response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				Expect(len(priorities)).To(Equal(2))
				Expect(priorities[0].Phase).ToNot(Equal("Updating"))
				Expect(priorities[1].Phase).ToNot(Equal("Updating"))
				Expect(priorities[1].Phase > priorities[0].Phase).To(BeTrue())
			})
		})

		Context("sync to/from priorities", func() {
			clusterId, _ := uuid.NewUUID()
			clusterId2, _ := uuid.NewUUID()

			It("sync flow", func() {
				var lowPriorityId string
				var mediumPriorityId string
				syncResponse, err := workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities := syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities)))
				for _, priority := range priorities {
					basePriority, found := predefinedPrioritiesMap[priority.Name]
					Expect(found).To(BeTrue())
					Expect(basePriority.Name).To(Equal(priority.Name))
					Expect(basePriority.Value).To(Equal(priority.Value))
					Expect(basePriority.Source).To(Equal(priority.Source))
					if priority.Name == "low" {
						lowPriorityId = priority.Id
					}
					if priority.Name == "medium" {
						mediumPriorityId = priority.Id
					}
				}

				statusResponse, err := workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}, {
						Id:     mediumPriorityId,
						Name:   "medium",
						Value:  80,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Not Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						status := priority.Status
						foundCluster := false
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Ready"))
								Expect(foundCluster).To(BeFalse())
								foundCluster = true
							}
						}
					}
					if priority.Name == "medium" {
						status := priority.Status
						foundCluster := false
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
								Expect(foundCluster).To(BeFalse())
								foundCluster = true
							}
						}
					}
				}

				statusResponse, err = workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Not Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						status := priority.Status
						foundCluster := false
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
								Expect(foundCluster).To(BeFalse())
								foundCluster = true
							}
						}
					}
					if priority.Name == "medium" {
						status := priority.Status
						foundCluster := false
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
								Expect(foundCluster).To(BeFalse())
								foundCluster = true
							}
						}
					}
				}
			})

			It("sync new cluster priority", func() {
				var lowPriorityId string
				var mediumPriorityId string
				syncResponse, err := workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities := syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities)))
				for _, priority := range priorities {
					basePriority, found := predefinedPrioritiesMap[priority.Name]
					Expect(found).To(BeTrue())
					Expect(basePriority.Name).To(Equal(priority.Name))
					Expect(basePriority.Value).To(Equal(priority.Value))
					Expect(basePriority.Source).To(Equal(priority.Source))
					if priority.Name == "low" {
						lowPriorityId = priority.Id
					}
					if priority.Name == "medium" {
						mediumPriorityId = priority.Id
					}
				}

				statusResponse, err := workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}, {
						Id:     mediumPriorityId,
						Name:   "medium",
						Value:  80,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Not Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  800,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))
				// update the status
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  801,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				// cluster defined priorities should be returned for user
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				lowFound := false
				mediumFound := false
				clusterDefinedFound := false
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						lowFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Ready"))
							}
						}
					}
					if priority.Name == "medium" {
						mediumFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
							}
						}
					}
					if priority.Name == "legacy" {
						clusterDefinedFound = true
						Expect(priority.Value).To(Equal(int32(801)))
						Expect(priority.Source).To(Equal(""))
						status := priority.Status
						Expect(len(status)).To(Equal(1))
						Expect(status[0].Phase).To(Equal("Ready"))
					}
				}
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeTrue())
				Expect(clusterDefinedFound).To(BeTrue())

				// cluster defined priorities should not be returned for agent
				syncResponse, err = workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities = syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities)))
				clusterDefinedFound = false
				for _, priority := range priorities {
					if priority.Name == "legacy" {
						clusterDefinedFound = true
					}
				}
				Expect(clusterDefinedFound).To(BeFalse())

				// delete cluster defined priority
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync:   true,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				lowFound = false
				mediumFound = false
				clusterDefinedFound = false
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						lowFound = true
					}
					if priority.Name == "medium" {
						mediumFound = true
					}
					if priority.Name == "legacy" {
						clusterDefinedFound = true
					}
				}
				// only cluster defined are deleted
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeTrue())
				Expect(clusterDefinedFound).To(BeFalse())
			})

			It("delete priorities", func() {
				var lowPriorityId string
				var mediumPriorityId string
				syncResponse, err := workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities := syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities)))
				for _, priority := range priorities {
					basePriority, found := predefinedPrioritiesMap[priority.Name]
					Expect(found).To(BeTrue())
					Expect(basePriority.Name).To(Equal(priority.Name))
					Expect(basePriority.Value).To(Equal(priority.Value))
					Expect(basePriority.Source).To(Equal(priority.Source))
					if priority.Name == "low" {
						lowPriorityId = priority.Id
					}
					if priority.Name == "medium" {
						mediumPriorityId = priority.Id
					}
				}

				statusResponse, err := workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}, {
						Id:     mediumPriorityId,
						Name:   "medium",
						Value:  80,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Not Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  800,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				// cluster defined priorities should be returned for user
				response, err := workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId.String(), "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities := response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				lowFound := false
				mediumFound := false
				clusterDefinedFound := false
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						lowFound = true
					}
					if priority.Name == "medium" {
						mediumFound = true
					}
					if priority.Name == "legacy" {
						clusterDefinedFound = true
					}
				}
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeTrue())
				Expect(clusterDefinedFound).To(BeTrue())

				statusResponse, err = workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}, {
						Id:     mediumPriorityId,
						Name:   "medium",
						Value:  80,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: DeletedPhase},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  800,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: DeletedPhase},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				// only "low" priority should be returned for get cluster
				response, err = workloadPropertiesService.GetWorkloadClusterPriorities(ctx, clusterId.String(), "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadClusterPriorities200Response).Priorities
				lowFound = false
				mediumFound = false
				clusterDefinedFound = false
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						lowFound = true
					}
					if priority.Name == "medium" {
						mediumFound = true
					}
					if priority.Name == "legacy" {
						clusterDefinedFound = true
					}
				}
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeFalse()) //
				Expect(clusterDefinedFound).To(BeFalse())

				// only "low"+ medium priority should be returned for get all
				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				tenantPriorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				lowFound = false
				mediumFound = false
				clusterDefinedFound = false
				for _, priority := range tenantPriorities {
					if priority.Name == "low" {
						lowFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Ready"))
							}
						}
					}
					if priority.Name == "medium" {
						mediumFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
							}
						}
					}
					if priority.Name == "legacy" {
						clusterDefinedFound = true
						Expect(priority.Value).To(Equal(int32(800)))
						Expect(priority.Source).To(Equal(""))
						status := priority.Status
						Expect(len(status)).To(Equal(1))
						Expect(status[0].Phase).To(Equal("Ready"))
					}
				}
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeTrue())
				Expect(clusterDefinedFound).To(BeFalse())

			})

			It("sync deleted priority", func() {

				newPriorityName := "newPriority"
				query := "INSERT INTO priorities (\"id\",\"name\", \"value\", \"source\", \"scope_type\", \"scope_id\", \"tenant_id\") VALUES"
				valuePriority := fmt.Sprintf("'%s', '%s', '%s', '%s', '%s', '%s', '%s'", uuid.New(), newPriorityName, "1212", RunaiSource, "tenant", testing.TestTenant, testing.TestTenant)
				rows, err := runaidb.DB().Raw(fmt.Sprintf("%s (%s);", query, valuePriority)).Rows()
				Expect(err).To(BeNil())
				Expect(rows.Err()).To(BeNil())

				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				var newPriority *api.WorkloadPriority = nil
				for i := range clusterPriorities {
					if clusterPriorities[i].Name == newPriorityName {
						newPriority = &clusterPriorities[i]
					}
				}
				Expect(newPriority).NotTo(BeNil())

				syncResponse, err := workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities := syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities) + 1))
				var newSyncPriority *api.SyncPriority = nil
				for i := range priorities {
					if priorities[i].Name == newPriorityName {
						newSyncPriority = &priorities[i]
					}
				}
				Expect(newSyncPriority).NotTo(BeNil())
				Expect(newSyncPriority.DeletedAt).To(BeNil())

				rows, err = runaidb.DB().Raw("UPDATE priorities SET deleted_at = $1 WHERE name = $2;",
					time.Now().Local(), newPriorityName).Rows()
				Expect(err).To(BeNil())
				Expect(rows.Err()).To(BeNil())

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				newPriority = nil
				for i := range clusterPriorities {
					if clusterPriorities[i].Name == newPriorityName {
						newPriority = &clusterPriorities[i]
					}
				}
				Expect(newPriority).To(BeNil())

				syncResponse, err = workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities = syncResponse.Body.([]api.SyncPriority)
				Expect(len(priorities)).To(Equal(len(predefinedPriorities) + 1))
				newSyncPriority = nil
				for i := range priorities {
					if priorities[i].Name == newPriorityName {
						newSyncPriority = &priorities[i]
					}
				}
				Expect(newSyncPriority).NotTo(BeNil())
				Expect(newSyncPriority.DeletedAt).NotTo(BeNil())
			})

			It("sync multi cluster priority", func() {
				var lowPriorityId string
				var mediumPriorityId string
				syncResponse, err := workloadPropertiesService.GetSyncWorkloadPriorities(ctx, clusterId.String())
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusOK))
				priorities := syncResponse.Body.([]api.SyncPriority)
				for _, priority := range priorities {
					if priority.Name == "low" {
						lowPriorityId = priority.Id
					}
					if priority.Name == "medium" {
						mediumPriorityId = priority.Id
					}
				}

				statusResponse, err := workloadPropertiesService.WorkloadPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Id:     lowPriorityId,
						Name:   "low",
						Value:  40,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}, {
						Id:     mediumPriorityId,
						Name:   "medium",
						Value:  80,
						Source: RunaiSource,
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Not Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  800,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))
				// update the status for cluster 2
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId2.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  11,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				// cluster defined priorities should be returned for user
				response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities := response.Body.(api.GetWorkloadPriorities200Response).Priorities
				lowFound := false
				mediumFound := false
				cluster1DefinedFound := false
				cluster2DefinedFound := false
				for _, priority := range clusterPriorities {
					if priority.Name == "low" {
						lowFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId.String() {
								Expect(status[i].Phase).To(Equal("Ready"))
							}
							if status[i].ClusterId == clusterId2.String() {
								Fail("status should not be present for cluster2")
							}
						}
					}
					if priority.Name == "medium" {
						mediumFound = true
						status := priority.Status
						for i := range status {
							if status[i].ClusterId == clusterId2.String() {
								Expect(status[i].Phase).To(Equal("Not Ready"))
							}
							if status[i].ClusterId == clusterId2.String() {
								Fail("status should not be present for cluster2")
							}
						}
					}
					if priority.Name == "legacy" {
						if priority.ScopeId == clusterId.String() {
							cluster1DefinedFound = true
							Expect(priority.Value).To(Equal(int32(800)))
							Expect(priority.Source).To(Equal(""))
							status := priority.Status
							Expect(len(status)).To(Equal(1))
							Expect(status[0].Phase).To(Equal("Ready"))
						}
						if priority.ScopeId == clusterId2.String() {
							cluster2DefinedFound = true
							Expect(priority.Value).To(Equal(int32(11)))
							Expect(priority.Source).To(Equal(""))
							status := priority.Status
							Expect(len(status)).To(Equal(1))
							Expect(status[0].Phase).To(Equal("Ready"))
						}
					}
				}
				Expect(lowFound).To(BeTrue())
				Expect(mediumFound).To(BeTrue())
				Expect(cluster1DefinedFound).To(BeTrue())
				Expect(cluster2DefinedFound).To(BeTrue())

				// update the status
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId2.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  111,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				cluster1DefinedFound = false
				cluster2DefinedFound = false
				for _, priority := range clusterPriorities {
					if priority.Name == "legacy" {
						if priority.ScopeId == clusterId.String() {
							cluster1DefinedFound = true
							Expect(priority.Value).To(Equal(int32(800)))
							Expect(priority.Source).To(Equal(""))
							status := priority.Status
							Expect(len(status)).To(Equal(1))
							Expect(status[0].Phase).To(Equal("Ready"))
						}
						if priority.ScopeId == clusterId2.String() {
							cluster2DefinedFound = true
							Expect(priority.Value).To(Equal(int32(111)))
							Expect(priority.Source).To(Equal(""))
							status := priority.Status
							Expect(len(status)).To(Equal(1))
							Expect(status[0].Phase).To(Equal("Ready"))
						}
					}
				}
				Expect(cluster1DefinedFound).To(BeTrue())
				Expect(cluster2DefinedFound).To(BeTrue())

				// delete cluster 1 priority
				statusResponse, err = workloadPropertiesService.WorkloadClusterPrioritiesStatus(ctx, clusterId.String(), api.PrioritiesForStatus{
					FullSync: false,
					Priorities: []api.PrioritiesForStatusPrioritiesInner{{
						Name:   "legacy",
						Value:  800,
						Source: "",
						Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Deleted"},
					}},
				})
				Expect(err).To(BeNil())
				Expect(statusResponse.Code).To(Equal(http.StatusNoContent))

				response, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				clusterPriorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities
				cluster1DefinedFound = false
				cluster2DefinedFound = false
				for _, priority := range clusterPriorities {
					if priority.Name == "legacy" {
						if priority.ScopeId == clusterId.String() {
							cluster1DefinedFound = true
						}
						if priority.ScopeId == clusterId2.String() {
							cluster2DefinedFound = true
							Expect(priority.Value).To(Equal(int32(111)))
							Expect(priority.Source).To(Equal(""))
							status := priority.Status
							Expect(len(status)).To(Equal(1))
							Expect(status[0].Phase).To(Equal("Ready"))
						}
					}
				}
				Expect(cluster1DefinedFound).To(BeFalse())
				Expect(cluster2DefinedFound).To(BeTrue())

			})

		})

		Context("Deletion cleanup", func() {
			It("Should verify priorities and priority statuses are deleted during tenant deletion", func() {
				// Setup test data
				clusterId := uuid.New()
				tenantId := fmt.Sprintf("%s", testing.TestTenant)

				// Create test priorities with different scopes
				priority1 := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-tenant-priority",
					Value:     100,
					Source:    "Test",
					ScopeType: dao.TenantScopeType,
					ScopeId:   "",
					TenantId:  tenantId,
				}
				priority2 := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-cluster-priority",
					Value:     200,
					Source:    "Test",
					ScopeType: dao.ClusterScopeType,
					ScopeId:   clusterId.String(),
					TenantId:  tenantId,
				}

				// Insert test priorities
				_, err := runaidb.DB().Create(&priority1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&priority2)
				Expect(err).To(BeNil())

				// Create priority statuses for the test priorities
				status1 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: priority1.ID,
					ClusterId:  clusterId,
					Phase:      "Ready",
				}
				status2 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: priority2.ID,
					ClusterId:  clusterId,
					Phase:      "Creating",
				}

				// Insert priority statuses
				_, err = runaidb.DB().Create(&status1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&status2)
				Expect(err).To(BeNil())

				// Create test workload type and workload type statuses
				workloadType := models.WorkloadType{
					ID:     uuid.New(),
					Name:   "test-workload-type",
					Group:  "test.group",
					Source: models.WorkloadTypeSourceCustom,
				}
				_, err = runaidb.DB().Create(&workloadType)
				Expect(err).To(BeNil())

				workloadTypeStatus1 := models.WorkloadTypeStatus{
					ID:             uuid.New(),
					WorkloadTypeId: workloadType.ID,
					ClusterId:      clusterId,
					TenantId:       tenantId,
					Phase:          "Ready",
					Version:        "v1",
				}
				workloadTypeStatus2 := models.WorkloadTypeStatus{
					ID:             uuid.New(),
					WorkloadTypeId: workloadType.ID,
					ClusterId:      clusterId,
					TenantId:       tenantId,
					Phase:          "Pending",
					Version:        "v2",
				}

				// Insert workload type statuses
				_, err = runaidb.DB().Create(&workloadTypeStatus1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&workloadTypeStatus2)
				Expect(err).To(BeNil())

				// Create test workload events
				tenantIdUint, _ := strconv.ParseUint(tenantId, 10, 32)
				event1 := models.Event{
					TenantID:                uint(tenantIdUint),
					ID:                      uuid.New(),
					ClusterID:               clusterId,
					Type:                    "Created",
					Message:                 "Test workload event 1",
					Reason:                  "Testing",
					CreatedAt:               time.Now(),
					Source:                  "Test",
					InvolvedObjectUid:       uuid.New().String(),
					InvolvedObjectKind:      "Pod",
					InvolvedObjectName:      "test-pod-1",
					InvolvedObjectNamespace: "default",
				}
				event2 := models.Event{
					TenantID:                uint(tenantIdUint),
					ID:                      uuid.New(),
					ClusterID:               clusterId,
					Type:                    "Started",
					Message:                 "Test workload event 2",
					Reason:                  "Testing",
					CreatedAt:               time.Now(),
					Source:                  "Test",
					InvolvedObjectUid:       uuid.New().String(),
					InvolvedObjectKind:      "Pod",
					InvolvedObjectName:      "test-pod-2",
					InvolvedObjectNamespace: "default",
				}

				// Insert workload events
				_, err = runaidb.DB().Create(&event1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&event2)
				Expect(err).To(BeNil())

				// Verify data exists before deletion
				var count int64
				err = runaidb.DB().Model(&models.Priority{}).Where("tenant_id = ?", tenantId).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(BeNumerically(">=", 2), "Should have at least 2 test priorities before deletion")

				// Verify priority statuses exist
				_, err = runaidb.DB().Raw(`SELECT COUNT(*) FROM priority_statuses ps 
					INNER JOIN priorities p ON ps.priority_id = p.id 
					WHERE p.tenant_id = ?`, tenantId).Scan(&count)
				Expect(err).To(BeNil())
				Expect(count).To(BeNumerically(">=", 2), "Should have at least 2 priority statuses before deletion")

				// Verify workload events exist
				err = runaidb.DB().Model(&models.Event{}).Where("tenant_id = ?", tenantIdUint).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(BeNumerically(">=", 2), "Should have at least 2 workload events before deletion")

				// Verify workload type statuses exist
				err = runaidb.DB().Model(&models.WorkloadTypeStatus{}).Where("tenant_id = ?", tenantId).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(BeNumerically(">=", 2), "Should have at least 2 workload type statuses before deletion")

				// Execute tenant deletion using the common package
				deleteRequest := events.TenantDelete{
					TenantId: tenantId,
				}
				err = common.HandleTenantDeleteRequest(ctx, &deleteRequest)
				Expect(err).To(BeNil(), "Tenant deletion should succeed")

				// Verify priorities are deleted (using Unscoped to check hard deletion)
				err = runaidb.DB().Unscoped().Model(&models.Priority{}).Where("tenant_id = ?", tenantId).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "All test priorities should be deleted")

				// Verify priority statuses are also deleted
				_, err = runaidb.DB().Raw(`SELECT COUNT(*) FROM priority_statuses ps 
					WHERE ps.priority_id IN (?, ?)`, priority1.ID, priority2.ID).Scan(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "All test priority statuses should be deleted")

				// Verify workload events are deleted
				err = runaidb.DB().Model(&models.Event{}).Where("id IN (?, ?)", event1.ID, event2.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Test workload events should be deleted")

				// Verify workload type statuses are deleted (using Unscoped to check hard deletion)
				err = runaidb.DB().Unscoped().Model(&models.WorkloadTypeStatus{}).
					Where("id IN (?, ?)", workloadTypeStatus1.ID, workloadTypeStatus2.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Test workload type statuses should be deleted")
			})

			It("Should verify priorities deletion handles foreign key constraints", func() {
				// This test specifically validates the fix for the foreign key constraint issue
				tenantId := fmt.Sprintf("%s", testing.TestTenant)
				clusterId := uuid.New()

				// Create a priority
				priority := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "fk-test-priority",
					Value:     150,
					Source:    "Test",
					ScopeType: dao.TenantScopeType,
					ScopeId:   "",
					TenantId:  tenantId,
				}

				_, err := runaidb.DB().Create(&priority)
				Expect(err).To(BeNil())

				// Create multiple priority statuses for the same priority
				status1 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: priority.ID,
					ClusterId:  clusterId,
					Phase:      "Ready",
				}
				status2 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: priority.ID,
					ClusterId:  uuid.New(), // Different cluster
					Phase:      "Failed",
				}

				_, err = runaidb.DB().Create(&status1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&status2)
				Expect(err).To(BeNil())

				// Call the deletion function directly to test the fix
				err = dao.DeletePrioritiesOfTenantId(ctx, tenantId)

				// This should not fail due to foreign key constraints anymore
				Expect(err).To(BeNil(), "DeletePrioritiesOfTenantId should handle foreign key constraints properly")

				// Verify priority is deleted
				var count int64
				err = runaidb.DB().Unscoped().Model(&models.Priority{}).Where("id = ?", priority.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Priority should be completely deleted")

				// Verify all related priority statuses are deleted
				err = runaidb.DB().Model(&models.PriorityStatus{}).Where("priority_id = ?", priority.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "All priority statuses should be deleted")
			})

			It("Should delete tenant priorities when external_workloads_type_config references them", func() {
				// This test validates that tenant deletion works when workload type configs reference priorities
				// This is the exact scenario that was causing the foreign key constraint violation
				tenantId := fmt.Sprintf("%s", testing.TestTenant)

				// Create a tenant-scoped priority
				priority := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      fmt.Sprintf("config-ref-priority-%s", rand.String(5)),
					Value:     175,
					Source:    "Test",
					ScopeType: dao.TenantScopeType,
					ScopeId:   "",
					TenantId:  tenantId,
				}

				_, err := runaidb.DB().Create(&priority)
				Expect(err).To(BeNil())

				// Create a workload type for testing
				workloadType, apiErr := dao.DbUpsertWorkloadType(ctx, runaidb.DB(), dao.WorkloadTypeCreateFieldsParsed{
					Group: "test.example.com",
					WorkloadTypeCreateUpdateFieldsParsed: dao.WorkloadTypeCreateUpdateFieldsParsed{
						WorkloadTypeName: fmt.Sprintf("TestWorkloadType-%s", rand.String(5)),
					},
				})
				Expect(apiErr).To(BeNil())

				// Get a category for the config
				category, catErr := dao.DbGetCategoryByName(ctx, 0, "Deploy")
				Expect(catErr).To(BeNil())
				categoryUUID := uuid.MustParse(category.Id)

				// Create an external_workloads_type_config that references the tenant priority
				config := models.ExternalWorkloadTypeConfig{
					ID:             uuid.New(),
					TenantId:       &tenantId,
					ScopeType:      dao.TenantScopeType,
					WorkloadTypeId: workloadType.ID,
					CategoryID:     &categoryUUID,
					PriorityID:     &priority.ID, // This references the tenant priority!
					CreatedAt:      time.Now(),
					CreatedBy:      ptr.To("Test"),
					UpdatedAt:      time.Now(),
					UpdatedBy:      ptr.To("Test"),
				}

				_, err = runaidb.DB().Create(&config)
				Expect(err).To(BeNil())

				// Verify the config was created and references the priority
				var count int64
				err = runaidb.DB().Model(&models.ExternalWorkloadTypeConfig{}).
					Where("id = ? AND priority_id = ?", config.ID, priority.ID).
					Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Config should exist and reference the priority")

				// Debug: Check what tenantId value is in the config
				var dbConfig models.ExternalWorkloadTypeConfig
				err = runaidb.DB().Where("id = ?", config.ID).First(&dbConfig)
				Expect(err).To(BeNil())
				if dbConfig.TenantId != nil {
					fmt.Printf("DEBUG: Config TenantId in DB: '%s', Deleting for TenantId: '%s'\n", *dbConfig.TenantId, tenantId)
				} else {
					fmt.Printf("DEBUG: Config TenantId in DB is NULL, Deleting for TenantId: '%s'\n", tenantId)
				}

				// Debug: Count configs before deletion
				err = runaidb.DB().Model(&models.ExternalWorkloadTypeConfig{}).
					Where("tenant_id = ?", tenantId).Count(&count)
				fmt.Printf("DEBUG: Configs with tenant_id='%s' BEFORE deletion: %d\n", tenantId, count)

				// Use the full tenant deletion flow - this is what happens in production
				deleteRequest := events.TenantDelete{
					TenantId: tenantId,
				}
				err = common.HandleTenantDeleteRequest(ctx, &deleteRequest)

				// Debug: Count configs after deletion attempt
				err = runaidb.DB().Unscoped().Model(&models.ExternalWorkloadTypeConfig{}).
					Where("id = ?", config.ID).Count(&count)
				fmt.Printf("DEBUG: Config still exists after HandleTenantDeleteRequest (Unscoped): %d\n", count)

				// Debug: Check if priority still exists
				err = runaidb.DB().Unscoped().Model(&models.Priority{}).
					Where("id = ?", priority.ID).Count(&count)
				fmt.Printf("DEBUG: Priority still exists after HandleTenantDeleteRequest (Unscoped): %d\n", count)

				// This should succeed without foreign key constraint violation
				Expect(err).To(BeNil(), "HandleTenantDeleteRequest should succeed even with config referencing priority")

				// Verify config is HARD deleted (not just soft deleted)
				err = runaidb.DB().Unscoped().Model(&models.ExternalWorkloadTypeConfig{}).
					Where("id = ?", config.ID).
					Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Config should be hard deleted by HandleTenantDeleteRequest")

				// Verify priority is deleted
				err = runaidb.DB().Unscoped().Model(&models.Priority{}).Where("id = ?", priority.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Priority should be completely deleted by HandleTenantDeleteRequest")
			})

			It("Should verify priorities and priority statuses are deleted during cluster deletion", func() {
				// Setup test data
				clusterId1 := uuid.New()
				clusterId2 := uuid.New()
				tenantId := fmt.Sprintf("%s", testing.TestTenant)

				// Create cluster-scoped priorities for cluster1 (should be deleted)
				clusterPriority1 := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-cluster1-priority-1",
					Value:     100,
					Source:    "Test",
					ScopeType: dao.ClusterScopeType,
					ScopeId:   clusterId1.String(),
					TenantId:  tenantId,
				}
				clusterPriority2 := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-cluster1-priority-2",
					Value:     200,
					Source:    "Test",
					ScopeType: dao.ClusterScopeType,
					ScopeId:   clusterId1.String(),
					TenantId:  tenantId,
				}

				// Create tenant-scoped priority (should NOT be deleted)
				tenantPriority := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-tenant-priority",
					Value:     300,
					Source:    "Test",
					ScopeType: dao.TenantScopeType,
					ScopeId:   "",
					TenantId:  tenantId,
				}

				// Create cluster-scoped priority for cluster2 (should NOT be deleted)
				clusterPriority3 := models.Priority{
					BaseModel: dbmodels.BaseModel{ID: uuid.New()},
					Name:      "test-cluster2-priority",
					Value:     400,
					Source:    "Test",
					ScopeType: dao.ClusterScopeType,
					ScopeId:   clusterId2.String(),
					TenantId:  tenantId,
				}

				// Insert test priorities
				_, err := runaidb.DB().Create(&clusterPriority1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&clusterPriority2)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&tenantPriority)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&clusterPriority3)
				Expect(err).To(BeNil())

				// Create priority statuses for cluster1 priorities
				status1 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: clusterPriority1.ID,
					ClusterId:  clusterId1,
					Phase:      "Ready",
				}
				status2 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: clusterPriority2.ID,
					ClusterId:  clusterId1,
					Phase:      "Creating",
				}
				// Create priority status for tenant priority on cluster1
				status3 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: tenantPriority.ID,
					ClusterId:  clusterId1,
					Phase:      "Ready",
				}
				// Create priority status for cluster2 priority (should not be deleted)
				status4 := models.PriorityStatus{
					ID:         uuid.New(),
					PriorityID: clusterPriority3.ID,
					ClusterId:  clusterId2,
					Phase:      "Ready",
				}

				// Insert priority statuses
				_, err = runaidb.DB().Create(&status1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&status2)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&status3)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&status4)
				Expect(err).To(BeNil())

				// Create test workload events for cluster1 (should be deleted)
				tenantIdUint, _ := strconv.ParseUint(tenantId, 10, 32)
				event1 := models.Event{
					TenantID:                uint(tenantIdUint),
					ID:                      uuid.New(),
					ClusterID:               clusterId1,
					Type:                    "Created",
					Message:                 "Test cluster1 event 1",
					Reason:                  "Testing",
					CreatedAt:               time.Now(),
					Source:                  "Test",
					InvolvedObjectUid:       uuid.New().String(),
					InvolvedObjectKind:      "Pod",
					InvolvedObjectName:      "test-pod-c1-1",
					InvolvedObjectNamespace: "default",
				}

				// Create test workload event for cluster2 (should NOT be deleted)
				event2 := models.Event{
					TenantID:                uint(tenantIdUint),
					ID:                      uuid.New(),
					ClusterID:               clusterId2,
					Type:                    "Started",
					Message:                 "Test cluster2 event",
					Reason:                  "Testing",
					CreatedAt:               time.Now(),
					Source:                  "Test",
					InvolvedObjectUid:       uuid.New().String(),
					InvolvedObjectKind:      "Pod",
					InvolvedObjectName:      "test-pod-c2-1",
					InvolvedObjectNamespace: "default",
				}

				// Insert workload events
				_, err = runaidb.DB().Create(&event1)
				Expect(err).To(BeNil())
				_, err = runaidb.DB().Create(&event2)
				Expect(err).To(BeNil())

				// Verify data exists before deletion
				var count int64

				// Verify cluster1 cluster-scoped priorities exist
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND scope_id = ?", dao.ClusterScopeType, clusterId1.String()).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(2)), "Should have 2 cluster1 priorities before deletion")

				// Verify tenant-scoped priority exists
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND tenant_id = ?", dao.TenantScopeType, tenantId).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Should have 1 tenant priority before deletion")

				// Verify cluster2 priority exists
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND scope_id = ?", dao.ClusterScopeType, clusterId2.String()).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Should have 1 cluster2 priority before deletion")

				// Verify priority statuses on cluster1 exist (should include tenant priority status on cluster1)
				err = runaidb.DB().Model(&models.PriorityStatus{}).Where("cluster_id = ?", clusterId1).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(3)), "Should have 3 priority statuses on cluster1 before deletion")

				// Verify workload events exist
				err = runaidb.DB().Model(&models.Event{}).Where("cluster_id = ?", clusterId1).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(BeNumerically(">=", 1), "Should have at least 1 workload event on cluster1 before deletion")

				// Execute cluster deletion using the common package
				deleteRequest := events.ClusterDelete{
					ClusterId: clusterId1.String(),
				}
				err = common.HandleClusterDeleteRequest(ctx, &deleteRequest)
				Expect(err).To(BeNil(), "Cluster deletion should succeed")

				// Verify cluster1 cluster-scoped priorities are soft deleted (deleted_at is NOT null)
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND scope_id = ? AND deleted_at IS NULL", dao.ClusterScopeType, clusterId1.String()).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Cluster1 cluster-scoped priorities should be soft deleted (deleted_at set)")

				// Verify they exist in the database but with deleted_at set
				err = runaidb.DB().Unscoped().Model(&models.Priority{}).Where("scope_type = ? AND scope_id = ? AND deleted_at IS NOT NULL", dao.ClusterScopeType, clusterId1.String()).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(2)), "Cluster1 cluster-scoped priorities should exist with deleted_at set")

				// Verify tenant-scoped priority still exists (should not be affected by cluster deletion)
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND tenant_id = ?", dao.TenantScopeType, tenantId).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Tenant-scoped priority should still exist after cluster deletion")

				// Verify cluster2 priority still exists
				err = runaidb.DB().Model(&models.Priority{}).Where("scope_type = ? AND scope_id = ?", dao.ClusterScopeType, clusterId2.String()).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Cluster2 priority should still exist after cluster1 deletion")

				// Verify cluster1 workload events are deleted
				err = runaidb.DB().Model(&models.Event{}).Where("id = ?", event1.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(0)), "Cluster1 workload events should be deleted")

				// Verify cluster2 workload events still exist
				err = runaidb.DB().Model(&models.Event{}).Where("id = ?", event2.ID).Count(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(1)), "Cluster2 workload events should still exist")
			})
		})
	})

	Context("Workload Types", func() {
		BeforeAll(func() {
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		})

		It("ListWorkloadTypes - returns all workload types", func() {
			apiResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{}, false)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			response, ok := apiResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			Expect(response.Types).ToNot(BeNil())
			Expect(len(response.Types)).To(BeNumerically(">=", 10))

			apiResponse, err = workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			categoriesResponse, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())

			apiResponse, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			prioritiesResponse, ok := apiResponse.Body.(api.GetWorkloadPriorities200Response)
			Expect(ok).To(BeTrue())

			deployCategory := findCategoryByName(categoriesResponse.Categories, "Deploy")
			trainCategory := findCategoryByName(categoriesResponse.Categories, "Train")
			buildCategory := findCategoryByName(categoriesResponse.Categories, "Build")
			veryHighPriority := findPriorityByName(prioritiesResponse.Priorities, "very-high")
			lowPriority := findPriorityByName(prioritiesResponse.Priorities, "low")
			highPriority := findPriorityByName(prioritiesResponse.Priorities, "high")

			// test a few specific workload types
			// first, verify the types are present
			workloadTypeNames := make(map[string]struct{})
			for _, wt := range response.Types {
				workloadTypeNames[wt.Name] = struct{}{}

				if !wt.Preemptibility.IsValid() {
					Fail(fmt.Sprintf("Workload type %s has invalid preemptibility value: %s", wt.Name, wt.Preemptibility))
				}
			}
			for _, typeName := range []string{
				"Deployment",
				"RayCluster",
				"Notebook",
				"TFJob",
				"PyTorchJob",
				"JAXJob",
				"MPIJob",
				"SeldonDeployment",
				"NIMCache",
				"NIMService",
				"SPOTRequest",
				"Pod",
				"InferenceService",
			} {
				_, found := workloadTypeNames[typeName]
				Expect(found).To(BeTrue(), fmt.Sprintf("%s should be present in workload types", typeName))
			}

			// then, verify their details
			for _, workloadType := range response.Types {
				switch workloadType.Name {
				case "Deployment":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "Deployment",
						CreatedBy:      RunaiSource,
						Group:          "apps",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: buildK8sDeploymentResourceInterface(),
							},
						},
					})
				case "RayCluster":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "RayCluster",
						CreatedBy:      RunaiSource,
						Group:          "ray.io",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildRayClusterV1()),
							},
							{
								Version:           "v1alpha1",
								ResourceInterface: convertKartaToMap(source.BuildRayClusterV1alpha1()),
							},
						},
					})
				case "Notebook":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     buildCategory.Id,
						CategoryName:   buildCategory.Name,
						PriorityId:     highPriority.Id,
						PriorityName:   highPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "Notebook",
						CreatedBy:      RunaiSource,
						Group:          "kubeflow.org",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1beta1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowNotebookV1beta1()),
							},
						},
					})
				case "DistributedWorkload":
					Expect(workloadType.Id).ToNot(Equal(""))
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "DistributedWorkload",
						CreatedBy:      RunaiSource,
						Group:          "run.ai",
						Source:         string(models.WorkloadTypeSourceNative),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version: "v2alpha1",
							},
						},
					})
				case "TrainingWorkload":
					Expect(workloadType.Id).ToNot(Equal(""))
					Expect(workloadType.CreatedAt).To(BeTemporally("~", time.Now(), 5*time.Minute))
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "TrainingWorkload",
						CreatedBy:      RunaiSource,
						Group:          "run.ai",
						Source:         string(models.WorkloadTypeSourceNative),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version: "v2alpha1",
							},
						},
					})
				case "Job":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "Job",
						CreatedBy:      RunaiSource,
						Group:          "batch",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: buildK8sJobResourceInterface(),
							},
						},
					})
				case "JAXJob":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "JAXJob",
						CreatedBy:      RunaiSource,
						Group:          "kubeflow.org",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowJAXJobV1()),
							},
						},
					})
				case "NIMCache":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "NIMCache",
						CreatedBy:      RunaiSource,
						Group:          "apps.nvidia.com",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1alpha1",
								ResourceInterface: convertKartaToMap(nvidia.BuildNIMCacheV1alpha1()),
							},
						},
					})
				case "NIMService":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "NIMService",
						CreatedBy:      RunaiSource,
						Group:          "apps.nvidia.com",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1alpha1",
								ResourceInterface: convertKartaToMap(nvidia.BuildNIMServiceV1alpha1()),
							},
						},
					})
				case "SPOTRequest":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "SPOTRequest",
						CreatedBy:      RunaiSource,
						Group:          "egx.nvidia.io",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(nvidia.BuildNVCFSPOTRequestV1()),
							},
						},
					})
				case "MPIJob":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "MPIJob",
						CreatedBy:      RunaiSource,
						Group:          "kubeflow.org",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowMPIJobV1()),
							},
							{
								Version:           "v2beta1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowMPIJobV2beta1()),
							},
						},
					})
				case "SeldonDeployment":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "SeldonDeployment",
						CreatedBy:      RunaiSource,
						Group:          "machinelearning.seldon.io",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildSeldonDeploymentV1()),
							},
							{
								Version:           "v1alpha2",
								ResourceInterface: convertKartaToMap(source.BuildSeldonDeploymentV1alpha2()),
							},
						},
					})
				case "Pod":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "Pod",
						CreatedBy:      RunaiSource,
						Group:          "",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: buildK8sPodResourceInterface(),
							},
						},
					})
				case "TFJob":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "TFJob",
						CreatedBy:      RunaiSource,
						Group:          "kubeflow.org",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowTFJobV1()),
							},
						},
					})
				case "PyTorchJob":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     trainCategory.Id,
						CategoryName:   trainCategory.Name,
						PriorityId:     lowPriority.Id,
						PriorityName:   lowPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
						Name:           "PyTorchJob",
						CreatedBy:      RunaiSource,
						Group:          "kubeflow.org",
						Source:         string(models.WorkloadTypeSourceExternal),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1",
								ResourceInterface: convertKartaToMap(source.BuildKubeflowPyTorchJobV1()),
							},
						},
					})
				case "InferenceService":
					validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
						CategoryId:     deployCategory.Id,
						CategoryName:   deployCategory.Name,
						PriorityId:     veryHighPriority.Id,
						PriorityName:   veryHighPriority.Name,
						Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
						Name:           "InferenceService",
						CreatedBy:      RunaiSource,
						Group:          "serving.kserve.io",
						Source:         string(models.WorkloadTypeSourceCustom),
						ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
							{
								Version:           "v1beta1",
								ResourceInterface: convertKartaToMap(source.BuildKServeInferenceServiceV1beta1()),
							},
						},
					})
				}
			}
		})

		It("ListWorkloadTypes - using externalTypesOnly param", func() {
			apiResponse, err := workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			categoriesResponse, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			apiResponse, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			prioritiesResponse, ok := apiResponse.Body.(api.GetWorkloadPriorities200Response)
			Expect(ok).To(BeTrue())

			deployCategory := findCategoryByName(categoriesResponse.Categories, "Deploy")
			trainCategory := findCategoryByName(categoriesResponse.Categories, "Train")
			buildCategory := findCategoryByName(categoriesResponse.Categories, "Build")
			veryHighPriority := findPriorityByName(prioritiesResponse.Priorities, "very-high")
			lowPriority := findPriorityByName(prioritiesResponse.Priorities, "low")
			highPriority := findPriorityByName(prioritiesResponse.Priorities, "high")

			apiResponse, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{}, false)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response, ok := apiResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			Expect(response.Types).ToNot(BeNil())
			Expect(len(response.Types)).To(BeNumerically(">=", 10))

			training, dist, interactive, inference := findRunaiWorkloadTypes(response.Types)
			Expect(training).ToNot(BeNil())
			Expect(dist).ToNot(BeNil())
			Expect(interactive).ToNot(BeNil())
			Expect(inference).ToNot(BeNil())

			validateWorkloadTypeConfig(*training, api.WorkloadTypeConfig{
				CategoryId:     trainCategory.Id,
				CategoryName:   trainCategory.Name,
				PriorityId:     lowPriority.Id,
				PriorityName:   lowPriority.Name,
				Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
				Name:           "TrainingWorkload",
				CreatedBy:      RunaiSource,
				Group:          "run.ai",
				Source:         string(models.WorkloadTypeSourceNative),
				ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version: "v2alpha1",
					},
				},
			})
			validateWorkloadTypeConfig(*dist, api.WorkloadTypeConfig{
				CategoryId:     trainCategory.Id,
				CategoryName:   trainCategory.Name,
				PriorityId:     lowPriority.Id,
				PriorityName:   lowPriority.Name,
				Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
				Name:           "DistributedWorkload",
				CreatedBy:      RunaiSource,
				Group:          "run.ai",
				Source:         string(models.WorkloadTypeSourceNative),
				ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version: "v2alpha1",
					},
				},
			})
			validateWorkloadTypeConfig(*interactive, api.WorkloadTypeConfig{
				CategoryId:     buildCategory.Id,
				CategoryName:   buildCategory.Name,
				PriorityId:     highPriority.Id,
				PriorityName:   highPriority.Name,
				Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
				Name:           "InteractiveWorkload",
				CreatedBy:      RunaiSource,
				Group:          "run.ai",
				Source:         string(models.WorkloadTypeSourceNative),
				ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version: "v2alpha1",
					},
				},
			})
			validateWorkloadTypeConfig(*inference, api.WorkloadTypeConfig{
				CategoryId:     deployCategory.Id,
				CategoryName:   deployCategory.Name,
				PriorityId:     veryHighPriority.Id,
				PriorityName:   veryHighPriority.Name,
				Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
				Name:           "InferenceWorkload",
				CreatedBy:      RunaiSource,
				Group:          "run.ai",
				Source:         string(models.WorkloadTypeSourceNative),
				ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version: "v2alpha1",
					},
				},
			})

			apiResponse, err = workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response, ok = apiResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			Expect(response.Types).ToNot(BeNil())

			training, dist, interactive, inference = findRunaiWorkloadTypes(response.Types)
			Expect(training).To(BeNil())
			Expect(dist).To(BeNil())
			Expect(interactive).To(BeNil())
			Expect(inference).To(BeNil())
		})

		Context("ListWorkloadTypes with filters", func() {
			It("filters by group and name with externalTypesOnly=false", func() {
				// filter by group run.ai -> should return only native run.ai workload types
				resp, err := workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"group==run.ai"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok := resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(BeNumerically(">=", 4))
				for _, t := range list.Types {
					Expect(t.Group).To(Equal("run.ai"))
					Expect(t.Name).ToNot(Equal("Job"))
				}

				// filter by name and group run.ai -> should return only the one native run.ai workload type
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"group==run.ai", "name==DistributedWorkload"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(1))
				Expect(list.Types[0].Name).To(Equal("DistributedWorkload"))
				Expect(list.Types[0].Group).To(Equal("run.ai"))

				// filter by name Job -> should return the Job type from batch group
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"name==Job"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(1))
				Expect(list.Types[0].Name).To(Equal("Job"))
				Expect(list.Types[0].Group).To(Equal("batch"))

				// filter by name and group -> should return the correct workload type
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"name==AmlJob", "group==amlarc.azureml.com"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(1))
				Expect(list.Types[0].Name).To(Equal("AmlJob"))
				Expect(list.Types[0].Group).To(Equal("amlarc.azureml.com"))

				// filter by group -> should return several workload types of same group
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"group==ray.io"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(3))
				for _, t := range list.Types {
					Expect(t.Group).To(Equal("ray.io"))
				}
			})

			It("filters by group and name with externalTypesOnly=true", func() {
				// filter by group run.ai -> should return none for external-only listing
				resp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{"group==run.ai"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok := resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(0))

				// filter by name Job -> should still return the external Job type
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{"name==Job"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(1))
				Expect(list.Types[0].Name).To(Equal("Job"))
				Expect(list.Types[0].Group).To(Equal("batch"))

				// filter by name InteractiveWorkload -> native, should not appear in external-only listing
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{"name==InteractiveWorkload"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(0))
			})

			It("filter by group and name - more tests", func() {
				// filter with group !=
				resp, err := workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"group!=kubeflow.org"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok := resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(BeNumerically(">=", 10))
				for _, t := range list.Types {
					Expect(t.Group).ToNot(Equal("kubeflow.org"))
				}

				// sanity - filter by name
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"name==Deployment"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(1))
				Expect(list.Types[0].Name).To(Equal("Deployment"))

				// filter by name with wrong case - should return nothing
				resp, err = workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{"name==deployment"}, false)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusOK))
				list, ok = resp.Body.(api.ListWorkloadTypes200Response)
				Expect(ok).To(BeTrue())
				Expect(len(list.Types)).To(Equal(0))
			})
		})

		It("GetWorkloadType - returns a specific workload type by ID", func() {
			// First, list all types to get a valid ID
			apiResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, false, []string{}, false)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			response, ok := apiResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			Expect(len(response.Types)).To(BeNumerically(">=", 10))

			typeId := ""
			nativeTypeId := ""
			for _, workloadType := range response.Types {
				if workloadType.Name == "TFJob" {
					typeId = workloadType.Id
					continue
				}
				if workloadType.Name == "InteractiveWorkload" {
					nativeTypeId = workloadType.Id
				}
			}
			Expect(typeId).ToNot(BeEmpty())
			Expect(nativeTypeId).ToNot(BeEmpty())

			apiResponse, err = workloadPropertiesService.GetWorkloadType(ctx, typeId)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			workloadType, ok := apiResponse.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(workloadType.Id).To(Equal(typeId))
			Expect(workloadType.Name).To(Equal("TFJob"))
			Expect(workloadType.CreatedBy).To(Equal(RunaiSource))
			Expect(workloadType.CreatedAt).ToNot(BeNil())
			Expect(workloadType.UpdatedBy).To(Equal(RunaiSource))
			Expect(workloadType.UpdatedAt).ToNot(BeNil())

			apiResponse, err = workloadPropertiesService.GetWorkloadType(ctx, nativeTypeId)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			nativeWorkloadType, ok := apiResponse.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(nativeWorkloadType.Id).To(Equal(nativeTypeId))
			Expect(nativeWorkloadType.Name).To(Equal("InteractiveWorkload"))
			Expect(workloadType.CreatedBy).To(Equal(RunaiSource))
			Expect(workloadType.CreatedAt).ToNot(BeNil())
			Expect(workloadType.UpdatedBy).To(Equal(RunaiSource))
			Expect(workloadType.UpdatedAt).ToNot(BeNil())

			apiResponse, err = workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			categoriesResponse, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			apiResponse, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			prioritiesResponse, ok := apiResponse.Body.(api.GetWorkloadPriorities200Response)
			Expect(ok).To(BeTrue())

			trainCategory := findCategoryByName(categoriesResponse.Categories, "Train")
			lowPriority := findPriorityByName(prioritiesResponse.Priorities, "low")

			validateWorkloadTypeConfig(workloadType, api.WorkloadTypeConfig{
				CategoryId:     trainCategory.Id,
				CategoryName:   trainCategory.Name,
				PriorityId:     lowPriority.Id,
				PriorityName:   lowPriority.Name,
				Preemptibility: api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
				Name:           "TFJob",
				CreatedBy:      RunaiSource,
				Group:          "kubeflow.org",
				Source:         string(models.WorkloadTypeSourceExternal),
				ResourceInterfaces: []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version:           "v1",
						ResourceInterface: convertKartaToMap(source.BuildKubeflowTFJobV1()),
					},
				},
			})
		})

		It("GetWorkloadType - returns 404 for non-existent type", func() {
			ctx := testing.SetSystemContext(testing.TestUser)
			nonExistentId := uuid.NewString()
			apiResponse, err := workloadPropertiesService.GetWorkloadType(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("GetWorkloadType - returns 400 for invalid uuid", func() {
			ctx := testing.SetSystemContext(testing.TestUser)
			apiResponse, err := workloadPropertiesService.GetWorkloadType(ctx, "invalid-uuid")
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
			Expect(apiResponse.Body.(api.Error).Message).To(ContainSubstring("uuid"))
		})

		It("Test InferenceService ResourceInterface and wokrload type", func() {
			apiResponse, err := workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			categoriesResponse, ok := apiResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			apiResponse, err = workloadPropertiesService.GetWorkloadPriorities(ctx, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			prioritiesResponse, ok := apiResponse.Body.(api.GetWorkloadPriorities200Response)
			Expect(ok).To(BeTrue())
			deployCategory := findCategoryByName(categoriesResponse.Categories, "Deploy")
			veryHighPriority := findPriorityByName(prioritiesResponse.Priorities, "very-high")

			inferenceServiceWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, "InferenceService", true)
			Expect(inferenceServiceWorkloadType).ToNot(BeNil())
			Expect(inferenceServiceWorkloadType.CategoryId).To(Equal(deployCategory.Id))
			Expect(inferenceServiceWorkloadType.CategoryName).To(Equal(deployCategory.Name))
			Expect(inferenceServiceWorkloadType.PriorityId).To(Equal(veryHighPriority.Id))
			Expect(inferenceServiceWorkloadType.PriorityName).To(Equal(veryHighPriority.Name))
			Expect(inferenceServiceWorkloadType.Name).To(Equal("InferenceService"))
			Expect(inferenceServiceWorkloadType.Group).To(Equal("serving.kserve.io"))
			Expect(inferenceServiceWorkloadType.UpdatedAt).To(BeTemporally("~", time.Now(), 5*time.Minute))
			Expect(inferenceServiceWorkloadType.UpdatedBy).To(Equal(RunaiSource))
			Expect(inferenceServiceWorkloadType.CreatedAt).To(BeTemporally("~", time.Now(), 5*time.Minute))
			Expect(inferenceServiceWorkloadType.CreatedBy).To(Equal(RunaiSource))
			Expect(inferenceServiceWorkloadType.Source).To(Equal(string(models.WorkloadTypeSourceCustom)))
			Expect(inferenceServiceWorkloadType.ResourceInterfaces).To(HaveLen(1))

			ri := inferenceServiceWorkloadType.ResourceInterfaces[0]
			Expect(ri.UpdatedAt).To(BeTemporally("~", time.Now(), 5*time.Minute))
			Expect(ri.Version).To(Equal("v1beta1"))
			Expect(ri.ResourceInterface).ToNot(BeNil())
			validateExpectedKserveInferenceService(ri.ResourceInterface)
		})
	})

	Context("UpdateWorkloadType tests", func() {
		var (
			priorities []api.WorkloadPriority
			categories []api.WorkloadCategory
		)

		BeforeAll(func() {

			response, err := workloadPropertiesService.GetWorkloadPriorities(ctx, "", "name", []string{})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			priorities = response.Body.(api.GetWorkloadPriorities200Response).Priorities

			response, err = workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			categories = response.Body.(api.ListCategories200Response).Categories

		})

		It("Add priority only for external workload", func() {
			listWorkloadsResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listWorkloadsResponse.Code).To(Equal(http.StatusOK))

			listResponse, ok := listWorkloadsResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			workloadTypeObj := listResponse.Types[0]
			toUpdateId := workloadTypeObj.Id
			currentPriorityId := workloadTypeObj.PriorityId
			currentCategoryId := workloadTypeObj.CategoryId
			currentPreemptibility := workloadTypeObj.Preemptibility

			// Get - just to validate the current state
			response, err := workloadPropertiesService.GetWorkloadType(ctx, toUpdateId)
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(updatedWorkloadType.PriorityId).To(Equal(currentPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(currentPreemptibility))

			var newPriorityId string

			// find any other different priority
			for _, priority := range priorities {
				if priority.Id != currentPriorityId && priority.ScopeType == "system" &&
					computeExpectedPreemptibility(ctx, testing.TestTenant, priority.Id) != currentPreemptibility {
					newPriorityId = priority.Id
					break
				}
			}
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: newPriorityId,
			}

			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(updatedWorkloadType.PriorityId).To(Equal(newPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId))
			expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, newPriorityId)
			Expect(updatedWorkloadType.Preemptibility).To(Equal(expectedPreemptibility))

			response, err = workloadPropertiesService.GetWorkloadType(ctx, toUpdateId)
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(updatedWorkloadType.PriorityId).To(Equal(newPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(expectedPreemptibility))

			// switch to different tenant - should be able to get the system-scoped config
			ctx = testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
			response, err = workloadPropertiesService.GetWorkloadType(ctx, toUpdateId)
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(updatedWorkloadType.PriorityId).To(Equal(currentPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(currentPreemptibility))

			// and also should find in list
			testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, updatedWorkloadType.Name, true)
			Expect(testWorkloadType).ToNot(BeNil())
			Expect(testWorkloadType.PriorityId).To(Equal(currentPriorityId))
			Expect(testWorkloadType.CategoryId).To(Equal(currentCategoryId))
			Expect(testWorkloadType.Preemptibility).To(Equal(currentPreemptibility))
			Expect(testWorkloadType.Id).To(Equal(toUpdateId))
		})

		It("Update priority and category for external workload", func() {
			listWorkloadsResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listWorkloadsResponse.Code).To(Equal(http.StatusOK))

			listResponse, ok := listWorkloadsResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			workloadTypeObj := listResponse.Types[0]
			toUpdateId := workloadTypeObj.Id

			currentCategoryId := workloadTypeObj.CategoryId
			currentPriorityId := workloadTypeObj.PriorityId
			var newPriorityId string
			var newCategoryId string

			// find any other different priority
			for _, priority := range priorities {
				if priority.Id != currentPriorityId {
					newPriorityId = priority.Id
					break
				}
			}
			// find any other different category
			for _, category := range categories {
				if category.Id != currentCategoryId {
					newCategoryId = category.Id
					break
				}
			}
			newPreemptibility := api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE
			if workloadTypeObj.Preemptibility == api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE {
				newPreemptibility = api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE
			}
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:     newPriorityId,
				CategoryId:     newCategoryId,
				Preemptibility: ptr.To(newPreemptibility),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.PriorityId).To(Equal(newPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(newCategoryId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(newPreemptibility))
		})

		It("Update only the Priority - validate the category stays the same", func() {
			listWorkloadsResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listWorkloadsResponse.Code).To(Equal(http.StatusOK))

			listResponse, ok := listWorkloadsResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			var workloadTypeObj api.WorkloadTypeConfig
			for _, wt := range listResponse.Types {
				if len(wt.ResourceInterfaces) > 0 {
					workloadTypeObj = wt
					break
				}
			}
			Expect(workloadTypeObj.Name).ToNot(BeEmpty())
			toUpdateId := workloadTypeObj.Id
			risBefore := workloadTypeObj.ResourceInterfaces

			var anotherNewPriorityId string
			for _, priority := range priorities {
				// make sure the preemptibility also changes
				if priority.Id != workloadTypeObj.PriorityId &&
					computeExpectedPreemptibility(ctx, testing.TestTenant, priority.Id) != workloadTypeObj.Preemptibility {
					anotherNewPriorityId = priority.Id
					break
				}
			}
			Expect(anotherNewPriorityId).ToNot(BeEmpty())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: anotherNewPriorityId,
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.PriorityId).To(Equal(anotherNewPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(workloadTypeObj.CategoryId))
			expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, anotherNewPriorityId)
			Expect(updatedWorkloadType.Preemptibility).To(Equal(expectedPreemptibility))
			Expect(updatedWorkloadType.ResourceInterfaces).To(ConsistOf(risBefore))
		})

		It("Update only the Preemptibility - validate the category and priority stays the same", func() {
			listWorkloadsResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listWorkloadsResponse.Code).To(Equal(http.StatusOK))

			listResponse, ok := listWorkloadsResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			workloadTypeObj := listResponse.Types[0]
			toUpdateId := workloadTypeObj.Id
			currentPriorityId := workloadTypeObj.PriorityId
			currentCategoryId := workloadTypeObj.CategoryId
			currentPreemptibility := workloadTypeObj.Preemptibility

			newPreemptibility := api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE
			if currentPreemptibility == api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE {
				newPreemptibility = api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE
			}

			updateFields := api.WorkloadTypeUpdateFields{
				Preemptibility: ptr.To(newPreemptibility),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.PriorityId).To(Equal(currentPriorityId))
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(newPreemptibility))
		})

		It("Update only the Category - validate the priority+preemptibility stays the same", func() {
			listWorkloadsResponse, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listWorkloadsResponse.Code).To(Equal(http.StatusOK))

			listResponse, ok := listWorkloadsResponse.Body.(api.ListWorkloadTypes200Response)
			Expect(ok).To(BeTrue())
			workloadTypeObj := listResponse.Types[0]
			toUpdateId := workloadTypeObj.Id
			currentPriorityId := workloadTypeObj.PriorityId
			currentPreemptibility := workloadTypeObj.Preemptibility

			// pick a different category
			var newCategoryId string
			for _, c := range categories {
				if c.Id != workloadTypeObj.CategoryId {
					newCategoryId = c.Id
					break
				}
			}

			updateFields := api.WorkloadTypeUpdateFields{
				CategoryId: newCategoryId,
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.CategoryId).To(Equal(newCategoryId))
			Expect(updatedWorkloadType.PriorityId).To(Equal(currentPriorityId))
			Expect(updatedWorkloadType.Preemptibility).To(Equal(currentPreemptibility))
		})

		It("Add priority and preemptibility only for native workload", func() {
			nativeWorkloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "DistributedWorkload", false)
			Expect(nativeWorkloadTypeConfig).ToNot(BeNil())

			toUpdateId := nativeWorkloadTypeConfig.Id
			currentPriorityId := nativeWorkloadTypeConfig.PriorityId
			currentCategoryId := nativeWorkloadTypeConfig.CategoryId
			var newPriorityId string
			var newPriorityName string

			// find any other different priority
			for _, priority := range priorities {
				if priority.Id != currentPriorityId {
					newPriorityId = priority.Id
					newPriorityName = priority.Name
					break
				}
			}
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:     newPriorityId,
				Preemptibility: ptr.To(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())

			// Validate all fields of the response object for native workload type update
			Expect(updatedWorkloadType.Id).To(Equal(toUpdateId), "Id should match the updated workload type id")
			Expect(updatedWorkloadType.Name).To(Equal("DistributedWorkload"), "Name should be DistributedWorkload")
			Expect(updatedWorkloadType.Group).To(Equal("run.ai"), "Group should be run.ai for native workload types")
			Expect(updatedWorkloadType.Source).To(Equal(string(models.WorkloadTypeSourceNative)), "Source should be native")
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId), "CategoryId should remain unchanged")
			Expect(updatedWorkloadType.CategoryName).ToNot(BeEmpty(), "CategoryName should not be empty")
			Expect(updatedWorkloadType.PriorityId).To(Equal(newPriorityId), "PriorityId should match the new priority")
			Expect(updatedWorkloadType.PriorityName).To(Equal(newPriorityName), "PriorityName should match the new priority name")
			Expect(updatedWorkloadType.Preemptibility).To(Equal(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
			Expect(updatedWorkloadType.CreatedBy).ToNot(BeEmpty(), "CreatedBy should not be empty")
			Expect(updatedWorkloadType.CreatedAt.IsZero()).To(BeFalse(), "CreatedAt should not be zero")

			// update again without preemptibility - should match the new priority anyway
			var anotherPreemptiblePriorityId string
			var anotherPreemptiblePriorityName string
			for _, priority := range priorities {
				if priority.Value < 100 && priority.Id != newPriorityId && priority.Id != currentPriorityId {
					anotherPreemptiblePriorityId = priority.Id
					anotherPreemptiblePriorityName = priority.Name
					break
				}
			}
			updateFields2 := api.WorkloadTypeUpdateFields{
				PriorityId: anotherPreemptiblePriorityId,
			}
			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, toUpdateId, updateFields2)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())

			// Validate all fields again after second update
			Expect(updatedWorkloadType.Id).To(Equal(toUpdateId), "Id should match the updated workload type id")
			Expect(updatedWorkloadType.Name).To(Equal("DistributedWorkload"), "Name should be DistributedWorkload")
			Expect(updatedWorkloadType.Group).To(Equal("run.ai"), "Group should be run.ai for native workload types")
			Expect(updatedWorkloadType.Source).To(Equal(string(models.WorkloadTypeSourceNative)), "Source should be native")
			Expect(updatedWorkloadType.CategoryId).To(Equal(currentCategoryId), "CategoryId should remain unchanged")
			Expect(updatedWorkloadType.CategoryName).ToNot(BeEmpty(), "CategoryName should not be empty")
			Expect(updatedWorkloadType.PriorityId).To(Equal(anotherPreemptiblePriorityId), "PriorityId should match the new priority")
			Expect(updatedWorkloadType.PriorityName).To(Equal(anotherPreemptiblePriorityName), "PriorityName should match the new priority name")
			Expect(updatedWorkloadType.Preemptibility).To(Equal(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE))
			Expect(updatedWorkloadType.CreatedBy).ToNot(BeEmpty(), "CreatedBy should not be empty")
			Expect(updatedWorkloadType.CreatedAt.IsZero()).To(BeFalse(), "CreatedAt should not be zero")
		})

		It("Update with invalid uuid should return 400", func() {
			invalidId := "invalid-uuid"
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: priorities[0].Id,
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, invalidId, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("workload Id must be a valid uuid"))
		})

		It("Update with non-existing uuid should return 404", func() {
			nonExistingUid := uuid.NewString()
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: priorities[0].Id,
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, nonExistingUid, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("Update with invalid priority uuid should return 400", func() {
			invalidId := "invalid-uuid"
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: invalidId,
			}

			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "SeldonDeployment", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("priority Id must be a valid uuid"))
		})

		It("Update with invalid category uuid should return 400", func() {
			invalidId := "invalid-uuidddd"
			updateFields := api.WorkloadTypeUpdateFields{
				CategoryId: invalidId,
			}

			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "VirtualMachineInstance", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("category Id must be a valid uuid"))
		})

		It("should fail to update native workload type with resource interfaces", func() {
			train := findWorkloadTypeByName(ctx, workloadPropertiesService, "TrainingWorkload", false)
			Expect(train).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         train.PriorityId,
				CategoryId:         train.CategoryId,
				ResourceInterfaces: ptr.To(buildRDsForGvk("run.ai", "v2alpha1", "TrainingWorkload")),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, train.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("cannot set resource interfaces for native workload types"))
		})

		It("should fail to update with wrong group for resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         workloadTypeConfig.PriorityId,
				CategoryId:         workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildRDsForGvk("wrong-group", "v1", "Job")),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("group in kartas must match the group of the workload type"))
		})

		It("should fail to update with empty group for resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         workloadTypeConfig.PriorityId,
				CategoryId:         workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildRDsForGvk("", "v1", "Job")),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("karta validation failed: root component must have full kind (group, version, kind)"))
		})

		It("should fail to update with wrong kind for resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         workloadTypeConfig.PriorityId,
				CategoryId:         workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildRDsForGvk(workloadTypeConfig.Group, "v1", "NotJob")),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("kind in kartas must match the name of the workload type"))
		})

		It("should fail to update with empty kind for resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         workloadTypeConfig.PriorityId,
				CategoryId:         workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildRDsForGvk(workloadTypeConfig.Group, "v1", "")),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("karta validation failed: root component must have full kind (group, version, kind)"))
		})

		It("should fail to update with duplicate versions in resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: workloadTypeConfig.PriorityId,
				CategoryId: workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildResourceInterfaces([]schema.GroupVersionKind{{
					Group: workloadTypeConfig.Group, Version: "v1", Kind: "Job",
				}, {
					Group: workloadTypeConfig.Group, Version: "v1", Kind: "Job",
				}})),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("duplicate version"))
		})

		It("should succeed to update with different versions in resource interfaces", func() {
			workloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "Job", true)
			Expect(workloadTypeConfig).ToNot(BeNil())
			risInSystemScope := workloadTypeConfig.ResourceInterfaces

			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: workloadTypeConfig.PriorityId,
				CategoryId: workloadTypeConfig.CategoryId,
				ResourceInterfaces: ptr.To(buildResourceInterfaces([]schema.GroupVersionKind{{
					Group: workloadTypeConfig.Group, Version: "v1", Kind: "Job",
				}, {
					Group: workloadTypeConfig.Group, Version: "v2", Kind: "Job",
				}, {
					Group: workloadTypeConfig.Group, Version: "v2alpha2", Kind: "Job",
				}})),
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			response, err = workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok := response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(len(updatedWorkloadType.ResourceInterfaces)).To(Equal(3))
			validateExpectedResourceInterfaces(updatedWorkloadType.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
				{
					Version:           "v1",
					ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v1", "Job"),
				},
				{
					Version:           "v2",
					ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v2", "Job"),
				},
				{
					Version:           "v2alpha2",
					ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v2alpha2", "Job"),
				}})

			updateFields.ResourceInterfaces = ptr.To((*updateFields.ResourceInterfaces)[:2]) // remove one version
			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(len(updatedWorkloadType.ResourceInterfaces)).To(Equal(2))
			validateExpectedResourceInterfaces(updatedWorkloadType.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
				{
					Version:           "v1",
					ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v1", "Job"),
				},
				{
					Version:           "v2",
					ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v2", "Job"),
				}})

			// try to update with different RI for same version - API succeeds but the DB entry is not changed
			riv1 := buildRIWithSomeSpec(workloadTypeConfig.Group, "v1", "Job")
			(*updateFields.ResourceInterfaces)[0] = riv1
			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(len(updatedWorkloadType.ResourceInterfaces)).To(Equal(2))
			updatedAtRi0 := updatedWorkloadType.ResourceInterfaces[0].UpdatedAt
			updatedWorkloadType.ResourceInterfaces[0].UpdatedAt = time.Time{}
			Expect(updatedWorkloadType.ResourceInterfaces[0]).ToNot(Equal(api.WorkloadTypeResourceInterfacesWithVersionInner{
				Version: "v1", ResourceInterface: riv1})) // not updated, still the old RI
			Expect(updatedWorkloadType.ResourceInterfaces[0]).To(Equal(api.WorkloadTypeResourceInterfacesWithVersionInner{
				Version: "v1", ResourceInterface: buildValidResourceInterfaceForGvk(workloadTypeConfig.Group, "v1", "Job")}))

			risInTenantScope := updatedWorkloadType.ResourceInterfaces
			risInTenantScope[0].UpdatedAt = updatedAtRi0

			// update without specifying RIs - should not change the existing ones in tenant scope
			updateFields = api.WorkloadTypeUpdateFields{
				PriorityId: workloadTypeConfig.PriorityId,
			}
			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.ResourceInterfaces).To(ConsistOf(risInTenantScope))

			// update with empty RIs - should inherit from system scope and have the RI from before
			updateFields = api.WorkloadTypeUpdateFields{
				PriorityId:         workloadTypeConfig.PriorityId,
				ResourceInterfaces: &[]map[string]interface{}{},
			}
			response, err = workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			updatedWorkloadType, ok = response.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(updatedWorkloadType.ResourceInterfaces).To(ConsistOf(risInSystemScope))
		})
	})

	Context("Create, Delete, Sync Workload Types", Ordered, func() {
		var (
			validCategoryId        string
			validPriorityId        string
			validCategoryId2       string
			validPriorityId2       string
			createdWorkloadTypeIds []string
		)

		BeforeAll(func() {
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})

			// use system to get valida category and priority
			ctxSystem := testing.SetSystemContext(testing.TestUser)

			// Get valid category and priority IDs for testing
			categoriesResp, err := workloadPropertiesService.ListCategories(ctxSystem)
			Expect(err).To(BeNil())
			Expect(categoriesResp.Code).To(Equal(http.StatusOK))
			categories := categoriesResp.Body.(api.ListCategories200Response).Categories
			Expect(len(categories)).To(BeNumerically(">=", 2))
			validCategoryId = categories[0].Id
			validCategoryId2 = categories[1].Id

			prioritiesResp, err := workloadPropertiesService.GetWorkloadPriorities(ctxSystem, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(prioritiesResp.Code).To(Equal(http.StatusOK))
			priorities := prioritiesResp.Body.(api.GetWorkloadPriorities200Response).Priorities
			Expect(len(priorities)).To(BeNumerically(">=", 2))
			validPriorityId = priorities[0].Id
			validPriorityId2 = priorities[1].Id
		})

		AfterEach(func() {
			// delete the created workload types
			for _, id := range createdWorkloadTypeIds {
				_, err := workloadPropertiesService.DeleteWorkloadType(ctx, id)
				Expect(err).To(BeNil())
			}
			createdWorkloadTypeIds = []string{}

			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		})

		Context("CreateWorkloadType", func() {
			It("happy flow - should successfully create a workload type with valid parameters and also update it", func() {
				timeNow := time.Now()

				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					Preemptibility:     ptr.To(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE),
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)
				Expect(workloadTypeConfig.Name).To(Equal("TestWorkloadType"))
				Expect(workloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(workloadTypeConfig.PriorityId).To(Equal(validPriorityId))
				Expect(workloadTypeConfig.Preemptibility).To(Equal(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
				Expect(workloadTypeConfig.Id).ToNot(BeEmpty())
				Expect(workloadTypeConfig.UpdatedAt).To(BeTemporally("~", timeNow, time.Second))
				Expect(workloadTypeConfig.UpdatedBy).To(Equal(testing.TestUser))
				Expect(workloadTypeConfig.CreatedAt).To(BeTemporally("~", timeNow, time.Second))
				Expect(workloadTypeConfig.CreatedBy).To(Equal(testing.TestUser))
				Expect(workloadTypeConfig.Source).To(Equal(string(models.WorkloadTypeSourceCustom)))

				// Get the created workload type to verify
				getResponse, err := workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))
				createdWorkloadType, ok := getResponse.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(createdWorkloadType.Name).To(Equal("TestWorkloadType"))
				Expect(createdWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(createdWorkloadType.PriorityId).To(Equal(validPriorityId))
				Expect(createdWorkloadType.Preemptibility).To(Equal(api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE))
				Expect(createdWorkloadType.Id).To(Equal(createdWorkloadTypeId))
				Expect(createdWorkloadType.CreatedBy).To(Equal(testing.TestUser))
				Expect(workloadTypeConfig.Source).To(Equal(string(models.WorkloadTypeSourceCustom)))

				updateFields := api.WorkloadTypeUpdateFields{
					PriorityId: validPriorityId2,
				}
				updateResponse, err := workloadPropertiesService.UpdateWorkloadType(ctx, createdWorkloadTypeId, updateFields)
				Expect(err).To(BeNil())
				Expect(updateResponse.Code).To(Equal(http.StatusOK))

				updatedWorkloadType, ok := updateResponse.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(updatedWorkloadType.Name).To(Equal("TestWorkloadType"))
				Expect(updatedWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(updatedWorkloadType.PriorityId).To(Equal(validPriorityId2))
				expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, validPriorityId2)
				Expect(updatedWorkloadType.Preemptibility).To(Equal(expectedPreemptibility))
			})

			It("should fail to create workload type with missing name", func() {
				createFields := api.WorkloadTypeCreateFields{
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType"),
					Group:              "apps",
					// Name is missing
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("name must be provided"))

				createFields = api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind(""),
					Group:              "apps",
					// Name is missing in ResourceInterfaces
				}

				response, err = workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("karta validation failed: root component must have full kind (group, version, kind)"))
			})

			It("should fail to create workload type with missing group", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType2"),
					// group is missing
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("group must be provided"))

				createFields = api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildRDsForGvk("", "v1", "TestWorkloadType2"),
					// group is missing in ResourceInterfaces
				}

				response, err = workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("karta validation failed: root component must have full kind (group, version, kind)"))
			})

			It("should fail to create workload type with missing version", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildRDsForGvk("apps", "", "TestWorkloadType2"),
					// version is missing
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("karta validation failed: root component must have full kind (group, version, kind)"))
			})

			It("should fail to create workload type with kind doesn't match Name", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("DifferentKind"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("kind in kartas must match the name of the workload type"))
			})

			It("should fail to create workload type with missing categoryId", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "TestWorkloadType2",
					Group:      "apps",
					PriorityId: validPriorityId,
					// CategoryId is missing
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType2"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("both categoryId and priorityId must be provided"))
			})

			It("should fail to create workload type with missing priorityId", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "TestWorkloadType3",
					Group:      "apps",
					CategoryId: validCategoryId,
					// PriorityId is missing
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType3"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("both categoryId and priorityId must be provided"))
			})

			It("should fail to create workload type with invalid categoryId format", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType4",
					Group:              "apps",
					CategoryId:         "invalid-uuid",
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType4"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("category Id must be a valid uuid"))
			})

			It("should fail to create workload type with invalid priorityId format", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType5",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         "invalid-uuid",
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType5"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("priority Id must be a valid uuid"))
			})

			It("should fail to create workload type with duplicate name+group", func() {
				// First create a workload type
				createFields := api.WorkloadTypeCreateFields{
					Name:               "DuplicateWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("DuplicateWorkloadType"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)

				// Try to create another with the same gvk
				duplicateFields := api.WorkloadTypeCreateFields{
					Name:               "DuplicateWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId2,
					PriorityId:         validPriorityId2,
					ResourceInterfaces: buildGenericRDForKind("DuplicateWorkloadType"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, duplicateFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusConflict))

				// now delete and re-create - should pass
				response, err = workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				workloadTypeConfig = createWorkloadType(ctx, workloadPropertiesService, duplicateFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)
			})

			It("should succeed to create workload type with same name but different group", func() {
				// First create a workload type
				createFields := api.WorkloadTypeCreateFields{
					Name:               "DuplicateWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("DuplicateWorkloadType"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)

				// Try to create another with the same name - different group - should succeed
				createFields2 := api.WorkloadTypeCreateFields{
					Name:               "DuplicateWorkloadType",
					Group:              "apps2",
					CategoryId:         validCategoryId2,
					PriorityId:         validPriorityId2,
					ResourceInterfaces: buildRDsForGvk("apps2", "v1", "DuplicateWorkloadType"),
				}
				workloadTypeConfig = createWorkloadType(ctx, workloadPropertiesService, createFields2)
				createdWorkloadTypeId = workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)
			})

			It("should fail to create native workload type", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "TrainingWorkload",
					Group:      "run.ai",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusConflict))

				// RI has the native's TrainingWorkload gvk - should fail
				createFields = api.WorkloadTypeCreateFields{
					Name:               "TrainingWorkload2",
					Group:              "run.ai.different",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildRDsForGvk("run.ai.different", "v1", "TrainingWorkload"),
				}
				response, err = workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("should create workload type with different category and priority", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "DifferentConfigWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId2,
					PriorityId:         validPriorityId2,
					Preemptibility:     ptr.To(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE),
					ResourceInterfaces: buildGenericRDForKind("DifferentConfigWorkloadType"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

				Expect(workloadTypeConfig.Name).To(Equal("DifferentConfigWorkloadType"))
				Expect(workloadTypeConfig.CategoryId).To(Equal(validCategoryId2))
				Expect(workloadTypeConfig.PriorityId).To(Equal(validPriorityId2))
				Expect(workloadTypeConfig.Preemptibility).To(Equal(api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE))
			})

			It("should verify created workload type appears in list", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)

				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, "TestWorkloadType", true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.Id).To(Equal(createdWorkloadTypeId))
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId))
			})

			It("should succeed to create TrainingWorkload with different group", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TrainingWorkload",
					Group:              "custom.run.ai",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildRDsForGvk("custom.run.ai", "v1", "TrainingWorkload"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)

				Expect(workloadTypeConfig.Name).To(Equal("TrainingWorkload"))
				Expect(workloadTypeConfig.Group).To(Equal("custom.run.ai"))
				Expect(workloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(workloadTypeConfig.PriorityId).To(Equal(validPriorityId))
				// preemptibility derived from priority when omitted
				expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, validPriorityId)
				Expect(workloadTypeConfig.Preemptibility).To(Equal(expectedPreemptibility))
				Expect(len(workloadTypeConfig.ResourceInterfaces)).To(Equal(1))
				Expect(workloadTypeConfig.ResourceInterfaces[0].Version).To(Equal("v1"))
				Expect(workloadTypeConfig.ResourceInterfaces[0].ResourceInterface).To(Equal(buildRDsForGvk("custom.run.ai", "v1", "TrainingWorkload")[0]))
			})

			It("should succeed to create without resource interfaces", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "NoResourceDefWorkloadType",
					Group:      "apps",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
					// No resource interfaces
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)

				Expect(workloadTypeConfig.Name).To(Equal("NoResourceDefWorkloadType"))
				Expect(workloadTypeConfig.Group).To(Equal("apps"))
				Expect(workloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(workloadTypeConfig.PriorityId).To(Equal(validPriorityId))
				// preemptibility derived from priority when omitted
				expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, validPriorityId)
				Expect(workloadTypeConfig.Preemptibility).To(Equal(expectedPreemptibility))
				Expect(len(workloadTypeConfig.ResourceInterfaces)).To(Equal(0))
			})

			It("should fail to create with duplicate versions in resource interfaces", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "DuplicateVersionWorkloadType",
					Group:      "apps",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
					ResourceInterfaces: buildResourceInterfaces([]schema.GroupVersionKind{{
						Group: "apps", Version: "v1", Kind: "DuplicateVersionWorkloadType",
					}, {
						Group: "apps", Version: "v1", Kind: "DuplicateVersionWorkloadType",
					}}),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("kartas contain duplicate version"))
			})

			It("should fail to create with wrong group in resource interfaces", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "WrongGroupWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildRDsForGvk("wrong-group", "v1", "WrongGroupWorkloadType"),
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("group in kartas must match the group of the workload type"))
			})

			It("should fail to create native workload type", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "TrainingWorkload",
					Group:      "run.ai",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
				}

				response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusConflict))
			})

			It("should succeed to create with different versions in resource interfaces", func() {
				name := "MultiVersionWorkloadType"
				createFields := api.WorkloadTypeCreateFields{
					Name:       name,
					Group:      "apps",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
					ResourceInterfaces: buildResourceInterfaces([]schema.GroupVersionKind{
						{Group: "apps", Version: "v1", Kind: name},
						{Group: "apps", Version: "v2", Kind: name},
						{Group: "apps", Version: "v2alpha2", Kind: name},
					}),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

				Expect(workloadTypeConfig.Name).To(Equal(name))
				Expect(len(workloadTypeConfig.ResourceInterfaces)).To(Equal(3))
				validateExpectedResourceInterfaces(workloadTypeConfig.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version:           "v1",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v1", name),
					},
					{
						Version:           "v2",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v2", name),
					},
					{
						Version:           "v2alpha2",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v2alpha2", name),
					},
				})

				getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
				Expect(err).To(BeNil())
				Expect(getResp.Code).To(Equal(http.StatusOK))
				fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(len(fetched.ResourceInterfaces)).To(Equal(3))
				validateExpectedResourceInterfaces(fetched.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version:           "v1",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v1", name),
					},
					{
						Version:           "v2",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v2", name),
					},
					{
						Version:           "v2alpha2",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v2alpha2", name),
					},
				})
			})

			It("update ResourceInterface from db (not allowed in API) - validate updatedAt changed", func() {
				// create a workload type
				name := "DBUpdateWorkloadType"
				createFields := api.WorkloadTypeCreateFields{
					Name:       name,
					Group:      "apps",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
					ResourceInterfaces: buildResourceInterfaces([]schema.GroupVersionKind{
						{Group: "apps", Version: "v1", Kind: name},
						{Group: "apps", Version: "v2", Kind: name},
					}),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

				Expect(workloadTypeConfig.Name).To(Equal(name))
				Expect(len(workloadTypeConfig.ResourceInterfaces)).To(Equal(2))
				validateExpectedResourceInterfaces(workloadTypeConfig.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version:           "v1",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v1", name),
					},
					{
						Version:           "v2",
						ResourceInterface: buildValidResourceInterfaceForGvk("apps", "v2", name),
					},
				})
				updatedAtV1 := workloadTypeConfig.ResourceInterfaces[0].UpdatedAt
				updatedAtV2 := workloadTypeConfig.ResourceInterfaces[1].UpdatedAt
				if workloadTypeConfig.ResourceInterfaces[0].Version == "v2" {
					updatedAtV1 = workloadTypeConfig.ResourceInterfaces[1].UpdatedAt
					updatedAtV2 = workloadTypeConfig.ResourceInterfaces[0].UpdatedAt
				}

				// update the RI directly in the DB - simulating a change outside the API
				err := dbUpdateWorkloadKartasForTest(1, dao.WorkloadTypeCreateUpdateFieldsParsed{
					WorkloadTypeId: uuid.MustParse(workloadTypeConfig.Id),
					Kartas: &[]models.Karta{
						buildResourceInterfaceCrdWithSomeSpec("apps", "v1", name),
						buildResourceInterfaceCrdWithSomeSpec("apps", "v2", name),
					},
				})
				Expect(err).To(BeNil())

				// fetch again - validate updatedAt changed
				getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
				Expect(err).To(BeNil())
				Expect(getResp.Code).To(Equal(http.StatusOK))
				fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(len(fetched.ResourceInterfaces)).To(Equal(2))
				validateExpectedResourceInterfaces(fetched.ResourceInterfaces, []api.WorkloadTypeResourceInterfacesWithVersionInner{
					{
						Version:           "v1",
						ResourceInterface: buildRIWithSomeSpec("apps", "v1", name),
					},
					{
						Version:           "v2",
						ResourceInterface: buildRIWithSomeSpec("apps", "v2", name),
					},
				})

				updatedAtV1After := fetched.ResourceInterfaces[0].UpdatedAt
				updatedAtV2After := fetched.ResourceInterfaces[1].UpdatedAt
				if fetched.ResourceInterfaces[0].Version == "v2" {
					updatedAtV1After = fetched.ResourceInterfaces[1].UpdatedAt
					updatedAtV2After = fetched.ResourceInterfaces[0].UpdatedAt
				}

				Expect(updatedAtV1After).To(BeTemporally(">", updatedAtV1))
				Expect(updatedAtV2After).To(BeTemporally(">", updatedAtV2))

				// List and verify same
				// use findWorkloadTypeByName
				listed := findWorkloadTypeByName(ctx, workloadPropertiesService, name, true)
				Expect(listed).ToNot(BeNil())
				Expect(len(listed.ResourceInterfaces)).To(Equal(2))
				updatedAtV1FromList := listed.ResourceInterfaces[0].UpdatedAt
				updatedAtV2FromList := listed.ResourceInterfaces[1].UpdatedAt
				if listed.ResourceInterfaces[0].Version == "v2" {
					updatedAtV1FromList = listed.ResourceInterfaces[1].UpdatedAt
					updatedAtV2FromList = listed.ResourceInterfaces[0].UpdatedAt
				}

				Expect(updatedAtV1FromList).To(Equal(updatedAtV1After))
				Expect(updatedAtV2FromList).To(Equal(updatedAtV2After))
			})
		})

		Context("DeleteWorkloadType", func() {
			It("should successfully delete a user-defined workload type", func() {
				// First create a workload type to delete
				createFields := api.WorkloadTypeCreateFields{
					Name:               "WorkloadTypeToDelete",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("WorkloadTypeToDelete"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				workloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeId)

				// Now delete it
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, workloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// validate the WorkloadType was deleted from the table (together with the config, that's obvious)
				workloadTypeUuid := uuid.MustParse(workloadTypeId)
				_, dbErr := dao.DbGetWorkloadTypeById(ctx, workloadTypeUuid)
				Expect(dbErr).ToNot(BeNil())
				Expect(dbErr.Code).To(BeEquivalentTo(http.StatusNotFound))
			})

			It("should fail to delete workload type with invalid UUID", func() {
				response, err := workloadPropertiesService.DeleteWorkloadType(ctx, "invalid-uuid")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("invalid workload type id"))
			})

			It("should fail to delete non-existent workload type", func() {
				nonExistentId := uuid.NewString()
				response, err := workloadPropertiesService.DeleteWorkloadType(ctx, nonExistentId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})

			It("should fail to delete native workload type - even as system", func() {
				// Get a native workload type (like TrainingWorkload)
				trainingWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, "TrainingWorkload", false)
				Expect(trainingWorkloadType).ToNot(BeNil())
				nativeWorkloadId := trainingWorkloadType.Id

				// Try to delete the native workload type
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, nativeWorkloadId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
				Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))

				// Try to delete the native workload type as system
				ctx = testing.SetSystemContext(testing.TestUser)
				deleteResponse, err = workloadPropertiesService.DeleteWorkloadType(ctx, nativeWorkloadId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
				Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))
			})

			It("should succeed to delete external workload type (non-user-defined) in tenant scope; should fail for system scope", func() {
				// Get an external workload type (like Deployment)
				deploymentWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, "Deployment", true)
				Expect(deploymentWorkloadType).ToNot(BeNil())
				externalWorkloadId := deploymentWorkloadType.Id

				// update the external workload type to create a config for tenant
				response, err := workloadPropertiesService.UpdateWorkloadType(ctx, externalWorkloadId, api.WorkloadTypeUpdateFields{
					CategoryId: validCategoryId2,
					PriorityId: validPriorityId2,
				})
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				// Try to delete the external workload type as system
				ctx = testing.SetSystemContext(testing.TestUser)
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, externalWorkloadId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
				Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))

				// Now switch back to tenant context - delete should NOT work
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				deleteResponse, err = workloadPropertiesService.DeleteWorkloadType(ctx, externalWorkloadId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
				Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))

				// now list - should still see the same workload type in the list
				foundWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, "Deployment", true)
				Expect(foundWorkloadType).ToNot(BeNil())
				Expect(foundWorkloadType.Id).To(Equal(externalWorkloadId))
				Expect(foundWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(foundWorkloadType.PriorityId).To(Equal(validPriorityId2))

				getResponse, err := workloadPropertiesService.GetWorkloadType(ctx, externalWorkloadId)
				Expect(err).To(BeNil())
				Expect(getResponse.Code).To(Equal(http.StatusOK))
				apiWorkloadType, ok := getResponse.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(apiWorkloadType.Id).To(Equal(externalWorkloadId))
				Expect(apiWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(apiWorkloadType.PriorityId).To(Equal(validPriorityId2))
			})

			It("should verify deleted workload type no longer appears in list", func() {
				// Create a workload type
				createFields := api.WorkloadTypeCreateFields{
					Name:               "WorkloadTypeToVerifyDeletion",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("WorkloadTypeToVerifyDeletion"),
				}

				workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				workloadTypeId := workloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeId)

				// Verify it exists in the list
				testWorkloadTypeBefore := findWorkloadTypeByName(ctx, workloadPropertiesService, "WorkloadTypeToVerifyDeletion", true)
				Expect(testWorkloadTypeBefore).ToNot(BeNil())

				// Delete it
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, workloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// Verify it no longer exists in the list
				testWorkloadTypeAfter := findWorkloadTypeByName(ctx, workloadPropertiesService, "WorkloadTypeToVerifyDeletion", true)
				Expect(testWorkloadTypeAfter).To(BeNil())

				response, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("Create and Delete with different scopes", func() {
			It("should succeed to create workload type in two different tenant scopes", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType2"),
				}

				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfig.CreatedBy).To(Equal(testing.TestUser))
				Expect(createdWorkloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(createdWorkloadTypeConfig.PriorityId).To(Equal(validPriorityId))
				expectedPreemptibility := computeExpectedPreemptibility(ctx, testing.TestTenant, validPriorityId)
				Expect(createdWorkloadTypeConfig.Preemptibility).To(Equal(expectedPreemptibility))

				// Verify it's visible in the current tenant
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())

				// switch to a different tenant context
				ctx = testing.SetContext(testing.TestTenant2, testing.AnotherUser, []string{})

				// try to get in a different tenant - should not be visible
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())

				// try to delete in a different tenant - should not be allowed
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusForbidden))

				// try to create same with different config - in a different tenant - should be allowed
				createFields2 := api.WorkloadTypeCreateFields{
					Name:               createFields.Name,
					Group:              "apps",
					CategoryId:         validCategoryId2,
					PriorityId:         validPriorityId2,
					ResourceInterfaces: buildGenericRDForKind(createFields.Name),
				}
				createdWorkloadTypeConfigTenant2 := createWorkloadType(ctx, workloadPropertiesService, createFields2)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeConfigTenant2.Id)
				Expect(createdWorkloadTypeConfigTenant2.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfigTenant2.CreatedBy).To(Equal(testing.AnotherUser))
				Expect(createdWorkloadTypeConfigTenant2.Id).To(Equal(createdWorkloadTypeId))
				Expect(createdWorkloadTypeConfigTenant2.CategoryId).To(Equal(validCategoryId2))
				Expect(createdWorkloadTypeConfigTenant2.PriorityId).To(Equal(validPriorityId2))
				expectedPreemptibilityTenant2 := computeExpectedPreemptibility(ctx, testing.TestTenant2, validPriorityId2)
				Expect(createdWorkloadTypeConfigTenant2.Preemptibility).To(Equal(expectedPreemptibilityTenant2))

				// get+list for tenant 2 and validate
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.Id).To(Equal(createdWorkloadTypeConfigTenant2.Id))
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId2))
				Expect(testWorkloadType.Preemptibility).To(Equal(expectedPreemptibilityTenant2))
				response, err := workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				testWorkloadType2, ok := response.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(testWorkloadType2.Id).To(Equal(createdWorkloadTypeId))
				Expect(testWorkloadType2.CategoryId).To(Equal(validCategoryId2))
				Expect(testWorkloadType2.PriorityId).To(Equal(validPriorityId2))
				Expect(testWorkloadType2.Preemptibility).To(Equal(expectedPreemptibilityTenant2))

				// get+list for tenant 1 and validate
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId))
				response, err = workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				testWorkloadType2, ok = response.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(testWorkloadType2.Id).To(Equal(createdWorkloadTypeId))
				Expect(testWorkloadType2.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadType2.PriorityId).To(Equal(validPriorityId))
				Expect(testWorkloadType2.Preemptibility).To(Equal(expectedPreemptibility))

				// delete in tenant 2
				ctx = testing.SetContext(testing.TestTenant2, testing.AnotherUser, []string{})
				deleteResponse, err = workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeConfigTenant2.Id)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// validate the WorkloadType was not deleted from the table
				workloadTypeUuid := uuid.MustParse(createdWorkloadTypeId)
				modelWorkloadType, dbErr := dao.DbGetWorkloadTypeById(ctx, workloadTypeUuid)
				Expect(dbErr).To(BeNil())
				Expect(modelWorkloadType).ToNot(BeNil())

				// should still appear in Get for tenant 1
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId))
				Expect(testWorkloadType.Preemptibility).To(Equal(expectedPreemptibility))

				// delete in tenant 1
				deleteResponse, err = workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeConfigTenant2.Id)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())
			})

			It("should succeed to create workload type in tenant scope and then in system scope", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadType2"),
				}

				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfig.CreatedBy).To(Equal(testing.TestUser))
				Expect(createdWorkloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(createdWorkloadTypeConfig.PriorityId).To(Equal(validPriorityId))

				// Verify it's visible in the current tenant
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				response, err := workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				// switch to system context
				ctx = testing.SetSystemContext(testing.AnotherUser)

				// try to get/list in system scope - should not be visible
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())
				response, err = workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))

				// try to delete in system scope - should not be allowed
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusForbidden))

				// try to create same with different config - in system scope - should be allowed
				createFields2 := api.WorkloadTypeCreateFields{
					Name:               createFields.Name,
					Group:              "apps",
					CategoryId:         validCategoryId2,
					PriorityId:         validPriorityId2,
					ResourceInterfaces: buildGenericRDForKind(createFields.Name),
				}
				createdWorkloadTypeConfigSystem := createWorkloadType(ctx, workloadPropertiesService, createFields2)
				Expect(createdWorkloadTypeConfigSystem.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfigSystem.CreatedBy).To(Equal(testing.AnotherUser))
				Expect(createdWorkloadTypeConfigSystem.Id).To(Equal(createdWorkloadTypeId))
				Expect(createdWorkloadTypeConfigSystem.CategoryId).To(Equal(validCategoryId2))
				Expect(createdWorkloadTypeConfigSystem.PriorityId).To(Equal(validPriorityId2))

				// get/list for system scope and validate
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.Id).To(Equal(createdWorkloadTypeConfigSystem.Id))
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId2))
				response, err = workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				testWorkloadType2, ok := response.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(testWorkloadType2.Id).To(Equal(createdWorkloadTypeConfigSystem.Id))
				Expect(testWorkloadType2.CategoryId).To(Equal(validCategoryId2))
				Expect(testWorkloadType2.PriorityId).To(Equal(validPriorityId2))

				// delete in system scope
				deleteResponse, err = workloadPropertiesService.DeleteWorkloadType(ctx, createdWorkloadTypeConfigSystem.Id)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// should be deleted for tenant 1 as well
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())

				// validate the WorkloadType was deleted from the table (together with the config, that's obvious)
				workloadTypeUuid := uuid.MustParse(createdWorkloadTypeConfigSystem.Id)
				_, dbErr := dao.DbGetWorkloadTypeById(ctx, workloadTypeUuid)
				Expect(dbErr).ToNot(BeNil())
				Expect(dbErr.Code).To(BeEquivalentTo(http.StatusNotFound))
				_, dbErr = dao.DbGetWorkloadTypeConfigById(ctx, 1, workloadTypeUuid.String())
				Expect(dbErr).ToNot(BeNil())
				Expect(dbErr.Code).To(BeEquivalentTo(http.StatusNotFound))
			})

			It("should fail to create workload type in tenant scope after it already exists in system scope (for all sources)", func() {
				// Create a system-scoped workload type
				ctx = testing.SetSystemContext(testing.TestUser)
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadTypeSystem",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadTypeSystem"),
				}
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeConfig.Id)

				workloadTypeNamesToTest := []string{"TestWorkloadTypeSystem", "Deployment", "DistributedWorkload"}

				// switch to tenant ctx
				ctx = testing.SetContext(testing.TestTenant, testing.AnotherUser, []string{})

				// test creating a tenant-scoped workload type with the same name
				for _, workloadTypeName := range workloadTypeNamesToTest {
					workloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, workloadTypeName, false)
					Expect(workloadType).ToNot(BeNil())

					version := ""
					for _, rd := range workloadType.ResourceInterfaces {
						version = rd.Version
						break
					}

					// Try to create a tenant-scoped workload type with the same gvk
					createFields = api.WorkloadTypeCreateFields{
						Name:       workloadType.Name,
						Group:      workloadType.Group,
						CategoryId: validCategoryId,
						PriorityId: validPriorityId,
					}
					if version != "" {
						createFields.ResourceInterfaces = buildRDsForGvk(workloadType.Group, version, workloadType.Name)
					}
					response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
					Expect(err).To(BeNil())
					Expect(response.Code).To(Equal(http.StatusConflict), fmt.Sprintf("Should have conflict for name: %s", workloadType.Name))
				}

				// switch to system scope so "AfterEach" can delete it
				ctx = testing.SetSystemContext(testing.TestUser)
			})

			It("should fail to create same workload type in system scope if it already exists in system scope (for all sources)", func() {
				ctx = testing.SetSystemContext(testing.TestUser)

				createFields := api.WorkloadTypeCreateFields{
					Name:               "SystemScopedWorkloadType",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("SystemScopedWorkloadType"),
				}
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeConfig.Id)

				workloadTypeNamesToTest := []string{"SystemScopedWorkloadType", "JAXJob", "InteractiveWorkload"}

				// test creating a workload type with the same name in the same scope - system scope
				for _, workloadTypeName := range workloadTypeNamesToTest {
					workloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, workloadTypeName, false)
					Expect(workloadType).ToNot(BeNil())

					version := ""
					for _, rd := range workloadType.ResourceInterfaces {
						version = rd.Version
						break
					}

					// Try to create a tenant-scoped workload type with the same gvk
					createFields = api.WorkloadTypeCreateFields{
						Name:       workloadType.Name,
						Group:      workloadType.Group,
						CategoryId: validCategoryId,
						PriorityId: validPriorityId,
					}
					if version != "" {
						createFields.ResourceInterfaces = buildRDsForGvk(workloadType.Group, version, workloadType.Name)
					}
					response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
					Expect(err).To(BeNil())
					Expect(response.Code).To(Equal(http.StatusConflict), fmt.Sprintf("Should have conflict for name: %s", workloadType.Name))
				}
			})

			It("should fail to delete system-scoped WorkloadTypes as-tenant (of all sources)", func() {
				// Create a system-scoped workload type
				createFields := api.WorkloadTypeCreateFields{
					Name:               "SystemScopedWorkloadType2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("SystemScopedWorkloadType2"),
				}

				ctx = testing.SetSystemContext(testing.TestUser)
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadTypeId)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfig.CreatedBy).To(Equal(testing.TestUser))
				Expect(createdWorkloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(createdWorkloadTypeConfig.PriorityId).To(Equal(validPriorityId))

				// Switch to tenant context
				ctx = testing.SetContext(testing.TestTenant, testing.AnotherUser, []string{})

				workloadTypeNamesToTest := []string{"SystemScopedWorkloadType2", "ReplicaSet", "DistributedInferenceWorkload"}

				// Try to delete the system-scoped workload type as a tenant user - should fail
				for i, workloadTypeName := range workloadTypeNamesToTest {
					workloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, workloadTypeName, false)
					Expect(workloadType).ToNot(BeNil())

					// Try to delete the system-scoped workload type
					deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, workloadType.Id)
					Expect(err).To(BeNil())

					if i == 0 {
						// for user_external - should fail with forbidden
						Expect(deleteResponse.Code).To(Equal(http.StatusForbidden), fmt.Sprintf("Should fail deleting for workloadTypeName: %s", workloadTypeName))
					} else {
						// for external and native - should fail with bad request
						Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest), fmt.Sprintf("Should fail deleting for workloadTypeName: %s", workloadTypeName))
						Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))
					}
				}

				// Verify it still exists as the system configured for tenant as well
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				response, err := workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))
				testWorkloadTypeConfig, ok := response.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(testWorkloadTypeConfig.Id).To(Equal(createdWorkloadTypeId))
				Expect(testWorkloadTypeConfig.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadTypeConfig.PriorityId).To(Equal(validPriorityId))

				// try to create as tenant - should fail cause already exists
				response, err = workloadPropertiesService.CreateWorkloadType(ctx, createFields)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusConflict), "Should not be able to create a workload type with the same name in tenant scope when it exists in system scope")
			})

			It("should fail to delete non-user-external sourced as system scope", func() {
				ctx = testing.SetSystemContext(testing.TestUser)

				workloadTypeNamesToTest := []string{"CronJob", "InferenceWorkload"}

				// Try to delete as system scope - should fail
				for _, workloadTypeName := range workloadTypeNamesToTest {
					workloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, workloadTypeName, false)
					Expect(workloadType).ToNot(BeNil())

					// Try to delete the system-scoped workload type
					deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, workloadType.Id)
					Expect(err).To(BeNil())

					// for other sources - should fail with bad request - not allowed to delete
					Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest), fmt.Sprintf("Should fail deleting for workloadTypeName: %s", workloadTypeName))
					Expect(deleteResponse.Body.(api.Error).Message).To(ContainSubstring("is not a user-defined type and cannot be deleted"))
				}
			})

			It("create as tenant -> create as system -> delete as system - should verify no longer appears in list for tenant", func() {
				// create as tenant
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadTypeForDelete",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadTypeForDelete"),
				}
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				testWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, testWorkloadTypeId)

				// create as system
				ctx = testing.SetSystemContext(testing.AnotherUser)
				createdWorkloadTypeConfigSystem := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfigSystem.Name).To(Equal(createFields.Name))
				Expect(createdWorkloadTypeConfigSystem.Id).To(Equal(testWorkloadTypeId))

				// delete as system
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, testWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// verify it no longer appears in list for system
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())

				// verify it no longer appears in list for tenant
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())

				response, err := workloadPropertiesService.GetWorkloadType(ctx, testWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})

			It("create as system -> update to create as tenant -> delete as system - should verify no longer appears in list for tenant", func() {
				// create as system
				ctx = testing.SetSystemContext(testing.AnotherUser)
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadTypeForDelete2",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadTypeForDelete2"),
				}
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				testWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, testWorkloadTypeId)

				// update to create as tenant
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				updateFields := api.WorkloadTypeUpdateFields{
					CategoryId: validCategoryId2,
					PriorityId: validPriorityId2,
				}
				updateResponse, err := workloadPropertiesService.UpdateWorkloadType(ctx, testWorkloadTypeId, updateFields)
				Expect(err).To(BeNil())
				Expect(updateResponse.Code).To(Equal(http.StatusOK))

				// delete as system
				ctx = testing.SetSystemContext(testing.AnotherUser)
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, testWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

				// verify it no longer appears in list for tenant
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).To(BeNil())

				response, err := workloadPropertiesService.GetWorkloadType(ctx, testWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})

			It("create as system -> update to create as tenant -> delete as tenant - should fail, verify still appears in list for system and for tenant", func() {
				// create as system
				ctx = testing.SetSystemContext(testing.AnotherUser)
				createFields := api.WorkloadTypeCreateFields{
					Name:               "TestWorkloadTypeForDelete3",
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind("TestWorkloadTypeForDelete3"),
				}
				createdWorkloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
				Expect(createdWorkloadTypeConfig.Name).To(Equal(createFields.Name))
				testWorkloadTypeId := createdWorkloadTypeConfig.Id
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, testWorkloadTypeId)

				// update to create as tenant
				ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
				updateFields := api.WorkloadTypeUpdateFields{
					CategoryId: validCategoryId2,
					PriorityId: validPriorityId2,
				}
				updateResponse, err := workloadPropertiesService.UpdateWorkloadType(ctx, testWorkloadTypeId, updateFields)
				Expect(err).To(BeNil())
				Expect(updateResponse.Code).To(Equal(http.StatusOK))

				// delete as tenant - should fail with forbidden
				deleteResponse, err := workloadPropertiesService.DeleteWorkloadType(ctx, testWorkloadTypeId)
				Expect(err).To(BeNil())
				Expect(deleteResponse.Code).To(Equal(http.StatusForbidden))

				// verify it still appears in list for tenant - with tenant config
				testWorkloadType := findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.Id).To(Equal(testWorkloadTypeId))
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId2))

				// verify it still appears in list for system - with system config
				ctx = testing.SetSystemContext(testing.AnotherUser)
				testWorkloadType = findWorkloadTypeByName(ctx, workloadPropertiesService, createFields.Name, true)
				Expect(testWorkloadType).ToNot(BeNil())
				Expect(testWorkloadType.Id).To(Equal(testWorkloadTypeId))
				Expect(testWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(testWorkloadType.PriorityId).To(Equal(validPriorityId))
			})

			It("should succeed to create and then update workload type in system scope", func() {
				// create in system scope
				ctx = testing.SetSystemContext(testing.TestUser)
				name := fmt.Sprintf("SystemCreateUpdate-%s", rand.String(5))
				createFields := api.WorkloadTypeCreateFields{
					Name:               name,
					Group:              "apps",
					CategoryId:         validCategoryId,
					PriorityId:         validPriorityId,
					ResourceInterfaces: buildGenericRDForKind(name),
				}
				createdWorkloadType := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, createdWorkloadType.Id)
				Expect(createdWorkloadType.Name).To(Equal(name))
				Expect(createdWorkloadType.CategoryId).To(Equal(validCategoryId))
				Expect(createdWorkloadType.PriorityId).To(Equal(validPriorityId))
				Expect(createdWorkloadType.Source).To(Equal(string(models.WorkloadTypeSourceCustom)))

				// update in system scope
				updateFields := api.WorkloadTypeUpdateFields{
					CategoryId: validCategoryId2,
					PriorityId: validPriorityId2,
				}
				updateResponse, err := workloadPropertiesService.UpdateWorkloadType(ctx, createdWorkloadType.Id, updateFields)
				Expect(err).To(BeNil())
				Expect(updateResponse.Code).To(Equal(http.StatusOK), fmt.Sprintf("response body: %+v", updateResponse.Body))
				updatedWorkloadType, ok := updateResponse.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(updatedWorkloadType.Id).To(Equal(createdWorkloadType.Id))
				Expect(updatedWorkloadType.Name).To(Equal(name))
				Expect(updatedWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(updatedWorkloadType.PriorityId).To(Equal(validPriorityId2))

				// verify via Get
				getResp, err := workloadPropertiesService.GetWorkloadType(ctx, createdWorkloadType.Id)
				Expect(err).To(BeNil())
				Expect(getResp.Code).To(Equal(http.StatusOK))
				gotWorkloadType, ok := getResp.Body.(api.WorkloadTypeConfig)
				Expect(ok).To(BeTrue())
				Expect(gotWorkloadType.Id).To(Equal(createdWorkloadType.Id))
				Expect(gotWorkloadType.Name).To(Equal(name))
				Expect(gotWorkloadType.CategoryId).To(Equal(validCategoryId2))
				Expect(gotWorkloadType.PriorityId).To(Equal(validPriorityId2))
			})
		})

		Context("SyncWorkloadTypes", func() {
			var (
				validClusterId string

				clusterId1 = testing.TestUuid1
				clusterId2 = testing.TestUuid2
				clusterId3 = testing.TestUuid3
			)

			BeforeEach(func() {
				validClusterId = uuid.NewString()
			})

			It("SyncWorkloadTypes - Success with FullSync=false", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusFailed, StatusCreating, StatusUpdating}
				initialIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// sync some more with FullSync=false
				newPhases := []string{StatusReady, StatusReady, StatusReady}
				newIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, false, newPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// validate all initial remain with their phases, and new ones were added/updated
				expectedIdToStatus := buildExpectedIdToStatus(append(initialIds, newIds...), append(initialPhases, newPhases...))
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)

				// validate without status
				listResp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, true)
				ExpectWithOffset(1, err).To(BeNil())
				ExpectWithOffset(1, listResp.Code).To(Equal(http.StatusOK))
				listBody, ok := listResp.Body.(api.ListWorkloadTypes200Response)
				ExpectWithOffset(1, ok).To(BeTrue())
				for _, workloadType := range listBody.Types {
					Expect(workloadType.ClusterStatuses).To(BeNil())
				}
			})

			It("SyncWorkloadTypes - Success with FullSync=true", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusFailed, StatusCreating, StatusUpdating, StatusReady}
				initialIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// sync some more with FullSync=true
				newPhases := []string{StatusReady, StatusReady, StatusReady}
				newIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, newPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// validate only the new ones exist; initial ones deleted
				expectedIdToStatus := buildExpectedIdToStatus(newIds, newPhases)
				for _, id := range initialIds {
					expectedIdToStatus[id] = nil
				}
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)
			})

			It("SyncWorkloadTypes - Delete workload types with DeletedPhase - fullSync false", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusCreating, StatusReady, StatusFailed, StatusReady}
				workloadTypeIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// sync some of them with Deleted phase
				updatePhases := []string{DeletedPhase, DeletedPhase, StatusFailed, DeletedPhase}
				syncWorkloadTypes(ctx, workloadPropertiesService, validClusterId, false, updatePhases, workloadTypeIds[:4])

				// expected: 0,1,3 deleted; 2 updated to StatusFailed; 4 unchanged
				expectedIdToStatus := map[string]*api.WorkloadTypeStatus{
					workloadTypeIds[0]: nil,
					workloadTypeIds[1]: nil,
					workloadTypeIds[2]: {Phase: updatePhases[2], Version: "v1"},
					workloadTypeIds[3]: nil,
					workloadTypeIds[4]: {Phase: initialPhases[4], Version: "v1"},
				}
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)
			})

			It("SyncWorkloadTypes - Delete workload types with DeletedPhase - fullSync true", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusCreating, StatusReady, StatusFailed, StatusReady}
				workloadTypeIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// sync some with Deleted phase
				updatePhases := []string{DeletedPhase, StatusFailed, DeletedPhase}
				syncWorkloadTypes(ctx, workloadPropertiesService, validClusterId, true, updatePhases, workloadTypeIds[:3])

				// expected: only index 1 remains with StatusFailed; others deleted (including 3,4 due to full sync)
				expectedIdToStatus := map[string]*api.WorkloadTypeStatus{}
				for i := range workloadTypeIds {
					if i == 1 {
						expectedIdToStatus[workloadTypeIds[i]] = &api.WorkloadTypeStatus{Phase: StatusFailed, Version: "v1"}
					} else {
						expectedIdToStatus[workloadTypeIds[i]] = nil
					}
				}
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)
			})

			It("SyncWorkloadTypes - Invalid cluster ID", func() {
				workloadTypesForSync := api.WorkloadTypesForSync{
					FullSync: false,
					WorkloadTypes: []api.WorkloadTypeForSync{
						{
							Id:      uuid.NewString(),
							Phase:   StatusReady,
							Version: "v1",
						},
					},
				}

				response, err := workloadPropertiesService.SyncWorkloadTypes(ctx, "invalid-cluster-id", workloadTypesForSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("invalid cluster id"))

				workloadTypesForSync.FullSync = true
				response, err = workloadPropertiesService.SyncWorkloadTypes(ctx, "invalid-cluster-id", workloadTypesForSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body).To(ContainSubstring("invalid cluster id"))
			})

			It("SyncWorkloadTypes - Empty workload types list", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusFailed, StatusCreating}
				initialIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				workloadTypesForSync := api.WorkloadTypesForSync{
					FullSync:      false,
					WorkloadTypes: []api.WorkloadTypeForSync{},
				}

				// now sync with empty list
				response, err := workloadPropertiesService.SyncWorkloadTypes(ctx, validClusterId, workloadTypesForSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// now validate the existing were not deleted
				expectedIdToStatus := buildExpectedIdToStatus(initialIds, initialPhases)
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)
			})

			It("SyncWorkloadTypes - Empty workload types list with full sync - delete all other statuses", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusFailed, StatusCreating}
				initialIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				workloadTypesForSync := api.WorkloadTypesForSync{
					FullSync:      true,
					WorkloadTypes: []api.WorkloadTypeForSync{},
				}

				// now sync with empty list and full sync
				response, err := workloadPropertiesService.SyncWorkloadTypes(ctx, validClusterId, workloadTypesForSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// now validate they were deleted
				expectedIdToStatus := map[string]*api.WorkloadTypeStatus{}
				for _, id := range initialIds {
					expectedIdToStatus[id] = nil
				}
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expectedIdToStatus)
			})

			It("SyncWorkloadTypes - Non-existent workload type IDs", func() {
				workloadTypes := []api.WorkloadTypeForSync{}
				phases := []string{StatusReady, StatusFailed, StatusCreating, StatusUpdating, StatusReady}

				for i := 0; i < 5; i++ {
					workloadTypes = append(workloadTypes, api.WorkloadTypeForSync{
						Id:      uuid.NewString(),
						Phase:   phases[i],
						Version: "v1",
					})
				}

				workloadTypesForSync := api.WorkloadTypesForSync{
					FullSync:      false,
					WorkloadTypes: workloadTypes,
				}

				response, err := workloadPropertiesService.SyncWorkloadTypes(ctx, validClusterId, workloadTypesForSync)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))

				// validate they were not written to db
				for _, workloadType := range workloadTypes {
					_, dbErr := dao.DbGetWorkloadTypeById(ctx, uuid.MustParse(workloadType.Id))
					Expect(dbErr).ToNot(BeNil(), fmt.Sprintf("Expected error for non-existent workload type ID: %s", workloadType.Id))
					Expect(dbErr.Code).To(BeEquivalentTo(http.StatusNotFound), fmt.Sprintf("Expected NotFound error for ID: %s", workloadType.Id))
				}
			})

			It("SyncWorkloadTypes - Invalid workload type ID format", func() {
				// sync some workload types first
				initialPhases := []string{StatusReady, StatusFailed, StatusCreating}
				workloadTypeIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					validClusterId, true, initialPhases, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				workloadTypeIdsToSync := []string{workloadTypeIds[0], "invalid-uuid-format", workloadTypeIds[1]}
				// reuse phases slice to update id0 and id1 accordingly
				syncWorkloadTypes(ctx, workloadPropertiesService, validClusterId, true, initialPhases, workloadTypeIdsToSync)

				// expected: id0 = initialPhases[0], id1 = initialPhases[1], id2 deleted due to full sync
				expected := map[string]*api.WorkloadTypeStatus{
					workloadTypeIds[0]: {Phase: initialPhases[0], Version: "v1"},
					workloadTypeIds[1]: {Phase: initialPhases[2], Version: "v1"},
					workloadTypeIds[2]: nil,
				}
				assertClusterStatuses(ctx, workloadPropertiesService, validClusterId, expected)
			})

			It("SyncWorkloadTypes - Multiple clusters with same workload types", func() {
				// Test sync with multiple clusters

				// sync some in cluster1
				initialPhasesCluster1 := []string{StatusReady, StatusFailed, StatusCreating, StatusReady, StatusReady}
				workloadTypeIds := createAndSyncWorkloadTypes(ctx, workloadPropertiesService,
					clusterId1, true, initialPhasesCluster1, validCategoryId, validPriorityId, &createdWorkloadTypeIds)

				// now sync the same ones in cluster2 - with different phases
				phasesCluster2 := []string{StatusReady, StatusFailed, StatusCreating, StatusUpdating, StatusUpdating}
				syncWorkloadTypes(ctx, workloadPropertiesService, clusterId2, true, phasesCluster2, workloadTypeIds)

				// now sync the same ones (partial) in cluster3 - with different phases
				phasesCluster3 := []string{StatusReady, StatusFailed, StatusCreating}
				syncWorkloadTypes(ctx, workloadPropertiesService, clusterId3, true, phasesCluster3, workloadTypeIds[:3])

				// validate the workload types were written to db for all clusters
				expectedIdToStatusCluster1 := buildExpectedIdToStatus(workloadTypeIds, initialPhasesCluster1)
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId1, expectedIdToStatusCluster1)
				expectedIdToStatusCluster2 := buildExpectedIdToStatus(workloadTypeIds, phasesCluster2)
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId2, expectedIdToStatusCluster2)
				expectedIdToStatusCluster3 := buildExpectedIdToStatus(workloadTypeIds[:3], phasesCluster3)
				expectedIdToStatusCluster3[workloadTypeIds[3]] = nil
				expectedIdToStatusCluster3[workloadTypeIds[4]] = nil
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId3, expectedIdToStatusCluster3)

				// validate some were deleted in cluster2, but not in cluster1 and cluster3
				phasesCluster2 = []string{StatusReady, StatusFailed, StatusCreating}
				syncWorkloadTypes(ctx, workloadPropertiesService, clusterId2, true, phasesCluster2, workloadTypeIds[:3])
				// expectedIdToStatusCluster1 and expectedIdToStatusCluster3 were already built - should stay the same
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId1, expectedIdToStatusCluster1)
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId3, expectedIdToStatusCluster3)
				expectedIdToStatusCluster2 = buildExpectedIdToStatus(workloadTypeIds[:3], phasesCluster2)
				for _, id := range workloadTypeIds[3:] {
					expectedIdToStatusCluster2[id] = nil
				}
				assertClusterStatuses(ctx, workloadPropertiesService, clusterId2, expectedIdToStatusCluster2)

				// using ctx of user with permissions only in clusters 1 and 2, validate the response
				ctxClusters12 := testing.SetContext(testing.TestTenant, testing.UserWithCluster1And2Permissions, []string{})
				expectedIdToStatusCluster3 = buildExpectedIdToNilStatus(workloadTypeIds)
				assertClusterStatuses(ctxClusters12, workloadPropertiesService, clusterId3, expectedIdToStatusCluster3)

				// expectedIdToStatusCluster1+2 were already built - should stay the same
				assertClusterStatuses(ctxClusters12, workloadPropertiesService, clusterId1, expectedIdToStatusCluster1)
				assertClusterStatuses(ctxClusters12, workloadPropertiesService, clusterId2, expectedIdToStatusCluster2)
			})

			It("SyncWorkloadTypes - with conditions and versions - validate in Get as well", func() {
				// prepare 3 workload types
				ids := []string{}
				for i := 0; i < 3; i++ {
					name := fmt.Sprintf("CondVerWorkloadType%d-%s", i, rand.String(3))
					createFields := api.WorkloadTypeCreateFields{
						Name:               name,
						Group:              "apps",
						CategoryId:         validCategoryId,
						PriorityId:         validPriorityId,
						ResourceInterfaces: buildGenericRDForKind(name),
					}
					created := createWorkloadType(ctx, workloadPropertiesService, createFields)
					ids = append(ids, created.Id)
					createdWorkloadTypeIds = append(createdWorkloadTypeIds, created.Id)
				}

				cluster1 := uuid.NewString()
				cluster2 := uuid.NewString()

				// build versions and conditions for both clusters
				versionsC1 := []string{"v1", "v2", "v2alpha1"}
				versionsC2 := []string{"v1beta1", "v2", "v3"}

				condsFor := func(tag string) []api.Condition {
					return []api.Condition{
						{Type: "Available", Status: "True", Message: fmt.Sprintf("available-%s", tag), Reason: "Synced"},
						{Type: "Failed", Status: "False", Message: fmt.Sprintf("no-error-%s", tag), Reason: "None"},
					}
				}

				conditionsC1 := [][]api.Condition{
					condsFor("c1-0"),
					condsFor("c1-1"),
					condsFor("c1-2"),
				}
				conditionsC2 := [][]api.Condition{
					condsFor("c2-0"),
					condsFor("c2-1"),
					condsFor("c2-2"),
				}

				// choose phases deterministically
				phasesC1 := []string{StatusReady, StatusFailed, StatusUpdating}
				phasesC2 := []string{StatusCreating, StatusReady, StatusFailed}

				// sync for both clusters with details
				syncWorkloadTypesWithDetails(ctx, workloadPropertiesService, cluster1, true, ids, phasesC1, versionsC1, conditionsC1)
				syncWorkloadTypesWithDetails(ctx, workloadPropertiesService, cluster2, true, ids, phasesC2, versionsC2, conditionsC2)

				// assert details via List and Get for cluster1
				expectedC1 := map[string]ExpectedWorkloadTypeStatusDetails{}
				for i, id := range ids {
					expectedC1[id] = ExpectedWorkloadTypeStatusDetails{Phase: phasesC1[i], Version: versionsC1[i], Conditions: conditionsC1[i]}
				}
				assertClusterStatusesWithDetails(ctx, workloadPropertiesService, cluster1, expectedC1)

				// assert details via List and Get for cluster2
				expectedC2 := map[string]ExpectedWorkloadTypeStatusDetails{}
				for i, id := range ids {
					expectedC2[id] = ExpectedWorkloadTypeStatusDetails{Phase: phasesC2[i], Version: versionsC2[i], Conditions: conditionsC2[i]}
				}
				assertClusterStatusesWithDetails(ctx, workloadPropertiesService, cluster2, expectedC2)
			})

			It("SyncWorkloadTypes - with different versions for same WorkloadType - see conditions being overriden", func() {
				createFields := api.WorkloadTypeCreateFields{
					Name:       "TypeWithVersions",
					Group:      "apps",
					CategoryId: validCategoryId,
					PriorityId: validPriorityId,
					ResourceInterfaces: buildResourceInterfaces([]schema.GroupVersionKind{
						{Group: "apps", Version: "v1", Kind: "TypeWithVersions"},
						{Group: "apps", Version: "v2", Kind: "TypeWithVersions"},
					}),
				}
				created := createWorkloadType(ctx, workloadPropertiesService, createFields)
				createdWorkloadTypeIds = append(createdWorkloadTypeIds, created.Id)

				workloadTypesForSync := api.WorkloadTypesForSync{
					FullSync: false,
					WorkloadTypes: []api.WorkloadTypeForSync{
						{
							Id:      created.Id,
							Phase:   "Ready",
							Version: "v1",
							Conditions: []api.Condition{{
								Type: "Available", Status: "True", Message: "available-v1", Reason: "Synced",
							}},
						},
						{
							Id:      created.Id,
							Phase:   "NotReady",
							Version: "v2",
							Conditions: []api.Condition{{
								Type: "Available", Status: "True", Message: "available-v1", Reason: "Synced",
							}, {
								Type: "SyncEnabled", Status: "False", Message: "available-v1", Reason: "Synced",
							}},
						},
					},
				}

				clusterId := uuid.NewString()
				response, err := workloadPropertiesService.SyncWorkloadTypes(ctx, clusterId, workloadTypesForSync)
				ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to sync workload types, error: %v", err))
				ExpectWithOffset(1, response.Code).To(Equal(http.StatusNoContent), fmt.Sprintf("Unexpected response code: %d", response.Code))

				// GET and validate phase and conditions per version
				response, err = workloadPropertiesService.GetWorkloadType(ctx, created.Id)
				ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to get workload type, error: %v", err))
				ExpectWithOffset(1, response.Code).To(Equal(http.StatusOK), fmt.Sprintf("Unexpected response code: %d", response.Code))
				workloadTypeConfig, ok := response.Body.(api.WorkloadTypeConfig)
				ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("Response body is not of type WorkloadTypeConfig: %T", response.Body))
				ExpectWithOffset(1, workloadTypeConfig.Id).To(Equal(created.Id))
				ExpectWithOffset(1, workloadTypeConfig.ClusterStatuses).ToNot(BeNil())
				ExpectWithOffset(1, len(*workloadTypeConfig.ClusterStatuses)).To(Equal(2))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].ClusterId).To(Equal(clusterId))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Version).To(Equal("v1"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Phase).To(Equal("Ready"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Conditions).To(Equal([]api.Condition{{
					Type: "Available", Status: "True", Message: "available-v1", Reason: "Synced",
				}}))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].ClusterId).To(Equal(clusterId))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Version).To(Equal("v2"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Phase).To(Equal("NotReady"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Conditions).To(Equal([]api.Condition{
					{Type: "Available", Status: "True", Message: "available-v1", Reason: "Synced"},
					{Type: "SyncEnabled", Status: "False", Message: "available-v1", Reason: "Synced"},
				}))

				// now send another sync with changed conditions and phase - validate they are updated
				workloadTypesForSync.WorkloadTypes[0].Phase = "NotReady"
				workloadTypesForSync.WorkloadTypes[0].Conditions = []api.Condition{{
					Type: "Available", Status: "False", Message: "not-available-v1", Reason: "Error",
				}}
				workloadTypesForSync.WorkloadTypes[1].Phase = "Ready"
				workloadTypesForSync.WorkloadTypes[1].Conditions = []api.Condition{{
					Type: "Available", Status: "True", Message: "available-v2", Reason: "Synced",
				}}

				timeNow := time.Now()
				response, err = workloadPropertiesService.SyncWorkloadTypes(ctx, clusterId, workloadTypesForSync)
				ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to sync workload types, error: %v", err))
				ExpectWithOffset(1, response.Code).To(Equal(http.StatusNoContent), fmt.Sprintf("Unexpected response code: %d", response.Code))

				// GET and validate updated phase and conditions per version
				response, err = workloadPropertiesService.GetWorkloadType(ctx, created.Id)
				ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to get workload type, error: %v", err))
				ExpectWithOffset(1, response.Code).To(Equal(http.StatusOK), fmt.Sprintf("Unexpected response code: %d", response.Code))
				workloadTypeConfig, ok = response.Body.(api.WorkloadTypeConfig)
				ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("Response body is not of type WorkloadTypeConfig: %T", response.Body))
				ExpectWithOffset(1, workloadTypeConfig.Id).To(Equal(created.Id))
				ExpectWithOffset(1, workloadTypeConfig.ClusterStatuses).ToNot(BeNil())
				ExpectWithOffset(1, len(*workloadTypeConfig.ClusterStatuses)).To(Equal(2))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].ClusterId).To(Equal(clusterId))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Version).To(Equal("v1"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Phase).To(Equal("NotReady"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].Conditions).To(Equal([]api.Condition{{
					Type: "Available", Status: "False", Message: "not-available-v1", Reason: "Error",
				}}))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[0].UpdatedAt).To(BeTemporally("~", timeNow, time.Second))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].ClusterId).To(Equal(clusterId))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Version).To(Equal("v2"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Phase).To(Equal("Ready"))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].Conditions).To(Equal([]api.Condition{{
					Type: "Available", Status: "True", Message: "available-v2", Reason: "Synced",
				}}))
				ExpectWithOffset(1, (*workloadTypeConfig.ClusterStatuses)[1].UpdatedAt).To(BeTemporally("~", timeNow, time.Second))
			})
		})
	})

	Context("Get/List - forbidden", Ordered, func() {
		var (
			unpermittedUserCtx = testing.SetContext(testing.TestTenant, testing.UnpermittedUser, []string{})
			emptyScopesUserCtx = testing.SetContext(testing.TestTenant, testing.UserWithEmptyScopes, []string{})
		)

		AfterAll(func() {
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		})

		It("should return NotFound for GetWorkloadType", func() {
			// find an existing workload type to test with
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
			nativeWorkloadTypeConfig := findWorkloadTypeByName(ctx, workloadPropertiesService, "DistributedWorkload", false)
			Expect(nativeWorkloadTypeConfig).ToNot(BeNil())
			response, err := workloadPropertiesService.GetWorkloadType(ctx, nativeWorkloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			// switch to unpermitted context - should fail
			response, err = workloadPropertiesService.GetWorkloadType(unpermittedUserCtx, nativeWorkloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))

			response, err = workloadPropertiesService.GetWorkloadType(emptyScopesUserCtx, nativeWorkloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should return forbidden for ListWorkloadTypes", func() {
			response, err := workloadPropertiesService.ListWorkloadTypes(unpermittedUserCtx, false, []string{}, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusForbidden))

			response, err = workloadPropertiesService.ListWorkloadTypes(emptyScopesUserCtx, false, []string{}, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should return forbidden for ListCategories", func() {
			response, err := workloadPropertiesService.ListCategories(unpermittedUserCtx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusForbidden))

			response, err = workloadPropertiesService.ListCategories(emptyScopesUserCtx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should return NotFound for GetCategory", func() {
			// find an existing category to test with
			ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
			categoriesResponse, err := workloadPropertiesService.ListCategories(ctx)
			Expect(err).To(BeNil())
			Expect(categoriesResponse.Code).To(Equal(http.StatusOK))
			categoriesList, ok := categoriesResponse.Body.(api.ListCategories200Response)
			Expect(ok).To(BeTrue())
			Expect(len(categoriesList.Categories)).To(BeNumerically(">", 0))
			Expect(categoriesList.Categories[0].Id).ToNot(BeEmpty())
			testCategoryId := categoriesList.Categories[0].Id

			response, err := workloadPropertiesService.GetCategoryById(ctx, testCategoryId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			// switch to unpermitted context - should fail
			response, err = workloadPropertiesService.GetCategoryById(unpermittedUserCtx, testCategoryId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))

			response, err = workloadPropertiesService.GetCategoryById(emptyScopesUserCtx, testCategoryId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})

	Context("kartas / resourceInterfaces backward compatibility", Ordered, func() {
		var (
			validCategoryId        string
			validPriorityId        string
			createdWorkloadTypeIds []string
		)

		BeforeAll(func() {
			ctxSystem := testing.SetSystemContext(testing.TestUser)

			categoriesResp, err := workloadPropertiesService.ListCategories(ctxSystem)
			Expect(err).To(BeNil())
			Expect(categoriesResp.Code).To(Equal(http.StatusOK))
			categories := categoriesResp.Body.(api.ListCategories200Response).Categories
			Expect(len(categories)).To(BeNumerically(">=", 1))
			validCategoryId = categories[0].Id

			prioritiesResp, err := workloadPropertiesService.GetWorkloadPriorities(ctxSystem, "", "", []string{})
			Expect(err).To(BeNil())
			Expect(prioritiesResp.Code).To(Equal(http.StatusOK))
			priorities := prioritiesResp.Body.(api.GetWorkloadPriorities200Response).Priorities
			Expect(len(priorities)).To(BeNumerically(">=", 1))
			validPriorityId = priorities[0].Id
		})

		AfterEach(func() {
			for _, id := range createdWorkloadTypeIds {
				_, err := workloadPropertiesService.DeleteWorkloadType(ctx, id)
				Expect(err).To(BeNil())
			}
			createdWorkloadTypeIds = []string{}
		})

		It("POST with only kartas - both fields populated in singular GET and list", func() {
			name := "BCKartasOnlyCreate"
			kartas := buildRDsForGvk("apps", "v1", name)

			createFields := api.WorkloadTypeCreateFields{
				Name:       name,
				Group:      "apps",
				CategoryId: validCategoryId,
				PriorityId: validPriorityId,
				Kartas:     kartas,
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			Expect(workloadTypeConfig.Kartas).To(HaveLen(1))
			Expect(workloadTypeConfig.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck // intentionally checking deprecated field for backward compat
			Expect(workloadTypeConfig.Kartas[0].Karta).To(Equal(kartas[0]))
			Expect(workloadTypeConfig.ResourceInterfaces[0].ResourceInterface).To(Equal(kartas[0])) //nolint:staticcheck

			getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(fetched.Kartas).To(HaveLen(1))
			Expect(fetched.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck
			Expect(fetched.Kartas[0].Karta).To(Equal(kartas[0]))
			Expect(fetched.ResourceInterfaces[0].ResourceInterface).To(Equal(kartas[0])) //nolint:staticcheck

			listResp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			listed := listResp.Body.(api.ListWorkloadTypes200Response).Types
			found := findWorkloadTypeInListById(listed, workloadTypeConfig.Id)
			Expect(found).ToNot(BeNil())
			Expect(found.Kartas).To(HaveLen(1))
			Expect(found.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck
			Expect(found.Kartas[0].Karta).To(Equal(kartas[0]))
			Expect(found.ResourceInterfaces[0].ResourceInterface).To(Equal(kartas[0])) //nolint:staticcheck
		})

		It("POST with only resourceInterfaces - both fields populated in singular GET and list", func() {
			name := "BCResourceInterfacesOnlyCreate"
			ris := buildRDsForGvk("apps", "v1", name)

			createFields := api.WorkloadTypeCreateFields{
				Name:               name,
				Group:              "apps",
				CategoryId:         validCategoryId,
				PriorityId:         validPriorityId,
				ResourceInterfaces: ris,
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			Expect(workloadTypeConfig.Kartas).To(HaveLen(1))
			Expect(workloadTypeConfig.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck
			Expect(workloadTypeConfig.Kartas[0].Karta).To(Equal(ris[0]))
			Expect(workloadTypeConfig.ResourceInterfaces[0].ResourceInterface).To(Equal(ris[0])) //nolint:staticcheck

			getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(fetched.Kartas).To(HaveLen(1))
			Expect(fetched.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck

			listResp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			listed := listResp.Body.(api.ListWorkloadTypes200Response).Types
			found := findWorkloadTypeInListById(listed, workloadTypeConfig.Id)
			Expect(found).ToNot(BeNil())
			Expect(found.Kartas).To(HaveLen(1))
			Expect(found.ResourceInterfaces).To(HaveLen(1)) //nolint:staticcheck
		})

		It("POST with both kartas and resourceInterfaces returns 400", func() {
			createFields := api.WorkloadTypeCreateFields{
				Name:               "BCBothFieldsCreate",
				Group:              "apps",
				CategoryId:         validCategoryId,
				PriorityId:         validPriorityId,
				Kartas:             buildRDsForGvk("apps", "v1", "BCBothFieldsCreate"),
				ResourceInterfaces: buildRDsForGvk("apps", "v1", "BCBothFieldsCreate"),
			}

			response, err := workloadPropertiesService.CreateWorkloadType(ctx, createFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("only one of kartas or resourceInterfaces may be set"))
		})

		It("PUT with only kartas - both fields populated in singular GET and list", func() {
			name := "BCKartasOnlyUpdate"
			createFields := api.WorkloadTypeCreateFields{
				Name:               name,
				Group:              "apps",
				CategoryId:         validCategoryId,
				PriorityId:         validPriorityId,
				ResourceInterfaces: buildRDsForGvk("apps", "v1", name),
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			updatedKartas := buildResourceInterfaces([]schema.GroupVersionKind{
				{Group: "apps", Version: "v1", Kind: name},
				{Group: "apps", Version: "v2", Kind: name},
			})
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: validPriorityId,
				CategoryId: validCategoryId,
				Kartas:     &updatedKartas,
			}

			updateResp, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.WorkloadTypeConfig)
			Expect(updated.Kartas).To(HaveLen(2))
			Expect(updated.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck

			getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(fetched.Kartas).To(HaveLen(2))
			Expect(fetched.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck

			listResp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			listed := listResp.Body.(api.ListWorkloadTypes200Response).Types
			found := findWorkloadTypeInListById(listed, workloadTypeConfig.Id)
			Expect(found).ToNot(BeNil())
			Expect(found.Kartas).To(HaveLen(2))
			Expect(found.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck
		})

		It("PUT with only resourceInterfaces - both fields populated in singular GET and list", func() {
			name := "BCResourceInterfacesOnlyUpdate"
			createFields := api.WorkloadTypeCreateFields{
				Name:       name,
				Group:      "apps",
				CategoryId: validCategoryId,
				PriorityId: validPriorityId,
				Kartas:     buildRDsForGvk("apps", "v1", name),
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			updatedRIs := buildResourceInterfaces([]schema.GroupVersionKind{
				{Group: "apps", Version: "v1", Kind: name},
				{Group: "apps", Version: "v2", Kind: name},
			})
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         validPriorityId,
				CategoryId:         validCategoryId,
				ResourceInterfaces: &updatedRIs,
			}

			updateResp, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.WorkloadTypeConfig)
			Expect(updated.Kartas).To(HaveLen(2))
			Expect(updated.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck

			getResp, err := workloadPropertiesService.GetWorkloadType(ctx, workloadTypeConfig.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched, ok := getResp.Body.(api.WorkloadTypeConfig)
			Expect(ok).To(BeTrue())
			Expect(fetched.Kartas).To(HaveLen(2))
			Expect(fetched.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck

			listResp, err := workloadPropertiesService.ListWorkloadTypes(ctx, true, []string{}, false)
			Expect(err).To(BeNil())
			listed := listResp.Body.(api.ListWorkloadTypes200Response).Types
			found := findWorkloadTypeInListById(listed, workloadTypeConfig.Id)
			Expect(found).ToNot(BeNil())
			Expect(found.Kartas).To(HaveLen(2))
			Expect(found.ResourceInterfaces).To(HaveLen(2)) //nolint:staticcheck
		})

		It("PUT with both kartas and resourceInterfaces returns 400", func() {
			name := "BCBothFieldsUpdate"
			createFields := api.WorkloadTypeCreateFields{
				Name:       name,
				Group:      "apps",
				CategoryId: validCategoryId,
				PriorityId: validPriorityId,
				Kartas:     buildRDsForGvk("apps", "v1", name),
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			both := buildRDsForGvk("apps", "v1", name)
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId:         validPriorityId,
				Kartas:             &both,
				ResourceInterfaces: &both,
			}

			response, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("only one of kartas or resourceInterfaces may be set"))
		})

		It("PUT with explicit empty kartas - clears existing entries (kartas wins over RIs)", func() {
			name := "BCEmptyKartasUpdate"
			createFields := api.WorkloadTypeCreateFields{
				Name:       name,
				Group:      "apps",
				CategoryId: validCategoryId,
				PriorityId: validPriorityId,
				Kartas:     buildRDsForGvk("apps", "v1", name),
			}
			workloadTypeConfig := createWorkloadType(ctx, workloadPropertiesService, createFields)
			createdWorkloadTypeIds = append(createdWorkloadTypeIds, workloadTypeConfig.Id)

			emptyKartas := []map[string]interface{}{}
			updateFields := api.WorkloadTypeUpdateFields{
				PriorityId: validPriorityId,
				Kartas:     &emptyKartas,
			}

			updateResp, err := workloadPropertiesService.UpdateWorkloadType(ctx, workloadTypeConfig.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
		})
	})
})

func findWorkloadTypeInListById(types []api.WorkloadTypeConfig, id string) *api.WorkloadTypeConfig {
	for i := range types {
		if types[i].Id == id {
			return &types[i]
		}
	}
	return nil
}

func createWorkloadType(ctx context.Context, service api.WorkloadPropertiesAPIServicer, createFields api.WorkloadTypeCreateFields) api.WorkloadTypeConfig {
	response, err := service.CreateWorkloadType(ctx, createFields)
	ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to create workload type, error: %v", err))
	ExpectWithOffset(1, response.Code).To(Equal(http.StatusCreated), fmt.Sprintf("Unexpected response code: %d; body: %v", response.Code, response.Body))

	workloadTypeConfig, ok := response.Body.(api.WorkloadTypeConfig)
	ExpectWithOffset(1, ok).To(BeTrue(), fmt.Sprintf("Response body is not of type WorkloadTypeConfig: %T", response.Body))
	ExpectWithOffset(1, workloadTypeConfig.Id).ToNot(BeEmpty(), fmt.Sprintf("Workload type ID should not be empty, got: %s", workloadTypeConfig.Id))

	return workloadTypeConfig
}

// computeExpectedPreemptibility returns expected preemptibility based on the priority value (>=100 => non-preemptible)
func computeExpectedPreemptibility(ctx context.Context, tenantId string, priorityId string) api.WorkloadTypePreemptibility {
	priority, dbErr := dao.GetPriorityById(ctx, tenantId, priorityId)
	ExpectWithOffset(1, dbErr).To(BeNil(), fmt.Sprintf("failed to get priority by id %s", priorityId))
	if priority.Value >= 100 {
		return api.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE
	}
	return api.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE
}

func findWorkloadTypeByName(ctx context.Context, service api.WorkloadPropertiesAPIServicer, workloadTypeName string, externalTypesOnly bool) *api.WorkloadTypeConfig {
	response, err := service.ListWorkloadTypes(ctx, externalTypesOnly, []string{}, false)
	Expect(err).To(BeNil())
	Expect(response.Code).To(Equal(http.StatusOK))

	listResponse, ok := response.Body.(api.ListWorkloadTypes200Response)
	Expect(ok).To(BeTrue())
	for _, workloadType := range listResponse.Types {
		if workloadType.Name == workloadTypeName {
			return &workloadType
		}
	}
	return nil

}

func syncWorkloadTypes(ctx context.Context, service api.WorkloadPropertiesAPIServicer,
	clusterId string, fullSync bool, phases []string, existingWorkloadTypeIdsToSync []string) {
	ExpectWithOffset(1, len(phases)).To(Equal(len(existingWorkloadTypeIdsToSync)), "Phases and workload type IDs must have the same length")

	workloadTypeStatuses := []api.WorkloadTypeForSync{}
	for i := 0; i < len(phases); i++ {
		workloadTypeStatuses = append(workloadTypeStatuses, api.WorkloadTypeForSync{
			Id:      existingWorkloadTypeIdsToSync[i],
			Phase:   phases[i],
			Version: "v1",
		})
	}

	workloadTypesForSync := api.WorkloadTypesForSync{
		FullSync:      fullSync,
		WorkloadTypes: workloadTypeStatuses,
	}

	response, err := service.SyncWorkloadTypes(ctx, clusterId, workloadTypesForSync)
	ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to sync workload types, error: %v", err))
	ExpectWithOffset(1, response.Code).To(Equal(http.StatusNoContent), fmt.Sprintf("Unexpected response code: %d", response.Code))
}

func createAndSyncWorkloadTypes(ctx context.Context, service api.WorkloadPropertiesAPIServicer,
	clusterId string, fullSync bool, phases []string,
	validCategoryId string, validPriorityId string, createdWorkloadTypeIds *[]string) []string {
	workloadTypeIds := []string{}

	workloadTypeStatuses := []api.WorkloadTypeForSync{}
	for i := 0; i < len(phases); i++ {
		name := fmt.Sprintf("TestWorkloadTypeForSync%d-%s", i, rand.String(3))
		createFields := api.WorkloadTypeCreateFields{
			Name:               name,
			Group:              "apps",
			CategoryId:         validCategoryId,
			PriorityId:         validPriorityId,
			ResourceInterfaces: buildGenericRDForKind(name),
		}
		createdWorkloadType := createWorkloadType(ctx, service, createFields)
		workloadTypeIds = append(workloadTypeIds, createdWorkloadType.Id)
		*createdWorkloadTypeIds = append(*createdWorkloadTypeIds, createdWorkloadType.Id)

		workloadTypeStatuses = append(workloadTypeStatuses, api.WorkloadTypeForSync{
			Id:      createdWorkloadType.Id,
			Phase:   phases[i],
			Version: "v1",
		})
	}

	workloadTypesForSync := api.WorkloadTypesForSync{
		FullSync:      fullSync,
		WorkloadTypes: workloadTypeStatuses,
	}

	response, err := service.SyncWorkloadTypes(ctx, clusterId, workloadTypesForSync)
	ExpectWithOffset(1, err).To(BeNil(), "Failed to sync workload types, error: %v", err)
	ExpectWithOffset(1, response.Code).To(Equal(http.StatusNoContent), "Unexpected response code: %d", response.Code)

	return workloadTypeIds
}

func buildExpectedIdToStatus(ids []string, phases []string) map[string]*api.WorkloadTypeStatus {
	ExpectWithOffset(1, len(ids)).To(Equal(len(phases)))
	expectedIdToStatus := map[string]*api.WorkloadTypeStatus{}
	for i := range ids {
		expectedIdToStatus[ids[i]] = &api.WorkloadTypeStatus{
			Phase:   phases[i],
			Version: "v1",
		}
	}
	return expectedIdToStatus
}

func buildExpectedIdToNilStatus(ids []string) map[string]*api.WorkloadTypeStatus {
	expectedIdToStatus := map[string]*api.WorkloadTypeStatus{}
	for _, id := range ids {
		expectedIdToStatus[id] = nil
	}
	return expectedIdToStatus
}

func assertClusterStatuses(ctx context.Context, service api.WorkloadPropertiesAPIServicer, clusterId string, expectedIdToStatus map[string]*api.WorkloadTypeStatus) {
	// Validate via list
	listResp, err := service.ListWorkloadTypes(ctx, true, []string{}, false)
	ExpectWithOffset(1, err).To(BeNil())
	ExpectWithOffset(1, listResp.Code).To(Equal(http.StatusOK))
	listBody, ok := listResp.Body.(api.ListWorkloadTypes200Response)
	ExpectWithOffset(1, ok).To(BeTrue())
	idToWorkloadTypeMap := map[string]api.WorkloadTypeConfig{}
	for _, workloadType := range listBody.Types {
		idToWorkloadTypeMap[workloadType.Id] = workloadType
	}

	// For each expected id, assert in both list and get
	for id, expectedStatus := range expectedIdToStatus {
		// from list
		workloadType, exists := idToWorkloadTypeMap[id]
		ExpectWithOffset(1, exists).To(BeTrue(), fmt.Sprintf("workload type %s not found in list", id))
		assertSingleClusterStatus(clusterId, expectedStatus, workloadType.ClusterStatuses)

		// from get
		getResp, err := service.GetWorkloadType(ctx, id)
		ExpectWithOffset(1, err).To(BeNil())
		ExpectWithOffset(1, getResp.Code).To(Equal(http.StatusOK))
		workloadType, ok = getResp.Body.(api.WorkloadTypeConfig)
		ExpectWithOffset(1, ok).To(BeTrue())
		assertSingleClusterStatus(clusterId, expectedStatus, workloadType.ClusterStatuses)
	}
}

func assertSingleClusterStatus(clusterId string, expectedStatus *api.WorkloadTypeStatus, clusterStatuses *[]api.WorkloadTypeStatus) {
	// find cluster status in the slice
	var found bool
	var phase string
	if clusterStatuses != nil {
		for _, s := range *clusterStatuses {
			if s.ClusterId == clusterId {
				found = true
				phase = s.Phase
				break
			}
		}
	}

	if expectedStatus == nil {
		ExpectWithOffset(2, found).To(BeFalse(), fmt.Sprintf("expected to not find status for cluster %s", clusterId))
	} else {
		ExpectWithOffset(2, found).To(BeTrue(), fmt.Sprintf("expected to find status for cluster %s", clusterId))
		ExpectWithOffset(2, phase).To(Equal(expectedStatus.Phase), "expected phase for cluster %s to be %s, got %s", clusterId, expectedStatus.Phase, phase)
	}
}

// Expected structure for verifying status details (phase, version, conditions)
type ExpectedWorkloadTypeStatusDetails struct {
	Phase      string
	Version    string
	Conditions []api.Condition
}

func syncWorkloadTypesWithDetails(
	ctx context.Context,
	service api.WorkloadPropertiesAPIServicer,
	clusterId string,
	fullSync bool,
	ids []string,
	phases []string,
	versions []string,
	conditions [][]api.Condition,
) {
	ExpectWithOffset(1, len(ids)).To(Equal(len(phases)))
	ExpectWithOffset(1, len(ids)).To(Equal(len(versions)))
	ExpectWithOffset(1, len(ids)).To(Equal(len(conditions)))

	workloadTypeStatuses := []api.WorkloadTypeForSync{}
	for i := 0; i < len(ids); i++ {
		workloadTypeStatuses = append(workloadTypeStatuses, api.WorkloadTypeForSync{
			Id:         ids[i],
			Phase:      phases[i],
			Version:    versions[i],
			Conditions: conditions[i],
		})
	}

	workloadTypesForSync := api.WorkloadTypesForSync{
		FullSync:      fullSync,
		WorkloadTypes: workloadTypeStatuses,
	}

	response, err := service.SyncWorkloadTypes(ctx, clusterId, workloadTypesForSync)
	ExpectWithOffset(1, err).To(BeNil(), fmt.Sprintf("Failed to sync workload types with details, error: %v", err))
	ExpectWithOffset(1, response.Code).To(Equal(http.StatusNoContent), fmt.Sprintf("Unexpected response code: %d", response.Code))
}

func assertClusterStatusesWithDetails(
	ctx context.Context,
	service api.WorkloadPropertiesAPIServicer,
	clusterId string,
	expectedIdToDetails map[string]ExpectedWorkloadTypeStatusDetails,
) {
	// Validate via list
	listResp, err := service.ListWorkloadTypes(ctx, true, []string{}, false)
	ExpectWithOffset(1, err).To(BeNil(), "error in ListWorkloadTypes")
	ExpectWithOffset(1, listResp.Code).To(Equal(http.StatusOK), "unexpected response code in ListWorkloadTypes")
	listBody, ok := listResp.Body.(api.ListWorkloadTypes200Response)
	ExpectWithOffset(1, ok).To(BeTrue())
	idToWorkloadTypeMap := map[string]api.WorkloadTypeConfig{}
	for _, workloadType := range listBody.Types {
		idToWorkloadTypeMap[workloadType.Id] = workloadType
	}

	for id, expected := range expectedIdToDetails {
		workloadType, exists := idToWorkloadTypeMap[id]
		ExpectWithOffset(1, exists).To(BeTrue(), fmt.Sprintf("workload type %s not found in list", id))
		assertSingleClusterStatusWithDetails(clusterId, expected, workloadType.ClusterStatuses)

		getResp, err := service.GetWorkloadType(ctx, id)
		ExpectWithOffset(1, err).To(BeNil(), "error in GetWorkloadType")
		ExpectWithOffset(1, getResp.Code).To(Equal(http.StatusOK), "unexpected response code in GetWorkloadType")
		workloadType, ok = getResp.Body.(api.WorkloadTypeConfig)
		ExpectWithOffset(1, ok).To(BeTrue())
		assertSingleClusterStatusWithDetails(clusterId, expected, workloadType.ClusterStatuses)
	}
}

func assertSingleClusterStatusWithDetails(
	clusterId string,
	expected ExpectedWorkloadTypeStatusDetails,
	clusterStatuses *[]api.WorkloadTypeStatus,
) {
	var found bool
	var actual api.WorkloadTypeStatus
	if clusterStatuses != nil {
		for _, s := range *clusterStatuses {
			if s.ClusterId == clusterId {
				found = true
				actual = s
				break
			}
		}
	}

	ExpectWithOffset(2, found).To(BeTrue(), fmt.Sprintf("expected to find status for cluster %s", clusterId))
	ExpectWithOffset(2, actual.Phase).To(Equal(expected.Phase), "expected phase for cluster %s to be %s, got %s", clusterId, expected.Phase, actual.Phase)
	ExpectWithOffset(2, actual.Version).To(Equal(expected.Version), "expected version for cluster %s to be %s, got %s", clusterId, expected.Version, actual.Version)

	// Compare conditions content-wise (order-insensitive, at least expected present)
	for _, expCond := range expected.Conditions {
		ExpectWithOffset(2, containsCondition(actual.Conditions, expCond)).To(BeTrue(),
			fmt.Sprintf("expected condition not found (type=%s,status=%s)", expCond.Type, expCond.Status))
	}
}

func containsCondition(actual []api.Condition, expected api.Condition) bool {
	for _, c := range actual {
		if c.Type == expected.Type && c.Status == expected.Status && c.Message == expected.Message && c.Reason == expected.Reason {
			return true
		}
	}
	return false
}

func buildGenericRDForKind(kind string) []map[string]interface{} {
	return buildRDsForGvk("apps", "v1", kind)
}

func buildRDsForGvk(group, version, kind string) []map[string]interface{} {
	return buildResourceInterfaces([]schema.GroupVersionKind{{Group: group, Version: version, Kind: kind}})
}

func buildResourceInterfaces(gvks []schema.GroupVersionKind) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, gvk := range gvks {
		result = append(result, buildValidResourceInterfaceForGvk(gvk.Group, gvk.Version, gvk.Kind))
	}
	return result
}

func buildValidResourceInterfaceForGvk(group, version, kind string) map[string]interface{} {
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"structureDefinition": map[string]interface{}{
				"rootComponent": map[string]interface{}{
					"kind": map[string]interface{}{
						"group":   group,
						"version": version,
						"kind":    kind,
					},
					"name": kind,
					"statusDefinition": map[string]interface{}{
						"phaseDefinition": map[string]interface{}{
							"path": ".status.phase",
						},
						"statusMappings": map[string]interface{}{},
					},
				},
			},
			"optimizationInstructions": map[string]interface{}{},
		},
		"metadata": map[string]interface{}{},
		"status":   map[string]interface{}{},
	}
}

func buildK8sDeploymentResourceInterface() map[string]interface{} {
	ri := kubernetes.BuildK8sDeploymentV1()
	bytes, err := json.Marshal(ri)
	Expect(err).To(BeNil(), "Failed to marshal Deployment ResourceInterface")
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	Expect(err).To(BeNil(), "Failed to unmarshal Deployment ResourceInterface")
	return result
}

func buildK8sPodResourceInterface() map[string]interface{} {
	ri := kubernetes.BuildK8sPodV1()
	bytes, err := json.Marshal(ri)
	Expect(err).To(BeNil(), "Failed to marshal Pod ResourceInterface")
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	Expect(err).To(BeNil(), "Failed to unmarshal Pod ResourceInterface")
	return result
}

func buildK8sJobResourceInterface() map[string]interface{} {
	ri := kubernetes.BuildK8sJobV1()
	bytes, err := json.Marshal(ri)
	Expect(err).To(BeNil(), "Failed to marshal Job ResourceInterface")
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	Expect(err).To(BeNil(), "Failed to unmarshal Job ResourceInterface")
	return result
}

func buildResourceInterfaceCrdWithSomeSpec(group, version, kind string) models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   group,
						Version: version,
						Kind:    kind,
					},
					Name: kind,
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: ptr.To(".spec.template"),
					},
					StatusDefinition: &kartav1alpha1.StatusDefinition{
						PhaseDefinition: &kartav1alpha1.PhaseDefinition{
							Path: ".status.phase",
						},
					},
				},
			},
		},
	}
}

func buildRIWithSomeSpec(group, version, kind string) map[string]interface{} {
	ri := ptr.To(buildResourceInterfaceCrdWithSomeSpec(group, version, kind))

	bytes, err := json.Marshal(ri)
	Expect(err).To(BeNil(), "Failed to marshal ResourceInterface")
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	Expect(err).To(BeNil(), "Failed to unmarshal ResourceInterface")
	return result
}

func validateExpectedResourceInterfaces(actual, expected []api.WorkloadTypeResourceInterfacesWithVersionInner) {
	Expect(actual).To(HaveLen(len(expected)), "ResourceInterfaces length mismatch")

	actualByVer := getResourceInterfacesByVersion(actual)
	expectedByVer := getResourceInterfacesByVersion(expected)

	for ver, expRI := range expectedByVer {
		actRI, exists := actualByVer[ver]
		Expect(exists).To(BeTrue(), fmt.Sprintf("ResourceInterface for version %s not found", ver))
		Expect(actRI.Version).To(Equal(expRI.Version), "Version mismatch for version %s", ver)
		Expect(actRI.ResourceInterface).To(HaveLen(len(expRI.ResourceInterface)), "ResourceInterface length mismatch for version %s", ver)

		// UpdatedAt comes from the db, so just validate it's recent and then ignore for deep equal
		Expect(actRI.UpdatedAt).To(BeTemporally("~", time.Now(), 2*time.Minute),
			"UpdatedAt mismatch for version %s: expected %v, got %v", ver, expRI.UpdatedAt, actRI.UpdatedAt)

		Expect(actRI.ResourceInterface).To(Equal(expRI.ResourceInterface), "ResourceInterface content mismatch for version %s", ver)
	}
}

func getResourceInterfacesByVersion(ris []api.WorkloadTypeResourceInterfacesWithVersionInner) map[string]api.WorkloadTypeResourceInterfacesWithVersionInner {
	result := map[string]api.WorkloadTypeResourceInterfacesWithVersionInner{}
	for _, ri := range ris {
		result[ri.Version] = ri
	}
	return result
}

// we don't allow update kartas using code/API, but we do expect maybe a migration to change the karta
// so this is a helper func for this
func dbUpdateWorkloadKartasForTest(tenantId uint, fields dao.WorkloadTypeCreateUpdateFieldsParsed) error {
	if fields.Kartas == nil || len(*fields.Kartas) == 0 {
		return nil
	}

	ctx := context.Background()
	timeNow := time.Now()

	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, karta := range *fields.Kartas {
		query := tx.Table(models.WorkloadTypeKarta{}.TableName()).
			Where("workload_type_id = ?", fields.WorkloadTypeId).
			Where("version = ?", karta.GroupVersionKind().Version)

		if isMgmtTenant {
			query = query.Where("scope_type = ?", api.SCOPETYPE_SYSTEM)
		} else {
			query = query.Where("tenant_id = ?", strconv.Itoa(int(tenantId))).
				Where("scope_type = ?", api.SCOPETYPE_TENANT)
		}

		if _, err := query.
			UpdateColumns(map[string]interface{}{
				"updated_at": timeNow,
				"karta":      karta,
			}); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func validateExpectedKserveInferenceService(actual map[string]interface{}) {
	// first, validate actual has expected structure
	actualRi := &models.Karta{}
	bytes, err := json.Marshal(actual)
	Expect(err).To(BeNil(), "Failed to marshal actual ResourceInterface")
	err = json.Unmarshal(bytes, actualRi)
	Expect(err).To(BeNil(), "Failed to unmarshal actual ResourceInterface")

	expectedRiKserveInferenceService := &models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "inferenceservice",
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   "serving.kserve.io",
						Version: "v1beta1",
						Kind:    "InferenceService",
					},
					StatusDefinition: &kartav1alpha1.StatusDefinition{
						ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{
							Path:             ".status.conditions",
							TypeFieldName:    "type",
							StatusFieldName:  "status",
							MessageFieldName: ptr.To("message"),
						},
						StatusMappings: kartav1alpha1.StatusMappings{
							Running: []kartav1alpha1.StatusMatcher{
								{
									ByConditions: []kartav1alpha1.ExpectedCondition{
										{
											Type:   "PredictorReady",
											Status: ptr.To("True"),
										},
										{
											Type:   "RoutesReady",
											Status: ptr.To("True"),
										},
										{
											Type:   "LatestDeploymentReady",
											Status: ptr.To("True"),
										},
									},
								},
							},
							Failed: []kartav1alpha1.StatusMatcher{
								{
									ByConditions: []kartav1alpha1.ExpectedCondition{
										{
											Type:   "PredictorReady",
											Status: ptr.To("False"),
										},
										{
											Type:   "PredictorConfigurationReady",
											Status: ptr.To("False"),
										},
										{
											Type:   "RoutesReady",
											Status: ptr.To("False"),
										},
									},
								},
							},
						},
					},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{
						Name:     "predictor",
						Kind:     &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
								NodeAffinityPath:      ptr.To(".spec.predictor.affinity.nodeAffinity"),
								PodAffinityPath:       ptr.To(".spec.predictor.affinity.podAffinity"),
								PriorityClassNamePath: ptr.To(".spec.predictor.priorityClassName"),
								SchedulerNamePath:     ptr.To(".spec.predictor.schedulerName"),
								LabelsPath:            ptr.To(".spec.predictor.labels"),
								AnnotationsPath:       ptr.To(".spec.predictor.annotations"),
								ContainerPath:         ptr.To(".spec.predictor | ( (.[]?  | select(type ==\"object\" and .storageUri )))"),
								ContainersPath:        ptr.To(".spec.predictor.containers"),
							},
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.predictor.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.predictor.maxReplicas"),
						},
						PodSelector: &kartav1alpha1.PodSelector{
							ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
								KeyPath: ".metadata.labels[\"component\"]",
								Value:   ptr.To("predictor"),
							},
						},
					},
					{
						Name:     "transformer",
						Kind:     &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
						OwnerRef: ptr.To("inferenceservice"),
						SpecDefinition: &kartav1alpha1.SpecDefinition{
							PodSpecPath:  ptr.To(".spec.transformer"),
							MetadataPath: ptr.To(".spec.transformer"),
						},
						ScaleDefinition: &kartav1alpha1.ScaleDefinition{
							MinReplicasPath: ptr.To(".spec.transformer.minReplicas"),
							MaxReplicasPath: ptr.To(".spec.transformer.maxReplicas"),
						},
						PodSelector: &kartav1alpha1.PodSelector{
							ComponentTypeSelector: &kartav1alpha1.ComponentTypeSelector{
								KeyPath: ".metadata.labels[\"component\"]",
								Value:   ptr.To("transformer"),
							},
						},
					},
				},
			},
			Instructions: kartav1alpha1.OptimizationInstructions{
				GangScheduling: &kartav1alpha1.GangSchedulingInstruction{
					PodGroups: []kartav1alpha1.PodGroupDefinition{
						{
							Name: "service",
							Members: []kartav1alpha1.PodGroupMemberDefinition{
								{
									ComponentName:   "predictor",
									GroupByKeyPaths: []string{".metadata.labels[\"serving.kserve.io/inferenceservice\"]"},
								},
								{
									ComponentName:   "transformer",
									GroupByKeyPaths: []string{".metadata.labels[\"serving.kserve.io/inferenceservice\"]"},
								},
							},
						},
					},
				},
			},
		},
	}

	Expect(actualRi.Spec.StructureDefinition.RootComponent).To(Equal(expectedRiKserveInferenceService.Spec.StructureDefinition.RootComponent), "KServe InferenceService RootComponent mismatch")
	Expect(actualRi.Spec.StructureDefinition.ChildComponents).To(Equal(expectedRiKserveInferenceService.Spec.StructureDefinition.ChildComponents), "KServe InferenceService ChildComponents mismatch")
	Expect(actualRi.Spec.StructureDefinition.AdditionalChildKinds).To(Equal(expectedRiKserveInferenceService.Spec.StructureDefinition.AdditionalChildKinds), "KServe InferenceService AdditionalChildKinds mismatch")
	Expect(actualRi.Spec.Instructions).To(Equal(expectedRiKserveInferenceService.Spec.Instructions), "KServe InferenceService Instructions GangScheduling mismatch")

	bytes, err = json.Marshal(expectedRiKserveInferenceService)
	Expect(err).To(BeNil(), "Failed to marshal expected ResourceInterface")
	var expected map[string]interface{}
	err = json.Unmarshal(bytes, &expected)
	Expect(err).To(BeNil(), "Failed to unmarshal expected ResourceInterface")
	Expect(actual).To(Equal(expected), "KServe InferenceService ResourceInterface mismatch")
}

// convertKartaToMap converts a models.Karta struct to map[string]interface{}
func convertKartaToMap(karta models.Karta) map[string]interface{} {
	bytes, err := json.Marshal(karta)
	Expect(err).To(BeNil(), "Failed to marshal Karta")
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	Expect(err).To(BeNil(), "Failed to unmarshal Karta")
	return result
}
