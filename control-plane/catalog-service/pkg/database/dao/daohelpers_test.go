/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"testing"

	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

func setupDaoMock(t *testing.T) dbtesting.Mocker {
	mocker := dbtesting.SetSQLMock(t)
	return mocker
}
