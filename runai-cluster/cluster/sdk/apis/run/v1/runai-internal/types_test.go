/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Internal", func() {
	Describe("SetDefaultsWhereNeeded", func() {
		It("should set defaults", func() {
			internal := &Internal{}
			internal.SetDefaultsWhereNeeded()
			Expect(internal.Global).NotTo(BeNil())
			Expect(internal.Global.CustomCA).NotTo(BeNil())
			Expect(internal.Global.CustomCA.Enabled).To(Equal(ptr.To[bool](false)))
			Expect(internal.Global.CustomCA.Secret).NotTo(BeNil())

			Expect(internal.Global.CustomCA.Secret.Name).To(Equal("runai-ca-cert"))
			Expect(internal.Global.CustomCA.Secret.Key).To(Equal("runai-ca.pem"))
			Expect(internal.Global.CustomCA.Secret.Optional).To(Equal(ptr.To[bool](true)))
		})
	})
})
