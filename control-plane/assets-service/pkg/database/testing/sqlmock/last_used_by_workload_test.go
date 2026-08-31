/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	lubwFirstOrCreate = `SELECT * FROM last_used_by_workload WHERE last_used_by_workload.project_id = ? AND last_used_by_workload.asset_id = ? AND last_used_by_workload.workload_type = ? ORDER BY last_used_by_workload.id LIMIT ?`
	lubwList          = `SELECT * FROM last_used_by_workload WHERE (last_used_by_workload.project_id = ? AND last_used_by_workload.workload_type = ?) AND asset_id in (?)`
	lubwInsert        = `INSERT INTO last_used_by_workload (id,tenant_id,project_id,asset_id,workload_type,workload_id,workload_creation_time) VALUES (?,?,?,?,?,?,?)`
)

func TestUpsertAssetLastUsedByWorkload(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(lubwFirstOrCreate, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.LastUsedByWorkloadRecord{})
	m.ExpectTransaction().Begin()
	m.ExpectExec(lubwInsert).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.UpsertAssetLastUsedByWorkload(auth.MockAuthInfo(), "a-1", "wtype", "w-1", int32(3), time.Now())
	require.Nil(t, apiErr)
}

func TestListUsageTimes(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(lubwList, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).
		Returns([]dbmodel.LastUsedByWorkloadRecord{{AssetId: "a-1"}})
	got, apiErr := dao.ListUsageTimes(auth.MockAuthInfo(), "wtype", int32(3), []string{"a-1"})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
