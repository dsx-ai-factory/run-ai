/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("NodePool Controller", func() {
	Describe("SetDefaultsWhereNeeded", func() {
		It("should set Enabled to false, since KRM owns the nodepool-controller", func() {
			c := &NodePoolController{}
			c.SetDefaultsWhereNeeded("registry", ptr.To[int32](1))
			Expect(c.Enabled).To(Equal(ptr.To(false)))
		})

		It("should keep Enabled false even when explicitly set to true", func() {
			c := &NodePoolController{Enabled: ptr.To(true)}
			c.SetDefaultsWhereNeeded("registry", ptr.To[int32](1))
			Expect(c.Enabled).To(Equal(ptr.To(false)))
		})
	})
})
