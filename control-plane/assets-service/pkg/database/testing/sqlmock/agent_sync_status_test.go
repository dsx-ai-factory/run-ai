package sqlmock

import (
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	assTenant  = int32(7)
	assCluster = "cluster-1"
	assAsset   = "asset-1"

	listAssetsForSyncSQL  = `SELECT assets_metadata.* FROM assets_metadata WHERE kind in (?) AND tenant_id = ? AND (cluster_id is null or cluster_id = ?) AND COALESCE(assets_metadata.private, FALSE) = false AND update_count > ? ORDER BY update_count`
	getSyncStatusSQL      = `SELECT * FROM sync_status WHERE sync_status.tenant_id = ? AND sync_status.cluster_id = ? AND sync_status.asset_id = ?`
	highestUpdateCountSQL = `SELECT last_value FROM assets_metadata_update_count`
	listSyncStatusSQL     = `SELECT * FROM sync_status WHERE tenant_id = ? AND asset_id in (?)`
	saveSyncStatusSQL     = `UPDATE sync_status SET tenant_id=?,cluster_id=?,asset_id=?,prev_name=?,spec=?,sent_to_agent_apply=?,sent_to_agent_delete=?,processed_by_agent=?,message=? WHERE id = ?`
)

func syncStatusRow() []dbmodel.SyncStatusRec {
	return []dbmodel.SyncStatusRec{{
		Id:               "ss-1",
		TenantId:         assTenant,
		ClusterId:        assCluster,
		AssetId:          assAsset,
		SentToAgentApply: 5,
	}}
}

func TestListAssetsForSync(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(listAssetsForSyncSQL).Returns([]dbmodel.AssetsToSync{})
	got, apiErr := dao.ListAssetsForSync(auth.MockAuthInfo(), assTenant, assCluster,
		[]api.AssetKind{api.ASSETKIND_PVC}, 5, 50, false)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetSyncStatusByClusterAndAssetId(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(getSyncStatusSQL).Returns(syncStatusRow())
	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	got, apiErr := dao.GetSyncStatusByClusterAndAssetId(auth.MockAuthInfo(), tx, assTenant, assCluster, assAsset)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateSyncStatusSendToAgent(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(getSyncStatusSQL).Returns(syncStatusRow())
	m.ExpectExec(saveSyncStatusSQL).Returns(1)
	m.ExpectTransaction().Commit()

	update := dbmodel.AssetsToSync{}
	update.Id = assAsset
	update.Kind = api.ASSETKIND_PVC
	update.UpdateCount = 9
	update.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	apiErr := dao.UpdateSyncStatusSendToAgent(auth.MockAuthInfo(), assTenant, assCluster,
		[]dbmodel.AssetsToSync{update})
	require.Nil(t, apiErr)
}

func TestHandleAgentSyncUpdates(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(getSyncStatusSQL).Returns(syncStatusRow())
	m.ExpectTransaction().Begin()
	m.ExpectExec(saveSyncStatusSQL).Returns(1)
	m.ExpectTransaction().Commit()

	updates := []api.ClusterSyncReportUpdate{{Id: assAsset, UpdateCount: 5}}
	dao.HandleAgentSyncUpdates(auth.MockAuthInfo(), assTenant, assCluster, updates)
}

func TestGetHighestUpdateCount(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(highestUpdateCountSQL).Returns([]struct{ LastValue int64 }{{5}})
	got, apiErr := dao.GetHighestUpdateCount(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Equal(t, int64(5), got)
}

func TestListSyncStatusByAssetIds(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(listSyncStatusSQL).Returns(syncStatusRow())
	got, apiErr := dao.ListSyncStatusByAssetIds(auth.MockAuthInfo(), assTenant, []string{assAsset})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
