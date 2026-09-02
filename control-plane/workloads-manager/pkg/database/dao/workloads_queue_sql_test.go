/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"net/http"
	"testing"

	authorization "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
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

func TestDbUpdateWorkloadOptimistic(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	mocker.ExpectTransaction().Begin()
	expectOptimisticUpdate(mocker, record).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount))
}

func TestDbUpdateWorkloadOptimistic_Conflict(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	mocker.ExpectTransaction().Begin()
	expectOptimisticUpdate(mocker, record).Returns(0)
	mocker.ExpectTransaction().Commit()

	err := DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount)
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, int(err.Code))
}

func TestDbUpdateWorkloadOptimistic_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	mocker.ExpectTransaction().Begin()
	expectOptimisticUpdate(mocker, record).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Rollback()

	err := DbUpdateWorkloadOptimistic(ctxWithTenant(), record, record.UpdateCount)
	require.Error(t, err)
	require.Equal(t, http.StatusServiceUnavailable, int(err.Code))
}

func TestDbUpsertWorkload(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlSaveWorkload).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpsertWorkload(ctxWithTenant(), record))
}

func TestDbUpsertWorkload_InsertWithoutID(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	record.ID = ""
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlInsertWorkload).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, DbUpsertWorkload(ctxWithTenant(), record))
	require.NotEmpty(t, record.ID)
}

func TestDbUpsertWorkload_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	record := testQueueRecord()
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlSaveWorkload).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Rollback()

	err := DbUpsertWorkload(ctxWithTenant(), record)
	require.Error(t, err)
	require.Equal(t, http.StatusServiceUnavailable, int(err.Code))
}

func TestDbGetWorkload(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, nil)
	require.Nil(t, err)
	require.Equal(t, testWorkloadID, got.ID)
	require.Equal(t, expected.Name, got.Name)
}

func TestDbGetWorkload_TenantScope(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, tenantScope())
	require.Nil(t, err)
	require.Equal(t, testWorkloadID, got.ID)
}

func TestDbGetWorkload_SystemScope(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, systemScope())
	require.Nil(t, err)
	require.Equal(t, testWorkloadID, got.ID)
}

func TestDbGetWorkload_ClusterDepartmentProjectScope(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlGetWorkloadScoped, testWorkloadID, testTenantID, 1).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, projectClusterDeptScope())
	require.Nil(t, err)
	require.Equal(t, testWorkloadID, got.ID)
}

func TestDbGetWorkload_DeniedScope(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlGetWorkloadDenied, testWorkloadID, testTenantID, 1).
		ReturnsError(errors.New("record not found"))

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, &authorization.PermittedScopes{})
	require.Error(t, err)
	require.Nil(t, got)
}

func TestDbGetWorkload_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlGetWorkloadByID, testWorkloadID, testTenantID, 1).
		ReturnsError(errors.New("boom"))

	got, err := DbGetWorkload(ctxWithTenant(), testWorkloadID, nil)
	require.Error(t, err)
	require.Equal(t, http.StatusServiceUnavailable, int(err.Code))
	require.Nil(t, got)
}

func TestDbListByIdsToMap(t *testing.T) {
	mocker := setupSQLMock(t)
	first := testQueueRecord()
	second := testQueueRecord()
	second.ID = testWorkloadID2
	second.Name = "wl-2"
	mocker.ExpectQuery(sqlListByIDs, testWorkloadID, testWorkloadID2, testTenantID).
		Returns([]models.WorkloadsQueueRecord{*first, *second})

	got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID, testWorkloadID2}, nil)
	require.Nil(t, err)
	require.Equal(t, 2, len(got))
	require.Equal(t, "wl-1", got[testWorkloadID].Name)
	require.Equal(t, "wl-2", got[testWorkloadID2].Name)
}

func TestDbListByIdsToMap_WithPermittedScopes(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlListByIDsScoped, testWorkloadID, testTenantID).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, projectClusterDeptScope())
	require.Nil(t, err)
	require.Equal(t, 1, len(got))
	require.Equal(t, testWorkloadID, got[testWorkloadID].ID)
}

func TestDbListByIdsToMap_Empty(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlListByIDsSingle, testWorkloadID, testTenantID).
		Returns([]models.WorkloadsQueueRecord{})

	got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, nil)
	require.Nil(t, err)
	require.Empty(t, got)
}

func TestDbListByIdsToMap_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlListByIDsSingle, testWorkloadID, testTenantID).
		ReturnsError(errors.New("boom"))

	got, err := DbListByIdsToMap(ctxWithTenant(), []string{testWorkloadID}, nil)
	require.Error(t, err)
	require.Nil(t, got)
}

func TestDbListUpdates(t *testing.T) {
	mocker := setupSQLMock(t)
	expected := testQueueRecord()
	mocker.ExpectQuery(sqlListUpdates, testTenantID, testClusterID, int64(3), 10).
		Returns([]models.WorkloadsQueueRecord{*expected})

	got, err := DbListUpdates(ctxWithTenant(), testClusterID, 3, 10)
	require.Nil(t, err)
	require.Equal(t, 1, len(got))
	require.Equal(t, testWorkloadID, got[0].ID)
}

func TestDbListUpdates_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlListUpdates, testTenantID, testClusterID, int64(0), 5).
		ReturnsError(errors.New("boom"))

	got, err := DbListUpdates(ctxWithTenant(), testClusterID, 0, 5)
	require.Error(t, err)
	require.Nil(t, got)
}

func TestCleanWorkloadsOfScope(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteWorkloadsTenant, testTenantID).Returns(2)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, nil))
}

func TestCleanWorkloadsOfScope_ByCluster(t *testing.T) {
	mocker := setupSQLMock(t)
	clusterID := testClusterID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteWorkloadsCluster, testTenantID, &clusterID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, &clusterID, nil, nil))
}

func TestCleanWorkloadsOfScope_ByDepartment(t *testing.T) {
	mocker := setupSQLMock(t)
	departmentID := testDepartmentID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteWorkloadsDept, testTenantID, &departmentID).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, &departmentID, nil))
}

func TestCleanWorkloadsOfScope_ByProject(t *testing.T) {
	mocker := setupSQLMock(t)
	projectID := testProjectID
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteWorkloadsProject, testTenantID, &projectID).Returns(3)
	mocker.ExpectTransaction().Commit()

	require.Nil(t, CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, &projectID))
}

func TestCleanWorkloadsOfScope_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteWorkloadsTenant, testTenantID).ReturnsError(errors.New("boom"))
	mocker.ExpectTransaction().Rollback()

	require.Error(t, CleanWorkloadsOfScope(ctxWithTenant(), testTenantID, nil, nil, nil))
}
