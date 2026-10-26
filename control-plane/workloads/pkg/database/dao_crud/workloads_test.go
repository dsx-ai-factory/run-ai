/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"encoding/json"
	"errors"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	crudWorkloadUpsertUpdate  = "UPDATE crud_workloads SET created_at=?,created_by=?,updated_at=?,updated_by=?,deleted_at=?,deleted_by=?,update_count=?,tenant_id=?,cluster_id=?,department_id=?,project_id=?,workload_type=?,requested_name=?,name=?,namespace=?,spec=?,master_spec=?,master_same_as_worker=?,desired_phase=?,assets_mapping=? WHERE crud_workloads.deleted_at IS NULL AND id = ?"
	crudWorkloadDesiredUpdate = "UPDATE crud_workloads SET desired_phase=?,updated_at=?,updated_by=? WHERE tenant_id = ? AND id = ?"
	crudWorkloadGetPrefix     = "SELECT * FROM crud_workloads WHERE tenant_id = ? AND id = ? AND workload_type = ?"
	crudWorkloadGetSuffix     = " ORDER BY crud_workloads.id LIMIT ?"
	crudWorkloadListSelect    = "SELECT * FROM crud_workloads WHERE tenant_id = ? AND workload_type = ? AND cluster_id = ? AND update_count > ? ORDER BY update_count LIMIT ?"
	crudWorkloadCleanupDelete = "DELETE FROM crud_workloads WHERE tenant_id = ? AND cluster_id = ? AND workload_type = ? AND desired_phase = ? AND update_count < ?"
	crudWorkloadProjectUpdate = "UPDATE crud_workloads SET desired_phase=? WHERE tenant_id = ? AND cluster_id = ? AND project_id = ?"
	crudWorkloadTenantDelete  = "DELETE FROM crud_workloads WHERE tenant_id = ?"
	crudWorkloadClusterDelete = "DELETE FROM crud_workloads WHERE tenant_id = ? AND cluster_id = ?"
	crudWorkloadByIdDelete    = "DELETE FROM crud_workloads WHERE crud_workloads.id = ? AND crud_workloads.tenant_id = ?"
	crudWorkloadSyncProject   = "DELETE FROM crud_workloads WHERE tenant_id = ? AND project_id = ?"
	crudWorkloadSyncFull      = "DELETE FROM crud_workloads WHERE tenant_id = ? AND cluster_id = ? AND id >= ? AND id <= ? AND NOT id IN (?,?) AND workload_type = ?"
)

type crudWorkloadRow struct {
	ID           uuid.UUID
	TenantId     int32
	ClusterId    string
	ProjectId    string
	WorkloadType string
	Name         string
	UpdateCount  int64
}

func crudWorkloadGetScopedSelect(where string) string {
	return crudWorkloadGetPrefix + where + crudWorkloadGetSuffix
}

var crudWorkloadFixtureId = uuid.MustParse("22222222-2222-2222-2222-222222222222")

func newCrudWorkloadFixture() *models_crud.Workload {
	return &models_crud.Workload{
		ID:           crudWorkloadFixtureId,
		TenantId:     1,
		ClusterId:    "cluster-1",
		DepartmentId: "dep-1",
		ProjectId:    "p1",
		WorkloadType: "workspace",
		Name:         "wl",
		Namespace:    "runai-p1",
	}
}

func crudWorkloadCtxWithSubject(tenantId uint, subject string) context.Context {
	return context.WithValue(testCtx(tenantId), runaictx.CfSubjectID, subject)
}

func expectDesiredPhaseUpdates(m dbtesting.Mocker, rowsAffected int64) {
	m.ExpectTransaction().Begin()
	m.ExpectExec(crudWorkloadDesiredUpdate).Returns(rowsAffected)
	m.ExpectTransaction().Commit()
}

