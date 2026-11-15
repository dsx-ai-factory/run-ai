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
	computeCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_compute.*`
	computeJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_compute on assets_metadata.id = assets_compute.asset_id`
	computeScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	computeGet  = computeCols + ` ` + computeJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + computeScope + ` AND assets_metadata.deleted_at IS NULL`
	computeList = computeCols + ` ` + computeJoins + ` WHERE assets_metadata.kind = ? AND ` + computeScope + ` AND assets_metadata.deleted_at IS NULL`

	computeInsert  = `INSERT INTO assets_compute (asset_id,cpu_core_limit,cpu_core_request,cpu_memory_limit,cpu_memory_request,extended_resources,gpu_compute_mode,gpu_devices_request,gpu_memory_limit,gpu_memory_request,gpu_portion_limit,gpu_portion_request,gpu_request_type,large_shm_request) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	computeHardDel = `DELETE FROM assets_compute WHERE assets_compute.asset_id = ?`
)

func computeRow() []dbmodel.ComputeFullRec { return []dbmodel.ComputeFullRec{{}} }

func TestGetComputeAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(computeGet, "c-1", "compute", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(computeRow())
	got, apiErr := dao.GetComputeAssetById(auth.MockAuthInfo(), "c-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListComputeAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(computeList, "compute", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(computeRow())
	got, apiErr := dao.ListComputeAssets(auth.MockAuthInfo(), nil)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateComputeAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-compute", "compute")
	m.ExpectExec(computeInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateComputeAsset(auth.MockAuthInfo(), api.ComputeCreationRequest{Meta: api.AssetCreationRequest{Name: "my-compute", Scope: api.SCOPE_PROJECT}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteComputeAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(computeHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteComputeAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_COMPUTE}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteComputeAsset_SoftUnsupported(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "compute")
	m.ExpectTransaction().Rollback()
	apiErr := dao.DeleteComputeAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: "compute"}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

const (
	computeSpecFirst = `SELECT assets_compute.cpu_core_limit,assets_compute.cpu_core_request,assets_compute.cpu_memory_limit,assets_compute.cpu_memory_request,assets_compute.extended_resources,assets_compute.gpu_compute_mode,assets_compute.gpu_devices_request,assets_compute.gpu_memory_limit,assets_compute.gpu_memory_request,assets_compute.gpu_portion_limit,assets_compute.gpu_portion_request,assets_compute.gpu_request_type,assets_compute.large_shm_request FROM assets_compute WHERE assets_compute.asset_id = ? ORDER BY assets_compute.asset_id LIMIT ?`
	computeSpecSave  = `UPDATE assets_compute SET cpu_core_limit=?,cpu_core_request=?,cpu_memory_limit=?,cpu_memory_request=?,extended_resources=?,gpu_compute_mode=?,gpu_devices_request=?,gpu_memory_limit=?,gpu_memory_request=?,gpu_portion_limit=?,gpu_portion_request=?,gpu_request_type=?,large_shm_request=? WHERE assets_compute.asset_id = ?`
)

func TestUpdateComputeAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(computeSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.ComputeAssetRec{{}})
	m.ExpectExec(computeSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateComputeAsset(auth.MockAuthInfo(), "c-1", api.ComputeUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
