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
	cmCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_config_map.*`
	cmJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_config_map on assets_metadata.id = assets_config_map.asset_id`
	cmScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	cmGet = cmCols + ` ` + cmJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + cmScope + ` AND assets_metadata.deleted_at IS NULL`

	cmInsert    = `INSERT INTO assets_config_map (asset_id,config_map,mount_path,sub_path,default_mode) VALUES (?,?,?,?,?)`
	cmSpecFirst = `SELECT assets_config_map.config_map,assets_config_map.mount_path,assets_config_map.sub_path,assets_config_map.default_mode FROM assets_config_map WHERE assets_config_map.asset_id = ? ORDER BY assets_config_map.asset_id LIMIT ?`
	cmHardDel   = `DELETE FROM assets_config_map WHERE assets_config_map.asset_id = ?`
)

func cmRow() []dbmodel.ConfigMapCombinedRec {
	return []dbmodel.ConfigMapCombinedRec{{
		AssetMeta:          api.AssetMeta{Id: "cm-1"},
		ConfigMapAssetSpec: api.ConfigMapAssetSpec{ConfigMap: strptr("cm")},
	}}
}

func TestGetConfigMapAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(cmGet, "cm-1", "config-map", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(cmRow())
	got, apiErr := dao.ConfigMapsDBAccessor.GetConfigMapAssetById(auth.MockAuthInfo(), "cm-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestCreateConfigMapAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-cm", "config-map")
	m.ExpectExec(cmInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.ConfigMapsDBAccessor.CreateConfigMapAsset(auth.MockAuthInfo(), api.ConfigMapCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-cm", Scope: api.SCOPE_PROJECT},
		Spec: api.ConfigMapAssetSpec{ConfigMap: strptr("cm")},
	}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateConfigMapAsset(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(cmSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.ConfigMapAssetUpdate{{}})
	m.ExpectTransaction().Commit()
	got, apiErr := dao.ConfigMapsDBAccessor.UpdateConfigMapAsset(auth.MockAuthInfo(), "cm-1",
		api.ConfigMapUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteConfigMapAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(cmHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.ConfigMapsDBAccessor.DeleteConfigMapAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_CONFIG_MAP}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteConfigMapAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "config-map")
	apiErr := dao.ConfigMapsDBAccessor.DeleteConfigMapAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_CONFIG_MAP}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}
