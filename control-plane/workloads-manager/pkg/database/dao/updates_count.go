/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func GetHighestUpdateCount(ctx context.Context) int64 {
	var currentValue int64

	result := dao_utils.DB().WithContext(ctx).Table(models.SequenceWorkloadsCrud).Select(models.SequenceLastValue).Scan(&currentValue)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get highest update count: %v", result.Error)
		return 0
	}

	return currentValue
}

func RestartUpdateCountForTest() {
	dao_utils.DB().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", models.SequenceWorkloadsCrud))
}
