/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/stretchr/testify/require"
)

const (
	sqlUpdateProcessedByAgent        = `UPDATE sync_status SET processed_by_agent=? WHERE (sync_status.workload_id = ? AND sync_status.tenant_id = ? AND sync_status.cluster_id = ? AND sync_status.sent_to_agent = ?) AND workload_id = ?`
	sqlUpdateSyncMessage             = `UPDATE sync_status SET sync_message=? WHERE (sync_status.workload_id = ? AND sync_status.tenant_id = ? AND sync_status.cluster_id = ? AND sync_status.sent_to_agent = ?) AND workload_id = ? AND workload_id = ?`
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

func clusterUpdateArgs(value any) []any {
	return []any{value, testWorkloadID, testTenantID, testClusterID, testUpdateCount, testWorkloadID}
}

func clusterMessageArgs(value any) []any {
	return []any{value, testWorkloadID, testTenantID, testClusterID, testUpdateCount, testWorkloadID, testWorkloadID}
}

func deletedSyncStatus() models.SyncStatusRecord {
	deleted := testSyncStatusRecord()
	deleted.Deleting = true
	deleted.ProcessedByAgent = testUpdateCount
	return deleted
}

func TestDbApplyClusterReport_NoUpdates(t *testing.T) {
	setupSQLMock(t)
	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, &api.ClusterSyncReport{}))
}

func TestDbApplyClusterReport(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		Returns([]models.SyncStatusRecord{})
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_IgnoresUnmatchedUpdate(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(0)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(0)
	mocker.ExpectQuery(sqlSelectDeletedFromClusterEmpty, true).
		Returns([]models.SyncStatusRecord{})
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_ContinuesOnProcessedByAgentFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).
		ReturnsError(errors.New("boom"))
	mocker.ExpectQuery(sqlSelectDeletedFromClusterEmpty, true).
		Returns([]models.SyncStatusRecord{})
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_DeletesCompletedDeletions(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		Returns([]models.SyncStatusRecord{deletedSyncStatus()})
	mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
	mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_LogsAgentDeleteFailureAndRemovesRecords(t *testing.T) {
	mocker := setupSQLMock(t)
	deleted := deletedSyncStatus()
	deleted.SyncMessage = convutils.Ref("agent could not delete")
	msg := "agent could not delete"

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(&msg)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		Returns([]models.SyncStatusRecord{deleted})
	mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
	mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, &msg)))
}

func TestDbApplyClusterReport_ContinuesOnLocateDeletionFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_ContinuesOnSyncStatusDeleteFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		Returns([]models.SyncStatusRecord{deletedSyncStatus()})
	mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).ReturnsError(errors.New("boom"))
	mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}

func TestDbApplyClusterReport_ContinuesOnQueueDeleteFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpdateProcessedByAgent, clusterUpdateArgs(testUpdateCount)...).Returns(1)
	mocker.ExpectExec(sqlUpdateSyncMessage, clusterMessageArgs(nil)...).Returns(1)
	mocker.ExpectQuery(sqlSelectDeletedFromCluster, testWorkloadID, true).
		Returns([]models.SyncStatusRecord{deletedSyncStatus()})
	mocker.ExpectExec(sqlDeleteSyncStatusByIDs, testWorkloadID).Returns(1)
	mocker.ExpectExec(sqlDeleteQueueByIDs, testWorkloadID).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbApplyClusterReport(ctxWithTenant(), testClusterID, clusterReport(testWorkloadID, testUpdateCount, nil)))
}
