/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sqlSelectHighestUpdateCount = `SELECT last_value FROM workloads_queue_sequence`
	sqlRestartUpdateCount       = `ALTER SEQUENCE workloads_queue_sequence RESTART WITH 1`
)

type lastValueRow struct {
	LastValue int64 `gorm:"column:last_value"`
}

var _ = Describe("GetHighestUpdateCount", func() {
	It("returns the sequence last value", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlSelectHighestUpdateCount).
			Returns([]lastValueRow{{LastValue: 99}})

		Expect(GetHighestUpdateCount(ctxWithTenant())).To(Equal(int64(99)))
	})

	It("returns 0 when the query fails", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectQuery(sqlSelectHighestUpdateCount).
			ReturnsError(errors.New("boom"))

		Expect(GetHighestUpdateCount(ctxWithTenant())).To(Equal(int64(0)))
	})
})

var _ = Describe("RestartUpdateCountForTest", func() {
	It("restarts the sequence", func() {
		mocker := setupSQLMock(GinkgoT())
		mocker.ExpectExec(sqlRestartUpdateCount).Returns(0)

		RestartUpdateCountForTest()
	})
})
