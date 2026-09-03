/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t *testing.T) dbtesting.Mocker {
	mocker := dbtesting.SetSQLMock(t)
	return mocker
}

func installPolicyMock(t *testing.T) {
	policies.SetMock(policies.SimpleMockClientFactory{})
	t.Cleanup(func() { policies.SetMock(policies.RealFactory{}) })
}

func strptr(s string) *string { return &s }

func zero() []struct{ C int64 } { return []struct{ C int64 }{{0}} }

const (
	sqlLockDeleteStale = `DELETE FROM assets_name_locks WHERE created_at < ?`
	sqlLockInsert      = `INSERT INTO assets_name_locks (id,name,kind,tenant_id,created_by,created_at) VALUES (?,?,?,?,?,?)`
	sqlLockDeleteById  = `DELETE FROM assets_name_locks WHERE assets_name_locks.id = ?`
	sqlMetaInsert      = `INSERT INTO assets_metadata (name,scope,cluster_id,department_id,project_id,auto_delete,id,kind,tenant_id,created_by,created_at,updated_by,updated_at,deleted_by,deleted_at,project_name,description,update_count,prev_name,private) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlMetaSelect      = `SELECT * FROM assets_metadata WHERE assets_metadata.id = ? AND assets_metadata.deleted_at IS NULL AND assets_metadata.id = ?`

	sqlMetaByIdIn  = `SELECT * FROM assets_metadata WHERE id IN (?)`
	sqlDeleteWST   = `DELETE FROM assets_supported_workload_types WHERE asset_id IN (?)`
	sqlSoftDelBy   = `UPDATE assets_metadata SET deleted_by=? WHERE tenant_id = ? AND id IN (?)`
	sqlSoftDelAt   = `UPDATE assets_metadata SET deleted_at=? WHERE tenant_id = ? AND id IN (?)`
	sqlMetaHardDel = `DELETE FROM assets_metadata WHERE id IN (?)`
)

func expectAssetMetaCreated(m dbtesting.Mocker, name, kind string) {
	m.ExpectTransaction().Begin()
	m.ExpectExec(sqlLockDeleteStale, dbtesting.AnyTime()).Returns(0)
	m.ExpectExec(sqlLockInsert, dbtesting.AnyArg(), name, kind, int32(0), "test", dbtesting.AnyTime()).Returns(1)
	m.ExpectTransaction().Commit()
	m.ExpectExec(sqlMetaInsert).Returns(1)
	m.ExpectQuery(sqlMetaSelect).Returns([]dbmodel.AssetMetaRec{})
	m.ExpectTransaction().Begin()
	m.ExpectExec(sqlLockDeleteById, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
}

func expectSoftDeleteAssetMeta(m dbtesting.Mocker, kind string) {
	m.ExpectQuery(sqlMetaByIdIn, dbtesting.AnyArg()).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1", Kind: api.AssetKind(kind)}}})
	m.ExpectExec(sqlDeleteWST, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(sqlSoftDelBy, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(sqlSoftDelAt, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
}

func expectSoftDeleteUnsupported(m dbtesting.Mocker, kind string) {
	m.ExpectQuery(sqlMetaByIdIn, dbtesting.AnyArg()).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1", Kind: api.AssetKind(kind)}}})
	m.ExpectExec(sqlDeleteWST, dbtesting.AnyArg()).Returns(1)
}

func expectHardDeleteAssetMeta(m dbtesting.Mocker) {
	m.ExpectQuery(sqlMetaByIdIn, dbtesting.AnyArg()).Returns([]dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1"}}})
	m.ExpectExec(sqlDeleteWST, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(sqlMetaHardDel, dbtesting.AnyArg()).Returns(1)
}

const (
	sqlPolicyComplianceDel = `DELETE FROM policy_compliance WHERE policy_compliance.asset_id = ?`
	sqlWorkloadAssetsDel   = `DELETE FROM workloads_assets WHERE asset_id = ?`
)

func expectDataSourceDeleteTail(m dbtesting.Mocker) {
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(sqlWorkloadAssetsDel, dbtesting.AnyArg()).Returns(1)
}

const (
	sqlUpdMetaSelect = `SELECT * FROM assets_metadata WHERE assets_metadata.id = ? AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND assets_metadata.deleted_at IS NULL`
	sqlUpdMetaSave   = `UPDATE assets_metadata SET name=?,scope=?,cluster_id=?,department_id=?,project_id=?,auto_delete=?,kind=?,tenant_id=?,created_by=?,created_at=?,updated_by=?,updated_at=?,deleted_by=?,deleted_at=?,project_name=?,description=?,update_count=?,prev_name=?,private=? WHERE assets_metadata.deleted_at IS NULL AND id = ?`
)

const updateAssetMetaName = "asset-name"

func expectUpdateAssetMeta(m dbtesting.Mocker) {
	row := []dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "asset-1", Name: updateAssetMetaName}}}
	m.ExpectQuery(sqlUpdMetaSelect).Returns(row)
	m.ExpectExec(sqlUpdMetaSave).Returns(1)
	m.ExpectQuery(sqlUpdMetaSelect).Returns(row)
}
