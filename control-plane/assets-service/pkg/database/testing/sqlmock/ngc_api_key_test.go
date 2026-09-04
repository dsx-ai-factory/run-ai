/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	ngcCols    = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_ngc_api_key.*`
	ngcJoins   = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_ngc_api_key on assets_metadata.id = assets_ngc_api_key.asset_id`
	ngcScope   = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	ngcPrivate = `(COALESCE(assets_metadata.private, FALSE) = false OR assets_metadata.created_by = ?)`

	ngcGet        = ngcCols + ` ` + ngcJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + ngcScope + ` AND ` + ngcPrivate + ` AND assets_metadata.deleted_at IS NULL`
	ngcGetDeleted = ngcCols + ` ` + ngcJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + ngcScope + ` AND ` + ngcPrivate
	ngcList       = ngcCols + ` ` + ngcJoins + ` WHERE assets_metadata.kind = ? AND ` + ngcScope + ` AND assets_metadata.deleted_at IS NULL`
	ngcListDel    = ngcCols + ` ` + ngcJoins + ` WHERE assets_metadata.kind = ? AND ` + ngcScope

	ngcInsert  = `INSERT INTO assets_ngc_api_key (asset_id,ngc_api_key,key_id) VALUES (?,?,?)`
	ngcHardDel = `DELETE FROM assets_ngc_api_key WHERE assets_ngc_api_key.asset_id = ?`
)

func ngcRow() []dbmodel.NgcApiKeyFullRecord { return []dbmodel.NgcApiKeyFullRecord{{}} }

func TestGetNgcApiKeyById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(ngcGet, "n-1", "ngcApiKey", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(ngcRow())
	got, apiErr := dao.GetNgcApiKeyById(auth.MockAuthInfo(), "n-1", dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetNgcApiKeyById_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(ngcGetDeleted, "n-1", "ngcApiKey", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(ngcRow())
	got, apiErr := dao.GetNgcApiKeyById(auth.MockAuthInfo(), "n-1", dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListNgcApiKeys(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(ngcList, "ngcApiKey", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(ngcRow())
	got, apiErr := dao.ListNgcApiKeys(auth.MockAuthInfo(), nil, dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListNgcApiKeys_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(ngcListDel, "ngcApiKey", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(ngcRow())
	_, apiErr := dao.ListNgcApiKeys(auth.MockAuthInfo(), nil, dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
}

func TestCreateNgcApiKey(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-ngc", "ngcApiKey")
	m.ExpectExec(ngcInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), int32(1)).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateNgcApiKey(auth.MockAuthInfo(), api.NgcApiKeyCreationRequest{Meta: api.AssetCreationRequest{Name: "my-ngc", Scope: api.SCOPE_PROJECT}}, 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteNgcApiKey_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(ngcHardDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteNgcApiKey(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_NGC_API_KEY}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestUpdateNgcApiKey(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectTransaction().Commit()
	m.ExpectQuery(ngcGet, dbtesting.AnyArg(), "ngcApiKey", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(ngcRow())
	got, apiErr := dao.UpdateNgcApiKey(auth.MockAuthInfo(), "n-1", api.NgcApiKeyUpdateRequest{Meta: api.CredentialUpdateRequest{Description: strptr("d")}}, 1)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteNgcApiKey_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteAssetMeta(m, "ngcApiKey")
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteNgcApiKey(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_NGC_API_KEY}, dao.SoftDelete)
	require.Nil(t, apiErr)
}
