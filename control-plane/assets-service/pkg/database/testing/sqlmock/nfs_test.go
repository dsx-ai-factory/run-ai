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
	nfsCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_nfs.*`
	nfsJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_nfs on assets_metadata.id = assets_nfs.asset_id`
	nfsScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	nfsGet   = nfsCols + ` ` + nfsJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + nfsScope + ` AND assets_metadata.deleted_at IS NULL`
	nfsList  = nfsCols + ` ` + nfsJoins + ` WHERE assets_metadata.kind = ? AND ` + nfsScope + ` AND assets_metadata.deleted_at IS NULL`
	nfsIns   = `INSERT INTO assets_nfs (asset_id,path,read_only,server,mount_path) VALUES (?,?,?,?,?)`
	nfsDel   = `DELETE FROM assets_nfs WHERE assets_nfs.asset_id = ?`
)

func TestGetNFSAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(nfsGet, "x-1", "nfs", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.NFSCombinedRec{{}})
	got, apiErr := dao.GetNFSAssetById(auth.MockAuthInfo(), "x-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListNFSAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(nfsList, "nfs", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.NFSCombinedRec{{}})
	got, apiErr := dao.ListNFSAssets(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateNFSAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "x", "nfs")
	m.ExpectExec(nfsIns).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateNFSAsset(auth.MockAuthInfo(), api.NfsCreationRequest{Meta: api.AssetCreationRequest{Name: "x", Scope: api.SCOPE_PROJECT}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteNFSAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(nfsDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteNFSAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "nfs"}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteNFSAsset_SoftUnsupported(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "nfs")
	m.ExpectTransaction().Rollback()
	apiErr := dao.DeleteNFSAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "nfs"}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

const (
	nfsSpecFirst = `SELECT assets_nfs.path,assets_nfs.read_only,assets_nfs.server,assets_nfs.mount_path FROM assets_nfs WHERE assets_nfs.asset_id = ? ORDER BY assets_nfs.asset_id LIMIT ?`
	nfsSpecSave  = `UPDATE assets_nfs SET path=?,read_only=?,server=?,mount_path=? WHERE assets_nfs.asset_id = ?`
)

func TestUpdateNFSAsset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(nfsSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.NFSAssetRec{{}})
	m.ExpectExec(nfsSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateNFSAsset(auth.MockAuthInfo(), "n-1", api.NfsUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
