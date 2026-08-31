/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	sqlAccessKeyInsert = `INSERT INTO assets_access_key (asset_id,existing_secret_name,access_key_id,secret_access_key,key_id) VALUES (?,?,?,?,?)`

	accessKeyJoins      = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_access_key on assets_metadata.id = assets_access_key.asset_id`
	accessKeySelectCols = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_access_key.*`
	accessKeyScope      = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	sqlGetById          = accessKeySelectCols + ` ` + accessKeyJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + accessKeyScope + ` AND assets_metadata.deleted_at IS NULL`
	sqlGetByIdUnscoped  = accessKeySelectCols + ` ` + accessKeyJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + accessKeyScope
	sqlList             = accessKeySelectCols + ` ` + accessKeyJoins + ` WHERE assets_metadata.kind = ? AND ` + accessKeyScope + ` AND assets_metadata.deleted_at IS NULL`
	sqlListUnscoped     = accessKeySelectCols + ` ` + accessKeyJoins + ` WHERE assets_metadata.kind = ? AND ` + accessKeyScope
)

func newAccessKeyRequest() api.AccessKeyCreationRequest {
	return api.AccessKeyCreationRequest{
		Meta: api.AssetCreationRequest{
			Name:  "my-access-key",
			Scope: api.SCOPE_PROJECT,
		},
		Spec: api.AccessKeyCreationSpec{
			AccessKeyId:     strptr("AKIA123"),
			SecretAccessKey: strptr("secret"),
		},
	}
}

func TestCreateAccessKey(t *testing.T) {
	mocker := setupDaoMock(t)

	mocker.ExpectTransaction().Begin()
	expectAssetMetaCreated(mocker, "my-access-key", "accessKey")
	mocker.ExpectExec(sqlAccessKeyInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), "AKIA123", "secret", int32(1)).Returns(1)
	mocker.ExpectTransaction().Commit()

	got, apiErr := dao.CreateAccessKey(auth.MockAuthInfo(), newAccessKeyRequest(), 1, nil)

	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.Equal(t, "AKIA123", *got.Spec.AccessKeyId)
}

func TestCreateAccessKey_AccessKeyInsertError(t *testing.T) {
	mocker := setupDaoMock(t)

	mocker.ExpectTransaction().Begin()
	expectAssetMetaCreated(mocker, "my-access-key", "accessKey")
	mocker.ExpectExec(sqlAccessKeyInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), "AKIA123", "secret", int32(1)).
		ReturnsError(errors.New("insert access key failed"))
	mocker.ExpectTransaction().Rollback()

	got, apiErr := dao.CreateAccessKey(auth.MockAuthInfo(), newAccessKeyRequest(), 1, nil)

	require.Nil(t, got)
	require.NotNil(t, apiErr)
	require.Equal(t, int32(503), apiErr.Code)
}

func TestGetAccessKeyById(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectQuery(sqlGetById, "asset-1", "accessKey", dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.AccessKeyFullRecord{{}})

	got, apiErr := dao.GetAccessKeyById(auth.MockAuthInfo(), "asset-1", dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetAccessKeyById_IncludeDeleted(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectQuery(sqlGetByIdUnscoped, "asset-1", "accessKey", dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.AccessKeyFullRecord{{}})

	got, apiErr := dao.GetAccessKeyById(auth.MockAuthInfo(), "asset-1", dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListAccessKeys(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectQuery(sqlList, "accessKey", dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.AccessKeyFullRecord{{}})

	got, apiErr := dao.ListAccessKeys(auth.MockAuthInfo(), nil, dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListAccessKeys_IncludeDeleted(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectQuery(sqlListUnscoped, "accessKey", dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.AccessKeyFullRecord{})

	_, apiErr := dao.ListAccessKeys(auth.MockAuthInfo(), nil, dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
}

const (
	sqlCountS3      = `SELECT count(*) FROM assets_s3 WHERE assets_s3.access_key_asset_id = ?`
	sqlCountSecret  = `SELECT count(*) FROM assets_secret_volume WHERE assets_secret_volume.credential_asset_id = ?`
	sqlCountEnv     = `SELECT count(*) FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_environment on assets_metadata.id = assets_environment.asset_id WHERE assets_metadata.kind = ? AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND EXISTS (SELECT 1 FROM jsonb_array_elements(environment_variables) AS elem WHERE elem->'credential'->>'assetId' = ?) AND assets_metadata.deleted_at IS NULL`
	sqlAccessKeyDel = `DELETE FROM assets_access_key WHERE assets_access_key.asset_id = ?`
)

func expectNotInUse(m dbtesting.Mocker) {
	m.ExpectQuery(sqlCountS3, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountSecret, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountEnv, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
}

func TestDeleteAccessKey_Soft(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectTransaction().Begin()
	expectNotInUse(mocker)
	mocker.ExpectQuery(sqlMetaByIdIn, dbtesting.AnyArg()).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ACCESS_KEY}}})
	mocker.ExpectExec(sqlDeleteWST, dbtesting.AnyArg()).Returns(1)
	mocker.ExpectExec(sqlSoftDelBy, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	mocker.ExpectExec(sqlSoftDelAt, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	mocker.ExpectTransaction().Commit()

	apiErr := dao.DeleteAccessKey(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ACCESS_KEY}, dao.SoftDelete)
	require.Nil(t, apiErr)
}

func TestDeleteAccessKey_Hard(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectTransaction().Begin()
	expectNotInUse(mocker)
	mocker.ExpectQuery(sqlMetaByIdIn, dbtesting.AnyArg()).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ACCESS_KEY}}})
	mocker.ExpectExec(sqlDeleteWST, dbtesting.AnyArg()).Returns(1)
	mocker.ExpectExec(`DELETE FROM assets_metadata WHERE id IN (?)`, dbtesting.AnyArg()).Returns(1)
	mocker.ExpectExec(sqlAccessKeyDel, dbtesting.AnyArg()).Returns(1)
	mocker.ExpectTransaction().Commit()

	apiErr := dao.DeleteAccessKey(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ACCESS_KEY}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteAccessKey_InUse(t *testing.T) {
	mocker := setupDaoMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(sqlCountS3, dbtesting.AnyArg()).Returns([]struct{ C int64 }{{1}})
	mocker.ExpectTransaction().Rollback()

	apiErr := dao.DeleteAccessKey(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ACCESS_KEY}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Equal(t, int32(412), apiErr.Code)
}

func TestCreateAccessKey_NameLockInsertError(t *testing.T) {
	mocker := setupDaoMock(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlLockDeleteStale, dbtesting.AnyTime()).Returns(0)
	mocker.ExpectExec(sqlLockInsert, dbtesting.AnyArg(), "my-access-key", "accessKey", int32(0), "test", dbtesting.AnyTime()).
		ReturnsError(errors.New("lock exists"))
	mocker.ExpectTransaction().Rollback()
	mocker.ExpectTransaction().Rollback()

	got, apiErr := dao.CreateAccessKey(auth.MockAuthInfo(), newAccessKeyRequest(), 1, nil)

	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

func TestUpdateAccessKey(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectTransaction().Commit()
	m.ExpectQuery(sqlGetById, dbtesting.AnyArg(), "accessKey", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.AccessKeyFullRecord{{}})
	got, apiErr := dao.UpdateAccessKey(auth.MockAuthInfo(), "a-1", api.AccessKeyUpdateRequest{Meta: api.CredentialUpdateRequest{Description: strptr("d")}}, 1)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
