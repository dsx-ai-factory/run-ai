/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	workloadsMgrClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloads_mgr_api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

// workloadsMgrMock is a minimal in-process mock for workloads_mgr.ClientInterface.
// It records the last BatchWorkloads call so tests can assert on it.
type workloadsMgrMock struct {
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
	succeeded := make([]string, len(batch.Ids))
	copy(succeeded, batch.Ids)
	return &workloads_mgr_api.BatchResponse{Succeeded: succeeded}, nil
}

func (m *workloadsMgrMock) CreateWorkloadV2(_ context.Context, _ workloads_mgr_api.WorkloadV2CreateRequest) (*workloads_mgr_api.WorkloadV2, *workloads_mgr_api.Error) {
	return nil, nil
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
		ID:                 id,
		TenantID:           tenantId,
		ClusterID:          clusterUUID,
		ProjectID:          projectId,
		Name:               "v2-test-" + id.String()[:8],
		Type:               "GenericJob",
		Phase:              api.PHASE_RUNNING,
		PhaseUpdatedAt:     time.Now(),
		SupportSuspend:     supportSuspend,
		SupportDelete:      true,
		SubmittedBy:        testing.TestUser, // matches the context subject so ownership-protection check passes
		Urls:               pq.StringArray{},
		Images:             pq.StringArray{},
		RequestedMigDevice: pq.StringArray{},
		RequestedNodepools: pq.StringArray{},
	}
	result := dao_utils.DB().Table("workloads").Create(&w)
	Expect(result.Error).To(BeNil(), "failed to insert test V2 workload")

	// Set SourceApi to WorkloadsV2 in a separate UPDATE (InternalCreatedWorkload holds this field).
	sourceApi := string(api.SOURCEAPI_WORKLOADS_V2)
	result = dao_utils.DB().Table("workloads").Where("id = ?", id).Update("source_api", sourceApi)
	Expect(result.Error).To(BeNil(), "failed to set source_api on test workload")

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
			result := dao_utils.DB().Table(models_crud.WorkloadsTableName).Where("id = ?", id).Count(&count)
			Expect(result.Error).To(BeNil())
			Expect(count).To(BeZero())

			var phase string
			result = dao_utils.DB().Table("workloads").Where("id = ?", id).Pluck("phase", &phase)
			Expect(result.Error).To(BeNil())
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
			dao_utils.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))

			_, err := ChangeDesiredStates(ctx, []string{id}, api.WORKLOADDESIREDPHASE_RUNNING)
			Expect(err).To(BeNil())

			Expect(mgr.lastAction).To(Equal(workloads_mgr_api.WORKLOADSV2BATCHACTION_RESUME))
			Expect(mgr.lastIds).To(ConsistOf(id))
		})

		It("should return 400 when SupportSuspend=false", func() {
			id := insertV2Workload(randoms.TenantId, randoms.ClusterId, randoms.ProjectIdStr, false)
			dao_utils.DB().Table("workloads").Where("id = ?", id).Update("phase", string(api.PHASE_STOPPED))

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
			// The error is swallowed; the outer call still returns a 2xx code.
			Expect(resp.Code).To(BeNumerically(">=", http.StatusOK))
			Expect(resp.Code).To(BeNumerically("<", http.StatusMultipleChoices))
		})
	})
})
