/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

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
	gsCols    = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_generic_secret.*`
	gsJoins   = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_generic_secret on assets_metadata.id = assets_generic_secret.asset_id`
	gsScope   = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	gsPrivate = `(COALESCE(assets_metadata.private, FALSE) = false OR assets_metadata.created_by = ?)`

	gsGet        = gsCols + ` ` + gsJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + gsScope + ` AND ` + gsPrivate + ` AND assets_metadata.deleted_at IS NULL`
	gsGetDeleted = gsCols + ` ` + gsJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + gsScope + ` AND ` + gsPrivate
	gsList       = gsCols + ` ` + gsJoins + ` WHERE assets_metadata.kind = ? AND ` + gsScope + ` AND assets_metadata.deleted_at IS NULL`
	gsListDel    = gsCols + ` ` + gsJoins + ` WHERE assets_metadata.kind = ? AND ` + gsScope

	gsInsert     = `INSERT INTO assets_generic_secret (asset_id,existing_secret_name,key_value_pairs,key_id) VALUES (?,?,?,?)`
	gsCountSecV  = `SELECT count(*) FROM assets_secret_volume WHERE assets_secret_volume.credential_asset_id = ?`
	gsHardDelRow = `DELETE FROM assets_generic_secret WHERE assets_generic_secret.asset_id = ?`
)

func gsRow() []dbmodel.GenericSecretFullRecord { return []dbmodel.GenericSecretFullRecord{{}} }

func TestGetGenericSecretById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gsGet, "g-1", "genericSecret", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gsRow())
	got, apiErr := dao.GetGenericSecretById(auth.MockAuthInfo(), "g-1", dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetGenericSecretById_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gsGetDeleted, "g-1", "genericSecret", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gsRow())
	got, apiErr := dao.GetGenericSecretById(auth.MockAuthInfo(), "g-1", dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListGenericSecrets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gsList, "genericSecret", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gsRow())
	got, apiErr := dao.ListGenericSecrets(auth.MockAuthInfo(), nil, dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListGenericSecrets_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gsListDel, "genericSecret", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gsRow())
	_, apiErr := dao.ListGenericSecrets(auth.MockAuthInfo(), nil, dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
}

func TestCreateGenericSecret(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-gs", "genericSecret")
	m.ExpectExec(gsInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), int32(1)).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateGenericSecret(auth.MockAuthInfo(), api.GenericSecretCreationRequest{Meta: api.AssetCreationRequest{Name: "my-gs", Scope: api.SCOPE_PROJECT}}, 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteGenericSecret_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(gsCountSecV, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountEnv, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountGit, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(gsHardDelRow, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteGenericSecret(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_GENERIC_SECRET}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestUpdateGenericSecret(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectTransaction().Commit()
	m.ExpectQuery(gsGet, dbtesting.AnyArg(), "genericSecret", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gsRow())
	got, apiErr := dao.UpdateGenericSecret(auth.MockAuthInfo(), "g-1", api.GenericSecretUpdateRequest{Meta: api.CredentialUpdateRequest{Description: strptr("d")}}, 1)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteGenericSecret_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(gsCountSecV, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountEnv, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountGit, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	expectSoftDeleteAssetMeta(m, "genericSecret")
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteGenericSecret(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_GENERIC_SECRET}, dao.SoftDelete)
	require.Nil(t, apiErr)
}
