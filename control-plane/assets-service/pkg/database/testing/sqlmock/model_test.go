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
	modelCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_model.*`
	modelJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_model on assets_metadata.id = assets_model.asset_id`
	modelScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	modelGet  = modelCols + ` ` + modelJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + modelScope + ` AND assets_metadata.deleted_at IS NULL`
	modelList = modelCols + ` ` + modelJoins + ` WHERE assets_metadata.kind = ? AND ` + modelScope + ` AND assets_metadata.deleted_at IS NULL`

	modelInsert    = `INSERT INTO assets_model (asset_id,title,description,thumbnail,huggingface_id,compute_asset_id,environment_asset_id,license) VALUES (?,?,?,?,?,?,?,?)`
	modelSpecFirst = `SELECT * FROM assets_model WHERE assets_model.asset_id = ? ORDER BY assets_model.asset_id LIMIT ?`
	modelSpecSave  = `UPDATE assets_model SET asset_id=?,title=?,description=?,thumbnail=?,huggingface_id=?,compute_asset_id=?,environment_asset_id=?,license=? WHERE assets_model.asset_id = ?`
	modelHardDel   = `DELETE FROM assets_model WHERE assets_model.asset_id = ?`
)

func modelRow() []dbmodel.ModelFullRec { return []dbmodel.ModelFullRec{{}} }

func TestGetModelAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(modelGet, "m-1", "model", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(modelRow())
	got, apiErr := dao.GetModelAssetById(auth.MockAuthInfo(), "m-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListModels(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(modelList, "model", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(modelRow())
	got, apiErr := dao.ListModels(auth.MockAuthInfo(), nil, "")
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateModelAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-model", "model")
	m.ExpectExec(modelInsert, dbtesting.AnyArg(), "t", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateModelAsset(auth.MockAuthInfo(), api.ModelCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-model", Scope: api.SCOPE_PROJECT},
		Spec: api.ModelAssetSpec{Title: "t"},
	}, nil, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateModelAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(modelSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.ModelAssetRec{{}})
	m.ExpectExec(modelSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateModelAssetById(auth.MockAuthInfo(), "m-1", api.ModelUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteModelAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(modelHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteModelAssetById(auth.MockAuthInfo(), "m-1", dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteModelAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "model")
	apiErr := dao.DeleteModelAssetById(auth.MockAuthInfo(), "m-1", dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}
