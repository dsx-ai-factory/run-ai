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
	regCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_registry.*`
	regJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_registry on assets_metadata.id = assets_registry.asset_id`
	regScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	regGet  = regCols + ` ` + regJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + regScope + ` AND assets_metadata.deleted_at IS NULL`
	regList = regCols + ` ` + regJoins + ` WHERE assets_metadata.kind = ? AND ` + regScope + ` AND assets_metadata.deleted_at IS NULL`

	regInsert    = `INSERT INTO assets_registry (asset_id,url,credential_kind,credential_id) VALUES (?,?,?,?)`
	regHardDel   = `DELETE FROM assets_registry WHERE assets_registry.asset_id = ?`
	regSpecFirst = `SELECT assets_registry.url,assets_registry.credential_kind,assets_registry.credential_id FROM assets_registry WHERE assets_registry.asset_id = ? ORDER BY assets_registry.asset_id LIMIT ?`
)

func regRow() []dbmodel.RegistryFullRecord {
	return []dbmodel.RegistryFullRecord{{AssetMeta: api.AssetMeta{Id: "asset-1", Name: updateAssetMetaName}, CredentialId: "cred-1"}}
}

func regPwdRow() []dbmodel.PasswordFullRecord {
	return []dbmodel.PasswordFullRecord{{PasswordSpec: api.PasswordSpec{User: strptr("u"), Password: strptr("p")}}}
}

func TestGetRegistryById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(regGet, "r-1", "registry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regRow())
	m.ExpectQuery(pwdGetPrivUse, "cred-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regPwdRow())
	got, apiErr := dao.GetRegistryById(auth.MockAuthInfo(), "r-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListRegistries(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(regList, "registry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regRow())
	m.ExpectQuery(pwdGetPrivUse, "cred-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regPwdRow())
	got, apiErr := dao.ListRegistries(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateRegistry(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-reg", "registry")
	expectAssetMetaCreated(m, "my-reg", "password")
	m.ExpectExec(pwdInsert, dbtesting.AnyArg(), true, dbtesting.AnyArg(), "u", "p", int32(1)).Returns(1)
	m.ExpectExec(regInsert, dbtesting.AnyArg(), "url", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateRegistry(auth.MockAuthInfo(), api.RegistryCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-reg", Scope: api.SCOPE_PROJECT},
		Spec: api.RegistrySpec{User: "u", Password: "p", Url: "url"},
	}, 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateRegistry(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(regSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.RegistryUpdate{{CredentialId: "cred-1"}})
	m.ExpectQuery(pwdGetPrivUse, "cred-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regPwdRow())
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateRegistry(auth.MockAuthInfo(), "r-1", api.RegistryUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteRegistry_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(regGet, "r-1", "registry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regRow())
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(regHardDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectQuery(pwdCountGit, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountGit2, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountSecV, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountEnv, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(pwdHardDelKey, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteRegistry(auth.MockAuthInfo(), "r-1", dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteRegistry_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(regGet, "r-1", "registry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(regRow())
	expectSoftDeleteUnsupported(m, "registry")
	apiErr := dao.DeleteRegistry(auth.MockAuthInfo(), "r-1", dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}
