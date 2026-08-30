package dao

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/stretchr/testify/require"
)

// idRow is a minimal row shape carrying just the primary key, used to feed
// SELECT results for queries that scan into Datavolume / DatavolumeView without
// pulling in the SharedScopes association column.
type idRow struct {
	ID uuid.UUID
}

func strPtr(s string) *string { return &s }

func fullAncestors() *authorization_utils.Ancestors {
	return &authorization_utils.Ancestors{
		TenantId:     "1",
		ClusterUuid:  strPtr("cluster-uuid-1"),
		DepartmentId: strPtr("dep-1"),
		ProjectId:    strPtr("proj-1"),
	}
}

const (
	dvViewSelect = "SELECT DISTINCT id,created_at,updated_at,created_by,updated_by,name,description," +
		"project_id,origin_pvc_name,project_name,department_id,tenant_id,cluster_id,phase,phase_message," +
		"conditions,datavolume_pvc_name,datavolume_pv_name,should_delete_original_volume,should_use_original_volume " +
		"FROM datavolumes_view"

	dvInsert = "INSERT INTO datavolumes (id,created_at,updated_at,created_by,updated_by,name,description," +
		"project_id,origin_pvc_name,should_delete_original_volume,should_use_original_volume,scope,project_name," +
		"department_name,department_id,tenant_id,cluster_id,internal,phase,phase_message,conditions," +
		"datavolume_pvc_name,datavolume_pv_name) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	dvClusterSelect = "SELECT datavolumes.id,datavolumes.created_at,datavolumes.updated_at,datavolumes.created_by," +
		"datavolumes.updated_by,datavolumes.name,datavolumes.description,datavolumes.project_id," +
		"datavolumes.origin_pvc_name,datavolumes.should_delete_original_volume,datavolumes.should_use_original_volume," +
		"datavolumes.scope,datavolumes.project_name,datavolumes.department_name,datavolumes.department_id," +
		"datavolumes.tenant_id,datavolumes.cluster_id,datavolumes.internal,datavolumes.phase FROM datavolumes"

	dvByIdSelect = "SELECT * FROM datavolumes WHERE tenant_id = ? AND datavolumes.id = ? ORDER BY datavolumes.id LIMIT ?"
	dvMetaSelect = "SELECT phase,project_id,department_id,cluster_id,tenant_id FROM datavolumes " +
		"WHERE tenant_id = ? AND datavolumes.id = ? ORDER BY datavolumes.id LIMIT ?"

	ssInsert = "INSERT INTO shared_scopes (datavolume_id,scope_type,scope_id,scope_name,tenant_id," +
		"created_at,created_by,project_id,department_id) VALUES (?,?,?,?,?,?,?,?,?)"
	ssDelete = "DELETE FROM shared_scopes WHERE tenant_id = ? AND datavolume_id = ? AND scope_id = ? AND scope_type = ?"

	syncUpdate = "UPDATE datavolumes SET conditions=?,datavolume_pv_name=?,datavolume_pvc_name=?,phase=?,phase_message=? " +
		"WHERE tenant_id = ? AND cluster_id = ? AND id = ? AND phase != ?"
	syncDelete                = "DELETE FROM datavolumes WHERE tenant_id = ? AND cluster_id = ? AND phase IN (?,?)"
	syncDeleteExcluding       = "DELETE FROM datavolumes WHERE tenant_id = ? AND cluster_id = ? AND phase IN (?,?) AND NOT id IN (?)"
	syncDeleteFromToExcluding = "DELETE FROM datavolumes WHERE tenant_id = ? AND cluster_id = ? AND phase IN (?,?) " +
		"AND id >= ? AND id <= ? AND NOT id IN (?)"

	countSelect = "SELECT COUNT(DISTINCT CONCAT(id, ', ', created_at, ', ', updated_at, ', ', created_by, ', ', " +
		"updated_by, ', ', name, ', ', description, ', ', project_id, ', ', origin_pvc_name, ', ', project_name, ', ', " +
		"department_id, ', ', tenant_id, ', ', cluster_id, ', ', phase, ', ', phase_message, ', ', conditions, ', ', " +
		"datavolume_pvc_name, ', ', datavolume_pv_name, ', ', should_delete_original_volume, ', ', " +
		"should_use_original_volume)) FROM datavolumes_view"

	pluckSelect          = "SELECT origin_pvc_name FROM datavolumes WHERE datavolumes.project_id = ? AND datavolumes.tenant_id = ?"
	pluckSelectWithScope = "SELECT origin_pvc_name FROM datavolumes WHERE (datavolumes.project_id = ? AND datavolumes.tenant_id = ?) AND ((false OR project_id IN ( 'proj-1')))"

	dvDelete = "DELETE FROM datavolumes WHERE tenant_id = ? AND phase IN (?,?) AND datavolumes.id = ?"

	updateStatusNotDeleting = "UPDATE datavolumes SET conditions=?,datavolume_pv_name=?,datavolume_pvc_name=?,phase=?," +
		"phase_message=? WHERE tenant_id = ? AND id = ? AND phase != ?"
	updateStatusDeleting = "UPDATE datavolumes SET conditions=?,datavolume_pv_name=?,datavolume_pvc_name=?,phase=?," +
		"phase_message=? WHERE tenant_id = ? AND id = ? AND phase = ?"

	updatePhaseToDeleting = "UPDATE datavolumes SET phase=?,phase_message=?,updated_by=?,updated_at=? " +
		"WHERE tenant_id = ? AND id = ?"

	patchUpdate = "UPDATE datavolumes SET description=?,updated_by=?,updated_at=? WHERE tenant_id = ? AND id = ? RETURNING *"
)

