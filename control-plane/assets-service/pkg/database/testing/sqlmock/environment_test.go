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
	envCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_environment.*`
	envJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_environment on assets_metadata.id = assets_environment.asset_id`
	envScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	envGet      = envCols + ` ` + envJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + envScope + ` AND assets_metadata.deleted_at IS NULL`
	envCredList = envCols + ` ` + envJoins + ` WHERE assets_metadata.kind = ? AND ` + envScope + ` AND EXISTS (SELECT 1 FROM jsonb_array_elements(environment_variables) AS elem WHERE elem->'credential'->>'assetId' = ?) AND assets_metadata.deleted_at IS NULL`

	envWSTInsert = `INSERT INTO assets_supported_workload_types (asset_id,workspace,training,inference,distributed,dist_framework) VALUES (?,?,?,?,?,?)`
	envInsert    = `INSERT INTO assets_environment (asset_id,args,command,environment_variables,run_as_gid,run_as_uid,supplemental_groups,allow_privilege_escalation,capabilities,create_home_dir,host_ipc,host_network,image,image_pull_policy,probes,read_only_root_filesystem,run_as_non_root,seccomp_profile_type,stdin,tty,uid_gid_source,working_dir,connections,override_uid_gid_in_workspace,routing_method) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	envHardDel   = `DELETE FROM assets_environment WHERE assets_environment.asset_id = ?`

	envWSTSelect = `SELECT assets_supported_workload_types.* FROM assets_supported_workload_types`
	envSpecFirst = `SELECT assets_environment.args,assets_environment.command,assets_environment.environment_variables,assets_environment.run_as_gid,assets_environment.run_as_uid,assets_environment.supplemental_groups,assets_environment.allow_privilege_escalation,assets_environment.capabilities,assets_environment.create_home_dir,assets_environment.host_ipc,assets_environment.host_network,assets_environment.image,assets_environment.image_pull_policy,assets_environment.probes,assets_environment.read_only_root_filesystem,assets_environment.run_as_non_root,assets_environment.seccomp_profile_type,assets_environment.stdin,assets_environment.tty,assets_environment.uid_gid_source,assets_environment.working_dir,assets_environment.connections,assets_environment.override_uid_gid_in_workspace,assets_environment.routing_method FROM assets_environment WHERE assets_environment.asset_id = ? ORDER BY assets_environment.asset_id LIMIT ?`
	envSpecSave  = `UPDATE assets_environment SET args=?,command=?,environment_variables=?,run_as_gid=?,run_as_uid=?,supplemental_groups=?,allow_privilege_escalation=?,capabilities=?,create_home_dir=?,host_ipc=?,host_network=?,image=?,image_pull_policy=?,probes=?,read_only_root_filesystem=?,run_as_non_root=?,seccomp_profile_type=?,stdin=?,tty=?,uid_gid_source=?,working_dir=?,connections=?,override_uid_gid_in_workspace=?,routing_method=? WHERE assets_environment.asset_id = ?`
)

func envRow() []dbmodel.EnvironmentCombinedRec {
	return []dbmodel.EnvironmentCombinedRec{{
		AssetMeta:            api.AssetMeta{Id: "e-1"},
		EnvironmentAssetSpec: api.EnvironmentAssetSpec{Image: strptr("img"), Connections: []api.Connection{{Name: "c"}}},
	}}
}

func expectEnvMetaCreated(m dbtesting.Mocker, name string) {
	m.ExpectTransaction().Begin()
	m.ExpectExec(sqlLockDeleteStale, dbtesting.AnyTime()).Returns(0)
	m.ExpectExec(sqlLockInsert, dbtesting.AnyArg(), name, "environment", int32(0), "test", dbtesting.AnyTime()).Returns(1)
	m.ExpectTransaction().Commit()
	m.ExpectExec(sqlMetaInsert).Returns(1)
	m.ExpectQuery(sqlMetaSelect).Returns([]dbmodel.AssetMetaRec{})
	m.ExpectExec(envWSTInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Begin()
	m.ExpectExec(sqlLockDeleteById, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
}

func TestGetEnvironmentAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(envGet, "e-1", "environment", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(envRow())
	got, apiErr := dao.GetEnvironmentAssetById(auth.MockAuthInfo(), "e-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.Equal(t, "img", *got.Spec.Image)
}

func TestListEnvironmentAssetsUsingCredentialAsEnvVar(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(envCredList, "environment", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(envRow())
	cid := "cred-1"
	got, apiErr := dao.ListEnvironmentAssetsUsingCredentialAsEnvVar(auth.MockAuthInfo(), &cid)
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestCreateEnvironmentAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectEnvMetaCreated(m, "my-env")
	m.ExpectExec(envInsert).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateEnvironmentAsset(auth.MockAuthInfo(), api.EnvironmentCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-env", Scope: api.SCOPE_PROJECT, WorkloadSupportedTypes: &api.WorkloadSupportedTypes{}},
		Spec: api.EnvironmentAssetSpec{Image: strptr("img")},
	}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateEnvironmentAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(envWSTSelect).Returns([]dbmodel.WorkloadSupportedTypesRec{{}})
	m.ExpectQuery(envSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.EnvironmentAssetUpdate{{}})
	m.ExpectExec(envSpecSave).Returns(1)
	m.ExpectExec(sqlPolicyComplianceDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdateEnvironmentAsset(auth.MockAuthInfo(), "e-1", api.EnvironmentUpdateRequest{Meta: api.AssetUpdateRequest{Name: updateAssetMetaName}}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteEnvironmentAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(envHardDel, dbtesting.AnyArg()).Returns(1)
	expectDataSourceDeleteTail(m)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteEnvironmentAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ENVIRONMENT}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteEnvironmentAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "environment")
	apiErr := dao.DeleteEnvironmentAsset(auth.MockAuthInfo(), &dbmodel.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_ENVIRONMENT}, dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}
