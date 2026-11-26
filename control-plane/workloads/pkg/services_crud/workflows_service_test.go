/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloads_mgr_api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The Argo Workflow workload type's category and priority, as seeded by the database migrations.
const (
	workflowDefaultCategory = "Deploy"
	workflowDefaultPriority = "very-high"
)

func createWorkflowRequest(name, projectId, clusterId, priorityClassName, category string) *api.WorkflowCreationRequest {
	request := &api.WorkflowCreationRequest{
		Name:      name,
		ProjectId: projectId,
		ClusterId: clusterId,
		Spec: api.WorkflowSpecSpec{
			Groups: []api.WorkflowGroupsInner{{
				Name: convutils.Ref("hello"),
				Tasks: []api.WorkflowTask{{
					Name:  convutils.Ref("hello"),
					Image: convutils.Ref("busybox:1.37"),
				}},
			}},
		},
	}
	if category != "" {
		request.Spec.Category = &category
	}
	if priorityClassName != "" {
		request.Spec.PriorityClass = &priorityClassName
	}
	return request
}

var _ = Describe("workflows service validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	var managerMock *workloadsMgrMock
	var managerWorkloadId string

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	workflowsService := NewWorkflowsService(mockTenantSettingsClient)
	workloadPropertiesService := services.NewWorkloadPropertiesService()

	countActions := func() int64 {
		var count int64
		Expect(runaidb.DB().Table("workloads_action_history").Where("tenant_id = ?", randoms.TenantId).Count(&count)).To(Succeed())
		return count
	}

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))

		originalManager := workloadsMgrClient.Get()
		DeferCleanup(func() { workloadsMgrClient.SetMock(originalManager) })
		managerWorkloadId = uuid.NewString()
		managerMock = &workloadsMgrMock{createResponse: &workloads_mgr_api.WorkloadV2{
			DesiredPhase: workloads_mgr_api.DESIREDPHASE_RUNNING,
			Metadata:     workloads_mgr_api.WorkloadV2Metadata{Id: managerWorkloadId, ProjectId: randoms.ProjectIdStr, ClusterId: randoms.ClusterId},
		}}
		workloadsMgrClient.SetMock(managerMock)

		syncPriorityAndValidate(ctx, workloadPropertiesService, randoms.TenantIdStr, workflowDefaultPriority, randoms.ClusterId)
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("create", func() {
		It("submits the workflow to workloads-manager with the workload type's defaults", func() {
			response, err := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			created := response.Body.(workloadsApi.Workflow)
			Expect(created.WorkloadId).To(Equal(managerWorkloadId))
			Expect(created.Name).To(Equal("wf-1"))
			Expect(created.RequestedName).To(Equal("wf-1"))
			Expect(created.ProjectId).To(Equal(randoms.ProjectIdStr))
			Expect(created.ClusterId).To(Equal(randoms.ClusterId))
			Expect(created.CreatedBy).To(Equal(testing.TestUser))
			Expect(created.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(*created.Spec.Category.Get()).To(Equal(workflowDefaultCategory))
			Expect(*created.Spec.PriorityClass.Get()).To(Equal(workflowDefaultPriority))

			Expect(managerMock.createRequest).NotTo(BeNil())
			metadata := managerMock.createRequest.Metadata
			Expect(*metadata.Name).To(Equal("wf-1"))
			Expect(*metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
			Expect(*metadata.Category.Get()).To(Equal(workflowDefaultCategory))
			Expect(*metadata.Priority.Get()).To(Equal(workflowDefaultPriority))
			Expect(managerMock.createRequest.Manifest).NotTo(HaveKey("status"))
		})

		It("generates the name from the requested prefix", func() {
			request := createWorkflowRequest("wf", randoms.ProjectIdStr, randoms.ClusterId, "", "")
			request.UseGivenNameAsPrefix = true

			response, _ := workflowsService.CreateWorkflow(ctx, *request)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			created := response.Body.(workloadsApi.Workflow)
			Expect(created.RequestedName).To(Equal("wf"))
			Expect(created.Name).To(HavePrefix("wf-"))
			Expect(*managerMock.createRequest.Metadata.Name).To(Equal(created.Name))
		})

		It("records the create with the manager-assigned ID", func() {
			response, _ := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
			Expect(response.Code).To(Equal(http.StatusAccepted))

			actions, _, actionErr := dao.DbGetWorkloadActions(ctx, uuid.MustParse(managerWorkloadId), 0, 0, "desc", "accepted_at", nil, "")
			Expect(actionErr).To(BeNil())
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].Action).To(Equal("create"))
			Expect(actions[0].Subject).To(Equal(testing.TestUser))
			Expect(actions[0].WorkloadName).To(Equal("wf-1"))
			Expect(actions[0].ProjectID).To(Equal(randoms.ProjectIdStr))
		})

		It("persists the workflow spec in a snapshot", func() {
			response, _ := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
			Expect(response.Code).To(Equal(http.StatusAccepted))

			var snapshot models_crud.WorkflowSnapshot
			Expect(runaidb.DB().Table(models_crud.WorkloadV2SnapshotsTableName).Where("id = ?", managerWorkloadId).First(&snapshot)).To(Succeed())
			Expect(snapshot.WorkloadType).To(Equal(Workflow))
			Expect(snapshot.Name).To(Equal("wf-1"))
			Expect(snapshot.Namespace).To(Equal(RunaiNsPrefix + "project-1"))
			Expect(snapshot.CreatedBy).To(Equal(testing.TestUser))
			Expect(snapshot.Spec.Groups).To(HaveLen(1))
			Expect(*snapshot.Spec.Groups[0].Name.Get()).To(Equal("hello"))
			Expect(*snapshot.Spec.Category.Get()).To(Equal(workflowDefaultCategory))
		})

		It("returns the workloads-manager error and does not record the create", func() {
			managerMock.returnErr = &workloads_mgr_api.Error{Code: http.StatusBadRequest, Message: "rejected by workloads-manager"}

			response, err := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.(*workloadsApi.SubmissionErrorV2).Message).To(Equal("rejected by workloads-manager"))
			Expect(countActions()).To(BeZero())
		})
	})

	Context("validations", func() {
		It("does not submit or record a workflow with an unknown priorityClass", func() {
			response, err := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "not-existing-priority-class", ""))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(managerMock.createRequest).To(BeNil())
			Expect(countActions()).To(BeZero())
		})

		It("rejects a non existing category", func() {
			response, _ := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", "not-existing-category"))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(managerMock.createRequest).To(BeNil())
		})

		It("rejects an invalid workflow name", func() {
			response, _ := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("Invalid_Name", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(managerMock.createRequest).To(BeNil())
		})

		It("rejects a cluster that is not the project's", func() {
			response, _ := workflowsService.CreateWorkflow(ctx, *createWorkflowRequest("wf-1", randoms.ProjectIdStr, uuid.NewString(), "", ""))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(managerMock.createRequest).To(BeNil())
		})

		It("rejects a compute profile with a GPU memory portion above 1", func() {
			request := createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", "")
			request.Spec.Compute = []api.WorkflowComputeProfile{{Name: convutils.Ref("gpu"), GpuMemoryPortionRequest: convutils.Ref(1.5)}}

			response, _ := workflowsService.CreateWorkflow(ctx, *request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.(*workloadsApi.SubmissionErrorV2).Message).To(ContainSubstring("compute[0].gpuMemoryPortionRequest"))
			Expect(managerMock.createRequest).To(BeNil())
		})

		It("rejects a group node pool that does not exist in the cluster", func() {
			request := createWorkflowRequest("wf-1", randoms.ProjectIdStr, randoms.ClusterId, "", "")
			request.Spec.Groups[0].NodePools = &[]string{"not-existing-node-pool"}

			response, _ := workflowsService.CreateWorkflow(ctx, *request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(managerMock.createRequest).To(BeNil())
		})
	})
})
