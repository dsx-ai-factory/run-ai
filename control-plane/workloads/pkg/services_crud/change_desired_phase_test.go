/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"net/http"
	"time"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloads_mgr_api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

// workloadsMgrMock is a minimal in-process mock for workloads_mgr.ClientInterface.
// It records the last BatchWorkloads call so tests can assert on it.
type workloadsMgrMock struct {
	createResponse *workloads_mgr_api.WorkloadV2
	batchResponse  *workloads_mgr_api.BatchResponse
	lastAction     workloads_mgr_api.WorkloadsV2BatchAction
	lastIds        []string
	returnErr      *workloads_mgr_api.Error
	getResponse    *workloads_mgr_api.WorkloadV2GetResponse
	getErr         *workloads_mgr_api.Error
	updateResponse *workloads_mgr_api.WorkloadV2
	updateErr      *workloads_mgr_api.Error
}

func (m *workloadsMgrMock) BatchWorkloads(_ context.Context, batch workloads_mgr_api.WorkloadsV2Batch) (*workloads_mgr_api.BatchResponse, *workloads_mgr_api.Error) {
	m.lastAction = batch.Action
	m.lastIds = batch.Ids
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	if m.batchResponse != nil {
		return m.batchResponse, nil
	}
	succeeded := make([]string, len(batch.Ids))
	copy(succeeded, batch.Ids)
	return &workloads_mgr_api.BatchResponse{Succeeded: succeeded}, nil
}

func (m *workloadsMgrMock) CreateWorkloadV2(_ context.Context, _ workloads_mgr_api.WorkloadV2CreateRequest) (*workloads_mgr_api.WorkloadV2, *workloads_mgr_api.Error) {
	return m.createResponse, m.returnErr
}

func (m *workloadsMgrMock) GetWorkloadV2ById(_ context.Context, _ string) (*workloads_mgr_api.WorkloadV2GetResponse, *workloads_mgr_api.Error) {
	return m.getResponse, m.getErr
}

func (m *workloadsMgrMock) UpdateWorkloadV2(_ context.Context, _ string, _ workloads_mgr_api.WorkloadV2UpdateRequest) (*workloads_mgr_api.WorkloadV2, *workloads_mgr_api.Error) {
	return nil, nil
}

func (m *workloadsMgrMock) UpdateWorkloadInternal(_ context.Context, _ string, _ workloads_mgr_api.WorkloadV2UpdateRequest) (*workloads_mgr_api.WorkloadV2, *workloads_mgr_api.Error) {
	return m.updateResponse, m.updateErr
}

// insertV2Workload inserts a minimal V2 workload row directly into the DB for testing.
func insertV2Workload(tenantId uint, clusterId, projectId string, supportSuspend bool) string {
	id := uuid.New()
	clusterUUID, parseErr := uuid.Parse(clusterId)
	Expect(parseErr).To(BeNil(), "clusterId must be a valid UUID")
	w := models.Workload{
		ID:                   id,
		TenantID:             tenantId,
		ClusterID:            clusterUUID,
		ProjectID:            projectId,
		Name:                 "v2-test-" + id.String()[:8],
		Type:                 "GenericJob",
		Phase:                api.PHASE_RUNNING,
		PhaseUpdatedAt:       time.Now(),
		SupportSuspend:       supportSuspend,
		SupportDelete:        true,
		DistributedFramework: string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
		SubmittedBy:          testing.TestUser, // matches the context subject so ownership-protection check passes
		Urls:                 pq.StringArray{},
		Images:               pq.StringArray{},
		RequestedMigDevice:   pq.StringArray{},
		RequestedNodepools:   pq.StringArray{},
	}
	_, err := runaidb.DB().Table("workloads").Create(&w)
	Expect(err).To(BeNil(), "failed to insert test V2 workload")

	// Set SourceApi to WorkloadsV2 in a separate UPDATE (InternalCreatedWorkload holds this field).
	sourceApi := string(api.SOURCEAPI_WORKLOADS_V2)
	_, err = runaidb.DB().Table("workloads").Where("id = ?", id).Update("source_api", sourceApi)
	Expect(err).To(BeNil(), "failed to set source_api on test workload")

	return id.String()
}

