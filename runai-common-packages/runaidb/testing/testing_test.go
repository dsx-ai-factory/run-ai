/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing_test

import (
	"testing"

	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

func TestSetOrmMock(t *testing.T) {
	db := dbtesting.SetOrmMock(t)
	require.NotNil(t, db)
}

func TestSetSQLMock(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	require.NotNil(t, mocker)
}
