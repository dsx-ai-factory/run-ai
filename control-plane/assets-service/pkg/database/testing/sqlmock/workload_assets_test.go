package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	wlaFind     = `SELECT * FROM assets_metadata WHERE id IN (?) AND assets_metadata.deleted_at IS NULL`
	wlaCredFind = `SELECT * FROM assets_metadata WHERE id IN (NULL) AND assets_metadata.deleted_at IS NULL`
	vcaMetaFind = `SELECT assets_metadata.*, assets_supported_workload_types.* FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id WHERE COALESCE(assets_metadata.private, FALSE) = false AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND id IN (?) AND assets_metadata.deleted_at IS NULL`
)

func TestListWorkloadAssets(t *testing.T) {
	m := setupDaoMock(t)
	db := runaidb.DB().WithContext(auth.MockAuthInfo())
	m.ExpectQuery(wlaFind, dbtesting.AnyArg()).Returns([]dbmodel.AssetsContainedInWorkload{})
	m.ExpectQuery(wlaCredFind).Returns([]dbmodel.CredentialsAssetSpecs{})
	assets, creds, apiErr := dao.ListWorkloadAssets([]string{"a-1"}, nil, db)
	require.Nil(t, apiErr)
	require.NotNil(t, assets)
	require.NotNil(t, creds)
}

func TestValidateContainedAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(vcaMetaFind, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.CombinedRec{{
		AssetMeta: api.AssetMeta{Id: "env-1", Kind: api.ASSETKIND_ENVIRONMENT, Scope: api.SCOPE_TENANT},
	}})
	ids, apiErr := dao.ValidateContainedAssets(auth.MockAuthInfo(), nil, api.SCOPE_TENANT,
		nil, nil, nil, &api.AssetsIds{Environment: "env-1"})
	require.Nil(t, apiErr)
	require.Equal(t, []string{"env-1"}, ids)
}
