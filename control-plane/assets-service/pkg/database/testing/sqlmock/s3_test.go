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
	s3Cols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_s3.*`
	s3Joins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_s3 on assets_metadata.id = assets_s3.asset_id`
	s3Scope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	s3Get            = s3Cols + ` ` + s3Joins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + s3Scope + ` AND assets_metadata.deleted_at IS NULL`
	s3List           = s3Cols + ` ` + s3Joins + ` WHERE assets_metadata.kind = ? AND ` + s3Scope + ` AND assets_metadata.deleted_at IS NULL`
	s3ByAccessKey    = s3Cols + ` ` + s3Joins + ` WHERE assets_metadata.kind = ? AND ` + s3Scope + ` AND assets_s3.access_key_asset_id = ? AND assets_metadata.deleted_at IS NULL ORDER BY assets_s3.access_key_asset_id`
	s3ByAnyAccessKey = s3Cols + ` ` + s3Joins + ` WHERE assets_metadata.kind = ? AND ` + s3Scope + ` AND assets_s3.access_key_asset_id IS NOT NULL AND assets_metadata.deleted_at IS NULL ORDER BY assets_s3.access_key_asset_id`

	s3Insert  = `INSERT INTO assets_s3 (asset_id,bucket,path,url,access_key_asset_id) VALUES (?,?,?,?,?)`
	s3HardDel = `DELETE FROM assets_s3 WHERE assets_s3.asset_id = ?`
)

func s3Row() []dbmodel.S3CombinedRec { return []dbmodel.S3CombinedRec{{}} }

func TestGetS3AssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(s3Get, "s-1", "s3", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(s3Row())
	got, apiErr := dao.GetS3AssetById(auth.MockAuthInfo(), "s-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListS3Assets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(s3List, "s3", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(s3Row())
	got, apiErr := dao.ListS3Assets(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListS3UsingAccessKey_Specific(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(s3ByAccessKey, "s3", dbtesting.AnyArg(), dbtesting.AnyArg(), "ak-1").Returns(s3Row())
	got, apiErr := dao.ListS3UsingAccessKey(auth.MockAuthInfo(), strptr("ak-1"))
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListS3UsingAccessKey_Any(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(s3ByAnyAccessKey, "s3", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(s3Row())
	got, apiErr := dao.ListS3UsingAccessKey(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateS3Asset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-s3", "s3")
	m.ExpectExec(s3Insert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateS3Asset(auth.MockAuthInfo(), api.S3CreationRequest{Meta: api.AssetCreationRequest{Name: "my-s3", Scope: api.SCOPE_PROJECT}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteS3Asset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(s3HardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteS3Asset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_S3}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteS3Asset_SoftUnsupported(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "s3")
	m.ExpectTransaction().Rollback()
	apiErr := dao.DeleteS3Asset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "s3"}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

const (
	s3SpecFirst = `SELECT assets_s3.bucket,assets_s3.path,assets_s3.url,assets_s3.access_key_asset_id FROM assets_s3 WHERE assets_s3.asset_id = ? ORDER BY assets_s3.asset_id LIMIT ?`
	s3SpecSave  = `UPDATE assets_s3 SET bucket=?,path=?,url=?,access_key_asset_id=? WHERE assets_s3.asset_id = ?`
)

func TestUpdateS3Asset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(s3SpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.S3AssetRec{{}})
	m.ExpectExec(s3SpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateS3Asset(auth.MockAuthInfo(), "s-1", api.S3UpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
