/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MultiWorkloadsError", func() {
	var (
		workloads []MinimalWorkloadInfo
		err       error
	)

	BeforeEach(func() {
		workloads = []MinimalWorkloadInfo{
			{Name: "w1", Project: "p1", Kind: "k1", ID: "id1"},
			{Name: "w2", Project: "p2", Kind: "k2", ID: "id2"},
		}
		err = NewMultiWorkloadsError(workloads)
	})

	Describe("NewMultiWorkloadsError", func() {
		It("should create a new MultiWorkloadsError with the given workloads", func() {
			Expect(err).To(BeAssignableToTypeOf(MultiWorkloadsError{}))
			Expect(err.(MultiWorkloadsError).Workloads).To(Equal(workloads)) //nolint:errorlint // just test if it fails it fails
		})
	})

	Describe("Error", func() {
		It("should return the correct error message", func() {
			expected := fmt.Sprintf("Found multiple workloads. Found: %v", workloads)
			Expect(err.Error()).To(Equal(expected))
		})
	})
})
