/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	autoDelSelect = `SELECT assets_metadata.*, assets_supported_workload_types.* FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id WHERE COALESCE(assets_metadata.private, FALSE) = false AND (assets_metadata.auto_delete = ? AND assets_metadata.kind = ? AND assets_metadata.tenant_id = ?) AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND id NOT in (select contained_asset_id from assets_template_content) AND id NOT in (select asset_id from workloads_assets) AND assets_metadata.created_at < ? AND assets_metadata.deleted_at IS NULL`

	autoDelSoft   = `UPDATE assets_metadata SET deleted_at=? WHERE id in (?) AND assets_metadata.deleted_at IS NULL`
	autoDelPvc    = `DELETE FROM assets_pvc WHERE asset_id in (?)`
	autoDelPolicy = `DELETE FROM policy_compliance WHERE asset_id in (?)`
	autoDelMeta   = `DELETE FROM assets_metadata WHERE id in (?)`
)

func autoDelCandidates() []dbmodel.AssetMetaRec {
	return []dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "auto-1"}}}
}

func TestAutoDeletePvcAssets_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(autoDelSelect).Returns(autoDelCandidates())
	m.ExpectTransaction().Begin()
	m.ExpectExec(autoDelSoft, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.AutoDeletePvcAssets(auth.MockAuthInfo(), dao.SoftDelete)
	require.Nil(t, apiErr)
}

func TestAutoDeletePvcAssets_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(autoDelSelect).Returns(autoDelCandidates())
	m.ExpectTransaction().Begin()
	m.ExpectExec(autoDelPvc, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(autoDelPolicy, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(autoDelMeta, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.AutoDeletePvcAssets(auth.MockAuthInfo(), dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestAutoDeletePvcAssets_None(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(autoDelSelect).Returns([]dbmodel.AssetMetaRec{})
	apiErr := dao.AutoDeletePvcAssets(auth.MockAuthInfo(), dao.SoftDelete)
	require.Nil(t, apiErr)
}
