package sqlmock

import (
	"testing"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	evUpsert  = `INSERT INTO events (id,tenant_id,cluster_id,asset_id,scope_type,scope_id,created_at,asset_kind,object_name,object_namespace,object_kind,object_uid,phase,type,message,reason,source) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`
	evDelete  = `DELETE FROM events WHERE tenant_id = ? AND asset_id IN (?)`
	evGet     = `SELECT * FROM events WHERE events.asset_id = ? LIMIT ?`
	evAssetId = "11111111-1111-1111-1111-111111111111"
)

func TestUpsertEvents(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectExec(evUpsert).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.UpsertEvents(auth.MockAuthInfo(), []*dbmodel.Event{{
		ID:      uuid.New(),
		AssetId: uuid.MustParse(evAssetId),
	}})
	require.Nil(t, apiErr)
}

func TestDeleteAssetEvents(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectExec(evDelete, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	apiErr := dao.DeleteAssetEvents(auth.MockAuthInfo(), tx, evAssetId)
	require.Nil(t, apiErr)
}

func TestDeleteAssetEventsByIds(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectExec(evDelete, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	apiErr := dao.DeleteAssetEventsByIds(auth.MockAuthInfo(), tx, []string{evAssetId})
	require.Nil(t, apiErr)
}

func TestDeleteAssetEventsByIds_Empty(t *testing.T) {
	setupDaoMock(t)
	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	apiErr := dao.DeleteAssetEventsByIds(auth.MockAuthInfo(), tx, []string{})
	require.Nil(t, apiErr)
}

func TestGetEventsByAssetId(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(evGet, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.Event{{
		ID:      uuid.New(),
		AssetId: uuid.MustParse(evAssetId),
	}})
	got, apiErr := dao.GetEventsByAssetId(auth.MockAuthInfo(), evAssetId, 0, 10)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
