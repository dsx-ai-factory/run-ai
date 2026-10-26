/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"bytes"
	"context"
	"encoding/json"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	workloads_mgr_api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/policy"

	"github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createWorkspaceRequest(name, projectId, clusterId, priorityCLassName, category string) *api.WorkspaceCreationRequest {
	request := &api.WorkspaceCreationRequest{
		Name:      name,
		ProjectId: projectId,
		ClusterId: clusterId,
		Spec: api.WorkspaceSpecSpec{
			Image: convutils.Ref("ubuntu"),
		},
	}
	if category != "" {
		request.Spec.Category = &category
	}
	if priorityCLassName != "" {
		request.Spec.PriorityClass = &priorityCLassName
	}
	return request
}

func syncPriorityAndValidate(ctx context.Context, service api.WorkloadPropertiesAPIServicer, tenantId, priorityClass, clusterId string) {
	// the priority must exist from the cluster scope
	priority, dbErr := dao.GetPriorityByName(ctx, tenantId, priorityClass)
	Expect(dbErr).To(BeNil())
	statusResponse, err := service.WorkloadPrioritiesStatus(ctx, clusterId, api.PrioritiesForStatus{
		FullSync: false,
		Priorities: []api.PrioritiesForStatusPrioritiesInner{{
			Id:     priority.ID.String(),
			Name:   priorityClass,
			Value:  40,
			Source: priority.Source,
			Status: api.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: "Ready"},
		}}})
	Expect(err).To(BeNil())
	Expect(statusResponse.Code).To(Equal(http.StatusNoContent))
}

