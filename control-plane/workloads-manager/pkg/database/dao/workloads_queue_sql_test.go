/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

const (
	sqlUpdateWorkloadOptimistic = `UPDATE workloads_queue SET desired_phase=?,effective_manifest=?,raw_manifest=?,updated_at=?,updated_by=? WHERE id = ? AND update_count = ? AND tenant_id = ?`
	sqlSaveWorkload             = `UPDATE workloads_queue SET created_at=?,created_by=?,updated_at=?,updated_by=?,deleted_at=?,deleted_by=?,tenant_id=?,cluster_id=?,department_id=?,department_name=?,project_id=?,project_name=?,name=?,category=?,priority_class=?,preemptibility=?,configuration=?,resource_group=?,resource_version=?,resource_kind=?,desired_phase=?,translated_from_v1=?,raw_manifest=?,effective_manifest=?,update_count=?,input_request_path=?,input_request_body=? WHERE workloads_queue.deleted_at IS NULL AND id = ?`
	sqlInsertWorkload           = `INSERT INTO workloads_queue (id,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,tenant_id,cluster_id,department_id,department_name,project_id,project_name,name,category,priority_class,preemptibility,configuration,resource_group,resource_version,resource_kind,desired_phase,translated_from_v1,raw_manifest,effective_manifest,update_count,input_request_path,input_request_body) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlGetWorkloadByID          = `SELECT * FROM workloads_queue WHERE workloads_queue.id = ? AND tenant_id = ? ORDER BY workloads_queue.id LIMIT ?`
	sqlGetWorkloadScoped        = `SELECT * FROM workloads_queue WHERE workloads_queue.id = ? AND tenant_id = ? AND ((false or cluster_id IN ( 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa') or department_id IN ( 'dept-1') or project_id IN ( 'proj-1'))) ORDER BY workloads_queue.id LIMIT ?`
	sqlGetWorkloadDenied        = `SELECT * FROM workloads_queue WHERE workloads_queue.id = ? AND tenant_id = ? AND (false) ORDER BY workloads_queue.id LIMIT ?`
	sqlListByIDs                = `SELECT * FROM workloads_queue WHERE id in (?,?) AND tenant_id = ? AND workloads_queue.deleted_at IS NULL`
	sqlListByIDsSingle          = `SELECT * FROM workloads_queue WHERE id in (?) AND tenant_id = ? AND workloads_queue.deleted_at IS NULL`
	sqlListByIDsScoped          = `SELECT * FROM workloads_queue WHERE id in (?) AND tenant_id = ? AND ((false or cluster_id IN ( 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa') or department_id IN ( 'dept-1') or project_id IN ( 'proj-1'))) AND workloads_queue.deleted_at IS NULL`
	sqlListUpdates              = `SELECT * FROM workloads_queue WHERE (workloads_queue.tenant_id = ? AND workloads_queue.cluster_id = ?) AND update_count > ? AND workloads_queue.deleted_at IS NULL ORDER BY update_count LIMIT ?`
	sqlDeleteWorkloadsTenant    = `DELETE FROM workloads_queue WHERE tenant_id = ?`
	sqlDeleteWorkloadsCluster   = `DELETE FROM workloads_queue WHERE tenant_id = ? AND cluster_id = ?`
	sqlDeleteWorkloadsDept      = `DELETE FROM workloads_queue WHERE tenant_id = ? AND department_id = ?`
	sqlDeleteWorkloadsProject   = `DELETE FROM workloads_queue WHERE tenant_id = ? AND project_id = ?`
)

func expectOptimisticUpdate(mocker dbtesting.Mocker, record *models.WorkloadsQueueRecord) dbtesting.ExecExpectation {
	return mocker.ExpectExec(sqlUpdateWorkloadOptimistic,
		record.DesiredPhase, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyTime(), record.UpdatedBy,
		record.ID, record.UpdateCount, record.TenantId,
	)
}

var _ = Describe("DbUpdateWorkloadOptimistic", func() {
	It("updates when the update count matches", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		mocker.ExpectTransaction().Begin()
		expectOptimisticUpdate(mocker, record).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount)).To(BeNil())
	})

	It("returns a 409 when no rows are affected", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		mocker.ExpectTransaction().Begin()
		expectOptimisticUpdate(mocker, record).Returns(0)
		mocker.ExpectTransaction().Commit()

		err := DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusConflict))
	})

	It("returns a 503 when the update fails", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		mocker.ExpectTransaction().Begin()
		expectOptimisticUpdate(mocker, record).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Rollback()

		err := DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
	})
})