// ---------- DbGetDatavolumes ----------

func TestDbGetDatavolumes_Success(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? ORDER BY id").Returns([]idRow{{ID: uuid.New()}})
	res, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, nil, 0, 0, "", "", nil, "", true)
	require.Nil(t, apiErr)
	require.Len(t, res, 1)
}

func TestDbGetDatavolumes_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? ORDER BY id").ReturnsError(errors.New("boom"))
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, nil, 0, 0, "", "", nil, "", true)
	require.NotNil(t, apiErr)
}

func TestDbGetDatavolumes_ExcludeInternal(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? AND internal != true ORDER BY id").Returns([]idRow{})
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, nil, 0, 0, "", "", nil, "", false)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumes_Search(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? AND (search_vector @@ to_tsquery('english', ? || ':*')) ORDER BY id").Returns([]idRow{})
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, nil, 0, 0, "", "", nil, "myvol", true)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumes_SortByOffsetLimit(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? ORDER BY name DESC NULLS LAST,id LIMIT ? OFFSET ?").Returns([]idRow{})
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, nil, 5, 10, "name", "desc", nil, "", true)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumes_UsableInProjectIdAncestors(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? AND ((project_id = 'proj-1' OR(scope_type = 'project' AND scope_id = 'proj-1') OR(scope_type = 'department' AND scope_id = 'dep-1') OR(scope_type = 'cluster' AND scope_id = 'cluster-uuid-1'))) ORDER BY id").Returns([]idRow{})
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, fullAncestors(), 0, 0, "", "", nil, "", true)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumes_UsableInProjectIdAncestors_Invalid(t *testing.T) {
	setupDaoMock(t)
	bad := &authorization_utils.Ancestors{TenantId: "1"} // missing project/dep/cluster
	_, apiErr := DbGetDatavolumes(testCtx("1", nil), 1, nil, bad, 0, 0, "", "", nil, "", true)
	require.NotNil(t, apiErr)
}

func TestDbGetDatavolumes_PermittedScopesFilter(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvViewSelect + " WHERE tenant_id = ? AND ((false OR project_id IN ( 'proj-1'))) ORDER BY id").Returns([]idRow{})
	ps := &authorization_api.PermittedScopes{Projects: []string{"proj-1"}}
	_, apiErr := DbGetDatavolumes(testCtx("1", ps), 1, ps, nil, 0, 0, "", "", nil, "", true)
	require.Nil(t, apiErr)
}

// ---------- DbGetDatavolumesInClusterWithPartialStatus ----------

func TestDbGetDatavolumesInClusterWithPartialStatus_Success(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvClusterSelect + " WHERE tenant_id = ? AND cluster_id = ? LIMIT ? OFFSET ?").Returns([]idRow{})
	_, apiErr := DbGetDatavolumesInClusterWithPartialStatus(testCtx("1", nil), 1, uuid.New(), 5, 10)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumesInClusterWithPartialStatus_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvClusterSelect + " WHERE tenant_id = ? AND cluster_id = ?").ReturnsError(errors.New("boom"))
	_, apiErr := DbGetDatavolumesInClusterWithPartialStatus(testCtx("1", nil), 1, uuid.New(), 0, 0)
	require.NotNil(t, apiErr)
}

// ---------- DbGetDatavolumeById / Meta ----------

