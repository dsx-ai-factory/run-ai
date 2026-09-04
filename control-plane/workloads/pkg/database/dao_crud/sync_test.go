/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	syncSequenceSelect = "SELECT last_value FROM crud_workloads_sequence"
)

type syncSequenceRow struct {
	LastValue int64
}

var _ = Describe("GetHighestUpdateCount", func() {
	It("returns the highest update count", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(syncSequenceSelect).Returns([]syncSequenceRow{{LastValue: 42}})

		value, apiErr := GetHighestUpdateCount(testCtx(1))

		Expect(apiErr).To(BeNil())
		Expect(value).To(Equal(int64(42)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(syncSequenceSelect).ReturnsError(errors.New("boom"))

		value, apiErr := GetHighestUpdateCount(testCtx(1))

		Expect(apiErr).ToNot(BeNil())
		Expect(value).To(Equal(int64(0)))
	})
})
