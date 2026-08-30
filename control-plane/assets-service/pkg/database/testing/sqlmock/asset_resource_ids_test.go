package sqlmock

import (
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/stretchr/testify/require"
)

const ariUpsert = `INSERT INTO assets_resources_ids (asset_id,resource_id,tenant_id,last_synced_at) VALUES (?,?,?,?) ON CONFLICT (asset_id,resource_id,tenant_id) DO UPDATE SET last_synced_at=excluded.last_synced_at`

func TestUpsertAssetResourceIds(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectExec(ariUpsert).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.UpsertAssetResourceIds(auth.MockAuthInfo(), []dbmodel.AssetResourceIdsRecord{{
		AssetId:      "a-1",
		ResourceId:   "r-1",
		LastSyncedAt: time.Now(),
	}})
	require.Nil(t, apiErr)
}
