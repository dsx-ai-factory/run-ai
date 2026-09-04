/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sqlUpdateClusterReport           = `UPDATE sync_status SET processed_by_agent=?,sync_message=? WHERE (sync_status.workload_id = ? AND sync_status.tenant_id = ? AND sync_status.cluster_id = ? AND sync_status.sent_to_agent = ?) AND workload_id = ?`
	sqlSelectDeletedFromCluster      = `SELECT * FROM sync_status WHERE workload_id in (?) AND deleting = ? AND sent_to_agent = processed_by_agent`
	sqlSelectDeletedFromClusterEmpty = `SELECT * FROM sync_status WHERE workload_id in (NULL) AND deleting = ? AND sent_to_agent = processed_by_agent`
	sqlDeleteSyncStatusByIDs         = `DELETE FROM sync_status WHERE workload_id in (?)`
	sqlDeleteQueueByIDs              = `DELETE FROM workloads_queue WHERE id in (?)`
)

func clusterReport(id string, updateCount int64, message *string) *api.ClusterSyncReport {
	return &api.ClusterSyncReport{
		Updates: []api.ClusterSyncReportUpdate{{
			Id:          id,
			UpdateCount: updateCount,
			Message:     message,
		}},
	}
}

func clusterReportUpdateArgs(message *string) []any {
	return []any{testUpdateCount, message, testWorkloadID, testTenantID, testClusterID, testUpdateCount, testWorkloadID}
}

func deletedSyncStatus() models.SyncStatusRecord {
	deleted := testSyncStatusRecord()
	deleted.Deleting = true
	deleted.ProcessedByAgent = testUpdateCount
	return deleted
}

var _ = Describe("DbApplyClusterReport", func() {
	It("returns immediately when there are no updates", func() {
		setupSQLMock(GinkgoT())

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, &api.ClusterSyncReport{})).To(BeNil())
	})

	It("updates processed_by_agent for a matching report", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			Returns([]models.SyncStatusRecord{})
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("ignores unmatched updates", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(0)
		mocker.ExpectQuery(sqlSelectDeletedFromClusterEmpty, true).
			Returns([]models.SyncStatusRecord{})
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("continues when processed_by_agent update fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).
			ReturnsError(errors.New("boom"))
		mocker.ExpectQuery(sqlSelectDeletedFromClusterEmpty, true).
			Returns([]models.SyncStatusRecord{})
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("deletes completed deletions", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			Returns([]models.SyncStatusRecord{deletedSyncStatus()})
		mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
		mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("logs agent delete failure and still removes records", func() {
		mocker := setupSQLMock(GinkgoT())
		deleted := deletedSyncStatus()
		deleted.SyncMessage = convutils.Ref("agent could not delete")
		msg := "agent could not delete"

		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(&msg)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			Returns([]models.SyncStatusRecord{deleted})
		mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
		mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, &msg))).To(BeNil())
	})

	It("continues when locating completed deletions fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("continues when sync_status delete fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			Returns([]models.SyncStatusRecord{deletedSyncStatus()})
		mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).ReturnsError(errors.New("boom"))
		mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})

	It("continues when queue delete fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectTransaction().Begin()
		mocker.ExpectExec(sqlUpdateClusterReport, clusterReportUpdateArgs(nil)...).Returns(1)
		mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
			Returns([]models.SyncStatusRecord{deletedSyncStatus()})
		mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
		mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).ReturnsError(errors.New("boom"))
		mocker.ExpectTransaction().Commit()

		Expect(DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil))).To(BeNil())
	})
})
