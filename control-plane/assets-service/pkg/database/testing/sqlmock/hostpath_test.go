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
	hostpathCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_host_path.*`
	hostpathJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_host_path on assets_metadata.id = assets_host_path.asset_id`
	hostpathScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	hostpathGet   = hostpathCols + ` ` + hostpathJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + hostpathScope + ` AND assets_metadata.deleted_at IS NULL`
	hostpathList  = hostpathCols + ` ` + hostpathJoins + ` WHERE assets_metadata.kind = ? AND ` + hostpathScope + ` AND assets_metadata.deleted_at IS NULL`
	hostpathIns   = `INSERT INTO assets_host_path (asset_id,path,read_only,mount_path,mount_propagation) VALUES (?,?,?,?,?)`
	hostpathDel   = `DELETE FROM assets_host_path WHERE assets_host_path.asset_id = ?`
)

func TestGetHostPathById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hostpathGet, "x-1", "hostPath", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.HostPathCombinedRec{{}})
	got, apiErr := dao.GetHostPathById(auth.MockAuthInfo(), "x-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListHostPathAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hostpathList, "hostPath", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.HostPathCombinedRec{{}})
	got, apiErr := dao.ListHostPathAssets(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateHostPathAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "x", "hostPath")
	m.ExpectExec(hostpathIns).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateHostPathAsset(auth.MockAuthInfo(), api.HostPathCreationRequest{Meta: api.AssetCreationRequest{Name: "x", Scope: api.SCOPE_PROJECT}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteHostPathAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(hostpathDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteHostPathAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "hostPath"}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteHostPathAsset_SoftUnsupported(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "hostPath")
	m.ExpectTransaction().Rollback()
	apiErr := dao.DeleteHostPathAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "hostPath"}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

const (
	hpSpecFirst = `SELECT assets_host_path.path,assets_host_path.read_only,assets_host_path.mount_path,assets_host_path.mount_propagation FROM assets_host_path WHERE assets_host_path.asset_id = ? ORDER BY assets_host_path.asset_id LIMIT ?`
	hpSpecSave  = `UPDATE assets_host_path SET path=?,read_only=?,mount_path=?,mount_propagation=? WHERE assets_host_path.asset_id = ?`
)

func TestUpdateHostPath(t *testing.T) {
	m := setupDaoMock(t)
	installPolicyMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(hpSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.HostPathAssetRec{{}})
	m.ExpectExec(hpSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateHostPath(auth.MockAuthInfo(), "h-1", api.HostPathUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
