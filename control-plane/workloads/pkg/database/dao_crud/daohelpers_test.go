/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"fmt"

	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t dbtesting.TestingT) dbtesting.Mocker {
	gormDB, mocker := dbtesting.NewGormMocker(t)
	dao_utils.SetMock(gormDB)
	t.Cleanup(dao_utils.ResetMock)
	return mocker
}

func testCtx(tenantId uint) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
}

type countRow struct {
	Count int64
}
