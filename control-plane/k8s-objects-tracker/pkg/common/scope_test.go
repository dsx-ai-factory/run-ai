/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateScopeType", func() {
	DescribeTable("accepted values",
		func(scopeType string) {
			Expect(ValidateScopeType(scopeType)).To(BeNil())
		},
		Entry("empty", ""),
		Entry("tenant", "tenant"),
		Entry("cluster", "cluster"),
		Entry("department", "department"),
		Entry("project", "project"),
	)

	DescribeTable("rejected values",
		func(scopeType string) {
			err := ValidateScopeType(scopeType)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("invalid scopeType"))
		},
		Entry("system", "system"),
		Entry("unknown", "unknown"),
		Entry("sql injection", "z') OR TRUE --"),
	)
})
