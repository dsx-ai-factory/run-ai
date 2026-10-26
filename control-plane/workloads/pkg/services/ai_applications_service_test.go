/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"net/http"
	"time"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	workload_settings "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/settings"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/google/uuid"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI applications API tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	aiApplicationsService := NewAIApplicationsService()
	workloadsService := NewWorkloadsService()
	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	podsService := NewPodsService(mockTenantSettingsClient)

	BeforeEach(func() {
		setting_pkg.Init(nil, nil)
		Expect(setting_pkg.Register[*workload_settings.ApplicationSubmission]()).To(BeNil())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, "2.25.0", nil)).
			WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.25.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
		ctx = testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
		orgUnitClient.SetMock(orgUnitClient.NewSimpleMock().
			WithProject("proj-1", orgUnitApi.Project{
				Id:       "proj-1",
				Name:     "project-1",
				ParentId: "dept-1",
			}).
			WithProject("proj-del", orgUnitApi.Project{
				Id:       "proj-del",
				Name:     "project-del",
				ParentId: "dept-1",
			}).
			// Department whose name is resolvable from org-unit-service, used to
			// verify the sync backward-compatibility fallback for older clusters
			// that do not send departmentName in the payload.
			WithDepartment("dept-fallback", "department-fallback"),
		)
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	Context("sync applications", Ordered, func() {
		clusterId, _ := uuid.NewUUID()
		app1Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e40")
		app1Name := "app-1"
		app2Id, _ := uuid.NewUUID()
		app2Name := "app-2"
		app3Id, _ := uuid.NewUUID()
		app3Name := "app-3"
		var application1 api.AiApplicationForSync
		var application2 api.AiApplicationForSync
		var application3 api.AiApplicationForSync

		AfterAll(func() {
			testing.CleanTables()
		})

		It("test add/update/delete sync", func() {
			application1 = api.AiApplicationForSync{
				Id:             app1Id.String(),
				Name:           app1Name,
				Source:         "helm",
				Type:           "chart1",
				DepartmentId:   "10",
				DepartmentName: "department-10",
				ProjectId:      "1",
				ProjectName:    "proj-1",
				K8sPhase:       "deployed",
				Resources: []api.AiApplicationSyncResource{
					{
						Id:   "123",
						Name: "res-1",
						Kind: "pvc",
					},
				},
				Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Name).To(Equal(application1.Name))
			Expect(applications[0].ID.String()).To(Equal(application1.Id))
			Expect(applications[0].Source).To(Equal(application1.Source))
			Expect(applications[0].ClusterId).To(Equal(clusterId.String()))
			Expect(applications[0].TenantId).To(Equal(testing.TestTenant))
			Expect(applications[0].Type).To(Equal(application1.Type))
			Expect(applications[0].ProjectName).To(Equal(application1.ProjectName))
			Expect(applications[0].DepartmentId).To(Equal(application1.DepartmentId))
			Expect(applications[0].DepartmentName).To(Equal(application1.DepartmentName))
			Expect(applications[0].ProjectId).To(Equal(application1.ProjectId))
			Expect(applications[0].CreatedAt.IsZero()).To(BeFalse())
			Expect(applications[0].UpdatedAt).NotTo(BeNil())
			Expect(applications[0].K8sPhase).To(Equal(application1.K8sPhase))
			createdAt := applications[0].CreatedAt
			resources := applications[0].Resources
			Expect(len(resources)).To(Equal(1))
			Expect(resources[0].Id).To(Equal("123"))
			Expect(resources[0].Name).To(Equal("res-1"))
			Expect(resources[0].Kind).To(Equal("pvc"))

			application1.K8sPhase = "other"
			application1.Resources = []api.AiApplicationSyncResource{
				{
					Id:   "122",
					Name: "res-2",
					Kind: "pvc",
				},
			}
			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err = dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications = *applicationsRes
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Name).To(Equal(application1.Name))
			Expect(applications[0].ID.String()).To(Equal(application1.Id))
			Expect(applications[0].Source).To(Equal(application1.Source))
			Expect(applications[0].ClusterId).To(Equal(clusterId.String()))
			Expect(applications[0].TenantId).To(Equal(testing.TestTenant))
			Expect(applications[0].Type).To(Equal(application1.Type))
			Expect(applications[0].ProjectName).To(Equal(application1.ProjectName))
			Expect(applications[0].DepartmentId).To(Equal(application1.DepartmentId))
			Expect(applications[0].DepartmentName).To(Equal(application1.DepartmentName))
			Expect(applications[0].ProjectId).To(Equal(application1.ProjectId))
			// created_at must survive a re-sync update (regression: Save previously
			// overwrote it with the zero value on every update).
			Expect(applications[0].CreatedAt.IsZero()).To(BeFalse())
			Expect(applications[0].CreatedAt).To(BeTemporally("==", createdAt))
			Expect(applications[0].UpdatedAt).NotTo(BeNil())
			Expect(applications[0].CreatedAt).NotTo(Equal(applications[0].UpdatedAt))
			Expect(applications[0].K8sPhase).To(Equal("other"))
			resources = applications[0].Resources
			Expect(len(resources)).To(Equal(1))
			Expect(resources[0].Id).To(Equal("122"))
			Expect(resources[0].Name).To(Equal("res-2"))
			Expect(resources[0].Kind).To(Equal("pvc"))

			application1.Action = api.AIAPPLICATIONSYNCACTION_DELETE
			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err = dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			Expect(len(*applicationsRes)).To(Equal(0))
		})

		It("resolves departmentName from org-unit when the sync payload omits it (older clusters)", func() {
			fallbackAppId, _ := uuid.NewUUID()
			applicationWithoutDeptName := api.AiApplicationForSync{
				Id:           fallbackAppId.String(),
				Name:         "app-fallback",
				Source:       "helm",
				Type:         "chart1",
				DepartmentId: "dept-fallback",
				// DepartmentName intentionally omitted, simulating an older cluster.
				ProjectId:   "proj-fallback",
				ProjectName: "project-fallback",
				K8sPhase:    "deployed",
				Action:      api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{applicationWithoutDeptName},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].ProjectId).To(Equal("proj-fallback"))
			// The name was resolved from org-unit-service (project's parent), not the payload.
			Expect(applications[0].DepartmentName).To(Equal("department-fallback"))

			testing.CleanTables()
		})

		It("test full sync", func() {
			application1 = api.AiApplicationForSync{
				Id:           app1Id.String(),
				Name:         app1Name,
				Source:       "helm",
				Type:         "chart1",
				DepartmentId: "10",
				ProjectId:    "1",
				ProjectName:  "proj-1",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application2 = api.AiApplicationForSync{
				Id:           app2Id.String(),
				Name:         app2Name,
				Source:       "helm",
				Type:         "chart2",
				DepartmentId: "10",
				ProjectId:    "1",
				ProjectName:  "proj-1",
				K8sPhase:     "deployed",
				Resources: []api.AiApplicationSyncResource{
					{
						Id:   "123",
						Name: "res-1",
						Kind: "pvc",
					},
				},
				Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application3 = api.AiApplicationForSync{
				Id:           app3Id.String(),
				Name:         app3Name,
				Source:       "helm",
				Type:         "chart2",
				DepartmentId: "10",
				ProjectId:    "1",
				ProjectName:  "proj-1",
				K8sPhase:     "deployed",
				Resources: []api.AiApplicationSyncResource{
					{
						Id:   "123",
						Name: "res-1",
						Kind: "pvc",
					},
					{
						Id:   "1234",
						Name: "res-2",
						Kind: "secret",
					},
				},
				Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
			}

			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Name).To(Equal(application1.Name))
			Expect(applications[0].ID.String()).To(Equal(application1.Id))
			Expect(applications[0].Source).To(Equal(application1.Source))
			Expect(applications[0].ClusterId).To(Equal(clusterId.String()))
			Expect(applications[0].TenantId).To(Equal(testing.TestTenant))
			Expect(applications[0].Type).To(Equal(application1.Type))
			Expect(applications[0].ProjectName).To(Equal(application1.ProjectName))
			Expect(applications[0].ProjectId).To(Equal(application1.ProjectId))
			Expect(applications[0].DepartmentId).To(Equal(application1.DepartmentId))
			Expect(applications[0].CreatedAt).NotTo(BeNil())
			Expect(applications[0].UpdatedAt).NotTo(BeNil())
			Expect(applications[0].K8sPhase).To(Equal(application1.K8sPhase))
			resources := applications[0].Resources
			Expect(len(resources)).To(Equal(0))

			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application2},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err = dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications = *applicationsRes
			Expect(len(applications)).To(Equal(2))

			expectedByName := map[string]api.AiApplicationForSync{
				application1.Name: application1,
				application2.Name: application2,
			}
			seen := map[string]bool{}
			for _, application := range applications {
				expected, ok := expectedByName[application.Name]
				Expect(ok).To(BeTrue())
				Expect(application.Name).To(Equal(expected.Name))
				Expect(application.ID.String()).To(Equal(expected.Id))
				Expect(seen[application.Name]).To(BeFalse()) // no duplicates
				seen[application.Name] = true
			}
			Expect(seen).To(HaveLen(len(expectedByName))) // no missing expected entries

			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     true,
				Applications: []api.AiApplicationForSync{application1, application3},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err = dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications = *applicationsRes
			Expect(len(applications)).To(Equal(2))
			var actualApplication1 models.ApplicationView
			var actualApplication3 models.ApplicationView
			for _, application := range applications {
				if application.Name == application1.Name {
					actualApplication1 = application
				} else if application.Name == application3.Name {
					actualApplication3 = application
				}
			}

			Expect(actualApplication1.Name).To(Equal(application1.Name))
			Expect(actualApplication1.ID.String()).To(Equal(application1.Id))
			Expect(actualApplication1.Source).To(Equal(application1.Source))
			Expect(actualApplication1.ClusterId).To(Equal(clusterId.String()))
			Expect(actualApplication1.TenantId).To(Equal(testing.TestTenant))
			Expect(actualApplication1.Type).To(Equal(application1.Type))
			Expect(actualApplication1.ProjectName).To(Equal(application1.ProjectName))
			Expect(actualApplication1.ProjectId).To(Equal(application1.ProjectId))
			Expect(actualApplication1.DepartmentId).To(Equal(application1.DepartmentId))
			Expect(actualApplication1.CreatedAt).NotTo(BeNil())
			Expect(actualApplication1.UpdatedAt).NotTo(BeNil())
			Expect(actualApplication1.K8sPhase).To(Equal(application1.K8sPhase))
			resources = actualApplication1.Resources
			Expect(len(resources)).To(Equal(0))

			Expect(actualApplication3.Name).To(Equal(application3.Name))
			Expect(actualApplication3.ID.String()).To(Equal(application3.Id))
			Expect(actualApplication3.Source).To(Equal(application3.Source))
			Expect(actualApplication3.ClusterId).To(Equal(clusterId.String()))
			Expect(actualApplication3.TenantId).To(Equal(testing.TestTenant))
			Expect(actualApplication3.Type).To(Equal(application3.Type))
			Expect(actualApplication3.ProjectName).To(Equal(application3.ProjectName))
			Expect(actualApplication3.ProjectId).To(Equal(application3.ProjectId))
			Expect(actualApplication3.DepartmentId).To(Equal(application3.DepartmentId))
			Expect(actualApplication3.CreatedAt).NotTo(BeNil())
			Expect(actualApplication3.UpdatedAt).NotTo(BeNil())
			Expect(actualApplication3.K8sPhase).To(Equal(application3.K8sPhase))
			resources = actualApplication3.Resources
			Expect(len(resources)).To(Equal(2))
			Expect(resources[0].Id).To(Equal("123"))
			Expect(resources[0].Name).To(Equal("res-1"))
			Expect(resources[0].Kind).To(Equal("pvc"))
			Expect(resources[1].Id).To(Equal("1234"))
			Expect(resources[1].Name).To(Equal("res-2"))
			Expect(resources[1].Kind).To(Equal("secret"))
		})
	})

	Context("get applications", Ordered, func() {
		clusterId, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e00")
		cluster2Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e01")
		app1Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e40")
		app1Name := "app-1"
		app2Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e41")
		app2Name := "app-2"
		app3Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e42")
		app3Name := "app-3"
		var application1 api.AiApplicationForSync
		var application2 api.AiApplicationForSync
		var application3 api.AiApplicationForSync

		BeforeAll(func() {
			application1 = api.AiApplicationForSync{
				Id:             app1Id.String(),
				Name:           app1Name,
				Source:         "helm",
				Type:           "chart1",
				DepartmentId:   "10",
				DepartmentName: "department-aaa",
				ProjectId:      "1",
				ProjectName:    "proj-1",
				K8sPhase:       "deployed",
				Resources:      nil,
				Action:         api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application2 = api.AiApplicationForSync{
				Id:             app2Id.String(),
				Name:           app2Name,
				Source:         "helm",
				Type:           "chart2",
				DepartmentId:   "10",
				DepartmentName: "department-aaa",
				ProjectId:      "1",
				ProjectName:    "proj-1",
				K8sPhase:       "deployed",
				Resources: []api.AiApplicationSyncResource{
					{
						Id:          "123",
						Name:        "res-1",
						ClusterId:   "123123123",
						ClusterName: "res-1-cluster",
						Kind:        "pvc",
					},
				},
				Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application3 = api.AiApplicationForSync{
				Id:             app3Id.String(),
				Name:           app3Name,
				Source:         "helm",
				Type:           "chart2",
				DepartmentId:   "101",
				DepartmentName: "department-bbb",
				ProjectId:      "11",
				ProjectName:    "proj-1",
				K8sPhase:       "deployed",
				Resources: []api.AiApplicationSyncResource{
					{
						Id:   "123",
						Name: "res-1",
						Kind: "pvc",
					},
					{
						Id:          "1234",
						Name:        "res-2",
						ClusterId:   "111234",
						ClusterName: "res-2-cluster",
						Kind:        "workload",
					},
				},
				Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
			}

			// Sync app1 and app2 separately with a small sleep between them so their
			// created_at timestamps are guaranteed to differ. This keeps the
			// "sort by createdAt" test below stable (otherwise an equal-timestamp
			// tie would fall back to the secondary "id DESC" sort and reverse the
			// expected app1/app2 order).
			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			time.Sleep(10 * time.Millisecond)
			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application2},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			// Wait 1 second to ensure different creation times
			time.Sleep(1 * time.Second)
			syncResponse, err = aiApplicationsService.SyncApplications(ctx, cluster2Id.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application3},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(3))
		})

		AfterAll(func() {
			testing.CleanTables()
		})

		It("test no filter/sort", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(3))
			Expect(applications[0].Id).To(Equal(application3.Id))
			Expect(applications[1].Id).To(Equal(application2.Id))
			Expect(applications[2].Id).To(Equal(application1.Id))
			application := applications[1]
			Expect(application.Name).To(Equal(application2.Name))
			Expect(application.Id).To(Equal(application2.Id))
			Expect(application.Source).To(Equal(application2.Source))
			Expect(application.Type).To(Equal(application2.Type))
			Expect(application.ClusterId).To(Equal(clusterId.String()))
			Expect(application.ProjectName).To(Equal(application2.ProjectName))
			Expect(application.ProjectId).To(Equal(application2.ProjectId))
			Expect(application.DepartmentId).To(Equal(application2.DepartmentId))
			Expect(application.DepartmentName).To(Equal(application2.DepartmentName))
			Expect(application.Phase).To(Equal(application2.K8sPhase))
			resources := application.Resources
			Expect(len(resources)).To(Equal(1))
			Expect(resources[0].Id).To(Equal("123"))
			Expect(resources[0].Name).To(Equal("res-1"))
			Expect(resources[0].Kind).To(Equal("pvc"))
		})

		It("test cluster filter", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", clusterId)}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(2))
			Expect(applications[0].Id).To(Equal(application2.Id))
			Expect(applications[1].Id).To(Equal(application1.Id))
		})

		It("test project filter", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{fmt.Sprintf("projectId==%s", "1")}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(2))
			Expect(applications[0].Id).To(Equal(application2.Id))
			Expect(applications[1].Id).To(Equal(application1.Id))
		})

		It("test department name filter", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{"departmentName==department-bbb"}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Id).To(Equal(application3.Id))
			Expect(applications[0].DepartmentName).To(Equal("department-bbb"))
		})

		It("test department id filter", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{"departmentId==101"}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Id).To(Equal(application3.Id))
			Expect(applications[0].DepartmentId).To(Equal("101"))
		})

		It("test department sort", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "departmentName", []string{}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(3))
			// Ascending by departmentName: "department-aaa" (app1, app2) before "department-bbb" (app3).
			// Within the equal "department-aaa" group the secondary sort is id DESC (app2 before app1).
			Expect(applications[0].Id).To(Equal(application2.Id))
			Expect(applications[1].Id).To(Equal(application1.Id))
			Expect(applications[2].Id).To(Equal(application3.Id))
		})

		It("test cluster sort", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "clusterId", []string{}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(3))
			Expect(applications[0].Id).To(Equal(application2.Id))
			Expect(applications[1].Id).To(Equal(application1.Id))
			Expect(applications[2].Id).To(Equal(application3.Id))
		})

		It("test sort by createdAt", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "createdAt", []string{}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(3))
			// Verify they are sorted by createdAt in ascending order
			Expect(applications[0].Id).To(Equal(application1.Id))
			Expect(applications[1].Id).To(Equal(application2.Id))
			Expect(applications[2].Id).To(Equal(application3.Id))
			// Verify the createdAt timestamps are in order
			Expect(applications[0].CreatedAt.Before(applications[1].CreatedAt) || applications[0].CreatedAt.Equal(applications[1].CreatedAt)).To(BeTrue())
			Expect(applications[1].CreatedAt.Before(applications[2].CreatedAt) || applications[1].CreatedAt.Equal(applications[2].CreatedAt)).To(BeTrue())
		})

		It("test filter by createdAt with >= operator", func() {
			// Get the third application (created after the 1-second sleep)
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{fmt.Sprintf("name==%s", app3Name)}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			apps := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(apps)).To(Equal(1))
			app3CreatedAt := apps[0].CreatedAt

			// Now filter for applications created at or after app3's timestamp
			// This should return only app3 since it was created 1 second after app1 and app2
			applicationsResponse, err = aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{fmt.Sprintf("createdAt>=%s", app3CreatedAt.Format(time.RFC3339))}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			// Should return only app3
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Id).To(Equal(application3.Id))
			// Verify all returned applications have createdAt >= the timestamp
			for _, app := range applications {
				Expect(app.CreatedAt.Equal(app3CreatedAt) || app.CreatedAt.After(app3CreatedAt)).To(BeTrue())
			}
		})

		It("list with no permitted scopes", func() {
			ctx2 := testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx2, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(0))
		})

		It("get with no permitted scopes", func() {
			ctx2 := testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})
			applicationsResponse, err := aiApplicationsService.GetApplication(ctx2, app1Id.String())
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusNotFound))
		})

		Describe("test verbosity", func() {
			createdWorkloadsIds := []string{}

			BeforeAll(func() {
				baseWorkload1 := SetupBaseWorkload(testing.TestUuid1, DefaultProjectId, cluster2Id.String())
				baseWorkload1.ApplicationId = ptr.To(app3Id.String())
				baseWorkload1.ApplicationName = ptr.To(app3Name)
				baseWorkload1.Endpoints = []api.EndpointSync{
					{
						Host:     "chat.example.com",
						Port:     8080,
						Path:     "/v1/completions",
						Protocol: "HTTPS",
						Url:      "https://chat.example.com:8080/v1/completions",
						SourceRef: api.EndpointSyncAllOfSourceRef{
							Kind: "Ingress",
							Name: "my-ingress",
						},
					},
					{
						Host:       "api.example.com",
						Port:       443,
						Path:       "/health",
						TlsEnabled: ptr.To(true),
						Protocol:   "HTTPS",
						Url:        "https://api.example.com/health",
					},
				}
				baseWorkload1.ResourceRefs = []api.ResourceRef{
					{Uid: "uid-pvc", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
					{Uid: "uid-sec", Name: "regcred", Kind: api.ResourceRefKind(runv1.StorageRefKindSecret)},
				}
				apiResponse, _ := workloadsService.UpsertWorkload(ctx, testing.TestUuid1, baseWorkload1)
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
				createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid1)

				id1 := uuid.NewString()
				basePod := SetupBasePod(id1, testing.TestUuid1, DefaultProjectId, cluster2Id.String())
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

				baseWorkload2 := SetupBaseWorkload(testing.TestUuid2, DefaultProjectId, cluster2Id.String())
				baseWorkload2.ApplicationId = ptr.To(app3Id.String())
				baseWorkload2.ApplicationName = ptr.To(app3Name)
				baseWorkload2.ResourceRefs = []api.ResourceRef{
					{Uid: "uid-pvc", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
				}
				apiResponse, _ = workloadsService.UpsertWorkload(ctx, testing.TestUuid2, baseWorkload2)
				Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
				createdWorkloadsIds = append(createdWorkloadsIds, testing.TestUuid2)

				id3 := uuid.NewString()
				basePod3 := SetupBasePod(id3, testing.TestUuid2, DefaultProjectId, cluster2Id.String())
				basePod3.RequestedResources.Cpu = &api.RequestResourceCores{
					Limit:   pointer.Float64(16),
					Request: pointer.Float64(8),
				}
				basePod3.RequestedResources.CpuMemory = &api.RequestResourceQuantity{
					Limit:   pointer.String("100Mi"),
					Request: pointer.String("1G"),
				}
				basePod3.AllocatedResources = &api.AllocatedResources{
					Gpu:       basePod.RequestedResources.Gpu.Request,
					Cpu:       basePod.RequestedResources.Cpu.Request,
					CpuMemory: basePod.RequestedResources.CpuMemory.Request,
				}
				basePod3.CompletedAt = convutils.Ref(time.Now())
				id4 := uuid.NewString()
				basePod4 := basePod
				basePod4.Id = id4
				basePod4.CurrentNodePool = pointer.String(OtherNodepool)
				basePod4.CompletedAt = convutils.Ref(basePod.CompletedAt.Add(1 * time.Hour))

				apiResponse, _ = podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
					ProjectId: DefaultProjectId,
					Pods: []api.PodSyncFields{
						basePod3,
						basePod4,
					},
				})
			})

			AfterAll(func() {
			})

			It("list applications brief", func() {
				applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_BRIEF, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", cluster2Id)}, "")
				Expect(err).To(BeNil())
				Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
				applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
				Expect(len(applications)).To(Equal(1))
				Expect(applications[0].Id).To(Equal(application3.Id))
				validateApplicationWithVerbosity(applications[0], application3, false)
			})

			It("list applications full", func() {
				applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{fmt.Sprintf("clusterId==%s", cluster2Id)}, "")
				Expect(err).To(BeNil())
				Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
				applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
				Expect(len(applications)).To(Equal(1))
				Expect(applications[0].Id).To(Equal(application3.Id))
				validateApplicationWithVerbosity(applications[0], application3, true)
			})

			It("get application", func() {
				applicationsResponse, err := aiApplicationsService.GetApplication(ctx, app3Id.String())
				Expect(err).To(BeNil())
				Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
				application := applicationsResponse.Body.(api.AiApplication)
				validateApplicationWithVerbosity(application, application3, true)
			})

			It("get application - validate all fields", func() {
				applicationsResponse, err := aiApplicationsService.GetApplication(ctx, app3Id.String())
				Expect(err).To(BeNil())
				Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
				application := applicationsResponse.Body.(api.AiApplication)

				// Validate basic application fields
				Expect(application.Id).To(Equal(app3Id.String()))
				Expect(application.Name).To(Equal(app3Name))
				Expect(application.Source).To(Equal("helm"))
				Expect(application.Type).To(Equal("chart2"))
				Expect(application.Phase).To(Equal("deployed"))
				Expect(application.ProjectId).To(Equal("11"))
				Expect(application.ProjectName).To(Equal("proj-1"))

				// Validate CreatedAt is set
				Expect(application.CreatedAt).NotTo(BeZero())
				Expect(application.CreatedAt).To(BeTemporally("<=", time.Now()))

				// Validate resources
				Expect(application.Resources).NotTo(BeNil())
				Expect(len(application.Resources)).To(Equal(2))
			})

			It("get application - validate per-workload resourceConnections", func() {
				applicationsResponse, err := aiApplicationsService.GetApplication(ctx, app3Id.String())
				Expect(err).To(BeNil())
				Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
				application := applicationsResponse.Body.(api.AiApplication)

				Expect(application.ResourceConnections).To(ConsistOf(
					api.AiApplicationResourceConnection{WorkloadId: testing.TestUuid1, Uid: "uid-pvc", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
					api.AiApplicationResourceConnection{WorkloadId: testing.TestUuid1, Uid: "uid-sec", Name: "regcred", Kind: api.ResourceRefKind(runv1.StorageRefKindSecret)},
					api.AiApplicationResourceConnection{WorkloadId: testing.TestUuid2, Uid: "uid-pvc", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
				))
			})
		})
	})

	Context("search applications", Ordered, func() {
		clusterId, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e10")
		app1Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e50")
		app2Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e51")
		app3Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e52")
		app4Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e53")
		app5Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e54")

		BeforeAll(func() {
			// Create 5 applications with 2 different projects and 2 different departments
			application1 := api.AiApplicationForSync{
				Id:           app1Id.String(),
				Name:         "ml-model-training",
				Source:       "helm",
				Type:         "tensorflow-app",
				DepartmentId: "dept001",
				ProjectId:    "project0001",
				ProjectName:  "project0001-ml",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application2 := api.AiApplicationForSync{
				Id:           app2Id.String(),
				Name:         "data-pipeline",
				Source:       "helm",
				Type:         "spark-app",
				DepartmentId: "dept001",
				ProjectId:    "project0001",
				ProjectName:  "project0001-data",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application3 := api.AiApplicationForSync{
				Id:           app3Id.String(),
				Name:         "web-service",
				Source:       "helm",
				Type:         "nginx-app",
				DepartmentId: "dept002",
				ProjectId:    "project0002",
				ProjectName:  "project0002-web",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application4 := api.AiApplicationForSync{
				Id:           app4Id.String(),
				Name:         "api-gateway",
				Source:       "helm",
				Type:         "nginx-app",
				DepartmentId: "dept002",
				ProjectId:    "project0002",
				ProjectName:  "project0002-api",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application5 := api.AiApplicationForSync{
				Id:           app5Id.String(),
				Name:         "tensorflow-inference",
				Source:       "helm",
				Type:         "tensorflow-app",
				DepartmentId: "dept001",
				ProjectId:    "project0001",
				ProjectName:  "project0001-inference",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}

			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1, application2, application3, application4, application5},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			// Verify all 5 applications were created
			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(5))
		})

		AfterAll(func() {
			testing.CleanTables()
		})

		It("test search by project name - should return only project0001 applications", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "project0001")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(3))

			// Verify all returned applications belong to project0001
			for _, app := range applications {
				Expect(app.ProjectId).To(Equal("project0001"))
			}

			// Verify the specific applications returned
			appIds := make(map[string]bool)
			for _, app := range applications {
				appIds[app.Id] = true
			}
			Expect(appIds[app1Id.String()]).To(BeTrue())
			Expect(appIds[app2Id.String()]).To(BeTrue())
			Expect(appIds[app5Id.String()]).To(BeTrue())
		})

		It("test search by application name - should return only matching names", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "tensorflow")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(2))

			// Verify all returned applications have 'tensorflow' in their name or type
			appIds := make(map[string]bool)
			for _, app := range applications {
				appIds[app.Id] = true
			}
			Expect(appIds[app1Id.String()]).To(BeTrue()) // ml-model-training with type tensorflow-app
			Expect(appIds[app5Id.String()]).To(BeTrue()) // tensorflow-inference
		})

		It("test search by application type - should return only nginx-app types", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "nginx")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(2))

			// Verify all returned applications are nginx-app type
			for _, app := range applications {
				Expect(app.Type).To(Equal("nginx-app"))
			}

			// Verify the specific applications returned
			appIds := make(map[string]bool)
			for _, app := range applications {
				appIds[app.Id] = true
			}
			Expect(appIds[app3Id.String()]).To(BeTrue())
			Expect(appIds[app4Id.String()]).To(BeTrue())
		})

		It("test search by name pattern - should return applications matching 'service'", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "service")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(1))
			Expect(applications[0].Id).To(Equal(app3Id.String()))
			Expect(applications[0].Name).To(Equal("web-service"))
		})

		It("test search with no matches - should return empty list", func() {
			applicationsResponse, err := aiApplicationsService.ListApplications(ctx, api.AIAPPLICATIONVERBOSITY_FULL, 0, 0, "", "", []string{}, "nonexistent")
			Expect(err).To(BeNil())
			Expect(applicationsResponse.Code).To(Equal(http.StatusOK))
			applications := applicationsResponse.Body.(api.ListApplications200Response).Applications
			Expect(len(applications)).To(Equal(0))
		})
	})

	Context("count applications", Ordered, func() {
		clusterId, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e20")
		cluster2Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e21")
		app1Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e60")
		app2Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e61")
		app3Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e62")
		app4Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e63")
		app5Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e64")
		app6Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e65")

		BeforeAll(func() {
			// Create 6 applications with different attributes for comprehensive testing
			application1 := api.AiApplicationForSync{
				Id:           app1Id.String(),
				Name:         "ml-training-app",
				Source:       "helm",
				Type:         "tensorflow",
				DepartmentId: "dept100",
				ProjectId:    "proj100",
				ProjectName:  "ml-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application2 := api.AiApplicationForSync{
				Id:           app2Id.String(),
				Name:         "data-processing",
				Source:       "helm",
				Type:         "spark",
				DepartmentId: "dept100",
				ProjectId:    "proj100",
				ProjectName:  "data-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application3 := api.AiApplicationForSync{
				Id:           app3Id.String(),
				Name:         "web-frontend",
				Source:       "helm",
				Type:         "nginx",
				DepartmentId: "dept200",
				ProjectId:    "proj200",
				ProjectName:  "web-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application4 := api.AiApplicationForSync{
				Id:           app4Id.String(),
				Name:         "api-backend",
				Source:       "helm",
				Type:         "nodejs",
				DepartmentId: "dept200",
				ProjectId:    "proj200",
				ProjectName:  "api-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			application5 := api.AiApplicationForSync{
				Id:           app5Id.String(),
				Name:         "tensorflow-serving",
				Source:       "helm",
				Type:         "tensorflow",
				DepartmentId: "dept100",
				ProjectId:    "proj100",
				ProjectName:  "inference-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}
			// Application 6 in a different cluster
			application6 := api.AiApplicationForSync{
				Id:           app6Id.String(),
				Name:         "monitoring-app",
				Source:       "helm",
				Type:         "prometheus",
				DepartmentId: "dept300",
				ProjectId:    "proj300",
				ProjectName:  "monitoring-project",
				K8sPhase:     "deployed",
				Resources:    nil,
				Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
			}

			// Sync applications to cluster1
			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application1, application2, application3, application4, application5},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			// Sync application6 to cluster2
			syncResponse, err = aiApplicationsService.SyncApplications(ctx, cluster2Id.String(), api.AiApplicationsForSync{
				FullSync:     false,
				Applications: []api.AiApplicationForSync{application6},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			// Verify all 6 applications were created
			applicationsRes, err := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(err).To(BeNil())
			applications := *applicationsRes
			Expect(len(applications)).To(Equal(6))
		})

		AfterAll(func() {
			testing.CleanTables()
		})

		It("count all applications - no filters", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(6)))
		})

		It("count applications with cluster filter", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(5)))

			countResponse, err = aiApplicationsService.CountApplications(ctx, []string{fmt.Sprintf("clusterId==%s", cluster2Id)}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count = countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(1)))
		})

		It("count applications with project filter", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"projectId==proj100"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(3)))

			countResponse, err = aiApplicationsService.CountApplications(ctx, []string{"projectId==proj200"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count = countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(2)))
		})

		It("count applications with type filter", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"type==tensorflow"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(2)))
		})

		It("count applications with name filter using contains", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"name=@app"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(2))) // ml-training-app, monitoring-app
		})

		It("count applications with name filter using starts with", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"name=^ml"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(1))) // ml-training-app
		})

		It("count applications with multiple filters", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{
				fmt.Sprintf("clusterId==%s", clusterId),
				"projectId==proj100",
			}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(3)))
		})

		It("count applications with search - by project name", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "ml-project")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(1)))
		})

		It("count applications with search - by type", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "tensorflow")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(2)))
		})

		It("count applications with search - by application name", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "frontend")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(1)))
		})

		It("count applications with search and filter combined", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"projectId==proj100"}, "tensorflow")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(2)))
		})

		It("count applications with search - no matches", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "nonexistent-search-term")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(0)))
		})

		It("count applications with filter - no matches", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"projectId==nonexistent-project"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountResponse).Count
			Expect(count).To(Equal(int64(0)))
		})

		It("count with invalid filter - should return bad request", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{"invalidField==value"}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("count should exclude soft-deleted applications", func() {
			countResponse, err := aiApplicationsService.CountApplications(ctx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			countBefore := countResponse.Body.(api.CountResponse).Count

			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:     app1Id.String(),
						Action: api.AIAPPLICATIONSYNCACTION_DELETE,
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			countResponse, err = aiApplicationsService.CountApplications(ctx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			countAfter := countResponse.Body.(api.CountResponse).Count
			Expect(countAfter).To(Equal(countBefore - 1))

			countResponse, err = aiApplicationsService.CountApplications(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			countFiltered := countResponse.Body.(api.CountResponse).Count
			Expect(countFiltered).To(Equal(countBefore - 2)) // -1 for cluster2's app, -1 for deleted app

			syncResponse, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:           app1Id.String(),
						Name:         "ml-training-app",
						Source:       "helm",
						Type:         "tensorflow",
						DepartmentId: "dept100",
						ProjectId:    "proj100",
						ProjectName:  "ml-project",
						K8sPhase:     "deployed",
						Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			countResponse, err = aiApplicationsService.CountApplications(ctx, []string{}, "")
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			countRestored := countResponse.Body.(api.CountResponse).Count
			Expect(countRestored).To(Equal(countBefore))
		})

		Describe("count applications with permitted scopes", func() {
			var ctx2 context.Context
			var ctx3 context.Context
			clusterId3, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e22")
			app7Id, _ := uuid.Parse("eea87015-19c3-4d03-b2b3-b632c5590e66")

			BeforeAll(func() {
				// Create context with different tenant
				ctx2 = testing.SetContext(testing.TestTenant2, testing.TestUser, []string{})

				// Create context with permitted scopes limited to specific projects
				ctx3 = testing.SetProjectOnlyContext(testing.TestTenant, testing.TestUser, []string{}, "proj100")

				// Create an application for tenant2
				application7 := api.AiApplicationForSync{
					Id:           app7Id.String(),
					Name:         "tenant2-app",
					Source:       "helm",
					Type:         "other",
					DepartmentId: "dept400",
					ProjectId:    "proj400",
					ProjectName:  "tenant2-project",
					K8sPhase:     "deployed",
					Resources:    nil,
					Action:       api.AIAPPLICATIONSYNCACTION_UPSERT,
				}
				syncResponse, err := aiApplicationsService.SyncApplications(ctx2, clusterId3.String(), api.AiApplicationsForSync{
					FullSync:     false,
					Applications: []api.AiApplicationForSync{application7},
				})
				Expect(err).To(BeNil())
				Expect(syncResponse.Code).To(Equal(http.StatusNoContent))
			})

			It("count applications with different tenant - should return 0", func() {
				countResponse, err := aiApplicationsService.CountApplications(ctx2, []string{}, "")
				Expect(err).To(BeNil())
				Expect(countResponse.Code).To(Equal(http.StatusOK))
				count := countResponse.Body.(api.CountResponse).Count
				Expect(count).To(Equal(int64(1))) // Only tenant2's application
			})

			It("count applications with limited permitted scopes - should only count allowed projects", func() {
				countResponse, err := aiApplicationsService.CountApplications(ctx3, []string{}, "")
				Expect(err).To(BeNil())
				Expect(countResponse.Code).To(Equal(http.StatusOK))
				count := countResponse.Body.(api.CountResponse).Count
				Expect(count).To(Equal(int64(3))) // Only applications in proj100
			})

			It("count applications with limited permitted scopes and filter", func() {
				countResponse, err := aiApplicationsService.CountApplications(ctx3, []string{"type==tensorflow"}, "")
				Expect(err).To(BeNil())
				Expect(countResponse.Code).To(Equal(http.StatusOK))
				count := countResponse.Body.(api.CountResponse).Count
				Expect(count).To(Equal(int64(2))) // tensorflow apps in proj100
			})

			It("count applications with limited permitted scopes and search", func() {
				countResponse, err := aiApplicationsService.CountApplications(ctx3, []string{}, "training")
				Expect(err).To(BeNil())
				Expect(countResponse.Code).To(Equal(http.StatusOK))
				count := countResponse.Body.(api.CountResponse).Count
				Expect(count).To(Equal(int64(1))) // ml-training-app in proj100
			})

			It("count applications with limited scopes filtering on non-permitted project - should return 0", func() {
				countResponse, err := aiApplicationsService.CountApplications(ctx3, []string{"projectId==proj200"}, "")
				Expect(err).To(BeNil())
				Expect(countResponse.Code).To(Equal(http.StatusOK))
				count := countResponse.Body.(api.CountResponse).Count
				Expect(count).To(Equal(int64(0))) // proj200 not in permitted scopes
			})
		})
	})

	Context("validate application spec", func() {
		It("should accept chartURL only", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					ChartURL: "https://example.com/chart.tgz",
				},
			}
			Expect(validateApplicationSpec(spec)).To(BeNil())
		})

		It("should accept repo triplet only", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "https://charts.example.com",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
			}
			Expect(validateApplicationSpec(spec)).To(BeNil())
		})

		It("should reject both chartURL and repo fields", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					ChartURL:     "https://example.com/chart.tgz",
					Repo:         "https://charts.example.com",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should reject when nothing is provided", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should reject partial repo triplet - missing chartVersion", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:      "https://charts.example.com",
					ChartName: "my-chart",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should reject partial repo triplet - missing chartName", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "https://charts.example.com",
					ChartVersion: "1.0.0",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should reject partial repo triplet - missing repo", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should reject repo with invalid scheme", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "ftp://charts.example.com",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should accept repo with oci scheme", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "oci://registry.example.com/charts",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
			}
			Expect(validateApplicationSpec(spec)).To(BeNil())
		})

		It("should accept valid auth asset ID", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "https://charts.example.com",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
				Auth: api.AiApplicationAuth{
					AssetId: uuid.NewString(),
				},
			}
			Expect(validateApplicationSpec(spec)).To(BeNil())
		})

		It("should reject invalid auth asset ID", func() {
			spec := &api.AiApplicationSpec{
				ChartRef: api.AiApplicationSpecChartRef{
					Repo:         "https://charts.example.com",
					ChartName:    "my-chart",
					ChartVersion: "1.0.0",
				},
				Auth: api.AiApplicationAuth{
					AssetId: "not-a-uuid",
				},
			}
			apiErr := validateApplicationSpec(spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})
	})

	Context("create application", Ordered, func() {
		clusterId, _ := uuid.Parse(testing.ClusterUuid1)

		AfterAll(func() {
			testing.CleanTables()
		})

		It("should reject creation when cluster version is below 2.25", func() {
			oldClusterId := testing.ClusterUuid3
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(oldClusterId, testing.BuildDisplayedCluster(oldClusterId, "2.24.0", nil)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			request := api.AiApplicationCreationRequest{
				Name:      "app-old-cluster",
				ClusterId: oldClusterId,
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/chart.tgz",
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			apiErr, ok := createResponse.Body.(api.Error)
			Expect(ok).To(BeTrue())
			Expect(apiErr.Message).To(ContainSubstring("incompatible with the version of cluster"))
		})

		It("should create an application with chartURL", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-url",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/chart.tgz",
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			application := createResponse.Body.(api.AiApplication)
			Expect(application.Name).To(Equal("app-url"))
			Expect(application.Source).To(Equal("runai"))
			Expect(application.Type).To(Equal("application"))
			Expect(application.ActionStatus).ToNot(BeNil())
			Expect(*application.ActionStatus).To(Equal("creating"))
			Expect(application.ProjectId).To(Equal("proj-1"))
			Expect(application.Spec.ChartRef.ChartURL).To(Equal("https://example.com/chart.tgz"))
			Expect(application.CreatedBy).To(Equal(testing.TestUser))
			Expect(application.Id).NotTo(BeEmpty())
		})

		It("should create an application with repo triplet", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-repo",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						Repo:         "https://charts.example.com",
						ChartName:    "redis",
						ChartVersion: "18.0.0",
					},
					Values: map[string]interface{}{
						"replicaCount": float64(2),
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			application := createResponse.Body.(api.AiApplication)
			Expect(application.Name).To(Equal("app-repo"))
			Expect(application.Spec.ChartRef.Repo).To(Equal("https://charts.example.com"))
			Expect(application.Spec.ChartRef.ChartName).To(Equal("redis"))
			Expect(application.Spec.ChartRef.ChartVersion).To(Equal("18.0.0"))
			Expect(application.Spec.Values).To(HaveKeyWithValue("replicaCount", float64(2)))
		})

		It("should create an application with auth asset ID", func() {
			assetId := uuid.NewString()
			request := api.AiApplicationCreationRequest{
				Name:      "app-private-repo",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						Repo:         "https://private.charts.example.com",
						ChartName:    "private-chart",
						ChartVersion: "1.0.0",
					},
					Auth: api.AiApplicationAuth{
						AssetId: assetId,
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			application := createResponse.Body.(api.AiApplication)
			Expect(application.Spec.Auth.AssetId).To(Equal(assetId))
		})

		It("should reject creation with invalid spec - both chartURL and repo", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-invalid",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL:     "https://example.com/chart.tgz",
						Repo:         "https://charts.example.com",
						ChartName:    "redis",
						ChartVersion: "18.0.0",
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject creation with empty spec", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-empty",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject creation with partial repo triplet", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-partial",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						Repo:      "https://charts.example.com",
						ChartName: "redis",
					},
				},
			}

			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
		})

		It("created application should be visible via list", func() {
			applicationsRes, dbErr := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(dbErr).To(BeNil())
			applications := *applicationsRes

			found := false
			for _, app := range applications {
				if app.Name == "app-url" {
					found = true
					Expect(app.TenantId).To(Equal(testing.TestTenant))
					Expect(app.ClusterId).To(Equal(clusterId.String()))
					Expect(app.Source).To(Equal("runai"))
					Expect(app.ActualPhase).To(Equal("creating"))
					Expect(app.DesiredPhase).NotTo(BeNil())
					Expect(*app.DesiredPhase).To(Equal("deployed"))
					Expect(app.Spec).NotTo(BeNil())
					Expect(app.Spec.ChartRef.ChartURL).To(Equal("https://example.com/chart.tgz"))
					break
				}
			}
			Expect(found).To(BeTrue())
		})
	})

	Context("delete application", Ordered, func() {
		clusterId, _ := uuid.Parse(testing.ClusterUuid1)
		var createdAppId string

		AfterEach(func() {
			testing.CleanTables()
		})

		It("should immediately soft-delete an unsynced application (actual_phase creating)", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-to-delete-unsynced",
				ClusterId: clusterId.String(),
				ProjectId: "proj-del",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/chart.tgz",
					},
				},
			}
			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))
			createdAppId = createResponse.Body.(api.AiApplication).Id

			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, createdAppId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))

			appUuid, _ := uuid.Parse(createdAppId)

			// Application should be gone from normal queries
			_, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).NotTo(BeNil())

			// But should exist as soft-deleted in the DB
			var softDeleted models.Application
			Expect(runaidb.DB().Unscoped().Where("id = ?", appUuid).First(&softDeleted)).To(BeNil())
			Expect(softDeleted.DeletedAt.Valid).To(BeTrue())

			// application_crud_sync row should still exist with desired_phase = deleted
			// so the agent can pick up the delete instruction on its next poll
			var crudSync models.ApplicationCrudSync
			Expect(runaidb.DB().Where("application_id = ?", appUuid).First(&crudSync)).To(BeNil())
			Expect(*crudSync.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
		})

		It("should mark synced application for async deletion (desired_phase deleted)", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "app-to-delete-synced",
				ClusterId: clusterId.String(),
				ProjectId: "proj-del",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/chart.tgz",
					},
				},
			}
			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))
			createdAppId = createResponse.Body.(api.AiApplication).Id

			// Simulate agent deploying the app on the cluster
			_, err = aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:          createdAppId,
						Name:        "app-to-delete-synced",
						Source:      "runai",
						Type:        "application",
						ProjectId:   "proj-del",
						ProjectName: "project-del",
						K8sPhase:    "deployed",
						Action:      api.AIAPPLICATIONSYNCACTION_UPSERT,
						ActionInfo: &api.AiApplicationSyncActionInfo{
							Status:    "completed",
							HelmPhase: "deployed",
						},
					},
				},
			})
			Expect(err).To(BeNil())

			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeployed))

			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, createdAppId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))

			app, dbErr = dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.DesiredPhase).NotTo(BeNil())
			Expect(*app.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
			Expect(app.DeletedAt.Valid).To(BeFalse())
		})

		It("should return not found for non-existent application", func() {
			fakeId := uuid.New().String()
			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, fakeId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("should return bad request for invalid UUID", func() {
			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, "not-a-uuid")
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("request applications updates (sync report)", Ordered, func() {
		clusterId, _ := uuid.Parse(testing.ClusterUuid1)
		cluster2Id, _ := uuid.Parse(testing.ClusterUuid2)

		AfterEach(func() {
			testing.CleanTables()
		})

		createApp := func(name string, cluster uuid.UUID, projectId string) string {
			request := api.AiApplicationCreationRequest{
				Name:      name,
				ClusterId: cluster.String(),
				ProjectId: projectId,
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/chart.tgz",
					},
				},
			}
			resp, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			return resp.Body.(api.AiApplication).Id
		}

		parseApplicationSyncResponse := func(httpResponse api.ImplResponse) api.AiApplicationSyncResponse {
			bytes, err := json.Marshal(httpResponse.Body)
			Expect(err).To(BeNil())
			response := &api.AiApplicationSyncResponse{}
			err = json.Unmarshal(bytes, response)
			Expect(err).To(BeNil())
			return *response
		}

		It("should return created applications as sync updates", func() {
			id1 := createApp("sync-app-1", clusterId, "proj-1")
			id2 := createApp("sync-app-2", clusterId, "proj-1")

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			resp := parseApplicationSyncResponse(syncResponse)
			Expect(len(resp.Applications)).To(Equal(2))
			Expect(resp.RequestInfo.AfterUpdate).To(Equal(int64(0)))
			Expect(resp.ResponseInfo.LowestUpdate).To(BeNumerically(">", 0))
			Expect(resp.ResponseInfo.HighestUpdate).To(BeNumerically(">=", resp.ResponseInfo.LowestUpdate))

			appIds := map[string]bool{}
			for _, syncInfo := range resp.Applications {
				appIds[syncInfo.Application.Id] = true
				Expect(syncInfo.Application.Source).To(Equal("runai"))
				Expect(syncInfo.Application.Type).To(Equal("application"))
				Expect(syncInfo.Application.Phase).To(Equal("creating"))
				Expect(syncInfo.Application.Spec.ChartRef.ChartURL).To(Equal("https://example.com/chart.tgz"))
			}
			Expect(appIds[id1]).To(BeTrue())
			Expect(appIds[id2]).To(BeTrue())
		})

		It("should return auth asset ID in sync updates", func() {
			assetId := uuid.NewString()
			request := api.AiApplicationCreationRequest{
				Name:      "sync-private-app",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						Repo:         "https://private.charts.example.com",
						ChartName:    "private-chart",
						ChartVersion: "1.0.0",
					},
					Auth: api.AiApplicationAuth{
						AssetId: assetId,
					},
				},
			}
			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			resp := parseApplicationSyncResponse(syncResponse)
			Expect(resp.Applications).To(HaveLen(1))
			Expect(resp.Applications[0].Application.Spec.Auth.AssetId).To(Equal(assetId))
		})

		It("should return only updates after a given update count", func() {
			createApp("sync-before", clusterId, "proj-1")

			firstSync, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			firstResp := parseApplicationSyncResponse(firstSync)
			Expect(len(firstResp.Applications)).To(Equal(1))

			createApp("sync-after", clusterId, "proj-1")

			secondSync, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, firstResp.ResponseInfo.HighestUpdate, clusterId.String(), api.ClusterSyncReport{})
			secondResp := parseApplicationSyncResponse(secondSync)
			Expect(len(secondResp.Applications)).To(Equal(1))
			Expect(secondResp.Applications[0].Application.Name).To(Equal("sync-after"))
		})

		It("should isolate sync updates by cluster", func() {
			createApp("cluster1-app", clusterId, "proj-1")
			createApp("cluster2-app", cluster2Id, "proj-1")

			sync1, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			resp1 := parseApplicationSyncResponse(sync1)
			Expect(len(resp1.Applications)).To(Equal(1))
			Expect(resp1.Applications[0].Application.Name).To(Equal("cluster1-app"))

			sync2, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, cluster2Id.String(), api.ClusterSyncReport{})
			resp2 := parseApplicationSyncResponse(sync2)
			Expect(len(resp2.Applications)).To(Equal(1))
			Expect(resp2.Applications[0].Application.Name).To(Equal("cluster2-app"))
		})

		It("should NOT soft-delete synced deleted applications via RequestApplicationsUpdates", func() {
			_ = createApp("app-keep", clusterId, "proj-1")
			id2 := createApp("app-delete", clusterId, "proj-1")

			// Agent deploys app-delete so it transitions past "creating"
			_, syncErr := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:          id2,
						Name:        "app-delete",
						Source:      "runai",
						Type:        "application",
						ProjectId:   "proj-1",
						ProjectName: "project-1",
						K8sPhase:    "deployed",
						Action:      api.AIAPPLICATIONSYNCACTION_UPSERT,
						ActionInfo: &api.AiApplicationSyncActionInfo{
							Status:    "completed",
							HelmPhase: "deployed",
						},
					},
				},
			})
			Expect(syncErr).To(BeNil())

			aiApplicationsService.DeleteApplication(ctx, id2)

			firstSync, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			firstResp := parseApplicationSyncResponse(firstSync)
			Expect(len(firstResp.Applications)).To(Equal(2))

			highestUpdate := firstResp.ResponseInfo.HighestUpdate
			secondSync, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, highestUpdate+1, clusterId.String(), api.ClusterSyncReport{})
			secondResp := parseApplicationSyncResponse(secondSync)
			Expect(len(secondResp.Applications)).To(Equal(0))

			// The deleted app should still be in the DB (not soft-deleted) because
			// only SyncApplications should soft-delete after cluster confirmation
			var app models.Application
			appUUID, _ := uuid.Parse(id2)
			Expect(runaidb.DB().Unscoped().Where("id = ?", appUUID).First(&app)).To(BeNil())
			Expect(app.DeletedAt.Valid).To(BeFalse())
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeleting))
		})

		It("should store 'not found' message without soft-deleting the application", func() {
			appId := createApp("app-not-found", clusterId, "proj-1")

			syncReport := api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:                appId,
						UpdateCount:       0,
						UpdateAttemptTime: time.Now(),
						Message:           convutils.Ref("Application \"app-not-found\" not found"),
					},
				},
			}

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), syncReport)
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			appUuid, _ := uuid.Parse(appId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.Name).To(Equal("app-not-found"))
			Expect(app.Message).NotTo(BeNil())
			Expect(*app.Message).To(ContainSubstring("not found"))
			Expect(app.DeletedAt.Valid).To(BeFalse())
		})

		It("should update message when cluster reports a non-'not found' error", func() {
			appId := createApp("app-error", clusterId, "proj-1")
			errorMessage := "deployment timed out"

			syncReport := api.ClusterSyncReport{
				Updates: []api.ClusterSyncReportUpdate{
					{
						Id:                appId,
						UpdateCount:       0,
						UpdateAttemptTime: time.Now(),
						Message:           convutils.Ref(errorMessage),
					},
				},
			}

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), syncReport)
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			appUuid, _ := uuid.Parse(appId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.Message).NotTo(BeNil())
			Expect(*app.Message).To(Equal(errorMessage))
		})

		It("should include deleted phase in sync items and set actual_phase to deleting for synced apps", func() {
			appId := createApp("app-to-del-sync", clusterId, "proj-1")

			// Agent deploys the app so it transitions past "creating"
			_, syncErr := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:          appId,
						Name:        "app-to-del-sync",
						Source:      "runai",
						Type:        "application",
						ProjectId:   "proj-1",
						ProjectName: "project-1",
						K8sPhase:    "deployed",
						Action:      api.AIAPPLICATIONSYNCACTION_UPSERT,
						ActionInfo: &api.AiApplicationSyncActionInfo{
							Status:    "completed",
							HelmPhase: "deployed",
						},
					},
				},
			})
			Expect(syncErr).To(BeNil())

			appUuid, _ := uuid.Parse(appId)
			appBefore, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(appBefore.ActualPhase).To(Equal(models.ApplicationPhaseDeployed))

			aiApplicationsService.DeleteApplication(ctx, appId)

			syncResponse, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			resp := parseApplicationSyncResponse(syncResponse)

			found := false
			for _, syncInfo := range resp.Applications {
				if syncInfo.Application.Id == appId {
					found = true
					Expect(syncInfo.Application.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
					Expect(syncInfo.Application.Phase).To(Equal(models.ApplicationPhaseDeleting))
					break
				}
			}
			Expect(found).To(BeTrue())

			appAfter, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(appAfter.ActualPhase).To(Equal(models.ApplicationPhaseDeleting))
			Expect(appAfter.DeletedAt.Valid).To(BeFalse())
		})

		It("should report system highest update count", func() {
			createApp("sys-count-app", clusterId, "proj-1")

			syncResponse, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			resp := parseApplicationSyncResponse(syncResponse)
			Expect(resp.SystemInfo.HighestUpdate).To(BeNumerically(">", 0))
			Expect(resp.SystemInfo.HighestUpdate).To(BeNumerically(">=", resp.ResponseInfo.HighestUpdate))
		})

		It("should only return runai-source applications and exclude non-runai sources", func() {
			runaiAppId := createApp("runai-source-app", clusterId, "proj-1")

			helmApp := models.Application{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
				Name:         "helm-source-app",
				Source:       "helm",
				Type:         "application",
				TenantId:     testing.TestTenant,
				ClusterId:    clusterId.String(),
				DepartmentId: "dept-1",
				ProjectId:    "proj-1",
				ProjectName:  "project-1",
				ActualPhase:  models.ApplicationPhaseDeployed,
			}
			dbErr := dao.CreateApplication(ctx, &helmApp, nil, nil)
			Expect(dbErr).To(BeNil())

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			resp := parseApplicationSyncResponse(syncResponse)
			Expect(len(resp.Applications)).To(Equal(1))
			Expect(resp.Applications[0].Application.Id).To(Equal(runaiAppId))
			Expect(resp.Applications[0].Application.Source).To(Equal("runai"))
		})

		It("should return empty when only non-runai source applications exist", func() {
			helmApp := models.Application{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
				Name:         "only-helm-app",
				Source:       "helm",
				Type:         "application",
				TenantId:     testing.TestTenant,
				ClusterId:    clusterId.String(),
				DepartmentId: "dept-1",
				ProjectId:    "proj-1",
				ProjectName:  "project-1",
				ActualPhase:  models.ApplicationPhaseDeployed,
			}
			dbErr := dao.CreateApplication(ctx, &helmApp, nil, nil)
			Expect(dbErr).To(BeNil())

			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			resp := parseApplicationSyncResponse(syncResponse)
			Expect(len(resp.Applications)).To(Equal(0))
		})

		It("agent status updates should not cause re-pull (no feedback loop)", func() {
			appId := createApp("feedback-test-app", clusterId, "proj-1")

			// Agent pulls for the first time and gets the new application
			firstSync, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			firstResp := parseApplicationSyncResponse(firstSync)
			Expect(len(firstResp.Applications)).To(Equal(1))
			Expect(firstResp.Applications[0].Application.Id).To(Equal(appId))

			watermark := firstResp.ResponseInfo.HighestUpdate

			// Agent reports k8s status back via SyncApplications (writes to applications table only)
			appUuid, _ := uuid.Parse(appId)
			statusSync, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:           appId,
						Name:         "feedback-test-app",
						Source:       "runai",
						Type:         "application",
						DepartmentId: "dept-1",
						ProjectId:    "proj-1",
						ProjectName:  "project-1",
						K8sPhase:     "deployed",
						Resources: []api.AiApplicationSyncResource{
							{Id: "res-1", Name: "pod-1", Kind: "pod"},
						},
						Action: api.AIAPPLICATIONSYNCACTION_UPSERT,
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(statusSync.Code).To(Equal(http.StatusNoContent))

			// Verify that the applications table was updated with the new status
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.K8sPhase).To(Equal("deployed"))

			// Agent polls again with the same watermark -- should get 0 results
			// because the SyncApplications write to the applications table does NOT
			// bump update_count in application_crud_sync
			secondSync, err := aiApplicationsService.RequestApplicationsUpdates(ctx, watermark, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			secondResp := parseApplicationSyncResponse(secondSync)
			Expect(len(secondResp.Applications)).To(Equal(0))
		})
	})

	Context("end-to-end deletion lifecycle (synced application)", Ordered, func() {
		clusterId, _ := uuid.Parse(testing.ClusterUuid1)
		var createdAppId string
		var agentWatermark int64

		parseApplicationSyncResponse := func(httpResponse api.ImplResponse) api.AiApplicationSyncResponse {
			bytes, err := json.Marshal(httpResponse.Body)
			Expect(err).To(BeNil())
			response := &api.AiApplicationSyncResponse{}
			err = json.Unmarshal(bytes, response)
			Expect(err).To(BeNil())
			return *response
		}

		AfterAll(func() {
			testing.CleanTables()
		})

		It("Step 1: create application", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "lifecycle-app",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/lifecycle-chart.tgz",
					},
				},
			}
			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			app := createResponse.Body.(api.AiApplication)
			createdAppId = app.Id
			Expect(app.ActionStatus).ToNot(BeNil())
			Expect(*app.ActionStatus).To(Equal(models.ApplicationPhaseCreating))
			Expect(app.Source).To(Equal(models.ApplicationSourceRunai))
		})

		It("Step 2: verify initial state in DB", func() {
			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseCreating))
			Expect(app.DesiredPhase).NotTo(BeNil())
			Expect(*app.DesiredPhase).To(Equal(models.ApplicationPhaseDeployed))
			Expect(app.DeletedAt.Valid).To(BeFalse())
		})

		It("Step 3: agent deploys app on cluster", func() {
			_, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:          createdAppId,
						Name:        "lifecycle-app",
						Source:      "runai",
						Type:        "application",
						ProjectId:   "proj-1",
						ProjectName: "project-1",
						K8sPhase:    "deployed",
						Action:      api.AIAPPLICATIONSYNCACTION_UPSERT,
						ActionInfo: &api.AiApplicationSyncActionInfo{
							Status:    "completed",
							HelmPhase: "deployed",
						},
					},
				},
			})
			Expect(err).To(BeNil())

			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeployed))
		})

		It("Step 4: user deletes synced application - only desired_phase changes", func() {
			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, createdAppId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))

			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(*app.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeployed))
			Expect(app.DeletedAt.Valid).To(BeFalse())
		})

		It("Step 5: agent pulls updates - actual_phase transitions to deleting", func() {
			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			resp := parseApplicationSyncResponse(syncResponse)
			found := false
			for _, syncInfo := range resp.Applications {
				if syncInfo.Application.Id == createdAppId {
					found = true
					Expect(syncInfo.Application.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
					Expect(syncInfo.Application.Phase).To(Equal(models.ApplicationPhaseDeleting))
					Expect(syncInfo.Application.Spec.ChartRef.ChartURL).To(Equal("https://example.com/lifecycle-chart.tgz"))
					break
				}
			}
			Expect(found).To(BeTrue())

			agentWatermark = resp.SystemInfo.HighestUpdate

			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeleting))
			Expect(app.DeletedAt.Valid).To(BeFalse())
		})

		It("Step 6: agent pulls with advanced watermark - app is NOT prematurely soft-deleted", func() {
			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, agentWatermark, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			resp := parseApplicationSyncResponse(syncResponse)
			Expect(len(resp.Applications)).To(Equal(0))

			appUuid, _ := uuid.Parse(createdAppId)
			app, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).To(BeNil())
			Expect(app.DeletedAt.Valid).To(BeFalse())
			Expect(*app.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
			Expect(app.ActualPhase).To(Equal(models.ApplicationPhaseDeleting))
		})

		It("Step 7: cluster confirms deletion via SyncApplications - app is soft-deleted", func() {
			syncResponse, err := aiApplicationsService.SyncApplications(ctx, clusterId.String(), api.AiApplicationsForSync{
				FullSync: false,
				Applications: []api.AiApplicationForSync{
					{
						Id:     createdAppId,
						Action: api.AIAPPLICATIONSYNCACTION_DELETE,
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			applicationsRes, dbErr := dao.GetApplications(ctx, testing.TestTenant, nil, 0, 0, "", "", []queryapi.Filter{}, "", false)
			Expect(dbErr).To(BeNil())
			for _, app := range *applicationsRes {
				Expect(app.ID.String()).NotTo(Equal(createdAppId))
			}

			var softDeletedApp models.Application
			appUUID, _ := uuid.Parse(createdAppId)
			Expect(runaidb.DB().Unscoped().Where("id = ?", appUUID).First(&softDeletedApp)).To(BeNil())
			Expect(softDeletedApp.DeletedAt.Valid).To(BeTrue())
		})
	})

	Context("end-to-end deletion lifecycle (unsynced application)", Ordered, func() {
		clusterId, _ := uuid.Parse(testing.ClusterUuid1)
		var createdAppId string

		AfterAll(func() {
			testing.CleanTables()
		})

		It("Step 1: create application", func() {
			request := api.AiApplicationCreationRequest{
				Name:      "unsynced-lifecycle-app",
				ClusterId: clusterId.String(),
				ProjectId: "proj-1",
				Spec: api.AiApplicationSpec{
					ChartRef: api.AiApplicationSpecChartRef{
						ChartURL: "https://example.com/unsynced-chart.tgz",
					},
				},
			}
			createResponse, err := aiApplicationsService.CreateApplication(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			app := createResponse.Body.(api.AiApplication)
			createdAppId = app.Id
			Expect(*app.ActionStatus).To(Equal(models.ApplicationPhaseCreating))
		})

		It("Step 2: user deletes before agent syncs - app is immediately soft-deleted", func() {
			deleteResponse, err := aiApplicationsService.DeleteApplication(ctx, createdAppId)
			Expect(err).To(BeNil())
			Expect(deleteResponse.Code).To(Equal(http.StatusAccepted))

			appUuid, _ := uuid.Parse(createdAppId)

			// Should be gone from normal queries
			_, dbErr := dao.GetApplicationById(ctx, testing.TestTenant, appUuid, nil)
			Expect(dbErr).NotTo(BeNil())

			// Soft-deleted in DB
			var softDeletedApp models.Application
			Expect(runaidb.DB().Unscoped().Where("id = ?", appUuid).First(&softDeletedApp)).To(BeNil())
			Expect(softDeletedApp.DeletedAt.Valid).To(BeTrue())

			// application_crud_sync row should still exist with desired_phase = deleted
			// so the agent picks up the delete instruction even if it already started deploying
			var crudSync models.ApplicationCrudSync
			Expect(runaidb.DB().Where("application_id = ?", appUuid).First(&crudSync)).To(BeNil())
			Expect(*crudSync.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
		})

		It("Step 3: agent should still see the delete instruction in sync updates", func() {
			syncResponse, err := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusOK))

			bytes, _ := json.Marshal(syncResponse.Body)
			resp := &api.AiApplicationSyncResponse{}
			json.Unmarshal(bytes, resp)

			found := false
			for _, syncInfo := range resp.Applications {
				if syncInfo.Application.Id == createdAppId {
					found = true
					Expect(syncInfo.Application.DesiredPhase).To(Equal(models.ApplicationPhaseDeleted))
					Expect(syncInfo.Application.Phase).To(Equal(models.ApplicationPhaseDeleting))
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("Step 4: subsequent agent poll returns no more sync items", func() {
			bytes, _ := json.Marshal(func() interface{} {
				syncResponse, _ := aiApplicationsService.RequestApplicationsUpdates(ctx, 0, clusterId.String(), api.ClusterSyncReport{})
				return syncResponse.Body
			}())
			resp := &api.AiApplicationSyncResponse{}
			json.Unmarshal(bytes, resp)
			watermark := resp.SystemInfo.HighestUpdate

			// Poll again with advanced watermark - agent should get no more items
			syncResponse2, err := aiApplicationsService.RequestApplicationsUpdates(ctx, watermark, clusterId.String(), api.ClusterSyncReport{})
			Expect(err).To(BeNil())

			bytes2, _ := json.Marshal(syncResponse2.Body)
			resp2 := &api.AiApplicationSyncResponse{}
			json.Unmarshal(bytes2, resp2)
			Expect(len(resp2.Applications)).To(Equal(0))
		})
	})
})

func validateApplicationWithVerbosity(actual api.AiApplication, expected api.AiApplicationForSync, isFull bool) {
	Expect(actual.Name).To(Equal(expected.Name))
	Expect(actual.Id).To(Equal(expected.Id))
	Expect(actual.Source).To(Equal(expected.Source))
	Expect(actual.Type).To(Equal(expected.Type))
	Expect(actual.ProjectName).To(Equal(expected.ProjectName))
	Expect(actual.ProjectId).To(Equal(expected.ProjectId))
	Expect(actual.Phase).To(Equal(expected.K8sPhase))
	resources := actual.Resources
	Expect(len(resources)).To(Equal(len(expected.Resources)))
	actualMap := map[string]api.AiApplicationResource{}
	for _, res := range actual.Resources {
		actualMap[res.Id] = res
	}
	for _, res := range expected.Resources {
		actualRes, ok := actualMap[res.Id]
		Expect(ok).To(BeTrue())
		Expect(actualRes.Name).To(Equal(res.Name))
		Expect(actualRes.Kind).To(Equal(res.Kind))
	}
	if isFull {
		Expect(actual.AllocatedResources).NotTo(BeNil())
		Expect(*actual.AllocatedResources.Gpu).To(Equal(float64(2)))
		Expect(*actual.AllocatedResources.Cpu).To(Equal(float64(16)))
		Expect(*actual.AllocatedResources.CpuMemory).To(Equal("4G"))
		Expect(actual.RequestedResources).NotTo(BeNil())
		Expect(*actual.RequestedResources.Gpu.Limit).To(Equal(float64(4)))
		Expect(*actual.RequestedResources.Gpu.Request).To(Equal(float64(2)))
		Expect(*actual.RequestedResources.Cpu.Limit).To(Equal(float64(32)))
		Expect(*actual.RequestedResources.Cpu.Request).To(Equal(float64(16)))
		Expect(*actual.RequestedResources.CpuMemory.Limit).To(Equal("300Mi"))
		Expect(*actual.RequestedResources.CpuMemory.Request).To(Equal("3G"))

		Expect(len(actual.Endpoints)).To(Equal(2))
		endpointsByHost := map[string]api.AiApplicationEndpoint{}
		for _, ep := range actual.Endpoints {
			endpointsByHost[ep.Host] = ep
		}
		ep1, ok := endpointsByHost["chat.example.com"]
		Expect(ok).To(BeTrue())
		Expect(ep1.WorkloadId).To(Equal(testing.TestUuid1))
		Expect(ep1.Port).To(Equal(int32(8080)))
		Expect(ep1.Path).To(Equal("/v1/completions"))
		Expect(ep1.TlsEnabled).To(BeNil())
		Expect(ep1.Protocol).To(Equal("HTTPS"))
		Expect(ep1.Url).To(Equal("https://chat.example.com:8080/v1/completions"))
		ep2, ok := endpointsByHost["api.example.com"]
		Expect(ok).To(BeTrue())
		Expect(ep2.WorkloadId).To(Equal(testing.TestUuid1))
		Expect(ep2.Port).To(Equal(int32(443)))
		Expect(ep2.Path).To(Equal("/health"))
		Expect(*ep2.TlsEnabled).To(BeTrue())
		Expect(ep2.Protocol).To(Equal("HTTPS"))
		Expect(ep2.Url).To(Equal("https://api.example.com/health"))
	} else {
		Expect(actual.AllocatedResources).To(BeNil())
		Expect(actual.RequestedResources).To(BeNil())
	}
}

var _ = Describe("applicationViewToApi resourceConnections mapping", func() {
	It("should flatten every per-workload resource ref into a uses edge", func() {
		view := models.ApplicationView{
			ResourceConnections: models.WorkloadResourceConnectionsArray{
				{
					WorkloadId: "wl-1",
					ResourceRefs: models.WorkloadResourceRefArray{
						{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
						{UID: "uid-2", Name: "regcred", Kind: string(runv1.StorageRefKindSecret)},
					},
				},
				{
					WorkloadId: "wl-2",
					ResourceRefs: models.WorkloadResourceRefArray{
						{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
					},
				},
			},
		}

		result := applicationViewToApi(view, api.AIAPPLICATIONVERBOSITY_BRIEF)

		Expect(result.ResourceConnections).To(Equal([]api.AiApplicationResourceConnection{
			{WorkloadId: "wl-1", Uid: "uid-1", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
			{WorkloadId: "wl-1", Uid: "uid-2", Name: "regcred", Kind: api.ResourceRefKind(runv1.StorageRefKindSecret)},
			{WorkloadId: "wl-2", Uid: "uid-1", Name: "data", Kind: api.ResourceRefKind(runv1.StorageRefKindPVC)},
		}))
	})

	It("should leave resourceConnections empty when the view has none", func() {
		result := applicationViewToApi(models.ApplicationView{}, api.AIAPPLICATIONVERBOSITY_BRIEF)

		Expect(result.ResourceConnections).To(BeEmpty())
	})
})