func TestDbGetDatavolumeById_Success(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvByIdSelect).Returns([]idRow{{ID: uuid.New()}})
	_, apiErr := DbGetDatavolumeById(testCtx("1", nil), 1, uuid.New(), false)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumeById_Error(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvByIdSelect).ReturnsError(errors.New("boom"))
	_, apiErr := DbGetDatavolumeById(testCtx("1", nil), 1, uuid.New(), false)
	require.NotNil(t, apiErr)
}

func TestDbGetDatavolumeById_WithSharedScopes(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvByIdSelect).Returns([]idRow{{ID: uuid.New()}})
	m.ExpectQuery(`SELECT * FROM shared_scopes WHERE shared_scopes.datavolume_id = ?`).Returns([]models.SharedScope{})
	_, apiErr := DbGetDatavolumeById(testCtx("1", nil), 1, uuid.New(), true)
	require.Nil(t, apiErr)
}

func TestDbGetDatavolumeMeta_Success(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(dvMetaSelect).Returns([]idRow{{ID: uuid.New()}})
	_, apiErr := DbGetDatavolumeMeta(testCtx("1", nil), 1, uuid.New())
	require.Nil(t, apiErr)
}

// ---------- DbCreateDatavolume ----------

func TestDbCreateDatavolume_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).Returns(1)
	tx.Commit()
	_, apiErr := DbCreateDatavolume(testCtx("1", nil), models.Datavolume{Name: "dv1"})
	require.Nil(t, apiErr)
}

func TestDbCreateDatavolume_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).ReturnsError(errors.New("boom"))
	tx.Rollback()
	_, apiErr := DbCreateDatavolume(testCtx("1", nil), models.Datavolume{Name: "dv1"})
	require.NotNil(t, apiErr)
}

// ---------- DbCreateAndShareDatavolume ----------

func TestDbCreateAndShareDatavolume_Department_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).Returns(1)
	m.ExpectExec(ssInsert).Returns(1)
	tx.Commit()
	dv := models.Datavolume{Name: "dv1", Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: "dep-1", DepartmentName: "dep"}
	_, apiErr := DbCreateAndShareDatavolume(testCtx("1", nil), 1, dv, "")
	require.Nil(t, apiErr)
}

func TestDbCreateAndShareDatavolume_Cluster_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).Returns(1)
	m.ExpectExec(ssInsert).Returns(1)
	tx.Commit()
	dv := models.Datavolume{Name: "dv1", Scope: api.SCOPETYPE_CLUSTER, ClusterId: uuid.New()}
	_, apiErr := DbCreateAndShareDatavolume(testCtx("1", nil), 1, dv, "cluster-name")
	require.Nil(t, apiErr)
}

func TestDbCreateAndShareDatavolume_CreateError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).ReturnsError(errors.New("boom"))
	tx.Rollback()
	dv := models.Datavolume{Name: "dv1", Scope: api.SCOPETYPE_DEPARTMENT}
	_, apiErr := DbCreateAndShareDatavolume(testCtx("1", nil), 1, dv, "")
	require.NotNil(t, apiErr)
}

// TestDbCreateAndShareDatavolume_ShareError documents a bug: on a shared-scope
// insert failure the code returns result.Error (the datavolume create result,
// which is nil) instead of insertResult.Error, so the error is swallowed and the
// transaction is committed rather than rolled back.
func TestDbCreateAndShareDatavolume_ShareError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvInsert).Returns(1)
	m.ExpectExec(ssInsert).ReturnsError(errors.New("boom"))
	tx.Commit()
	dv := models.Datavolume{Name: "dv1", Scope: api.SCOPETYPE_DEPARTMENT}
	_, apiErr := DbCreateAndShareDatavolume(testCtx("1", nil), 1, dv, "")
	require.Nil(t, apiErr)
}

// ---------- DbDeleteDatavolume ----------

func TestDbDeleteDatavolume_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvDelete).Returns(1)
	tx.Commit()
	apiErr := DbDeleteDatavolume(testCtx("1", nil), 1, uuid.New())
	require.Nil(t, apiErr)
}

func TestDbDeleteDatavolume_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(dvDelete).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DbDeleteDatavolume(testCtx("1", nil), 1, uuid.New())
	require.NotNil(t, apiErr)
}

// ---------- DbUpdateStatus ----------

func TestDbUpdateStatus_NotDeleting_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(updateStatusNotDeleting).Returns(1)
	tx.Commit()
	apiErr := DbUpdateStatus(testCtx("1", nil), 1, uuid.New(), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY})
	require.Nil(t, apiErr)
}

func TestDbUpdateStatus_Deleting_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(updateStatusDeleting).Returns(1)
	tx.Commit()
	apiErr := DbUpdateStatus(testCtx("1", nil), 1, uuid.New(), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_DELETING})
	require.Nil(t, apiErr)
}

