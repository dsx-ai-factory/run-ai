package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	svCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_secret_volume.*`
	svJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_secret_volume on assets_metadata.id = assets_secret_volume.asset_id`
	svScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	svGet = svCols + ` ` + svJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + svScope + ` AND assets_metadata.deleted_at IS NULL`

	svInsert    = `INSERT INTO assets_secret_volume (asset_id,credential_asset_id,mount_path,default_mode) VALUES (?,?,?,?)`
	svSpecFirst = `SELECT assets_secret_volume.credential_asset_id,assets_secret_volume.mount_path,assets_secret_volume.default_mode FROM assets_secret_volume WHERE assets_secret_volume.asset_id = ? ORDER BY assets_secret_volume.asset_id LIMIT ?`
	svHardDel   = `DELETE FROM assets_secret_volume WHERE assets_secret_volume.asset_id = ?`

	svListByCred     = svCols + ` ` + svJoins + ` WHERE assets_metadata.kind = ? AND ` + svScope + ` AND assets_secret_volume.credential_asset_id = ? AND assets_metadata.deleted_at IS NULL ORDER BY assets_secret_volume.credential_asset_id`
	svTenantSettings = `SELECT * FROM tenant_settings WHERE tenant_settings.key = ?`
)

func svRow() []dbmodel.SecretAssetCombinedRec {
	return []dbmodel.SecretAssetCombinedRec{{
		AssetMeta:       api.AssetMeta{Id: "sv-1"},
		SecretAssetSpec: api.SecretAssetSpec{CredentialAssetId: strptr("cred-1")},
	}}
}

func TestGetSecretVolumeAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(svGet, "sv-1", "secret-volume", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(svRow())
	got, apiErr := dao.NewSecretVolumeDBAccessor().GetSecretVolumeAssetById(auth.MockAuthInfo(), "sv-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestCreateSecretVolumeAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-sv", "secret-volume")
	m.ExpectExec(svInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.NewSecretVolumeDBAccessor().CreateSecretVolumeAsset(auth.MockAuthInfo(), api.SecretAssetCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-sv", Scope: api.SCOPE_PROJECT},
		Spec: api.SecretAssetSpec{CredentialAssetId: strptr("cred-1")},
	}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateSecretVolumeAsset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(svSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.SecretAssetUpdate{{}})
	m.ExpectTransaction().Commit()
	got, apiErr := dao.NewSecretVolumeDBAccessor().UpdateSecretVolumeAsset(auth.MockAuthInfo(), "sv-1",
		api.SecretAssetUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteSecretVolumeAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(svHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.NewSecretVolumeDBAccessor().DeleteSecretVolumeAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_SECRET_VOLUME}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteSecretVolumeAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "secret-volume")
	apiErr := dao.NewSecretVolumeDBAccessor().DeleteSecretVolumeAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_SECRET_VOLUME}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

func TestListSecretVolumeUsingCredentialId(t *testing.T) {
	m := setupDaoMock(t)
	dao.ForceTenantCloneForTesting(0)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(svTenantSettings, dbtesting.AnyArg()).Returns([]dbmodel.TenantSettingsRec{{Value: "5"}})
	m.ExpectTransaction().Rollback()
	m.ExpectQuery(svListByCred, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(svRow())
	got, apiErr := dao.ListSecretVolumeUsingCredentialId(auth.MockAuthInfo(), strptr("cred-1"))
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
