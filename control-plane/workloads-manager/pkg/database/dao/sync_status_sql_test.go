/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/stretchr/testify/require"
)

const (
	sqlGetSyncStatus           = `SELECT * FROM sync_status WHERE sync_status.workload_id = ? AND sync_status.tenant_id = ? ORDER BY sync_status.workload_id LIMIT ?`
	sqlUpsertSyncStatus        = `INSERT INTO sync_status (workload_id,tenant_id,cluster_id,department_id,project_id,deleting,sent_to_agent,processed_by_agent,sync_message) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT (workload_id) DO UPDATE SET sent_to_agent=excluded.sent_to_agent,deleting=excluded.deleting`
	sqlDeleteSyncStatusTenant  = `DELETE FROM sync_status WHERE tenant_id = ?`
	sqlDeleteSyncStatusCluster = `DELETE FROM sync_status WHERE tenant_id = ? AND cluster_id = ?`
	sqlDeleteSyncStatusDept    = `DELETE FROM sync_status WHERE tenant_id = ? AND department_id = ?`
	sqlDeleteSyncStatusProject = `DELETE FROM sync_status WHERE tenant_id = ? AND project_id = ?`
)

func TestDbGetSyncStatus(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testSyncStatusRecord()
	mocker.ExpectQuery(sqlGetSyncStatus, testWorkloadID, testTenantID, 1).
		Returns([]models.SyncStatusRecord{expected})

	got, err := DbGetSyncStatus(ctxWithTenant(), testWorkloadID)
	require.Nil(t, err)
	require.Equal(t, testWorkloadID, got.WorkloadId)
	require.Equal(t, testUpdateCount, got.SentToAgent)
}

func TestDbGetSyncStatus_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlGetSyncStatus, testWorkloadID, testTenantID, 1).
		ReturnsError(errors.New("boom"))

	got, err := DbGetSyncStatus(ctxWithTenant(), testWorkloadID)
	require.Error(t, err)
	require.Equal(t, http.StatusServiceUnavailable, int(err.Code))
	require.Nil(t, got)
}

func TestDbUpdateSentToAgent_EmptyBatch(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{}, nil))
}

func TestDbUpdateSentToAgent(t *testing.T) {
	mocker := setupSQLMock(t)
	scope := map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()}

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpsertSyncStatus).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
		Workloads: []api.WorkloadV2Update{{
			Id:           testWorkloadID,
			UpdateCount:  testUpdateCount,
			DesiredPhase: api.DESIREDPHASE_RUNNING,
		}},
	}, scope))
}

func TestDbUpdateSentToAgent_MissingFromScope(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpsertSyncStatus).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
		Workloads: []api.WorkloadV2Update{{
			Id:           testWorkloadID,
			UpdateCount:  testUpdateCount,
			DesiredPhase: api.DESIREDPHASE_DELETED,
		}},
	}, map[string]models.WorkloadsQueueRecord{}))
}

func TestDbUpdateSentToAgent_ContinuesOnUpsertFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpsertSyncStatus).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
		Workloads: []api.WorkloadV2Update{{
			Id:           testWorkloadID,
			UpdateCount:  testUpdateCount,
			DesiredPhase: api.DESIREDPHASE_RUNNING,
		}},
	}, map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()}))
}

func TestDbUpdateSentToAgent_ContinuesWhenNoRowsAffected(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlUpsertSyncStatus).Returns(0)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateSentToAgent(ctxWithTenant(), testClusterID, api.WorkloadsV2Updates{
		Workloads: []api.WorkloadV2Update{{
			Id:           testWorkloadID,
			UpdateCount:  testUpdateCount,
			DesiredPhase: api.DESIREDPHASE_RUNNING,
		}},
	}, map[string]models.WorkloadsQueueRecord{testWorkloadID: *testQueueRecord()}))
}

func TestCleanSyncStatusOfScope(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSyncStatusTenant, testTenantID).Returns(2)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, nil))
}

func TestCleanSyncStatusOfScope_ByCluster(t *testing.T) {
	mocker := setupSQLMock(t)
	clusterID := testClusterID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSyncStatusCluster, testTenantID, &clusterID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, &clusterID, nil, nil))
}

func TestCleanSyncStatusOfScope_ByDepartment(t *testing.T) {
	mocker := setupSQLMock(t)
	departmentID := testDepartmentID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSyncStatusDept, testTenantID, &departmentID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, &departmentID, nil))
}

func TestCleanSyncStatusOfScope_ByProject(t *testing.T) {
	mocker := setupSQLMock(t)
	projectID := testProjectID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSyncStatusProject, testTenantID, &projectID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, &projectID))
}

func TestCleanSyncStatusOfScope_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSyncStatusTenant, testTenantID).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Rollback()

	err := CleanSyncStatusOfScope(ctxWithTenant(), testTenantID, nil, nil, nil)
	require.Error(t, err)
	require.Equal(t, http.StatusServiceUnavailable, int(err.Code))
}