var _ = Describe("DbUpsertWorkload", func() {
	It("saves an existing workload", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlSaveWorkload).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpsertWorkload(ctxWithTenant(), record)).To(BeNil())
	})

	It("inserts a workload without an ID", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		record.ID = ""
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlInsertWorkload).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpsertWorkload(ctxWithTenant(), record)).To(BeNil())
		Expect(record.ID).ToNot(BeEmpty())
	})

	It("returns a 503 when the save fails", func() {
		mocker := setupSQLMock(GinkgoT())
		record := testQueueRecord()
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlSaveWorkload).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Rollback()

		err := DbUpsertWorkload(ctxWithTenant(), record)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
	})
})

var _ = Describe("DbGetWorkload", func() {
	It("returns the matching workload", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, nil)
		Expect(err).To(BeNil())
		Expect(got.ID).To(Equal(testWorkloadID))
		Expect(got.Name).To(Equal(expected.Name))
	})

	It("applies tenant scope", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, tenantScope())
		Expect(err).To(BeNil())
		Expect(got.ID).To(Equal(testWorkloadID))
	})

	It("applies system scope", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, systemScope())
		Expect(err).To(BeNil())
		Expect(got.ID).To(Equal(testWorkloadID))
	})

	It("applies cluster, department, and project scope", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlGetWorkloadScoped, testWorkloadID, testTenantID, 1).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, projectClusterDeptScope())
		Expect(err).To(BeNil())
		Expect(got.ID).To(Equal(testWorkloadID))
	})

	It("returns an error for a denied scope", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlGetWorkloadDenied, testWorkloadID, testTenantID, 1).
			ReturnsError(errors.New("record not found"))

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, &authorization.PermittedScopes{})
		Expect(err).ToNot(BeNil())
		Expect(got).To(BeNil())
	})

	It("returns a 503 when the query fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
			ReturnsError(errors.New("boom"))

		got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, nil)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
		Expect(got).To(BeNil())
	})
})

var _ = Describe("DbListByIdsToMap", func() {
	It("returns a map keyed by workload ID", func() {
		mocker := setupSQLMock(GinkgoT())
		first := testQueueRecord()
		second := testQueueRecord()
		second.ID = testWorkloadID2
		second.Name = "wl-2"
		mocker.ExpectQuery(sqlListByIDs, testWorkloadID, testWorkloadID2, testTenantID).
			Returns([]models.WorkloadsQueueRecord{*first, *second})

		got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID, testWorkloadID2}, nil)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))
		Expect(got[testWorkloadID].Name).To(Equal("wl-1"))
		Expect(got[testWorkloadID2].Name).To(Equal("wl-2"))
	})

	It("applies permitted scopes", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlListByIDsScoped, testWorkloadID, testTenantID).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, projectClusterDeptScope())
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[testWorkloadID].ID).To(Equal(testWorkloadID))
	})

	It("returns an empty map when there are no rows", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlListByIDsSingle, testWorkloadID, testTenantID).
			Returns([]models.WorkloadsQueueRecord{})

		got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, nil)
		Expect(err).To(BeNil())
		Expect(got).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlListByIDsSingle, testWorkloadID, testTenantID).
			ReturnsError(errors.New("boom"))

		got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, nil)
		Expect(err).ToNot(BeNil())
		Expect(got).To(BeNil())
	})
})

var _ = Describe("DbListUpdates", func() {
	It("returns matching updates", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testQueueRecord()
		mocker.ExpectQuery(sqlListUpdates, testTenantID, testClusterID, int64(3), 10).
			Returns([]models.WorkloadsQueueRecord{*expected})

		got, err := DbListUpdates(ctxWithTenant(), testClusterID, 3, 10)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[0].ID).To(Equal(testWorkloadID))
	})

	It("returns an error when the query fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlListUpdates, testTenantID, testClusterID, int64(0), 5).
			ReturnsError(errors.New("boom"))

		got, err := DbListUpdates(ctxWithTenant(), testClusterID, 0, 5)
		Expect(err).ToNot(BeNil())
		Expect(got).To(BeNil())
	})
})

var _ = Describe("CleanWorkloadsOfScope", func() {
	It("deletes by tenant", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteWorkloadsTenant, testTenantID).Returns(2)
		mocker.ExpectTransaction().Commit()

		Expect(CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, nil)).To(BeNil())
	})

	It("deletes by cluster", func() {
		mocker := setupSQLMock(GinkgoT())
		clusterID := testClusterID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteWorkloadsCluster, testTenantID, &clusterID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, &clusterID, nil, nil)).To(BeNil())
	})

	It("deletes by department", func() {
		mocker := setupSQLMock(GinkgoT())
		departmentID := testDepartmentID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteWorkloadsDept, testTenantID, &departmentID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, &departmentID, nil)).To(BeNil())
	})

	It("deletes by project", func() {
		mocker := setupSQLMock(GinkgoT())
		projectID := testProjectID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteWorkloadsProject, testTenantID, &projectID).Returns(3)
		mocker.ExpectTransaction().Commit()

		Expect(CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, &projectID)).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteWorkloadsTenant, testTenantID).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Rollback()

		Expect(CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, nil)).ToNot(BeNil())
	})
})
