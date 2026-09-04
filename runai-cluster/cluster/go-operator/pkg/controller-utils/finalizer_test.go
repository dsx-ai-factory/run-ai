/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

var _ = Describe("Finalizer", func() {
	var (
		resource *runaiv1.RunaiConfig
	)

	BeforeEach(func() {
		resource = &runaiv1.RunaiConfig{}
	})

	Context("HasFinalizer()", func() {
		It("should return false when a finalizer is missing", func() {
			Expect(HasFinalizer(resource, runaiv1.Finalizer)).To(BeFalse())
		})

		It("should return true when a finalizer exists", func() {
			resource.SetFinalizers([]string{runaiv1.Finalizer})
			Expect(HasFinalizer(resource, runaiv1.Finalizer)).To(BeTrue())
		})
	})

	Context("AddFinalizer()", func() {
		It("should add a finalizer", func() {
			Expect(resource.GetFinalizers()).To(BeEmpty())
			AddFinalizer(resource, runaiv1.Finalizer)
			Expect(resource.GetFinalizers()).To(ContainElement(runaiv1.Finalizer))
		})
	})

	Context("RemoveFinalizer()", func() {
		It("should remove a finalizer", func() {
			AddFinalizer(resource, runaiv1.Finalizer)
			RemoveFinalizer(resource, runaiv1.Finalizer)
			Expect(resource.GetFinalizers()).To(BeEmpty())
		})
	})
})
