/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbtesting

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	syncUpdateCounter = "effective_policies_update_count"
)

func RestartUpdateCounter() {
	_, _ = runaidb.DB().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", syncUpdateCounter))
}