func TestDbUpdateStatus_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(updateStatusNotDeleting).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DbUpdateStatus(testCtx("1", nil), 1, uuid.New(), models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY})
	require.NotNil(t, apiErr)
}

// ---------- DbUpdatePhaseToDeleting ----------

func TestDbUpdatePhaseToDeleting_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(updatePhaseToDeleting).Returns(1)
	tx.Commit()
	apiErr := DbUpdatePhaseToDeleting(testCtx("1", nil), 1, uuid.New(), "subject")
	require.Nil(t, apiErr)
}

func TestDbUpdatePhaseToDeleting_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(updatePhaseToDeleting).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DbUpdatePhaseToDeleting(testCtx("1", nil), 1, uuid.New(), "subject")
	require.NotNil(t, apiErr)
}

// ---------- DatavolumeNameAvailability ----------

func TestDatavolumeNameAvailability_Available(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(`SELECT count(*) FROM datavolumes WHERE name = ? AND cluster_id = ?`).Returns([]countRow{{Count: 0}})
	available, apiErr := DatavolumeNameAvailability(testCtx("1", nil), "dv1", "cluster-1")
	require.Nil(t, apiErr)
	require.True(t, available)
}

func TestDatavolumeNameAvailability_Conflict(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(`SELECT count(*) FROM datavolumes WHERE name = ? AND cluster_id = ?`).Returns([]countRow{{Count: 2}})
	available, apiErr := DatavolumeNameAvailability(testCtx("1", nil), "dv1", "cluster-1")
	require.Nil(t, apiErr)
	require.False(t, available)
}

func TestDatavolumeNameAvailability_Error(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(`SELECT count(*) FROM datavolumes WHERE name = ? AND cluster_id = ?`).ReturnsError(errors.New("boom"))
	_, apiErr := DatavolumeNameAvailability(testCtx("1", nil), "dv1", "cluster-1")
	require.NotNil(t, apiErr)
}

// ---------- GetUsedPvcsInPermittedScopes ----------

func TestGetUsedPvcsInPermittedScopes_Success(t *testing.T) {
	m := setupDaoMock(t)
	ps := &authorization_api.PermittedScopes{Projects: []string{"proj-1"}}
	m.ExpectQuery(pluckSelectWithScope).Returns([]struct{ OriginPvcName string }{})
	_, apiErr := GetUsedPvcsInPermittedScopes(testCtx("1", ps), "proj-1")
	require.Nil(t, apiErr)
}

func TestGetUsedPvcsInPermittedScopes_NoPermittedScopes(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pluckSelect).Returns([]struct{ OriginPvcName string }{})
	_, apiErr := GetUsedPvcsInPermittedScopes(testCtx("1", nil), "proj-1")
	require.Nil(t, apiErr)
}

func TestGetUsedPvcsInPermittedScopes_Error(t *testing.T) {
	m := setupDaoMock(t)
	ps := &authorization_api.PermittedScopes{Projects: []string{"proj-1"}}
	m.ExpectQuery(pluckSelectWithScope).ReturnsError(errors.New("boom"))
	_, apiErr := GetUsedPvcsInPermittedScopes(testCtx("1", ps), "proj-1")
	require.NotNil(t, apiErr)
}

// ---------- DbSyncDatavolumeStatuses ----------

func syncDv() []models.Datavolume {
	return []models.Datavolume{{BaseModel: models.BaseModel{ID: uuid.New()}, DatavolumeStatus: models.DatavolumeStatus{Phase: api.DATAVOLUMEPHASE_READY}}}
}

func TestDbSyncDatavolumeStatuses_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(syncUpdate).Returns(1)
	m.ExpectExec(syncDeleteExcluding).Returns(1)
	tx.Commit()
	apiErr := DbSyncDatavolumeStatuses(testCtx("1", nil), 1, "cluster-1", syncDv(), nil, nil)
	require.Nil(t, apiErr)
}

func TestDbSyncDatavolumeStatuses_WithFromToExcluding(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(syncUpdate).Returns(1)
	m.ExpectExec(syncDeleteFromToExcluding).Returns(1)
	tx.Commit()
	from := uuid.New()
	to := uuid.New()
	apiErr := DbSyncDatavolumeStatuses(testCtx("1", nil), 1, "cluster-1", syncDv(), &from, &to)
	require.Nil(t, apiErr)
}

