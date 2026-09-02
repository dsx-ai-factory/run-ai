/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	sqlSelectHighestUpdateCount = `SELECT last_value FROM workloads_queue_sequence`
	sqlRestartUpdateCount       = `ALTER SEQUENCE workloads_queue_sequence RESTART WITH 1`
)

type lastValueRow struct {
	LastValue int64 `gorm:"column:last_value"`
}

func TestGetHighestUpdateCount(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlSelectHighestUpdateCount).
		Returns([]lastValueRow{{LastValue: 99}})

	got := GetHighestUpdateCount(ctxWithTenant())
	require.Equal(t, int64(99), got)
}

func TestGetHighestUpdateCount_OnDBFailure(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectQuery(sqlSelectHighestUpdateCount).
		ReturnsError(errors.New("boom"))

	got := GetHighestUpdateCount(ctxWithTenant())
	require.Equal(t, int64(0), got)
}

func TestRestartUpdateCountForTest(t *testing.T) {
	mocker := setupSQLMock(t)
	mocker.ExpectExec(sqlRestartUpdateCount).Returns(0)

	RestartUpdateCountForTest()
}
