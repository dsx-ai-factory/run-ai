/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

func GetHighestUpdateCount(ctx context.Context) int64 {
	var currentValue int64

	_, err := runaidb.DB().WithContext(ctx).Table(models.SequenceWorkloadsCrud).Select(models.SequenceLastValue).Scan(&currentValue)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get highest update count: %v", err)
		return 0
	}

	return currentValue
}

func RestartUpdateCountForTest() {
	db := runaidb.DB()
	if db == nil {
		panic("runaidb.DB() is nil; the test suite must call runaidb.Init before specs that touch the database")
	}
	_, _ = db.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", models.SequenceWorkloadsCrud))
}