var _ = Describe("workspaces service validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	workspacesService := NewWorkspacesService(mockTenantSettingsClient)
	trainingsService := NewTrainingsService(mockTenantSettingsClient)
	workloadPropertiesService := services.NewWorkloadPropertiesService()
	policy.SetMock(policy.NewSimpleMock())

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	It("records public HTTP workspace actions and retains them after deletion", func() {
		const subject = "history-service-account"
		const subjectType = "ServiceAccount"
		requestCtx := context.WithValue(ctx, runaictx.CfSubjectID, subject)
		requestCtx = context.WithValue(requestCtx, runaictx.CfSubjectType, subjectType)
		router := api.NewRouter(api.NewWorkspacesAPIController(workspacesService), api.NewWorkloadActionsAPIController(services.NewWorkloadActionsService()))
		body, err := json.Marshal(createWorkspaceRequest("http-workspace", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/v1/workloads/workspaces", bytes.NewReader(body)).WithContext(requestCtx)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusAccepted), response.Body.String())
		var created workloadsApi.SupersetWorkload
		Expect(json.Unmarshal(response.Body.Bytes(), &created)).To(Succeed())
		id := uuid.MustParse(created.WorkloadId)
		_, err = runaidb.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_RUNNING))
		Expect(err).NotTo(HaveOccurred())
		for _, action := range []string{"suspend", "resume", "delete"} {
			method, path := http.MethodPost, "/api/v1/workloads/workspaces/"+id.String()+"/"+action
			if action == "resume" {
				_, err = runaidb.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))
				Expect(err).NotTo(HaveOccurred())
			}
			if action == "delete" {
				method, path = http.MethodDelete, "/api/v1/workloads/workspaces/"+id.String()
			}
			response = httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(method, path, nil).WithContext(requestCtx))
			Expect(response.Code).To(Equal(http.StatusAccepted), response.Body.String())
		}
		readHistory := func() api.GetWorkloadActionsResponse {
			response = httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/workloads/"+id.String()+"/actions", nil).WithContext(requestCtx))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			var page api.GetWorkloadActionsResponse
			Expect(json.Unmarshal(response.Body.Bytes(), &page)).To(Succeed())
			return page
		}
		page := readHistory()
		Expect(page.Actions).To(HaveLen(4))
		actions := make([]api.WorkloadActionType, 0, len(page.Actions))
		for _, action := range page.Actions {
			actions = append(actions, action.Action)
			Expect(action.Subject).To(Equal(subject))
			Expect(action.SubjectType).To(Equal(subjectType))
			Expect(action.AcceptedAt.IsZero()).To(BeFalse())
		}
		Expect(actions).To(ConsistOf(
			api.WorkloadActionType("create"),
			api.WorkloadActionType("suspend"),
			api.WorkloadActionType("resume"),
			api.WorkloadActionType("delete"),
		))
	})

	It("records a manager-backed workspace create with the manager-assigned ID", func() {
		original := anyWorkloadSubmitters[Workspace]
		DeferCleanup(func() { anyWorkloadSubmitters[Workspace] = original })
		anyWorkloadSubmitters[Workspace] = &fakeAnyWorkloadSubmitter{methods: SubmissionMethods{AnyWorkload: true}, submission: &Submission{}}
		id := uuid.NewString()
		workloadsMgrClient.SetMock(&workloadsMgrMock{createResponse: &workloads_mgr_api.WorkloadV2{DesiredPhase: workloads_mgr_api.DESIREDPHASE_RUNNING, Metadata: workloads_mgr_api.WorkloadV2Metadata{Id: id, Name: "manager-normalized-workspace", ProjectId: randoms.ProjectIdStr, ClusterId: randoms.ClusterId}}})
		response, err := workspacesService.CreateWorkspace(ctx, *createWorkspaceRequest("manager-workspace", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
		created := response.Body.(workloadsApi.SupersetWorkload)
		Expect(created.WorkloadId).To(Equal(id))
		Expect(created.Name).To(Equal("manager-workspace"))
		rows, _, apiErr := dao.DbGetWorkloadActions(ctx, uuid.MustParse(id), 0, 0)
		Expect(apiErr).To(BeNil())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Action).To(Equal("create"))
		Expect(rows[0].WorkloadName).To(Equal("manager-workspace"))
	})

	It("keeps an accepted create successful when history persistence fails", func() {
		_, err := runaidb.DB().Exec("ALTER TABLE workloads_action_history RENAME TO unavailable_action_history")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			_, err := runaidb.DB().Exec("ALTER TABLE unavailable_action_history RENAME TO workloads_action_history")
			Expect(err).NotTo(HaveOccurred())
		})
		response, err := workspacesService.CreateWorkspace(ctx, *createWorkspaceRequest("history-unavailable", randoms.ProjectIdStr, randoms.ClusterId, "", ""))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
	})

	It("does not record a rejected workspace create", func() {
		response, err := workspacesService.CreateWorkspace(ctx, *createWorkspaceRequest("bad-workspace", randoms.ProjectIdStr, randoms.ClusterId, "unknown-priority", ""))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		var count int64
		Expect(runaidb.DB().Table("workloads_action_history").Where("tenant_id = ?", randoms.TenantId).Count(&count)).To(Succeed())
		Expect(count).To(BeZero())
	})
	Context("create", func() {
		It("get should be sensitive to workload type", func() {

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.21.9999", convutils.Ref(true))))
			request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, constants.PriorityClassBuild, "Build")

			createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)

			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createPayload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(createPayload.WorkloadId).NotTo(BeNil())
			actions, _, actionErr := dao.DbGetWorkloadActions(ctx, uuid.MustParse(createPayload.WorkloadId), 0, 0)
			Expect(actionErr).To(BeNil())
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].Action).To(Equal("create"))
			Expect(actions[0].Subject).To(Equal(testing.TestUser))
			Expect(actions[0].SubjectType).To(Equal("user"))
			Expect(actions[0].WorkloadName).To(Equal(createPayload.Name))
			Expect(actions[0].ProjectID).To(Equal(randoms.ProjectIdStr))

			//  expecting the actual phase to be Creating...
			createPayload.ActualPhase = convutils.Ref(workloadsApi.PHASE_CREATING)

			// get workspace should return the correct workspace
			getResponse, _ := workspacesService.GetWorkspace(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload).To(BeEquivalentTo(createPayload))

			Expect(getPayload.Spec.PriorityClass).To(Not(BeNil()))
			Expect(getPayload.Spec.Labels).To(BeNil())

			// get training should return 404 not found
			getResponse, _ = trainingsService.GetTraining(ctx, createPayload.WorkloadId)
			Expect(getResponse.Code).To(Equal(http.StatusNotFound))
		})
	})

	It("Invalid priorityClass", func() {
		priorityClass := "not-existing-priority-class"
		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, "")

		createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
	})

	It("Non existing Category", func() {
		category := "not-existing-category"
		priorityClass := "low"
		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, category)
		syncPriorityAndValidate(ctx, workloadPropertiesService, randoms.TenantIdStr, priorityClass, randoms.ClusterId)
		createResponse, err := workspacesService.CreateWorkspace(ctx, *request)
		Expect(err).To(BeNil())
		Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
	})

	It("Not set Category", func() {
		priorityClass := "low"
		tenantId := randoms.TenantIdStr
		clusterId := randoms.ClusterId
		syncPriorityAndValidate(ctx, workloadPropertiesService, tenantId, priorityClass, clusterId)

		category := ""

		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, category)

		createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		supersetWorkload := createResponse.Body.(workloadsApi.SupersetWorkload)
		Expect(supersetWorkload).To(Not(BeNil()))
		Expect(*supersetWorkload.Spec.Category.Get()).To(Equal("Build"))
	})

	It("Validate priorityClass for version below 2.22", func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.21.9999", convutils.Ref(true))))

		priorityClass := "not-existing-priority-class"
		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, "")

		createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusBadRequest))

		priorityClass = "low"
		request = createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, "")

		createResponse, _ = workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
	})

	It("Invalid priorityClass for version above equal 2.22", func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22.0", convutils.Ref(true))))

		// "medium" priority exists in tenant scope but was not created in the cluster scope
		priorityClass := "medium"
		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, "")

		createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
	})

	It("Valid priorityClass for version above equal 2.22", func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, "2.22.0", convutils.Ref(true))))

		priorityClass := "low"
		syncPriorityAndValidate(ctx, workloadPropertiesService, randoms.TenantIdStr, priorityClass, randoms.ClusterId)

		request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, priorityClass, "")
		createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
		Expect(createResponse.Code).To(Equal(http.StatusAccepted))
	})

	Context("delete", func() {
		It("should be reflected in get", func() {
			request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, "", "")

			createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// simulate real-life time b/w creating and deleting
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ := workspacesService.DeleteWorkspace(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DeletedAt).To(Not(BeNil()))
		})
	})

	Context("suspend", func() {
		It("should be reflected in get", func() {
			request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, "", "")

			createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// attempting to suspend will fail, because we need the status to be running
			phaseResponse, _ := workspacesService.SuspendWorkspace(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusBadRequest))
			errPayload := phaseResponse.Body.(*api.Error)
			Expect(errPayload.Message).To(ContainSubstring("suspend request is not permitted for Creating workloads"))

			// forcefully change the phase to: running
			nowPlus1Min := time.Now().Add(1 * time.Minute)
			dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, nowPlus1Min)
			Expect(dbErr).To(BeNil())

			// simulate real-life time b/w creating and suspending
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ = workspacesService.SuspendWorkspace(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(BeEquivalentTo(workloadsApi.WORKLOADDESIREDPHASE_STOPPED))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(BeEquivalentTo(workloadsApi.PHASE_STOPPING))
		})
	})

	Context("resume", func() {
		It("should be reflected in get", func() {
			request := createWorkspaceRequest("ws-1", randoms.ProjectIdStr, randoms.ClusterId, "", "")

			createResponse, _ := workspacesService.CreateWorkspace(ctx, *request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			payloadCreate := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(payloadCreate.WorkloadId).NotTo(BeNil())

			workloadId := payloadCreate.WorkloadId

			// desired state -> running, actual phase -> creating
			getResponse, _ := workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload := getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(Equal(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(Equal(workloadsApi.PHASE_CREATING))

			// attempting to suspend will fail, because we need the status to be sopped
			phaseResponse, _ := workspacesService.ResumeWorkspace(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusBadRequest))
			errPayload := phaseResponse.Body.(*api.Error)
			Expect(errPayload.Message).To(ContainSubstring("resume request is permitted only for stopped workloads (currently the workload is in Creating phase)"))

			// forcefully change the phase to: running
			nowPlus1Min := time.Now().Add(1 * time.Minute)
			dbErr := dao.DbUpdateWorkloadPhase(ctx, randoms.TenantId, workloadId, api.WorkloadPatchFields{Phase: api.PHASE_STOPPED}, nowPlus1Min)
			Expect(dbErr).To(BeNil())

			// simulate real-life time b/w creating and suspending
			time.Sleep(time.Millisecond * 200)

			phaseResponse, _ = workspacesService.ResumeWorkspace(ctx, workloadId)
			Expect(phaseResponse.Code).To(Equal(http.StatusAccepted))

			getResponse, _ = workspacesService.GetWorkspace(ctx, workloadId)
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			getPayload = getResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(getPayload.DesiredPhase).To(BeEquivalentTo(workloadsApi.WORKLOADDESIREDPHASE_RUNNING))
			Expect(getPayload.ActualPhase).NotTo(BeNil())
			Expect(*getPayload.ActualPhase).To(BeEquivalentTo(workloadsApi.PHASE_RESUMING))
		})
	})
})
