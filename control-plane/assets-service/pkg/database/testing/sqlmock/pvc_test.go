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
	pvcCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_pvc.*`
	pvcJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_pvc on assets_metadata.id = assets_pvc.asset_id`
	pvcScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	pvcGet        = pvcCols + ` ` + pvcJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pvcScope + ` AND assets_metadata.deleted_at IS NULL`
	pvcGetDeleted = pvcCols + ` ` + pvcJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pvcScope

	pvcInsert   = `INSERT INTO assets_pvc (asset_id,path,claim_info,existing_pvc,claim_name,read_only,ephemeral,data_sharing) VALUES (?,?,?,?,?,?,?,?)`
	pvcSpecSave = `UPDATE assets_pvc SET asset_id=?,path=?,claim_info=?,existing_pvc=?,claim_name=?,read_only=?,ephemeral=?,data_sharing=? WHERE assets_pvc.asset_id = ?`
	pvcHardDel  = `DELETE FROM assets_pvc WHERE assets_pvc.asset_id = ?`
)

func pvcRow() []dbmodel.PVCCombinedRec {
	return []dbmodel.PVCCombinedRec{{
		AssetMeta:    api.AssetMeta{Id: "p-1"},
		PvcAssetSpec: api.PvcAssetSpec{ClaimName: strptr("cn")},
	}}
}

func TestGetPvcAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pvcGet, "p-1", "pvc", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(pvcRow())
	got, apiErr := dao.GetPvcAssetById(auth.MockAuthInfo(), "p-1", dao.DatasourcesOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetPvcAssetById_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pvcGetDeleted, "p-1", "pvc", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(pvcRow())
	got, apiErr := dao.GetPvcAssetById(auth.MockAuthInfo(), "p-1", dao.DatasourcesOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestCreatePvcAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-pvc", "pvc")
	m.ExpectExec(pvcInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreatePvcAsset(auth.MockAuthInfo(), api.PvcCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-pvc", Scope: api.SCOPE_PROJECT},
		Spec: api.PvcAssetSpec{ClaimName: strptr("cn")},
	}, nil, nil, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdatePvcAsset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectExec(pvcSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdatePvcAsset(auth.MockAuthInfo(), "p-1", api.AssetUpdateRequest{Name: updateAssetMetaName}, &dbmodel.PVCAssetRec{AssetId: "p-1"}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeletePvcAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(pvcHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeletePvcAsset(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_PVC}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeletePvcAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteAssetMeta(m, "pvc")
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeletePvcAsset(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_PVC}, dao.SoftDelete)
	require.Nil(t, apiErr)
}
