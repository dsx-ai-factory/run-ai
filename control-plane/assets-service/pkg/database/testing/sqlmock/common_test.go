/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	wblSelect      = `SELECT id FROM workloads WHERE tenant_id = ? AND project_id = ? ORDER BY id ASC LIMIT ?`
	wblAssetsDel   = `DELETE FROM workloads_assets WHERE workload_id IN (?)`
	wblWorkloadDel = `DELETE FROM workloads WHERE id IN (?)`
)

func workloadIDBatch(ids ...string) []struct {
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

func TestDeleteWorkloadsByScope(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(wblSelect, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(workloadIDBatch("wl-1"))
	m.ExpectExec(wblAssetsDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(wblWorkloadDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectQuery(wblSelect, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(workloadIDBatch())
	m.ExpectTransaction().Commit()

	err := dao.DeleteWorkloadsByScope(auth.MockAuthInfo(), api.SCOPETYPE_PROJECT, "proj-1")
	require.Nil(t, err)
}

func TestDeleteWorkloadsByScope_None(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectQuery(wblSelect, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(workloadIDBatch())
	m.ExpectTransaction().Commit()

	err := dao.DeleteWorkloadsByScope(auth.MockAuthInfo(), api.SCOPETYPE_PROJECT, "proj-1")
	require.Nil(t, err)
}
