/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"strconv"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type location struct {
	tenantId  uint
	clusterId string
	projectId string
}

var _ = Describe("workload curd - event tests", func() {
	var randoms *testutils.RandomValues
	var ctx, ctx2ndTenant context.Context
	var tenantId1, tenantId2 uint

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	workspacesService := NewWorkspacesService(mockTenantSettingsClient)

	var projectId1 = "100001"
	var projectId2 = "100002"
	var projectId3 = "100003"
	var clusterId1 = uuid.New()
	var clusterId2 = uuid.New()

	AfterEach(func() {
		testing.CleanTables()
	})

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		authMock := testing_crud.NewRandomsBasedAuthMockClient()
		orgunit.SetMock(authMock.OrgUnitMock())

		tenantId1 = randoms.TenantId
		tenantId2 = randoms.TenantId + 1 // +1 to ensure different tenant name
		ctx2ndTenant = testing.SetContext(strconv.Itoa(int(tenantId2)), testing.TestUser, []string{})

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	Context("main flow", func() {
		It("project deletion event", func() {

			expectedWorkloads := make(map[string]location)
			var notExpectedWorkloads []string

			// internal test function injects a workload and sets all mock + expected structure
			injectWorkload := func(ctx context.Context, tenantId uint, clusterId uuid.UUID, projectId string, name string) {

				org_unit.SetMock(org_unit.NewSimpleMock().WithProject(projectId, func() openapi.Project {
					return openapi.Project{
						Id:   projectId,
						Name: "project-1",
						Parent: *openapi.NewNullableDepartmentProjectParent(&openapi.DepartmentProjectParent{
							Id:   randoms.DepartmentId,
							Name: randoms.DepartmentId,
						}),
						ParentId:  randoms.DepartmentId,
						ClusterId: clusterId.String(),
					}
				}()))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(clusterId.String(), testing.BuildDisplayedCluster(clusterId.String(), testing.TestClusterVersionNativeSubmission, nil)))

				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(clusterId.String()))

				request := api.WorkspaceCreationRequest{
					Name:      name,
					ProjectId: projectId,
					ClusterId: clusterId.String(),
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))
				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeNil())
				expectedWorkloads[createPayload.WorkloadId] = location{tenantId, clusterId.String(), projectId}
			}

			// internal test function verifies that expected workload exists and not expected workload aren't (deleted)
			expectWorkloads := func(expectedWorkloads map[string]location, notExpectedWorkloads []string) {
				for workloadId, location := range expectedWorkloads {
					workload, err := dao_crud.DbGetWorkload(ctx, location.tenantId, Workspace, workloadId, nil)
					Expect(err).To(BeNil())
					Expect(workload.DesiredPhase).NotTo(Equal(workloadsApi.WORKLOADDESIREDPHASE_DELETED))
				}

				for _, workloadId := range notExpectedWorkloads {
					workload, err := dao_crud.DbGetWorkload(ctx, randoms.TenantId, Workspace, workloadId, nil)
					Expect(err).To(BeNil())
					Expect(*workload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_DELETED))
				}
			}

			// 2 clusters with 3 projects with the same name
			// each project has 2 workloads
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId1, "ws-1-1")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId1, "ws-1-2")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId2, "ws-1-3")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId2, "ws-1-4")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId3, "ws-1-5")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId3, "ws-1-6")

			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId1, "ws-2-1")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId1, "ws-2-2")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId2, "ws-2-3")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId2, "ws-2-4")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId3, "ws-2-5")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId3, "ws-2-6")

			// Verify that all workloads exists
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)

			// Test Case 1: delete project 1 from cluster 1
			deleteRequest1 := events.ProjectDeleteRequest{
				ClusterId: clusterId1.String(),
				ProjectId: projectId1,
			}
			err := common.HandleProjectDeleteRequest(ctx, &deleteRequest1)
			Expect(err).To(BeNil())
			// Verify that ONLY workloads from project 1 in cluster 1 were deleted
			notExpectedWorkloads = composeNotExpectedWorkloadWithProjectDeleteRequest(expectedWorkloads, &deleteRequest1)
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)

		})
		It("cluster deletion event", func() {

			expectedWorkloads := make(map[string]location)
			var notExpectedWorkloads []string

			// internal test function injects a workload and sets all mock + expected structure
			injectWorkload := func(ctx context.Context, tenantId uint, clusterId uuid.UUID, projectId string, name string) {

				org_unit.SetMock(org_unit.NewSimpleMock().WithProject(projectId, func() openapi.Project {
					return openapi.Project{
						Id:   projectId,
						Name: "project-1",
						Parent: *openapi.NewNullableDepartmentProjectParent(&openapi.DepartmentProjectParent{
							Id:   randoms.DepartmentId,
							Name: randoms.DepartmentId,
						}),
						ParentId:  randoms.DepartmentId,
						ClusterId: clusterId.String(),
					}
				}()))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(clusterId.String(), testing.BuildDisplayedCluster(clusterId.String(), testing.TestClusterVersionNativeSubmission, nil)))

				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(clusterId.String()))

				request := api.WorkspaceCreationRequest{
					Name:      name,
					ProjectId: projectId,
					ClusterId: clusterId.String(),
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))
				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeEmpty())
				expectedWorkloads[createPayload.WorkloadId] = location{tenantId, clusterId.String(), projectId}
			}

			// internal test function verifies that expected workload exists and not expected workload aren't (deleted)
			expectWorkloads := func(expectedWorkloads map[string]location, notExpectedWorkloads []string) {
				for workloadId, location := range expectedWorkloads {
					workload, err := dao_crud.DbGetWorkload(ctx, location.tenantId, Workspace, workloadId, nil)
					Expect(err).To(BeNil())
					Expect(workload.DesiredPhase).NotTo(Equal(workloadsApi.WORKLOADDESIREDPHASE_DELETED))
				}

				for _, workloadId := range notExpectedWorkloads {
					_, err := dao_crud.DbGetWorkload(ctx, randoms.TenantId, Workspace, workloadId, nil)
					Expect(err).NotTo(BeNil())
					Expect(err.Code).To(Equal(int32(http.StatusNotFound)))
				}
			}

			// 2 clusters with 3 projects with the same name
			// each project has 2 workloads
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId1, "ws-1-1")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId1, "ws-1-2")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId2, "ws-1-3")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId2, "ws-1-4")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId3, "ws-1-5")
			injectWorkload(ctx, randoms.TenantId, clusterId1, projectId3, "ws-1-6")

			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId1, "ws-2-1")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId1, "ws-2-2")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId2, "ws-2-3")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId2, "ws-2-4")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId3, "ws-2-5")
			injectWorkload(ctx, randoms.TenantId, clusterId2, projectId3, "ws-2-6")

			// Verify that all workloads exists
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)

			// Test Case 1: delete project 1 from cluster 1
			deleteRequest := events.ClusterDelete{
				ClusterId: clusterId1.String(),
			}
			err := common.HandleClusterDeleteRequest(ctx, &deleteRequest)
			Expect(err).To(BeNil())
			// Verify that ONLY workloads from project 1 in cluster 1 were deleted
			notExpectedWorkloads = composeNotExpectedWorkloadWithClusterDeleteRequest(expectedWorkloads, &deleteRequest)
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)

		})

		It("tenant deletion event", func() {
			expectedWorkloads := make(map[string]location)
			var notExpectedWorkloads []string

			// internal test function injects a workload and sets all mock + expected structure
			injectWorkload := func(ctx context.Context, tenantId uint, clusterId uuid.UUID, projectId string, name string) {

				org_unit.SetMock(org_unit.NewSimpleMock().WithProject(projectId, func() openapi.Project {
					return openapi.Project{
						Id:   projectId,
						Name: "project-1",
						Parent: *openapi.NewNullableDepartmentProjectParent(&openapi.DepartmentProjectParent{
							Id:   randoms.DepartmentId,
							Name: randoms.DepartmentId,
						}),
						ParentId:  randoms.DepartmentId,
						ClusterId: clusterId.String(),
					}
				}()))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(clusterId.String(), testing.BuildDisplayedCluster(clusterId.String(), testing.TestClusterVersionNativeSubmission, nil)))

				clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(clusterId.String()))

				request := api.WorkspaceCreationRequest{
					Name:      name,
					ProjectId: projectId,
					ClusterId: clusterId.String(),
					Spec: api.WorkspaceSpecSpec{
						Image: convutils.Ref("ubuntu"),
					},
				}
				createResponse, _ := workspacesService.CreateWorkspace(ctx, request)
				Expect(createResponse.Code).To(Equal(http.StatusAccepted))
				createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
				Expect(createPayload.WorkloadId).NotTo(BeEmpty())
				expectedWorkloads[createPayload.WorkloadId] = location{tenantId, clusterId.String(), projectId}
			}

			// internal test function verifies that expected workload exists and not expected workload aren't (deleted)
			expectWorkloads := func(expectedWorkloads map[string]location, notExpectedWorkloads []string) {
				for workloadId, location := range expectedWorkloads {
					workload, err := dao_crud.DbGetWorkload(ctx, location.tenantId, Workspace, workloadId, nil)
					Expect(err).To(BeNil())
					Expect(workload.DesiredPhase).NotTo(Equal(workloadsApi.WORKLOADDESIREDPHASE_DELETED))
				}

				for _, workloadId := range notExpectedWorkloads {
					_, err := dao_crud.DbGetWorkload(ctx, randoms.TenantId, Workspace, workloadId, nil)
					Expect(err).NotTo(BeNil())
					Expect(err.Code).To(Equal(int32(http.StatusNotFound)))
				}
			}

			// 2 tenants with 3 projects with the same name
			// each project has 2 workloads
			injectWorkload(ctx, tenantId1, clusterId1, projectId1, "ws-1-1")
			injectWorkload(ctx, tenantId1, clusterId1, projectId1, "ws-1-2")
			injectWorkload(ctx, tenantId1, clusterId1, projectId2, "ws-1-3")
			injectWorkload(ctx, tenantId1, clusterId1, projectId2, "ws-1-4")
			injectWorkload(ctx, tenantId1, clusterId1, projectId3, "ws-1-5")
			injectWorkload(ctx, tenantId1, clusterId1, projectId3, "ws-1-6")

			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId1, "ws-2-1")
			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId1, "ws-2-2")
			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId2, "ws-2-3")
			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId2, "ws-2-4")
			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId3, "ws-2-5")
			injectWorkload(ctx2ndTenant, tenantId2, clusterId2, projectId3, "ws-2-6")

			// Verify that all workloads exists
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)

			// Test Case 1: delete the first tenant
			deleteRequest := events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			}
			err := common.HandleTenantDeleteRequest(ctx, &deleteRequest)
			Expect(err).To(BeNil())
			// Verify that ONLY workloads from the first tenant were deleted
			notExpectedWorkloads = composeNotExpectedWorkloadWithTenantDeleteRequest(expectedWorkloads, &deleteRequest)
			expectWorkloads(expectedWorkloads, notExpectedWorkloads)
		})
	})
})

