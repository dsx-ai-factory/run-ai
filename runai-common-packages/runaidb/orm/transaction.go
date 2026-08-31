/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orm

import "github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"

func FinalizeTransaction(tx DB, committed *bool) {
	if !*committed {
		logger.GetLogger().Errorf("transaction was not committed, undoing all changes")
		if err := tx.Rollback(); err != nil {
			logger.GetLogger().Errorf("failed to rollback uncommitted transaction: %v", err)
		}
		return
	}

	if r := recover(); r != nil {
		logger.GetLogger().Errorf("panic in transaction: %v, undoing all changes", r)
		if err := tx.Rollback(); err != nil {
			logger.GetLogger().Errorf("failed to rollback transaction after panic: %v", err)
		}
	}
}
