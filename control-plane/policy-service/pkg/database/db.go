/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// FinalizeTransaction closes the transaction, rollback it in case of error or panic
func FinalizeTransaction(tx orm.DB, committed *bool) {
	orm.FinalizeTransaction(tx, committed)
}
