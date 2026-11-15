/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const assetMetaScopeSelect = `SELECT id FROM assets_metadata WHERE tenant_id = ? AND project_id = ? AND kind = ? AND assets_metadata.deleted_at IS NULL ORDER BY id ASC LIMIT ?`

func assetIDBatch(ids ...string) []struct {
	Id string `gorm:"column:id"`
} {
	out := make([]struct {
		Id string `gorm:"column:id"`
	}, 0, len(ids))
	for _, id := range ids {
		out = append(out, struct {
			Id string `gorm:"column:id"`
		}{Id: id})
	}
	return out
}

func TestDeleteAssetsByScope(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(`SELECT id FROM assets_metadata WHERE project_id = ? AND kind = ? AND tenant_id = ? ORDER BY id ASC LIMIT ?`, "proj-1", "workload-template", int32(0), 100).Returns(assetIDBatch())
	m.ExpectTransaction().Commit()

	const kinds = 17
	for i := 0; i < kinds; i++ {
		m.ExpectTransaction().Begin()
		m.ExpectQuery(assetMetaScopeSelect, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(assetIDBatch())
		m.ExpectTransaction().Commit()
	}

	err := dao.DeleteAssetsByScope(auth.MockAuthInfo(), api.SCOPETYPE_PROJECT, "proj-1")
	require.Nil(t, err)
}

func TestDeleteAssetsByScopeLegacyFailure(t *testing.T) {
	m := setupDaoMock(t)
	wantErr := errors.New("legacy template deletion failed")
	m.ExpectTransaction().Begin()
	m.ExpectQuery(`SELECT id FROM assets_metadata WHERE project_id = ? AND kind = ? AND tenant_id = ? ORDER BY id ASC LIMIT ?`, "proj-1", "workload-template", int32(0), 100).Returns(assetIDBatch("template-1"))
	m.ExpectExec(`DELETE FROM assets_template_content WHERE template_asset_id IN (?)`, "template-1").Returns(1)
	m.ExpectExec(`DELETE FROM assets_template WHERE asset_id IN (?)`, "template-1").ReturnsError(wantErr)
	m.ExpectTransaction().Rollback()

	err := dao.DeleteAssetsByScope(auth.MockAuthInfo(), api.SCOPETYPE_PROJECT, "proj-1")
	require.ErrorIs(t, err, wantErr)
}

func TestValidateClaimNameUniqueness(t *testing.T) {
	m := setupDaoMock(t)

	q := database.NewParametrizedQuery()
	q.Add("assets_metadata.tenant_id = ?", int32(0))
	q.Add("AND assets_pvc.claim_name = ?", "my-claim")

	m.ExpectQuery(`SELECT count(*) FROM assets_pvc INNER JOIN assets_metadata ON assets_metadata.id = assets_pvc.asset_id WHERE assets_metadata.tenant_id = ? AND assets_pvc.claim_name = ?`, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())

	tx := runaidb.DB().WithContext(auth.MockAuthInfo())
	err := dao.ValidateClaimNameUniqueness(auth.MockAuthInfo(), tx, "my-claim", q)
	require.Nil(t, err)
}
