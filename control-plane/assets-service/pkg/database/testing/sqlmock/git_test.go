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
	gitCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_git.*`
	gitJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_git on assets_metadata.id = assets_git.asset_id`
	gitScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	gitGet      = gitCols + ` ` + gitJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + gitScope + ` AND assets_metadata.deleted_at IS NULL`
	gitList     = gitCols + ` ` + gitJoins + ` WHERE assets_metadata.kind = ? AND ` + gitScope + ` AND assets_metadata.deleted_at IS NULL`
	gitByPwd    = gitCols + ` ` + gitJoins + ` WHERE assets_metadata.kind = ? AND ` + gitScope + ` AND ((assets_git.password_asset_id = ? OR assets_git.credential_id = ?)) AND assets_metadata.deleted_at IS NULL ORDER BY assets_git.credential_id`
	gitByAnyPwd = gitCols + ` ` + gitJoins + ` LEFT JOIN assets_metadata AS am ON assets_git.credential_id = am.id WHERE assets_metadata.kind = ? AND ` + gitScope + ` AND ((assets_git.password_asset_id IS NOT NULL OR (assets_git.credential_id IS NOT NULL AND am.kind = ?))) AND assets_metadata.deleted_at IS NULL ORDER BY assets_git.credential_id`
	gitByGS     = gitCols + ` ` + gitJoins + ` WHERE assets_metadata.kind = ? AND ` + gitScope + ` AND (assets_git.credential_id = ?) AND assets_metadata.deleted_at IS NULL ORDER BY assets_git.credential_id`

	gitInsert = `INSERT INTO assets_git (asset_id,repository,branch,revision,path,password_asset_id,credential_id) VALUES (?,?,?,?,?,?,?)`
	gitDel    = `DELETE FROM assets_git WHERE assets_git.asset_id = ?`
)

func gitRow() []dbmodel.GitCombinedRec { return []dbmodel.GitCombinedRec{{}} }

func TestGetGitAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gitGet, "g-1", "git", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gitRow())
	got, apiErr := dao.GetGitAssetById(auth.MockAuthInfo(), "g-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListGitAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gitList, "git", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gitRow())
	got, apiErr := dao.ListGitAssets(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListGitUsingPassword_Specific(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gitByPwd, "git", dbtesting.AnyArg(), dbtesting.AnyArg(), "p", "p").Returns(gitRow())
	got, apiErr := dao.ListGitUsingPassword(auth.MockAuthInfo(), strptr("p"))
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListGitUsingPassword_Any(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gitByAnyPwd, "git", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(gitRow())
	got, apiErr := dao.ListGitUsingPassword(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListGitUsingGenericSecret_Specific(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(gitByGS, "git", dbtesting.AnyArg(), dbtesting.AnyArg(), "s").Returns(gitRow())
	got, apiErr := dao.ListGitUsingGenericSecret(auth.MockAuthInfo(), strptr("s"))
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateGitAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "x", "git")
	m.ExpectExec(gitInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateGitAsset(auth.MockAuthInfo(), api.GitCreationRequest{Meta: api.AssetCreationRequest{Name: "x", Scope: api.SCOPE_PROJECT}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteGitAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(gitDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteGitAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_GIT}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteGitAsset_SoftUnsupported(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "git")
	m.ExpectTransaction().Rollback()
	apiErr := dao.DeleteGitAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "git"}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

const (
	gitSpecFirst = `SELECT assets_git.repository,assets_git.branch,assets_git.revision,assets_git.path,assets_git.password_asset_id,assets_git.credential_id FROM assets_git WHERE assets_git.asset_id = ? ORDER BY assets_git.asset_id LIMIT ?`
	gitSpecSave  = `UPDATE assets_git SET repository=?,branch=?,revision=?,path=?,password_asset_id=?,credential_id=? WHERE assets_git.asset_id = ?`
)

func TestUpdateGitAsset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(gitSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.GitAssetRec{{}})
	m.ExpectExec(gitSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateGitAsset(auth.MockAuthInfo(), "g-1", api.GitUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
