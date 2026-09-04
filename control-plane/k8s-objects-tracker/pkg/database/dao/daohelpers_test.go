/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t dbtesting.TestingT) dbtesting.Mocker {
	return dbtesting.SetSQLMock(t)
}

type countRow struct {
	Count int64
}

func ctxWithTenant(tenantId int32) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
}
