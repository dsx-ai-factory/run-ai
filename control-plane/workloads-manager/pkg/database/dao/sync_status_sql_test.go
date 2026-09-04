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
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
)

const (
	sqlGetSyncStatus           = `SELECT * FROM sync_status WHERE sync_status.workload_id = ? AND sync_status.tenant_id = ? ORDER BY sync_status.workload_id LIMIT ?`
	sqlUpsertSyncStatus        = `INSERT INTO sync_status (workload_id,tenant_id,cluster_id,department_id,project_id,deleting,sent_to_agent,processed_by_agent,sync_message) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (workload_id) DO UPDATE SET sent_to_agent=excluded.sent_to_agent,deleting=excluded.deleting`
	sqlDeleteSyncStatusTenant  = `DELETE FROM sync_status WHERE tenant_id = ?`
	sqlDeleteSyncStatusCluster = `DELETE FROM sync_status WHERE tenant_id = ? AND cluster_id = ?`
	sqlDeleteSyncStatusDept    = `DELETE FROM sync_status WHERE tenant_id = ? AND department_id = ?`
	sqlDeleteSyncStatusProject = `DELETE FROM sync_status WHERE tenant_id = ? AND project_id = ?`
)

var _ = Describe("DbGetSyncStatus", func() {
	It("returns the matching sync status", func() {
		mocker := setupSQLMock(GinkgoT())
		expected := testSyncStatusRecord()
		mocker.ExpectQuery(sqlGetSyncStatus, testWorkloadID, testTenantID, 1).
			Returns([]models.SyncStatusRecord{expected})

		got, err := DbGetSyncStatus(ctxWithTenant(), testWorkloadID)
		Expect(err).To(BeNil())
		Expect(got.WorkloadId).To(Equal(testWorkloadID))
		Expect(got.SentToAgent).To(Equal(testUpdateCount))
	})

	It("returns a 503 when the query fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlGetSyncStatus, testWorkloadID, testTenantID, 1).
			ReturnsError(errors.New("boom"))

		got, err := DbGetSyncStatus(ctxWithTenant(), testWorkloadID)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
		Expect(got).To(BeNil())
	})
})

var _ = Describe("DbUpdateSentToAgent", func() {
	It("commits an empty batch without upserting", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{}, nil)).To(BeNil())
	})

	It("upserts sync status for updates in scope", func() {
		mocker := setupSQLMock(GinkgoT())
		scope := map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()}

		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpsertSyncStatus).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
			Workloads: []api.WorkloadV2Update{{
				Id:           testWorkloadID,
				UpdateCount:  testUpdateCount,
				DesiredPhase: api.DESIREDPHASE_RUNNING,
			}},
		}, scope)).To(BeNil())
	})

	It("still upserts when the workload is missing from scope", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpsertSyncStatus).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
			Workloads: []api.WorkloadV2Update{{
				Id:           testWorkloadID,
				UpdateCount:  testUpdateCount,
				DesiredPhase: api.DESIREDPHASE_DELETED,
			}},
		}, map[string]models.WorkloadsQueueRecord{})).To(BeNil())
	})

	It("continues when the upsert fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpsertSyncStatus).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
			Workloads: []api.WorkloadV2Update{{
				Id:           testWorkloadID,
				UpdateCount:  testUpdateCount,
				DesiredPhase: api.DESIREDPHASE_RUNNING,
			}},
		}, map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()})).To(BeNil())
	})

	It("continues when no rows are affected", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpsertSyncStatus).Returns(0)
		mocker.ExpectTransaction().Commit()

		Expect(DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
			Workloads: []api.WorkloadV2Update{{
				Id:           testWorkloadID,
				UpdateCount:  testUpdateCount,
				DesiredPhase: api.DESIREDPHASE_RUNNING,
			}},
		}, map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()})).To(BeNil())
	})
})

var _ = Describe("CleanSyncStatusOfScope", func() {
	It("deletes by tenant", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteSyncStatusTenant, testTenantID).Returns(2)
		mocker.ExpectTransaction().Commit()

		Expect(CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, nil)).To(BeNil())
	})

	It("deletes by cluster", func() {
		mocker := setupSQLMock(GinkgoT())
		clusterID := testClusterID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteSyncStatusCluster, testTenantID, &clusterID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, &clusterID, nil, nil)).To(BeNil())
	})

	It("deletes by department", func() {
		mocker := setupSQLMock(GinkgoT())
		departmentID := testDepartmentID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteSyncStatusDept, testTenantID, &departmentID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, &departmentID, nil)).To(BeNil())
	})

	It("deletes by project", func() {
		mocker := setupSQLMock(GinkgoT())
		projectID := testProjectID
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteSyncStatusProject, testTenantID, &projectID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, &projectID)).To(BeNil())
	})

	It("returns a 503 when the delete fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlDeleteSyncStatusTenant, testTenantID).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Rollback()

		err := CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, nil)
		Expect(err).ToNot(BeNil())
		Expect(int(err.Code)).To(Equal(http.StatusServiceUnavailable))
	})
})
