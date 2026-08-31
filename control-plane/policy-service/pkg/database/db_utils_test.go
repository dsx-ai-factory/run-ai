/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("db_utils", func() {
	Context("error raised", func() {
		It("should return the correct error", func() {
			err := Error(DELETE, EffectivePolicy)
			Expect(int(err.Code)).To(Equal(int(http.StatusServiceUnavailable)))
			Expect(err.Message).To(ContainSubstring("failed to delete requested EffectivePolicy"))
		})
	})
})