var _ = Describe("ChangeDesiredStates — V2 proxy routing", func() {
	var (
		ctx     context.Context
		randoms *testutils.RandomValues
		mgr     *workloadsMgrMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		policies.SetMock(policies.SimpleMockClientFactory{})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)
		orgunit.SetMock(authMock.OrgUnitMock())
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, ptr.To(true))))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))

		mgr = &workloadsMgrMock{}
		workloadsMgrClient.SetMock(mgr)
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	DescribeTable("records exactly one accepted workload state action", func(native bool, workloadType string, phase api.WorkloadDesiredPhase, expected string) {
		id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)
		_, err := runaidb.DB().Table("workloads").Where("id = ?", id).Update("type", workloadType)
		Expect(err).NotTo(HaveOccurred())
		if native {
			_, err = runaidb.DB().Table("workloads").Where("id = ?", id).Update("source_api", "WorkloadsV1")
			Expect(err).NotTo(HaveOccurred())
		}
		if phase == api.WORKLOADDESIREDPHASE_RUNNING {
			_, err = runaidb.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))
			Expect(err).NotTo(HaveOccurred())
		}
		response, err := ChangeDesiredStates(ctx, []string{id}, phase)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
		result := response.Body.(workloadsApi.BatchResponse)
		Expect(result.Succeeded).To(ConsistOf(id))
		Expect(result.Failed).To(BeEmpty())
		if native {
			Expect(mgr.lastIds).To(BeEmpty())
		} else {
			Expect(mgr.lastIds).To(ConsistOf(id))
		}
		rows, _, apiErr := dao.DbGetWorkloadActions(ctx, uuid.MustParse(id), 0, 0)
		Expect(apiErr).To(BeNil())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Action).To(Equal(expected))
		Expect(rows[0].Subject).To(Equal(testing.TestUser))
		Expect(rows[0].SubjectType).To(Equal("user"))
	},
		Entry("native workspace suspend", true, Workspace, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("native workspace resume", true, Workspace, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("native workspace delete", true, Workspace, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("manager-backed workspace suspend", false, Workspace, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("manager-backed workspace resume", false, Workspace, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("manager-backed workspace delete", false, Workspace, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("native training suspend", true, Training, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("native training resume", true, Training, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("native training delete", true, Training, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("manager-backed training suspend", false, Training, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("manager-backed training resume", false, Training, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("manager-backed training delete", false, Training, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("native distributed suspend", true, Distributed, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("native distributed resume", true, Distributed, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("native distributed delete", true, Distributed, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("manager-backed distributed suspend", false, Distributed, api.WORKLOADDESIREDPHASE_STOPPED, "suspend"),
		Entry("manager-backed distributed resume", false, Distributed, api.WORKLOADDESIREDPHASE_RUNNING, "resume"),
		Entry("manager-backed distributed delete", false, Distributed, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("native inference delete", true, Inference, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("manager-backed inference delete", false, Inference, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("native distributed inference delete", true, DistributedInference, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
		Entry("manager-backed distributed inference delete", false, DistributedInference, api.WORKLOADDESIREDPHASE_DELETED, "delete"),
	)
	It("does not record a manager-backed workspace state change rejected before delivery", func() {
		id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, false)
		_, err := runaidb.DB().Table("workloads").Where("id = ?", id).Update("type", Workspace)
		Expect(err).NotTo(HaveOccurred())
		response, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_STOPPED)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(mgr.lastIds).To(BeEmpty())
		expectWorkloadActions(ctx, id)
	})

	It("does not record a rejected native workspace state change", func() {
		id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)
		_, err := runaidb.DB().Table("workloads").Where("id = ?", id).Updates(map[string]any{
			"type":       Workspace,
			"source_api": "WorkloadsV1",
			"phase":      api.PHASE_RUNNING,
		})
		Expect(err).NotTo(HaveOccurred())

		response, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_RUNNING)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(mgr.lastIds).To(BeEmpty())
		expectWorkloadActions(ctx, id)
	})

	It("records only accepted workspace IDs in a mixed manager batch", func() {
		accepted := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)
		rejected := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)
		_, err := runaidb.DB().Table("workloads").Where("id IN ?", []string{accepted, rejected}).Update("type", Workspace)
		Expect(err).NotTo(HaveOccurred())
		mgr.batchResponse = &workloads_mgr_api.BatchResponse{Succeeded: []string{accepted}, Failed: []workloads_mgr_api.BatchResponseFailureItem{{Id: rejected, Code: http.StatusBadRequest, Message: "rejected"}}}
		response, err := ChangeDesiredStates(ctx, []string{accepted, rejected}, api.WORKLOADDESIREDPHASE_STOPPED)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
		result := response.Body.(workloadsApi.BatchResponse)
		Expect(result.Succeeded).To(ConsistOf(accepted))
		Expect(result.Failed).To(ConsistOf(workloadsApi.BatchResponseFailureItem{Id: rejected, Code: http.StatusBadRequest, Message: "rejected"}))
		rows, _, apiErr := dao.DbGetWorkloadActions(ctx, uuid.MustParse(accepted), 0, 0)
		Expect(apiErr).To(BeNil())
		Expect(rows).To(HaveLen(1))
		rows, _, apiErr = dao.DbGetWorkloadActions(ctx, uuid.MustParse(rejected), 0, 0)
		Expect(apiErr).To(BeNil())
		Expect(rows).To(BeEmpty())
	})

	It("collects missing workload failures across concurrent batches", func() {
		processor := newWorkloadBatchProcessor(
			&authorization_api.PermittedScopes{System: ptr.To(true)},
			api.WORKLOADDESIREDPHASE_DELETED,
			randoms.TenantId,
		)
		missingIDs := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
		response := processor.processMultipleWorkloads(ctx, missingIDs, 1)
		Expect(response.Code).To(Equal(http.StatusAccepted))
		result := response.Body.(workloadsApi.BatchResponse)
		Expect(result.Succeeded).To(BeEmpty())
		Expect(result.Failed).To(HaveLen(len(missingIDs)))
		failedIDs := make([]string, 0, len(result.Failed))
		for _, failure := range result.Failed {
			failedIDs = append(failedIDs, failure.Id)
		}
		Expect(failedIDs).To(ConsistOf(missingIDs))
	})

	It("denies a workload when permitted scopes are unavailable", func() {
		Expect(isPermittedForWorkload(&models.WorkloadView{}, nil)).To(BeFalse())
	})

	Context("suspend request (STOPPED desired phase)", func() {
		It("should proxy to workloads-manager with SUSPEND action when SupportSuspend=true", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)

			_, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_STOPPED)
			Expect(err).To(BeNil())

			Expect(mgr.lastAction).To(Equal(workloads_mgr_api.WORKLOADSV2BATCHACTION_SUSPEND))
			Expect(mgr.lastIds).To(ConsistOf(id))
		})

		It("should not create a crud_workloads row for V2 workloads", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)

			_, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_STOPPED)
			Expect(err).To(BeNil())

			var count int64
			Expect(runaidb.DB().Table(models_crud.WorkloadsTableName).Where("id = ?", id).Count(&count)).To(BeNil())
			Expect(count).To(BeZero())

			var phase string
			Expect(runaidb.DB().Table("workloads").Where("id = ?", id).Pluck("phase", &phase)).To(BeNil())
			Expect(phase).To(Equal(string(api.PHASE_STOPPING)))
		})

		It("should return 400 when SupportSuspend=false", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, false)

			resp, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_STOPPED)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			// workloads-manager was NOT called
			Expect(mgr.lastIds).To(BeEmpty())
		})
	})

	Context("resume request (RUNNING desired phase)", func() {
		It("should proxy to workloads-manager with RESUME action when SupportSuspend=true", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)
			// put it in Stopped state so the resume is valid
			runaidb.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))

			_, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_RUNNING)
			Expect(err).To(BeNil())

			Expect(mgr.lastAction).To(Equal(workloads_mgr_api.WORKLOADSV2BATCHACTION_RESUME))
			Expect(mgr.lastIds).To(ConsistOf(id))
		})

		It("should return 400 when SupportSuspend=false", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, false)
			runaidb.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))

			resp, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_RUNNING)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(mgr.lastIds).To(BeEmpty())
		})
	})

	Context("delete request (DELETED desired phase)", func() {
		It("should proxy to workloads-manager with DELETE action regardless of SupportSuspend", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, false)

			_, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_DELETED)
			Expect(err).To(BeNil())

			Expect(mgr.lastAction).To(Equal(workloads_mgr_api.WORKLOADSV2BATCHACTION_DELETE))
			Expect(mgr.lastIds).To(ConsistOf(id))
		})
	})

	Context("workloads-manager batch error", func() {
		It("should swallow a BatchWorkloads transport error and return an empty result", func() {
			// processGenericWorkloadsBatch logs the error and returns without populating
			// Succeeded or Failed, so the caller should still get a 2xx with an empty list.
			mgr.returnErr = &workloads_mgr_api.Error{Code: http.StatusServiceUnavailable, Message: "upstream unavailable"}
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, true)

			resp, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_STOPPED)
			Expect(err).To(BeNil())
			// The manager was reached (error path exercised).
			Expect(mgr.lastIds).To(ConsistOf(id))
			Expect(resp.Code).To(BeNumerically(">=", http.StatusOK))
			Expect(resp.Code).To(BeNumerically("<", http.StatusMultipleChoices))
			expectWorkloadActions(ctx, id)
		})
	})
})