func composeNotExpectedWorkloadWithProjectDeleteRequest(expectedWorkloads map[string]location, msg *events.ProjectDeleteRequest) []string {
	var notExpectedWorkloads []string
	// go over all workload
	for workloadId, location := range expectedWorkloads {
		// if the workload is in the location that was target to deletion, move it to the not expected workload list
		if location.projectId == msg.ProjectId && location.clusterId == msg.ClusterId {
			notExpectedWorkloads = append(notExpectedWorkloads, workloadId)
		}
	}

	// purge "deleted" workload from the expected map
	for _, workloadId := range notExpectedWorkloads {
		delete(expectedWorkloads, workloadId)
	}

	return notExpectedWorkloads
}

func composeNotExpectedWorkloadWithClusterDeleteRequest(expectedWorkloads map[string]location, msg *events.ClusterDelete) []string {
	var notExpectedWorkloads []string
	// go over all workload
	for workloadId, location := range expectedWorkloads {
		// if the workload is in the location that was target to deletion, move it to the not expected workload list
		if location.clusterId == msg.ClusterId {
			notExpectedWorkloads = append(notExpectedWorkloads, workloadId)
		}
	}

	// purge "deleted" workload from the expected map
	for _, workloadId := range notExpectedWorkloads {
		delete(expectedWorkloads, workloadId)
	}

	return notExpectedWorkloads
}

func composeNotExpectedWorkloadWithTenantDeleteRequest(expectedWorkloads map[string]location, msg *events.TenantDelete) []string {
	var notExpectedWorkloads []string
	// go over all workload
	for workloadId, location := range expectedWorkloads {
		// if the workload is in the location that was target to deletion, move it to the not expected workload list
		if parsedId, err := strconv.ParseUint(msg.TenantId, 10, 64); err == nil && location.tenantId == uint(parsedId) {
			notExpectedWorkloads = append(notExpectedWorkloads, workloadId)
		}
	}

	// purge "deleted" workload from the expected map
	for _, workloadId := range notExpectedWorkloads {
		delete(expectedWorkloads, workloadId)
	}

	return notExpectedWorkloads
}
