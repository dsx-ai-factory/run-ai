package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const assetMetaScopeSelect = `SELECT id FROM assets_metadata WHERE tenant_id = ? AND project_id = ? AND kind = ? AND assets_metadata.deleted_at IS NULL ORDER BY id ASC LIMIT ?`

func assetIDBatch(ids ...string) []struct {
	Id string `gorm:"column:id"`
} {
	out := make([]struct {
		Id string `gorm:"column:id"`
	}, 0, len(ids))
	for _, id := range ids {
		out = append(out, struct {
			Id string `gorm:"column:id"`
		}{Id: id})
	}
	return out
}

func TestDeleteAssetsByScope(t *testing.T) {
	m := setupDaoMock(t)

	const kinds = 17
	for i := 0; i < kinds; i++ {
		m.ExpectTransaction().Begin()
		m.ExpectQuery(assetMetaScopeSelect, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(assetIDBatch())
		m.ExpectTransaction().Commit()
	}

	err := dao.DeleteAssetsByScope(auth.MockAuthInfo(), api.SCOPETYPE_PROJECT, "proj-1")
	require.Nil(t, err)
}

func TestValidateClaimNameUniqueness(t *testing.T) {
	m := setupDaoMock(t)

	q := database.NewParametrizedQuery()
	q.Add("assets_metadata.tenant_id = ?", int32(0))
	q.Add("AND assets_pvc.claim_name = ?", "my-claim")

	m.ExpectQuery(`SELECT count(*) FROM assets_pvc INNER JOIN assets_metadata ON assets_metadata.id = assets_pvc.asset_id WHERE assets_metadata.tenant_id = ? AND assets_pvc.claim_name = ?`, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())

	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	err := dao.ValidateClaimNameUniqueness(auth.MockAuthInfo(), tx, "my-claim", q)
	require.Nil(t, err)
}