func TestDbSyncDatavolumeStatuses_UpdateError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(syncUpdate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DbSyncDatavolumeStatuses(testCtx("1", nil), 1, "cluster-1", syncDv(), nil, nil)
	require.NotNil(t, apiErr)
}

func TestDbSyncDatavolumeStatuses_DeleteError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(syncDelete).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DbSyncDatavolumeStatuses(testCtx("1", nil), 1, "cluster-1", nil, nil, nil)
	require.NotNil(t, apiErr)
}

// ---------- DbUpsertDatavolumeSharedScopes ----------

func TestDbUpsertDatavolumeSharedScopes_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(ssDelete).Returns(1)
	m.ExpectExec(ssInsert).Returns(1)
	tx.Commit()
	toDelete := []models.SharedScope{{ScopeId: "s1", ScopeType: api.SCOPETYPE_PROJECT}}
	toCreate := []models.SharedScope{{ScopeId: "s2", ScopeType: api.SCOPETYPE_PROJECT}}
	_, apiErr := DbUpsertDatavolumeSharedScopes(testCtx("1", nil), 1, uuid.New(), toCreate, toDelete)
	require.Nil(t, apiErr)
}

func TestDbUpsertDatavolumeSharedScopes_DeleteError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(ssDelete).ReturnsError(errors.New("boom"))
	tx.Rollback()
	toDelete := []models.SharedScope{{ScopeId: "s1", ScopeType: api.SCOPETYPE_PROJECT}}
	_, apiErr := DbUpsertDatavolumeSharedScopes(testCtx("1", nil), 1, uuid.New(), nil, toDelete)
	require.NotNil(t, apiErr)
}

func TestDbUpsertDatavolumeSharedScopes_CreateError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(ssInsert).ReturnsError(errors.New("boom"))
	tx.Rollback()
	toCreate := []models.SharedScope{{ScopeId: "s2", ScopeType: api.SCOPETYPE_PROJECT}}
	_, apiErr := DbUpsertDatavolumeSharedScopes(testCtx("1", nil), 1, uuid.New(), toCreate, nil)
	require.NotNil(t, apiErr)
}

// ---------- DbPatchDatavolume ----------

func TestDbPatchDatavolume_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(patchUpdate).Returns([]idRow{{ID: uuid.New()}})
	tx.Commit()
	_, apiErr := DbPatchDatavolume(testCtx("1", nil), 1, uuid.New(), "subject", "new desc")
	require.Nil(t, apiErr)
}

func TestDbPatchDatavolume_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(patchUpdate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	_, apiErr := DbPatchDatavolume(testCtx("1", nil), 1, uuid.New(), "subject", "new desc")
	require.NotNil(t, apiErr)
}

// ---------- DbCountDatavolumes ----------

func TestDbCountDatavolumes_Success(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(countSelect + " WHERE tenant_id = ?").Returns([]countRow{{Count: 3}})
	count, apiErr := DbCountDatavolumes(testCtx("1", nil), 1, nil, nil, nil, "")
	require.Nil(t, apiErr)
	require.Equal(t, int64(3), count)
}

func TestDbCountDatavolumes_Search(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(countSelect + " WHERE tenant_id = ? AND (search_vector @@ to_tsquery('english', ? || ':*'))").Returns([]countRow{{Count: 1}})
	_, apiErr := DbCountDatavolumes(testCtx("1", nil), 1, nil, nil, nil, "myvol")
	require.Nil(t, apiErr)
}

func TestDbCountDatavolumes_UsableInProjectIdAncestors(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(countSelect + " WHERE tenant_id = ? AND ((project_id = 'proj-1' OR(scope_type = 'project' AND scope_id = 'proj-1') OR(scope_type = 'department' AND scope_id = 'dep-1') OR(scope_type = 'cluster' AND scope_id = 'cluster-uuid-1')))").Returns([]countRow{{Count: 1}})
	_, apiErr := DbCountDatavolumes(testCtx("1", nil), 1, nil, fullAncestors(), nil, "")
	require.Nil(t, apiErr)
}

func TestDbCountDatavolumes_InvalidAncestors(t *testing.T) {
	setupDaoMock(t)
	bad := &authorization_utils.Ancestors{TenantId: "1"}
	_, apiErr := DbCountDatavolumes(testCtx("1", nil), 1, nil, bad, nil, "")
	require.NotNil(t, apiErr)
}

func TestDbCountDatavolumes_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(countSelect + " WHERE tenant_id = ?").ReturnsError(errors.New("boom"))
	_, apiErr := DbCountDatavolumes(testCtx("1", nil), 1, nil, nil, nil, "")
	require.NotNil(t, apiErr)
}