var _ = Describe("DbUpsertWorkload", func() {
	It("upserts the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadUpsertUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertWorkload(testCtx(1), newCrudWorkloadFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the upsert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadUpsertUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertWorkload(testCtx(1), newCrudWorkloadFixture())

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpdateDesiredPhase", func() {
	It("updates the desired phase", func() {
		m := setupDaoMock(GinkgoT())
		expectDesiredPhaseUpdates(m, 1)

		updateTime, apiErr := DbUpdateDesiredPhase(crudWorkloadCtxWithSubject(1, "user-1"), 1, "workspace",
			crudWorkloadFixtureId.String(), nil, api.WORKLOADDESIREDPHASE_DELETED)

		Expect(apiErr).To(BeNil())
		Expect(updateTime).ToNot(BeNil())
	})

	It("returns a not found error when no rows are affected", func() {
		m := setupDaoMock(GinkgoT())
		expectDesiredPhaseUpdates(m, 0)

		updateTime, apiErr := DbUpdateDesiredPhase(testCtx(1), 1, "workspace",
			crudWorkloadFixtureId.String(), nil, api.WORKLOADDESIREDPHASE_DELETED)

		Expect(updateTime).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(404)))
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadDesiredUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		updateTime, apiErr := DbUpdateDesiredPhase(testCtx(1), 1, "workspace",
			crudWorkloadFixtureId.String(), nil, api.WORKLOADDESIREDPHASE_DELETED)

		Expect(updateTime).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkload", func() {
	It("returns the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(crudWorkloadGetScopedSelect("")).Returns([]crudWorkloadRow{{ID: crudWorkloadFixtureId, Name: "wl"}})

		workload, apiErr := DbGetWorkload(testCtx(1), 1, "workspace", crudWorkloadFixtureId.String(), nil)

		Expect(apiErr).To(BeNil())
		Expect(workload).ToNot(BeNil())
		Expect(workload.Name).To(Equal("wl"))
	})

	It("returns an error when the record is not found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(crudWorkloadGetScopedSelect("")).Returns([]crudWorkloadRow{})

		workload, apiErr := DbGetWorkload(testCtx(1), 1, "workspace", crudWorkloadFixtureId.String(), nil)

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(crudWorkloadGetScopedSelect("")).ReturnsError(errors.New("boom"))

		workload, apiErr := DbGetWorkload(testCtx(1), 1, "workspace", crudWorkloadFixtureId.String(), nil)

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("addPermittedScopesQueries", func() {
	systemTrue := true
	systemFalse := false

	DescribeTable("builds the scope filter",
		func(scopes *authorization.PermittedScopes, where string) {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(crudWorkloadGetScopedSelect(where)).
				Returns([]crudWorkloadRow{{ID: crudWorkloadFixtureId, Name: "wl"}})

			workload, apiErr := DbGetWorkload(testCtx(1), 1, "workspace", crudWorkloadFixtureId.String(), scopes)

			Expect(apiErr).To(BeNil())
			Expect(workload).ToNot(BeNil())
		},
		Entry("no scopes at all", nil, ""),
		Entry("system scope", &authorization.PermittedScopes{System: &systemTrue}, ""),
		Entry("tenant scope", &authorization.PermittedScopes{System: &systemFalse, Tenants: []string{"1"}}, ""),
		Entry("empty scopes", &authorization.PermittedScopes{}, " AND (false)"),
		Entry("cluster scope", &authorization.PermittedScopes{Clusters: []string{"c1"}}, " AND ((false or cluster_id IN ( 'c1')))"),
		Entry("department scope", &authorization.PermittedScopes{Departments: []string{"d1"}}, " AND ((false or department_id IN ( 'd1')))"),
		Entry("project scope", &authorization.PermittedScopes{Projects: []string{"p1"}}, " AND ((false or project_id IN ( 'p1')))"),
		Entry("cluster, department and project scopes together",
			&authorization.PermittedScopes{Clusters: []string{"c1"}, Departments: []string{"d1"}, Projects: []string{"p1"}},
			" AND ((false or cluster_id IN ( 'c1') or department_id IN ( 'd1') or project_id IN ( 'p1')))"),
	)
})

var _ = Describe("DbListUpdates", func() {
	It("returns the updated workloads", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(crudWorkloadListSelect).Returns([]crudWorkloadRow{
			{ID: crudWorkloadFixtureId, UpdateCount: 3},
		})

		records, apiErr := DbListUpdates(testCtx(1), 1, "workspace", "cluster-1", 2, 100)

		Expect(apiErr).To(BeNil())
		Expect(records).To(HaveLen(1))
		Expect(records[0].UpdateCount).To(Equal(int64(3)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(crudWorkloadListSelect).ReturnsError(errors.New("boom"))

		records, apiErr := DbListUpdates(testCtx(1), 1, "workspace", "cluster-1", 2, 100)

		Expect(records).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("CleanupObsoleteWorkloadUpdates", func() {
	It("deletes the obsolete updates", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadCleanupDelete).Returns(2)
		m.ExpectTransaction().Commit()

		apiErr := CleanupObsoleteWorkloadUpdates(testCtx(1), 1, "workspace", "cluster-1", 5)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadCleanupDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := CleanupObsoleteWorkloadUpdates(testCtx(1), 1, "workspace", "cluster-1", 5)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DeleteWorkloadsOfProjectId", func() {
	It("marks the project workloads as deleted", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadProjectUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadsOfProjectId(testCtx(1), "1", "cluster-1", "p1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadProjectUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadsOfProjectId(testCtx(1), "1", "cluster-1", "p1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteWorkloadsOfTenantId", func() {
	It("deletes the tenant workloads", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadTenantDelete).Returns(3)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadsOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadTenantDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadsOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteWorkloadsOfClusterId", func() {
	It("deletes the cluster workloads", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadClusterDelete).Returns(3)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadsOfClusterId(testCtx(1), "1", "cluster-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadClusterDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadsOfClusterId(testCtx(1), "1", "cluster-1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteWorkloadById", func() {
	It("deletes the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadByIdDelete).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadById(testCtx(1), crudWorkloadFixtureId)

		Expect(err).ToNot(HaveOccurred())
	})

	It("succeeds when no rows are affected", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadByIdDelete).Returns(0)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadById(testCtx(1), crudWorkloadFixtureId)

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadByIdDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadById(testCtx(1), crudWorkloadFixtureId)

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DbAddChangeStateRecord", func() {
	It("writes the change state record", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadUpsertUpdate,
			dbtesting.AnyTime(), "user-1", dbtesting.AnyTime(), "user-1", nil, "", int64(0),
			int64(1), "33333333-3333-3333-3333-333333333333", "dep-1", "p1", "workspace", "",
			"wl", "runai-p1", `{"image":"img-1"}`, nil, nil, "Deleted", nil,
			crudWorkloadFixtureId).Returns(1)
		m.ExpectTransaction().Commit()

		view := &models.WorkloadView{
			ID:           crudWorkloadFixtureId,
			TenantID:     1,
			ClusterID:    uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			DepartmentID: "dep-1",
			ProjectID:    "p1",
			Name:         "wl",
			Namespace:    "runai-p1",
			Images:       pq.StringArray{"img-1", "img-2"},
		}

		updateTime, apiErr := DbAddChangeStateRecord(crudWorkloadCtxWithSubject(1, "user-1"), "workspace", view,
			api.WORKLOADDESIREDPHASE_DELETED)

		Expect(apiErr).To(BeNil())
		Expect(updateTime).ToNot(BeNil())
	})

	It("writes the record when there are no images and no subject", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadUpsertUpdate,
			dbtesting.AnyTime(), "", dbtesting.AnyTime(), "", nil, "", int64(0),
			int64(1), "00000000-0000-0000-0000-000000000000", "", "", "workspace", "",
			"wl", "", `{}`, nil, nil, "Deleted", nil,
			crudWorkloadFixtureId).Returns(1)
		m.ExpectTransaction().Commit()

		view := &models.WorkloadView{
			ID:       crudWorkloadFixtureId,
			TenantID: 1,
			Name:     "wl",
		}

		updateTime, apiErr := DbAddChangeStateRecord(testCtx(1), "workspace", view, api.WORKLOADDESIREDPHASE_DELETED)

		Expect(apiErr).To(BeNil())
		Expect(updateTime).ToNot(BeNil())
	})

	It("returns an error when the write fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadUpsertUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		view := &models.WorkloadView{ID: crudWorkloadFixtureId, TenantID: 1, Name: "wl"}

		updateTime, apiErr := DbAddChangeStateRecord(testCtx(1), "workspace", view, api.WORKLOADDESIREDPHASE_DELETED)

		Expect(updateTime).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbSyncWorkloads", func() {
	It("returns an error when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbSyncWorkloads(testCtx(1), 1, "p1", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns a bad request error when no scope is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Rollback()

		apiErr := DbSyncWorkloads(testCtx(1), 1, "", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(400)))
	})

	It("syncs the workloads of a project", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadSyncProject).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloads(testCtx(1), 1, "p1", "", nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
	})

	It("syncs with an id range and filters", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadSyncFull).Returns(1)
		m.ExpectTransaction().Commit()

		from := uuid.MustParse("00000000-0000-0000-0000-000000000000")
		to := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
		filters := []queryapi.Filter{
			{Field: "workload_type", Operator: "=", Values: []string{"workspace"}},
		}

		apiErr := DbSyncWorkloads(testCtx(1), 1, "", "cluster-1", &to, &from, []string{"a", "b"}, filters)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadSyncProject).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbSyncWorkloads(testCtx(1), 1, "p1", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(crudWorkloadSyncProject).Returns(1)

		apiErr := DbSyncWorkloads(testCtx(1), 1, "p1", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = DescribeTable("GPU memory DB boundary", func(snapshot bool) {
	m := setupDaoMock(GinkgoT())
	var spec, storedSpec workloadsApi.SupersetSpec
	var master, storedMaster workloadsApi.MasterSpec
	Expect(json.Unmarshal([]byte(`{"compute":{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5,"gpuComputeMode":"sm-sharing"}}`), &spec)).To(Succeed())
	Expect(json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5,"gpuComputeMode":"sm-sharing"}}`), &storedSpec)).To(Succeed())
	Expect(json.Unmarshal([]byte(`{"compute":{"gpuMemoryPortionRequest":0.75}}`), &master)).To(Succeed())
	Expect(json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":0.75}}`), &storedMaster)).To(Succeed())
	before, err := json.Marshal(spec)
	Expect(err).ToNot(HaveOccurred())
	beforeMaster, err := json.Marshal(master)
	Expect(err).ToNot(HaveOccurred())
	stored, err := json.Marshal(storedSpec)
	Expect(err).ToNot(HaveOccurred())
	storedMasterJSON, err := json.Marshal(storedMaster)
	Expect(err).ToNot(HaveOccurred())

	query, count, specIndex := crudWorkloadUpsertUpdate, 21, 15
	if snapshot {
		query, count, specIndex = snapshotUpsertUpdate, 17, 12
	}
	args := make([]any, count)
	for i := range args {
		args[i] = dbtesting.AnyArg()
	}
	args[specIndex], args[specIndex+1] = string(stored), string(storedMasterJSON)
	m.ExpectTransaction().Begin()
	m.ExpectExec(query, args...).Returns(1)
	m.ExpectTransaction().Commit()
	if snapshot {
		record := newSnapshotFixture()
		record.Spec, record.MasterSpec = &spec, &master
		Expect(DbUpsertWorkloadV2Snapshot(testCtx(1), record)).To(BeNil())
		Expect(record.Spec).To(BeIdenticalTo(&spec))
		Expect(record.MasterSpec).To(BeIdenticalTo(&master))
	} else {
		record := newCrudWorkloadFixture()
		record.Spec, record.MasterSpec = &spec, &master
		Expect(DbUpsertWorkload(testCtx(1), record)).To(BeNil())
		Expect(record.Spec).To(BeIdenticalTo(&spec))
		Expect(record.MasterSpec).To(BeIdenticalTo(&master))
	}
	after, err := json.Marshal(spec)
	Expect(err).ToNot(HaveOccurred())
	Expect(after).To(Equal(before))
	afterMaster, err := json.Marshal(master)
	Expect(err).ToNot(HaveOccurred())
	Expect(afterMaster).To(Equal(beforeMaster))
},
	Entry("stores workload aliases without changing the caller's spec", false),
	Entry("stores snapshot aliases without changing the caller's spec", true),
)
