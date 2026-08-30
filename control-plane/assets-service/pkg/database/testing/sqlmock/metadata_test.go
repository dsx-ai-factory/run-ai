package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/stretchr/testify/require"
)

const (
	metaLeftJoinWST = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id`
	metaScope       = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	sqlMetaById = `SELECT assets_metadata.*, assets_supported_workload_types.* ` + metaLeftJoinWST +
		` WHERE ` + metaScope +
		` AND (COALESCE(assets_metadata.private, FALSE) = false OR assets_metadata.created_by = ?) AND assets_metadata.id = ?`

	sqlListAssetMeta = `SELECT * FROM assets_metadata WHERE assets_metadata.kind = ? AND ` + metaScope +
		` AND assets_metadata.deleted_at IS NULL`

	sqlListMetaByIds = `SELECT assets_metadata.*, assets_supported_workload_types.* ` + metaLeftJoinWST +
		` WHERE (assets_metadata.auto_delete is null or assets_metadata.auto_delete = ?)` +
		` AND COALESCE(assets_metadata.private, FALSE) = false AND ` + metaScope +
		` AND id IN (?) AND assets_metadata.deleted_at IS NULL`

	sqlCountAssetMeta = `SELECT count(*) as count FROM assets_metadata WHERE assets_metadata.kind = ? AND ` + metaScope +
		` AND assets_metadata.deleted_at IS NULL`

	sqlFindPVCEventResourceId = `SELECT asset_id FROM assets_resources_ids WHERE assets_resources_ids.resource_id = ? ORDER BY assets_resources_ids.asset_id LIMIT ?`
	sqlFindPVCEventMeta       = `SELECT * FROM assets_metadata WHERE assets_metadata.id = ? AND assets_metadata.deleted_at IS NULL AND assets_metadata.id = ? ORDER BY assets_metadata.id LIMIT ?`

	sqlNameAvailability = `SELECT count(*) FROM assets_metadata WHERE assets_metadata.name = ? AND assets_metadata.deleted_at IS NULL`
)

func TestGetAssetMetaById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlMetaById).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "a-1"}}})
	got, apiErr := dao.GetAssetMetaById(auth.MockAuthInfo(), "a-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListAssetMeta(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlListAssetMeta).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "a-1", Kind: api.ASSETKIND_PVC}}})
	got, apiErr := dao.ListAssetMeta(auth.MockAuthInfo(), dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: api.ASSETKIND_PVC}})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListAssetsMetaByIds(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlListMetaByIds).Returns([]dbmodel.CombinedRec{{AssetMeta: api.AssetMeta{Id: "a-1"}}})
	got, apiErr := dao.ListAssetsMetaByIds(auth.MockAuthInfo(), []string{"a-1"}, false, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestCountAssetMeta(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlCountAssetMeta).Returns(zero())
	kind := api.ASSETKIND_PVC
	db, apiErr := dao.CountAssetMeta(auth.MockAuthInfo(), &kind, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, db)
	var count int64
	require.NoError(t, db.Count(&count))
}

func TestFindAssetMetaOfPVCEvent(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlFindPVCEventResourceId).Returns([]struct {
		AssetId string
	}{{AssetId: "a-1"}})
	m.ExpectQuery(sqlFindPVCEventMeta).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "a-1"}}})
	got, apiErr := dao.FindAssetMetaOfPVCEvent(auth.MockAuthInfo(), &api.EventSyncFields{
		Id:             "e-1",
		InvolvedObject: api.InvolvedObject{Uid: "uid-1"},
	})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestAssetNameAvailability(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlNameAvailability).Returns(zero())
	q := database.NewParametrizedQuery()
	q.Add("assets_metadata.name = ?", "some-name")
	ok, apiErr := dao.AssetNameAvailability(auth.MockAuthInfo(), api.ASSETKIND_PVC, q)
	require.Nil(t, apiErr)
	require.True(t, ok)
}
